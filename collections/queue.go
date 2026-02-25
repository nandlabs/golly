package collections

// Queue is an interface that represents a queue data structure
type Queue[T any] interface {
	Collection[T]
	// Remove and return the element at the front of the queue
	Dequeue() (T, error)
	// Add an element to the queue
	Enqueue(elem T) error
	// Return the element at the front of the queue without removing it
	Front() (T, error)
}

type queueImpl[T any] struct {
	List[T]
}

// NewQueue creates a new Queue
func NewArrayQueue[T any]() Queue[T] {
	return &queueImpl[T]{NewArrayList[T]()}
}

// Remove and return the element at the front of the queue
func (q *queueImpl[T]) Dequeue() (v T, err error) {
	if q.IsEmpty() {
		err = ErrEmptyCollection
	} else {
		v, err = q.RemoveFirst()
	}
	return
}

// Add an element to the queue
func (q *queueImpl[T]) Enqueue(elem T) error {
	return q.AddLast(elem)
}

// Return the element at the front of the queue without removing it
func (q *queueImpl[T]) Front() (v T, err error) {
	if q.IsEmpty() {
		err = ErrEmptyCollection
	} else {
		return q.GetFirst()
	}
	return
}

// syncQueueImpl is a thread-safe version of Queue
type syncQueueImpl[T any] struct {
	*SyncedArrayList[T]
}

// NewSyncQueue creates a new thread-safe Queue
func NewSyncQueue[T any]() Queue[T] {
	return &syncQueueImpl[T]{NewSyncedArrayList[T]()}
}

// Remove and return the element at the front of the queue
func (sq *syncQueueImpl[T]) Dequeue() (T, error) {
	return sq.RemoveFirst()
}

// Add an element to the queue
func (sq *syncQueueImpl[T]) Enqueue(elem T) error {
	return sq.AddLast(elem)

}

// Return the element at the front of the queue without removing it
func (sq *syncQueueImpl[T]) Front() (T, error) {

	return sq.GetFirst()
}
