// Package auth provides identity primitives — JWT mint/verify, password
// hashing, and session-token helpers — that are datastore-agnostic. It is
// the foundation for golly/turbo's JWT middleware and any application that
// wants a sane, stdlib-first identity layer without pulling in a third-party
// JWT library.
//
// Stdlib only, except for x/crypto/argon2 used by the password hasher
// (already in tree as a direct dep of golly).
package auth

import (
	"crypto"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"strings"
	"time"
)

// Algorithm names supported by Sign / Verify. Pin your allowlist explicitly
// at the verifier — never derive it from the token header alone.
const (
	AlgHS256 = "HS256"
	AlgHS384 = "HS384"
	AlgHS512 = "HS512"
	AlgRS256 = "RS256"
	AlgRS384 = "RS384"
	AlgRS512 = "RS512"
	AlgEdDSA = "EdDSA"
)

// Standard JWT errors.
var (
	ErrInvalidToken   = errors.New("auth/jwt: invalid token")
	ErrInvalidSig     = errors.New("auth/jwt: invalid signature")
	ErrAlgNotAllowed  = errors.New("auth/jwt: algorithm not allowed")
	ErrExpired        = errors.New("auth/jwt: token expired")
	ErrNotYetValid    = errors.New("auth/jwt: token not yet valid")
	ErrIssuerMismatch = errors.New("auth/jwt: issuer mismatch")
	ErrAudMismatch    = errors.New("auth/jwt: audience mismatch")
	ErrKeyNotFound    = errors.New("auth/jwt: signing key not found")
)

// Header is the JOSE header for a JWS-style JWT.
type Header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ,omitempty"`
	Kid string `json:"kid,omitempty"`
}

// Claims is the JWT payload. RegisteredClaims covers the RFC 7519 reserved
// names; custom fields go in Extra and are merged into the payload at sign
// time and surfaced back on verify.
type Claims struct {
	Issuer    string         `json:"iss,omitempty"`
	Subject   string         `json:"sub,omitempty"`
	Audience  Audience       `json:"aud,omitempty"`
	ExpiresAt *NumericDate   `json:"exp,omitempty"`
	NotBefore *NumericDate   `json:"nbf,omitempty"`
	IssuedAt  *NumericDate   `json:"iat,omitempty"`
	ID        string         `json:"jti,omitempty"`
	Extra     map[string]any `json:"-"`
}

// Audience is a single string OR a slice — both shapes are valid per RFC 7519.
type Audience []string

// MarshalJSON renders a single-element Audience as a bare string.
func (a Audience) MarshalJSON() ([]byte, error) {
	switch len(a) {
	case 0:
		return []byte("null"), nil
	case 1:
		return json.Marshal(a[0])
	default:
		return json.Marshal([]string(a))
	}
}

// UnmarshalJSON accepts string OR []string.
func (a *Audience) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*a = nil
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*a = Audience{s}
		return nil
	}
	var arr []string
	if err := json.Unmarshal(b, &arr); err != nil {
		return err
	}
	*a = arr
	return nil
}

// NumericDate is RFC 7519 §2 NumericDate: seconds since epoch.
type NumericDate struct{ time.Time }

func NewNumericDate(t time.Time) *NumericDate { return &NumericDate{t.UTC()} }

func (n *NumericDate) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", n.Unix())), nil
}

func (n *NumericDate) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	// Accept integer or float (Go's encoding/json may emit a float for some).
	var f float64
	if err := json.Unmarshal(b, &f); err != nil {
		return err
	}
	n.Time = time.Unix(int64(f), 0).UTC()
	return nil
}

// Signer produces signatures for one algorithm.
type Signer interface {
	Alg() string
	Sign(signingInput []byte) ([]byte, error)
}

// Verifier validates signatures for one algorithm.
type Verifier interface {
	Alg() string
	Verify(signingInput, sig []byte) error
}

// --- HS (HMAC) ---

// HSSigner is an HMAC-SHA{256,384,512} signer + verifier (symmetric).
type HSSigner struct {
	alg string
	key []byte
}

// NewHSSigner returns an HS signer/verifier for alg in {AlgHS256, AlgHS384, AlgHS512}.
func NewHSSigner(alg string, key []byte) (*HSSigner, error) {
	if _, err := hsHash(alg); err != nil {
		return nil, err
	}
	if len(key) == 0 {
		return nil, fmt.Errorf("auth/jwt: HS key must be non-empty")
	}
	return &HSSigner{alg: alg, key: key}, nil
}

func (s *HSSigner) Alg() string { return s.alg }

func (s *HSSigner) Sign(signingInput []byte) ([]byte, error) {
	h, _ := hsHash(s.alg)
	mac := hmac.New(h, s.key)
	_, _ = mac.Write(signingInput)
	return mac.Sum(nil), nil
}

