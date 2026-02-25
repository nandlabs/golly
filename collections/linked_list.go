package collections

import (
	"errors"
	"fmt"
	"sync"

	"oss.nandlabs.io/golly/assertion"
)

// LinkedList is a generic list implementation using a linked list
type LinkedList[T any] struct {
	head *node[T]
	tail *node[T]
	size int
}

// node is a single element in the linked list
type node[T any] struct {
	value T
	next  *node[T]
}

// NewLinkedList creates a new LinkedList
func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{size: 0}
}

// Add an element to the list
func (ll *LinkedList[T]) Add(elem T) error {
	newNode := &node[T]{value: elem}
	if ll.head == nil {
		ll.head = newNode
		ll.tail = newNode
	} else {
		ll.tail.next = newNode
		ll.tail = newNode
	}
	ll.size++
	return nil
}

// Add All elements from another list to this list
func (ll *LinkedList[T]) AddAll(list Collection[T]) error {
	it := list.Iterator()
	for it.HasNext() {
		if err := ll.Add(it.Next()); err != nil {
			return err
		}
	}
	return nil
}

// AddAt adds an element at the specified index
func (ll *LinkedList[T]) AddAt(index int, elem T) error {
	if index < 0 || index > ll.size {
		return errors.New("index out of range")
	}
	newNode := &node[T]{value: elem}
	if index == 0 {
		newNode.next = ll.head
		ll.head = newNode
	} else {
		prev := ll.head
		for i := 0; i < index-1; i++ {
			prev = prev.next
		}
		newNode.next = prev.next
		prev.next = newNode
	}
	ll.size++
	return nil
}

// AddFirst adds an element at the beginning of the list
func (ll *LinkedList[T]) AddFirst(elem T) error {
	return ll.AddAt(0, elem)
}

// AddLast adds an element at the end of the list
func (ll *LinkedList[T]) AddLast(elem T) error {
	return ll.Add(elem)
}

// Clear removes all elements from the list
func (ll *LinkedList[T]) Clear() {
	ll.head = nil
	ll.tail = nil
	ll.size = 0
}

// Contains checks if an element is in the list
func (ll *LinkedList[T]) Contains(elem T) bool {
	current := ll.head
	for current != nil {
		if assertion.Equal(current.value, elem) {
			return true
		}
		current = current.next
	}
	return false
}

// Get returns the element at the specified index
func (ll *LinkedList[T]) Get(index int) (v T, err error) {
	if index < 0 || index >= ll.size {
		err = errors.New("index out of range")
		return
	}
	current := ll.head
	for i := 0; i < index; i++ {
		current = current.next
	}
	v = current.value
	return
}

// GetFirst returns the first element in the list
func (ll *LinkedList[T]) GetFirst() (T, error) {
	return ll.Get(0)
}

// GetLast returns the last element in the list
func (ll *LinkedList[T]) GetLast() (T, error) {
	return ll.Get(ll.size - 1)
}

// IndexOf returns the index of the specified element
func (ll *LinkedList[T]) IndexOf(elem T) int {
	current := ll.head
	for i := 0; current != nil; i++ {
		if assertion.Equal(current.value, elem) {
			return i
		}
		current = current.next
	}
	return -1
}

// IsEmpty checks if the list is empty
func (ll *LinkedList[T]) IsEmpty() bool {
	return ll.size == 0
}

// LastIndexOf returns the last index of the specified element
func (ll *LinkedList[T]) LastIndexOf(elem T) int {
	var index = -1
	current := ll.head
	for i := 0; current != nil; i++ {
		if assertion.Equal(current.value, elem) {
			index = i
		}
		current = current.next
	}
	return index
}

// Remove an element from the list
func (ll *LinkedList[T]) Remove(elem T) bool {
	if ll.head == nil {
		return false
	}
	if assertion.Equal(ll.head.value, elem) {
		ll.head = ll.head.next
		ll.size--
		return true
	}
	prev := ll.head
	current := ll.head.next
	for current != nil {
		if assertion.Equal(current.value, elem) {
			prev.next = current.next
			ll.size--
			return true
		}
		prev = current
		current = current.next
	}
	return false
}

