package ws

import (
	"crypto/tls"
	"time"
)

const (
	defaultReadBufferSize   = 4096
	defaultWriteBufferSize  = 4096
	defaultMaxMessageSize   = 64 * 1024 // 64KB
	defaultPingInterval     = 30 * time.Second
	defaultPongTimeout      = 10 * time.Second
	defaultWriteTimeout     = 10 * time.Second
	defaultHandshakeTimeout = 10 * time.Second
)

// config holds the shared configuration for server handlers and client connections.
type config struct {
	readBufferSize   int
	writeBufferSize  int
	maxMessageSize   int64
	pingInterval     time.Duration
	pongTimeout      time.Duration
	writeTimeout     time.Duration
	handshakeTimeout time.Duration
	tlsConfig        *tls.Config
	autoReconnect    bool
	maxReconnectWait time.Duration
	checkOrigin      func(origin string) bool
}

func defaultConfig() *config {
	return &config{
		readBufferSize:   defaultReadBufferSize,
		writeBufferSize:  defaultWriteBufferSize,
		maxMessageSize:   defaultMaxMessageSize,
		pingInterval:     defaultPingInterval,
		pongTimeout:      defaultPongTimeout,
		writeTimeout:     defaultWriteTimeout,
		handshakeTimeout: defaultHandshakeTimeout,
		maxReconnectWait: 30 * time.Second,
	}
}

// Option is a function that configures a WebSocket handler or client.
type Option func(*config)

// WithReadBufferSize sets the size of the read buffer in bytes.
func WithReadBufferSize(size int) Option {
	return func(c *config) {
		if size > 0 {
			c.readBufferSize = size
		}
	}
}

// WithWriteBufferSize sets the size of the write buffer in bytes.
func WithWriteBufferSize(size int) Option {
	return func(c *config) {
		if size > 0 {
			c.writeBufferSize = size
		}
	}
}

// WithMaxMessageSize sets the maximum allowed message size in bytes.
func WithMaxMessageSize(size int64) Option {
	return func(c *config) {
		if size > 0 {
			c.maxMessageSize = size
		}
	}
}

// WithPingInterval sets the interval between ping frames sent to the peer.
// Set to 0 to disable pings.
func WithPingInterval(d time.Duration) Option {
	return func(c *config) {
		c.pingInterval = d
	}
}

// WithPongTimeout sets the time to wait for a pong response before closing.
func WithPongTimeout(d time.Duration) Option {
	return func(c *config) {
		if d > 0 {
			c.pongTimeout = d
		}
	}
}

// WithWriteTimeout sets the timeout for write operations.
func WithWriteTimeout(d time.Duration) Option {
	return func(c *config) {
		if d > 0 {
			c.writeTimeout = d
		}
	}
}

// WithHandshakeTimeout sets the timeout for the WebSocket handshake.
func WithHandshakeTimeout(d time.Duration) Option {
	return func(c *config) {
		if d > 0 {
			c.handshakeTimeout = d
		}
	}
}

// WithTLSConfig sets the TLS configuration for secure WebSocket connections.
func WithTLSConfig(tc *tls.Config) Option {
	return func(c *config) {
		c.tlsConfig = tc
	}
}

// WithAutoReconnect enables automatic reconnection for the client.
func WithAutoReconnect(enabled bool) Option {
	return func(c *config) {
		c.autoReconnect = enabled
	}
}

// WithMaxReconnectWait sets the maximum wait time between reconnection attempts.
func WithMaxReconnectWait(d time.Duration) Option {
	return func(c *config) {
		if d > 0 {
			c.maxReconnectWait = d
		}
	}
}

// WithCheckOrigin sets a function to validate the Origin header during the handshake.
// If nil, all origins are accepted.
func WithCheckOrigin(fn func(origin string) bool) Option {
	return func(c *config) {
		c.checkOrigin = fn
	}
}
