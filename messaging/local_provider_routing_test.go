package messaging

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"oss.nandlabs.io/golly/testing/assert"
)

// TestRouting_KeyPicksSameNamedGroup verifies that identical routing
// keys deterministically route to the same named group and never
// straddle groups.
func TestRouting_KeyPicksSameNamedGroup(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://routing-key-groups")

	var toGroupA, toGroupB atomic.Int32
	var wg sync.WaitGroup
	const total = 20
	wg.Add(total)

	assert.NoError(t, lp.AddListener(u, func(Message) {
		toGroupA.Add(1)
		wg.Done()
	}, NewOptionsBuilder().AddNamedListener("group-a").Build()...))
	assert.NoError(t, lp.AddListener(u, func(Message) {
		toGroupB.Add(1)
		wg.Done()
	}, NewOptionsBuilder().AddNamedListener("group-b").Build()...))

	sendOpts := NewOptionsBuilder().AddRoutingKey("stable-key").Build()
	for i := 0; i < total; i++ {
		msg, _ := NewLocalMessage()
		assert.NoError(t, lp.Send(u, msg, sendOpts...))
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out; a=%d b=%d", toGroupA.Load(), toGroupB.Load())
	}
	// Exactly one group should have received all messages.
	if toGroupA.Load() != 0 && toGroupB.Load() != 0 {
		t.Errorf("keyed dispatch straddled groups: a=%d b=%d", toGroupA.Load(), toGroupB.Load())
	}
	if toGroupA.Load()+toGroupB.Load() != int32(total) {
		t.Errorf("expected total=%d, got a+b=%d", total, toGroupA.Load()+toGroupB.Load())
	}
}

// TestRouting_UnnamedListenersStillBroadcast verifies unnamed listeners
// keep their broadcast semantics even when a routing key is present.
func TestRouting_UnnamedListenersStillBroadcast(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://routing-key-broadcast")

	var unnamedA, unnamedB atomic.Int32
	var wg sync.WaitGroup
	wg.Add(2)
	assert.NoError(t, lp.AddListener(u, func(Message) { unnamedA.Add(1); wg.Done() }))
	assert.NoError(t, lp.AddListener(u, func(Message) { unnamedB.Add(1); wg.Done() }))

	msg, _ := NewLocalMessage()
	if k, ok := msg.(Keyed); ok {
		k.SetRoutingKey("does-not-matter")
	}
	assert.NoError(t, lp.Send(u, msg))

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("unnamed listeners did not both fire: a=%d b=%d", unnamedA.Load(), unnamedB.Load())
	}
	if unnamedA.Load() != 1 || unnamedB.Load() != 1 {
		t.Errorf("expected both unnamed listeners to fire once; a=%d b=%d",
			unnamedA.Load(), unnamedB.Load())
	}
}

// TestRouting_NoKeyRetainsPerGroupBehavior verifies that when no routing
// key is set, every named group fires one listener (the pre-v1.7 fan-out
// behavior).
func TestRouting_NoKeyRetainsPerGroupBehavior(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://routing-no-key")

	var groupA, groupB atomic.Int32
	var wg sync.WaitGroup
	wg.Add(2) // both groups should fire once each
	assert.NoError(t, lp.AddListener(u, func(Message) { groupA.Add(1); wg.Done() },
		NewOptionsBuilder().AddNamedListener("a").Build()...))
	assert.NoError(t, lp.AddListener(u, func(Message) { groupB.Add(1); wg.Done() },
		NewOptionsBuilder().AddNamedListener("b").Build()...))

	msg, _ := NewLocalMessage()
	assert.NoError(t, lp.Send(u, msg))

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("named groups did not both fire: a=%d b=%d", groupA.Load(), groupB.Load())
	}
	if groupA.Load() != 1 || groupB.Load() != 1 {
		t.Errorf("expected each named group to fire once; a=%d b=%d",
			groupA.Load(), groupB.Load())
	}
}

// TestRouting_KeyDeterministicWithinGroup verifies that when a single
// named group has multiple listeners, an identical routing key
// deterministically selects the same listener index every time.
func TestRouting_KeyDeterministicWithinGroup(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://routing-within-group")

	const listeners = 4
	var counts [listeners]atomic.Int32
	var wg sync.WaitGroup
	const total = 12
	wg.Add(total)

	for i := 0; i < listeners; i++ {
		idx := i
		assert.NoError(t, lp.AddListener(u, func(Message) {
			counts[idx].Add(1)
			wg.Done()
		}, NewOptionsBuilder().AddNamedListener("workers").Build()...))
	}

	sendOpts := NewOptionsBuilder().AddRoutingKey("customer-42").Build()
	for i := 0; i < total; i++ {
		msg, _ := NewLocalMessage()
		assert.NoError(t, lp.Send(u, msg, sendOpts...))
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}
	nonZero := 0
	for i := 0; i < listeners; i++ {
		if counts[i].Load() > 0 {
			nonZero++
		}
	}
	if nonZero != 1 {
		t.Errorf("expected exactly one listener to fire (sticky by key); got %d non-zero counts", nonZero)
	}
}
