package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"strings"
	"testing"
	"time"
)

// ---- JWT: round-trip per algorithm ----

func TestJWT_HS256_RoundTrip(t *testing.T) {
	s, err := NewHSSigner(AlgHS256, []byte("super-secret-key-32-bytes-of-zzzzzz"))
	if err != nil {
		t.Fatal(err)
	}
	claims := &Claims{
		Issuer:    "test",
		Subject:   "u-1",
		Audience:  Audience{"api"},
		ExpiresAt: NewNumericDate(time.Now().Add(5 * time.Minute)),
		Extra:     map[string]any{"role": "admin"},
	}
	tok, err := Sign(s, claims, "k1")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	got, err := Verify(tok, VerifyOptions{
		Algs:             []string{AlgHS256},
		VerifierFallback: s,
		Issuer:           "test",
		Audience:         "api",
	})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got.Subject != "u-1" {
		t.Errorf("subject roundtrip wrong: %q", got.Subject)
	}
	if got.Extra["role"] != "admin" {
		t.Errorf("extra roundtrip wrong: %v", got.Extra)
	}
}

func TestJWT_RS256_RoundTrip(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewRSSigner(AlgRS256, priv)
	if err != nil {
		t.Fatal(err)
	}
	v, err := NewRSVerifier(AlgRS256, &priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	tok, _ := Sign(s, &Claims{Subject: "u"}, "")
	if _, err := Verify(tok, VerifyOptions{Algs: []string{AlgRS256}, VerifierFallback: v}); err != nil {
		t.Errorf("RS verify: %v", err)
	}
}

func TestJWT_EdDSA_RoundTrip(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	s := NewEdDSASigner(priv)
	v := NewEdDSAVerifier(priv.Public().(ed25519.PublicKey))
	tok, _ := Sign(s, &Claims{Subject: "u"}, "")
	if _, err := Verify(tok, VerifyOptions{Algs: []string{AlgEdDSA}, VerifierFallback: v}); err != nil {
		t.Errorf("EdDSA verify: %v", err)
	}
}

// ---- JWT: security guarantees ----

func TestJWT_RejectsAlgNotInAllowlist(t *testing.T) {
	s, _ := NewHSSigner(AlgHS256, []byte("k"))
	tok, _ := Sign(s, &Claims{}, "")

	// Verifier allows only RS256 → must reject HS256 token.
	_, err := Verify(tok, VerifyOptions{Algs: []string{AlgRS256}, VerifierFallback: s})
	if !errors.Is(err, ErrAlgNotAllowed) {
		t.Errorf("expected ErrAlgNotAllowed; got %v", err)
	}
}

func TestJWT_RejectsEmptyAllowlist(t *testing.T) {
	s, _ := NewHSSigner(AlgHS256, []byte("k"))
	tok, _ := Sign(s, &Claims{}, "")
	_, err := Verify(tok, VerifyOptions{VerifierFallback: s})
	if !errors.Is(err, ErrAlgNotAllowed) {
		t.Errorf("expected ErrAlgNotAllowed with empty allowlist; got %v", err)
	}
}

func TestJWT_RejectsTampered(t *testing.T) {
	s, _ := NewHSSigner(AlgHS256, []byte("k"))
	tok, _ := Sign(s, &Claims{Subject: "u"}, "")
	// Flip the last char.
	tampered := tok[:len(tok)-1] + flip(tok[len(tok)-1])
	_, err := Verify(tampered, VerifyOptions{Algs: []string{AlgHS256}, VerifierFallback: s})
	if !errors.Is(err, ErrInvalidSig) {
		t.Errorf("expected ErrInvalidSig; got %v", err)
	}
}

func flip(c byte) string {
	if c == 'a' {
		return "b"
	}
	return "a"
}

func TestJWT_ExpiredAndNotYetValid(t *testing.T) {
	s, _ := NewHSSigner(AlgHS256, []byte("k"))

	expired := &Claims{ExpiresAt: NewNumericDate(time.Now().Add(-time.Hour))}
	tok, _ := Sign(s, expired, "")
	if _, err := Verify(tok, VerifyOptions{Algs: []string{AlgHS256}, VerifierFallback: s}); !errors.Is(err, ErrExpired) {
		t.Errorf("expected ErrExpired; got %v", err)
	}

	future := &Claims{NotBefore: NewNumericDate(time.Now().Add(time.Hour))}
	tok2, _ := Sign(s, future, "")
	if _, err := Verify(tok2, VerifyOptions{Algs: []string{AlgHS256}, VerifierFallback: s}); !errors.Is(err, ErrNotYetValid) {
		t.Errorf("expected ErrNotYetValid; got %v", err)
	}
}

func TestJWT_IssuerAndAudience(t *testing.T) {
	s, _ := NewHSSigner(AlgHS256, []byte("k"))
	tok, _ := Sign(s, &Claims{Issuer: "good", Audience: Audience{"api"}}, "")

	if _, err := Verify(tok, VerifyOptions{Algs: []string{AlgHS256}, VerifierFallback: s, Issuer: "bad"}); !errors.Is(err, ErrIssuerMismatch) {
		t.Errorf("expected ErrIssuerMismatch; got %v", err)
	}
	if _, err := Verify(tok, VerifyOptions{Algs: []string{AlgHS256}, VerifierFallback: s, Audience: "missing"}); !errors.Is(err, ErrAudMismatch) {
		t.Errorf("expected ErrAudMismatch; got %v", err)
	}
}

func TestJWT_KeysetByKID(t *testing.T) {
	k1, _ := NewHSSigner(AlgHS256, []byte("k1key"))
	k2, _ := NewHSSigner(AlgHS256, []byte("k2key"))
	tok, _ := Sign(k2, &Claims{Subject: "u"}, "k2")

	keyset := func(kid string) (Verifier, error) {
		switch kid {
		case "k1":
			return k1, nil
		case "k2":
			return k2, nil
		}
		return nil, ErrKeyNotFound
	}
	if _, err := Verify(tok, VerifyOptions{Algs: []string{AlgHS256}, Keyset: keyset}); err != nil {
		t.Errorf("expected verify with kid=k2; got %v", err)
	}
}

func TestJWT_Audience_StringOrArray(t *testing.T) {
	s, _ := NewHSSigner(AlgHS256, []byte("k"))
	// Single-element audience marshals as a bare string.
	tok, _ := Sign(s, &Claims{Audience: Audience{"api"}}, "")
	if !strings.Contains(tok, ".") {
		t.Fatal("malformed token")
	}
	got, err := Verify(tok, VerifyOptions{Algs: []string{AlgHS256}, VerifierFallback: s})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Audience) != 1 || got.Audience[0] != "api" {
		t.Errorf("aud roundtrip wrong: %v", got.Audience)
	}
}

