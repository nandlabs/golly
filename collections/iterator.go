package collections

// Iterator is an interface that allows iterating over a collection
type Iterator[T any] interface {
	// HasNext returns true if there are more elements in the collection
	HasNext() bool
	// Next returns the next element in the collection
	Next() T
	// Remove removes the last element returned by the iterator from the collection
	Remove()
}

type Iterable[T any] interface {
	Iterator() Iterator[T]
}
