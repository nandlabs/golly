package wsmux

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"oss.nandlabs.io/golly/ws"
)

// --- unit tests: routing & error paths without a real connection ---

func TestMux_RoutesByChannel(t *testing.T) {
	mux := New()
	var (
		gotChan    string
		gotPayload string
	)
	mux.HandleFunc("greet", func(env *Envelope, _ Replier) {
		gotChan = env.Channel
		var s string
		_ = json.Unmarshal(env.Payload, &s)
		gotPayload = s
	})

	// stub conn — handler does NOT use the replier, so a zero conn is safe.
	raw, _ := json.Marshal(Envelope{Channel: "greet", Payload: json.RawMessage(`"hello"`)})
	mux.ServeMessage(&ws.Conn{}, ws.NewTextMessage(raw))

	if gotChan != "greet" {
		t.Errorf("channel = %q, want greet", gotChan)
	}
	if gotPayload != "hello" {
		t.Errorf("payload = %q, want hello", gotPayload)
	}
}

func TestMux_UnknownChannelCallsOnUnknown(t *testing.T) {
	mux := New()
	called := false
	mux.OnUnknown(func(_ *ws.Conn, env *Envelope, err error) {
		called = true
		if !errors.Is(err, ErrUnknownChannel) {
			t.Errorf("err = %v, want ErrUnknownChannel", err)
		}
		if env.Channel != "ghost" {
			t.Errorf("env.Channel = %q, want ghost", env.Channel)
		}
	})

	raw, _ := json.Marshal(Envelope{Channel: "ghost"})
	mux.ServeMessage(&ws.Conn{}, ws.NewTextMessage(raw))

	if !called {
		t.Fatal("OnUnknown was not invoked")
	}
}

func TestMux_DecodeErrorOnBadJSON(t *testing.T) {
	mux := New()
	var captured error
	mux.OnDecodeError(func(_ *ws.Conn, _ []byte, err error) {
		captured = err
	})

	mux.ServeMessage(&ws.Conn{}, ws.NewTextMessage([]byte(`{not json`)))

	if captured == nil {
		t.Fatal("OnDecodeError was not invoked")
	}
	if !strings.Contains(captured.Error(), "invalid envelope JSON") {
		t.Errorf("err = %v, want it to mention invalid envelope JSON", captured)
	}
}

func TestMux_DecodeErrorOnMissingChannel(t *testing.T) {
	mux := New()
	var captured error
	mux.OnDecodeError(func(_ *ws.Conn, _ []byte, err error) {
		captured = err
	})

	mux.ServeMessage(&ws.Conn{}, ws.NewTextMessage([]byte(`{"type":"x"}`)))

	if captured == nil {
		t.Fatal("OnDecodeError was not invoked for missing channel")
	}
	if !strings.Contains(captured.Error(), "missing channel") {
		t.Errorf("err = %v, want it to mention missing channel", captured)
	}
}

func TestMux_RejectsBinaryFrames(t *testing.T) {
	mux := New()
	var captured error
	mux.OnDecodeError(func(_ *ws.Conn, _ []byte, err error) {
		captured = err
	})

	mux.ServeMessage(&ws.Conn{}, ws.NewBinaryMessage([]byte{0x01, 0x02}))

	if captured == nil || !strings.Contains(captured.Error(), "non-text") {
		t.Errorf("expected non-text decode error; got %v", captured)
	}
}

func TestMux_HandleOverwrites(t *testing.T) {
	mux := New()
	calls := 0
	mux.HandleFunc("x", func(_ *Envelope, _ Replier) { calls += 10 })
	mux.HandleFunc("x", func(_ *Envelope, _ Replier) { calls++ })

	raw, _ := json.Marshal(Envelope{Channel: "x"})
	mux.ServeMessage(&ws.Conn{}, ws.NewTextMessage(raw))

	if calls != 1 {
		t.Errorf("expected the second handler to win (calls=1); got %d", calls)
	}
}

// --- integration test: full request/reply over a real WebSocket ---

func TestMux_RequestReply_EndToEnd(t *testing.T) {
	mux := New()
	mux.HandleFunc("ping", func(env *Envelope, r Replier) {
		var s string
		_ = json.Unmarshal(env.Payload, &s)
		_ = r.Reply("pong", map[string]string{"echo": s})
	})

	handler := ws.NewHandler(ws.WithPingInterval(0))
	handler.OnMessage(mux.ServeMessage)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	client, err := ws.Dial(context.Background(), wsURL, ws.WithPingInterval(0))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer client.Close()

	req, _ := json.Marshal(Envelope{
		Type:    "req",
		Channel: "ping",
		ID:      "abc-123",
		Payload: json.RawMessage(`"hi"`),
	})
	if err := client.Send(ws.NewTextMessage(req)); err != nil {
		t.Fatalf("send: %v", err)
	}

	select {
	case m := <-client.Messages():
		var env Envelope
		if err := json.Unmarshal(m.Data, &env); err != nil {
			t.Fatalf("decode reply: %v", err)
		}
		if env.Type != "pong" {
			t.Errorf("reply type = %q, want pong", env.Type)
		}
		if env.Channel != "ping" {
			t.Errorf("reply channel = %q, want ping (echoed)", env.Channel)
		}
		if env.ID != "abc-123" {
			t.Errorf("reply id = %q, want abc-123 (correlation)", env.ID)
		}
		var body map[string]string
		_ = json.Unmarshal(env.Payload, &body)
		if body["echo"] != "hi" {
			t.Errorf("reply payload.echo = %q, want hi", body["echo"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no reply received")
	}
}

// TestMux_ConcurrentRegistrations verifies handler registration is safe under
// concurrent calls (the lock matters when handlers re-register dynamically).
func TestMux_ConcurrentRegistrations(t *testing.T) {
	mux := New()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mux.HandleFunc("c", func(_ *Envelope, _ Replier) {})
			mux.OnUnknown(func(_ *ws.Conn, _ *Envelope, _ error) {})
		}(i)
	}
	wg.Wait()

	// Just make sure it still routes correctly after the storm.
	raw, _ := json.Marshal(Envelope{Channel: "c"})
	mux.ServeMessage(&ws.Conn{}, ws.NewTextMessage(raw))
}
