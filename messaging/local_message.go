package messaging

import "sync/atomic"

// LocalMessage is the in-tree Message implementation used by the chan://
// provider. It embeds BaseMessage — so it inherits Header / Body / Keyed
// support — and augments it with an optional acknowledgment channel and
// delivery-attempt counter used by the ack-aware dispatcher.
//
// The acknowledgment plumbing is opt-in: LocalMessage created via
// NewLocalMessage has no ack channel and Rsvp is a no-op — matching the
// pre-v1.7 behavior. When the dispatcher tracks delivery (RequireAckOpt),
// it attaches an ack channel via bindAck before firing the listener.
type LocalMessage struct {
	*BaseMessage
	ack      chan ackSignal
	attempts int32
}

// ackSignal is the internal envelope carried on the ack channel: it
// records whether the listener acknowledged (yes) and, on Nack, whether
// the message should be redelivered.
type ackSignal struct {
	ok      bool
	requeue bool
}

// NewLocalMessage returns a fresh LocalMessage suitable for chan://
// sends. The returned message has no ack channel bound; the dispatcher
// binds one when the target URL is configured with RequireAckOpt.
func NewLocalMessage() (msg Message, err error) {
	var baseMsg *BaseMessage
	baseMsg, err = NewBaseMessage()
	if err == nil {
		msg = &LocalMessage{
			BaseMessage: baseMsg,
		}
	}
	return
}

// Rsvp acknowledges (yes=true) or negatively acknowledges (yes=false)
// the message. When no ack channel is bound (default) it is a no-op —
// matching the historical LocalMessage semantics.
//
// With an ack channel bound (see bindAck), the first Rsvp call after
// dispatch signals the dispatcher; subsequent calls are dropped. On
// Nack, the RequeueOpt option controls whether the provider redelivers
// the message or routes it to the DeadLetter destination.
func (lm *LocalMessage) Rsvp(yes bool, options ...Option) (err error) {
	if lm.ack == nil {
		return
	}
	sig := ackSignal{ok: yes}
	if !yes {
		if v, has := GetOptValue[bool](RequeueOpt, options...); has {
			sig.requeue = v
		}
	}
	// Non-blocking: subsequent Rsvp calls are ignored once a signal is queued.
	select {
	case lm.ack <- sig:
	default:
	}
	return
}

// bindAck attaches an ack channel to the message. Called by the
// dispatcher just before firing the listener when RequireAckOpt is set
// on the URL. The channel is buffered to size 1 so Rsvp is non-blocking
// even if the dispatcher's timeout has already fired.
func (lm *LocalMessage) bindAck(ch chan ackSignal) {
	lm.ack = ch
}

// deliveryAttempts returns the number of times this message has been
// dispatched. The counter is incremented by the dispatcher before each
// attempt and read by the DeadLetter routing logic.
func (lm *LocalMessage) deliveryAttempts() int32 {
	return atomic.LoadInt32(&lm.attempts)
}

// incrementAttempts atomically increments the delivery-attempt counter
// and returns the new value.
func (lm *LocalMessage) incrementAttempts() int32 {
	return atomic.AddInt32(&lm.attempts, 1)
}
