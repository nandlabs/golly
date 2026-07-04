package messaging

import (
	"sync/atomic"
	"testing"
	"time"

	"oss.nandlabs.io/golly/testing/assert"
)

// TestRequireAck_DispatcherWaitsForAck verifies that with RequireAckOpt
// set the dispatcher does not deliver the next message until the current
// one is acknowledged.
func TestRequireAck_DispatcherWaitsForAck(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://require-ack-wait")

	release := make(chan struct{})
	var order atomic.Int32
	firstStarted := make(chan struct{}, 1)
	secondStarted := make(chan struct{}, 1)

	opts := NewOptionsBuilder().AddRequireAck().AddAckTimeout(5 * time.Second).Build()
	assert.NoError(t, lp.AddListener(u, func(m Message) {
		n := order.Add(1)
		if n == 1 {
			firstStarted <- struct{}{}
			<-release
		} else {
			secondStarted <- struct{}{}
		}
		_ = m.Rsvp(true)
	}, opts...))

	for i := 0; i < 2; i++ {
		msg, _ := NewLocalMessage()
		_, _ = msg.SetBodyStr("ack-me")
		assert.NoError(t, lp.Send(u, msg))
	}

	// The first listener must start before the second.
	<-firstStarted
	select {
	case <-secondStarted:
		t.Fatal("second listener started before first ack — dispatcher did not wait")
	case <-time.After(80 * time.Millisecond):
	}
	close(release)
	<-secondStarted
}

// TestRequireAck_AckTimeoutTriggersRequeue verifies that when a listener
// fails to Rsvp within AckTimeout the dispatcher treats it as a
// retryable failure and redelivers.
func TestRequireAck_AckTimeoutTriggersRequeue(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://ack-timeout")

	var seen atomic.Int32
	done := make(chan struct{}, 1)

	opts := NewOptionsBuilder().
		AddRequireAck().
		AddAckTimeout(50 * time.Millisecond).
		AddMaxDeliveryAttempts(3).
		Build()

	assert.NoError(t, lp.AddListener(u, func(m Message) {
		n := seen.Add(1)
		if n < 3 {
			// Don't Rsvp — force the timeout path.
			return
		}
		_ = m.Rsvp(true)
		select {
		case done <- struct{}{}:
		default:
		}
	}, opts...))

	msg, _ := NewLocalMessage()
	_, _ = msg.SetBodyStr("timeout-retry")
	assert.NoError(t, lp.Send(u, msg))

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("message never acked after retries; seen=%d", seen.Load())
	}
	if seen.Load() != 3 {
		t.Errorf("expected 3 deliveries, got %d", seen.Load())
	}
}

// TestRequireAck_MaxAttemptsRoutesToDLQ verifies that after
// MaxDeliveryAttempts exhausts, the message routes to the configured
// DeadLetter destination.
func TestRequireAck_MaxAttemptsRoutesToDLQ(t *testing.T) {
	// Register a DLQ listener on the default manager (single provider covers
	// both source and DLQ URLs since they share the chan scheme).
	dlqURL := "chan://ack-dlq-target"
	dlqArrived := make(chan Message, 1)
	assert.NoError(t, GetManager().AddListener(mustParseURL(t, dlqURL), func(m Message) {
		select {
		case dlqArrived <- m:
		default:
		}
	}))
	t.Cleanup(func() {
		_ = GetManager().RemoveListeners(mustParseURL(t, dlqURL))
	})

	u := mustParseURL(t, "chan://ack-dlq-source")
	opts := NewOptionsBuilder().
		AddRequireAck().
		AddAckTimeout(30 * time.Millisecond).
		AddMaxDeliveryAttempts(2).
		AddDeadLetter(dlqURL).
		Build()

	assert.NoError(t, GetManager().AddListener(u, func(m Message) {
		// Never ack — always Nack with requeue.
		_ = m.Rsvp(false, NewOptionsBuilder().AddRequeue().Build()...)
	}, opts...))
	t.Cleanup(func() { _ = GetManager().RemoveListeners(u) })

	msg, _ := NewLocalMessage()
	_, _ = msg.SetBodyStr("route-to-dlq")
	assert.NoError(t, GetManager().Send(u, msg))

	select {
	case m := <-dlqArrived:
		if m.ReadAsStr() != "route-to-dlq" {
			t.Errorf("expected DLQ body %q, got %q", "route-to-dlq", m.ReadAsStr())
		}
	case <-time.After(3 * time.Second):
		t.Fatal("message never reached DLQ")
	}
}

// TestRequireAck_DeliveryCounter verifies deliveryAttempts increments
// exactly once per dispatch attempt.
func TestRequireAck_DeliveryCounter(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://ack-counter")

	var lastAttempts atomic.Int32
	done := make(chan struct{}, 1)

	opts := NewOptionsBuilder().
		AddRequireAck().
		AddAckTimeout(30 * time.Millisecond).
		AddMaxDeliveryAttempts(4).
		Build()

	assert.NoError(t, lp.AddListener(u, func(m Message) {
		if lm, ok := m.(*LocalMessage); ok {
			lastAttempts.Store(lm.deliveryAttempts())
		}
		if lastAttempts.Load() < 3 {
			return // no Rsvp — timeout will requeue
		}
		_ = m.Rsvp(true)
		select {
		case done <- struct{}{}:
		default:
		}
	}, opts...))

	msg, _ := NewLocalMessage()
	assert.NoError(t, lp.Send(u, msg))
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("message never succeeded")
	}
	if lastAttempts.Load() != 3 {
		t.Errorf("expected final attempts counter 3, got %d", lastAttempts.Load())
	}
}

// TestRequireAck_NackWithoutRequeueDrops verifies that a Nack without
// RequeueOpt (and no DLQ) drops the message immediately without any
// redelivery attempt.
func TestRequireAck_NackWithoutRequeueDrops(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://nack-no-requeue")

	var seen atomic.Int32
	opts := NewOptionsBuilder().
		AddRequireAck().
		AddMaxDeliveryAttempts(5).
		Build()

	assert.NoError(t, lp.AddListener(u, func(m Message) {
		seen.Add(1)
		_ = m.Rsvp(false) // no Requeue option
	}, opts...))

	msg, _ := NewLocalMessage()
	assert.NoError(t, lp.Send(u, msg))
	time.Sleep(150 * time.Millisecond)
	if seen.Load() != 1 {
		t.Errorf("expected 1 delivery, got %d", seen.Load())
	}
}
