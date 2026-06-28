package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// VerifyFunc is the user-supplied function that validates a raw token and
// returns the authenticated claims (whatever type fits the app — typically
// *jwt.Claims from golly/auth). Returning an error rejects the request with
// 401.
//
// Keeping this as a callback (not a hard import on golly/auth) lets turbo
// stay decoupled from any specific JWT library and lets consumers pin the
// allowlist of algorithms / issuers / audiences themselves.
type VerifyFunc func(token string) (claims any, err error)

// JWTConfig configures the JWT authenticator. Verify is required; the rest
// have sane defaults.
type JWTConfig struct {
	Verify     VerifyFunc
	HeaderName string // default: Authorization
	Scheme     string // default: Bearer  (set "" to take the whole header value)
	ContextKey any    // where validated claims are stashed; defaults to claimsContextKey{}
	OnError    func(w http.ResponseWriter, r *http.Request, err error)
	OnMissing  func(w http.ResponseWriter, r *http.Request)
}

// Errors raised by the JWT authenticator.
var (
	ErrMissingAuthHeader = errors.New("turbo/auth/jwt: missing Authorization header")
	ErrWrongScheme       = errors.New("turbo/auth/jwt: wrong authorization scheme")
	ErrEmptyToken        = errors.New("turbo/auth/jwt: empty token")
)

// claimsContextKey is the default key used when the caller doesn't supply one.
type claimsContextKey struct{}

// JWT returns an Authenticator that extracts a bearer token from the
// configured header, runs Verify(token), and on success injects the claims
// into the request context. On failure it writes 401 (or calls the
// OnError/OnMissing override).
func JWT(cfg JWTConfig) Authenticator {
	if cfg.Verify == nil {
		panic("turbo/auth/jwt: Verify is required")
	}
	if cfg.HeaderName == "" {
		cfg.HeaderName = "Authorization"
	}
	scheme := strings.TrimSpace(cfg.Scheme)
	if cfg.Scheme == "" {
		scheme = "Bearer"
	}
	ckey := cfg.ContextKey
	if ckey == nil {
		ckey = claimsContextKey{}
	}
	onErr := cfg.OnError
	if onErr == nil {
		onErr = defaultOnError
	}
	onMissing := cfg.OnMissing
	if onMissing == nil {
		onMissing = func(w http.ResponseWriter, r *http.Request) {
			defaultOnError(w, r, ErrMissingAuthHeader)
		}
	}

	return &jwtAuthenticator{
		verify:     cfg.Verify,
		header:     cfg.HeaderName,
		scheme:     scheme,
		contextKey: ckey,
		onError:    onErr,
		onMissing:  onMissing,
	}
}

// ClaimsFrom retrieves claims previously stashed by JWT(...) from the
// request context. The second return is false if no claims are present.
//
// If you used a custom ContextKey, call ctx.Value(myKey) directly instead.
func ClaimsFrom(ctx context.Context) (any, bool) {
	v := ctx.Value(claimsContextKey{})
	return v, v != nil
}

// jwtAuthenticator implements Authenticator.
type jwtAuthenticator struct {
	verify     VerifyFunc
	header     string
	scheme     string
	contextKey any
	onError    func(w http.ResponseWriter, r *http.Request, err error)
	onMissing  func(w http.ResponseWriter, r *http.Request)
}

func (a *jwtAuthenticator) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get(a.header)
		if raw == "" {
			a.onMissing(w, r)
			return
		}
		token := raw
		if a.scheme != "" {
			prefix := a.scheme + " "
			if !strings.HasPrefix(raw, prefix) {
				a.onError(w, r, ErrWrongScheme)
				return
			}
			token = strings.TrimSpace(raw[len(prefix):])
		}
		if token == "" {
			a.onError(w, r, ErrEmptyToken)
			return
		}
		claims, err := a.verify(token)
		if err != nil {
			a.onError(w, r, err)
			return
		}
		ctx := context.WithValue(r.Context(), a.contextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func defaultOnError(w http.ResponseWriter, _ *http.Request, err error) {
	http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
}
