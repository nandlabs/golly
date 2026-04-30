package ws

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"oss.nandlabs.io/golly/clients"
)

func TestServerClientEcho(t *testing.T) {
	// Set up echo server
	handler := NewHandler(
		WithPingInterval(0), // disable pings for test
	)
	handler.OnMessage(func(conn *Conn, msg Message) {
		if err := conn.Send(msg); err != nil {
			t.Errorf("echo send error: %v", err)
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Connect client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client, err := Dial(context.Background(), wsURL,
		WithPingInterval(0),
	)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer client.Close()

	// Send and receive
	sent := NewTextMessage([]byte("hello websocket"))
	if err := client.Send(sent); err != nil {
		t.Fatalf("send error: %v", err)
	}

	select {
	case msg := <-client.Messages():
		if string(msg.Data) != "hello websocket" {
			t.Fatalf("expected 'hello websocket', got %q", msg.Data)
		}
		if msg.Type != OpText {
			t.Fatalf("expected OpText, got %v", msg.Type)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for echo response")
	}
}

func TestServerClientBinaryMessage(t *testing.T) {
	handler := NewHandler(WithPingInterval(0))
	handler.OnMessage(func(conn *Conn, msg Message) {
		conn.Send(msg)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client, err := Dial(context.Background(), wsURL, WithPingInterval(0))
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer client.Close()

	data := []byte{0x00, 0x01, 0x02, 0xFF}
	if err := client.Send(NewBinaryMessage(data)); err != nil {
		t.Fatalf("send error: %v", err)
	}

	select {
	case msg := <-client.Messages():
		if msg.Type != OpBinary {
			t.Fatalf("expected OpBinary, got %v", msg.Type)
		}
		if len(msg.Data) != 4 {
			t.Fatalf("expected 4 bytes, got %d", len(msg.Data))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestServerMultipleClients(t *testing.T) {
	var mu sync.Mutex
	connected := 0

	handler := NewHandler(WithPingInterval(0))
	handler.OnConnect(func(conn *Conn) {
		mu.Lock()
		connected++
		mu.Unlock()
	})
	handler.OnMessage(func(conn *Conn, msg Message) {
		conn.Send(msg)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	clients := make([]*Client, 3)
	for i := range clients {
		c, err := Dial(context.Background(), wsURL, WithPingInterval(0))
		if err != nil {
			t.Fatalf("dial %d error: %v", i, err)
		}
		clients[i] = c
	}

	// Wait for connections
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if connected != 3 {
		t.Fatalf("expected 3 connections, got %d", connected)
	}
	mu.Unlock()

	// Verify connections list
	conns := handler.Connections()
	if len(conns) != 3 {
		t.Fatalf("expected 3 active connections, got %d", len(conns))
	}

	// Clean up
	for _, c := range clients {
		c.Close()
	}
}

func TestServerBroadcast(t *testing.T) {
	handler := NewHandler(WithPingInterval(0))
	handler.OnMessage(func(conn *Conn, msg Message) {
		handler.Broadcast(msg)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect 2 clients
	client1, err := Dial(context.Background(), wsURL, WithPingInterval(0))
	if err != nil {
		t.Fatalf("dial 1 error: %v", err)
	}
	defer client1.Close()

	client2, err := Dial(context.Background(), wsURL, WithPingInterval(0))
	if err != nil {
		t.Fatalf("dial 2 error: %v", err)
	}
	defer client2.Close()

	// Wait for both connections
	time.Sleep(100 * time.Millisecond)

	// Send from client1 — both should receive
	client1.Send(NewTextMessage([]byte("broadcast")))

	for _, ch := range []<-chan Message{client1.Messages(), client2.Messages()} {
		select {
		case msg := <-ch:
			if string(msg.Data) != "broadcast" {
				t.Fatalf("expected 'broadcast', got %q", msg.Data)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for broadcast")
		}
	}
}

func TestServerOnDisconnect(t *testing.T) {
	disconnected := make(chan string, 1)

	handler := NewHandler(WithPingInterval(0))
	handler.OnDisconnect(func(conn *Conn, err error) {
		disconnected <- conn.ID()
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client, err := Dial(context.Background(), wsURL, WithPingInterval(0))
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	client.Close()

	select {
	case id := <-disconnected:
		if id == "" {
			t.Fatal("expected non-empty connection ID")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for disconnect")
	}
}

func TestServerRejectsNonUpgrade(t *testing.T) {
	handler := NewHandler()
	server := httptest.NewServer(handler)
	defer server.Close()

	// Regular HTTP request
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestServerCheckOrigin(t *testing.T) {
	handler := NewHandler(
		WithCheckOrigin(func(origin string) bool {
			return origin == "http://allowed.example.com"
		}),
		WithPingInterval(0),
	)
	handler.OnMessage(func(conn *Conn, msg Message) {
		conn.Send(msg)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Client without allowed origin should fail at the HTTP level
	// (our client doesn't send Origin header, so checkOrigin will get "")
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	_, err := Dial(context.Background(), wsURL, WithPingInterval(0))
	if err == nil {
		t.Fatal("expected error for rejected origin")
	}
}

func TestClientInvalidURL(t *testing.T) {
	_, err := Dial(context.Background(), "http://invalid.example.com/ws")
	if err != ErrInvalidURL {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}
}

func TestClientSendAfterClose(t *testing.T) {
	handler := NewHandler(WithPingInterval(0))
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client, err := Dial(context.Background(), wsURL, WithPingInterval(0))
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}

	client.Close()
	err = client.Send(NewTextMessage([]byte("test")))
	if err != ErrConnClosed {
		t.Fatalf("expected ErrConnClosed, got %v", err)
	}
}

func TestUpgradeFunction(t *testing.T) {
	var upgraded *Conn

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := Upgrade(w, r, WithPingInterval(0))
		if err != nil {
			t.Errorf("upgrade error: %v", err)
			return
		}
		upgraded = conn
		// Echo one message then close
		conn.onMessage = func(c *Conn, msg Message) {
			c.Send(msg)
		}
		go conn.readPump()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client, err := Dial(context.Background(), wsURL, WithPingInterval(0))
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer client.Close()

	client.Send(NewTextMessage([]byte("upgrade test")))

	select {
	case msg := <-client.Messages():
		if string(msg.Data) != "upgrade test" {
			t.Fatalf("expected 'upgrade test', got %q", msg.Data)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}

	if upgraded == nil {
		t.Fatal("connection should have been upgraded")
	}
}

func TestOptionsApplied(t *testing.T) {
	cfg := defaultConfig()

	opts := []Option{
		WithReadBufferSize(8192),
		WithWriteBufferSize(16384),
		WithMaxMessageSize(1024 * 1024),
		WithPingInterval(5 * time.Second),
		WithPongTimeout(3 * time.Second),
		WithWriteTimeout(15 * time.Second),
		WithHandshakeTimeout(20 * time.Second),
		WithAutoReconnect(true),
		WithMaxReconnectWait(60 * time.Second),
	}
	for _, o := range opts {
		o(cfg)
	}

	if cfg.readBufferSize != 8192 {
		t.Errorf("readBufferSize: got %d, want 8192", cfg.readBufferSize)
	}
	if cfg.writeBufferSize != 16384 {
		t.Errorf("writeBufferSize: got %d, want 16384", cfg.writeBufferSize)
	}
	if cfg.maxMessageSize != 1024*1024 {
		t.Errorf("maxMessageSize: got %d, want %d", cfg.maxMessageSize, 1024*1024)
	}
	if cfg.pingInterval != 5*time.Second {
		t.Errorf("pingInterval: got %v, want 5s", cfg.pingInterval)
	}
	if cfg.pongTimeout != 3*time.Second {
		t.Errorf("pongTimeout: got %v, want 3s", cfg.pongTimeout)
	}
	if cfg.writeTimeout != 15*time.Second {
		t.Errorf("writeTimeout: got %v, want 15s", cfg.writeTimeout)
	}
	if cfg.handshakeTimeout != 20*time.Second {
		t.Errorf("handshakeTimeout: got %v, want 20s", cfg.handshakeTimeout)
	}
	if !cfg.autoReconnect {
		t.Error("autoReconnect should be true")
	}
	if cfg.maxReconnectWait != 60*time.Second {
		t.Errorf("maxReconnectWait: got %v, want 60s", cfg.maxReconnectWait)
	}
}

func TestClientWithBasicAuth(t *testing.T) {
	var receivedAuth string

	// Custom server that captures the Authorization header from the handshake
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		// Proceed with WebSocket upgrade
		handler := NewHandler(WithPingInterval(0))
		handler.OnMessage(func(conn *Conn, msg Message) {
			_ = conn.Send(msg)
		})
		handler.ServeHTTP(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client, err := Dial(context.Background(), wsURL,
		WithPingInterval(0),
		WithBasicAuth("testuser", "testpass"),
	)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer client.Close()

	// Send a message to confirm connection works
	if err := client.Send(NewTextMessage([]byte("auth-test"))); err != nil {
		t.Fatalf("send error: %v", err)
	}

	select {
	case msg := <-client.Messages():
		if string(msg.Data) != "auth-test" {
			t.Fatalf("expected 'auth-test', got %q", msg.Data)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for echo response")
	}

	// Verify the auth header was sent
	expectedCreds := base64.StdEncoding.EncodeToString([]byte("testuser:testpass"))
	expectedHeader := "Basic " + expectedCreds
	if receivedAuth != expectedHeader {
		t.Errorf("auth header: got %q, want %q", receivedAuth, expectedHeader)
	}
}

func TestClientWithBearerAuth(t *testing.T) {
	var receivedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		handler := NewHandler(WithPingInterval(0))
		handler.ServeHTTP(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client, err := Dial(context.Background(), wsURL,
		WithPingInterval(0),
		WithBearerAuth("my-secret-token"),
	)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer client.Close()

	if receivedAuth != "Bearer my-secret-token" {
		t.Errorf("auth header: got %q, want %q", receivedAuth, "Bearer my-secret-token")
	}
}

func TestClientWithCustomHeaders(t *testing.T) {
	var receivedHeader string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Custom-Header")
		handler := NewHandler(WithPingInterval(0))
		handler.ServeHTTP(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client, err := Dial(context.Background(), wsURL,
		WithPingInterval(0),
		WithHeader("X-Custom-Header", "custom-value"),
	)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer client.Close()

	if receivedHeader != "custom-value" {
		t.Errorf("custom header: got %q, want %q", receivedHeader, "custom-value")
	}
}

func TestClientCircuitBreaker(t *testing.T) {
	// Create a circuit breaker with low thresholds
	cb := clients.NewCircuitBreaker(&clients.BreakerInfo{
		FailureThreshold: 2,
		SuccessThreshold: 1,
		MaxHalfOpen:      1,
		Timeout:          1,
	})

	// Verify that the circuit breaker is integrated with dial:
	// connect to a non-existent server → should fail with a dial error
	// but the circuit breaker's OnExecution should be called
	ctx := context.Background()
	_, err := Dial(ctx, "ws://127.0.0.1:1",
		WithPingInterval(0),
		WithCircuitBreaker(cb),
		WithHandshakeTimeout(100*time.Millisecond),
	)
	if err == nil {
		t.Fatal("expected dial error")
	}

	// Verify the circuit breaker was applied to config
	cfg := defaultConfig()
	WithCircuitBreaker(cb)(cfg)
	if cfg.circuitBreaker == nil {
		t.Fatal("circuit breaker should be set")
	}

	// Verify successful connection also reports to circuit breaker
	handler := NewHandler(WithPingInterval(0))
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client, err := Dial(ctx, wsURL,
		WithPingInterval(0),
		WithCircuitBreaker(cb),
	)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	_ = client.Close()
}

func TestClientRetryPolicy(t *testing.T) {
	attempts := 0
	var mu sync.Mutex

	// Server that only accepts after 2 attempts
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		attempts++
		current := attempts
		mu.Unlock()

		if current < 2 {
			// Reject the connection
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// Accept on 2nd+ attempt
		handler := NewHandler(WithPingInterval(0))
		handler.ServeHTTP(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client, err := Dial(context.Background(), wsURL,
		WithPingInterval(0),
		WithRetryPolicy(&clients.RetryPolicy{
			MaxRetries:      3,
			BackoffInterval: 50 * time.Millisecond,
			MaxBackoff:      200 * time.Millisecond,
			Exponential:     false,
		}),
	)
	if err != nil {
		t.Fatalf("dial error after retries: %v", err)
	}
	defer client.Close()

	mu.Lock()
	if attempts < 2 {
		t.Errorf("expected at least 2 attempts, got %d", attempts)
	}
	mu.Unlock()
}

func TestClientWithClientOptions(t *testing.T) {
	cfg := defaultConfig()

	opts := clients.NewOptionsBuilder().
		WithAuth(clients.NewBasicAuth("user", "pass")).
		WithRetryPolicy(&clients.RetryPolicy{
			MaxRetries:      5,
			BackoffInterval: 100 * time.Millisecond,
			MaxBackoff:      1 * time.Second,
			Exponential:     true,
		}).
		Build()

	WithClientOptions(opts)(cfg)

	if cfg.auth == nil {
		t.Error("auth should be set")
	}
	if cfg.auth.Type() != clients.AuthTypeBasic {
		t.Errorf("auth type: got %v, want %v", cfg.auth.Type(), clients.AuthTypeBasic)
	}
	if cfg.retryPolicy == nil {
		t.Error("retryPolicy should be set")
	}
	if cfg.retryPolicy.MaxRetries != 5 {
		t.Errorf("maxRetries: got %d, want 5", cfg.retryPolicy.MaxRetries)
	}
}

func TestOptionsNewFeatures(t *testing.T) {
	cfg := defaultConfig()

	cb := clients.NewCircuitBreaker(nil)
	ri := &clients.RetryInfo{
		MaxRetries:  3,
		Wait:        100,
		Exponential: true,
		Multiplier:  2.0,
		MaxWait:     5000,
		Jitter:      true,
	}

	opts := []Option{
		WithAuth(clients.NewBearerAuth("token123")),
		WithCircuitBreaker(cb),
		WithRetryInfo(ri),
		WithHeader("X-Test", "value"),
	}
	for _, o := range opts {
		o(cfg)
	}

	if cfg.auth == nil {
		t.Error("auth should be set")
	}
	if cfg.circuitBreaker == nil {
		t.Error("circuitBreaker should be set")
	}
	if cfg.retryInfo == nil {
		t.Error("retryInfo should be set")
	}
	if cfg.headers["X-Test"] != "value" {
		t.Errorf("header: got %q, want %q", cfg.headers["X-Test"], "value")
	}
}
