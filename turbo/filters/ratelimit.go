package filters

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimitConfig configures the per-principal token-bucket rate limiter.
//
// Tokens-per-second + Burst control the bucket; Principal extracts the
// rate-limit key from a request (user id, API key, remote IP, etc.). If
// Principal returns an empty string the request is allowed through
// unconditionally — useful for letting unauthenticated paths bypass.
type RateLimitConfig struct {
	// Rate is the steady-state tokens per second added to each bucket.
	Rate float64
	// Burst is the maximum bucket size (and the per-key initial tokens).
	Burst int
	// Principal extracts the rate-limit key. The default keys by
	// req.RemoteAddr — useful for naïve IP throttling but typically you
	// want to swap this for an auth-derived key.
	Principal func(r *http.Request) string
	// OnLimit is called when a request is rejected. Defaults to writing
	// 429 with a Retry-After header equal to the time until the next token.
	OnLimit func(w http.ResponseWriter, r *http.Request, retryAfter time.Duration)
	// Now overrides the time source (testing).
	Now func() time.Time
}

// RateLimit returns an http middleware that enforces a token-bucket policy.
// State is held in process — for distributed deployments wire up a store
// that satisfies the unexported bucketStore interface (extension point
// reserved; see roadmap).
//
//	router.AddGlobalFilter(filters.RateLimit(filters.RateLimitConfig{
//	    Rate:      5,        // 5 requests/sec steady state
//	    Burst:     20,       // bucket size
//	    Principal: filters.PrincipalFromHeader("X-User-Id"),
//	}))
func RateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	if cfg.Rate <= 0 {
		panic("turbo/filters: RateLimit Rate must be > 0")
	}
	if cfg.Burst <= 0 {
		cfg.Burst = 1
	}
	if cfg.Principal == nil {
		cfg.Principal = func(r *http.Request) string { return r.RemoteAddr }
	}
	if cfg.OnLimit == nil {
		cfg.OnLimit = defaultOnLimit
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}

	store := &memBucketStore{
		rate:    cfg.Rate,
		burst:   float64(cfg.Burst),
		now:     cfg.Now,
		buckets: make(map[string]*bucket),
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := cfg.Principal(r)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}
			if ok, retryAfter := store.take(key); !ok {
				cfg.OnLimit(w, r, retryAfter)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// PrincipalFromHeader returns a Principal extractor that reads the given
// header. Empty/missing → empty key (the request bypasses the limiter).
func PrincipalFromHeader(name string) func(*http.Request) string {
	return func(r *http.Request) string {
		return strings.TrimSpace(r.Header.Get(name))
	}
}

// --- internals ---

type bucket struct {
	tokens float64
	last   time.Time
}

type memBucketStore struct {
	mu      sync.Mutex
	rate    float64
	burst   float64
	now     func() time.Time
	buckets map[string]*bucket
}

// take attempts to remove one token from key's bucket. Returns
// (allowed, retryAfter). retryAfter is 0 when allowed.
func (s *memBucketStore) take(key string) (bool, time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	b, ok := s.buckets[key]
	if !ok {
		b = &bucket{tokens: s.burst, last: now}
		s.buckets[key] = b
	}
	// Refill.
	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * s.rate
		if b.tokens > s.burst {
			b.tokens = s.burst
		}
		b.last = now
	}
	if b.tokens >= 1 {
		b.tokens--
		return true, 0
	}
	// Compute when the next token arrives.
	need := 1 - b.tokens
	seconds := need / s.rate
	return false, time.Duration(seconds * float64(time.Second))
}

func defaultOnLimit(w http.ResponseWriter, _ *http.Request, retryAfter time.Duration) {
	secs := int(retryAfter.Seconds())
	if secs < 1 {
		secs = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(secs))
	http.Error(w, fmt.Sprintf("rate limit exceeded; retry after %ds", secs), http.StatusTooManyRequests)
}
