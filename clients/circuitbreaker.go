package clients

import (
	"errors"
	"sync/atomic"
	"time"
)

// CircuitBreaker states
const (
	circuitClosed           uint32 = iota // Circuit is closed and requests can flow through
	circuitHalfOpen                       // Circuit is partially open and allows limited requests for testing
	circuitOpen                           // Circuit is open and requests are blocked
	defaultTimeout          = 300
	defaultMaxHalfOpen      = 5
	defaultSuccessThreshold = 3
	defaultFailureThreshold = 3
)

// ErrCBOpen is the error returned when the circuit breaker is open and unable to process requests.
var ErrCBOpen = errors.New("the Circuit breaker is open and unable to process request")

// BreakerInfo holds the configuration parameters for the CircuitBreaker.
type BreakerInfo struct {
	FailureThreshold uint64 // Number of consecutive failures required to open the circuit
	SuccessThreshold uint64 // Number of consecutive successes required to close the circuit
	MaxHalfOpen      uint32 // Maximum number of requests allowed in the half-open state
	Timeout          uint32 // Timeout duration for the circuit to transition from open to half-open state
}

// CircuitBreaker is a struct that represents a circuit breaker.
type CircuitBreaker struct {
	*BreakerInfo
	currentState    uint32 // Current state of the circuit breaker
	successCounter  uint64 // Counter for successful requests
	failureCounter  uint64 // Counter for failed requests
	halfOpenCounter uint32 // Counter for requests in the half-open state
}

// NewCircuitBreaker creates a new CircuitBreaker instance with the provided BreakerInfo.
// If no BreakerInfo is provided, default values will be used.
func NewCircuitBreaker(info *BreakerInfo) (cb *CircuitBreaker) {
	// Set default values if not provided
	if info == nil {
		info = &BreakerInfo{}
	}

	if info.SuccessThreshold == 0 {
		info.SuccessThreshold = defaultSuccessThreshold
	}
	if info.FailureThreshold == 0 {
		info.FailureThreshold = defaultFailureThreshold
	}
	if info.MaxHalfOpen == 0 {
		info.MaxHalfOpen = defaultMaxHalfOpen
	}

	if info.Timeout == 0 {
		info.Timeout = defaultTimeout
	}

	return &CircuitBreaker{
		BreakerInfo:     info,
		successCounter:  0,
		failureCounter:  0,
		halfOpenCounter: 0,
		currentState:    circuitClosed,
	}
}

// CanExecute checks if a request can be executed based on the current state of the circuit breaker.
// It returns an error if the circuit is open or if the maximum number of requests in the half-open state is reached.
func (cb *CircuitBreaker) CanExecute() (err error) {
	state := cb.getState()
	if state == circuitOpen {
		err = ErrCBOpen
	} else if state == circuitHalfOpen {
		val := atomic.AddUint32(&cb.halfOpenCounter, 1)
		if val > cb.MaxHalfOpen {
			cb.updateState(circuitHalfOpen, circuitOpen)
			err = ErrCBOpen
		}
	}
	return
}

// OnExecution is called after a request is executed.
// It updates the success or failure counters based on the result of the request.
// It also checks if the circuit needs to transition to a different state based on the counters and thresholds.
func (cb *CircuitBreaker) OnExecution(success bool) {
	var val uint64
	state := cb.getState()
	if success {
		val = atomic.AddUint64(&cb.successCounter, 1)
		if state == circuitHalfOpen {
			if val >= cb.SuccessThreshold {
				cb.updateState(circuitHalfOpen, circuitClosed)
			}
		}
	} else {
		val = atomic.AddUint64(&cb.failureCounter, 1)
		// Check if the failure threshold is reached
		if state == circuitClosed {
			if val >= cb.FailureThreshold {
				cb.updateState(circuitClosed, circuitOpen)
			}
		}
	}
}

// Reset resets the circuit breaker to its initial state.
func (cb *CircuitBreaker) Reset() {
	atomic.StoreUint32(&cb.currentState, circuitClosed)
	atomic.StoreUint64(&cb.failureCounter, 0)
	atomic.StoreUint64(&cb.successCounter, 0)
	atomic.StoreUint32(&cb.halfOpenCounter, 0)
}

// updateState updates the state of the circuit breaker atomically.
// It also resets the success and failure counters and starts a timer to transition to the half-open state.
func (cb *CircuitBreaker) updateState(oldState, newState uint32) {
	if atomic.CompareAndSwapUint32(&cb.currentState, oldState, newState) {
		atomic.StoreUint64(&cb.successCounter, 0)
		atomic.StoreUint64(&cb.failureCounter, 0)
		atomic.StoreUint32(&cb.halfOpenCounter, 0)
		// Check if moving to circuitOpen state
		if newState == circuitOpen {
			// Start Timer for HalfOpen
			go func() {
				time.Sleep(time.Second * time.Duration(cb.Timeout))
				cb.updateState(circuitOpen, circuitHalfOpen)
			}()
		}
	}
}

// getState returns the current state of the circuit breaker.
func (cb *CircuitBreaker) getState() (s uint32) {
	atomic.LoadUint32(&s)
	return
}
