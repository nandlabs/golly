package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- AccessLog ---

func TestAccessLog_PassesThroughAndCapturesStatus(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	})
	wrapped := AccessLog()(handler)
	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)
	if !called {
		t.Fatal("AccessLog did not call the underlying handler")
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "ok")
	}
}

// statusRecorder default 200 when the handler writes a body without
// calling WriteHeader explicitly — net/http's documented behavior.
func TestStatusRecorder_DefaultsTo200OnWriteWithoutHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rec, status: http.StatusOK}
	_, _ = sr.Write([]byte("hello"))
	if sr.status != http.StatusOK {
		t.Errorf("status = %d, want 200", sr.status)
	}
	if sr.bytes != 5 {
		t.Errorf("bytes = %d, want 5", sr.bytes)
	}
}

// --- Recover ---

func TestRecover_HandlerPanicYields500(t *testing.T) {
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})
	wrapped := Recover()(handler)
	req := httptest.NewRequest(http.MethodGet, "/explode", nil)
	rec := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Recover did not swallow panic: %v", r)
		}
	}()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "boom") {
		t.Errorf("body should include the panic value; got %q", rec.Body.String())
	}
}

func TestRecover_NoPanicIsPassthrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
	wrapped := Recover()(handler)
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want 202", rec.Code)
	}
}

func TestRecover_PropagatesAbortHandler(t *testing.T) {
	// http.ErrAbortHandler is the documented escape hatch — Recover
	// must let it through so net/http closes the connection silently.
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	})
	wrapped := Recover()(handler)
	req := httptest.NewRequest(http.MethodGet, "/abort", nil)
	rec := httptest.NewRecorder()

	defer func() {
		r := recover()
		if r != http.ErrAbortHandler {
			t.Errorf("expected http.ErrAbortHandler to propagate, got %v", r)
		}
	}()
	wrapped.ServeHTTP(rec, req)
	t.Fatal("expected re-panic with ErrAbortHandler")
}
