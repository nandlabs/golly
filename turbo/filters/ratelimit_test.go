package filters

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimit_BurstAndRefill(t *testing.T) {
	now := time.Unix(0, 0)
	mw := RateLimit(RateLimitConfig{
		Rate:      1, // 1 token/sec
		Burst:     3,
		Principal: PrincipalFromHeader("X-User"),
		Now:       func() time.Time { return now },
	})
	var served int32
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&served, 1)
		w.WriteHeader(http.StatusOK)
	}))

	// First 3 requests within burst should all succeed.
	for i := 0; i < 3; i++ {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("X-User", "alice")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("req %d: status = %d, want 200", i, w.Code)
		}
	}
	// 4th should be rejected.
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-User", "alice")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("4th req: status = %d, want 429", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Error("Retry-After header missing on 429")
	}
	// Advance time by 1.1s — one token refilled — next should pass.
	now = now.Add(1100 * time.Millisecond)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-User", "alice")
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("after refill: status = %d, want 200", w.Code)
	}
	if atomic.LoadInt32(&served) != 4 {
		t.Errorf("served = %d, want 4", served)
	}
}

func TestRateLimit_PerPrincipalIsolation(t *testing.T) {
	mw := RateLimit(RateLimitConfig{
		Rate:      0.001, // basically frozen
		Burst:     1,
		Principal: PrincipalFromHeader("X-User"),
	})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))

	// alice spends her token.
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest(http.MethodGet, "/", nil)
	r1.Header.Set("X-User", "alice")
	h.ServeHTTP(w1, r1)
	if w1.Code != 200 {
		t.Fatalf("alice 1: %d", w1.Code)
	}
	// alice's 2nd is denied.
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	r2.Header.Set("X-User", "alice")
	h.ServeHTTP(w2, r2)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("alice 2: %d, want 429", w2.Code)
	}
	// bob is independent.
	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest(http.MethodGet, "/", nil)
	r3.Header.Set("X-User", "bob")
	h.ServeHTTP(w3, r3)
	if w3.Code != 200 {
		t.Errorf("bob 1: %d, want 200", w3.Code)
	}
}

func TestRateLimit_EmptyPrincipalBypass(t *testing.T) {
	mw := RateLimit(RateLimitConfig{
		Rate:      0.001,
		Burst:     1,
		Principal: func(r *http.Request) string { return "" }, // always empty
	})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		if w.Code != http.StatusOK {
			t.Errorf("req %d: status = %d, want 200 (empty principal should bypass)", i, w.Code)
		}
	}
}
