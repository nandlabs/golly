package ws

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// connID is an atomic counter for generating unique connection IDs.
var connID atomic.Uint64

// MessageHandler is called when a message is received on a connection.
type MessageHandler func(conn *Conn, msg Message)

// ConnHandler is called on connection lifecycle events.
type ConnHandler func(conn *Conn)

// DisconnectHandler is called when a connection is closed.
type DisconnectHandler func(conn *Conn, err error)

// Conn represents a WebSocket connection.
type Conn struct {
	id       string
	conn     net.Conn
	reader   *bufio.Reader
	writeMu  sync.Mutex
	cfg      *config
	closed   atomic.Bool
	closeCh  chan struct{}
	isServer bool

	onMessage    MessageHandler
	onDisconnect DisconnectHandler
}

// newConn wraps a net.Conn into a WebSocket Conn.
func newConn(netConn net.Conn, cfg *config, isServer bool) *Conn {
	id := connID.Add(1)
	return &Conn{
		id:       fmt.Sprintf("ws-%d", id),
		conn:     netConn,
		reader:   bufio.NewReaderSize(netConn, cfg.readBufferSize),
		cfg:      cfg,
		closeCh:  make(chan struct{}),
		isServer: isServer,
	}
}

// ID returns the unique identifier for this connection.
func (c *Conn) ID() string {
	return c.id
}

// RemoteAddr returns the remote network address of the peer.
func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// Send sends a message to the peer. It is safe to call from multiple goroutines.
func (c *Conn) Send(msg Message) error {
	if c.closed.Load() {
		return ErrConnClosed
	}
	return c.writeMessage(msg.Type, msg.Data)
}

// Close sends a close frame and closes the underlying connection.
func (c *Conn) Close() error {
	return c.closeWithCode(CloseNormal, "")
}

// closeWithCode sends a close frame with the given code and reason.
func (c *Conn) closeWithCode(code CloseCode, reason string) error {
	if !c.closed.CompareAndSwap(false, true) {
		return ErrConnClosed
	}
	close(c.closeCh)

	// Best-effort close frame
	payload := makeClosePayload(code, reason)
	_ = c.writeMessage(OpClose, payload)

	return c.conn.Close()
}

// writeMessage writes a single-frame message to the connection.
func (c *Conn) writeMessage(opcode Opcode, data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if c.cfg.writeTimeout > 0 {
		if err := c.conn.SetWriteDeadline(time.Now().Add(c.cfg.writeTimeout)); err != nil {
			return err
		}
	}

	f := &frame{
		fin:     true,
		opcode:  opcode,
		masked:  !c.isServer, // clients MUST mask, servers MUST NOT mask
		payload: data,
	}
	return writeFrame(c.conn, f)
}

// readPump reads frames from the connection and dispatches messages.
func (c *Conn) readPump() {
	defer func() {
		var disconnectErr error
		if r := recover(); r != nil {
			disconnectErr = fmt.Errorf("ws: panic in read pump: %v", r)
		}
		if c.onDisconnect != nil {
			c.onDisconnect(c, disconnectErr)
		}
		_ = c.closeWithCode(CloseAbnormal, "")
	}()

	var msgType Opcode
	var msgBuf []byte

	for {
		f, err := readFrame(c.reader, c.cfg.maxMessageSize)
		if err != nil {
			if c.closed.Load() {
				return
			}
			if err == io.EOF || isNetClose(err) {
				return
			}
			logger.DebugF("ws: read error on %s: %v", c.id, err)
			return
		}

		switch f.opcode {
		case OpClose:
			// Respond with close frame
			_ = c.closeWithCode(CloseNormal, "")
			return

		case OpPing:
			if err := c.writeMessage(OpPong, f.payload); err != nil {
				logger.DebugF("ws: pong write error on %s: %v", c.id, err)
				return
			}

		case OpPong:
			// Reset pong deadline handled by ping loop

		case OpText, OpBinary:
			if f.fin {
				// Complete single-frame message
				if c.onMessage != nil {
					c.onMessage(c, Message{Type: f.opcode, Data: f.payload})
				}
			} else {
				// Start of fragmented message
				msgType = f.opcode
				msgBuf = append([]byte{}, f.payload...)
			}

		case OpContinuation:
			msgBuf = append(msgBuf, f.payload...)
			// Check accumulated size
			if c.cfg.maxMessageSize > 0 && int64(len(msgBuf)) > c.cfg.maxMessageSize {
				_ = c.closeWithCode(CloseMessageTooBig, "message too large")
				return
			}
			if f.fin {
				if c.onMessage != nil {
					c.onMessage(c, Message{Type: msgType, Data: msgBuf})
				}
				msgBuf = nil
			}

		default:
			_ = c.closeWithCode(CloseProtocolError, "unknown opcode")
			return
		}
	}
}

// pingLoop sends periodic ping frames to the peer.
func (c *Conn) pingLoop() {
	if c.cfg.pingInterval <= 0 {
		return
	}
	ticker := time.NewTicker(c.cfg.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.writeMessage(OpPing, nil); err != nil {
				logger.DebugF("ws: ping error on %s: %v", c.id, err)
				return
			}
		case <-c.closeCh:
			return
		}
	}
}

// isNetClose returns true if the error indicates a closed network connection.
func isNetClose(err error) bool {
	if err == nil {
		return false
	}
	// Check for net.OpError with "use of closed network connection"
	if opErr, ok := err.(*net.OpError); ok {
		return opErr.Err.Error() == "use of closed network connection"
	}
	return false
}
