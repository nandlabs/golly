package clients

import (
	"math"
	"time"

	"oss.nandlabs.io/golly/config"
)

type RetryPolicy struct {
	MaxRetries      int
	BackoffInterval time.Duration
	MaxBackoff      time.Duration
	Exponential     bool
}

func (r *RetryPolicy) WaitTime(retryCount int) time.Duration {
	backoff := r.BackoffInterval
	if r.Exponential {
		multiple := math.Pow(2, float64(retryCount))
		backoff = time.Duration(multiple) * r.BackoffInterval
		backoff = min(backoff, r.MaxBackoff)
	}
	return backoff
}

type ClientOptions struct {
	// Attributes
	Attributes config.Attributes
	// RetryPolicy holds the retry configuration for the client
	RetryPolicy *RetryPolicy
	// CircuitBreaker holds the circuit breaker configuration for the client
	CircuitBreaker *CircuitBreaker
	// Auth holds the authentication configuration for the client
	Auth AuthProvider
}

type OptionsBuilder struct {
	options ClientOptions
}

func NewOptionsBuilder() *OptionsBuilder {
	return &OptionsBuilder{
		options: ClientOptions{
			Attributes:     config.NewMapAttributes(),
			CircuitBreaker: nil,
			Auth:           nil,
			RetryPolicy:    nil,
		},
	}
}

// WithAtributes sets the attributes for the client.
// Parameters:
//   - attributes: The attributes to set.
//
// Returns:
//
//	*OptionsBuilder: The options builder.
func (b *OptionsBuilder) WithAttributes(attributes config.Attributes) *OptionsBuilder {
	b.options.Attributes = attributes
	return b
}

// WithAuth sets the authenticator for the client.
// Parameters:
//   - authenticator: The authenticator to set.
//
// Returns:
//
//	*OptionsBuilder: The options builder.
func (b *OptionsBuilder) WithAuth(authenticator AuthProvider) *OptionsBuilder {
	b.options.Auth = authenticator
	return b
}

// WithRetryPolicy sets the retry policy for the client.
// Parameters:
//   - retryPolicy: The retry policy to set.
//
// Returns:
//	*OptionsBuilder: The options builder.

func (b *OptionsBuilder) WithRetryPolicy(retryPolicy *RetryPolicy) *OptionsBuilder {
	b.options.RetryPolicy = retryPolicy
	return b
}

// WithCircuitBreaker sets the circuit breaker for the client.
// Parameters:
//   - circuitBreaker: The circuit breaker to set.
//
// Returns:
//	*OptionsBuilder: The options builder.

func (b *OptionsBuilder) WithCircuitBreaker(circuitBreaker *CircuitBreaker) *OptionsBuilder {
	b.options.CircuitBreaker = circuitBreaker
	return b
}

// AddRetryPolicy adds a retry policy to the client.
// Parameters:
//   - maxRetries: The maximum number of retries allowed.
//   - backoffIntervalMs: The wait time between retries in milliseconds. If set to <=0, the client will retry immediately.

func (b *OptionsBuilder) RetryPolicy(maxRetries int, backoffIntervalMs int, exponential bool, maxBackoffInMs int) *OptionsBuilder {
	b.options.RetryPolicy = &RetryPolicy{
		MaxRetries:      maxRetries,
		BackoffInterval: time.Duration(backoffIntervalMs) * time.Millisecond,
		Exponential:     exponential,
		MaxBackoff:      time.Duration(maxBackoffInMs) * time.Millisecond,
	}
	return b
}

// AddCircuitBreaker adds a circuit breaker to the client.
// Parameters:
//   - failureThreshold: The number of consecutive failures required to open the circuit.
//   - successThreshold: The number of consecutive successes required to close the circuit.
//   - maxHalfOpen: The maximum number of requests allowed in the half-open state.
//   - timeout: The timeout duration for the circuit to transition from open to half-open state.
func (b *OptionsBuilder) CircuitBreaker(failureThreshold uint64, successThreshold uint64, maxHalfOpen uint32, timeout uint32) *OptionsBuilder {
	breakerInfo := &BreakerInfo{
		FailureThreshold: failureThreshold,
		SuccessThreshold: successThreshold,
		MaxHalfOpen:      maxHalfOpen,
		Timeout:          timeout,
	}
	b.options.CircuitBreaker = NewCircuitBreaker(breakerInfo)
	return b
}

// AddBasicAuth adds basic authentication to the client.
// This method will override any existing authentication configuration.
// Parameters:
//   - user: The username.
//   - pass: The password.
func (b *OptionsBuilder) BasicAuth(user, pass string) *OptionsBuilder {
	b.options.Auth = &BasicAuth{
		user: user,
		pass: pass,
	}
	return b
}

// AddTokenAuth adds token authentication to the client.
// This method will override any existing authentication configuration.
// Parameters:
//   - token: The token.
func (b *OptionsBuilder) TokenBearerAuth(token string) *OptionsBuilder {
	b.options.Auth = &TokenBearerAuth{
		token: token,
	}
	return b
}

// Build creates a new ClientOptions with the provided configuration.
// Returns:
//
//	*ClientOptions: The client.
func (b *OptionsBuilder) Build() *ClientOptions {
	return &b.options
}
