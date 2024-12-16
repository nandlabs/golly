package collections

import (
	"errors"
	"fmt"
	"sync"

	"oss.nandlabs.io/golly/assertion"
)

// ArrayList is a generic list implementation using an array
type ArrayList[T any] struct {
	elements []T
}

// NewArrayList creates a new ArrayList
func NewArrayList[T any]() *ArrayList[T] {
	return &ArrayList[T]{elements: make([]T, 0)}
}

// Add an element to the list
func (al *ArrayList[T]) Add(elem T) (err error) {
	al.elements = append(al.elements, elem)
	return
}

// AddAll adds all elements from another list to this list
func (l *ArrayList[T]) AddAll(list Collection[T]) (err error) {
	it := list.Iterator()
	for it.HasNext() {
		l.Add(it.Next())
	}
	return
}

// AddAt adds an element at the specified index
func (l *ArrayList[T]) AddAt(index int, elem T) error {
	if index < 0 || index > len(l.elements) {
		return errors.New("index out of range")
	}
	l.elements = append(l.elements[:index], append([]T{elem}, l.elements[index:]...)...)
	return nil
}

// AddLast adds an element at the end of the list
func (l *ArrayList[T]) AddFirst(elem T) (err error) {
	l.elements = append([]T{elem}, l.elements...)
	return
}

// AddLast adds an element at the end of the list
func (l *ArrayList[T]) AddLast(elem T) (err error) {
	l.elements = append(l.elements, elem)
	return
}

// Clear removes all elements from the list
func (l *ArrayList[T]) Clear() {
	l.elements = make([]T, 0)
}

// Contains checks if an element is in the list
func (l *ArrayList[T]) Contains(elem T) bool {
	for _, e := range l.elements {
		if assertion.Equal(e, elem) {
			return true
		}
	}
	return false
}

// Get returns the element at the specified index
func (l *ArrayList[T]) Get(index int) (v T, err error) {
	if index < 0 || index >= len(l.elements) {
		err = errors.New("index out of range")
	} else {
		v = l.elements[index]
	}
	return
}

// GetFirst returns the first element in the list
func (l *ArrayList[T]) GetFirst() (v T, err error) {
	if len(l.elements) == 0 {
		err = errors.New("list is empty")
	} else {
		v = l.elements[0]
	}
	return
}

// GetLast returns the last element in the list
func (l *ArrayList[T]) GetLast() (v T, err error) {
	if len(l.elements) == 0 {
		err = errors.New("list is empty")
	} else {
		v = l.elements[len(l.elements)-1]
	}
	return
}

func (l *ArrayList[T]) IndexOf(elem T) int {
	for i, e := range l.elements {
		if assertion.Equal(e, elem) {
			return i
		}
	}
	return -1
}

// IsEmpty checks if the list is empty
func (l *ArrayList[T]) IsEmpty() bool {
	return len(l.elements) == 0
}

// Iterator returns an Iterator for the list
func (l *ArrayList[T]) Iterator() Iterator[T] {
	return &arrayListIterator[T]{list: l, index: 0}
}

// LastIndexOf returns the last index of the specified element
func (l *ArrayList[T]) LastIndexOf(elem T) int {
	for i := len(l.elements) - 1; i >= 0; i-- {
		if assertion.Equal(l.elements[i], elem) {
			return i
		}
	}
	return -1
}

// Clear removes all elements from the list
func (l *ArrayList[T]) Remove(elem T) bool {
	//find the index of the element. Loop through the elements and remove the element
	for i, e := range l.elements {
		if assertion.Equal(e, elem) {
			if i == len(l.elements)-1 {
				l.elements = l.elements[:i]
			} else {
				l.elements = append(l.elements[:i], l.elements[i+1:]...)
			}
			return true
		}
	}
	return false
}

// RemoveAt removes the element at the specified index
func (l *ArrayList[T]) RemoveAt(index int) (v T, err error) {
	if index < 0 || index >= len(l.elements) {
		err = errors.New("index out of range")
		return

	}
	v = l.elements[index]
	l.elements = append(l.elements[:index], l.elements[index+1:]...)
	return
}

// RemoveFirst removes the first element from the list
func (l *ArrayList[T]) RemoveFirst() (T, error) {
	return l.RemoveAt(0)
}

// RemoveLast removes the last element from the list
func (l *ArrayList[T]) RemoveLast() (T, error) {
	return l.RemoveAt(len(l.elements) - 1)
}

// Size returns the number of elements in the list
func (l *ArrayList[T]) Size() int {
	return len(l.elements)
}

// String returns a string representation of the list
func (l *ArrayList[T]) String() string {
	return fmt.Sprintf("%v", l.elements)
}

// arrayListIterator is an iterator for the ArrayList
type arrayListIterator[T any] struct {
	list  *ArrayList[T]
	index int
}

// HasNext returns true if there are more elements in the collection
func (it *arrayListIterator[T]) HasNext() bool {
	return it.index < len(it.list.elements)
}

func (it *arrayListIterator[T]) Remove() {

	it.list.elements = append(it.list.elements[:it.index], it.list.elements[it.index+1:]...)
	it.index--
}

// Next returns the next element in the collection
func (it *arrayListIterator[T]) Next() T {
	elem := it.list.elements[it.index]
	it.index++
	return elem
}

// SyncedArrayList is a synchronized version of the ArrayList
type SyncedArrayList[T any] struct {
	list  *ArrayList[T]
	mutex sync.RWMutex
}

