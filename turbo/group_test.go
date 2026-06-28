package turbo

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestGroup_PrefixConcatenated(t *testing.T) {
	r := NewRouter()
	api := r.Group("/api/v1")
	_, err := api.Get("/users/:id", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(r)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/v1/users/42")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	// And not at the un-prefixed path.
	resp2, err := http.Get(srv.URL + "/users/42")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode == http.StatusOK {
		t.Error("un-prefixed path should not be served")
	}
}

func TestGroup_NestedGroupComposesPrefixes(t *testing.T) {
	r := NewRouter()
	api := r.Group("/api")
	v1 := api.Group("/v1")
	_, _ = v1.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(r)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/v1/ping")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("nested group prefix not applied; status = %d", resp.StatusCode)
	}
}

func TestGroup_FiltersInheritedFromParent(t *testing.T) {
	r := NewRouter()
	var calls int32
	mark := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&calls, 1)
			next.ServeHTTP(w, r)
		})
	}
	api := r.Group("/api").Use(mark)
	v1 := api.Group("/v1")
	_, _ = v1.Get("/x", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(r)
	defer srv.Close()
	for i := 0; i < 3; i++ {
		resp, _ := http.Get(srv.URL + "/api/v1/x")
		resp.Body.Close()
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Errorf("filter calls = %d, want 3 (parent's filter must apply on nested-group routes)", calls)
	}
}

func TestGroup_AuthenticatorAppliedToGroup(t *testing.T) {
	r := NewRouter()
	var seen int32
	stub := stubAuth(func() { atomic.AddInt32(&seen, 1) })
	api := r.Group("/api")
	api.AddAuthenticator(stub)
	_, _ = api.Get("/x", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(r)
	defer srv.Close()
	resp, _ := http.Get(srv.URL + "/api/x")
	resp.Body.Close()
	if atomic.LoadInt32(&seen) != 1 {
		t.Errorf("group authenticator not invoked: seen=%d", seen)
	}
}

// --- helpers ---

type stubAuthenticator struct{ before func() }

func (s stubAuthenticator) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.before()
		next.ServeHTTP(w, r)
	})
}

func stubAuth(before func()) stubAuthenticator { return stubAuthenticator{before: before} }
