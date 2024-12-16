package collections

// type HashSet[T comparable] struct {
// 	elements map[T]struct{}
// }

// // NewHashSet creates a new HashSet
// func NewHashSet[T comparable]() *HashSet[T] {
// 	return &HashSet[T]{elements: make(map[T]struct{})}
// }

// // Add an element to the set
// func (hs *HashSet[T]) Add(elem T) {
// 	hs.elements[elem] = struct{}{}
// }

// // Clear removes all elements from the set
// func (hs *HashSet[T]) Clear() {
// 	hs.elements = make(map[T]struct{})
// }

// // Contains checks if an element is in the set
// func (hs *HashSet[T]) Contains(elem T) bool {
// 	_, ok := hs.elements[elem]
// 	return ok
// }

// // Iterator returns an Iterator for the set
// func (hs *HashSet[T]) Iterator() Iterator[T] {
// 	return NewHashSetIterator[T](hs)
// }

// // Remove an element from the set
// func (hs *HashSet[T]) Remove(elem T) bool {
// 	if hs.Contains(elem) {
// 		delete(hs.elements, elem)
// 		return true
// 	}
// 	return false
// }

// // Size returns the number of elements in the set
// func (hs *HashSet[T]) Size() int {
// 	return len(hs.elements)
// }

// // AddAll adds all elements from another set to this set
// func (hs *HashSet[T]) AddAll(set *HashSet[T]) {
// 	for elem := range set.elements {
// 		hs.Add(elem)
// 	}
// }

// // Union returns a new set containing all elements from this set and another set
// func (hs *HashSet[T]) Union(set *HashSet[T]) *HashSet[T] {
// 	result := NewHashSet[T]()
// 	result.AddAll(hs)
// 	result.AddAll(set)
// 	return result
// }

// // Intersection returns a new set containing only elements that are present in both this set and another set
// func (hs *HashSet[T]) Intersection(set *HashSet[T]) *HashSet[T] {
// 	result := NewHashSet[T]()
// 	for elem := range hs.elements {
// 		if set.Contains(elem) {
// 			result.Add(elem)
// 		}
// 	}
// 	return result
// }

// // Difference returns a new set containing only elements that are present in this set but not in another set
// func (hs *HashSet[T]) Difference(set *HashSet[T]) *HashSet[T] {
// 	result := NewHashSet[T]()
// 	for elem := range hs.elements {
// 		if !set.Contains(elem) {
// 			result.Add(elem)
// 		}
// 	}
// 	return result
// }

// // SymmetricDifference returns a new set containing only elements that are present in either this set or another set, but not in both
// func (hs *HashSet[T]) SymmetricDifference(set *HashSet[T]) *HashSet[T] {
// 	result := NewHashSet[T]()
// 	for elem := range hs.elements {
// 		if !set.Contains(elem) {
// 			result.Add(elem)
// 		}
// 	}
// 	for elem := range set.elements {
// 		if !hs.Contains(elem) {
// 			result.Add(elem)
// 		}
// 	}
// 	return result
// }

// // IsSubsetOf checks if this set is a subset of another set
// func (hs *HashSet[T]) IsSubsetOf(set *HashSet[T]) bool {
// 	for elem := range hs.elements {
// 		if !set.Contains(elem) {
// 			return false
// 		}
// 	}
// 	return true
// }

// // IsSupersetOf checks if this set is a superset of another set
// func (hs *HashSet[T]) IsSupersetOf(set *HashSet[T]) bool {
// 	return set.IsSubsetOf(hs)
// }

// // IsProperSubsetOf checks if this set is a proper subset of another set
// func (hs *HashSet[T]) IsProperSubsetOf(set *HashSet[T]) bool {
// 	return hs.Size() < set.Size() && hs.IsSubsetOf(set)
// }

// // IsProperSupersetOf checks if this set is a proper superset of another set
// func (hs *HashSet[T]) IsProperSupersetOf(set *HashSet[T]) bool {
// 	return hs.Size() > set.Size() && hs.IsSupersetOf(set)
// }

// // IsDisjointWith checks if this set has no elements in common with another set
// func (hs *HashSet[T]) IsDisjointWith(set *HashSet[T]) bool {
// 	for elem := range hs.elements {
// 		if set.Contains(elem) {
// 			return false
// 		}
// 	}
// 	return true
// }

// // Equal checks if this set is equal to another set
// func (hs *HashSet[T]) Equal(set *HashSet[T]) bool {
// 	return hs.IsSubsetOf(set) && hs.IsSupersetOf(set)
// }

// // Clone returns a shallow copy of this set
// func (hs *HashSet[T]) Clone() *HashSet[T] {
// 	result := NewHashSet[T]()
// 	result.AddAll(hs)
// 	return result
// }

// type HashSetIterator[T comparable] struct {
// 	elements map[T]struct{}
// 	index    int
// }

// // NewHashSetIterator creates a new HashSetIterator
// func NewHashSetIterator[T comparable](set *HashSet[T]) *HashSetIterator[T] {
// 	return &HashSetIterator[T]{elements: set.elements}
// }

// // HasNext returns true if there are more elements in the collection
// func (it *HashSetIterator[T]) HasNext() bool {
// 	return it.index < len(it.elements)
// }

// // Next returns the next element in the collection
// func (it *HashSetIterator[T]) Next() T {
// 	for elem := range it.elements {
// 		if it.index == 0 {
// 			it.index++
// 			return elem
// 		}
// 		it.index++
// 	}
// 	return nil
// }
