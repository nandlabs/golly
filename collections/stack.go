package collections

import "fmt"

// Stack is an interface that represents a stack data structure
type Stack[T any] interface {
	Collection[T]
	// Push an element onto the stack
	Push(elem T)
	// Pop and return the element at the top of the stack
	Pop() (T, error)
	// Return the element at the top of the stack without removing it
	Peek() (T, error)
}

// StackImpl is a basic implementation of the Stack interface
type stackImpl[T any] struct {
	*ArrayList[T]
}

// NewStack creates a new Stack
func NewStack[T any]() Stack[T] {
	return &stackImpl[T]{NewArrayList[T]()}
}

// Push an element onto the stack
func (s *stackImpl[T]) Push(elem T) {
	_ = s.AddLast(elem)
}

// Pop and return the element at the top of the stack
func (s *stackImpl[T]) Pop() (v T, err error) {
	if s.IsEmpty() {
		err = ErrEmptyCollection
		return
	}
	elem, _ := s.GetLast()
	s.Remove(elem)
	return elem, nil
}

// Peek returns the element at the top of the stack without removing it
func (s *stackImpl[T]) Peek() (v T, err error) {
	if s.IsEmpty() {
		err = ErrEmptyCollection
		return
	}
	elem, _ := s.GetLast()
	return elem, nil
}

// Iterator returns an Iterator for the stack
func (s *stackImpl[T]) Iterator() Iterator[T] {
	fmt.Println("Stack Iterator")
	return &stackIterator[T]{s.ArrayList, s.Size() - 1}
}

type stackIterator[T any] struct {
	*ArrayList[T]
	index int
}

// HasNext returns true if there are more elements in the stack
func (si *stackIterator[T]) HasNext() bool {
	return si.index > -1
}

// Next returns the next element in the stack
func (si *stackIterator[T]) Next() T {
	elem := si.elements[si.index]
	si.index--
	return elem
}

// Remove removes the last element returned by the iterator
func (si *stackIterator[T]) Remove() {
	_, _ = si.RemoveAt(si.index)
	si.index--
}

// syncStack is a thread-safe version of the Stack interface
type syncStackImpl[T any] struct {
	*SyncedArrayList[T]
}

// NewSyncStack creates a new synchronized Stack
func NewSyncStack[T any]() Stack[T] {
	return &syncStackImpl[T]{NewSyncedArrayList[T]()}
}

// Push an element onto the stack
func (ss *syncStackImpl[T]) Push(elem T) {
	_ = ss.AddLast(elem)
}

// Pop and return the element at the top of the stack
func (ss *syncStackImpl[T]) Pop() (T, error) {
	return ss.RemoveLast()
}

// Peek returns the element at the top of the stack without removing it
func (ss *syncStackImpl[T]) Peek() (T, error) {
	return ss.GetLast()
}

// Iterator returns an Iterator for the synchronized stack
func (ss *syncStackImpl[T]) Iterator() Iterator[T] {
	fmt.Println("Synced Stack Iterator")
	return &syncStackIterator[T]{list: ss.SyncedArrayList, index: ss.Size() - 1}
}

// syncStackIterator is a thread-safe version of the stackIterator
type syncStackIterator[T any] struct {
	list  *SyncedArrayList[T]
	index int
}

// HasNext returns true if there are more elements in the stack
func (ssi *syncStackIterator[T]) HasNext() bool {
	ssi.list.mutex.RLock()
	defer ssi.list.mutex.RUnlock()
	return ssi.index > -1
}

// Next returns the next element in the stack
func (ssi *syncStackIterator[T]) Next() T {
	ssi.list.mutex.RLock()
	defer ssi.list.mutex.RUnlock()
	// TODO: Handle error
	elem, _ := ssi.list.Get(ssi.index)
	ssi.index--
	return elem
}

// Remove removes the last element returned by the iterator
func (ssi *syncStackIterator[T]) Remove() {

	if ssi.index > -1 && ssi.index < ssi.list.Size() {
		_, _ = ssi.list.RemoveAt(ssi.index)
	}
	ssi.index--
}
