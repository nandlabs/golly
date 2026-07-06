package genai

import (
	"context"
	"errors"
	"net"
	"strings"
)

// ErrNoRoute is returned when no strategy can produce a viable route for a task
// (e.g. no model satisfies the required capabilities and budget).
var ErrNoRoute = errors.New("genai: no route for task")

// RetryableError is an optional interface a provider error may implement to
// declare, authoritatively, whether it should trigger fallover to a different
// model. When present it wins over heuristic classification.
type RetryableError interface {
	error
	Retryable() bool
}

// retryable reports whether an error (with an optional partial GenResponse)
// should trigger fallover to a different model. Transient and capacity errors
// (429, 5xx, network, model-unavailable) are retryable; client and permanent
// errors (400, invalid schema, content policy, auth) are not — a different
// model would fail identically, wasting latency and money.
func retryable(err error, resp *GenResponse) bool {
	if err == nil {
		return false
	}

	// Authoritative provider signal wins.
	var re RetryableError
	if errors.As(err, &re) {
		return re.Retryable()
	}

	// A content-filter finish reason is never retryable: a different model will
	// not un-block the content.
	if resp != nil {
		for _, c := range resp.Candidates {
			if c.FinishReason == FinishReasonContentFilter {
				return false
			}
		}
	}

	// Context cancellation / deadline is a caller-driven stop, not a per-model
	// transient — do not fall over, the next model would hit the same deadline.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Network-level errors are transient.
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// Heuristic string classification for providers that surface status codes
	// in error text (the in-tree impl providers do exactly this).
	msg := strings.ToLower(err.Error())
	for _, permanent := range permanentMarkers {
		if strings.Contains(msg, permanent) {
			return false
		}
	}
	for _, transient := range transientMarkers {
		if strings.Contains(msg, transient) {
			return true
		}
	}

	// Unknown errors are treated as permanent: better to surface an unexpected
	// failure than to silently burn the fallback budget on it.
	return false
}

// permanentMarkers are substrings that indicate a non-retryable client error.
// Checked before transientMarkers so an explicit 4xx wins.
var permanentMarkers = []string{
	"status 400", "status 401", "status 403", "status 404", "status 422",
	"content_filter", "content policy", "invalid schema", "invalid_request",
	"unauthorized", "forbidden", "bad request", "permission",
}

// transientMarkers are substrings that indicate a retryable transient error.
var transientMarkers = []string{
	"status 429", "status 500", "status 502", "status 503", "status 504",
	"too many requests", "rate limit", "timeout", "timed out", "unavailable",
	"overloaded", "capacity", "connection reset", "temporarily",
}
