package messaging

import (
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"oss.nandlabs.io/golly/testing/assert"
)

// newTestProvider creates a fresh LocalProvider for isolated testing.
func newTestProvider(t *testing.T) *LocalProvider {
	t.Helper()
	lp := &LocalProvider{}
	err := lp.Setup()
	assert.NoError(t, err)
	return lp
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("failed to parse URL %q: %v", raw, err)
	}
	return u
}

// --- Basic Send / Receive ---

func TestLocalProvider_Send(t *testing.T) {
	lms := GetManager()
	msg, err := lms.NewMessage("chan")
	assert.NoError(t, err)
	input := "this is a test string"
	_, err = msg.SetBodyStr(input)
	if err != nil {
		t.Errorf("Error SetBodyStr:: %v", err)
	}
	uri, _ := url.Parse("chan://localhost:8080")
	got := lms.Send(uri, msg)
	if got != nil {
		t.Errorf("Error got :: %v", got)
	}
	uriErr, _ := url.Parse("http://localhost:8080")
	got = lms.Send(uriErr, msg)
	if got.Error() != "unsupported scheme http" {
		t.Errorf("Error got :: %v", got)
	}
}

func TestLocalProvider_SendBatch(t *testing.T) {
	lms := GetManager()
	msg1, err := lms.NewMessage("chan")
	assert.NoError(t, err)
	_, err = msg1.SetBodyStr("this is a test string 1")
	assert.NoError(t, err)

	msg2, err := lms.NewMessage("chan")
	assert.NoError(t, err)
	_, err = msg2.SetBodyStr("this is a test string 2")
	assert.NoError(t, err)

	uri, _ := url.Parse("chan://send-batch-host")
	msgs := []Message{msg1, msg2}
	got := lms.SendBatch(uri, msgs)
	if got != nil {
		t.Errorf("Error got :: %v", got)
	}
}

// --- Send and Receive round-trip ---

func TestLocalProvider_SendReceive(t *testing.T) {
	lp := newTestProvider(t)
	uri := mustParseURL(t, "chan://send-recv")

	msg, err := NewLocalMessage()
	assert.NoError(t, err)
	_, err = msg.SetBodyStr("hello")
	assert.NoError(t, err)

	err = lp.Send(uri, msg)
	assert.NoError(t, err)

	received, err := lp.Receive(uri)
	assert.NoError(t, err)
	if received.ReadAsStr() != "hello" {
		t.Errorf("expected %q, got %q", "hello", received.ReadAsStr())
	}
}

func TestLocalProvider_ReceiveBatch(t *testing.T) {
	lp := newTestProvider(t)
	uri := mustParseURL(t, "chan://recv-batch")

	// Send 3 messages
	for i := 0; i < 3; i++ {
		msg, err := NewLocalMessage()
		assert.NoError(t, err)
		_, err = msg.SetBodyStr("msg")
		assert.NoError(t, err)
		err = lp.Send(uri, msg)
		assert.NoError(t, err)
	}

	// Small delay to ensure all messages are buffered
	time.Sleep(10 * time.Millisecond)

	msgs, err := lp.ReceiveBatch(uri)
	assert.NoError(t, err)
	if len(msgs) < 1 {
		t.Fatalf("expected at least 1 message, got %d", len(msgs))
	}
	if len(msgs) > 3 {
		t.Fatalf("expected at most 3 messages, got %d", len(msgs))
	}
}

// --- Listener ---

