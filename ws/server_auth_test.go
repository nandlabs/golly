package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type userKey struct{}

// TestUpgradeAuth_Accepts verifies that a hook returning ok=true allows the
// upgrade and that the returned context is reachable via Conn.Context() in
// the OnConnect/OnMessage handlers.
func TestUpgradeAuth_Accepts(t *testing.T) {
	gotCtx := make(chan context.Context, 1)
	handler := NewHandler(
		WithPingInterval(0),
		WithUpgradeAuth(func(r *http.Request) (bool, context.Context) {
			if r.Header.Get("X-User") == "" {
				return false, nil
			}
			return true, context.WithValue(r.Context(), userKey{}, r.Header.Get("X-User"))
		}),
	)
	handler.OnConnect(func(c *Conn) { gotCtx <- c.Context() })
	handler.OnMessage(func(c *Conn, m Message) { _ = c.Send(m) })

	srv := httptest.NewServer(handler)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	client, err := Dial(context.Background(), wsURL, WithPingInterval(0), WithHeader("X-User", "alice"))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer client.Close()

	select {
	case ctx := <-gotCtx:
		if user, _ := ctx.Value(userKey{}).(string); user != "alice" {
			t.Errorf("Conn.Context() user = %q, want alice", user)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server never observed connection")
	}
}

// TestUpgradeAuth_Rejects verifies the handshake is refused with 401 when the
// hook returns ok=false; Dial must surface an error.
func TestUpgradeAuth_Rejects(t *testing.T) {
	handler := NewHandler(
		WithPingInterval(0),
		WithUpgradeAuth(func(r *http.Request) (bool, context.Context) {
			return false, nil
		}),
	)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	_, err := Dial(context.Background(), wsURL, WithPingInterval(0))
	if err == nil {
		t.Fatal("expected dial to fail when UpgradeAuth rejects")
	}
}

// TestUpgradeAuth_NotConfigured verifies the default context is
// context.Background() when no hook is wired (existing behavior preserved).
func TestUpgradeAuth_NotConfigured(t *testing.T) {
	gotCtx := make(chan context.Context, 1)
	handler := NewHandler(WithPingInterval(0))
	handler.OnConnect(func(c *Conn) { gotCtx <- c.Context() })

	srv := httptest.NewServer(handler)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	client, err := Dial(context.Background(), wsURL, WithPingInterval(0))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer client.Close()

	select {
	case ctx := <-gotCtx:
		if ctx == nil {
			t.Fatal("Conn.Context() returned nil; want context.Background()")
		}
		if ctx.Err() != nil {
			t.Errorf("default context should be live; got err %v", ctx.Err())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server never observed connection")
	}
}

// TestHub_SendEndToEnd uses two real connections registered in a Hub and
// verifies a targeted Send delivers only to the matching key.
func TestHub_SendEndToEnd(t *testing.T) {
	hub := NewHub()
	handler := NewHandler(WithPingInterval(0))

	// Tag each new connection with a header-derived user id.
	connReady := make(chan *Conn, 2)
	handler.OnConnect(func(c *Conn) {
		connReady <- c
	})
	handler.OnDisconnect(func(c *Conn, _ error) {
		hub.Remove(c)
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	// Dial two clients. Close them all at once via a single deferred loop
	// (gocritic.deferInLoop avoids per-iteration defers).
	clients := make([]*Client, 2)
	defer func() {
		for _, c := range clients {
			if c != nil {
				_ = c.Close()
			}
		}
	}()
	for i := range clients {
		c, err := Dial(context.Background(), wsURL, WithPingInterval(0))
		if err != nil {
			t.Fatalf("dial %d: %v", i, err)
		}
		clients[i] = c
	}

	// Collect the server-side connections and put them into the Hub under
	// distinct keys.
	serverConns := make([]*Conn, 0, 2)
	for i := 0; i < 2; i++ {
		select {
		case sc := <-connReady:
			serverConns = append(serverConns, sc)
		case <-time.After(2 * time.Second):
			t.Fatal("server never accepted connection")
		}
	}
	hub.Add(serverConns[0], "user-a")
	hub.Add(serverConns[1], "user-b")

	// Send only to user-a.
	n := hub.Send("user-a", NewTextMessage([]byte("hello-a")))
	if n != 1 {
		t.Errorf("Send to user-a delivered to %d, want 1", n)
	}

	select {
	case m := <-clients[0].Messages():
		if string(m.Data) != "hello-a" {
			t.Errorf("client 0 got %q, want hello-a", m.Data)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("client 0 never received message")
	}

	// Client 1 must NOT receive anything.
	select {
	case m := <-clients[1].Messages():
		t.Errorf("client 1 unexpectedly received %q", m.Data)
	case <-time.After(200 * time.Millisecond):
		// good — no message
	}
}
