package messaging

import (
	"testing"
	"time"

	"oss.nandlabs.io/golly/testing/assert"
)

// TestBackpressure_LocalBufferSizeAppliesAtCreation verifies the option
// is honored on first-create of the destination.
func TestBackpressure_LocalBufferSizeAppliesAtCreation(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://backpressure-buf-size")

	opts := NewOptionsBuilder().AddLocalBufferSize(4).Build()

	// Fill the small buffer.
	for i := 0; i < 4; i++ {
		msg, _ := NewLocalMessage()
		assert.NoError(t, lp.Send(u, msg, opts...))
	}
	// Next non-blocking Send must overflow.
	msg, _ := NewLocalMessage()
	if err := lp.Send(u, msg, opts...); err != ErrChannelFull {
		t.Errorf("expected ErrChannelFull, got %v", err)
	}
}

// TestBackpressure_BlockOnFullBlocksUntilRoom verifies BlockOnFullOpt
// makes Send wait for a slot rather than returning ErrChannelFull.
func TestBackpressure_BlockOnFullBlocksUntilRoom(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://backpressure-block")

	// Use a buffer of 1 so we can drive the block-until-room path deterministically.
	initOpts := NewOptionsBuilder().AddLocalBufferSize(1).Build()
	blockOpts := NewOptionsBuilder().AddBlockOnFull().Build()

	// Prime the buffer.
	msg, _ := NewLocalMessage()
	assert.NoError(t, lp.Send(u, msg, initOpts...))

	// Kick off a blocking Send in a goroutine — it should hang until we drain.
	sendDone := make(chan error, 1)
	go func() {
		msg2, _ := NewLocalMessage()
		sendDone <- lp.Send(u, msg2, blockOpts...)
	}()

	// Give Send a moment to start blocking.
	select {
	case <-sendDone:
		t.Fatal("blocking Send returned before buffer drained")
	case <-time.After(30 * time.Millisecond):
	}

	// Drain — this frees a slot for the pending Send.
	_, err := lp.Receive(u)
	assert.NoError(t, err)

	select {
	case err := <-sendDone:
		assert.NoError(t, err)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("blocking Send never completed after drain")
	}
}

// TestBackpressure_WithoutBlockOnFullReturnsChannelFull is the
// regression test for the default (v1.6) behavior: an ErrChannelFull
// short-circuits instead of blocking.
func TestBackpressure_WithoutBlockOnFullReturnsChannelFull(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://backpressure-no-block")

	opts := NewOptionsBuilder().AddLocalBufferSize(1).Build()
	msg, _ := NewLocalMessage()
	assert.NoError(t, lp.Send(u, msg, opts...))

	msg2, _ := NewLocalMessage()
	err := lp.Send(u, msg2, opts...)
	if err != ErrChannelFull {
		t.Errorf("expected ErrChannelFull without BlockOnFull, got %v", err)
	}
}

// TestBackpressure_SetDefaultBufferSize verifies the global override
// changes the buffer size for subsequently created destinations.
func TestBackpressure_SetDefaultBufferSize(t *testing.T) {
	lp := newTestProvider(t)
	lp.SetDefaultBufferSize(3)

	u := mustParseURL(t, "chan://backpressure-default-override")

	// Fill 3 slots — the value we just set.
	for i := 0; i < 3; i++ {
		msg, _ := NewLocalMessage()
		assert.NoError(t, lp.Send(u, msg))
	}
	msg, _ := NewLocalMessage()
	if err := lp.Send(u, msg); err != ErrChannelFull {
		t.Errorf("expected ErrChannelFull at 4th send with default=3, got %v", err)
	}
}

// TestBackpressure_LocalBufferSizeIgnoredAfterCreation verifies that
// re-passing LocalBufferSizeOpt does not resize an existing channel.
func TestBackpressure_LocalBufferSizeIgnoredAfterCreation(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://backpressure-second-call")

	// Create with 2 slots.
	init := NewOptionsBuilder().AddLocalBufferSize(2).Build()
	msg, _ := NewLocalMessage()
	assert.NoError(t, lp.Send(u, msg, init...))

	// Attempt to "grow" to 100 slots — should still be 2.
	grow := NewOptionsBuilder().AddLocalBufferSize(100).Build()
	msg2, _ := NewLocalMessage()
	assert.NoError(t, lp.Send(u, msg2, grow...))

	// 3rd send must overflow (proves we still have the original size of 2).
	msg3, _ := NewLocalMessage()
	if err := lp.Send(u, msg3, grow...); err != ErrChannelFull {
		t.Errorf("expected ErrChannelFull (option ignored after creation), got %v", err)
	}
}
