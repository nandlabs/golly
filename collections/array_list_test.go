package collections

import (
	"testing"

	"oss.nandlabs.io/golly/testing/assert"
)

func TestNewArrayList(t *testing.T) {
	list := NewArrayList[int]()
	assert.NotNil(t, list)
	assert.Equal(t, 0, list.Size())
	assert.True(t, list.IsEmpty())
}
func TestArrayList_Add(t *testing.T) {
	list := NewArrayList[int]()
	err := list.Add(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, list.Size())
	assert.True(t, list.Contains(1))

	err = list.Add(2)
	assert.Nil(t, err)
	assert.Equal(t, 2, list.Size())
	assert.True(t, list.Contains(2))

	err = list.Add(3)
	assert.Nil(t, err)
	assert.Equal(t, 3, list.Size())
	assert.True(t, list.Contains(3))
}
func TestArrayList_AddAll(t *testing.T) {
	list1 := NewArrayList[int]()
	list1.Add(1)
	list1.Add(2)
	list1.Add(3)

	list2 := NewArrayList[int]()
	list2.Add(4)
	list2.Add(5)

	err := list1.AddAll(list2)
	assert.Nil(t, err)
	assert.Equal(t, 5, list1.Size())
	assert.True(t, list1.Contains(4))
	assert.True(t, list1.Contains(5))
}

