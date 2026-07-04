package messaging

import (
	"fmt"
	"net/url"
	"time"

	"oss.nandlabs.io/golly/clients"
)

// RetryPolicy is the retry configuration honored by WithRetry. It is a
// re-export of clients.RetryInfo so callers building messaging pipelines
// don't need to import clients directly; the semantics (fixed /
// exponential / jittered backoff, MaxWait cap) are identical.
type RetryPolicy = clients.RetryInfo

// DefaultRetryPolicy returns a reasonable default policy — 3 attempts,
// 100ms base wait, exponential backoff with jitter, capped at 2s.
// Callers who want different semantics should construct their own
// RetryPolicy directly.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries:  3,
		Wait:        100,
		Exponential: true,
		Multiplier:  2,
		MaxWait:     2000,
		Jitter:      true,
	}
}

// WithRetry wraps a handler with retry-on-panic and dead-letter routing.
// The returned handler:
//
//  1. Invokes handler(msg). If handler runs to completion, the wrapper
//     returns and the message is considered delivered.
//
//  2. If handler panics, the wrapper recovers, waits per policy, and
//     retries — up to policy.MaxRetries attempts in total.
//
//  3. When all retries are exhausted, the wrapper sends msg to dlqURL
//     via the package-level Manager. If the DLQ send itself fails, the
//     failure is logged at ERROR — the wrapper never re-panics.
//
// dlqURL may be nil: retries still run per policy, but exhaustion drops
// the message (logged at WARN). This matches the "no DLQ configured"
// path documented for broker satellites.
//
// The wrapper is deliberately opinionated: it does not consult
// Message.Rsvp because handlers of type func(Message) have no
// standardized error-return channel. Handlers that need to signal
// "please retry me" should panic (e.g. via a custom retry sentinel) or
// use the ack-aware flow (RequireAckOpt) which the LocalProvider
// dispatcher redelivers on Nack + Requeue.
func WithRetry(handler func(Message), policy RetryPolicy, dlqURL *url.URL) func(Message) {
	if handler == nil {
		return func(Message) {}
	}
	// Normalize the retry cap so a zero-value policy still tries once.
	maxAttempts := policy.MaxRetries
	if maxAttempts < 0 {
		maxAttempts = 0
	}
	return func(msg Message) {
		var lastPanic any
		// The first attempt is index 0; MaxRetries additional retries follow.
		for attempt := 0; attempt <= maxAttempts; attempt++ {
			if attempt > 0 {
				wait := policy.WaitTime(attempt - 1)
				if wait > 0 {
					time.Sleep(wait)
				}
			}
			done := runHandlerSafely(handler, msg)
			if done == nil {
				return // success
			}
			lastPanic = done
			logger.WarnF("messaging: handler panic on attempt %d/%d: %v",
				attempt+1, maxAttempts+1, done)
		}
		// Retries exhausted.
		if dlqURL != nil {
			if err := GetManager().Send(dlqURL, msg); err != nil {
				logger.ErrorF("messaging: dead-letter send to %s failed after %d attempts (last panic: %v): %v",
					dlqURL.String(), maxAttempts+1, lastPanic, err)
				return
			}
			logger.InfoF("messaging: routed message to dead-letter %s after %d attempts (last panic: %v)",
				dlqURL.String(), maxAttempts+1, lastPanic)
			return
		}
		logger.WarnF("messaging: dropping message after %d attempts (last panic: %v); no dead-letter configured",
			maxAttempts+1, lastPanic)
	}
}

// runHandlerSafely invokes handler(msg) and returns any panic value it
// raised, or nil on clean completion. The recovered value is returned
// so the caller can decide whether to retry or route to DLQ.
func runHandlerSafely(handler func(Message), msg Message) (recovered any) {
	defer func() {
		if r := recover(); r != nil {
			recovered = r
		}
	}()
	handler(msg)
	return nil
}

// retrySentinel is a small helper type callers can panic with to
// distinguish "please retry me" from unexpected failures. It is optional
// — WithRetry treats every panic the same.
type retrySentinel struct{ reason string }

func (r retrySentinel) Error() string { return fmt.Sprintf("retry: %s", r.reason) }

// RetrySignal returns an error value that, when passed to panic() from
// inside a handler, will be recovered by WithRetry and treated as a
// retry request. This lets handlers cleanly express "retry me" without
// panicking with an ambiguous value.
func RetrySignal(reason string) error { return retrySentinel{reason: reason} }
