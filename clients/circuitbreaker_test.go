package clients

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// newTestBreaker returns a breaker with small, deterministic thresholds and a
// 1-second open timeout (the smallest the seconds-based Timeout field allows).
func newTestBreaker() *CircuitBreaker {
	return NewCircuitBreaker(&BreakerInfo{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		MaxHalfOpen:      5,
		Timeout:          1,
	})
}

func TestCircuitBreaker_StartsClosed(t *testing.T) {
	cb := newTestBreaker()
	if err := cb.CanExecute(); err != nil {
		t.Fatalf("fresh breaker should permit execution, got %v", err)
	}
	if got := cb.getState(); got != circuitClosed {
		t.Fatalf("fresh breaker state = %d, want circuitClosed", got)
	}
}

// The core regression: before the getState() fix the breaker never tripped no
// matter how many failures were reported.
func TestCircuitBreaker_OpensAfterFailureThreshold(t *testing.T) {
	cb := newTestBreaker()
	// Two failures — still under the threshold of 3.
	cb.OnExecution(false)
	cb.OnExecution(false)
	if err := cb.CanExecute(); err != nil {
		t.Fatalf("breaker should still be closed after 2/3 failures, got %v", err)
	}
	// Third failure trips it.
	cb.OnExecution(false)
	if err := cb.CanExecute(); !errors.Is(err, ErrCBOpen) {
		t.Fatalf("breaker should be open after 3 failures, CanExecute = %v (want ErrCBOpen)", err)
	}
	if got := cb.getState(); got != circuitOpen {
		t.Fatalf("state = %d, want circuitOpen", got)
	}
}

// Consecutive semantics: an interleaved success resets the failure streak.
func TestCircuitBreaker_SuccessResetsFailureStreak(t *testing.T) {
	cb := newTestBreaker()
	cb.OnExecution(false)
	cb.OnExecution(false) // 2 consecutive failures
	cb.OnExecution(true)  // streak broken → counter reset
	cb.OnExecution(false)
	cb.OnExecution(false) // only 2 consecutive again
	if err := cb.CanExecute(); err != nil {
		t.Fatalf("breaker should still be closed — the success reset the streak; got %v", err)
	}
	cb.OnExecution(false) // now 3 consecutive
	if err := cb.CanExecute(); !errors.Is(err, ErrCBOpen) {
		t.Fatalf("breaker should open on the 3rd consecutive failure, got %v", err)
	}
}

// open → half-open (via the timer) → closed (via successes).
func TestCircuitBreaker_RecoversThroughHalfOpen(t *testing.T) {
	cb := newTestBreaker()
	for i := 0; i < 3; i++ {
		cb.OnExecution(false)
	}
	if err := cb.CanExecute(); !errors.Is(err, ErrCBOpen) {
		t.Fatalf("precondition: breaker should be open, got %v", err)
	}

	// Wait for the open→half-open timer (Timeout = 1s) with headroom.
	waitForState(t, cb, circuitHalfOpen, 3*time.Second)

	// In half-open, probes are permitted.
	if err := cb.CanExecute(); err != nil {
		t.Fatalf("half-open should permit a probe, got %v", err)
	}
	// SuccessThreshold (2) consecutive successes close the circuit.
	cb.OnExecution(true)
	cb.OnExecution(true)
	if got := cb.getState(); got != circuitClosed {
		t.Fatalf("state after 2 successful probes = %d, want circuitClosed", got)
	}
	if err := cb.CanExecute(); err != nil {
		t.Fatalf("closed breaker should permit execution, got %v", err)
	}
}

// A failed probe during half-open snaps the circuit straight back open —
// before the fix this was silently ignored and probing continued.
func TestCircuitBreaker_HalfOpenFailureReopens(t *testing.T) {
	cb := newTestBreaker()
	for i := 0; i < 3; i++ {
		cb.OnExecution(false)
	}
	waitForState(t, cb, circuitHalfOpen, 3*time.Second)

	if err := cb.CanExecute(); err != nil {
		t.Fatalf("half-open should permit the first probe, got %v", err)
	}
	cb.OnExecution(false) // probe fails
	if got := cb.getState(); got != circuitOpen {
		t.Fatalf("a failed probe should reopen the circuit; state = %d, want circuitOpen", got)
	}
	if err := cb.CanExecute(); !errors.Is(err, ErrCBOpen) {
		t.Fatalf("reopened breaker should block, got %v", err)
	}
}

// Half-open throttles concurrent probes to MaxHalfOpen.
func TestCircuitBreaker_HalfOpenThrottlesProbes(t *testing.T) {
	cb := NewCircuitBreaker(&BreakerInfo{
		FailureThreshold: 1,
		SuccessThreshold: 100, // never close during the test
		MaxHalfOpen:      3,
		Timeout:          1,
	})
	cb.OnExecution(false) // threshold 1 → open immediately
	waitForState(t, cb, circuitHalfOpen, 3*time.Second)

	permitted := 0
	for i := 0; i < 10; i++ {
		if cb.CanExecute() == nil {
			permitted++
		}
	}
	if permitted != int(cb.MaxHalfOpen) {
		t.Fatalf("half-open permitted %d probes, want MaxHalfOpen=%d", permitted, cb.MaxHalfOpen)
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := newTestBreaker()
	for i := 0; i < 3; i++ {
		cb.OnExecution(false)
	}
	if err := cb.CanExecute(); !errors.Is(err, ErrCBOpen) {
		t.Fatalf("precondition: breaker should be open, got %v", err)
	}
	cb.Reset()
	if got := cb.getState(); got != circuitClosed {
		t.Fatalf("after Reset state = %d, want circuitClosed", got)
	}
	if err := cb.CanExecute(); err != nil {
		t.Fatalf("after Reset breaker should permit execution, got %v", err)
	}
}

func TestCircuitBreaker_Defaults(t *testing.T) {
	cb := NewCircuitBreaker(nil)
	if cb.FailureThreshold != defaultFailureThreshold ||
		cb.SuccessThreshold != defaultSuccessThreshold ||
		cb.MaxHalfOpen != defaultMaxHalfOpen ||
		cb.Timeout != defaultTimeout {
		t.Fatalf("defaults not applied: %+v", cb.BreakerInfo)
	}
	if cb.getState() != circuitClosed {
		t.Fatal("default breaker should start closed")
	}
}

// Race coverage: concurrent OnExecution / CanExecute must not race and must
// still trip the breaker deterministically. Run under -race.
func TestCircuitBreaker_ConcurrentReports(t *testing.T) {
	cb := NewCircuitBreaker(&BreakerInfo{
		FailureThreshold: 50,
		SuccessThreshold: 2,
		MaxHalfOpen:      5,
		Timeout:          300,
	})
	var wg sync.WaitGroup
	var failures int64
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cb.OnExecution(false)
			atomic.AddInt64(&failures, 1)
			_ = cb.CanExecute()
		}()
	}
	wg.Wait()
	// 200 concurrent failures vastly exceed the threshold of 50 → open.
	if err := cb.CanExecute(); !errors.Is(err, ErrCBOpen) {
		t.Fatalf("breaker should be open after 200 concurrent failures, got %v", err)
	}
}

// waitForState polls the breaker until it reaches want or the deadline elapses.
func waitForState(t *testing.T, cb *CircuitBreaker, want uint32, within time.Duration) {
	t.Helper()
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if cb.getState() == want {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("breaker did not reach state %d within %s (still %d)", want, within, cb.getState())
}
