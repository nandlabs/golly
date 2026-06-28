package turbo

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSSE_Headers(t *testing.T) {
	w := httptest.NewRecorder()
	sse, err := NewSSE(w)
	if err != nil {
		t.Fatal(err)
	}
	defer sse.Close()
	h := w.Header()
	if h.Get("Content-Type") != "text/event-stream" {
		t.Errorf("Content-Type = %q", h.Get("Content-Type"))
	}
	if h.Get("Cache-Control") != "no-cache" {
		t.Errorf("Cache-Control wrong")
	}
	if h.Get("X-Accel-Buffering") != "no" {
		t.Errorf("X-Accel-Buffering should be 'no' to defeat nginx buffering")
	}
}

func TestSSE_SendStringAndJSON(t *testing.T) {
	w := httptest.NewRecorder()
	sse, err := NewSSE(w)
	if err != nil {
		t.Fatal(err)
	}
	if err := sse.Send("greet", "hello"); err != nil {
		t.Fatalf("send string: %v", err)
	}
	if err := sse.Send("payload", map[string]int{"n": 42}); err != nil {
		t.Fatalf("send map: %v", err)
	}
	out := w.Body.String()
	if !strings.Contains(out, "event: greet\ndata: hello\n\n") {
		t.Errorf("missing greet record:\n%s", out)
	}
	if !strings.Contains(out, "event: payload\ndata: {\"n\":42}\n\n") {
		t.Errorf("missing payload record:\n%s", out)
	}
}

func TestSSE_MultilineDataPrefixesEveryLine(t *testing.T) {
	w := httptest.NewRecorder()
	sse, _ := NewSSE(w)
	_ = sse.Send("multi", "line1\nline2\nline3")
	out := w.Body.String()
	if !strings.Contains(out, "data: line1\ndata: line2\ndata: line3\n") {
		t.Errorf("each line should be prefixed:\n%s", out)
	}
}

func TestSSE_CloseEmitsCloseEvent(t *testing.T) {
	w := httptest.NewRecorder()
	sse, _ := NewSSE(w)
	_ = sse.Close()
	if !strings.Contains(w.Body.String(), "event: close\ndata: bye\n\n") {
		t.Errorf("close should emit close event; got:\n%s", w.Body.String())
	}
	// Send after Close should error.
	if err := sse.Send("x", "y"); err == nil {
		t.Error("send after close should error")
	}
}

func TestSSE_NoFlusherReturnsErr(t *testing.T) {
	// Explicitly-non-Flusher wrapper (embedding would promote Flush() from
	// the recorder and defeat the test).
	rec := httptest.NewRecorder()
	w := &nonFlusherWriter{rec: rec}
	if _, err := NewSSE(w); err != ErrUnflushable {
		t.Errorf("expected ErrUnflushable; got %v", err)
	}
}

type nonFlusherWriter struct {
	rec *httptest.ResponseRecorder
}

func (n *nonFlusherWriter) Header() http.Header         { return n.rec.Header() }
func (n *nonFlusherWriter) Write(b []byte) (int, error) { return n.rec.Write(b) }
func (n *nonFlusherWriter) WriteHeader(s int)           { n.rec.WriteHeader(s) }

// Real end-to-end with an actual httptest.Server to prove streaming works.
func TestSSE_EndToEnd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sse, err := NewSSE(w)
		if err != nil {
			t.Errorf("NewSSE: %v", err)
			return
		}
		defer sse.Close()
		for i := 0; i < 3; i++ {
			_ = sse.Send("tick", i)
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	got := string(body)
	for i := 0; i < 3; i++ {
		want := "event: tick\ndata: " + string(rune('0'+i))
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in body:\n%s", want, got)
		}
	}
	if !strings.Contains(got, "event: close") {
		t.Errorf("missing close event:\n%s", got)
	}
}
