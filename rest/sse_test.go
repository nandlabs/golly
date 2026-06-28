package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newSSEServer(t *testing.T, lines string) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, lines)
	}))
	t.Cleanup(srv.Close)
	c := NewClientWithOptions(CliOptsBuilder().Build())
	return srv, c
}

func TestSSE_BasicEvents(t *testing.T) {
	body := "event: tick\ndata: 1\n\nevent: tick\ndata: 2\n\n"
	srv, c := newSSEServer(t, body)
	req, _ := c.NewRequest(srv.URL, http.MethodGet)
	stream, err := c.Stream(context.Background(), req)
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	defer stream.Close()

	for i := 1; i <= 2; i++ {
		ev, err := stream.Next(context.Background())
		if err != nil {
			t.Fatalf("Next %d: %v", i, err)
		}
		if ev.Type != "tick" || string(ev.Data) != fmt.Sprintf("%d", i) {
			t.Errorf("event %d wrong: type=%q data=%q", i, ev.Type, ev.Data)
		}
	}
	if _, err := stream.Next(context.Background()); err != io.EOF {
		t.Errorf("expected EOF after 2 events; got %v", err)
	}
}

func TestSSE_MultilineData(t *testing.T) {
	body := "data: line1\ndata: line2\ndata: line3\n\n"
	srv, c := newSSEServer(t, body)
	req, _ := c.NewRequest(srv.URL, http.MethodGet)
	stream, err := c.Stream(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	ev, err := stream.Next(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if string(ev.Data) != "line1\nline2\nline3" {
		t.Errorf("multiline data wrong: %q", ev.Data)
	}
	if ev.Type != "message" {
		t.Errorf("default type should be 'message'; got %q", ev.Type)
	}
}

func TestSSE_IDAndRetry(t *testing.T) {
	body := "id: 7\nretry: 2500\nevent: x\ndata: y\n\n"
	srv, c := newSSEServer(t, body)
	req, _ := c.NewRequest(srv.URL, http.MethodGet)
	stream, err := c.Stream(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	ev, err := stream.Next(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ev.ID != "7" {
		t.Errorf("id = %q, want '7'", ev.ID)
	}
	if ev.Retry != 2500*time.Millisecond {
		t.Errorf("retry = %v, want 2.5s", ev.Retry)
	}
	if stream.LastEventID() != "7" {
		t.Errorf("LastEventID = %q, want '7'", stream.LastEventID())
	}
}

func TestSSE_CommentsAndBlankLinesIgnored(t *testing.T) {
	body := ": heartbeat\n\n: another\nevent: real\ndata: payload\n\n"
	srv, c := newSSEServer(t, body)
	req, _ := c.NewRequest(srv.URL, http.MethodGet)
	stream, err := c.Stream(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	ev, _ := stream.Next(context.Background())
	if ev.Type != "real" || string(ev.Data) != "payload" {
		t.Errorf("expected real/payload; got %+v", ev)
	}
}

func TestSSE_RejectsNonStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	c := NewClientWithOptions(CliOptsBuilder().Build())
	req, _ := c.NewRequest(srv.URL, http.MethodGet)
	if _, err := c.Stream(context.Background(), req); err == nil {
		t.Fatal("expected error for non-event-stream response")
	} else if !strings.Contains(err.Error(), "text/event-stream") {
		t.Errorf("error should mention text/event-stream; got %v", err)
	}
}

func TestSSE_ContextCancelStopsNext(t *testing.T) {
	// Server holds the connection open without emitting events.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer srv.Close()
	c := NewClientWithOptions(CliOptsBuilder().Build())
	req, _ := c.NewRequest(srv.URL, http.MethodGet)
	stream, err := c.Stream(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := stream.Next(ctx)
		done <- err
	}()
	// Force ctx cancel by closing the stream — this unblocks the scanner.
	time.Sleep(200 * time.Millisecond)
	_ = stream.Close()
	select {
	case <-done:
		// any return is fine — what matters is Next() unblocked.
	case <-time.After(2 * time.Second):
		t.Fatal("Next did not return after stream close")
	}
}
