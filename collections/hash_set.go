package collections

import (
	"fmt"
	"sync"
)

// Byte value to be used as a placeholder in the map

// HashSet is a an implementation of a Set interface using a hash map.
type HashSet[T comparable] struct {
	hashMap map[T]any
}

// NewHashSet creates a new HashSet.
func NewHashSet[T comparable]() *HashSet[T] {
	return &HashSet[T]{hashMap: make(map[T]any)}
}

// Add an element to the set.
func (hs *HashSet[T]) Add(elem T) error {
	hs.hashMap[elem] = nil
	return nil
}

// AddAll adds all elements from another set to this set.
func (hs *HashSet[T]) AddAll(set Collection[T]) error {
	it := set.Iterator()
	for it.HasNext() {
		if err := hs.Add(it.Next()); err != nil {
			return err
		}
	}
	return nil
}

// Clear removes all elements from the set.
func (hs *HashSet[T]) Clear() {
	hs.hashMap = make(map[T]any)
}

// Contains checks if the set contains an element.
func (hs *HashSet[T]) Contains(elem T) bool {
	_, ok := hs.hashMap[elem]
	return ok
}

// Remove
func (hs *HashSet[T]) Remove(elem T) bool {
	_, ok := hs.hashMap[elem]
	if ok {
		delete(hs.hashMap, elem)
		return true
	} else {
		return false
	}
}

// Size returns the number of elements in the set.
func (hs *HashSet[T]) Size() int {
	return len(hs.hashMap)
}

// ContainsAll checks if the set contains all elements from another set.
func (hs *HashSet[T]) ContainsAll(set Set[T]) bool {
	it := set.Iterator()
	for it.HasNext() {
		if !hs.Contains(it.Next()) {
			return false
		}
	}
	return true
}

// IsEmpty checks if the set is empty.
func (hs *HashSet[T]) IsEmpty() bool {
	return len(hs.hashMap) == 0
}

// Iterator returns an iterator over the elements in the set.
func (hs *HashSet[T]) Iterator() Iterator[T] {
	keys := make([]T, 0, len(hs.hashMap))
	for key := range hs.hashMap {
		keys = append(keys, key)
	}
	return &hashSetIterator[T]{keys: keys, hashMap: hs}
}

// String returns a string representation of the set.
func (hs *HashSet[T]) String() string {
	str := "{"
	it := hs.Iterator()
	for it.HasNext() {
		str += fmt.Sprintf("%v", it.Next())
		if it.HasNext() {
			str += ", "
		}
	}
	str += "}"
	return str
}

// Union returns a new set containing all elements from this set and another set.
func (hs *HashSet[T]) Union(set Set[T]) Set[T] {
	union := NewHashSet[T]()
	_ = union.AddAll(hs)
	_ = union.AddAll(set)
	return union
}

// Intersection returns a new set containing only the elements that are in both this set and another set.
func (hs *HashSet[T]) Intersection(set Set[T]) Set[T] {
	intersection := NewHashSet[T]()
	it := hs.Iterator()
	for it.HasNext() {
		elem := it.Next()
		if set.Contains(elem) {
			_ = intersection.Add(elem)
		}
	}
	return intersection
}

// Difference returns a new set containing only the elements that are in this set but not in another set.
func (hs *HashSet[T]) Difference(set Set[T]) Set[T] {
	difference := NewHashSet[T]()
	it := hs.Iterator()
	for it.HasNext() {
		elem := it.Next()
		if !set.Contains(elem) {
			_ = difference.Add(elem)
		}
	}
	return difference
}

// IsSubset checks if this set is a subset of another set.
func (hs *HashSet[T]) IsSubset(set Set[T]) bool {
	it := hs.Iterator()
	for it.HasNext() {
		if !set.Contains(it.Next()) {
			return false
		}
	}
	return true
}

// IsSuperset checks if this set is a superset of another set.
func (hs *HashSet[T]) IsSuperset(set Set[T]) bool {
	return set.IsSubset(hs)
}

// IsProperSubset checks if this set is a proper subset of another set.
func (hs *HashSet[T]) IsProperSubset(set Set[T]) bool {
	return hs.IsSubset(set) && !hs.IsSuperset(set)
}

// IsProperSuperset checks if this set is a proper superset of another set.
func (hs *HashSet[T]) IsProperSuperset(set Set[T]) bool {
	return hs.IsSuperset(set) && !hs.IsSubset(set)
}

// IsDisjoint checks if this set has no elements in common with another set.
func (hs *HashSet[T]) IsDisjoint(set Set[T]) bool {
	it := hs.Iterator()
	for it.HasNext() {
		if set.Contains(it.Next()) {
			return false
		}
	}
	return true
}

// SymmetricDifference returns a new set containing only the elements that are in either this set or another set, but not in both.
func (hs *HashSet[T]) SymmetricDifference(set Set[T]) Set[T] {
	union := hs.Union(set)
	intersection := hs.Intersection(set)
	return union.Difference(intersection)
}

type hashSetIterator[T comparable] struct {
	keys    []T
	index   int
	hashMap *HashSet[T]
}

