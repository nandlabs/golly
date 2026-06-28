package ws

import (
	"sync"
)

// Hub is an optional connection registry that maps arbitrary string keys
// (e.g. user ids, room names, conversation ids) to *Conn, with targeted
// fan-out via Send. It is safe for concurrent use.
//
// A Hub does not replace Handler.Broadcast — use Broadcast for "everybody",
// Hub for "this user's devices" or "this conversation."
//
// Typical usage:
//
//	hub := ws.NewHub()
//	handler.OnConnect(func(c *ws.Conn) {
//	    userID := userIDFromContext(c.Context())
//	    hub.Add(c, userID, "room:"+roomFromQuery(c))
//	})
//	handler.OnDisconnect(func(c *ws.Conn, _ error) {
//	    hub.Remove(c)
//	})
//	// Later, deliver a private message:
//	hub.Send("user-42", ws.NewTextMessage([]byte(`{"new":"message"}`)))
type Hub struct {
	mu     sync.RWMutex
	byKey  map[string]map[*Conn]struct{} // key -> set of conns
	byConn map[*Conn]map[string]struct{} // conn -> set of keys (for fast Remove)
}

// NewHub returns an empty, ready-to-use Hub.
func NewHub() *Hub {
	return &Hub{
		byKey:  make(map[string]map[*Conn]struct{}),
		byConn: make(map[*Conn]map[string]struct{}),
	}
}

// Add associates a connection with one or more routing keys. Calling Add again
// with the same conn and overlapping keys is a no-op for those keys; new keys
// are appended (the conn ends up in the union of all keys it was Added with).
// Passing zero keys still registers the conn — useful so a later Remove cleans
// it up cleanly even if no keys were associated.
func (h *Hub) Add(c *Conn, keys ...string) {
	if c == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.byConn[c]; !ok {
		h.byConn[c] = make(map[string]struct{})
	}
	for _, k := range keys {
		if _, ok := h.byKey[k]; !ok {
			h.byKey[k] = make(map[*Conn]struct{})
		}
		h.byKey[k][c] = struct{}{}
		h.byConn[c][k] = struct{}{}
	}
}

// Remove drops a connection from every key it was associated with. Safe to
// call multiple times; a conn that isn't registered is a no-op.
func (h *Hub) Remove(c *Conn) {
	if c == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	keys, ok := h.byConn[c]
	if !ok {
		return
	}
	for k := range keys {
		if set, ok := h.byKey[k]; ok {
			delete(set, c)
			if len(set) == 0 {
				delete(h.byKey, k)
			}
		}
	}
	delete(h.byConn, c)
}

// Send delivers msg to every connection associated with key and returns the
// number of successful deliveries. Errors writing to individual conns are
// logged at debug level and do not abort the fan-out — one slow / dead conn
// must not block delivery to the others.
func (h *Hub) Send(key string, msg Message) int {
	// Snapshot under the read lock then send outside it so that:
	//  (a) a long-running Send can't block concurrent Add/Remove, and
	//  (b) a handler that calls Hub.Remove inside its OnDisconnect (which a
	//      failing Send may trigger) doesn't deadlock.
	h.mu.RLock()
	set, ok := h.byKey[key]
	if !ok || len(set) == 0 {
		h.mu.RUnlock()
		return 0
	}
	conns := make([]*Conn, 0, len(set))
	for c := range set {
		conns = append(conns, c)
	}
	h.mu.RUnlock()

	delivered := 0
	for _, c := range conns {
		if err := c.Send(msg); err != nil {
			logger.DebugF("ws/hub: send to %s on key %q: %v", c.ID(), key, err)
			continue
		}
		delivered++
	}
	return delivered
}

// Rooms returns the routing keys a connection is associated with. The returned
// slice is a snapshot and may be modified by the caller without affecting the
// Hub. Order is unspecified.
func (h *Hub) Rooms(c *Conn) []string {
	if c == nil {
		return nil
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	keys, ok := h.byConn[c]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(keys))
	for k := range keys {
		out = append(out, k)
	}
	return out
}

// Len returns the number of connections currently registered in the Hub
// (regardless of key membership).
func (h *Hub) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.byConn)
}

// CountKey returns the number of connections currently associated with key.
func (h *Hub) CountKey(key string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.byKey[key])
}
