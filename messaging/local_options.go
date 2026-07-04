package messaging

// Local-provider option keys. These are honored by the in-tree LocalProvider
// (chan:// scheme) and, where the semantics fit, may also be honored by
// satellite providers that opt in. They are additive: unset options fall
// back to the historical behavior of the LocalProvider.
const (
	// RequireAckOpt is a bool. When passed to Receiver.AddListener, the
	// LocalProvider dispatcher waits for the listener to call Message.Rsvp
	// (or an ack timeout) before moving on to the next message. Without it,
	// dispatch is fire-and-forget (at-most-once).
	//
	// Value: bool. Default false.
	RequireAckOpt = "RequireAck"

	// RequeueOpt is a bool. When passed to Message.Rsvp(false, opts...) it
	// asks the provider to redeliver the message (subject to
	// MaxDeliveryAttemptsOpt). Without it, a Nack routes to the DeadLetter
	// destination (if configured) or drops the message.
	//
	// Value: bool. Default false.
	RequeueOpt = "Requeue"

	// ConsumerConcurrencyOpt is an int. When passed to
	// Receiver.AddListener, the LocalProvider caps the number of in-flight
	// listener invocations for that URL at the given value. Without it,
	// each message spawns a fresh goroutine per matching listener
	// (unbounded).
	//
	// Value: int (> 0). Default: unbounded.
	ConsumerConcurrencyOpt = "ConsumerConcurrency"

	// LocalBufferSizeOpt is an int. When the first Send / AddListener /
	// Receive call for a URL specifies this option, the destination channel
	// is allocated with the given buffer size instead of the package
	// default. Once the destination exists, further attempts to change the
	// size are ignored.
	//
	// Value: int (> 0). Default: 256 (or the value set via
	// SetDefaultBufferSize).
	LocalBufferSizeOpt = "LocalBufferSize"

	// BlockOnFullOpt is a bool. When true, Send blocks until the
	// destination channel has room (or ctx cancels via SendCtx). When
	// false (default), Send returns ErrChannelFull immediately when the
	// buffer is full.
	//
	// Value: bool. Default false.
	BlockOnFullOpt = "BlockOnFull"

	// RoutingKeyOpt is a string. Setting it on a send-time Option supplies
	// a routing key without requiring the caller to type-assert the
	// Message to Keyed. When both are set, the explicit option wins for
	// the individual Send call; the message's own routing key remains
	// unchanged.
	//
	// Value: string.
	RoutingKeyOpt = "RoutingKey"
)

// --- OptionsBuilder additions for the local-provider option keys ---

// AddRequireAck signals that the LocalProvider dispatcher must wait for
// Message.Rsvp before dispatching the next message on this URL. See
// RequireAckOpt.
func (ob *OptionsBuilder) AddRequireAck() *OptionsBuilder {
	return ob.Add(RequireAckOpt, true)
}

// AddRequeue is used by handlers calling Message.Rsvp(false, ...) to ask
// the provider to redeliver the message rather than drop or dead-letter
// it. See RequeueOpt.
func (ob *OptionsBuilder) AddRequeue() *OptionsBuilder {
	return ob.Add(RequeueOpt, true)
}

// AddConsumerConcurrency caps the LocalProvider's in-flight listener
// invocations at n per URL. n must be > 0. See ConsumerConcurrencyOpt.
func (ob *OptionsBuilder) AddConsumerConcurrency(n int) *OptionsBuilder {
	if n <= 0 {
		return ob
	}
	return ob.Add(ConsumerConcurrencyOpt, n)
}

// AddLocalBufferSize requests a per-destination channel buffer of size n
// (honored only on first creation of the destination). n must be > 0.
// See LocalBufferSizeOpt.
func (ob *OptionsBuilder) AddLocalBufferSize(n int) *OptionsBuilder {
	if n <= 0 {
		return ob
	}
	return ob.Add(LocalBufferSizeOpt, n)
}

// AddBlockOnFull requests that Send block until the buffer has room
// instead of returning ErrChannelFull. See BlockOnFullOpt.
func (ob *OptionsBuilder) AddBlockOnFull() *OptionsBuilder {
	return ob.Add(BlockOnFullOpt, true)
}

// AddRoutingKey attaches a routing key to a Send call. Providers that
// support keyed dispatch (LocalProvider with multiple named groups,
// Kafka partitioners, etc.) use it to select a stable destination.
// See RoutingKeyOpt.
func (ob *OptionsBuilder) AddRoutingKey(key string) *OptionsBuilder {
	return ob.Add(RoutingKeyOpt, key)
}
