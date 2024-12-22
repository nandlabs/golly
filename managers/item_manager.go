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
