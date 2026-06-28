package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeClaims struct{ user string }

func TestJWT_OK(t *testing.T) {
	a := JWT(JWTConfig{
		Verify: func(tok string) (any, error) {
			if tok != "good" {
				return nil, errors.New("nope")
			}
			return &fakeClaims{user: "alice"}, nil
		},
	})
	called := false
	h := a.Apply(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		v, ok := ClaimsFrom(r.Context())
		if !ok {
			t.Errorf("claims missing from ctx")
		}
		if c, ok := v.(*fakeClaims); !ok || c.user != "alice" {
			t.Errorf("claims wrong: %v", v)
		}
		w.WriteHeader(http.StatusOK)
	}))
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer good")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if !called {
		t.Error("downstream handler was not invoked")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestJWT_MissingHeaderUnauthorized(t *testing.T) {
	a := JWT(JWTConfig{Verify: func(string) (any, error) { return nil, nil }})
	w := httptest.NewRecorder()
	a.Apply(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("handler should not run")
	})).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestJWT_WrongScheme(t *testing.T) {
	a := JWT(JWTConfig{Verify: func(string) (any, error) { return nil, nil }})
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Basic xyz")
	w := httptest.NewRecorder()
	a.Apply(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Basic scheme should be rejected; status = %d", w.Code)
	}
}

func TestJWT_VerifyFails(t *testing.T) {
	a := JWT(JWTConfig{
		Verify: func(tok string) (any, error) { return nil, errors.New("bad sig") },
	})
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer x")
	w := httptest.NewRecorder()
	a.Apply(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("must not run") })).ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestJWT_CustomScheme(t *testing.T) {
	a := JWT(JWTConfig{
		Scheme: "Token",
		Verify: func(tok string) (any, error) { return tok, nil },
	})
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Token abc123")
	w := httptest.NewRecorder()
	a.Apply(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

func TestJWT_CustomContextKey(t *testing.T) {
	type myKey struct{}
	a := JWT(JWTConfig{
		ContextKey: myKey{},
		Verify:     func(tok string) (any, error) { return "claims!", nil },
	})
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer x")
	w := httptest.NewRecorder()
	a.Apply(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if v := r.Context().Value(myKey{}); v != "claims!" {
			t.Errorf("custom key value wrong: %v", v)
		}
		// Default ClaimsFrom should miss the value (different key).
		if _, ok := ClaimsFrom(r.Context()); ok {
			t.Error("default ClaimsFrom should miss the custom key")
		}
	})).ServeHTTP(w, r)
}
