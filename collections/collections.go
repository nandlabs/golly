package collections

import (
	"errors"
	"time"
)

// ErrEmptyCollection is an error that is returned when a collection is empty
var ErrEmptyCollection error = errors.New("collection is empty")

// ErrFullCollection is an error that is returned when a collection is full
var ErrFullCollection error = errors.New("collection is full")

// ErrElementNotFound is an error that is returned when an element is not found in the collection
var ErrElementNotFound error = errors.New("element not found")

// ErrIndexOutOfBounds is an error that is returned when an index is out of bounds
var ErrIndexOutOfBounds error = errors.New("index out of bounds")

// ErrInvalidCapacity is an error that is returned when an invalid capacity is specified
var ErrInvalidCapacity error = errors.New("invalid capacity")

// ErrInvalidIndex is an error that is returned when an invalid index is specified
var ErrInvalidIndex error = errors.New("invalid index")

//Collection is a generic interface that defines a collection of elements with various methods to manipulate them.
//The Collection interface uses a type parameter T to represent the type of elements stored in the collection.

type Collection[T any] interface {
	Iterable[T]
	// Add an element to the collection
	Add(elem T) error
	// AddAll adds all elements from another collection to this collection
	AddAll(coll Collection[T]) error
	// AddFirst adds an element at the beginning of the list
	AddFirst(elem T) error
	// AddLast adds an element at the end of the list
	AddLast(elem T) error
	// Clear removes all elements from the list
	Clear()
	// Contains checks if an element is in the collection
	Contains(elem T) bool
	// Return true if the collection is empty
	IsEmpty() bool
	// Remove an element from the collection
	Remove(elem T) bool
	// Return the number of elements in the collection
	Size() int
	// String returns a string representation of the collection
	String() string
}

// BoundCollection is a generic interface that defines a bounded collection of elements with various methods to manipulate them.
// The BoundCollection interface uses a type parameter T to represent the type of elements stored in the collection.

type BoundCollection[T any] interface {
	Collection[T]
	// Return the maximum capacity of the collection
	Capacity() int
	// Offer adds an element to the collection if it is not full
	Offer(elem T, time time.Duration) bool
	// OfferAndWait adds an element to the collection if it is not full, blocking until space is available
	OfferAndWait(elem T)
}
