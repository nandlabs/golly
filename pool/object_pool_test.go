package pool

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// --- helpers ---

var errCreationFailed = errors.New("creation failed")

func intCreator() (int, error) {
	return 42, nil
}

var nextID atomic.Int64

func uniqueIntCreator() (int, error) {
	return int(nextID.Add(1)), nil
}

func failingCreator() (int, error) {
	return 0, errCreationFailed
}

func noopDestroyer(v int) error {
	return nil
}

// --- NewPool ---

func TestNewPool_ValidConfig(t *testing.T) {
	p, err := NewPool(intCreator, noopDestroyer, 2, 5, 3)
	if err != nil {
		t.Fatalf("NewPool returned error: %v", err)
	}
	if p == nil {
		t.Fatal("NewPool returned nil")
	}
	if p.Min() != 2 {
		t.Errorf("Min() = %d, want 2", p.Min())
	}
	if p.Max() != 5 {
		t.Errorf("Max() = %d, want 5", p.Max())
	}
	if p.MaxWait() != 3 {
		t.Errorf("MaxWait() = %d, want 3", p.MaxWait())
	}
}

func TestNewPool_NilCreator(t *testing.T) {
	_, err := NewPool(nil, noopDestroyer, 0, 5, 1)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestNewPool_ZeroMax(t *testing.T) {
	_, err := NewPool(intCreator, noopDestroyer, 0, 0, 1)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestNewPool_MinExceedsMax_Clamped(t *testing.T) {
	p, err := NewPool(intCreator, noopDestroyer, 10, 5, 1)
	if err != nil {
		t.Fatalf("NewPool returned error: %v", err)
	}
	if p.Min() != 5 {
		t.Errorf("Min() = %d, want 5 (clamped to max)", p.Min())
	}
}

func TestNewPool_NilDestroyer(t *testing.T) {
	p, err := NewPool(intCreator, nil, 0, 5, 1)
	if err != nil {
		t.Fatalf("NewPool should allow nil destroyer, got: %v", err)
	}
	if p.Destroyer() != nil {
		t.Error("Destroyer() should be nil")
	}
}

// --- Start ---

func TestStart_PreCreatesMinObjects(t *testing.T) {
	p, _ := NewPool(uniqueIntCreator, noopDestroyer, 3, 5, 1)
	if err := p.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer p.Close()

	if p.Current() != 3 {
		t.Errorf("Current() = %d, want 3 after Start", p.Current())
	}
	if p.HighWaterMark() != 3 {
		t.Errorf("HighWaterMark() = %d, want 3", p.HighWaterMark())
	}
}

func TestStart_FailingCreator(t *testing.T) {
	p, _ := NewPool(failingCreator, noopDestroyer, 2, 5, 1)
	err := p.Start()
	if err == nil {
		t.Fatal("Start() should return error when creator fails")
	}
}

// --- Checkout ---

func TestCheckout_ReturnsPreCreatedObject(t *testing.T) {
	p, _ := NewPool(intCreator, noopDestroyer, 1, 5, 1)
	p.Start()
	defer p.Close()

	v, err := p.Checkout()
	if err != nil {
		t.Fatalf("Checkout() error: %v", err)
	}
	if v != 42 {
		t.Errorf("Checkout() = %d, want 42", v)
	}
}

func TestCheckout_CreatesNewWhenPoolEmpty(t *testing.T) {
	p, _ := NewPool(uniqueIntCreator, noopDestroyer, 0, 5, 1)
	p.Start()
	defer p.Close()

	v, err := p.Checkout()
	if err != nil {
		t.Fatalf("Checkout() error: %v", err)
	}
	if v == 0 {
		t.Error("Checkout() returned zero value, expected a created object")
	}
	if p.Current() != 1 {
		t.Errorf("Current() = %d, want 1", p.Current())
	}
}

func TestCheckout_ReturnsErrorWhenPoolExhausted(t *testing.T) {
	p, _ := NewPool(uniqueIntCreator, noopDestroyer, 0, 1, 1)
	p.Start()
	defer p.Close()

	// Exhaust the pool
	_, err := p.Checkout()
	if err != nil {
		t.Fatalf("first Checkout() error: %v", err)
	}

	// Second checkout should timeout (maxWait=1s)
	start := time.Now()
	_, err = p.Checkout()
	elapsed := time.Since(start)

	if !errors.Is(err, ErrCacheFull) {
		t.Fatalf("expected ErrCacheFull, got %v", err)
	}
	if elapsed < 900*time.Millisecond {
		t.Errorf("expected to wait ~1s, waited %v", elapsed)
	}
}

func TestCheckout_ReturnsErrorWhenCreatorFails(t *testing.T) {
	calls := 0
	creator := func() (int, error) {
		calls++
		if calls > 1 {
			return 0, errCreationFailed
		}
		return 1, nil
	}

	p, _ := NewPool(creator, noopDestroyer, 0, 5, 1)
	p.Start()
	defer p.Close()

	// First checkout succeeds
	v, err := p.Checkout()
	if err != nil {
		t.Fatalf("first Checkout() error: %v", err)
	}
	p.Checkin(v)

	// Checkout the pre-existing object from pool
	_, err = p.Checkout()
	if err != nil {
		t.Fatalf("second Checkout() error: %v", err)
	}

	// Third checkout needs creation and should fail
	_, err = p.Checkout()
	if err == nil {
		t.Fatal("expected error from failing creator")
	}
}

func TestCheckout_AfterClose(t *testing.T) {
	p, _ := NewPool(intCreator, noopDestroyer, 0, 5, 1)
	p.Start()
	p.Close()

	_, err := p.Checkout()
	if !errors.Is(err, ErrPoolClosed) {
		t.Fatalf("expected ErrPoolClosed, got %v", err)
	}
}

// --- Checkin ---

func TestCheckin_ReturnsObjectToPool(t *testing.T) {
	p, _ := NewPool(uniqueIntCreator, noopDestroyer, 0, 2, 1)
	p.Start()
	defer p.Close()

	v1, _ := p.Checkout()
	v2, _ := p.Checkout()

	if p.Current() != 2 {
		t.Errorf("Current() = %d, want 2", p.Current())
	}

	p.Checkin(v1)
	// After checkin, the object should be available for checkout again
	v3, err := p.Checkout()
	if err != nil {
		t.Fatalf("Checkout after Checkin error: %v", err)
	}
	if v3 != v1 {
		t.Errorf("expected to get back %d, got %d", v1, v3)
	}
	p.Checkin(v2)
	p.Checkin(v3)
}

func TestCheckin_UnknownObject_NoOp(t *testing.T) {
	p, _ := NewPool(intCreator, noopDestroyer, 0, 5, 1)
	p.Start()
	defer p.Close()

	// Checking in an object that was never checked out should be a no-op
	p.Checkin(999)
	if p.Current() != 0 {
		t.Errorf("Current() = %d, want 0", p.Current())
	}
}

// --- Delete ---

func TestDelete_RemovesAndDestroysObject(t *testing.T) {
	destroyed := make([]int, 0)
	destroyer := func(v int) error {
		destroyed = append(destroyed, v)
		return nil
	}

	p, _ := NewPool(uniqueIntCreator, destroyer, 0, 5, 1)
	p.Start()
	defer p.Close()

	v, _ := p.Checkout()
	if p.Current() != 1 {
		t.Errorf("Current() = %d, want 1", p.Current())
	}

	p.Delete(v)
	if p.Current() != 0 {
		t.Errorf("Current() after Delete = %d, want 0", p.Current())
	}
	if len(destroyed) != 1 || destroyed[0] != v {
		t.Errorf("destroyer not called correctly, destroyed=%v", destroyed)
	}
}

func TestDelete_UnknownObject_NoOp(t *testing.T) {
	p, _ := NewPool(intCreator, noopDestroyer, 0, 5, 1)
	p.Start()
	defer p.Close()

	p.Delete(999) // should not panic
}

// --- Clear ---

func TestClear_RemovesIdleObjects(t *testing.T) {
	destroyed := make([]int, 0)
	destroyer := func(v int) error {
		destroyed = append(destroyed, v)
		return nil
	}

	p, _ := NewPool(uniqueIntCreator, destroyer, 3, 5, 1)
	p.Start()
	defer p.Close()

	if p.Current() != 3 {
		t.Fatalf("Current() = %d, want 3", p.Current())
	}

	// Checkout one so it's in-use (not cleared)
	v, _ := p.Checkout()

	p.Clear()

	// Only the in-use object should remain
	if p.Current() != 1 {
		t.Errorf("Current() after Clear = %d, want 1", p.Current())
	}
	if len(destroyed) != 2 {
		t.Errorf("expected 2 idle objects destroyed, got %d", len(destroyed))
	}

	p.Checkin(v)
}

// --- Close ---

func TestClose_DestroysAllObjects(t *testing.T) {
	destroyed := make([]int, 0)
	mu := sync.Mutex{}
	destroyer := func(v int) error {
		mu.Lock()
		destroyed = append(destroyed, v)
		mu.Unlock()
		return nil
	}

	p, _ := NewPool(uniqueIntCreator, destroyer, 2, 5, 1)
	p.Start()

	// Checkout one so we have 1 in-use + 1 idle
	_, _ = p.Checkout()

	err := p.Close()
	if err != nil {
		t.Fatalf("Close() error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(destroyed) != 2 {
		t.Errorf("expected 2 objects destroyed, got %d", len(destroyed))
	}
}

// --- HighWaterMark ---

func TestHighWaterMark_TracksPeak(t *testing.T) {
	p, _ := NewPool(uniqueIntCreator, noopDestroyer, 0, 5, 1)
	p.Start()
	defer p.Close()

	v1, _ := p.Checkout()
	v2, _ := p.Checkout()
	v3, _ := p.Checkout()

	if p.HighWaterMark() != 3 {
		t.Errorf("HighWaterMark() = %d, want 3", p.HighWaterMark())
	}

	p.Checkin(v1)
	p.Checkin(v2)
	p.Checkin(v3)

	// HighWaterMark should not decrease
	if p.HighWaterMark() != 3 {
		t.Errorf("HighWaterMark() after checkin = %d, want 3", p.HighWaterMark())
	}
}

// --- Setters ---

func TestSetMin(t *testing.T) {
	p, _ := NewPool(intCreator, noopDestroyer, 1, 10, 1)
	p.Start()
	defer p.Close()

	p.SetMin(5)
	if p.Min() != 5 {
		t.Errorf("Min() = %d, want 5", p.Min())
	}

	// Setting min > max should be ignored
	p.SetMin(20)
	if p.Min() != 5 {
		t.Errorf("Min() = %d, want 5 (should not change)", p.Min())
	}
}

func TestSetMax(t *testing.T) {
	p, _ := NewPool(intCreator, noopDestroyer, 1, 5, 1)
	p.Start()
	defer p.Close()

	p.SetMax(10)
	if p.Max() != 10 {
		t.Errorf("Max() = %d, want 10", p.Max())
	}

	// Setting max < min should be ignored
	p.SetMax(0)
	if p.Max() != 10 {
		t.Errorf("Max() = %d, want 10 (should not change)", p.Max())
	}
}

func TestSetIdleTimeout(t *testing.T) {
	p, _ := NewPool(intCreator, noopDestroyer, 0, 5, 1)
	p.Start()
	defer p.Close()

	p.SetIdleTimeout(30)
	if p.IdleTimeout() != 30 {
		t.Errorf("IdleTimeout() = %d, want 30", p.IdleTimeout())
	}
}

func TestSetMaxWait(t *testing.T) {
	p, _ := NewPool(intCreator, noopDestroyer, 0, 5, 1)
	p.Start()
	defer p.Close()

	p.SetMaxWait(10)
	if p.MaxWait() != 10 {
		t.Errorf("MaxWait() = %d, want 10", p.MaxWait())
	}
}

// --- Concurrent access ---

func TestConcurrentCheckoutCheckin(t *testing.T) {
	p, _ := NewPool(uniqueIntCreator, noopDestroyer, 0, 10, 5)
	p.Start()
	defer p.Close()

	var wg sync.WaitGroup
	errCount := atomic.Int64{}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := p.Checkout()
			if err != nil {
				errCount.Add(1)
				return
			}
			// Simulate some work
			time.Sleep(10 * time.Millisecond)
			p.Checkin(v)
		}()
	}
	wg.Wait()

	if errCount.Load() > 0 {
		t.Logf("Note: %d goroutines got errors (expected if > 10 concurrent)", errCount.Load())
	}

	// All objects should be returned
	current := p.Current()
	if current > 10 {
		t.Errorf("Current() = %d, want <= 10", current)
	}
}

func TestCheckoutWaitsForCheckin(t *testing.T) {
	p, _ := NewPool(uniqueIntCreator, noopDestroyer, 0, 1, 5)
	p.Start()
	defer p.Close()

	v, _ := p.Checkout()

	// Start a goroutine that will checkin after 500ms
	go func() {
		time.Sleep(500 * time.Millisecond)
		p.Checkin(v)
	}()

	start := time.Now()
	v2, err := p.Checkout()
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Checkout() error: %v", err)
	}
	if v2 != v {
		t.Errorf("expected same object back, got %d vs %d", v2, v)
	}
	if elapsed < 400*time.Millisecond {
		t.Errorf("expected to wait ~500ms, waited %v", elapsed)
	}
	p.Checkin(v2)
}
