package messaging

import (
	"net/url"
	"time"
)

// Observer is a metrics / tracing hook interface. Providers that
// implement ObservableProvider fire an Observer's callbacks around Send,
// Receive, Ack and Nack so callers can attach Prometheus counters,
// OpenTelemetry spans, log tapping, etc. without wrapping every provider
// method.
//
// All callbacks must be non-blocking and safe to invoke from arbitrary
// goroutines. Providers guard nil Observer values internally, so callers
// may pass nil to disable observation without a nil check.
type Observer interface {
	// OnSend is called after a Send attempt, whether it succeeded or
	// failed. err is nil on success. latency measures the elapsed time
	// from Send entry to return.
	OnSend(u *url.URL, msg Message, err error, latency time.Duration)

	// OnReceive is called after a Receive (or dispatched listener) call
	// yields a message. err is non-nil when the receive itself failed
	// (e.g. provider closed) and msg is nil in that case.
	OnReceive(u *url.URL, msg Message, err error)

	// OnAck is called when a listener acknowledges a message via
	// Message.Rsvp(true, ...). Only providers that track acks (e.g.
	// LocalProvider with RequireAckOpt) fire this callback.
	OnAck(u *url.URL, msg Message)

	// OnNack is called when a listener negatively acknowledges a message
	// via Message.Rsvp(false, ...). requeue reflects the RequeueOpt
	// value on the Rsvp call. Only providers that track acks fire this
	// callback.
	OnNack(u *url.URL, msg Message, requeue bool)
}

// ObservableProvider is an optional capability a Provider may implement
// to accept an Observer for metrics / tracing hooks. Callers should
// prefer this over wrapping every Producer / Receiver method externally.
//
// The zero-value Observer contract: passing nil clears any previously
// registered observer and disables the hooks.
type ObservableProvider interface {
	Provider
	// SetObserver installs (or clears, with nil) the observer. It is
	// safe to call at any point in the provider's lifecycle — providers
	// snapshot the current observer at each hook site.
	SetObserver(Observer)
}

// SetObserver is a convenience wrapper: it type-asserts p to
// ObservableProvider and installs obs when supported. Returns true when
// the provider accepted the observer and false when it did not
// implement ObservableProvider.
func SetObserver(p Provider, obs Observer) bool {
	if op, ok := p.(ObservableProvider); ok {
		op.SetObserver(obs)
		return true
	}
	return false
}

// noopObserver is a sentinel used by providers so hook sites can call
// through unconditionally without a nil check. It is not exported —
// callers that want a no-op should pass nil.
type noopObserver struct{}

func (noopObserver) OnSend(*url.URL, Message, error, time.Duration) {}
func (noopObserver) OnReceive(*url.URL, Message, error)             {}
func (noopObserver) OnAck(*url.URL, Message)                        {}
func (noopObserver) OnNack(*url.URL, Message, bool)                 {}
