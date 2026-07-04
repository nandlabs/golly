package messaging

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"oss.nandlabs.io/golly/testing/assert"
)

// TestConsumerConcurrency_CapsInFlight verifies that with
// ConsumerConcurrencyOpt=N no more than N listener invocations are
// active at the same time.
func TestConsumerConcurrency_CapsInFlight(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://concurrency-cap")

	const cap = 3
	const total = 20

	var inFlight atomic.Int32
	var maxInFlight atomic.Int32
	var wg sync.WaitGroup
	wg.Add(total)

	opts := NewOptionsBuilder().AddConsumerConcurrency(cap).Build()
	assert.NoError(t, lp.AddListener(u, func(Message) {
		cur := inFlight.Add(1)
		// Track max concurrently observed.
		for {
			m := maxInFlight.Load()
			if cur <= m || maxInFlight.CompareAndSwap(m, cur) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		inFlight.Add(-1)
		wg.Done()
	}, opts...))

	for i := 0; i < total; i++ {
		msg, _ := NewLocalMessage()
		assert.NoError(t, lp.Send(u, msg))
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("listeners did not drain within timeout")
	}
	if maxInFlight.Load() > int32(cap) {
		t.Errorf("expected max in-flight <= %d, got %d", cap, maxInFlight.Load())
	}
	if maxInFlight.Load() == 0 {
		t.Errorf("expected some in-flight overlap; got 0")
	}
}

// TestConsumerConcurrency_UnsetIsUnbounded verifies the default
// (no-option) behavior is not accidentally bounded — many in-flight
// goroutines are permitted.
func TestConsumerConcurrency_UnsetIsUnbounded(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://concurrency-unbounded")

	const total = 10
	var maxInFlight atomic.Int32
	var inFlight atomic.Int32
	var wg sync.WaitGroup
	wg.Add(total)

	assert.NoError(t, lp.AddListener(u, func(Message) {
		cur := inFlight.Add(1)
		for {
			m := maxInFlight.Load()
			if cur <= m || maxInFlight.CompareAndSwap(m, cur) {
				break
			}
		}
		time.Sleep(30 * time.Millisecond)
		inFlight.Add(-1)
		wg.Done()
	}))

	for i := 0; i < total; i++ {
		msg, _ := NewLocalMessage()
		assert.NoError(t, lp.Send(u, msg))
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out")
	}
	// Without a cap we should see substantial overlap.
	if maxInFlight.Load() < 3 {
		t.Errorf("expected substantial overlap without cap; got %d", maxInFlight.Load())
	}
}