func TestLocalProvider_AddListener(t *testing.T) {
	lp := newTestProvider(t)
	uri := mustParseURL(t, "chan://listener-test")

	var received atomic.Value
	done := make(chan struct{})

	err := lp.AddListener(uri, func(output Message) {
		received.Store(output.ReadAsStr())
		close(done)
	})
	assert.NoError(t, err)

	msg, err := NewLocalMessage()
	assert.NoError(t, err)
	_, err = msg.SetBodyStr("listener message")
	assert.NoError(t, err)
	err = lp.Send(uri, msg)
	assert.NoError(t, err)

	select {
	case <-done:
		if received.Load().(string) != "listener message" {
			t.Errorf("expected %q, got %q", "listener message", received.Load().(string))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("listener did not receive message within timeout")
	}
}

func TestLocalProvider_AddMultipleUnnamedListeners(t *testing.T) {
	lp := newTestProvider(t)
	uri := mustParseURL(t, "chan://multi-listener")

	var count atomic.Int32
	var wg sync.WaitGroup
	numListeners := 3
	wg.Add(numListeners)

	for i := 0; i < numListeners; i++ {
		err := lp.AddListener(uri, func(msg Message) {
			count.Add(1)
			wg.Done()
		})
		assert.NoError(t, err)
	}

	msg, err := NewLocalMessage()
	assert.NoError(t, err)
	_, err = msg.SetBodyStr("broadcast")
	assert.NoError(t, err)
	err = lp.Send(uri, msg)
	assert.NoError(t, err)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if count.Load() != int32(numListeners) {
			t.Errorf("expected %d listener calls, got %d", numListeners, count.Load())
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out: only %d of %d listeners fired", count.Load(), numListeners)
	}
}

func TestLocalProvider_AddNamedListener(t *testing.T) {
	lp := newTestProvider(t)
	uri := mustParseURL(t, "chan://named-listener")

	var received atomic.Int32
	done := make(chan struct{}, 1)

	// Add 2 listeners under the same name; only 1 should fire per message
	for i := 0; i < 2; i++ {
		err := lp.AddListener(uri, func(msg Message) {
			received.Add(1)
			select {
			case done <- struct{}{}:
			default:
			}
		}, Option{Key: "NamedListener", Value: "worker-group"})
		assert.NoError(t, err)
	}

	msg, err := NewLocalMessage()
	assert.NoError(t, err)
	_, err = msg.SetBodyStr("named test")
	assert.NoError(t, err)
	err = lp.Send(uri, msg)
	assert.NoError(t, err)

	select {
	case <-done:
		// Give a brief window for any duplicate dispatch
		time.Sleep(50 * time.Millisecond)
		if received.Load() != 1 {
			t.Errorf("expected exactly 1 named listener call, got %d", received.Load())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("named listener did not fire within timeout")
	}
}

// --- Concurrency ---

func TestLocalProvider_ConcurrentSendReceive(t *testing.T) {
	lp := newTestProvider(t)
	uri := mustParseURL(t, "chan://concurrent")
	numMessages := 100

	// Concurrent senders
	var sendWg sync.WaitGroup
	sendWg.Add(numMessages)
	for i := 0; i < numMessages; i++ {
		go func() {
			defer sendWg.Done()
			msg, _ := NewLocalMessage()
			_, _ = msg.SetBodyStr("concurrent")
			_ = lp.Send(uri, msg)
		}()
	}
	sendWg.Wait()

	// Drain all messages
	received := 0
	for received < numMessages {
		msgs, err := lp.ReceiveBatch(uri)
		assert.NoError(t, err)
		received += len(msgs)
	}
	if received != numMessages {
		t.Errorf("expected %d messages, got %d", numMessages, received)
	}
}

func TestLocalProvider_ConcurrentAddListener(t *testing.T) {
	lp := newTestProvider(t)
	uri := mustParseURL(t, "chan://concurrent-listener")

	var count atomic.Int32
	numListeners := 5
	var listenerWg sync.WaitGroup
	listenerWg.Add(numListeners)

	// Concurrently add multiple listeners
	for i := 0; i < numListeners; i++ {
		go func() {
			defer listenerWg.Done()
			_ = lp.AddListener(uri, func(msg Message) {
				count.Add(1)
			})
		}()
	}
	listenerWg.Wait()

	msg, _ := NewLocalMessage()
	_, _ = msg.SetBodyStr("fan-out")
	_ = lp.Send(uri, msg)

	// Wait for listeners to fire
	time.Sleep(200 * time.Millisecond)
	if count.Load() < 1 {
		t.Errorf("expected at least 1 listener call, got %d", count.Load())
	}
}

// --- Channel Full ---

func TestLocalProvider_ChannelFull(t *testing.T) {
	lp := newTestProvider(t)
	uri := mustParseURL(t, "chan://full-test")

	// Fill the buffered channel
	for i := 0; i < defaultChannelBufSize; i++ {
		msg, _ := NewLocalMessage()
		_, _ = msg.SetBodyStr("fill")
		err := lp.Send(uri, msg)
		assert.NoError(t, err)
	}

	// Next send should return ErrChannelFull
	msg, _ := NewLocalMessage()
	_, _ = msg.SetBodyStr("overflow")
	err := lp.Send(uri, msg)
	if err != ErrChannelFull {
		t.Errorf("expected ErrChannelFull, got %v", err)
	}
}

// --- Close ---

func TestLocalProvider_Close(t *testing.T) {
	lp := newTestProvider(t)
	uri := mustParseURL(t, "chan://close-test")

	// Send a message so channel is created
	msg, _ := NewLocalMessage()
	_, _ = msg.SetBodyStr("before close")
	_ = lp.Send(uri, msg)

	err := lp.Close()
	assert.NoError(t, err)

	// After close, getChan should return ErrProviderClosed
	_, err = lp.getChan(uri)
	if err != ErrProviderClosed {
		t.Errorf("expected ErrProviderClosed, got %v", err)
	}
}

func TestLocalProvider_SendAfterClose(t *testing.T) {
	lp := newTestProvider(t)
	uri := mustParseURL(t, "chan://send-after-close")

	err := lp.Close()
	assert.NoError(t, err)

	msg, _ := NewLocalMessage()
	_, _ = msg.SetBodyStr("should fail")
	err = lp.Send(uri, msg)
	if err != ErrProviderClosed {
		t.Errorf("expected ErrProviderClosed, got %v", err)
	}
}

func TestLocalProvider_ReceiveAfterClose(t *testing.T) {
	lp := newTestProvider(t)
	uri := mustParseURL(t, "chan://recv-after-close")

	// Create channel and close
	msg, _ := NewLocalMessage()
	_, _ = msg.SetBodyStr("data")
	_ = lp.Send(uri, msg)
	_ = lp.Close()

	// Receive should drain buffered message or return ErrProviderClosed
	_, err := lp.Receive(uri)
	// Could be nil (got buffered msg) or ErrProviderClosed (chan already drained)
	if err != nil && err != ErrProviderClosed {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- Schemes / ID ---

func TestLocalProvider_Schemes(t *testing.T) {
	lp := newTestProvider(t)
	schemes := lp.Schemes()
	if len(schemes) != 1 || schemes[0] != "chan" {
		t.Errorf("expected [chan], got %v", schemes)
	}
}

func TestLocalProvider_Id(t *testing.T) {
	lp := newTestProvider(t)
	if lp.Id() != "local-channel" {
		t.Errorf("expected local-channel, got %s", lp.Id())
	}
}

func TestLocalProvider_NewMessage(t *testing.T) {
	lp := newTestProvider(t)
	msg, err := lp.NewMessage("chan")
	assert.NoError(t, err)
	if msg == nil {
		t.Error("expected non-nil message")
	}
}

// --- Manager integration ---

func TestLocalProvider_ManagerAddListener(t *testing.T) {
	lms := GetManager()
	uri := mustParseURL(t, "chan://manager-listener")
	input := "listener via manager"

	var received atomic.Value
	done := make(chan struct{})

	err := lms.AddListener(uri, func(output Message) {
		received.Store(output.ReadAsStr())
		close(done)
	})
	assert.NoError(t, err)

	msg, err := lms.NewMessage("chan")
	assert.NoError(t, err)
	_, err = msg.SetBodyStr(input)
	assert.NoError(t, err)
	_ = lms.Send(uri, msg)

	select {
	case <-done:
		if received.Load().(string) != input {
			t.Errorf("expected %q, got %q", input, received.Load().(string))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("listener did not receive message within timeout")
	}
}

// --- Listener removal (ListenerRemover) ---

func TestLocalProvider_RemoveListeners_DropsAll(t *testing.T) {
	lp := newTestProvider(t)
	uri := mustParseURL(t, "chan://remove-all")

	var fired atomic.Int32
	err := lp.AddListener(uri, func(Message) { fired.Add(1) })
	assert.NoError(t, err)
	err = lp.AddListener(uri, func(Message) { fired.Add(1) })
	assert.NoError(t, err)

	// Remove all listeners; subsequent messages must not fire any.
	err = lp.RemoveListeners(uri)
	assert.NoError(t, err)

	msg, err := NewLocalMessage()
	assert.NoError(t, err)
	_, _ = msg.SetBodyStr("ignored")
	err = lp.Send(uri, msg)
	assert.NoError(t, err)

	// Give the dispatch goroutine time to (not) fire anything.
	time.Sleep(150 * time.Millisecond)
	if fired.Load() != 0 {
		t.Fatalf("expected 0 listener fires after RemoveListeners, got %d", fired.Load())
	}
}

func TestLocalProvider_RemoveListeners_Idempotent(t *testing.T) {
	lp := newTestProvider(t)
	uri := mustParseURL(t, "chan://remove-idempotent")

	// Remove with no listeners registered yet.
	assert.NoError(t, lp.RemoveListeners(uri))
	// Remove twice.
	assert.NoError(t, lp.RemoveListeners(uri))

	// Add a listener, remove, remove again.
	assert.NoError(t, lp.AddListener(uri, func(Message) {}))
	assert.NoError(t, lp.RemoveListeners(uri))
	assert.NoError(t, lp.RemoveListeners(uri))
}

func TestLocalProvider_RemoveNamedListener_LeavesOthers(t *testing.T) {
	lp := newTestProvider(t)
	uri := mustParseURL(t, "chan://remove-named")

	var unnamedFired atomic.Int32
	var keptNamedFired atomic.Int32
	var droppedNamedFired atomic.Int32

	assert.NoError(t, lp.AddListener(uri, func(Message) { unnamedFired.Add(1) }))
	assert.NoError(t, lp.AddListener(uri, func(Message) { keptNamedFired.Add(1) },
		NewOptionsBuilder().AddNamedListener("keep").Build()...))
	assert.NoError(t, lp.AddListener(uri, func(Message) { droppedNamedFired.Add(1) },
		NewOptionsBuilder().AddNamedListener("drop").Build()...))

	// Remove only the "drop" named group.
	assert.NoError(t, lp.RemoveNamedListener(uri, "drop"))

	msg, _ := NewLocalMessage()
	_, _ = msg.SetBodyStr("partial")
	assert.NoError(t, lp.Send(uri, msg))

	// Wait for dispatch.
	time.Sleep(150 * time.Millisecond)

	if unnamedFired.Load() != 1 {
		t.Errorf("unnamed listener should have fired once; got %d", unnamedFired.Load())
	}
	if keptNamedFired.Load() != 1 {
		t.Errorf("kept named listener should have fired once; got %d", keptNamedFired.Load())
	}
	if droppedNamedFired.Load() != 0 {
		t.Errorf("dropped named listener should NOT have fired; got %d", droppedNamedFired.Load())
	}
}

func TestLocalProvider_RemoveListeners_ClosedProvider(t *testing.T) {
	lp := &LocalProvider{}
	assert.NoError(t, lp.Setup())
	assert.NoError(t, lp.Close())
	uri := mustParseURL(t, "chan://closed")
	if err := lp.RemoveListeners(uri); err != ErrProviderClosed {
		t.Errorf("expected ErrProviderClosed, got %v", err)
	}
	if err := lp.RemoveNamedListener(uri, "foo"); err != ErrProviderClosed {
		t.Errorf("expected ErrProviderClosed, got %v", err)
	}
}

func TestManager_RemoveListeners_DelegatesToProvider(t *testing.T) {
	mgr := GetManager()
	uri := mustParseURL(t, "chan://manager-remove")

	var fired atomic.Int32
	assert.NoError(t, mgr.AddListener(uri, func(Message) { fired.Add(1) }))
	assert.NoError(t, mgr.RemoveListeners(uri))

	msg, _ := mgr.NewMessage("chan")
	_, _ = msg.SetBodyStr("nope")
	assert.NoError(t, mgr.Send(uri, msg))
	time.Sleep(150 * time.Millisecond)
	if fired.Load() != 0 {
		t.Fatalf("manager RemoveListeners did not propagate; %d fires", fired.Load())
	}
}

// TestManager_RemoveListeners_UnsupportedProvider verifies that a provider
// not implementing ListenerRemover triggers ErrListenerRemovalUnsupported.
func TestManager_RemoveListeners_UnsupportedProvider(t *testing.T) {
	// Construct a fresh manager so we don't pollute the default one.
	m := &managerImpl{knownProviders: make(map[string]Provider)}
	// A bare Provider — Schemes/Id only, no listener-removal support.
	bare := &bareProvider{}
	m.Register(bare)
	uri := mustParseURL(t, "bare://x")
	err := m.RemoveListeners(uri)
	if err != ErrListenerRemovalUnsupported {
		t.Fatalf("expected ErrListenerRemovalUnsupported, got %v", err)
	}
	err = m.RemoveNamedListener(uri, "n")
	if err != ErrListenerRemovalUnsupported {
		t.Fatalf("expected ErrListenerRemovalUnsupported, got %v", err)
	}
}

// bareProvider implements Provider but explicitly NOT ListenerRemover.
type bareProvider struct{}

func (p *bareProvider) Close() error                                             { return nil }
func (p *bareProvider) Id() string                                               { return "bare" }
func (p *bareProvider) Schemes() []string                                        { return []string{"bare"} }
func (p *bareProvider) Setup() error                                             { return nil }
func (p *bareProvider) NewMessage(string, ...Option) (Message, error)            { return nil, nil }
func (p *bareProvider) Send(*url.URL, Message, ...Option) error                  { return nil }
func (p *bareProvider) SendBatch(*url.URL, []Message, ...Option) error           { return nil }
func (p *bareProvider) Receive(*url.URL, ...Option) (Message, error)             { return nil, nil }
func (p *bareProvider) ReceiveBatch(*url.URL, ...Option) ([]Message, error)      { return nil, nil }
func (p *bareProvider) AddListener(*url.URL, func(msg Message), ...Option) error { return nil }
