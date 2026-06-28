package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

// Session represents an opaque server-side session (not a JWT). The actual
// session id is a random 32-byte URL-safe string; the data lives in a Store
// keyed by it.
type Session struct {
	ID        string
	Subject   string // user id, etc.
	CreatedAt time.Time
	ExpiresAt time.Time      // absolute expiry
	LastSeen  time.Time      // for sliding expiry
	Data      map[string]any // app-specific
}

// Active reports whether the session is non-empty, non-expired.
func (s *Session) Active(now time.Time) bool {
	return s != nil && s.ID != "" && now.Before(s.ExpiresAt)
}

// SessionStore is the persistence interface for opaque sessions. The default
// in-memory MemorySessionStore is suitable for tests and single-process apps.
type SessionStore interface {
	Get(id string) (*Session, bool)
	Put(s *Session) error
	Delete(id string) error
}

// MemorySessionStore is a goroutine-safe in-memory SessionStore.
type MemorySessionStore struct {
	mu sync.RWMutex
	m  map[string]*Session
}

// NewMemorySessionStore returns an empty in-memory store.
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{m: make(map[string]*Session)}
}

func (s *MemorySessionStore) Get(id string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.m[id]
	return sess, ok
}

func (s *MemorySessionStore) Put(sess *Session) error {
	if sess == nil || sess.ID == "" {
		return errors.New("auth/session: empty session id")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[sess.ID] = sess
	return nil
}

func (s *MemorySessionStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, id)
	return nil
}

// SessionConfig governs how the SessionManager mints and renews sessions.
type SessionConfig struct {
	// Absolute is the maximum total lifetime of a session, regardless of
	// activity. After this, the session must be re-created.
	Absolute time.Duration
	// Sliding extends ExpiresAt on Touch up to Absolute. Set to 0 to
	// disable sliding (sessions are pure absolute-expiry tokens).
	Sliding time.Duration
}

// SessionManager mints and validates opaque sessions backed by a SessionStore.
type SessionManager struct {
	store SessionStore
	cfg   SessionConfig
	now   func() time.Time
}

// NewSessionManager returns a SessionManager. Uses NewMemorySessionStore if
// store is nil.
func NewSessionManager(store SessionStore, cfg SessionConfig) *SessionManager {
	if store == nil {
		store = NewMemorySessionStore()
	}
	if cfg.Absolute <= 0 {
		cfg.Absolute = 24 * time.Hour
	}
	return &SessionManager{store: store, cfg: cfg, now: time.Now}
}

// New mints a fresh session for subject and stores it.
func (m *SessionManager) New(subject string, data map[string]any) (*Session, error) {
	id, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	now := m.now()
	// With sliding configured, initial expiry is now+Sliding (capped at
	// hardCap=now+Absolute). Without sliding, it's the hardCap.
	exp := now.Add(m.cfg.Absolute)
	if m.cfg.Sliding > 0 {
		if sliding := now.Add(m.cfg.Sliding); sliding.Before(exp) {
			exp = sliding
		}
	}
	sess := &Session{
		ID:        id,
		Subject:   subject,
		CreatedAt: now,
		LastSeen:  now,
		ExpiresAt: exp,
		Data:      data,
	}
	if err := m.store.Put(sess); err != nil {
		return nil, err
	}
	return sess, nil
}

// Validate looks up id, verifies it's active. Returns (nil, false) when not
// found or expired (also deletes expired sessions).
func (m *SessionManager) Validate(id string) (*Session, bool) {
	sess, ok := m.store.Get(id)
	if !ok {
		return nil, false
	}
	if !sess.Active(m.now()) {
		_ = m.store.Delete(id)
		return nil, false
	}
	return sess, true
}

// Touch records activity. If Sliding is non-zero, extends ExpiresAt by
// Sliding (capped at CreatedAt+Absolute) so active sessions don't expire.
func (m *SessionManager) Touch(id string) error {
	sess, ok := m.Validate(id)
	if !ok {
		return errors.New("auth/session: session not active")
	}
	now := m.now()
	sess.LastSeen = now
	if m.cfg.Sliding > 0 {
		newExp := now.Add(m.cfg.Sliding)
		hardCap := sess.CreatedAt.Add(m.cfg.Absolute)
		if newExp.After(hardCap) {
			newExp = hardCap
		}
		sess.ExpiresAt = newExp
	}
	return m.store.Put(sess)
}

// Revoke deletes the session, immediately invalidating it.
func (m *SessionManager) Revoke(id string) error { return m.store.Delete(id) }

// randomToken returns a base64.RawURL-encoded random byte string of nBytes
// of entropy (the encoded string is longer than nBytes).
func randomToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
