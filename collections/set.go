package collections

// Set is a generic interface that defines a collection of unique elements with various methods to manipulate them.
// The Set interface uses a type parameter T to represent the type of elements stored in the set.
// It provides the following methods:
//

type Set[T comparable] interface {
	// Extends Collection[T]
	Collection[T]
	// Union returns a new set containing all elements from this set and another set
	Union(set Set[T]) Set[T]
	// Intersection returns a new set containing only the elements that are in both this set and another set
	Intersection(set Set[T]) Set[T]
	// Difference returns a new set containing only the elements that are in this set but not in another set
	Difference(set Set[T]) Set[T]
	// SymmetricDifference returns a new set containing only the elements that are in either this set or another set, but not in both
	SymmetricDifference(set Set[T]) Set[T]
	// IsSubset checks if this set is a subset of another set
	IsSubset(set Set[T]) bool
	// IsSuperset checks if this set is a superset of another set
	IsSuperset(set Set[T]) bool
	// IsProperSubset checks if this set is a proper subset of another set
	IsProperSubset(set Set[T]) bool
	// IsProperSuperset checks if this set is a proper superset of another set
	IsProperSuperset(set Set[T]) bool
	// IsDisjoint checks if this set has no elements in common with another set
	IsDisjoint(set Set[T]) bool
	// Remove an element from the set
	Remove(elem T) bool
	// Size returns the number of elements in the set
	Size() int
}
