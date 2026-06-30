package cache

import (
	"context"
	"sync"
	"time"
)

// InMemory is a goroutine-safe in-memory Cache. Entries with a TTL are
// expired lazily on access; an optional janitor goroutine sweeps stale
// entries on a fixed interval to bound memory use under churn (enable
// via WithJanitor).
//
// Suitable for single-process caches. For cross-process / cross-node
// caching, implement Cache with a network backend (Redis, Memcached,
// cloud caches) in a separate module.
type InMemory[K comparable, V any] struct {
	mu       sync.RWMutex
	entries  map[K]entry[V]
	closed   bool
	stopJani chan struct{}
	now      func() time.Time // injectable clock for tests
}

type entry[V any] struct {
	value   V
	expires time.Time // zero = never expires
}

// InMemoryOption configures an InMemory cache at construction time.
type InMemoryOption[K comparable, V any] func(*InMemory[K, V])

// WithJanitor runs a background goroutine that calls Sweep every interval.
// Pass 0 (the default) or a negative duration to disable the janitor.
// The goroutine stops on Close.
func WithJanitor[K comparable, V any](interval time.Duration) InMemoryOption[K, V] {
	return func(c *InMemory[K, V]) {
		if interval <= 0 {
			return
		}
		stop := make(chan struct{})
		c.stopJani = stop
		// Capture `stop` in the goroutine instead of reading c.stopJani —
		// avoids a race with Close, which mutates c.stopJani under lock.
		go c.janitorLoop(interval, stop)
	}
}

// withClock injects a clock for deterministic tests; unexported on purpose.
func withClock[K comparable, V any](now func() time.Time) InMemoryOption[K, V] {
	return func(c *InMemory[K, V]) { c.now = now }
}

// NewInMemory constructs an empty in-memory cache. Options are applied in
// order; later options override earlier ones for the same field.
func NewInMemory[K comparable, V any](opts ...InMemoryOption[K, V]) *InMemory[K, V] {
	c := &InMemory[K, V]{
		entries: make(map[K]entry[V]),
		now:     time.Now,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Get returns the value for key or (zero, false) when absent or expired.
func (c *InMemory[K, V]) Get(_ context.Context, key K) (V, bool) {
	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()
	var zero V
	if !ok {
		return zero, false
	}
	if c.isExpired(e) {
		// Best-effort lazy delete. Held under read lock above, so promote.
		c.mu.Lock()
		// Re-check under write lock — another goroutine may have replaced it.
		if cur, ok := c.entries[key]; ok && c.isExpired(cur) {
			delete(c.entries, key)
		}
		c.mu.Unlock()
		return zero, false
	}
	return e.value, true
}

// Set stores value under key with no expiry.
func (c *InMemory[K, V]) Set(ctx context.Context, key K, value V) error {
	return c.SetWithTTL(ctx, key, value, NoExpiry)
}

// SetWithTTL stores value under key with the given TTL.
func (c *InMemory[K, V]) SetWithTTL(_ context.Context, key K, value V, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return ErrClosed
	}
	e := entry[V]{value: value}
	if ttl > 0 {
		e.expires = c.now().Add(ttl)
	}
	c.entries[key] = e
	return nil
}

// Delete removes key. Returns true when the key was present.
func (c *InMemory[K, V]) Delete(_ context.Context, key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.entries[key]; ok {
		delete(c.entries, key)
		return true
	}
	return false
}

// Has reports whether key is present and not expired.
func (c *InMemory[K, V]) Has(ctx context.Context, key K) bool {
	_, ok := c.Get(ctx, key)
	return ok
}

// Clear removes every entry.
func (c *InMemory[K, V]) Clear(_ context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return ErrClosed
	}
	c.entries = make(map[K]entry[V])
	return nil
}

// Len returns the number of stored entries, including any that are
// expired but not yet swept.
func (c *InMemory[K, V]) Len(_ context.Context) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Sweep removes all expired entries and returns the count removed.
// Implements Sweeper.
func (c *InMemory[K, V]) Sweep() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return 0
	}
	n := 0
	for k, e := range c.entries {
		if c.isExpired(e) {
			delete(c.entries, k)
			n++
		}
	}
	return n
}

// Close stops the janitor (if any) and marks the cache closed. Subsequent
// Set/SetWithTTL/Clear return ErrClosed; Get/Has/Len/Delete continue to
// work on the existing snapshot for graceful drain (idempotent).
func (c *InMemory[K, V]) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	if c.stopJani != nil {
		// Closing signals the janitor goroutine (which captured the same
		// channel as a local) to exit. We leave c.stopJani non-nil so a
		// second Close() doesn't re-close (guarded by c.closed) and so
		// concurrent readers observe a stable value.
		close(c.stopJani)
	}
	return nil
}

func (c *InMemory[K, V]) isExpired(e entry[V]) bool {
	return !e.expires.IsZero() && !c.now().Before(e.expires)
}

// janitorLoop sweeps expired entries every interval until stop is closed.
// stop is passed in (not read from c.stopJani) so the loop has no race
// with Close, which mutates c.stopJani under the write lock.
func (c *InMemory[K, V]) janitorLoop(interval time.Duration, stop <-chan struct{}) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			c.Sweep()
		case <-stop:
			return
		}
	}
}

// Compile-time assertions that InMemory satisfies the public interfaces.
var (
	_ Cache[string, int] = (*InMemory[string, int])(nil)
	_ Sweeper            = (*InMemory[string, int])(nil)
)
