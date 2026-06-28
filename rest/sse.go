package rest

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SSEEvent is a single parsed Server-Sent Event.
type SSEEvent struct {
	// ID is the event id (echoed back as Last-Event-ID on reconnect).
	ID string
	// Type is the event type (defaults to "message" if not supplied).
	Type string
	// Data is the raw payload — for multi-line events the lines are joined
	// with '\n' as per the SSE spec.
	Data []byte
	// Retry, when non-zero, indicates the server-requested reconnect delay.
	Retry time.Duration
}

// SSEStream consumes a text/event-stream response, yielding one event per
// Next(). Close releases the underlying body. SSEStream is not safe for
// concurrent Next() calls from multiple goroutines.
type SSEStream struct {
	body    io.ReadCloser
	scanner *bufio.Scanner

	// In-flight event being assembled across consecutive lines.
	curID    string
	curType  string
	curData  []byte
	curRetry time.Duration

	// lastEventID is the id of the last fully-delivered event; on
	// auto-reconnect it's sent back as Last-Event-ID.
	lastEventID string
}

// Next blocks until the next event is fully assembled, the stream ends
// (io.EOF), or ctx is canceled. The returned SSEEvent is owned by the
// caller; the internal buffer is reset for the next event.
//
// Pass a fresh request context — typically the same one used for Stream().
func (s *SSEStream) Next(ctx context.Context) (*SSEEvent, error) {
	if s == nil || s.scanner == nil {
		return nil, io.EOF
	}
	for {
		// Honor cancellation between lines (Scanner is blocking).
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if !s.scanner.Scan() {
			// End-of-stream — flush any pending event, then return EOF.
			if err := s.scanner.Err(); err != nil {
				return nil, err
			}
			if len(s.curData) > 0 || s.curID != "" || s.curType != "" {
				ev := s.flush()
				return ev, nil
			}
			return nil, io.EOF
		}
		line := s.scanner.Text()
		// Empty line terminates an event.
		if line == "" {
			if len(s.curData) > 0 || s.curID != "" || s.curType != "" || s.curRetry > 0 {
				return s.flush(), nil
			}
			continue
		}
		// Comments (lines starting with ':') are ignored per the spec.
		if strings.HasPrefix(line, ":") {
			continue
		}
		field, value := splitField(line)
		switch field {
		case "id":
			s.curID = value
			s.lastEventID = value
		case "event":
			s.curType = value
		case "data":
			if len(s.curData) > 0 {
				s.curData = append(s.curData, '\n')
			}
			s.curData = append(s.curData, value...)
		case "retry":
			if ms, err := strconv.Atoi(value); err == nil && ms > 0 {
				s.curRetry = time.Duration(ms) * time.Millisecond
			}
		}
	}
}

// Close releases the underlying response body. Safe to call multiple times.
func (s *SSEStream) Close() error {
	if s == nil || s.body == nil {
		return nil
	}
	err := s.body.Close()
	s.body = nil
	return err
}

// LastEventID returns the id of the most recently fully-delivered event.
// Useful when resuming a stream manually.
func (s *SSEStream) LastEventID() string {
	if s == nil {
		return ""
	}
	return s.lastEventID
}

// flush builds the SSEEvent from the assembled fields and resets state.
func (s *SSEStream) flush() *SSEEvent {
	ev := &SSEEvent{
		ID:    s.curID,
		Type:  s.curType,
		Data:  s.curData,
		Retry: s.curRetry,
	}
	if ev.Type == "" {
		ev.Type = "message"
	}
	s.curID = ""
	s.curType = ""
	s.curData = nil
	s.curRetry = 0
	return ev
}

// splitField splits an SSE line into (field, value). Per the spec, the
// optional leading space after the ':' is stripped.
func splitField(line string) (string, string) {
	idx := strings.IndexByte(line, ':')
	if idx < 0 {
		return line, ""
	}
	field := line[:idx]
	value := line[idx+1:]
	value = strings.TrimPrefix(value, " ")
	return field, value
}

// ErrNotEventStream is returned by Stream when the response's Content-Type
// is not text/event-stream.
var ErrNotEventStream = errors.New("rest/sse: response is not text/event-stream")

// Stream sends req via c and, on a successful streaming response, returns
// an SSEStream over the response body. The caller MUST Close() the stream
// to free the underlying connection.
//
// The response's Content-Type must start with "text/event-stream"; otherwise
// the body is closed and ErrNotEventStream is returned.
func (c *Client) Stream(_ context.Context, req *Request) (*SSEStream, error) {
	if req == nil {
		return nil, fmt.Errorf("rest/sse: request is nil")
	}
	// Accept-friendly nudge for servers that content-negotiate.
	req.AddHeader("Accept", "text/event-stream")

	resp, err := c.Execute(req)
	if err != nil {
		return nil, fmt.Errorf("rest/sse: execute: %w", err)
	}
	raw := resp.Raw()
	if raw == nil || raw.Body == nil {
		return nil, fmt.Errorf("rest/sse: empty response")
	}
	if !resp.IsSuccess() {
		_ = raw.Body.Close()
		return nil, fmt.Errorf("rest/sse: status %d", resp.StatusCode())
	}
	if ct := raw.Header.Get("Content-Type"); !strings.HasPrefix(strings.ToLower(ct), "text/event-stream") {
		_ = raw.Body.Close()
		return nil, fmt.Errorf("%w (got %q)", ErrNotEventStream, ct)
	}
	scanner := bufio.NewScanner(raw.Body)
	// SSE lines can be up to whatever the server sends; raise the cap
	// generously so giant payloads don't trip Scanner's default 64KB limit.
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	return &SSEStream{body: raw.Body, scanner: scanner}, nil
}

// ensure http is imported (we declare ErrNotEventStream above; http used by Raw).
var _ = http.MethodGet
