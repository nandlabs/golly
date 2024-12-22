# Collections

Collections is a package that provides a collection of generic data structures, including Stack, Queue, List, Set, and their synchronized versions. This package is designed to be type-safe and easy to use with Go's generics.

---

- [Installation](#installation)
- [Collections Interface](#collections-interface)
  - [Methods](#methods)
- [Supported Collections](#supported-collections)
  - [Stack](#stack)
    - [Basic Stack](#basic-stack)
    - [Synchronized Stack](#synchronized-stack)
  - [Queue](#queue)
    - [Basic Queue](#basic-queue)
    - [Synchronized Queue](#synchronized-queue)
  - [List](#list)
    - [Basic List](#basic-list)
    - [Synchronized List](#synchronized-list)
  - [Set](#set)
    - [Basic Set](#basic-set)
    - [Synchronized Set](#synchronized-set)
- [Implementations](#implementations)
  - [ArrayList](#arraylist)
  - [LinkedList](#linkedlist)
  - [HashSet](#hashset)
  - [Synchronized Versions](#synchronized-versions)

---

## Installation

To install the package, use the following command:

```sh
go get github.com/nandlabs/golly/collections
```

## Collections Interface

The `Collection` interface is a generic interface that defines a collection of elements with various methods to manipulate them. It uses a type parameter `T` to represent the type of elements stored in the collection.

### Methods

- `Add(elem T) error`: Adds an element to the collection.
- `AddAll(coll Collection[T]) error`: Adds all elements from another collection to this collection.
- `Clear()`: Removes all elements from the collection.
- `Contains(elem T) bool`: Checks if an element is in the collection.
- `IsEmpty() bool`: Returns true if the collection is empty.
- `Remove(elem T) bool`: Removes an element from the collection.
- `Size() int`: Returns the number of elements in the collection.

## Supported Collections

### Stack

A stack is a LIFO (Last In, First Out) data structure. The package provides both a basic and a synchronized stack implementation.

#### Basic Stack

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/collections"
)

func main() {
    stack := collections.NewStack[int]()
    stack.Push(1)
    stack.Push(2)
    stack.Push(3)

    fmt.Println(stack.Pop()) // Output: 3
    fmt.Println(stack.Peek()) // Output: 2
}
```

#### Synchronized Stack

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/collections"
)

func main() {
    stack := collections.NewSyncStack[int]()
    stack.Push(1)
    stack.Push(2)
    stack.Push(3)

    fmt.Println(stack.Pop()) // Output: 3
    fmt.Println(stack.Peek()) // Output: 2
}
```

### Queue

A queue is a FIFO (First In, First Out) data structure. The package provides both a basic and a synchronized queue implementation.

#### Basic Queue

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/collections"
)

func main() {
    queue := collections.NewArrayQueue[int]()
    queue.Enqueue(1)
    queue.Enqueue(2)
    queue.Enqueue(3)

    fmt.Println(queue.Dequeue()) // Output: 1
    fmt.Println(queue.Front()) // Output: 2
}
```

#### Synchronized Queue

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/collections"
)

func main() {
    queue := collections.NewSyncQueue[int]()
    queue.Enqueue(1)
    queue.Enqueue(2)
    queue.Enqueue(3)

    fmt.Println(queue.Dequeue()) // Output: 1
    fmt.Println(queue.Front()) // Output: 2
}
```

### List

A list is a collection of elements that can be accessed by index. The package provides both a basic and a synchronized list implementation.

#### Basic List

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/collections"
)

func main() {
    list := collections.NewArrayList[int]()
    list.Add(1)
    list.Add(2)
    list.Add(3)

    fmt.Println(list.Get(0)) // Output: 1
    fmt.Println(list.GetLast()) // Output: 3
}
```

#### Synchronized List

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/collections"
)

func main() {
    list := collections.NewSyncedArrayList[int]()
    list.Add(1)
    list.Add(2)
    list.Add(3)

    fmt.Println(list.Get(0)) // Output: 1
    fmt.Println(list.GetLast()) // Output: 3
}
```

### Set

A set is a collection of unique elements. The package provides both a basic and a synchronized set implementation.

#### Basic Set

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/collections"
)

func main() {
    set := collections.NewHashSet[int]()
    set.Add(1)
    set.Add(2)
    set.Add(3)

    fmt.Println(set.Contains(2)) // Output: true
    fmt.Println(set.Size()) // Output: 3
}
```

#### Synchronized Set

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/collections"
)

func main() {
    set := collections.NewSyncSet[int]()
    set.Add(1)
    set.Add(2)
    set.Add(3)

    fmt.Println(set.Contains(2)) // Output: true
    fmt.Println(set.Size()) // Output: 3
}
```

## Implementations

### ArrayList

The `ArrayList` is a generic list implementation using an array. It provides methods to add, remove, and access elements by index.

### LinkedList

The `LinkedList` is a generic list implementation using a linked list. It provides methods to add, remove, and access elements by index.

### HashSet

The `HashSet` is a generic set implementation using a hash map. It provides methods to add, remove, and check for the presence of elements.

### Synchronized Versions

Each data structure has a synchronized version that is thread-safe. These synchronized versions use mutexes to ensure safe concurrent access.