// HasNext returns true if the iteration has more elements.
func (hsi *hashSetIterator[T]) HasNext() bool {
	return hsi.index < len(hsi.keys)
}

// Next returns the next element in the iteration.
func (hsi *hashSetIterator[T]) Next() T {
	elem := hsi.keys[hsi.index]
	hsi.index++
	return elem
}

// Remove removes the last element returned by the iterator.
func (hsi *hashSetIterator[T]) Remove() {
	if (hsi.index >= len(hsi.keys)) || (hsi.index < 0) {
		return
	} else if hsi.index < len(hsi.keys)-1 {
		copy(hsi.keys[hsi.index:], hsi.keys[hsi.index+1:])

	}
	hsi.index--

	hsi.keys = append(hsi.keys[:hsi.index], hsi.keys[hsi.index+1:]...)
	hsi.hashMap.Remove(hsi.keys[hsi.index])
}

type SyncSet[T comparable] struct {
	// set is the underlying set
	set Set[T]
	// mutex is used to synchronize access to the underlying set
	mutex sync.RWMutex
}

// NewSyncSet creates a new synchronized set.
func NewSyncSet[T comparable]() Set[T] {
	return &SyncSet[T]{set: NewHashSet[T]()}
}

// AsSyncSet wraps a set with a mutex to create a synchronized set.
func AsSyncSet[T comparable](set Set[T]) Set[T] {
	return &SyncSet[T]{set: set}
}

// Add adds an element to the set.
func (ss *SyncSet[T]) Add(elem T) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	return ss.set.Add(elem)
}

// AddAll adds all elements from another Colelction to this set.
func (ss *SyncSet[T]) AddAll(set Collection[T]) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	return ss.set.AddAll(set)
}

// Clear removes all elements from the set.
func (ss *SyncSet[T]) Clear() {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	ss.set.Clear()
}

// Contains checks if the set contains an element.
func (ss *SyncSet[T]) Contains(elem T) bool {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	return ss.set.Contains(elem)
}

// Remove removes an element from the set.
func (ss *SyncSet[T]) Remove(elem T) bool {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	return ss.set.Remove(elem)
}

// Size returns the number of elements in the set.
func (ss *SyncSet[T]) Size() int {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	return ss.set.Size()
}

// IsEmpty checks if the set is empty.
func (ss *SyncSet[T]) IsEmpty() bool {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	return ss.set.IsEmpty()
}

// Iterator returns an iterator over the elements in the set.
func (ss *SyncSet[T]) Iterator() Iterator[T] {

	return &syncHashSetIterator[T]{iterator: ss.set.Iterator(), mutex: &ss.mutex}
}

// String returns a string representation of the set.
func (ss *SyncSet[T]) String() string {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	return ss.set.String()
}

// Union returns a new set containing all elements from this set and another set.
func (ss *SyncSet[T]) Union(set Set[T]) Set[T] {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	return ss.set.Union(set)
}

// Intersection returns a new set containing only the elements that are in both this set and another set.
func (ss *SyncSet[T]) Intersection(set Set[T]) Set[T] {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	return ss.set.Intersection(set)
}

// Difference returns a new set containing only the elements that are in this set but not in another set.
func (ss *SyncSet[T]) Difference(set Set[T]) Set[T] {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	return ss.set.Difference(set)
}

// IsSubset checks if this set is a subset of another set.
func (ss *SyncSet[T]) IsSubset(set Set[T]) bool {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	return ss.set.IsSubset(set)
}

// IsSuperset checks if this set is a superset of another set.
func (ss *SyncSet[T]) IsSuperset(set Set[T]) bool {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	return ss.set.IsSuperset(set)
}

// IsProperSubset checks if this set is a proper subset of another set.
func (ss *SyncSet[T]) IsProperSubset(set Set[T]) bool {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	return ss.set.IsProperSubset(set)
}

// IsProperSuperset checks if this set is a proper superset of another set.
func (ss *SyncSet[T]) IsProperSuperset(set Set[T]) bool {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	return ss.set.IsProperSuperset(set)
}

// IsDisjoint checks if this set has no elements in common with another set.
func (ss *SyncSet[T]) IsDisjoint(set Set[T]) bool {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	return ss.set.IsDisjoint(set)
}

// SymmetricDifference returns a new set containing only the elements that are in either this set or another set, but not in both.
func (ss *SyncSet[T]) SymmetricDifference(set Set[T]) Set[T] {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	return ss.set.SymmetricDifference(set)
}

// syncHashSetIterator is an iterator for a synchronized set.
type syncHashSetIterator[T comparable] struct {
	iterator Iterator[T]
	mutex    *sync.RWMutex
}

// HasNext returns true if the iteration has more elements.
func (si *syncHashSetIterator[T]) HasNext() bool {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	return si.iterator.HasNext()
}

// Next returns the next element in the iteration.
func (si *syncHashSetIterator[T]) Next() T {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	return si.iterator.Next()
}

// Remove removes the last element returned by the iterator.
func (si *syncHashSetIterator[T]) Remove() {
	si.mutex.Lock()
	defer si.mutex.Unlock()
	si.iterator.Remove()
}
