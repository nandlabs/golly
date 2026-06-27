package ws

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

const websocketGUID = "258EAFA5-E914-47DA-95CA-5AB5AC11653B"

// Handler is a WebSocket server handler that upgrades HTTP connections
// and manages WebSocket connections.
type Handler struct {
	cfg          *config
	mu           sync.RWMutex
	connections  map[string]*Conn
	onMessage    MessageHandler
	onConnect    ConnHandler
	onDisconnect DisconnectHandler
}

// NewHandler creates a new WebSocket handler with the given options.
func NewHandler(opts ...Option) *Handler {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	return &Handler{
		cfg:         cfg,
		connections: make(map[string]*Conn),
	}
}

// OnMessage registers a handler for incoming messages.
func (h *Handler) OnMessage(fn MessageHandler) {
	h.onMessage = fn
}

// OnConnect registers a handler called when a new connection is established.
func (h *Handler) OnConnect(fn ConnHandler) {
	h.onConnect = fn
}

// OnDisconnect registers a handler called when a connection is closed.
func (h *Handler) OnDisconnect(fn DisconnectHandler) {
	h.onDisconnect = fn
}

// Connections returns a snapshot of all active connections.
func (h *Handler) Connections() []*Conn {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns := make([]*Conn, 0, len(h.connections))
	for _, c := range h.connections {
		conns = append(conns, c)
	}
	return conns
}

// Broadcast sends a message to all connected clients.
func (h *Handler) Broadcast(msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.connections {
		if err := c.Send(msg); err != nil {
			logger.DebugF("ws: broadcast error to %s: %v", c.ID(), err)
		}
	}
}

// ServeHTTP implements the http.Handler interface. It performs the WebSocket
// upgrade handshake and begins handling the connection.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !isWebSocketUpgrade(r) {
		http.Error(w, "expected websocket upgrade", http.StatusBadRequest)
		return
	}

	// Validate origin
	if h.cfg.checkOrigin != nil {
		origin := r.Header.Get("Origin")
		if !h.cfg.checkOrigin(origin) {
			http.Error(w, "origin not allowed", http.StatusForbidden)
			return
		}
	}

	// Authenticate the handshake before doing anything destructive (hijack).
	var authCtx context.Context
	if h.cfg.upgradeAuth != nil {
		ok, ctx := h.cfg.upgradeAuth(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		authCtx = ctx
	}

	// Get the WebSocket key
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		http.Error(w, "missing Sec-WebSocket-Key", http.StatusBadRequest)
		return
	}

	// Hijack the connection
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "server does not support hijacking", http.StatusInternalServerError)
		return
	}
	netConn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write upgrade response
	acceptKey := computeAcceptKey(key)
	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"
	if _, err := bufrw.WriteString(resp); err != nil {
		_ = netConn.Close()
		return
	}
	if err := bufrw.Flush(); err != nil {
		_ = netConn.Close()
		return
	}

	// Create WebSocket connection
	conn := newConn(netConn, h.cfg, true)
	conn.setContext(authCtx)
	// Use buffered reader from hijack if it has buffered data
	if bufrw.Reader.Buffered() > 0 {
		conn.reader = bufrw.Reader
	}

	// Register connection
	h.addConn(conn)

	// Set up disconnect handler
	conn.onMessage = h.onMessage
	conn.onDisconnect = func(c *Conn, err error) {
		h.removeConn(c)
		if h.onDisconnect != nil {
			h.onDisconnect(c, err)
		}
	}

	// Notify connect handler
	if h.onConnect != nil {
		h.onConnect(conn)
	}

	logger.InfoF("ws: client connected %s from %s", conn.ID(), conn.RemoteAddr())

	// Start read/write pumps
	go conn.pingLoop()
	go conn.readPump()
}

func (h *Handler) addConn(c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[c.id] = c
}

func (h *Handler) removeConn(c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.connections, c.id)
}

// isWebSocketUpgrade checks if the request is a valid WebSocket upgrade request.
func isWebSocketUpgrade(r *http.Request) bool {
	return r.Method == http.MethodGet &&
		headerContains(r.Header, "Connection", "upgrade") &&
		headerContains(r.Header, "Upgrade", "websocket") &&
		r.Header.Get("Sec-WebSocket-Version") == "13"
}

// headerContains checks if a header value contains a specific token (case-insensitive).
func headerContains(h http.Header, key, token string) bool {
	for _, v := range h[http.CanonicalHeaderKey(key)] {
		for _, s := range strings.Split(v, ",") {
			if strings.EqualFold(strings.TrimSpace(s), token) {
				return true
			}
		}
	}
	return false
}

// computeAcceptKey computes the Sec-WebSocket-Accept value per RFC 6455 Section 4.2.2.
func computeAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key))
	h.Write([]byte(websocketGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Upgrade upgrades an HTTP connection to a WebSocket connection without the full Handler.
// This is useful for one-off upgrades outside the Handler pattern.
func Upgrade(w http.ResponseWriter, r *http.Request, opts ...Option) (*Conn, error) {
	if !isWebSocketUpgrade(r) {
		return nil, ErrHandshakeFailed
	}

	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}

	var authCtx context.Context
	if cfg.upgradeAuth != nil {
		ok, ctx := cfg.upgradeAuth(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return nil, ErrUnauthorized
		}
		authCtx = ctx
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		return nil, ErrHandshakeFailed
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("ws: server does not support hijacking")
	}
	netConn, bufrw, err := hj.Hijack()
	if err != nil {
		return nil, fmt.Errorf("ws: hijack failed: %w", err)
	}

	acceptKey := computeAcceptKey(key)
	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"
	if _, err := bufrw.WriteString(resp); err != nil {
		_ = netConn.Close()
		return nil, fmt.Errorf("ws: write handshake response: %w", err)
	}
	if err := bufrw.Flush(); err != nil {
		_ = netConn.Close()
		return nil, fmt.Errorf("ws: flush handshake response: %w", err)
	}

	conn := newConn(netConn, cfg, true)
	conn.setContext(authCtx)
	if bufrw.Reader.Buffered() > 0 {
		conn.reader = bufrw.Reader
	}

	return conn, nil
}
