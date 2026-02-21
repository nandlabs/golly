// Package main demonstrates the collections package with generic data structures.
package main

import (
	"fmt"

	"oss.nandlabs.io/golly/collections"
)

func main() {
	// --- ArrayList ---
	fmt.Println("=== ArrayList ===")
	list := collections.NewArrayList[string]()
	_ = list.Add("Alice")
	_ = list.Add("Bob")
	_ = list.Add("Charlie")
	fmt.Println("Size:", list.Size())

	val, _ := list.Get(1)
	fmt.Println("Get(1):", val)

	_ = list.Remove("Bob")
	fmt.Println("After Remove('Bob'), Size:", list.Size())

	// --- LinkedList ---
	fmt.Println("\n=== LinkedList ===")
	ll := collections.NewLinkedList[int]()
	_ = ll.Add(10)
	_ = ll.Add(20)
	_ = ll.Add(30)
	fmt.Println("Size:", ll.Size())
	first, _ := ll.Get(0)
	fmt.Println("First:", first)

	// --- HashSet ---
	fmt.Println("\n=== HashSet ===")
	set := collections.NewHashSet[string]()
	_ = set.Add("go")
	_ = set.Add("rust")
	_ = set.Add("go") // duplicate, ignored
	fmt.Println("Size:", set.Size())
	fmt.Println("Contains 'go':", set.Contains("go"))
	fmt.Println("Contains 'java':", set.Contains("java"))

	// --- Stack (LIFO) ---
	fmt.Println("\n=== Stack ===")
	stack := collections.NewStack[string]()
	stack.Push("first")
	stack.Push("second")
	stack.Push("third")
	top, _ := stack.Peek()
	fmt.Println("Peek:", top)
	popped, _ := stack.Pop()
	fmt.Println("Pop:", popped)
	fmt.Println("Size after pop:", stack.Size())

	// --- Queue (FIFO) ---
	fmt.Println("\n=== Queue ===")
	queue := collections.NewArrayQueue[int]()
	_ = queue.Enqueue(1)
	_ = queue.Enqueue(2)
	_ = queue.Enqueue(3)
	front, _ := queue.Front()
	fmt.Println("Front:", front)
	dequeued, _ := queue.Dequeue()
	fmt.Println("Dequeue:", dequeued)
	fmt.Println("Size after dequeue:", queue.Size())

	// --- Thread-safe collections ---
	fmt.Println("\n=== Thread-Safe ArrayList ===")
	syncList := collections.NewSyncedArrayList[int]()
	_ = syncList.Add(100)
	_ = syncList.Add(200)
	fmt.Println("Sync list size:", syncList.Size())
}
