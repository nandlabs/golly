package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestInMemory_SetGet(t *testing.T) {
	c := NewInMemory[string, int]()
	ctx := context.Background()

	if err := c.Set(ctx, "a", 1); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if v, ok := c.Get(ctx, "a"); !ok || v != 1 {
		t.Fatalf("Get(a) = (%d, %v), want (1, true)", v, ok)
	}
	if _, ok := c.Get(ctx, "missing"); ok {
		t.Fatalf("Get(missing) should report ok=false")
	}
}

func TestInMemory_TTLExpiry(t *testing.T) {
	now := time.Now()
	clock := &fakeClock{now: now}
	c := NewInMemory[string, string](withClock[string, string](clock.Now))
	ctx := context.Background()

	if err := c.SetWithTTL(ctx, "k", "v", 100*time.Millisecond); err != nil {
		t.Fatalf("SetWithTTL: %v", err)
	}
	// Within TTL: present
	if _, ok := c.Get(ctx, "k"); !ok {
		t.Fatalf("expected Get(k) ok within TTL")
	}
	// Advance past TTL — lazy expiry on next Get
	clock.advance(101 * time.Millisecond)
	if _, ok := c.Get(ctx, "k"); ok {
		t.Fatalf("expected Get(k) ok=false past TTL")
	}
	// Entry should also be removed from underlying map (lazy delete).
	if got := c.Len(ctx); got != 0 {
		t.Fatalf("expected Len=0 after lazy delete; got %d", got)
	}
}

func TestInMemory_NoExpiryBehaviour(t *testing.T) {
	clock := &fakeClock{now: time.Now()}
	c := NewInMemory[string, int](withClock[string, int](clock.Now))
	ctx := context.Background()

	// Both Set and SetWithTTL(..., NoExpiry) must mean "never expires".
	_ = c.Set(ctx, "a", 1)
	_ = c.SetWithTTL(ctx, "b", 2, NoExpiry)

	clock.advance(365 * 24 * time.Hour) // 1 year later
	for _, k := range []string{"a", "b"} {
		if _, ok := c.Get(ctx, k); !ok {
			t.Errorf("Get(%q) expired but should be NoExpiry", k)
		}
	}
}

func TestInMemory_Delete(t *testing.T) {
	c := NewInMemory[int, int]()
	ctx := context.Background()
	_ = c.Set(ctx, 1, 10)
	if !c.Delete(ctx, 1) {
		t.Fatalf("Delete(1) should return true when present")
	}
	if c.Delete(ctx, 1) {
		t.Fatalf("Delete(1) should return false after removal")
	}
	if _, ok := c.Get(ctx, 1); ok {
		t.Fatalf("Get(1) should be absent after Delete")
	}
}

func TestInMemory_Has(t *testing.T) {
	clock := &fakeClock{now: time.Now()}
	c := NewInMemory[string, int](withClock[string, int](clock.Now))
	ctx := context.Background()
	_ = c.SetWithTTL(ctx, "k", 1, 50*time.Millisecond)
	if !c.Has(ctx, "k") {
		t.Fatalf("Has(k) should be true within TTL")
	}
	clock.advance(60 * time.Millisecond)
	if c.Has(ctx, "k") {
		t.Fatalf("Has(k) should be false past TTL")
	}
}

func TestInMemory_Clear(t *testing.T) {
	c := NewInMemory[string, int]()
	ctx := context.Background()
	_ = c.Set(ctx, "a", 1)
	_ = c.Set(ctx, "b", 2)
	if err := c.Clear(ctx); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if c.Len(ctx) != 0 {
		t.Fatalf("expected Len=0 after Clear; got %d", c.Len(ctx))
	}
}

func TestInMemory_Sweep(t *testing.T) {
	clock := &fakeClock{now: time.Now()}
	c := NewInMemory[string, int](withClock[string, int](clock.Now))
	ctx := context.Background()

	_ = c.SetWithTTL(ctx, "a", 1, 10*time.Millisecond)
	_ = c.SetWithTTL(ctx, "b", 2, 10*time.Millisecond)
	_ = c.Set(ctx, "c", 3) // no expiry

	clock.advance(20 * time.Millisecond)
	if n := c.Sweep(); n != 2 {
		t.Fatalf("Sweep removed %d, want 2", n)
	}
	if c.Len(ctx) != 1 {
		t.Fatalf("expected Len=1 after Sweep; got %d", c.Len(ctx))
	}
}

func TestInMemory_ClosedRejectsWrites(t *testing.T) {
	c := NewInMemory[string, int]()
	if err := c.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// Close is idempotent.
	if err := c.Close(); err != nil {
		t.Fatalf("Close (second call): %v", err)
	}
	ctx := context.Background()
	if err := c.Set(ctx, "k", 1); err != ErrClosed {
		t.Fatalf("Set after Close should return ErrClosed; got %v", err)
	}
	if err := c.SetWithTTL(ctx, "k", 1, time.Second); err != ErrClosed {
		t.Fatalf("SetWithTTL after Close should return ErrClosed; got %v", err)
	}
	if err := c.Clear(ctx); err != ErrClosed {
		t.Fatalf("Clear after Close should return ErrClosed; got %v", err)
	}
}

func TestInMemory_Janitor(t *testing.T) {
	// Real-time janitor (small interval); not using fakeClock here because
	// the janitor uses time.Ticker which can't be faked without extra
	// indirection — and the loop itself is the thing under test.
	c := NewInMemory[string, int](WithJanitor[string, int](20 * time.Millisecond))
	ctx := context.Background()
	_ = c.SetWithTTL(ctx, "k", 1, 10*time.Millisecond)
	// Wait for the janitor to sweep at least once after expiry.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if c.Len(ctx) == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if c.Len(ctx) != 0 {
		t.Fatalf("janitor never swept expired entry; Len=%d", c.Len(ctx))
	}
	if err := c.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestInMemory_Concurrent(t *testing.T) {
	c := NewInMemory[int, int]()
	ctx := context.Background()
	const writers, readers, iters = 8, 8, 500

	var wg sync.WaitGroup
	wg.Add(writers + readers)

	var stops atomic.Int32
	for w := 0; w < writers; w++ {
		go func(w int) {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				if i%5 == 0 {
					_ = c.SetWithTTL(ctx, w*100+i, i, time.Millisecond)
				} else {
					_ = c.Set(ctx, w*100+i, i)
				}
				if i%17 == 0 {
					c.Delete(ctx, w*100+i)
				}
			}
		}(w)
	}
	for r := 0; r < readers; r++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				_, _ = c.Get(ctx, i)
				c.Has(ctx, i)
				stops.Add(1)
			}
		}()
	}
	wg.Wait()
	if stops.Load() != int32(readers*iters) {
		t.Fatalf("reader goroutines did not all complete: %d", stops.Load())
	}
}

// fakeClock is a monotonically advancing test clock.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (f *fakeClock) Now() time.Time { f.mu.Lock(); defer f.mu.Unlock(); return f.now }
func (f *fakeClock) advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
}
