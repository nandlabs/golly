# MESSAGING Client

This is a flexible and extensible messaging client designed to provide a unified interface for producing and consuming messages across various messaging platforms. It allows developers to seamlessly integrate messaging functionality into their applications without being tied to a specific messaging service.

---

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Delivery guarantees](#delivery-guarantees)
- [Capability interfaces](#capability-interfaces)
- [Option keys](#option-keys)
- [Extending the library](#extending-the-library)

---

## Features

- General producer interface for sending messages to different messaging platforms.
- General consumer interface for receiving and processing messages from different messaging platforms.
- Context-aware Send / Receive / AddListener (v1.7).
- Ack-aware dispatch on the in-tree `chan://` provider with configurable redelivery + dead-lettering (v1.7).
- Routing-key aware dispatch via the optional `Keyed` message capability (v1.7).
- Bounded consumer concurrency + configurable channel backpressure (v1.7).
- Metrics / tracing hooks via the `Observer` interface (v1.7).
- Adapter into `lifecycle.Component` for graceful startup / shutdown (v1.7).
- Handler retry + dead-letter wrapper (`WithRetry`) (v1.7).
- Supports QualityOfService features such as
  - CircuitBreaker
  - Retry
- Easy-to-use message abstraction for consistent handling of messages across platforms.
- Can be extended to multiple messaging platforms with easily pluggable interfaces, including:
  - AMQP (Advanced Message Queuing Protocol)
  - Apache Kafka
  - AWS SNS (Simple Notification Service)
  - AWS SQS (Simple Queue Service)
  - GCM (Google Cloud Messaging)
  - GCP Pub/Sub (Google Cloud Pub/Sub)

## Installation

To install the messaging client, use the following command:

```bash
go get oss.nandlabs.io/golly/clients/messaging
```

## Usage

1. Import the library into your Go project:

    ```go
    import "oss.nandlabs.io/golly/clients/messaging"
    ```

2. Initialize the messaging provider for a specific platform. For example, to use the AMQP extension:

    ```go
    type AMQPProvider struct {} // implements the Provider interface defined under the library
    amqpProvider := &AMQPProvider{}

    manager := messaging.GetManager()
    manager.Register(amqpProvider)
    ```

3. Send a message

    ```go
    msg, _ := manager.NewMessage("chan")
    _, _ = msg.SetBodyStr("Hello, World!")
    destination, _ := url.Parse("chan://demo")
    if err := manager.Send(destination, msg); err != nil {
        // handle error
    }
    ```

4. Receive a message via an on-demand call or a listener

    ```go
    // Listener form
    _ = manager.AddListener(destination, func(m messaging.Message) {
        // process m
    })

    // On-demand form
    m, err := manager.Receive(destination)
    ```

5. Ack-aware consumption (v1.7)

    ```go
    _ = manager.AddListener(destination, func(m messaging.Message) {
        if err := process(m); err != nil {
            _ = m.Rsvp(false, messaging.NewOptionsBuilder().AddRequeue().Build()...)
            return
        }
        _ = m.Rsvp(true)
    }, messaging.NewOptionsBuilder().
        AddRequireAck().
        AddAckTimeout(5*time.Second).
        AddMaxDeliveryAttempts(3).
        AddDeadLetter("chan://demo-dlq").
        Build()...)
    ```

## Delivery guarantees

| Provider              | Default             | Configurable via                                   |
| --------------------- | ------------------- | -------------------------------------------------- |
| LocalProvider (chan)  | At-most-once        | `RequireAckOpt` → at-least-once with redelivery    |
| SQS / Pub-Sub / Kafka | At-least-once       | `DeliveryGuaranteeOpt` (AtMostOnce / ExactlyOnce)  |
| JetStream             | At-least-once       | `DeliveryGuaranteeOpt` (ExactlyOnce with EOS)      |

Local provider details:

- **Without `RequireAckOpt`**: fire-and-forget. `Send` returns as soon as the message is queued. Every unnamed listener runs per message; each named group fires one listener per message.
- **With `RequireAckOpt`**: the dispatcher waits for `Message.Rsvp` (or `AckTimeoutOpt`, default 30s). Nack + `RequeueOpt` requeues (up to `MaxDeliveryAttemptsOpt`, default 3). Terminal failures route to `DeadLetterOpt` or drop.

## Capability interfaces

Every v1.7 addition is a **capability interface** — providers opt in by implementing it alongside `Provider`. Type-assert (or use the package helpers) to detect support:

| Interface            | Purpose                                            | Package helper                                   |
| -------------------- | -------------------------------------------------- | ------------------------------------------------ |
| `ProducerCtx`        | Context-aware Send / SendBatch                     | `messaging.SendCtx`, `messaging.SendBatchCtx`    |
| `ReceiverCtx`        | Context-aware Receive / ReceiveBatch / AddListener | `messaging.ReceiveCtx`, `messaging.AddListenerCtx` |
| `ListenerRemover`    | RemoveListeners / RemoveNamedListener              | Direct type assertion                            |
| `Keyed` (on Message) | Routing key for partitioned dispatch               | `messaging.RoutingKeyOf`                         |
| `ObservableProvider` | SetObserver for metrics / tracing                  | `messaging.SetObserver`                          |
| `LifecycleProvider`  | Provider + `lifecycle.Component`                   | `messaging.AsLifecycleComponent`                 |
| `Drainer`            | Graceful in-flight wait                            | `AsLifecycleComponent` calls it internally       |

## Option keys

New in v1.7 (all additive):

| Key                    | Type            | Meaning                                                        |
| ---------------------- | --------------- | -------------------------------------------------------------- |
| `RequireAckOpt`        | bool            | Dispatcher waits for `Message.Rsvp` before moving on           |
| `RequeueOpt`           | bool            | Passed to `Rsvp(false, …)` — asks the provider to redeliver    |
| `ConsumerConcurrencyOpt` | int           | Cap in-flight listener invocations per URL                     |
| `LocalBufferSizeOpt`   | int             | Per-destination buffer size (LocalProvider, at first creation) |
| `BlockOnFullOpt`       | bool            | Send blocks instead of returning `ErrChannelFull`              |
| `RoutingKeyOpt`        | string          | Attach a routing key to the Send call                          |

Pre-existing (broker-side advisory):

| Key                        | Type              | Meaning                                            |
| -------------------------- | ----------------- | -------------------------------------------------- |
| `DeliveryGuaranteeOpt`     | DeliveryGuarantee | AtMostOnce / AtLeastOnce / ExactlyOnce             |
| `DeadLetterOpt`            | string            | Dead-letter destination                            |
| `MaxDeliveryAttemptsOpt`   | int               | Redelivery cap before DLQ                          |
| `ConsumerModeOpt`          | ConsumerMode      | Push vs pull                                       |
| `AckTimeoutOpt`            | time.Duration     | Per-message ack window                             |

## Extending the library

To add support for additional messaging platforms, implement `Provider` and — where the semantics fit — one or more of the capability interfaces listed above. Concrete providers are free to ignore option keys they cannot honor (the LocalProvider ignores broker-side keys like `ConsumerModeOpt`, for example), but should return an explicit error at construction time when a request they cannot satisfy is unavoidable.

Refer to the in-tree `LocalProvider` implementation for a reference wiring of all capability interfaces (`ProducerCtx`, `ReceiverCtx`, `ListenerRemover`, `ObservableProvider`, `Drainer`).
