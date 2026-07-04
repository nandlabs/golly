package messaging

import (
	"testing"
	"time"

	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/golly/testing/assert"
)

// TestAsLifecycleComponent_StartStop verifies that the wrapper's Start
// calls Setup on the underlying provider and Stop calls Close.
func TestAsLifecycleComponent_StartStop(t *testing.T) {
	lp := &LocalProvider{}
	comp := AsLifecycleComponent(lp)

	if comp.Id() == "" {
		t.Error("expected non-empty component Id")
	}
	if comp.Id() != lp.Id() {
		t.Errorf("expected component Id to match provider Id, got %q vs %q", comp.Id(), lp.Id())
	}

	assert.NoError(t, comp.Start())
	if comp.State() != lifecycle.Running {
		t.Errorf("expected Running after Start, got %v", comp.State())
	}

	// Send / receive to verify the provider was actually initialized.
	uri := mustParseURL(t, "chan://lifecycle-basic")
	msg, err := NewLocalMessage()
	assert.NoError(t, err)
	_, _ = msg.SetBodyStr("via-lifecycle")
	assert.NoError(t, lp.Send(uri, msg))
	got, err := lp.Receive(uri)
	assert.NoError(t, err)
	if got.ReadAsStr() != "via-lifecycle" {
		t.Errorf("expected %q, got %q", "via-lifecycle", got.ReadAsStr())
	}

	assert.NoError(t, comp.Stop())
	if comp.State() != lifecycle.Stopped {
		t.Errorf("expected Stopped after Stop, got %v", comp.State())
	}

	// Provider is now closed — Send should surface ErrProviderClosed.
	msg2, _ := NewLocalMessage()
	if err := lp.Send(uri, msg2); err != ErrProviderClosed {
		t.Errorf("expected ErrProviderClosed after Stop, got %v", err)
	}
}

// TestAsLifecycleComponent_DrainWaitsForInFlight verifies the graceful
// shutdown path: an in-flight require-ack listener that has not yet
// returned should hold Stop until drain timeout expires.
func TestAsLifecycleComponent_DrainWaitsForInFlight(t *testing.T) {
	lp := &LocalProvider{}
	comp := AsLifecycleComponentWithTimeout(lp, 200*time.Millisecond)
	assert.NoError(t, comp.Start())

	u := mustParseURL(t, "chan://lifecycle-drain")
	release := make(chan struct{})
	fired := make(chan struct{}, 1)

	assert.NoError(t, lp.AddListener(u, func(m Message) {
		fired <- struct{}{}
		<-release
	}, NewOptionsBuilder().AddRequireAck().AddAckTimeout(2*time.Second).Build()...))

	msg, _ := NewLocalMessage()
	_, _ = msg.SetBodyStr("in-flight")
	assert.NoError(t, lp.Send(u, msg))

	// Wait until the listener has started running.
	select {
	case <-fired:
	case <-time.After(time.Second):
		t.Fatal("listener never fired")
	}

	// Stop should now block for the drain-timeout budget.
	start := time.Now()
	stopDone := make(chan error, 1)
	go func() { stopDone <- comp.Stop() }()

	// Ensure Stop is actually waiting.
	select {
	case <-stopDone:
		t.Fatal("Stop returned before drain timeout while listener was in flight")
	case <-time.After(50 * time.Millisecond):
	}

	// Release the listener; Stop should then complete quickly.
	close(release)
	select {
	case err := <-stopDone:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Stop never returned after listener released")
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Errorf("Stop took too long: %v", elapsed)
	}
}

// TestLocalProvider_DrainReturnsImmediatelyWhenIdle verifies Drain is a
// fast no-op when no dispatches are in flight.
func TestLocalProvider_DrainReturnsImmediatelyWhenIdle(t *testing.T) {
	lp := newTestProvider(t)
	start := time.Now()
	if err := lp.Drain(100 * time.Millisecond); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
		t.Errorf("Drain took too long on idle provider: %v", elapsed)
	}
}