func TestSyncedArrayList_AddAll(t *testing.T) {
	list1 := NewSyncedArrayList[int]()
	list1.Add(1)
	list1.Add(2)
	list1.Add(3)

	list2 := NewSyncedArrayList[int]()
	list2.Add(4)
	list2.Add(5)

	err := list1.AddAll(list2)
	assert.Nil(t, err)
	assert.Equal(t, 5, list1.Size())
	assert.True(t, list1.Contains(4))
	assert.True(t, list1.Contains(5))
}
func TestArrayList_AddAt(t *testing.T) {
	list := NewArrayList[int]()
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

func TestArrayList_AddFirst(t *testing.T) {
	list := NewArrayList[int]()
	err := list.AddFirst(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, list.Size())
	assert.True(t, list.Contains(1))

	err = list.AddFirst(2)
	assert.Nil(t, err)
	assert.Equal(t, 2, list.Size())
	assert.True(t, list.Contains(2))
	assert.Equal(t, 2, list.elements[0])
}

func TestArrayList_AddLast(t *testing.T) {
	list := NewArrayList[int]()
	err := list.AddLast(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, list.Size())
	assert.True(t, list.Contains(1))

	err = list.AddLast(2)
	assert.Nil(t, err)
	assert.Equal(t, 2, list.Size())
	assert.True(t, list.Contains(2))
	assert.Equal(t, 2, list.elements[1])
}

func TestArrayList_Clear(t *testing.T) {
	list := NewArrayList[int]()
	list.Add(1)
	list.Add(2)
	list.Clear()
	assert.Equal(t, 0, list.Size())
	assert.True(t, list.IsEmpty())
}

func TestArrayList_Get(t *testing.T) {
	list := NewArrayList[int]()
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

func TestArrayList_GetFirst(t *testing.T) {
	list := NewArrayList[int]()
	list.Add(1)
	list.Add(2)

	val, err := list.GetFirst()
	assert.Nil(t, err)
	assert.Equal(t, 1, val)

	list.Clear()
	_, err = list.GetFirst()
	assert.NotNil(t, err)
}

func TestArrayList_GetLast(t *testing.T) {
	list := NewArrayList[int]()
	list.Add(1)
	list.Add(2)

	val, err := list.GetLast()
	assert.Nil(t, err)
	assert.Equal(t, 2, val)

	list.Clear()
	_, err = list.GetLast()
	assert.NotNil(t, err)
}

func TestArrayList_IndexOf(t *testing.T) {
	list := NewArrayList[int]()
	list.Add(1)
	list.Add(2)
	list.Add(3)

	index := list.IndexOf(2)
	assert.Equal(t, 1, index)

	index = list.IndexOf(4)
	assert.Equal(t, -1, index)
}

func TestArrayList_LastIndexOf(t *testing.T) {
	list := NewArrayList[int]()
	list.Add(1)
	list.Add(2)
	list.Add(2)
	list.Add(3)

	index := list.LastIndexOf(2)
	assert.Equal(t, 2, index)

	index = list.LastIndexOf(4)
	assert.Equal(t, -1, index)
}

func TestArrayList_RemoveAt(t *testing.T) {
	list := NewArrayList[int]()
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

func TestArrayList_RemoveFirst(t *testing.T) {
	list := NewArrayList[int]()
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

func TestArrayList_RemoveLast(t *testing.T) {
	list := NewArrayList[int]()
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

func TestArrayList_Iterator(t *testing.T) {
	list := NewArrayList[int]()
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

func TestSyncedArrayList_Add(t *testing.T) {
	list := NewSyncedArrayList[int]()
	err := list.Add(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, list.Size())
	assert.True(t, list.Contains(1))

	err = list.Add(2)
	assert.Nil(t, err)
	assert.Equal(t, 2, list.Size())
	assert.True(t, list.Contains(2))

	err = list.Add(3)
	assert.Nil(t, err)
	assert.Equal(t, 3, list.Size())
	assert.True(t, list.Contains(3))
}

func TestSyncedArrayList_AddAt(t *testing.T) {
	list := NewSyncedArrayList[int]()
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

func TestSyncedArrayList_AddFirst(t *testing.T) {
	list := NewSyncedArrayList[int]()
	err := list.AddFirst(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, list.Size())
	assert.True(t, list.Contains(1))

	err = list.AddFirst(2)
	assert.Nil(t, err)
	assert.Equal(t, 2, list.Size())
	assert.True(t, list.Contains(2))
	assert.Equal(t, 2, list.list.elements[0])
}

func TestSyncedArrayList_AddLast(t *testing.T) {
	list := NewSyncedArrayList[int]()
	err := list.AddLast(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, list.Size())
	assert.True(t, list.Contains(1))

	err = list.AddLast(2)
	assert.Nil(t, err)
	assert.Equal(t, 2, list.Size())
	assert.True(t, list.Contains(2))
	assert.Equal(t, 2, list.list.elements[1])
}

func TestSyncedArrayList_Clear(t *testing.T) {
	list := NewSyncedArrayList[int]()
	list.Add(1)
	list.Add(2)
	list.Clear()
	assert.Equal(t, 0, list.Size())
	assert.True(t, list.IsEmpty())
}

func TestSyncedArrayList_Get(t *testing.T) {
	list := NewSyncedArrayList[int]()
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

func TestSyncedArrayList_GetFirst(t *testing.T) {
	list := NewSyncedArrayList[int]()
	list.Add(1)
	list.Add(2)

	val, err := list.GetFirst()
	assert.Nil(t, err)
	assert.Equal(t, 1, val)

	list.Clear()
	_, err = list.GetFirst()
	assert.NotNil(t, err)
}

func TestSyncedArrayList_GetLast(t *testing.T) {
	list := NewSyncedArrayList[int]()
	list.Add(1)
	list.Add(2)

	val, err := list.GetLast()
	assert.Nil(t, err)
	assert.Equal(t, 2, val)

	list.Clear()
	_, err = list.GetLast()
	assert.NotNil(t, err)
}

func TestSyncedArrayList_IndexOf(t *testing.T) {
	list := NewSyncedArrayList[int]()
	list.Add(1)
	list.Add(2)
	list.Add(3)

	index := list.IndexOf(2)
	assert.Equal(t, 1, index)

	index = list.IndexOf(4)
	assert.Equal(t, -1, index)
}

func TestSyncedArrayList_LastIndexOf(t *testing.T) {
	list := NewSyncedArrayList[int]()
	list.Add(1)
	list.Add(2)
	list.Add(2)
	list.Add(3)

	index := list.LastIndexOf(2)
	assert.Equal(t, 2, index)

	index = list.LastIndexOf(4)
	assert.Equal(t, -1, index)
}

func TestSyncedArrayList_RemoveAt(t *testing.T) {
	list := NewSyncedArrayList[int]()
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

func TestSyncedArrayList_RemoveFirst(t *testing.T) {
	list := NewSyncedArrayList[int]()
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

func TestSyncedArrayList_RemoveLast(t *testing.T) {
	list := NewSyncedArrayList[int]()
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

func TestSyncedArrayList_Iterator(t *testing.T) {
	list := NewSyncedArrayList[int]()
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
func TestArrayList_Remove(t *testing.T) {
	list := NewArrayList[int]()
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

func TestSyncedArrayList_Remove(t *testing.T) {
	list := NewSyncedArrayList[int]()
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
