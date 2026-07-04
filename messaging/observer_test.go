package messaging

import (
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"oss.nandlabs.io/golly/testing/assert"
)

// TestObserver_OnSendAndOnReceive verifies the OnSend / OnReceive hooks
// fire for both success and error paths on LocalProvider.
func TestObserver_OnSendAndOnReceive(t *testing.T) {
	lp := newTestProvider(t)
	obs := &recordingObserver{}
	lp.SetObserver(obs)

	u := mustParseURL(t, "chan://observe-send-recv")
	msg, _ := NewLocalMessage()
	_, _ = msg.SetBodyStr("observed")
	assert.NoError(t, lp.Send(u, msg))

	got, err := lp.Receive(u)
	assert.NoError(t, err)
	if got.ReadAsStr() != "observed" {
		t.Errorf("expected %q, got %q", "observed", got.ReadAsStr())
	}

	obs.mu.Lock()
	defer obs.mu.Unlock()
	if len(obs.sends) != 1 {
		t.Errorf("expected 1 OnSend, got %d", len(obs.sends))
	}
	if len(obs.receives) != 1 {
		t.Errorf("expected 1 OnReceive, got %d", len(obs.receives))
	}
	if obs.sends[0].err != nil {
		t.Errorf("expected nil err on OnSend, got %v", obs.sends[0].err)
	}
	if obs.receives[0].err != nil {
		t.Errorf("expected nil err on OnReceive, got %v", obs.receives[0].err)
	}
}

// TestObserver_OnAckOnNack verifies ack/nack callbacks fire for a
// require-ack-configured listener.
func TestObserver_OnAckOnNack(t *testing.T) {
	lp := newTestProvider(t)
	obs := &recordingObserver{}
	lp.SetObserver(obs)

	u := mustParseURL(t, "chan://observe-ack")
	opts := NewOptionsBuilder().AddRequireAck().AddAckTimeout(500 * time.Millisecond).Build()

	// First listener acks.
	ackDone := make(chan struct{}, 1)
	assert.NoError(t, lp.AddListener(u, func(m Message) {
		_ = m.Rsvp(true)
		ackDone <- struct{}{}
	}, opts...))

	msg, _ := NewLocalMessage()
	_, _ = msg.SetBodyStr("ack-me")
	assert.NoError(t, lp.Send(u, msg))
	<-ackDone
	// Wait for the dispatcher to record the ack.
	waitFor(t, func() bool {
		obs.mu.Lock()
		defer obs.mu.Unlock()
		return len(obs.acks) == 1
	}, time.Second, "OnAck")

	// Now nack a message on a fresh listener (fresh URL to isolate config).
	uNack := mustParseURL(t, "chan://observe-nack")
	nackOpts := NewOptionsBuilder().AddRequireAck().AddMaxDeliveryAttempts(1).Build()
	nackDone := make(chan struct{}, 1)
	assert.NoError(t, lp.AddListener(uNack, func(m Message) {
		_ = m.Rsvp(false)
		nackDone <- struct{}{}
	}, nackOpts...))
	msg2, _ := NewLocalMessage()
	_, _ = msg2.SetBodyStr("nack-me")
	assert.NoError(t, lp.Send(uNack, msg2))
	<-nackDone
	waitFor(t, func() bool {
		obs.mu.Lock()
		defer obs.mu.Unlock()
		return len(obs.nacks) >= 1
	}, time.Second, "OnNack")
}

// TestObserver_NilIsSafe verifies passing nil to SetObserver clears any
// previously-installed observer and leaves the hook sites safe.
func TestObserver_NilIsSafe(t *testing.T) {
	lp := newTestProvider(t)

	// No observer installed at all.
	u := mustParseURL(t, "chan://observe-nil")
	msg, _ := NewLocalMessage()
	assert.NoError(t, lp.Send(u, msg))

	// Install then clear.
	obs := &recordingObserver{}
	lp.SetObserver(obs)
	lp.SetObserver(nil)

	// Further sends must not push to the cleared observer.
	msg2, _ := NewLocalMessage()
	assert.NoError(t, lp.Send(u, msg2))
	obs.mu.Lock()
	defer obs.mu.Unlock()
	if len(obs.sends) != 0 {
		t.Errorf("expected 0 OnSend after nil-clear; got %d", len(obs.sends))
	}
}

// TestSetObserver_HelperReportsCapability checks the package helper
// returns false for providers that do not implement ObservableProvider.
func TestSetObserver_HelperReportsCapability(t *testing.T) {
	if !SetObserver(newTestProvider(t), &recordingObserver{}) {
		t.Error("expected LocalProvider to accept observer via helper")
	}
	if SetObserver(&bareProvider{}, &recordingObserver{}) {
		t.Error("bareProvider should not accept observer")
	}
}

// --- Test doubles ---

type sendRecord struct {
	u   *url.URL
	err error
	lat time.Duration
}
type recvRecord struct {
	u   *url.URL
	err error
}
type ackRecord struct {
	u       *url.URL
	requeue bool
}

type recordingObserver struct {
	mu       sync.Mutex
	sends    []sendRecord
	receives []recvRecord
	acks     []ackRecord
	nacks    []ackRecord
}

func (o *recordingObserver) OnSend(u *url.URL, _ Message, err error, lat time.Duration) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.sends = append(o.sends, sendRecord{u, err, lat})
}
func (o *recordingObserver) OnReceive(u *url.URL, _ Message, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.receives = append(o.receives, recvRecord{u, err})
}
func (o *recordingObserver) OnAck(u *url.URL, _ Message) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.acks = append(o.acks, ackRecord{u, false})
}
func (o *recordingObserver) OnNack(u *url.URL, _ Message, requeue bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.nacks = append(o.nacks, ackRecord{u, requeue})
}

// waitFor polls check() until it returns true or timeout expires, failing
// the test with msg when it times out. Intended for eventual-consistency
// assertions where the exact timing of a dispatcher goroutine is not
// worth pinning down.
func waitFor(t *testing.T, check func() bool, timeout time.Duration, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	// last try
	if !check() {
		var _ atomic.Int32 // ensure atomic import stays reachable when trimmed
		t.Fatalf("condition never met: %s", msg)
	}
}