func (s *HSSigner) Verify(signingInput, sig []byte) error {
	expected, _ := s.Sign(signingInput)
	if !hmac.Equal(expected, sig) {
		return ErrInvalidSig
	}
	return nil
}

func hsHash(alg string) (func() hash.Hash, error) {
	switch alg {
	case AlgHS256:
		return sha256.New, nil
	case AlgHS384:
		return sha512.New384, nil
	case AlgHS512:
		return sha512.New, nil
	}
	return nil, fmt.Errorf("auth/jwt: unsupported HS alg %q", alg)
}

// --- RS (RSA) ---

// RSSigner is an RSA-PKCS1v15-SHA{256,384,512} signer.
type RSSigner struct {
	alg string
	key *rsa.PrivateKey
}

// NewRSSigner returns an RS signer for alg in {AlgRS256, AlgRS384, AlgRS512}.
// To verify-only with a public key, use NewRSVerifier instead.
func NewRSSigner(alg string, key *rsa.PrivateKey) (*RSSigner, error) {
	if _, err := rsHash(alg); err != nil {
		return nil, err
	}
	if key == nil {
		return nil, fmt.Errorf("auth/jwt: RS private key is nil")
	}
	return &RSSigner{alg: alg, key: key}, nil
}

func (s *RSSigner) Alg() string { return s.alg }

func (s *RSSigner) Sign(signingInput []byte) ([]byte, error) {
	h, hashID := rsHashPair(s.alg)
	sum := h()
	_, _ = sum.Write(signingInput)
	return rsa.SignPKCS1v15(nil, s.key, hashID, sum.Sum(nil))
}

func (s *RSSigner) Verify(signingInput, sig []byte) error {
	pub := &s.key.PublicKey
	return rsVerify(s.alg, pub, signingInput, sig)
}

// RSVerifier validates RSA signatures with a public key only.
type RSVerifier struct {
	alg string
	pub *rsa.PublicKey
}

// NewRSVerifier returns a verifier-only with a public key.
func NewRSVerifier(alg string, pub *rsa.PublicKey) (*RSVerifier, error) {
	if _, err := rsHash(alg); err != nil {
		return nil, err
	}
	if pub == nil {
		return nil, fmt.Errorf("auth/jwt: RS public key is nil")
	}
	return &RSVerifier{alg: alg, pub: pub}, nil
}

func (v *RSVerifier) Alg() string                    { return v.alg }
func (v *RSVerifier) Verify(input, sig []byte) error { return rsVerify(v.alg, v.pub, input, sig) }

func rsVerify(alg string, pub *rsa.PublicKey, input, sig []byte) error {
	h, hashID := rsHashPair(alg)
	sum := h()
	_, _ = sum.Write(input)
	if err := rsa.VerifyPKCS1v15(pub, hashID, sum.Sum(nil), sig); err != nil {
		return ErrInvalidSig
	}
	return nil
}

func rsHash(alg string) (func() hash.Hash, error) {
	switch alg {
	case AlgRS256:
		return sha256.New, nil
	case AlgRS384:
		return sha512.New384, nil
	case AlgRS512:
		return sha512.New, nil
	}
	return nil, fmt.Errorf("auth/jwt: unsupported RS alg %q", alg)
}

func rsHashPair(alg string) (func() hash.Hash, crypto.Hash) {
	switch alg {
	case AlgRS256:
		return sha256.New, crypto.SHA256
	case AlgRS384:
		return sha512.New384, crypto.SHA384
	}
	return sha512.New, crypto.SHA512
}

// --- EdDSA ---

// EdDSASigner is an Ed25519 signer + verifier.
type EdDSASigner struct {
	priv ed25519.PrivateKey
	pub  ed25519.PublicKey
}

func NewEdDSASigner(priv ed25519.PrivateKey) *EdDSASigner {
	return &EdDSASigner{priv: priv, pub: priv.Public().(ed25519.PublicKey)}
}

func (s *EdDSASigner) Alg() string { return AlgEdDSA }
func (s *EdDSASigner) Sign(input []byte) ([]byte, error) {
	return ed25519.Sign(s.priv, input), nil
}
func (s *EdDSASigner) Verify(input, sig []byte) error { return edVerify(s.pub, input, sig) }

// EdDSAVerifier validates Ed25519 signatures with a public key only.
type EdDSAVerifier struct{ pub ed25519.PublicKey }

func NewEdDSAVerifier(pub ed25519.PublicKey) *EdDSAVerifier { return &EdDSAVerifier{pub: pub} }
func (v *EdDSAVerifier) Alg() string                        { return AlgEdDSA }
func (v *EdDSAVerifier) Verify(input, sig []byte) error     { return edVerify(v.pub, input, sig) }

