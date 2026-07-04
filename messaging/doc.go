// Package messaging provides a set of utilities for working with messaging systems.
// It includes functionality for sending and receiving messages, as well as managing
// message queues, on top of a small provider-abstraction (Producer / Receiver /
// Provider). Concrete providers ship in-tree (chan:// LocalProvider) or as
// satellite modules (SQS, Pub/Sub, Kafka, JetStream, AMQP, MQTT, …).
//
// # Core interfaces
//
// The provider abstraction pivots on three interfaces:
//
//   - Producer: Send / SendBatch — write messages to a URL.
//   - Receiver: Receive / ReceiveBatch / AddListener — read messages, either
//     on demand or via a callback.
//   - Provider: extends Producer + Receiver + io.Closer and adds Id / Schemes /
//     Setup / NewMessage.
//
// Every additional feature layered on top is a capability interface providers
// opt into by implementing alongside the base contract. Callers detect the
// capability with a type assertion (or use a package-level helper that does the
// assertion for them). The full v1.7 capability set:
//
//   - ProducerCtx / ReceiverCtx — context-aware send / receive (see context.go).
//   - ListenerRemover — RemoveListeners / RemoveNamedListener (added in v1.6).
//   - Keyed (on Message) — routing key for partitioned dispatch (see keyed.go).
//   - ObservableProvider — SetObserver for metrics / tracing hooks (observer.go).
//   - LifecycleProvider / Drainer — integration with lifecycle.Component
//     (lifecycle.go).
//
// None of these change the base interfaces' signatures — v1.7 is additive.
//
// # Delivery guarantees
//
// Delivery semantics vary per provider. The three modes we care about:
//
//   - At-most-once — messages may be lost on provider failure but never
//     duplicated. Cheap. Suitable for metrics, telemetry, transient signals.
//
//   - At-least-once — messages are never lost but may be redelivered on
//     provider or consumer failure. The default for most durable brokers.
//     Pair with idempotent consumers.
//
//   - Exactly-once — broker-side dedup + transactional consumers. Strongest,
//     highest overhead.
//
// LocalProvider (chan://) semantics:
//
//   - Without RequireAckOpt: at-most-once, fire-and-forget. Send returns as
//     soon as the message is queued; the dispatcher fires each listener in a
//     goroutine (bounded by ConsumerConcurrencyOpt when set) and does not
//     wait for Rsvp. Handler failures are silently dropped.
//
//   - With RequireAckOpt on AddListener: at-least-once with configurable
//     redelivery. The dispatcher waits (up to AckTimeoutOpt) for the listener
//     to call Message.Rsvp. A Nack + RequeueOpt requeues the message (up to
//     MaxDeliveryAttemptsOpt). Nack without requeue, terminal ack-timeout, or
//     attempts-exhausted route to the DeadLetterOpt destination via the
//     package-level Manager; when no DLQ is configured the message is dropped.
//
//   - Backpressure: Send returns ErrChannelFull by default when the buffer is
//     full. BlockOnFullOpt makes Send block until room is available or ctx
//     cancels (via SendCtx). LocalBufferSizeOpt sets the per-destination
//     buffer at creation; SetDefaultBufferSize overrides the global default.
//
// Satellite providers (JetStream, Kafka, SQS, Pub/Sub, AMQP, …) default to
// at-least-once. Callers can request AtMostOnce or ExactlyOnce via
// DeliveryGuaranteeOpt (see broker_options.go); providers that cannot honor
// the requested mode should return an error at construction time.
//
// # Context propagation
//
// Every operation that can wait — Send with BlockOnFullOpt, Receive,
// ReceiveBatch, AddListener — has a ctx-aware variant reachable through the
// ProducerCtx / ReceiverCtx capability interfaces. Callers should prefer the
// package-level helpers (SendCtx, ReceiveCtx, …) over direct type assertions
// — the helpers fall back to the non-ctx methods with best-effort ctx-race
// cancellation when a provider does not implement the capability.
//
// # Lifecycle integration
//
// A Provider can be adapted into a lifecycle.Component via AsLifecycleComponent
// (or AsLifecycleComponentWithTimeout for a custom drain budget). The wrapper's
// Start() calls provider.Setup(); Stop() calls provider.Drain(timeout) when
// the provider implements Drainer, then provider.Close(). Providers that
// implement LifecycleProvider directly can be registered as-is.
//
// # Observability
//
// LocalProvider implements ObservableProvider. Callers can install a metrics
// / tracing hook with SetObserver and see:
//
//   - OnSend — after each Send, with the resulting err and latency.
//   - OnReceive — after each Receive-side yield, including dispatcher-driven
//     listener invocations.
//   - OnAck / OnNack — around Message.Rsvp when the dispatcher tracks acks
//     (RequireAckOpt).
//
// Passing nil clears any previously registered observer; provider hook sites
// always run through the sentinel no-op so callers do not have to guard nil.
//
// # Handler retry + DLQ
//
// WithRetry(handler, policy, dlqURL) returns a handler wrapper that recovers
// from listener panics, retries per policy (exponential backoff / jitter /
// MaxWait cap all supported), and routes to dlqURL on exhaustion — with
// dlqURL nil, exhaustion drops the message. Retry semantics are strictly
// panic-based; the RetrySignal helper lets handlers panic with an unambiguous
// value.
package messaging
