# Managers Package

The `managers` package provides utility features for managing items in Go.

This package includes support for registering, unregistering, and retrieving items in a thread-safe manner.

## Installation

To install the package, use the following command:

```sh
go get github.com/nandlabs/golly/managers
```

## Index

- [Installation](#installation)
- [Usage](#usage)
  - [Registering Items](#registering-items)
  - [Unregistering Items](#unregistering-items)
  - [Retrieving Items](#retrieving-items)
  - [Listing All Items](#listing-all-items)
- [Components](#components)
  - [ItemManager](#itemmanager)
- [License](#license)

## Usage

### Registering Items

To register an item, use the `Register` method of the `ItemManager` interface. Here is an example:

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/managers"
)

func main() {
    manager := managers.NewItemManager[string]()
    manager.Register("item1", "This is item 1")

    item := manager.Get("item1")
    fmt.Println("Registered item:", item)
}
```

### Unregistering Items

To unregister an item, use the `Unregister` method of the `ItemManager` interface. Here is an example:

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/managers"
)

func main() {
    manager := managers.NewItemManager[string]()
    manager.Register("item1", "This is item 1")

    manager.Unregister("item1")
    item := manager.Get("item1")
    fmt.Println("Unregistered item:", item) // Output: Unregistered item:
}
```

### Retrieving Items

To retrieve an item, use the `Get` method of the `ItemManager` interface. Here is an example:

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/managers"
)

func main() {
    manager := managers.NewItemManager[string]()
    manager.Register("item1", "This is item 1")

    item := manager.Get("item1")
    fmt.Println("Retrieved item:", item)
}
```

### Listing All Items

To list all items, use the `Items` method of the `ItemManager` interface. Here is an example:

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/managers"
)

func main() {
    manager := managers.NewItemManager[string]()
    manager.Register("item1", "This is item 1")
    manager.Register("item2", "This is item 2")

    items := manager.Items()
    fmt.Println("All items:", items)
}
```

## Components

### ItemManager

The `ItemManager` interface provides methods for managing a collection of items. It includes methods for registering, unregistering, retrieving, and listing items.

```go
package managers

import "sync"

type ItemManager[T any] interface {
    Register(name string, item T)
    Unregister(name string)
    Get(name string) T
    Items() []T
}

type itemManager[T any] struct {
    items map[string]T
    mutex sync.RWMutex
}

func (it *itemManager[T]) Register(name string, item T) {
    it.mutex.Lock()
    defer it.mutex.Unlock()
    it.items[name] = item
}

func (it *itemManager[T]) Unregister(name string) {
    it.mutex.Lock()
    defer it.mutex.Unlock()
    delete(it.items, name)
}

func (it *itemManager[T]) Get(name string) T {
    it.mutex.RLock()
    defer it.mutex.RUnlock()
    item := it.items[name]

    return item
}

func (it *itemManager[T]) Items() []T {
    it.mutex.RLock()
    defer it.mutex.RUnlock()
    var items []T
    for _, item := range it.items {
        items = append(items, item)
    }

    return items
}

func NewItemManager[T any]() ItemManager[T] {
    return &itemManager[T]{
        items: make(map[string]T),
    }
}
```
