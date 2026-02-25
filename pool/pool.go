package pool

import (
	"errors"
)

var ErrCacheFull = errors.New("cache is full")
var ErrPoolClosed = errors.New("pool is closed")
var ErrObjectNotFound = errors.New("object not found in pool")
var ErrInvalidConfig = errors.New("invalid pool configuration")

// ObjectHandler is a generic function type that takes an argument of type T and returns an error.
// It is used for handling objects in a cache (e.g. destroying / cleaning up).
type ObjectHandler[T any] func(T) error

// ObjectCreator is a function type for creating new objects of type T.
type ObjectCreator[T any] func() (T, error)

// Pool is a generic interface for object pooling.
type Pool[T any] interface {
	// Creator returns the function used to create new objects.
	Creator() ObjectCreator[T]
	// Destroyer returns the function used to destroy objects.
	Destroyer() ObjectHandler[T]
	// Checkout retrieves an object for use. Blocks up to MaxWait seconds if the pool is exhausted.
	Checkout() (T, error)
	// Checkin returns an object to the pool.
	Checkin(T)
	// Delete removes a specific object from the pool and destroys it.
	Delete(T)
	// Clear removes all idle objects from the pool and destroys them.
	Clear()
	// Min returns the minimum number of objects to keep in the pool.
	Min() int
	// SetMin sets the minimum number of objects.
	SetMin(int)
	// Max returns the maximum number of objects allowed.
	Max() int
	// SetMax sets the maximum number of objects allowed.
	SetMax(int)
	// Current returns the total number of live objects (idle + in-use).
	Current() int
	// HighWaterMark returns the peak number of concurrent objects observed.
	HighWaterMark() int
	// LowWaterMark returns the minimum number of concurrent objects observed after initial fill.
	LowWaterMark() int
	// IdleTimeout returns the idle timeout in seconds.
	IdleTimeout() int
	// SetIdleTimeout sets the idle timeout in seconds.
	SetIdleTimeout(int)
	// MaxWait returns the maximum wait time in seconds for an object to become available.
	MaxWait() int
	// SetMaxWait sets the maximum wait time in seconds.
	SetMaxWait(int)
	// Start initializes the pool and pre-creates Min objects.
	Start() error
	// Close drains the pool and destroys all objects.
	Close() error
}

// PooledObject wraps an object with a unique identifier for tracking.
type PooledObject[T any] struct {
	// id is the unique identifier for the object.
	id int
	// obj is the object of type T that is being wrapped.
	obj T
}
