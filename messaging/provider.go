package messaging

import (
	"io"
	"net/url"
)

// Producer interface is used to send message(s) to a specific provider
type Producer interface {
	// Send function sends an individual message to the url
	Send(*url.URL, Message, ...Option) error
	// SendBatch sends a batch of messages to the url
	SendBatch(*url.URL, []Message, ...Option) error
}

// Receiver interface provides the functions for receiving a message(s)
type Receiver interface {
	// Receive function performs on-demand receive of a single message.
	// This function may or may not wait for the messages to arrive. This is purely dependent on the implementation.
	Receive(*url.URL, ...Option) (Message, error)
	// ReceiveBatch function performs on-demand receive of a batch of messages.
	// This function may or may not wait for the messages to arrive. This is purely dependent on the implementation.
	ReceiveBatch(*url.URL, ...Option) ([]Message, error)
	// AddListener registers a listener for the message
	AddListener(*url.URL, func(msg Message), ...Option) error
}

// ListenerRemover is an optional capability a Receiver may implement to
// support tearing down previously-registered listeners. Providers signal
// support by satisfying this interface in addition to Receiver; callers
// should type-assert before invoking.
//
// Implementations must be idempotent — removing listeners that aren't
// registered (or for an unknown URL) is not an error.
type ListenerRemover interface {
	// RemoveListeners removes every listener registered for the URL.
	// The destination channel itself remains open so Send/Receive continue
	// to work. Returns nil if no listeners were registered (idempotent).
	RemoveListeners(*url.URL) error
	// RemoveNamedListener removes only the named-listener group for the URL,
	// leaving other listeners on the same URL in place. Returns nil if the
	// URL or name is unknown (idempotent).
	RemoveNamedListener(*url.URL, string) error
}

// Provider interface exposes methods for a messaging provider
// It includes Producer and Receiver interfaces
// It also includes Schemes method to get the supported schemes,
// Setup method to perform initial setup and NewMessage method to create a new message
type Provider interface {
	// Extends io.Closer
	io.Closer
	// Producer Interface included
	Producer
	// Receiver interface included
	Receiver
	// Id returns the id of the provider
	Id() string
	// Schemes is array of URL schemes supported by this provider
	Schemes() []string
	// Setup method called
	Setup() error

	// NewMessage function creates a new message that can be used by the clients. It expects the scheme to be provided
	NewMessage(string, ...Option) (Message, error)
}
