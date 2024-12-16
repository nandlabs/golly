// Package collections provides a collection of generic data structures.

package collections

// List is a generic interface that defines a collection of elements with various methods to manipulate them.
// The List interface uses a type parameter T to represent the type of elements stored in the list.
// It provides the following methods:
//
// Add(elem T):
//   Adds an element to the list.
//
// Clear():
//   Removes all elements from the list.
//
// Contains(elem T) bool:
//   Checks if an element is in the list. Returns true if the element is found, otherwise false.
//
// Get(index int) T:
//   Returns the element at the specified index. The index is zero-based.
//
// Iterator() Iterator[T]:
//   Returns an Iterator for the list, which can be used to traverse the elements.
//
// IndexOf(elem T) int:
//   Returns the index of the specified element. If the element is not found, it returns -1.
//
// Remove(elem T) bool:
//   Removes an element from the list. Returns true if the element was successfully removed, otherwise false.
//
// Size() int:
//   Returns the number of elements in the list.

type List[T any] interface {
	Collection[T]
	// AddAt adds an element at the specified index
	AddAt(index int, elem T) error
	// Get returns the element at the specified index
	Get(index int) (T, error)
	// GetFirst returns the first element in the list
	GetFirst() (T, error)
	// GetLast returns the last element in the list
	GetLast() (T, error)
	// IndexOf returns the index of the specified element
	IndexOf(elem T) int
	// IsEmpty checks if the list is empty
	IsEmpty() bool
	// LastIndexOf returns the last index of the specified element
	LastIndexOf(elem T) int
	// RemoveAt removes the element at the specified index
	RemoveAt(index int) (T, error)
	// RemoveFirst removes the first element from the list
	RemoveFirst() (T, error)
	// RemoveLast removes the last element from the list
	RemoveLast() (T, error)
}