// ---- Password hasher ----

func TestPassword_HashVerifyRoundtrip(t *testing.T) {
	h := DefaultPasswordHasher()
	// Reduce parameters for test speed.
	h.Memory = 8192
	h.Iterations = 1
	hash, err := h.Hash("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if err := VerifyPassword("correct-horse-battery-staple", hash); err != nil {
		t.Errorf("expected match; got %v", err)
	}
	if err := VerifyPassword("wrong", hash); !errors.Is(err, ErrPasswordMismatch) {
		t.Errorf("expected ErrPasswordMismatch; got %v", err)
	}
}

func TestPassword_RejectsMalformedHash(t *testing.T) {
	if err := VerifyPassword("x", "not-a-phc-string"); err == nil {
		t.Error("expected error on malformed hash")
	}
}

// ---- Sessions ----

func TestSession_NewValidateRevoke(t *testing.T) {
	mgr := NewSessionManager(nil, SessionConfig{Absolute: time.Hour})
	sess, err := mgr.New("user-42", map[string]any{"k": "v"})
	if err != nil {
		t.Fatal(err)
	}
	if sess.ID == "" || sess.Subject != "user-42" {
		t.Errorf("bad session: %+v", sess)
	}
	got, ok := mgr.Validate(sess.ID)
	if !ok || got.Subject != "user-42" {
		t.Errorf("validate failed: ok=%v sess=%+v", ok, got)
	}
	_ = mgr.Revoke(sess.ID)
	if _, ok := mgr.Validate(sess.ID); ok {
		t.Error("revoked session should not validate")
	}
}

func TestSession_TouchSlidesExpiry(t *testing.T) {
	mgr := NewSessionManager(nil, SessionConfig{Absolute: time.Hour, Sliding: 10 * time.Minute})
	sess, _ := mgr.New("u", nil)
	origExp := sess.ExpiresAt
	// Move "now" forward.
	mgr.now = func() time.Time { return time.Now().Add(5 * time.Minute) }
	_ = mgr.Touch(sess.ID)
	got, _ := mgr.Validate(sess.ID)
	if !got.ExpiresAt.After(origExp) {
		t.Errorf("touch should extend expiry; orig=%v new=%v", origExp, got.ExpiresAt)
	}
}

func TestSession_AbsoluteExpiry(t *testing.T) {
	mgr := NewSessionManager(nil, SessionConfig{Absolute: time.Hour})
	sess, _ := mgr.New("u", nil)
	mgr.now = func() time.Time { return time.Now().Add(2 * time.Hour) }
	if _, ok := mgr.Validate(sess.ID); ok {
		t.Error("expired session should not validate")
	}
}
