package turbo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- #8 method registrars ---

func TestRouter_PatchHeadOptionsRegistrars(t *testing.T) {
	r := NewRouter()
	hit := map[string]int{}
	mk := func(method string) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, req *http.Request) {
			hit[method]++
			w.WriteHeader(http.StatusOK)
		}
	}
	if _, err := r.Patch("/x", mk(PATCH)); err != nil {
		t.Fatalf("Patch register: %v", err)
	}
	if _, err := r.Head("/x", mk(HEAD)); err != nil {
		t.Fatalf("Head register: %v", err)
	}
	if _, err := r.Options("/x", mk(OPTIONS)); err != nil {
		t.Fatalf("Options register: %v", err)
	}

	for _, m := range []string{PATCH, HEAD, OPTIONS} {
		req := httptest.NewRequest(m, "/x", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("%s: status = %d, want 200", m, w.Result().StatusCode)
		}
		if hit[m] != 1 {
			t.Errorf("%s handler invoked %d times, want 1", m, hit[m])
		}
	}
}

func TestGroup_PatchHeadOptionsRegistrars(t *testing.T) {
	r := NewRouter()
	g := r.Group("/api")
	called := false
	noop := func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}
	for _, register := range []func(string, func(http.ResponseWriter, *http.Request)) (*Route, error){
		g.Patch, g.Head, g.Options,
	} {
		called = false
		_, err := register("/p", noop)
		if err != nil {
			t.Fatalf("register: %v", err)
		}
	}
	for _, m := range []string{PATCH, HEAD, OPTIONS} {
		called = false
		req := httptest.NewRequest(m, "/api/p", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("%s on group: status = %d, want 200", m, w.Result().StatusCode)
		}
		if !called {
			t.Errorf("%s on group: handler not invoked", m)
		}
	}
}

// --- #7 405 with Allow header ---

func TestRouter_405_SetsAllowHeader(t *testing.T) {
	r := NewRouter()
	noop := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }
	_, _ = r.Get("/users", noop)
	_, _ = r.Post("/users", noop)
	_, _ = r.Patch("/users", noop)

	req := httptest.NewRequest(http.MethodDelete, "/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", w.Result().StatusCode)
	}
	got := w.Result().Header.Get("Allow")
	want := "GET, PATCH, POST" // sorted, comma-separated per RFC
	if got != want {
		t.Errorf("Allow header = %q, want %q", got, want)
	}
}

// --- #7 trailing-slash policy ---

func TestRouter_TrailingSlash_DefaultLenient(t *testing.T) {
	r := NewRouter()
	hit := 0
	_, _ = r.Get("/users", func(w http.ResponseWriter, _ *http.Request) {
		hit++
		w.WriteHeader(http.StatusOK)
	})

	// "/users/" should also match because StrictSlash defaults to false.
	req := httptest.NewRequest(http.MethodGet, "/users/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// refinePath inside ServeHTTP may issue a redirect for trailing-slash
	// normalisation; either outcome is acceptable here as long as the
	// handler eventually runs or the redirect points somewhere sensible.
	// What we're really asserting: no 404.
	if w.Result().StatusCode == http.StatusNotFound {
		t.Fatalf("default StrictSlash=false but /users/ returned 404")
	}
}

func TestRouter_TrailingSlash_DoesNotPromote404To405(t *testing.T) {
	// Regression: when StrictSlash=false, the trailing-slash retry
	// must NOT adopt a match that has no handler for the request's
	// method — otherwise a 404 silently becomes a 405.
	r := NewRouter()
	_, _ = r.Put("/api/widget/:id", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPut, "/api/widget/", nil) // no id
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404 (retry must not promote to 405)", w.Result().StatusCode)
	}
}

func TestRouter_StrictSlash_TrueIsRespected(t *testing.T) {
	r := NewRouter().StrictSlash(true)
	_, _ = r.Get("/users", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// With strict slashes, /users/ must NOT match /users.
	// (refinePath in ServeHTTP returns 301 for trailing-slash mismatch,
	// which is the existing behavior — we just want to make sure the
	// retry-path is suppressed when strict is on.)
	req := httptest.NewRequest(http.MethodGet, "/users/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// 301 (redirect) or 404 — definitely not 200.
	if w.Result().StatusCode == http.StatusOK {
		t.Errorf("StrictSlash(true) but /users/ matched /users (status %d)", w.Result().StatusCode)
	}
}

// --- #7 custom 404/405 setters ---

func TestRouter_SetNotFoundHandler_CalledOnMissingRoute(t *testing.T) {
	r := NewRouter()
	r.SetNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("custom 404"))
	}))
	req := httptest.NewRequest(http.MethodGet, "/nothing-here", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusTeapot {
		t.Errorf("custom 404 not invoked: status = %d", w.Result().StatusCode)
	}
	body := w.Body.String()
	if !strings.Contains(body, "custom 404") {
		t.Errorf("custom 404 body not written: %q", body)
	}
}

func TestRouter_SetMethodNotAllowedHandler_CalledAndAllowStillSet(t *testing.T) {
	r := NewRouter()
	_, _ = r.Get("/x", func(w http.ResponseWriter, _ *http.Request) {})
	r.SetMethodNotAllowedHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusTeapot {
		t.Errorf("custom 405 not invoked: status = %d", w.Result().StatusCode)
	}
	// Even with a custom 405 handler, the Allow header is router policy
	// (RFC 9110 §15.5.6) and should still be set by the router.
	if got := w.Result().Header.Get("Allow"); got != "GET" {
		t.Errorf("Allow = %q, want %q", got, "GET")
	}
}

// --- helpers tests ---

func TestToggleTrailingSlash(t *testing.T) {
	cases := []struct{ in, out string }{
		{"", ""},
		{"/", ""},
		{"/x", "/x/"},
		{"/x/", "/x"},
		{"/a/b/c", "/a/b/c/"},
		{"/a/b/c/", "/a/b/c"},
	}
	for _, c := range cases {
		if got := toggleTrailingSlash(c.in); got != c.out {
			t.Errorf("toggleTrailingSlash(%q) = %q, want %q", c.in, got, c.out)
		}
	}
}

func TestAllowedMethods_Sorted(t *testing.T) {
	r := &Route{handlers: map[string]http.Handler{
		"POST":  nil,
		"GET":   nil,
		"PATCH": nil,
	}}
	got := allowedMethods(r)
	want := "GET, PATCH, POST"
	if got != want {
		t.Errorf("allowedMethods = %q, want %q", got, want)
	}
	if allowedMethods(nil) != "" {
		t.Error("allowedMethods(nil) should be empty")
	}
	if allowedMethods(&Route{}) != "" {
		t.Error("allowedMethods(empty route) should be empty")
	}
}
