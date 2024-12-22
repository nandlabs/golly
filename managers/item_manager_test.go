package managers

import (
	"testing"

	"oss.nandlabs.io/golly/testing/assert"
)

func TestItemManager_Get(t *testing.T) {
	manager := NewItemManager[int]()
	manager.Register("item1", 1)
	manager.Register("item2", 2)

	item := manager.Get("item1")
	assert.Equal(t, 1, item)

	item = manager.Get("item2")
	assert.Equal(t, 2, item)

	item = manager.Get("item3")
	assert.Equal(t, 0, item) // Assuming zero value for int type
}

func TestItemManager_Items(t *testing.T) {
	manager := NewItemManager[int]()
	manager.Register("item1", 1)
	manager.Register("item2", 2)
	manager.Register("item3", 3)

	items := manager.Items()
	expectedItems := []int{1, 2, 3}

	assert.Equal(t, expectedItems, items)
}

func TestItemManager_Items_Empty(t *testing.T) {
	manager := NewItemManager[int]()

	items := manager.Items()
	expectedItems := []int{}

	assert.ElementsMatch(t, items, expectedItems...)
}

func TestItemManager_Items_AfterUnregister(t *testing.T) {
	manager := NewItemManager[int]()
	manager.Register("item1", 1)
	manager.Register("item2", 2)
	manager.Register("item3", 3)

	manager.Unregister("item2")

	items := manager.Items()
	expectedItems := []int{1, 3}

	assert.ElementsMatch(t, items, expectedItems...)
}
