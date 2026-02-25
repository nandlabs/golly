package clients

import (
	"math"
	"math/rand/v2"
	"time"
)

// RetryInfo represents the retry configuration for a client.
// It supports fixed, exponential, and jittered backoff strategies.
type RetryInfo struct {
	MaxRetries int // Maximum number of retries allowed.
	Wait       int // Base wait time in milliseconds between retries.

	// Exponential enables exponential backoff when true.
	// The wait time is multiplied by Multiplier^retryCount on each attempt.
	Exponential bool

	// Multiplier is the factor by which the wait time is multiplied on each
	// successive retry. Defaults to 2 if set to <= 0 when Exponential is true.
	Multiplier float64

	// MaxWait is the upper bound for the wait time in milliseconds.
	// When set to > 0, the computed backoff will never exceed this value.
	MaxWait int

	// Jitter adds randomized jitter to the wait time when true.
	// A random duration between 0 and the computed backoff is added,
	// which helps prevent thundering-herd problems.
	Jitter bool
}

// WaitTime calculates the wait duration for the given retry attempt (0-indexed).
//
// With Exponential=false, it returns a fixed duration of Wait milliseconds.
//
// With Exponential=true, the backoff is computed as:
//
//	backoff = Wait * Multiplier^retryCount
//
// The result is capped at MaxWait (if > 0) and optionally jittered.
func (r *RetryInfo) WaitTime(retryCount int) time.Duration {
	backoff := time.Duration(r.Wait) * time.Millisecond

	if r.Exponential {
		multiplier := r.Multiplier
		if multiplier <= 0 {
			multiplier = 2
		}
		factor := math.Pow(multiplier, float64(retryCount))
		backoff = time.Duration(float64(backoff) * factor)

		if r.MaxWait > 0 {
			maxBackoff := time.Duration(r.MaxWait) * time.Millisecond
			backoff = min(backoff, maxBackoff)
		}
	}

	if r.Jitter && backoff > 0 {
		jitter := time.Duration(rand.Int64N(int64(backoff)))
		backoff += jitter
	}

	return backoff
}
