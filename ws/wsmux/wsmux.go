// Package wsmux layers a small JSON envelope and per-channel handler routing
// over the raw byte streams that ws.Conn exposes.
//
// The envelope is intentionally minimal:
//
//	{
//	  "type":    "string, e.g. \"msg\" or \"req\"",   // free-form, app decides
//	  "channel": "string routing key",                 // mandatory; handler is keyed by this
//	  "id":      "string correlation id (optional)",   // present on request/reply pairs
//	  "payload": <any JSON value>                      // app-specific body
//	}
//
// Apps can use it as-is for request/reply or pub/sub-style multiplexing over
// a single WebSocket without writing their own framing.
package wsmux

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"oss.nandlabs.io/golly/ws"
)

// Envelope is the wire shape every message in a wsmux conversation uses.
// Payload is left as json.RawMessage so handlers can decode it into the
// type they expect without an extra round-trip through any.
type Envelope struct {
	Type    string          `json:"type,omitempty"`
	Channel string          `json:"channel"`
	ID      string          `json:"id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Handler receives an envelope addressed to its channel and may reply via
// Replier. ctx is the ws.Conn's context (carries auth, etc.).
type Handler func(env *Envelope, r Replier)

// Replier sends a response to the originating peer over the same connection,
// preserving the request's id so the requester can correlate.
type Replier interface {
	// Reply sends a typed response with the same id as the inbound envelope.
	// Useful for request/reply patterns.
	Reply(typ string, payload any) error
	// Send sends a brand-new envelope (caller-chosen channel/type/id/payload)
	// on the same connection. Useful for unsolicited messages.
	Send(env Envelope) error
	// Conn returns the underlying *ws.Conn.
	Conn() *ws.Conn
}

// ErrUnknownChannel is returned (and surfaced via OnUnknown) when an inbound
// envelope's channel has no registered handler.
var ErrUnknownChannel = errors.New("wsmux: unknown channel")

// Mux routes incoming WebSocket messages to handlers by envelope.Channel.
// It is safe to register handlers concurrently. Routing itself happens on the
// ws.Conn's read goroutine, so handlers should return quickly or dispatch
// long work to their own goroutine.
type Mux struct {
	mu        sync.RWMutex
	handlers  map[string]Handler
	onUnknown func(c *ws.Conn, env *Envelope, err error)
	onDecode  func(c *ws.Conn, raw []byte, err error)
}

// New returns an empty Mux.
func New() *Mux {
	return &Mux{
		handlers: make(map[string]Handler),
	}
}

// Handle registers a handler for envelopes whose Channel equals channel.
// Registering the same channel twice overwrites the prior handler.
func (m *Mux) Handle(channel string, h Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[channel] = h
}

// HandleFunc is a convenience for Handle with a plain function value.
func (m *Mux) HandleFunc(channel string, fn func(env *Envelope, r Replier)) {
	m.Handle(channel, Handler(fn))
}

// OnUnknown registers a fallback invoked when an envelope arrives on a channel
// with no registered handler. If nil (default) the envelope is silently dropped.
func (m *Mux) OnUnknown(fn func(c *ws.Conn, env *Envelope, err error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onUnknown = fn
}

// OnDecodeError registers a fallback invoked when an inbound message is not
// valid JSON or doesn't satisfy the envelope schema. If nil the message is
// silently dropped.
func (m *Mux) OnDecodeError(fn func(c *ws.Conn, raw []byte, err error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onDecode = fn
}

// ServeMessage decodes msg as an Envelope and dispatches it to the registered
// handler. Plug this into ws.Handler.OnMessage:
//
//	handler.OnMessage(mux.ServeMessage)
func (m *Mux) ServeMessage(c *ws.Conn, msg ws.Message) {
	if msg.Type != ws.OpText {
		// Envelopes are JSON; binary frames are not for this mux.
		if cb := m.decodeErrCallback(); cb != nil {
			cb(c, msg.Data, fmt.Errorf("wsmux: non-text frame (opcode %d)", msg.Type))
		}
		return
	}
	env, err := decode(msg.Data)
	if err != nil {
		if cb := m.decodeErrCallback(); cb != nil {
			cb(c, msg.Data, err)
		}
		return
	}

	m.mu.RLock()
	h, ok := m.handlers[env.Channel]
	unknown := m.onUnknown
	m.mu.RUnlock()

	if !ok {
		if unknown != nil {
			unknown(c, env, ErrUnknownChannel)
		}
		return
	}

	h(env, &replier{conn: c, inbound: env})
}

// decodeErrCallback returns the current OnDecodeError callback under the
// read lock. Pulled out to keep ServeMessage tidy.
func (m *Mux) decodeErrCallback() func(c *ws.Conn, raw []byte, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.onDecode
}

// SendEnvelope is a helper that JSON-encodes env and writes it to c as a text
// frame. Use it for unsolicited server-pushed messages.
func SendEnvelope(c *ws.Conn, env Envelope) error {
	b, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("wsmux: marshal envelope: %w", err)
	}
	return c.Send(ws.NewTextMessage(b))
}

// decode parses raw JSON into an Envelope and validates that Channel is set.
func decode(raw []byte) (*Envelope, error) {
	env := &Envelope{}
	if err := json.Unmarshal(raw, env); err != nil {
		return nil, fmt.Errorf("wsmux: invalid envelope JSON: %w", err)
	}
	if env.Channel == "" {
		return nil, errors.New("wsmux: envelope missing channel")
	}
	return env, nil
}

// replier is the default Replier implementation. It writes responses on the
// same ws.Conn that delivered the inbound envelope.
type replier struct {
	conn    *ws.Conn
	inbound *Envelope
}

func (r *replier) Conn() *ws.Conn { return r.conn }

func (r *replier) Reply(typ string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("wsmux: marshal reply payload: %w", err)
	}
	return SendEnvelope(r.conn, Envelope{
		Type:    typ,
		Channel: r.inbound.Channel,
		ID:      r.inbound.ID,
		Payload: body,
	})
}

func (r *replier) Send(env Envelope) error {
	return SendEnvelope(r.conn, env)
}
