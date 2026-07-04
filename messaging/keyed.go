package messaging

// Keyed is an optional capability a Message may implement to expose a
// routing key. Providers use the key for stable, partitioned dispatch:
//
//   - LocalProvider: when multiple named-listener groups are registered
//     for a URL, hash(key) % len(groups) selects a single group instead of
//     firing one listener per group.
//
//   - Broker satellites (Kafka, JetStream, RabbitMQ, …) map the key to
//     the broker's own partitioning primitive (partition selector,
//     subject token, routing key).
//
// Messages that do not implement Keyed are treated as unkeyed: providers
// fall back to whatever unkeyed policy they document (random pick, round
// robin, sticky-first, …). The package-level helper RoutingKeyOf returns
// "" for those messages so callers don't have to type-assert.
type Keyed interface {
	// RoutingKey returns the message's current routing key, or "" if
	// none is set.
	RoutingKey() string
	// SetRoutingKey stores the routing key on the message. Callers may
	// call this at any point before dispatch; providers snapshot the key
	// when they need it.
	SetRoutingKey(key string)
}

// RoutingKeyOf returns the message's routing key if it implements Keyed,
// or "" otherwise. This is the recommended way to read the routing key
// from provider code — it avoids sprinkling type assertions across the
// dispatch path.
func RoutingKeyOf(msg Message) string {
	if k, ok := msg.(Keyed); ok {
		return k.RoutingKey()
	}
	return ""
}