// NewSyncedArrayList creates a new SyncedArrayList
func NewSyncedArrayList[T any]() *SyncedArrayList[T] {
	return &SyncedArrayList[T]{list: NewArrayList[T]()}
}

// Add an element to the list
func (sal *SyncedArrayList[T]) Add(elem T) error {
	sal.mutex.Lock()
	defer sal.mutex.Unlock()
	return sal.list.Add(elem)
}

// AddAll adds all elements from another list to this list
func (sal *SyncedArrayList[T]) AddAll(list Collection[T]) error {
	sal.mutex.Lock()
	defer sal.mutex.Unlock()
	return sal.list.AddAll(list)
}

// AddAt adds an element at the specified index
func (sal *SyncedArrayList[T]) AddAt(index int, elem T) error {
	sal.mutex.Lock()
	defer sal.mutex.Unlock()
	return sal.list.AddAt(index, elem)
}

// AddFirst adds an element at the beginning of the list
func (sal *SyncedArrayList[T]) AddFirst(elem T) error {
	sal.mutex.Lock()
	defer sal.mutex.Unlock()
	return sal.list.AddFirst(elem)
}

// AddLast adds an element at the end of the list
func (sal *SyncedArrayList[T]) AddLast(elem T) error {
	sal.mutex.Lock()
	defer sal.mutex.Unlock()
	return sal.list.AddLast(elem)
}

// Clear removes all elements from the list
func (sal *SyncedArrayList[T]) Clear() {
	sal.mutex.Lock()
	defer sal.mutex.Unlock()
	sal.list.Clear()
}

// Contains checks if an element is in the list
func (sal *SyncedArrayList[T]) Contains(elem T) bool {
	sal.mutex.RLock()
	defer sal.mutex.RUnlock()
	return sal.list.Contains(elem)
}

// Get returns the element at the specified index
func (sal *SyncedArrayList[T]) Get(index int) (T, error) {
	sal.mutex.RLock()
	defer sal.mutex.RUnlock()
	return sal.list.Get(index)
}

// GetFirst returns the first element in the list
func (sal *SyncedArrayList[T]) GetFirst() (T, error) {
	sal.mutex.RLock()
	defer sal.mutex.RUnlock()
	return sal.list.GetFirst()
}

// GetLast returns the last element in the list
func (sal *SyncedArrayList[T]) GetLast() (T, error) {
	sal.mutex.RLock()
	defer sal.mutex.RUnlock()
	return sal.list.GetLast()
}

// IndexOf returns the index of the specified element
func (sal *SyncedArrayList[T]) IndexOf(elem T) int {
	sal.mutex.RLock()
	defer sal.mutex.RUnlock()
	return sal.list.IndexOf(elem)
}

// Remove an element from the list
func (sal *SyncedArrayList[T]) Remove(elem T) bool {
	sal.mutex.Lock()
	defer sal.mutex.Unlock()
	return sal.list.Remove(elem)
}

// RemoveAt removes the element at the specified index
func (sal *SyncedArrayList[T]) RemoveAt(index int) (T, error) {
	sal.mutex.Lock()
	defer sal.mutex.Unlock()
	return sal.list.RemoveAt(index)
}

// RemoveFirst removes the first element from the list
func (sal *SyncedArrayList[T]) RemoveFirst() (T, error) {
	sal.mutex.Lock()
	defer sal.mutex.Unlock()
	return sal.list.RemoveFirst()
}

// RemoveLast removes the last element from the list
func (sal *SyncedArrayList[T]) RemoveLast() (T, error) {
	sal.mutex.Lock()
	defer sal.mutex.Unlock()
	return sal.list.RemoveLast()
}

// Size returns the number of elements in the list
func (sal *SyncedArrayList[T]) Size() int {
	sal.mutex.RLock()
	defer sal.mutex.RUnlock()
	return sal.list.Size()
}

// Iterator returns an Iterator for the list
func (sal *SyncedArrayList[T]) Iterator() Iterator[T] {
	return &syncArrayListIterator[T]{list: sal, index: 0}
}

// IsEmpty checks if the list is empty
func (sal *SyncedArrayList[T]) IsEmpty() bool {
	sal.mutex.RLock()
	defer sal.mutex.RUnlock()
	return sal.list.Size() == 0
}

// LastIndexOf returns the last index of the specified element
func (sal *SyncedArrayList[T]) LastIndexOf(elem T) int {
	sal.mutex.RLock()
	defer sal.mutex.RUnlock()
	return sal.list.LastIndexOf(elem)
}

// String returns a string representation of the list
func (sal *SyncedArrayList[T]) String() string {
	sal.mutex.RLock()
	defer sal.mutex.RUnlock()
	return sal.list.String()
}

type syncArrayListIterator[T any] struct {
	list  *SyncedArrayList[T]
	index int
}

// HasNext returns true if there are more elements in the collection
func (it *syncArrayListIterator[T]) HasNext() bool {
	it.list.mutex.RLock()
	defer it.list.mutex.RUnlock()
	return it.index < len(it.list.list.elements)
}

// Next returns the next element in the collection
func (it *syncArrayListIterator[T]) Next() T {
	it.list.mutex.RLock()
	defer it.list.mutex.RUnlock()
	elem := it.list.list.elements[it.index]
	it.index++
	return elem
}

// Remove removes the last element returned by the iterator from the collection
func (it *syncArrayListIterator[T]) Remove() {
	it.list.mutex.Lock()
	defer it.list.mutex.Unlock()
	it.list.list.elements = append(it.list.list.elements[:it.index], it.list.list.elements[it.index+1:]...)
	it.index--
}
