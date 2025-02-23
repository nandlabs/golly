package clients

import "time"

type RetryPolicy struct {
	MaxRetries      int
	BackoffInterval time.Duration
}

type ClientOptions struct {
	// RetryInfo holds the retry configuration for the client
	RetryInfo *RetryInfo
	// CircuitBreaker holds the circuit breaker configuration for the client
	CircuitBreaker *CircuitBreaker
}
