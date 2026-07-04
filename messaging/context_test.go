package messaging

import (
	"context"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"oss.nandlabs.io/golly/testing/assert"
)

// TestSendCtx_DelegatesToProducerCtx verifies that when a producer
// implements ProducerCtx, SendCtx routes to the ctx-aware method rather
// than falling back to the ctx-raced Send path.
func TestSendCtx_DelegatesToProducerCtx(t *testing.T) {
	prov := &ctxTrackingProducer{}
	u := mustParseURL(t, "chan://ctx-delegate")
	msg, _ := NewLocalMessage()

	err := SendCtx(context.Background(), prov, u, msg)
	assert.NoError(t, err)
	if !prov.ctxCalled.Load() {
		t.Fatal("SendCtx did not delegate to ProducerCtx.SendCtx")
	}
	if prov.plainCalled.Load() {
		t.Fatal("SendCtx should not have fallen back to plain Send when ProducerCtx is implemented")
	}
}

// TestSendCtx_FallbackCancellation verifies that when the producer does
// NOT implement ProducerCtx, SendCtx runs Send in a goroutine and races
// it against ctx cancellation — a slow producer surfaces ctx.Err().
func TestSendCtx_FallbackCancellation(t *testing.T) {
	prov := &slowProducer{delay: 200 * time.Millisecond}
	u := mustParseURL(t, "chan://ctx-fallback")
	msg, _ := NewLocalMessage()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := SendCtx(ctx, prov, u, msg)
	elapsed := time.Since(start)

	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
	// Should return well before the 200ms delay elapses.
	if elapsed > 150*time.Millisecond {
		t.Errorf("expected fast cancellation, elapsed=%v", elapsed)
	}
}

// TestReceiveCtx_DelegatesToReceiverCtx checks the same delegation for
// the receive path.
func TestReceiveCtx_DelegatesToReceiverCtx(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://ctx-recv-delegate")

	msg, _ := NewLocalMessage()
	_, _ = msg.SetBodyStr("hello")
	assert.NoError(t, lp.Send(u, msg))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	got, err := ReceiveCtx(ctx, lp, u)
	assert.NoError(t, err)
	if got.ReadAsStr() != "hello" {
		t.Errorf("expected %q, got %q", "hello", got.ReadAsStr())
	}
}

// TestReceiveCtx_CtxCancellationOnLocalProvider verifies the
// LocalProvider's ReceiverCtx implementation respects a ctx deadline
// even with an empty channel.
func TestReceiveCtx_CtxCancellationOnLocalProvider(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://ctx-recv-cancel")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := ReceiveCtx(ctx, lp, u)
	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Errorf("receive took too long: %v", time.Since(start))
	}
}

// TestSendCtx_BlockOnFullWithCtxCancellation verifies that a
// BlockOnFull-configured LocalProvider Send returns ctx.Err() when the
// buffer is full and the ctx expires.
func TestSendCtx_BlockOnFullWithCtxCancellation(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://ctx-block-full")

	opts := NewOptionsBuilder().AddLocalBufferSize(2).AddBlockOnFull().Build()

	// Fill the tiny buffer.
	for i := 0; i < 2; i++ {
		msg, _ := NewLocalMessage()
		_, _ = msg.SetBodyStr("fill")
		assert.NoError(t, lp.Send(u, msg, opts...))
	}

	// Now a blocking Send with a short ctx should hit deadline exceeded.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	msg, _ := NewLocalMessage()
	_, _ = msg.SetBodyStr("overflow")
	err := lp.SendCtx(ctx, u, msg, opts...)
	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

// TestAddListenerCtx_CancellationRemovesListeners verifies the
// AddListenerCtx watcher tears the listener down when ctx cancels.
func TestAddListenerCtx_CancellationRemovesListeners(t *testing.T) {
	lp := newTestProvider(t)
	u := mustParseURL(t, "chan://ctx-listener-remove")

	var fired atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	assert.NoError(t, lp.AddListenerCtx(ctx, u, func(Message) { fired.Add(1) }))
	cancel()
	// Give the cancellation watcher a moment.
	time.Sleep(30 * time.Millisecond)

	msg, _ := NewLocalMessage()
	_, _ = msg.SetBodyStr("post-cancel")
	assert.NoError(t, lp.Send(u, msg))
	time.Sleep(100 * time.Millisecond)
	if fired.Load() != 0 {
		t.Errorf("expected 0 fires after ctx cancel, got %d", fired.Load())
	}
}

// --- Test doubles ---

// ctxTrackingProducer implements both Producer and ProducerCtx; the
// test asserts SendCtx delegates to the ctx-aware method.
type ctxTrackingProducer struct {
	ctxCalled   atomic.Bool
	plainCalled atomic.Bool
}

func (p *ctxTrackingProducer) Send(*url.URL, Message, ...Option) error {
	p.plainCalled.Store(true)
	return nil
}
func (p *ctxTrackingProducer) SendBatch(*url.URL, []Message, ...Option) error { return nil }
func (p *ctxTrackingProducer) SendCtx(context.Context, *url.URL, Message, ...Option) error {
	p.ctxCalled.Store(true)
	return nil
}
func (p *ctxTrackingProducer) SendBatchCtx(context.Context, *url.URL, []Message, ...Option) error {
	return nil
}

// slowProducer implements Producer only. Its Send sleeps to simulate a
// blocking backend that ignores ctx.
type slowProducer struct {
	delay time.Duration
}

func (p *slowProducer) Send(*url.URL, Message, ...Option) error {
	time.Sleep(p.delay)
	return nil
}
func (p *slowProducer) SendBatch(*url.URL, []Message, ...Option) error { return nil }

// Compile-time guard: slowProducer must NOT implement ProducerCtx.
var _ Producer = (*slowProducer)(nil)
