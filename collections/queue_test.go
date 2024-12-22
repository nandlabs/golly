package collections

import (
	"testing"

	"oss.nandlabs.io/golly/testing/assert"
)

func TestArrayQueue_Enqueue(t *testing.T) {
	queue := NewArrayQueue[int]()
	err := queue.Enqueue(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, queue.Size())
	assert.False(t, queue.IsEmpty())
}

func TestArrayQueue_Dequeue(t *testing.T) {
	queue := NewArrayQueue[int]()
	queue.Enqueue(1)
	queue.Enqueue(2)

	val, err := queue.Dequeue()
	assert.Nil(t, err)
	assert.Equal(t, 1, val)
	assert.Equal(t, 1, queue.Size())

	val, err = queue.Dequeue()
	assert.Nil(t, err)
	assert.Equal(t, 2, val)
	assert.Equal(t, 0, queue.Size())
	assert.True(t, queue.IsEmpty())

	_, err = queue.Dequeue()
	assert.NotNil(t, err)
}

func TestArrayQueue_Front(t *testing.T) {
	queue := NewArrayQueue[int]()
	queue.Enqueue(1)
	queue.Enqueue(2)

	val, err := queue.Front()
	assert.Nil(t, err)
	assert.Equal(t, 1, val)

	queue.Dequeue()
	val, err = queue.Front()
	assert.Nil(t, err)
	assert.Equal(t, 2, val)

	queue.Dequeue()
	_, err = queue.Front()
	assert.NotNil(t, err)
}

func TestSyncQueue_Enqueue(t *testing.T) {
	queue := NewSyncQueue[int]()
	err := queue.Enqueue(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, queue.Size())
	assert.False(t, queue.IsEmpty())
}

func TestSyncQueue_Dequeue(t *testing.T) {
	queue := NewSyncQueue[int]()
	queue.Enqueue(1)
	queue.Enqueue(2)

	val, err := queue.Dequeue()
	assert.Nil(t, err)
	assert.Equal(t, 1, val)
	assert.Equal(t, 1, queue.Size())

	val, err = queue.Dequeue()
	assert.Nil(t, err)
	assert.Equal(t, 2, val)
	assert.Equal(t, 0, queue.Size())
	assert.True(t, queue.IsEmpty())

	_, err = queue.Dequeue()
	assert.NotNil(t, err)
}

func TestSyncQueue_Front(t *testing.T) {
	queue := NewSyncQueue[int]()
	queue.Enqueue(1)
	queue.Enqueue(2)

	val, err := queue.Front()
	assert.Nil(t, err)
	assert.Equal(t, 1, val)

	queue.Dequeue()
	val, err = queue.Front()
	assert.Nil(t, err)
	assert.Equal(t, 2, val)

	queue.Dequeue()
	_, err = queue.Front()
	assert.NotNil(t, err)
}
