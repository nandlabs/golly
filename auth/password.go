package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// PasswordHasher hashes and verifies passwords using argon2id with the OWASP
// 2024 recommended defaults. The output is a self-describing PHC-style
// string ("$argon2id$...") so verification doesn't need the parameters
// passed in separately — they're embedded in the hash.
type PasswordHasher struct {
	Memory      uint32 // KiB; default 47104 (46 MiB)
	Iterations  uint32 // default 1
	Parallelism uint8  // default 1
	SaltLen     uint32 // bytes; default 16
	KeyLen      uint32 // bytes; default 32
}

// DefaultPasswordHasher returns a PasswordHasher with the OWASP-recommended
// argon2id parameters as of 2024 (m=46 MiB, t=1, p=1).
func DefaultPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		Memory:      47104,
		Iterations:  1,
		Parallelism: 1,
		SaltLen:     16,
		KeyLen:      32,
	}
}

// ErrPasswordMismatch is returned by Verify when the password does not match.
var ErrPasswordMismatch = errors.New("auth/password: password mismatch")

// Hash returns a PHC-encoded argon2id hash:
//
//	$argon2id$v=19$m=47104,t=1,p=1$<base64-salt>$<base64-hash>
func (h *PasswordHasher) Hash(password string) (string, error) {
	if h == nil {
		h = DefaultPasswordHasher()
	}
	salt := make([]byte, h.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("auth/password: read salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, h.Iterations, h.Memory, h.Parallelism, h.KeyLen)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.Memory, h.Iterations, h.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// VerifyPassword checks password against an encoded hash. Returns nil on
// match, ErrPasswordMismatch on mismatch, or an error if the encoded hash
// is malformed. Uses a constant-time comparison.
func VerifyPassword(password, encoded string) error {
	parts := strings.Split(encoded, "$")
	// Expected: ["", "argon2id", "v=19", "m=...,t=...,p=...", "<salt>", "<hash>"]
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return fmt.Errorf("auth/password: invalid encoded hash format")
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return fmt.Errorf("auth/password: invalid version field: %w", err)
	}
	if version != argon2.Version {
		return fmt.Errorf("auth/password: incompatible argon2 version %d (this build supports %d)", version, argon2.Version)
	}
	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return fmt.Errorf("auth/password: invalid params: %w", err)
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("auth/password: invalid salt: %w", err)
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return fmt.Errorf("auth/password: invalid hash: %w", err)
	}

	got := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(want)))
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return ErrPasswordMismatch
	}
	return nil
}
