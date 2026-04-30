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

	"oss.nandlabs.io/golly/clients"
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
// If a CircuitBreaker is configured, it will be checked before connecting.
// If a RetryPolicy or RetryInfo is configured, failed connections will be retried.
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

	if err := c.connectWithRetry(ctx); err != nil {
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

// connectWithRetry wraps connect with circuit breaker and retry logic.
func (c *Client) connectWithRetry(ctx context.Context) error {
	// Check circuit breaker
	if c.cfg.circuitBreaker != nil {
		if err := c.cfg.circuitBreaker.CanExecute(); err != nil {
			return fmt.Errorf("ws: circuit breaker open: %w", err)
		}
	}

	err := c.connect(ctx)

	// Report to circuit breaker
	if c.cfg.circuitBreaker != nil {
		c.cfg.circuitBreaker.OnExecution(err == nil)
	}

	if err == nil {
		return nil
	}

	// Retry with RetryInfo (enhanced)
	if c.cfg.retryInfo != nil {
		return c.retryConnect(ctx, err, c.retryInfoWait)
	}

	// Retry with RetryPolicy (legacy)
	if c.cfg.retryPolicy != nil {
		return c.retryConnect(ctx, err, c.retryPolicyWait)
	}

	return err
}

// retryWaitFunc returns the wait duration for a given retry count.
type retryWaitFunc func(int) time.Duration

func (c *Client) retryInfoWait(attempt int) time.Duration {
	return c.cfg.retryInfo.WaitTime(attempt)
}

func (c *Client) retryPolicyWait(attempt int) time.Duration {
	return c.cfg.retryPolicy.WaitTime(attempt)
}

func (c *Client) retryConnect(ctx context.Context, lastErr error, waitFn retryWaitFunc) error {
	maxRetries := 0
	if c.cfg.retryInfo != nil {
		maxRetries = c.cfg.retryInfo.MaxRetries
	} else if c.cfg.retryPolicy != nil {
		maxRetries = c.cfg.retryPolicy.MaxRetries
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		wait := waitFn(attempt)
		logger.DebugF("ws: retry %d/%d after %v (last error: %v)", attempt+1, maxRetries, wait, lastErr)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}

		// Check circuit breaker before retry
		if c.cfg.circuitBreaker != nil {
			if err := c.cfg.circuitBreaker.CanExecute(); err != nil {
				return fmt.Errorf("ws: circuit breaker open during retry: %w", err)
			}
		}

		err := c.connect(ctx)

		// Report to circuit breaker
		if c.cfg.circuitBreaker != nil {
			c.cfg.circuitBreaker.OnExecution(err == nil)
		}

		if err == nil {
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("ws: connection failed after %d retries: %w", maxRetries, lastErr)
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

	var handshakeBuf strings.Builder
	fmt.Fprintf(&handshakeBuf, "GET %s HTTP/1.1\r\n", path)
	fmt.Fprintf(&handshakeBuf, "Host: %s\r\n", u.Host)
	handshakeBuf.WriteString("Upgrade: websocket\r\n")
	handshakeBuf.WriteString("Connection: Upgrade\r\n")
	fmt.Fprintf(&handshakeBuf, "Sec-WebSocket-Key: %s\r\n", wsKey)
	handshakeBuf.WriteString("Sec-WebSocket-Version: 13\r\n")

	// Add authentication headers
	if c.cfg.auth != nil {
		if err := writeAuthHeaders(&handshakeBuf, c.cfg.auth); err != nil {
			_ = netConn.Close()
			return fmt.Errorf("ws: auth header failed: %w", err)
		}
	}

	// Add custom headers
	for k, v := range c.cfg.headers {
		fmt.Fprintf(&handshakeBuf, "%s: %s\r\n", k, v)
	}

	handshakeBuf.WriteString("\r\n")

	if _, err := netConn.Write([]byte(handshakeBuf.String())); err != nil {
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

		// Check circuit breaker before reconnect
		if c.cfg.circuitBreaker != nil {
			if err := c.cfg.circuitBreaker.CanExecute(); err != nil {
				logger.DebugF("ws: circuit breaker open, skipping reconnect")
				wait *= 2
				if wait > c.cfg.maxReconnectWait {
					wait = c.cfg.maxReconnectWait
				}
				continue
			}
		}

		logger.InfoF("ws: attempting reconnection to %s", c.url)
		err := c.connect(context.Background())

		// Report to circuit breaker
		if c.cfg.circuitBreaker != nil {
			c.cfg.circuitBreaker.OnExecution(err == nil)
		}

		if err != nil {
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

// apiKeyHeaderer is an optional interface for auth providers that supply a custom header name.
type apiKeyHeaderer interface {
	HeaderName() string
}

// writeAuthHeaders writes authentication headers to the handshake request buffer.
func writeAuthHeaders(buf *strings.Builder, auth clients.AuthProvider) error {
	switch auth.Type() {
	case clients.AuthTypeBasic:
		user, err := auth.User()
		if err != nil {
			return err
		}
		pass, err := auth.Pass()
		if err != nil {
			return err
		}
		creds := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
		fmt.Fprintf(buf, "Authorization: Basic %s\r\n", creds)
	case clients.AuthTypeBearer:
		token, err := auth.Token()
		if err != nil {
			return err
		}
		fmt.Fprintf(buf, "Authorization: Bearer %s\r\n", token)
	case clients.AuthTypeAPIKey:
		token, err := auth.Token()
		if err != nil {
			return err
		}
		headerName := "X-API-Key"
		if h, ok := auth.(apiKeyHeaderer); ok && h.HeaderName() != "" {
			headerName = h.HeaderName()
		}
		fmt.Fprintf(buf, "%s: %s\r\n", headerName, token)
	}
	return nil
}
