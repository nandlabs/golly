package messaging

import "time"

// Broker-targeted option keys. These are recognized by durable / brokered
// Providers (JetStream, Kafka, RabbitMQ, …) that ship as satellite modules.
// The in-tree LocalProvider ignores them — they're advisory hints, not
// hard requirements for the Provider interface.
const (
	// DeliveryGuaranteeOpt selects the broker's delivery semantics.
	// Value: DeliveryGuarantee (AtMostOnce / AtLeastOnce / ExactlyOnce).
	DeliveryGuaranteeOpt = "DeliveryGuarantee"

	// DeadLetterOpt names a destination topic/queue/stream that receives
	// messages whose primary delivery exhausts retries or violates schema.
	// Value: string (broker-specific identifier).
	DeadLetterOpt = "DeadLetter"

	// MaxDeliveryAttemptsOpt caps how many times a broker will redeliver a
	// message before routing it to the DeadLetter destination (when set).
	// Value: int (> 0). Brokers without redelivery semantics ignore it.
	MaxDeliveryAttemptsOpt = "MaxDeliveryAttempts"

	// ConsumerModeOpt selects between long-poll/push or pull-based consumption.
	// Value: ConsumerMode (ConsumerPush / ConsumerPull).
	ConsumerModeOpt = "ConsumerMode"

	// AckTimeoutOpt is the per-message acknowledgment deadline; messages
	// not ack'd within this window are redelivered (subject to
	// MaxDeliveryAttempts). Value: time.Duration.
	AckTimeoutOpt = "AckTimeout"
)

// DeliveryGuarantee is the broker-side delivery semantic for produced messages.
type DeliveryGuarantee string

const (
	// AtMostOnce is the cheapest mode — messages may be lost on broker
	// failure, but never duplicated. Suitable for metrics, telemetry, etc.
	AtMostOnce DeliveryGuarantee = "at-most-once"

	// AtLeastOnce is the default for most durable brokers — messages are
	// never lost but may be redelivered on broker / consumer failure.
	// Pair with idempotent consumers.
	AtLeastOnce DeliveryGuarantee = "at-least-once"

	// ExactlyOnce requires broker-side dedup + transactional consumers
	// (e.g. JetStream exactly-once, Kafka EOS). Strongest guarantee,
	// highest overhead.
	ExactlyOnce DeliveryGuarantee = "exactly-once"
)

// ConsumerMode selects between server-pushed delivery and consumer-pulled fetch.
type ConsumerMode string

const (
	// ConsumerPush is the default for most brokers — the server streams
	// messages as they arrive; the consumer's handler is invoked per message.
	ConsumerPush ConsumerMode = "push"

	// ConsumerPull asks the consumer to explicitly fetch batches at its
	// own pace. Useful for back-pressure control on slow consumers.
	ConsumerPull ConsumerMode = "pull"
)

// --- OptionsBuilder additions ---

// AddDeliveryGuarantee sets the broker's delivery semantics for produced
// messages. Brokers without the requested mode (e.g. AtMostOnce on a strict
// JetStream stream) should return an error from Provider.NewProducer.
func (ob *OptionsBuilder) AddDeliveryGuarantee(g DeliveryGuarantee) *OptionsBuilder {
	return ob.Add(DeliveryGuaranteeOpt, g)
}

// AddDeadLetter routes messages that exhaust retries (or violate schema) to
// destination. The destination format is broker-specific (a subject in NATS,
// a topic in Kafka, etc.). Pair with AddMaxDeliveryAttempts for an explicit
// redelivery cap.
func (ob *OptionsBuilder) AddDeadLetter(destination string) *OptionsBuilder {
	return ob.Add(DeadLetterOpt, destination)
}

// AddMaxDeliveryAttempts caps redelivery attempts before routing to the
// DeadLetter destination (when configured). n must be > 0.
func (ob *OptionsBuilder) AddMaxDeliveryAttempts(n int) *OptionsBuilder {
	if n <= 0 {
		return ob
	}
	return ob.Add(MaxDeliveryAttemptsOpt, n)
}

// AddConsumerMode selects push vs pull delivery on consumers that support
// both (most JetStream / Kafka consumers).
func (ob *OptionsBuilder) AddConsumerMode(m ConsumerMode) *OptionsBuilder {
	return ob.Add(ConsumerModeOpt, m)
}

// AddAckTimeout sets the per-message acknowledgment window. Messages not
// ack'd within d are redelivered (subject to MaxDeliveryAttempts).
func (ob *OptionsBuilder) AddAckTimeout(d time.Duration) *OptionsBuilder {
	if d <= 0 {
		return ob
	}
	return ob.Add(AckTimeoutOpt, d)
}