// RemoveAt removes the element at the specified index
func (ll *LinkedList[T]) RemoveAt(index int) (v T, err error) {
	if index < 0 || index >= ll.size {
		err = errors.New("index out of range")
		return
	}
	var value T
	if index == 0 {
		value = ll.head.value
		ll.head = ll.head.next
		ll.size--
		return value, nil
	}
	prev := ll.head
	for i := 0; i < index-1; i++ {
		prev = prev.next
	}
	v = prev.next.value
	prev.next = prev.next.next
	ll.size--

	return
}

// RemoveFirst removes the first element from the list
func (ll *LinkedList[T]) RemoveFirst() (T, error) {
	return ll.RemoveAt(0)
}

// RemoveLast removes the last element from the list
func (ll *LinkedList[T]) RemoveLast() (T, error) {
	return ll.RemoveAt(ll.size - 1)
}

// Size returns the number of elements in the list
func (ll *LinkedList[T]) Size() int {
	return ll.size
}

// Iterator returns an Iterator for the list
func (ll *LinkedList[T]) Iterator() Iterator[T] {
	return &linkedListIterator[T]{list: ll, current: ll.head}

}

// linkedListIterator is an iterator for a linked list
type linkedListIterator[T any] struct {
	list    *LinkedList[T]
	current *node[T]
}

// HasNext returns true if there are more elements in the collection
func (li *linkedListIterator[T]) HasNext() bool {
	return li.current != nil
}

// Next returns the next element in the collection
func (li *linkedListIterator[T]) Next() (v T) {
	if li.current == nil {
		return
	}
	v = li.current.value
	li.current = li.current.next
	return
}

// Remove removes the last element returned by the iterator from the collection
func (li *linkedListIterator[T]) Remove() {
	li.list.Remove(li.current.value)
}

// String returns a string representation of the list
func (ll *LinkedList[T]) String() string {
	var str string
	current := ll.head
	for current != nil {
		str += fmt.Sprintf("%v ", current.value)
		current = current.next
	}
	return str
}

type SyncedLinkedList[T any] struct {
	list  *LinkedList[T]
	mutex sync.RWMutex
}

// NewSyncedLinkedList creates a new SyncedLinkedList
func NewSyncedLinkedList[T any]() *SyncedLinkedList[T] {
	return &SyncedLinkedList[T]{list: NewLinkedList[T]()}
}

// Add an element to the list
func (sll *SyncedLinkedList[T]) Add(elem T) {
	sll.mutex.Lock()
	defer sll.mutex.Unlock()
	_ = sll.list.Add(elem)
}

// AddAll adds all elements from another list to this list
func (sll *SyncedLinkedList[T]) AddAll(list Collection[T]) error {
	sll.mutex.Lock()
	defer sll.mutex.Unlock()
	return sll.list.AddAll(list)
}

// AddAt adds an element at the specified index
func (sll *SyncedLinkedList[T]) AddAt(index int, elem T) error {
	sll.mutex.Lock()
	defer sll.mutex.Unlock()
	return sll.list.AddAt(index, elem)

}

// AddFirst adds an element at the beginning of the list
func (sll *SyncedLinkedList[T]) AddFirst(elem T) error {
	sll.mutex.Lock()
	defer sll.mutex.Unlock()
	return sll.list.AddFirst(elem)
}

// AddLast adds an element at the end of the list
func (sll *SyncedLinkedList[T]) AddLast(elem T) error {
	sll.mutex.Lock()
	defer sll.mutex.Unlock()
	return sll.list.AddLast(elem)
}

// Clear removes all elements from the list
func (sll *SyncedLinkedList[T]) Clear() {
	sll.mutex.Lock()
	defer sll.mutex.Unlock()
	sll.list.Clear()
}

