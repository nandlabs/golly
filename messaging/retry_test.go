package messaging

import (
	"sync/atomic"
	"testing"
	"time"

	"oss.nandlabs.io/golly/testing/assert"
)

// TestWithRetry_HappyPath verifies the wrapper calls handler exactly
// once when no panic occurs.
func TestWithRetry_HappyPath(t *testing.T) {
	var called atomic.Int32
	h := WithRetry(func(Message) {
		called.Add(1)
	}, DefaultRetryPolicy(), nil)

	msg, _ := NewLocalMessage()
	h(msg)
	if called.Load() != 1 {
		t.Errorf("expected 1 call, got %d", called.Load())
	}
}

// TestWithRetry_RetriesOnPanic verifies the wrapper retries after a
// panic and eventually succeeds when the handler stabilizes.
func TestWithRetry_RetriesOnPanic(t *testing.T) {
	var attempts atomic.Int32
	policy := RetryPolicy{MaxRetries: 3, Wait: 1}
	h := WithRetry(func(Message) {
		if attempts.Add(1) < 3 {
			panic("transient")
		}
	}, policy, nil)

	msg, _ := NewLocalMessage()
	h(msg) // must not panic out of the wrapper
	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts.Load())
	}
}

// TestWithRetry_RoutesToDLQOnExhaustion verifies that when retries are
// exhausted and a dlqURL is configured, the message is sent to the DLQ
// via the package-level Manager.
func TestWithRetry_RoutesToDLQOnExhaustion(t *testing.T) {
	// Fresh DLQ URL — use the default manager's LocalProvider.
	dlq := mustParseURL(t, "chan://retry-dlq")
	// Attach a listener so we can observe the DLQ arrival.
	arrived := make(chan struct{}, 1)
	assert.NoError(t, GetManager().AddListener(dlq, func(Message) {
		select {
		case arrived <- struct{}{}:
		default:
		}
	}))

	policy := RetryPolicy{MaxRetries: 1, Wait: 1}
	var attempts atomic.Int32
	h := WithRetry(func(Message) {
		attempts.Add(1)
		panic("always fails")
	}, policy, dlq)

	msg, _ := NewLocalMessage()
	_, _ = msg.SetBodyStr("dlq-me")
	h(msg)

	if attempts.Load() != 2 {
		t.Errorf("expected 2 attempts (initial + 1 retry), got %d", attempts.Load())
	}
	select {
	case <-arrived:
	case <-time.After(2 * time.Second):
		t.Fatal("DLQ never received the exhausted message")
	}
	_ = GetManager().RemoveListeners(dlq)
}

// TestWithRetry_DropsWhenNoDLQConfigured verifies the wrapper doesn't
// panic when retries exhaust without a DLQ — it just logs and returns.
func TestWithRetry_DropsWhenNoDLQConfigured(t *testing.T) {
	var attempts atomic.Int32
	policy := RetryPolicy{MaxRetries: 1, Wait: 1}
	h := WithRetry(func(Message) {
		attempts.Add(1)
		panic("no dlq")
	}, policy, nil)

	msg, _ := NewLocalMessage()
	h(msg) // must return cleanly
	if attempts.Load() != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts.Load())
	}
}

// TestWithRetry_NilHandlerIsSafe verifies WithRetry doesn't panic when
// handed a nil handler.
func TestWithRetry_NilHandlerIsSafe(t *testing.T) {
	h := WithRetry(nil, DefaultRetryPolicy(), nil)
	if h == nil {
		t.Fatal("expected non-nil handler wrapper")
	}
	msg, _ := NewLocalMessage()
	h(msg) // no panic
}

// TestRetrySignal_ReturnsErrorValue verifies the RetrySignal helper
// produces a useful error value.
func TestRetrySignal_ReturnsErrorValue(t *testing.T) {
	err := RetrySignal("upstream 5xx")
	if err == nil {
		t.Fatal("expected non-nil retry signal")
	}
	if err.Error() != "retry: upstream 5xx" {
		t.Errorf("unexpected retry signal message: %q", err.Error())
	}
}
