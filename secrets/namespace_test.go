package secrets

import (
	"context"
	"errors"
	"sync"
	"testing"
)

// fakeStore is a tiny in-memory Store implementation used in tests.
type fakeStore struct {
	mu           sync.RWMutex
	m            map[string]*Credential
	name         string
	lastReadKey  string
	lastWriteKey string
}

func newFakeStore() *fakeStore        { return &fakeStore{m: map[string]*Credential{}, name: "fake"} }
func (f *fakeStore) Provider() string { return f.name }
func (f *fakeStore) Get(key string, _ context.Context) (*Credential, error) {
	f.mu.Lock()
	f.lastReadKey = key
	c, ok := f.m[key]
	f.mu.Unlock()
	if !ok {
		return nil, errors.New("not found")
	}
	return c, nil
}
func (f *fakeStore) Write(key string, c *Credential, _ context.Context) error {
	f.mu.Lock()
	f.m[key] = c
	f.lastWriteKey = key
	f.mu.Unlock()
	return nil
}

func TestNamespace_PrefixesKeys(t *testing.T) {
	inner := newFakeStore()
	ns := Namespace(inner, "tenant/T1/user/U7")

	ctx := context.Background()
	cred := &Credential{Value: []byte("v")}
	if err := ns.Write("api-key", cred, ctx); err != nil {
		t.Fatal(err)
	}
	wantKey := "tenant/T1/user/U7/api-key"
	if inner.lastWriteKey != wantKey {
		t.Errorf("inner write key = %q, want %q", inner.lastWriteKey, wantKey)
	}
	// Get with the bare key reads the qualified one.
	if _, err := ns.Get("api-key", ctx); err != nil {
		t.Errorf("get: %v", err)
	}
	if inner.lastReadKey != wantKey {
		t.Errorf("inner read key = %q, want %q", inner.lastReadKey, wantKey)
	}
}

func TestNamespace_NestedNamespacesConcatenate(t *testing.T) {
	inner := newFakeStore()
	tenant := Namespace(inner, "tenant/T")
	user := Namespace(tenant, "user/U")
	_ = user.Write("k", &Credential{Value: []byte("x")}, context.Background())
	if inner.lastWriteKey != "tenant/T/user/U/k" {
		t.Errorf("nested key = %q", inner.lastWriteKey)
	}
}

func TestNamespace_StripsLeadingSlashOnKey(t *testing.T) {
	inner := newFakeStore()
	ns := Namespace(inner, "ns")
	_ = ns.Write("/leaked", &Credential{Value: []byte("x")}, context.Background())
	if inner.lastWriteKey != "ns/leaked" {
		t.Errorf("leading-slash escape: %q", inner.lastWriteKey)
	}
}

func TestNamespace_EmptyPrefixIsNoOp(t *testing.T) {
	inner := newFakeStore()
	ns := Namespace(inner, "")
	if ns != inner {
		t.Errorf("empty prefix without authorizer should return inner unchanged")
	}
}

func TestNamespace_EmptyPrefixWithAuthorizerStillWraps(t *testing.T) {
	inner := newFakeStore()
	authCalls := 0
	ns := Namespace(inner, "", WithAuthorizer(func(_ context.Context, _ Op, _ string) error {
		authCalls++
		return nil
	}))
	_ = ns.Write("k", &Credential{Value: []byte("v")}, context.Background())
	_, _ = ns.Get("k", context.Background())
	if authCalls != 2 {
		t.Errorf("authorizer should have been called 2x; got %d", authCalls)
	}
}

func TestAuthorizer_BlocksOperations(t *testing.T) {
	inner := newFakeStore()
	ns := Namespace(inner, "x", WithAuthorizer(func(_ context.Context, op Op, key string) error {
		if op == OpGet && key == "x/secret" {
			return ErrForbidden
		}
		return nil
	}))
	// Write is allowed.
	if err := ns.Write("secret", &Credential{Value: []byte("v")}, context.Background()); err != nil {
		t.Errorf("write should be allowed: %v", err)
	}
	// Get is forbidden.
	if _, err := ns.Get("secret", context.Background()); !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden; got %v", err)
	}
	// Inner store should NOT have been read.
	if inner.lastReadKey != "" {
		t.Errorf("authorizer should short-circuit before inner.Get; got %q", inner.lastReadKey)
	}
}

func TestNamespace_NilInnerReturnsNil(t *testing.T) {
	if got := Namespace(nil, "x"); got != nil {
		t.Errorf("Namespace(nil, ...) should return nil; got %T", got)
	}
}
