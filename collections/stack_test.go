package collections

import (
	"testing"

	"oss.nandlabs.io/golly/testing/assert"
)

func TestStack_Push(t *testing.T) {
	stack := NewStack[int]()
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)

	assert.Equal(t, 3, stack.Size())
	v, e := stack.Peek()
	assert.Nil(t, e)
	assert.Equal(t, 3, v)
}

func TestStack_Pop(t *testing.T) {
	stack := NewStack[int]()
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)

	val, err := stack.Pop()
	assert.Nil(t, err)
	assert.Equal(t, 3, val)
	assert.Equal(t, 2, stack.Size())

	val, err = stack.Pop()
	assert.Nil(t, err)
	assert.Equal(t, 2, val)
	assert.Equal(t, 1, stack.Size())

	val, err = stack.Pop()
	assert.Nil(t, err)
	assert.Equal(t, 1, val)
	assert.Equal(t, 0, stack.Size())

	_, err = stack.Pop()
	assert.NotNil(t, err)
}

func TestStack_Peek(t *testing.T) {
	stack := NewStack[int]()
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)

	val, err := stack.Peek()
	assert.Nil(t, err)
	assert.Equal(t, 3, val)
	assert.Equal(t, 3, stack.Size())

	stack.Pop()
	val, err = stack.Peek()
	assert.Nil(t, err)
	assert.Equal(t, 2, val)
	assert.Equal(t, 2, stack.Size())
}

func TestStack_Iterator(t *testing.T) {
	stack := NewStack[int]()
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)

	it := stack.Iterator()
	assert.True(t, it.HasNext())
	assert.Equal(t, 3, it.Next())
	assert.True(t, it.HasNext())
	assert.Equal(t, 2, it.Next())
	assert.True(t, it.HasNext())
	assert.Equal(t, 1, it.Next())
	assert.False(t, it.HasNext())
}

func TestSyncStack_Push(t *testing.T) {
	stack := NewSyncStack[int]()
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)

	assert.Equal(t, 3, stack.Size())
	v, e := stack.Peek()
	assert.Nil(t, e)
	assert.Equal(t, 3, v)
}

func TestSyncStack_Pop(t *testing.T) {
	stack := NewSyncStack[int]()
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)

	val, err := stack.Pop()
	assert.Nil(t, err)
	assert.Equal(t, 3, val)
	assert.Equal(t, 2, stack.Size())

	val, err = stack.Pop()
	assert.Nil(t, err)
	assert.Equal(t, 2, val)
	assert.Equal(t, 1, stack.Size())

	val, err = stack.Pop()
	assert.Nil(t, err)
	assert.Equal(t, 1, val)
	assert.Equal(t, 0, stack.Size())

	_, err = stack.Pop()
	assert.NotNil(t, err)
}

func TestSyncStack_Peek(t *testing.T) {
	stack := NewSyncStack[int]()
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)

	val, err := stack.Peek()
	assert.Nil(t, err)
	assert.Equal(t, 3, val)
	assert.Equal(t, 3, stack.Size())

	stack.Pop()
	val, err = stack.Peek()
	assert.Nil(t, err)
	assert.Equal(t, 2, val)
	assert.Equal(t, 2, stack.Size())
}

func TestSyncStack_Iterator(t *testing.T) {
	stack := NewSyncStack[int]()
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)

	it := stack.Iterator()
	assert.True(t, it.HasNext())
	assert.Equal(t, 3, it.Next())
	assert.True(t, it.HasNext())
	assert.Equal(t, 2, it.Next())
	assert.True(t, it.HasNext())
	assert.Equal(t, 1, it.Next())
	assert.False(t, it.HasNext())
}
