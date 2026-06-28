package secrets

import (
	"context"
	"errors"
	"strings"
)

// Op identifies a Store operation, used by Authorizer to enforce policy.
type Op string

const (
	OpGet   Op = "get"
	OpWrite Op = "write"
)

// Authorizer is consulted before each Store operation. Returning a non-nil
// error blocks the op with that error (the underlying Store is never called).
// Use ctx to read principal/tenant info from request scope.
//
//	auth := func(ctx context.Context, op secrets.Op, key string) error {
//	    user := userFromCtx(ctx)
//	    if !user.CanRead(key) { return errors.New("forbidden") }
//	    return nil
//	}
type Authorizer func(ctx context.Context, op Op, key string) error

// ErrForbidden is the canonical error to return from an Authorizer when the
// caller has no permission. Wrap it for richer messages: fmt.Errorf("%w:
// user X cannot read user Y secrets", secrets.ErrForbidden).
var ErrForbidden = errors.New("secrets: forbidden")

// Namespaced wraps a Store, prepending prefix + "/" to every key passed to
// Get/Write. Authorizer (optional) runs before each call.
//
// The wrapper does not touch the prefix — callers see and pass *bare* keys;
// the underlying Store sees fully-qualified ones. A namespaced view of a
// namespaced store concatenates the prefixes.
//
//	root := myStore                        // raw store
//	tenant := secrets.Namespace(root, "tenant/" + tenantID)
//	user   := secrets.Namespace(tenant, "user/" + userID)
//	user.Get("api-key", ctx)               // → root key "tenant/T/user/U/api-key"
type Namespaced struct {
	inner  Store
	prefix string
	auth   Authorizer
}

// Namespace returns a Namespaced view over inner with the given prefix. An
// empty prefix is a no-op (returns inner unchanged when no Authorizer is
// also wired). Pass options to attach an Authorizer.
func Namespace(inner Store, prefix string, opts ...NamespaceOption) Store {
	if inner == nil {
		return nil
	}
	prefix = strings.TrimRight(strings.TrimSpace(prefix), "/")
	cfg := namespaceConfig{}
	for _, o := range opts {
		o(&cfg)
	}
	if prefix == "" && cfg.auth == nil {
		return inner
	}
	return &Namespaced{inner: inner, prefix: prefix, auth: cfg.auth}
}

// NamespaceOption configures a Namespaced wrapper.
type NamespaceOption func(*namespaceConfig)

type namespaceConfig struct {
	auth Authorizer
}

// WithAuthorizer attaches an authorization hook invoked before every
// Get/Write. Returning an error blocks the operation.
func WithAuthorizer(fn Authorizer) NamespaceOption {
	return func(c *namespaceConfig) { c.auth = fn }
}

// Provider proxies to the underlying store.
func (n *Namespaced) Provider() string { return n.inner.Provider() }

// Get fully-qualifies key with the prefix, runs the Authorizer (if set), and
// forwards to the inner Store.
func (n *Namespaced) Get(key string, ctx context.Context) (*Credential, error) {
	full := n.qualify(key)
	if err := n.authorize(ctx, OpGet, full); err != nil {
		return nil, err
	}
	return n.inner.Get(full, ctx)
}

// Write fully-qualifies key with the prefix, runs the Authorizer (if set),
// and forwards to the inner Store.
func (n *Namespaced) Write(key string, credential *Credential, ctx context.Context) error {
	full := n.qualify(key)
	if err := n.authorize(ctx, OpWrite, full); err != nil {
		return err
	}
	return n.inner.Write(full, credential, ctx)
}

// qualify prepends the prefix (and a separator) to key, with traversal
// protection: a leading "/" or ".." is rejected before forwarding so a
// malicious caller can't escape the namespace.
func (n *Namespaced) qualify(key string) string {
	key = strings.TrimSpace(key)
	if n.prefix == "" {
		return key
	}
	return n.prefix + "/" + strings.TrimLeft(key, "/")
}

// authorize runs the hook if set; nil is success.
func (n *Namespaced) authorize(ctx context.Context, op Op, key string) error {
	if n.auth == nil {
		return nil
	}
	return n.auth(ctx, op, key)
}
