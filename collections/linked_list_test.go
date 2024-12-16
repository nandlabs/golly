package collections

import (
	"testing"

	"oss.nandlabs.io/golly/testing/assert"
)

func TestNewLinkedList(t *testing.T) {
	list := NewLinkedList[int]()
	assert.NotNil(t, list)
	assert.Equal(t, 0, list.Size())
	assert.True(t, list.IsEmpty())
}

func TestLinkedList_Add(t *testing.T) {
	list := NewLinkedList[int]()
	list.Add(1)
	assert.Equal(t, 1, list.Size())
	assert.True(t, list.Contains(1))

	list.Add(2)
	assert.Equal(t, 2, list.Size())
	assert.True(t, list.Contains(2))

	list.Add(3)
	assert.Equal(t, 3, list.Size())
	assert.True(t, list.Contains(3))
}

func TestLinkedList_AddAll(t *testing.T) {
	list1 := NewLinkedList[int]()
	list1.Add(1)
	list1.Add(2)
	list1.Add(3)

	list2 := NewLinkedList[int]()
	list2.Add(4)
	list2.Add(5)

	list1.AddAll(list2)
	assert.Equal(t, 5, list1.Size())
	assert.True(t, list1.Contains(4))
	assert.True(t, list1.Contains(5))
}

func TestLinkedList_AddAt(t *testing.T) {
	list := NewLinkedList[int]()
	err := list.AddAt(0, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, list.Size())
	assert.True(t, list.Contains(1))

	err = list.AddAt(1, 2)
	assert.Nil(t, err)
	assert.Equal(t, 2, list.Size())
	assert.True(t, list.Contains(2))

	err = list.AddAt(1, 3)
	assert.Nil(t, err)
	assert.Equal(t, 3, list.Size())
	assert.True(t, list.Contains(3))

	err = list.AddAt(5, 4)
	assert.NotNil(t, err)
	assert.Equal(t, 3, list.Size())
}

func TestLinkedList_AddFirst(t *testing.T) {
	list := NewLinkedList[int]()
	list.AddFirst(1)
	assert.Equal(t, 1, list.Size())
	assert.True(t, list.Contains(1))

	list.AddFirst(2)
	assert.Equal(t, 2, list.Size())
	assert.True(t, list.Contains(2))
	assert.Equal(t, 2, list.head.value)
}

func TestLinkedList_AddLast(t *testing.T) {
	list := NewLinkedList[int]()
	list.AddLast(1)
	assert.Equal(t, 1, list.Size())
	assert.True(t, list.Contains(1))

	list.AddLast(2)
	assert.Equal(t, 2, list.Size())
	assert.True(t, list.Contains(2))
	assert.Equal(t, 2, list.tail.value)
}

func TestLinkedList_Clear(t *testing.T) {
	list := NewLinkedList[int]()
	list.Add(1)
	list.Add(2)
	list.Clear()
	assert.Equal(t, 0, list.Size())
	assert.True(t, list.IsEmpty())
}

func TestLinkedList_Contains(t *testing.T) {
	list := NewLinkedList[int]()
	list.Add(1)
	list.Add(2)
	assert.True(t, list.Contains(1))
	assert.True(t, list.Contains(2))
	assert.False(t, list.Contains(3))
}

func TestLinkedList_Get(t *testing.T) {
	list := NewLinkedList[int]()
	list.Add(1)
	list.Add(2)

	val, err := list.Get(0)
	assert.Nil(t, err)
	assert.Equal(t, 1, val)

	val, err = list.Get(1)
	assert.Nil(t, err)
	assert.Equal(t, 2, val)

	_, err = list.Get(2)
	assert.NotNil(t, err)
}

func TestLinkedList_GetFirst(t *testing.T) {
	list := NewLinkedList[int]()
	list.Add(1)
	list.Add(2)

	val, err := list.GetFirst()
	assert.Nil(t, err)
	assert.Equal(t, 1, val)

	list.Clear()
	_, err = list.GetFirst()
	assert.NotNil(t, err)
}

func TestLinkedList_GetLast(t *testing.T) {
	list := NewLinkedList[int]()
	list.Add(1)
	list.Add(2)

	val, err := list.GetLast()
	assert.Nil(t, err)
	assert.Equal(t, 2, val)

	list.Clear()
	_, err = list.GetLast()
	assert.NotNil(t, err)
}

func TestLinkedList_IndexOf(t *testing.T) {
	list := NewLinkedList[int]()
	list.Add(1)
	list.Add(2)
	list.Add(3)

	index := list.IndexOf(2)
	assert.Equal(t, 1, index)

	index = list.IndexOf(4)
	assert.Equal(t, -1, index)
}

func TestLinkedList_LastIndexOf(t *testing.T) {
	list := NewLinkedList[int]()
	list.Add(1)
	list.Add(2)
	list.Add(2)
	list.Add(3)

	index := list.LastIndexOf(2)
	assert.Equal(t, 2, index)

	index = list.LastIndexOf(4)
	assert.Equal(t, -1, index)
}

func TestLinkedList_Remove(t *testing.T) {
	list := NewLinkedList[int]()
	list.Add(1)
	list.Add(2)
	list.Add(3)

	removed := list.Remove(2)
	assert.True(t, removed)
	assert.Equal(t, 2, list.Size())
	assert.False(t, list.Contains(2))

	removed = list.Remove(4)
	assert.False(t, removed)
	assert.Equal(t, 2, list.Size())
}

func TestLinkedList_RemoveAt(t *testing.T) {
	list := NewLinkedList[int]()
	list.Add(1)
	list.Add(2)
	list.Add(3)

	val, err := list.RemoveAt(1)
	assert.Nil(t, err)
	assert.Equal(t, 2, val)
	assert.Equal(t, 2, list.Size())
	assert.False(t, list.Contains(2))

	_, err = list.RemoveAt(5)
	assert.NotNil(t, err)
}

func TestLinkedList_RemoveFirst(t *testing.T) {
	list := NewLinkedList[int]()
	list.Add(1)
	list.Add(2)
	list.Add(3)

	val, err := list.RemoveFirst()
	assert.Nil(t, err)
	assert.Equal(t, 1, val)
	assert.Equal(t, 2, list.Size())
	assert.False(t, list.Contains(1))

	list.Clear()
	_, err = list.RemoveFirst()
	assert.NotNil(t, err)
}

func TestLinkedList_RemoveLast(t *testing.T) {
	list := NewLinkedList[int]()
	list.Add(1)
	list.Add(2)
	list.Add(3)

	val, err := list.RemoveLast()
	assert.Nil(t, err)
	assert.Equal(t, 3, val)
	assert.Equal(t, 2, list.Size())
	assert.False(t, list.Contains(3))

	list.Clear()
	_, err = list.RemoveLast()
	assert.NotNil(t, err)
}

func TestLinkedList_Iterator(t *testing.T) {
	list := NewLinkedList[int]()
	list.Add(1)
	list.Add(2)
	list.Add(3)

	it := list.Iterator()
	assert.True(t, it.HasNext())
	assert.Equal(t, 1, it.Next())
	assert.True(t, it.HasNext())
	assert.Equal(t, 2, it.Next())
	assert.True(t, it.HasNext())
	assert.Equal(t, 3, it.Next())
	assert.False(t, it.HasNext())
}
