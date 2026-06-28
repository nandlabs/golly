package turbo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

// SSEWriter is a thin wrapper around http.ResponseWriter that emits
// Server-Sent Events. Construct one per request via NewSSE; it sets the
// proper headers and wraps the underlying http.Flusher so each Send is
// immediately pushed to the client.
//
// Typical usage:
//
//	router.Get("/events", func(w http.ResponseWriter, r *http.Request) {
//	    sse, err := turbo.NewSSE(w)
//	    if err != nil { /* handler doesn't support Flusher */ return }
//	    defer sse.Close()
//	    for token := range tokens(r.Context()) {
//	        if err := sse.Send("token", token); err != nil { return }
//	    }
//	})
type SSEWriter struct {
	mu      sync.Mutex
	w       http.ResponseWriter
	flusher http.Flusher
	closed  bool
}

// ErrUnflushable is returned by NewSSE when the response writer doesn't
// implement http.Flusher (e.g. wrapped in a buffering middleware).
var ErrUnflushable = errors.New("turbo/sse: response writer does not support flushing")

// NewSSE wraps w as an SSEWriter. It sets headers (Content-Type:
// text/event-stream, Cache-Control: no-cache, Connection: keep-alive,
// X-Accel-Buffering: no) and writes them immediately so proxies don't
// buffer the stream.
func NewSSE(w http.ResponseWriter) (*SSEWriter, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, ErrUnflushable
	}
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no") // disable nginx proxy buffering
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	return &SSEWriter{w: w, flusher: flusher}, nil
}

// Send emits an SSE event. event is the event name (empty for default).
// data is JSON-marshaled; strings/[]byte are written verbatim to preserve
// caller formatting (useful for line-multiline payloads).
//
// Wire format per event:
//
//	event: <event>\n
//	data:  <line1>\n
//	data:  <line2>\n
//	\n
func (s *SSEWriter) Send(event string, data any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("turbo/sse: writer closed")
	}
	if event != "" {
		if _, err := fmt.Fprintf(s.w, "event: %s\n", event); err != nil {
			return err
		}
	}
	body, err := dataToString(data)
	if err != nil {
		return err
	}
	// SSE: each line of data must be prefixed with "data: ".
	for line := range splitLines(body) {
		if _, err := fmt.Fprintf(s.w, "data: %s\n", line); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprint(s.w, "\n"); err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

// Flush forces a flush of the underlying writer. Send already flushes after
// each event; this is for use after manual writes via Raw().
func (s *SSEWriter) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		s.flusher.Flush()
	}
}

// Close marks the writer closed and emits a terminal "event: close" so
// well-behaved EventSource clients stop reconnecting.
func (s *SSEWriter) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	_, _ = fmt.Fprint(s.w, "event: close\ndata: bye\n\n")
	s.flusher.Flush()
	s.closed = true
	return nil
}

// dataToString renders data for an SSE data: line. string and []byte pass
// through; anything else is JSON-marshaled.
func dataToString(data any) (string, error) {
	switch x := data.(type) {
	case nil:
		return "", nil
	case string:
		return x, nil
	case []byte:
		return string(x), nil
	}
	b, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("turbo/sse: marshal data: %w", err)
	}
	return string(b), nil
}

// splitLines yields lines of s without trailing \n.
func splitLines(s string) func(yield func(string) bool) {
	return func(yield func(string) bool) {
		start := 0
		for i := 0; i < len(s); i++ {
			if s[i] == '\n' {
				if !yield(s[start:i]) {
					return
				}
				start = i + 1
			}
		}
		if start <= len(s) {
			if !yield(s[start:]) {
				return
			}
		}
	}
}
