package ws

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Client represents a WebSocket client connection.
type Client struct {
	cfg          *config
	url          string
	conn         *Conn
	mu           sync.Mutex
	messages     chan Message
	done         chan struct{}
	onMessage    MessageHandler
	onConnect    func()
	onDisconnect func(error)
	closed       bool
}

// Dial creates a new WebSocket client connection to the given URL.
// The URL scheme must be "ws" or "wss".
func Dial(ctx context.Context, rawURL string, opts ...Option) (*Client, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}

	c := &Client{
		cfg:      cfg,
		url:      rawURL,
		messages: make(chan Message, 64),
		done:     make(chan struct{}),
	}

	if err := c.connect(ctx); err != nil {
		return nil, err
	}

	return c, nil
}

// OnMessage registers a handler for incoming messages.
func (c *Client) OnMessage(fn MessageHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onMessage = fn
	if c.conn != nil {
		c.conn.onMessage = func(conn *Conn, msg Message) {
			fn(conn, msg)
			select {
			case c.messages <- msg:
			default:
			}
		}
	}
}

// OnConnect registers a callback for when a connection is established.
func (c *Client) OnConnect(fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onConnect = fn
}

// OnDisconnect registers a callback for when the connection is lost.
func (c *Client) OnDisconnect(fn func(error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onDisconnect = fn
}

// Send sends a message to the server.
func (c *Client) Send(msg Message) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return ErrConnClosed
	}
	return conn.Send(msg)
}

// Messages returns a channel that receives incoming messages.
func (c *Client) Messages() <-chan Message {
	return c.messages
}

// Conn returns the underlying WebSocket connection.
func (c *Client) Conn() *Conn {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn
}

// Close closes the client connection and stops any reconnection attempts.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return ErrConnClosed
	}
	c.closed = true
	close(c.done)
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) connect(ctx context.Context) error {
	u, err := url.Parse(c.url)
	if err != nil {
		return ErrInvalidURL
	}

	var scheme string
	switch u.Scheme {
	case "ws":
		scheme = "tcp"
	case "wss":
		scheme = "tcp"
	default:
		return ErrInvalidURL
	}

	host := u.Host
	if !strings.Contains(host, ":") {
		if u.Scheme == "wss" {
			host += ":443"
		} else {
			host += ":80"
		}
	}

	// Dial with timeout
	dialer := &net.Dialer{Timeout: c.cfg.handshakeTimeout}
	var netConn net.Conn
	if u.Scheme == "wss" {
		tlsCfg := c.cfg.tlsConfig
		if tlsCfg == nil {
			tlsCfg = &tls.Config{}
		}
		if tlsCfg.ServerName == "" {
			hostname := u.Hostname()
			tlsCfg = tlsCfg.Clone()
			tlsCfg.ServerName = hostname
		}
		netConn, err = tls.DialWithDialer(dialer, scheme, host, tlsCfg)
	} else {
		netConn, err = dialer.DialContext(ctx, scheme, host)
	}
	if err != nil {
		return fmt.Errorf("ws: dial failed: %w", err)
	}

	// Perform client handshake
	wsKey := generateKey()
	path := u.RequestURI()
	if path == "" {
		path = "/"
	}

	handshake := fmt.Sprintf("GET %s HTTP/1.1\r\n"+
		"Host: %s\r\n"+
		"Upgrade: websocket\r\n"+
		"Connection: Upgrade\r\n"+
		"Sec-WebSocket-Key: %s\r\n"+
		"Sec-WebSocket-Version: 13\r\n\r\n",
		path, u.Host, wsKey)

	if _, err := netConn.Write([]byte(handshake)); err != nil {
		_ = netConn.Close()
		return fmt.Errorf("ws: handshake write failed: %w", err)
	}

	// Read handshake response
	br := bufio.NewReaderSize(netConn, c.cfg.readBufferSize)
	if err := c.readHandshakeResponse(br, wsKey); err != nil {
		_ = netConn.Close()
		return err
	}

	conn := newConn(netConn, c.cfg, false)
	conn.reader = br

	c.mu.Lock()
	c.conn = conn
	onMsg := c.onMessage
	onConnect := c.onConnect
	c.mu.Unlock()

	// Set up message handler that feeds the channel
	conn.onMessage = func(wsConn *Conn, msg Message) {
		if onMsg != nil {
			onMsg(wsConn, msg)
		}
		select {
		case c.messages <- msg:
		default:
		}
	}

	// Set up disconnect handler
	conn.onDisconnect = func(wsConn *Conn, err error) {
		c.mu.Lock()
		onDisconnect := c.onDisconnect
		shouldReconnect := c.cfg.autoReconnect && !c.closed
		c.conn = nil
		c.mu.Unlock()

		if onDisconnect != nil {
			onDisconnect(err)
		}

		if shouldReconnect {
			go c.reconnectLoop()
		}
	}

	if onConnect != nil {
		onConnect()
	}

	logger.InfoF("ws: connected to %s", c.url)

	go conn.pingLoop()
	go conn.readPump()

	return nil
}

func (c *Client) readHandshakeResponse(br *bufio.Reader, wsKey string) error {
	// Read status line
	line, err := br.ReadString('\n')
	if err != nil {
		return fmt.Errorf("ws: reading status line: %w", err)
	}
	if !strings.Contains(line, "101") {
		return ErrHandshakeFailed
	}

	// Read headers
	expectedAccept := computeAcceptKey(wsKey)
	gotAccept := false
	for {
		line, err = br.ReadString('\n')
		if err != nil {
			return fmt.Errorf("ws: reading headers: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(strings.ToLower(line), "sec-websocket-accept:") {
			val := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			if val == expectedAccept {
				gotAccept = true
			}
		}
	}

	if !gotAccept {
		return ErrHandshakeFailed
	}
	return nil
}

func (c *Client) reconnectLoop() {
	wait := time.Second
	for {
		select {
		case <-c.done:
			return
		case <-time.After(wait):
		}

		c.mu.Lock()
		if c.closed {
			c.mu.Unlock()
			return
		}
		c.mu.Unlock()

		logger.InfoF("ws: attempting reconnection to %s", c.url)
		if err := c.connect(context.Background()); err != nil {
			logger.DebugF("ws: reconnect failed: %v", err)
			// Exponential backoff
			wait *= 2
			if wait > c.cfg.maxReconnectWait {
				wait = c.cfg.maxReconnectWait
			}
			continue
		}
		logger.InfoF("ws: reconnected to %s", c.url)
		return
	}
}

// generateKey generates a random 16-byte base64-encoded WebSocket key.
func generateKey() string {
	key := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		// Fallback (should never happen)
		return base64.StdEncoding.EncodeToString([]byte("golly-ws-key1234"))
	}
	return base64.StdEncoding.EncodeToString(key)
}