// Contains checks if an element is in the list
func (sll *SyncedLinkedList[T]) Contains(elem T) bool {
	sll.mutex.RLock()
	defer sll.mutex.RUnlock()
	return sll.list.Contains(elem)
}

// Get returns the element at the specified index
func (sll *SyncedLinkedList[T]) Get(index int) (T, error) {
	sll.mutex.RLock()
	defer sll.mutex.RUnlock()
	return sll.list.Get(index)
}

// GetFirst returns the first element in the list
func (sll *SyncedLinkedList[T]) GetFirst() (T, error) {
	sll.mutex.RLock()
	defer sll.mutex.RUnlock()
	return sll.list.GetFirst()
}

// GetLast returns the last element in the list
func (sll *SyncedLinkedList[T]) GetLast() (T, error) {
	sll.mutex.RLock()
	defer sll.mutex.RUnlock()
	return sll.list.GetLast()
}

// IndexOf returns the index of the specified element
func (sll *SyncedLinkedList[T]) IndexOf(elem T) int {
	sll.mutex.RLock()
	defer sll.mutex.RUnlock()
	return sll.list.IndexOf(elem)
}

// Remove an element from the list
func (sll *SyncedLinkedList[T]) Remove(elem T) bool {
	sll.mutex.Lock()
	defer sll.mutex.Unlock()
	return sll.list.Remove(elem)
}

// RemoveAt removes the element at the specified index
func (sll *SyncedLinkedList[T]) RemoveAt(index int) (T, error) {
	sll.mutex.Lock()
	defer sll.mutex.Unlock()
	return sll.list.RemoveAt(index)
}

// RemoveFirst removes the first element from the list
func (sll *SyncedLinkedList[T]) RemoveFirst() (T, error) {
	sll.mutex.Lock()
	defer sll.mutex.Unlock()
	return sll.list.RemoveFirst()
}

// RemoveLast removes the last element from the list
func (sll *SyncedLinkedList[T]) RemoveLast() (T, error) {
	sll.mutex.Lock()
	defer sll.mutex.Unlock()
	return sll.list.RemoveLast()
}

// Size returns the number of elements in the list
func (sll *SyncedLinkedList[T]) Size() int {
	sll.mutex.RLock()
	defer sll.mutex.RUnlock()
	return sll.list.Size()
}

// Iterator returns an Iterator for the list
func (sll *SyncedLinkedList[T]) Iterator() Iterator[T] {

	return &syncedLinkedListIterator[T]{list: sll, index: 0}
}

// IsEmpty checks if the list is empty
func (sll *SyncedLinkedList[T]) IsEmpty() bool {
	sll.mutex.RLock()
	defer sll.mutex.RUnlock()
	return sll.list.IsEmpty()
}

// LastIndexOf returns the last index of the specified element
func (sll *SyncedLinkedList[T]) LastIndexOf(elem T) int {
	sll.mutex.RLock()
	defer sll.mutex.RUnlock()
	return sll.list.LastIndexOf(elem)
}

// String returns a string representation of the list
func (sll *SyncedLinkedList[T]) String() string {
	sll.mutex.RLock()
	defer sll.mutex.RUnlock()
	return sll.list.String()
}

type syncedLinkedListIterator[T any] struct {
	list  *SyncedLinkedList[T]
	index int
}

// HasNext returns true if there are more elements in the collection
func (sli *syncedLinkedListIterator[T]) HasNext() bool {
	sli.list.mutex.RLock()
	defer sli.list.mutex.RUnlock()
	return sli.index < sli.list.Size()
}

// Next returns the next element in the collection
func (sli *syncedLinkedListIterator[T]) Next() (v T) {
	sli.list.mutex.RLock()
	defer sli.list.mutex.RUnlock()
	v, _ = sli.list.Get(sli.index)
	sli.index++
	return
}

// Remove removes the last element returned by the iterator from the collection
func (sli *syncedLinkedListIterator[T]) Remove() {
	sli.list.mutex.Lock()
	defer sli.list.mutex.Unlock()
	sli.index--
	_, _ = sli.list.RemoveAt(sli.index)
}