func edVerify(pub ed25519.PublicKey, input, sig []byte) error {
	if !ed25519.Verify(pub, input, sig) {
		return ErrInvalidSig
	}
	return nil
}

// --- Sign / Verify ---

// Sign produces a compact-form JWT using signer.
func Sign(signer Signer, claims *Claims, kid string) (string, error) {
	if signer == nil {
		return "", errors.New("auth/jwt: signer is nil")
	}
	header := Header{Alg: signer.Alg(), Typ: "JWT", Kid: kid}
	hb, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	pb, err := marshalClaims(claims)
	if err != nil {
		return "", err
	}
	hEnc := base64.RawURLEncoding.EncodeToString(hb)
	pEnc := base64.RawURLEncoding.EncodeToString(pb)
	signingInput := []byte(hEnc + "." + pEnc)
	sig, err := signer.Sign(signingInput)
	if err != nil {
		return "", err
	}
	return string(signingInput) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

// marshalClaims merges Extra into the registered-claims JSON object.
func marshalClaims(c *Claims) ([]byte, error) {
	if c == nil {
		return []byte("{}"), nil
	}
	// Marshal then unmarshal-into-map so we can splice Extra in.
	first, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	m := map[string]any{}
	if err := json.Unmarshal(first, &m); err != nil {
		return nil, err
	}
	for k, v := range c.Extra {
		m[k] = v
	}
	return json.Marshal(m)
}

// VerifyOptions configure Verify. Algs is a hard allowlist — algorithms in
// the token header that are not on this list are rejected (prevents
// algorithm confusion). Keyset chooses the verifier per kid; if Keyset is
// nil, VerifierFallback is used regardless of kid.
type VerifyOptions struct {
	Algs             []string
	Keyset           func(kid string) (Verifier, error)
	VerifierFallback Verifier
	Issuer           string        // optional; if non-empty must match exactly
	Audience         string        // optional; if non-empty must appear in aud
	Leeway           time.Duration // clock skew allowance
	// Now overrides the time source for testing.
	Now func() time.Time
}

// Verify parses, signature-checks, and claim-validates a compact-form JWT,
// returning the parsed Claims. It rejects 'none' implicitly and any
// algorithm not in opts.Algs.
func Verify(token string, opts VerifyOptions) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	hb, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var header Header
	if err = json.Unmarshal(hb, &header); err != nil {
		return nil, ErrInvalidToken
	}

	if len(opts.Algs) == 0 {
		return nil, ErrAlgNotAllowed
	}
	allowed := false
	for _, a := range opts.Algs {
		if a == header.Alg {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, ErrAlgNotAllowed
	}

	var verifier Verifier
	if opts.Keyset != nil {
		verifier, err = opts.Keyset(header.Kid)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrKeyNotFound, err)
		}
	} else {
		verifier = opts.VerifierFallback
	}
	if verifier == nil {
		return nil, ErrKeyNotFound
	}
	if verifier.Alg() != header.Alg {
		return nil, ErrAlgNotAllowed
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}
	if sigErr := verifier.Verify([]byte(parts[0]+"."+parts[1]), sig); sigErr != nil {
		return nil, sigErr
	}

	pb, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	c, err := unmarshalClaims(pb)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	now := time.Now()
	if opts.Now != nil {
		now = opts.Now()
	}
	if c.ExpiresAt != nil && now.After(c.ExpiresAt.Add(opts.Leeway)) {
		return nil, ErrExpired
	}
	if c.NotBefore != nil && now.Add(opts.Leeway).Before(c.NotBefore.Time) {
		return nil, ErrNotYetValid
	}
	if opts.Issuer != "" && c.Issuer != opts.Issuer {
		return nil, ErrIssuerMismatch
	}
	if opts.Audience != "" {
		ok := false
		for _, a := range c.Audience {
			if a == opts.Audience {
				ok = true
				break
			}
		}
		if !ok {
			return nil, ErrAudMismatch
		}
	}
	return c, nil
}

// unmarshalClaims first decodes the registered claims, then captures any
// remaining fields into Extra so callers don't lose data.
func unmarshalClaims(pb []byte) (*Claims, error) {
	c := &Claims{}
	if err := json.Unmarshal(pb, c); err != nil {
		return nil, err
	}
	all := map[string]any{}
	if err := json.Unmarshal(pb, &all); err != nil {
		return nil, err
	}
	for _, reserved := range []string{"iss", "sub", "aud", "exp", "nbf", "iat", "jti"} {
		delete(all, reserved)
	}
	if len(all) > 0 {
		c.Extra = all
	}
	return c, nil
}
