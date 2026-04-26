package secrets

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// KeyMetadata contains information about a managed key
type KeyMetadata struct {
	KeyID     string    // Unique identifier for the key
	Version   int       // Version number of this key
	CreatedAt time.Time // When the key was created
	RotatedAt time.Time // When the key was last rotated (if ever)
	ExpiresAt time.Time // When the key expires
	IsActive  bool      // Whether this key is currently in use
	Algorithm string    // Encryption algorithm used with this key
	KeySize   int       // Key size in bits
}

// KeyRotationPolicy defines when and how keys should be rotated
type KeyRotationPolicy struct {
	// AutoRotate enables automatic key rotation
	AutoRotate bool

	// RotationIntervalDays specifies how often to rotate keys (in days)
	// Set to 0 to disable interval-based rotation
	RotationIntervalDays int

	// RotationSchedule is a cron-like expression for scheduled rotations
	// Examples: "0 0 * * 0" (weekly), "0 0 1 * *" (monthly)
	// Set to empty string to disable scheduled rotation
	RotationSchedule string

	// MaxKeyAgeDays is the maximum age of a key before it must be rotated
	// Used as a safety mechanism in addition to RotationIntervalDays
	MaxKeyAgeDays int

	// NotifyBefore specifies how many days before expiration to notify (0 = never)
	NotifyBeforeDays int

	// KeepOldVersions specifies how many old key versions to keep for decryption
	KeepOldVersions int
}

// KeyManager handles key lifecycle including rotation, versioning, and storage
type KeyManager interface {
	// CreateKey creates a new key with the given ID and algorithm
	CreateKey(ctx context.Context, keyID string, algorithm string, keySize int) error

	// GetActiveKey returns the currently active key for the given key ID
	GetActiveKey(ctx context.Context, keyID string) (*KeyMetadata, error)

	// GetKey retrieves a specific key version
	GetKey(ctx context.Context, keyID string, version int) (*KeyMetadata, error)

	// ListKeys returns all key versions for a given key ID
	ListKeys(ctx context.Context, keyID string) ([]KeyMetadata, error)

	// ListActiveKeys returns all currently active keys in the system
	ListActiveKeys(ctx context.Context) ([]KeyMetadata, error)

	// RotateKey performs a key rotation, creating a new key version
	// and optionally re-encrypting credentials with the new key
	RotateKey(ctx context.Context, keyID string) error

	// SetKeyPolicy sets the rotation policy for a key
	SetKeyPolicy(ctx context.Context, keyID string, policy *KeyRotationPolicy) error

	// GetKeyPolicy retrieves the rotation policy for a key
	GetKeyPolicy(ctx context.Context, keyID string) (*KeyRotationPolicy, error)

	// RevokeKey revokes a key version, preventing further use
	RevokeKey(ctx context.Context, keyID string, version int) error

	// CheckRotationDue checks if any keys are due for rotation based on their policies
	CheckRotationDue(ctx context.Context) ([]string, error)
}

// LocalKeyManager implements KeyManager for local file-based key storage
type LocalKeyManager struct {
	keys      map[string]*keyVersions                      // Map of keyID -> versions
	policies  map[string]*KeyRotationPolicy                // Key rotation policies
	mutex     sync.RWMutex                                 // Thread safety
	keyStore  string                                       // Path to key storage directory
	keyLoader func(id string, version int) ([]byte, error) // Custom key loader
}

// keyVersions holds all versions of a key
type keyVersions struct {
	versions map[int]*KeyMetadata
	latest   int // Version number of the latest key
}

// NewLocalKeyManager creates a new local key manager
func NewLocalKeyManager(keyStorePath string) *LocalKeyManager {
	return &LocalKeyManager{
		keys:     make(map[string]*keyVersions),
		policies: make(map[string]*KeyRotationPolicy),
		keyStore: keyStorePath,
	}
}

// CreateKey creates a new key with initial version 1
func (km *LocalKeyManager) CreateKey(ctx context.Context, keyID string, algorithm string, keySize int) error {
	km.mutex.Lock()
	defer km.mutex.Unlock()

	if _, exists := km.keys[keyID]; exists {
		return fmt.Errorf("key %q already exists", keyID)
	}

	metadata := &KeyMetadata{
		KeyID:     keyID,
		Version:   1,
		CreatedAt: time.Now(),
		IsActive:  true,
		Algorithm: algorithm,
		KeySize:   keySize,
	}

	km.keys[keyID] = &keyVersions{
		versions: map[int]*KeyMetadata{1: metadata},
		latest:   1,
	}

	// Set default rotation policy
	km.policies[keyID] = &KeyRotationPolicy{
		AutoRotate:           false,
		RotationIntervalDays: 90,
		MaxKeyAgeDays:        365,
		KeepOldVersions:      5,
	}

	return nil
}

// GetActiveKey returns the current active key for a key ID
func (km *LocalKeyManager) GetActiveKey(ctx context.Context, keyID string) (*KeyMetadata, error) {
	km.mutex.RLock()
	defer km.mutex.RUnlock()

	versions, exists := km.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key %q not found", keyID)
	}

	if metadata, exists := versions.versions[versions.latest]; exists {
		return metadata, nil
	}

	return nil, fmt.Errorf("no active version for key %q", keyID)
}

// GetKey retrieves a specific key version
func (km *LocalKeyManager) GetKey(ctx context.Context, keyID string, version int) (*KeyMetadata, error) {
	km.mutex.RLock()
	defer km.mutex.RUnlock()

	versions, exists := km.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key %q not found", keyID)
	}

	metadata, exists := versions.versions[version]
	if !exists {
		return nil, fmt.Errorf("version %d of key %q not found", version, keyID)
	}

	return metadata, nil
}

// ListKeys returns all versions of a key
func (km *LocalKeyManager) ListKeys(ctx context.Context, keyID string) ([]KeyMetadata, error) {
	km.mutex.RLock()
	defer km.mutex.RUnlock()

	versions, exists := km.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key %q not found", keyID)
	}

	result := make([]KeyMetadata, 0, len(versions.versions))
	for _, metadata := range versions.versions {
		result = append(result, *metadata)
	}

	return result, nil
}

// ListActiveKeys returns all currently active keys
func (km *LocalKeyManager) ListActiveKeys(ctx context.Context) ([]KeyMetadata, error) {
	km.mutex.RLock()
	defer km.mutex.RUnlock()

	var result []KeyMetadata
	for _, versions := range km.keys {
		if metadata, exists := versions.versions[versions.latest]; exists && metadata.IsActive {
			result = append(result, *metadata)
		}
	}

	return result, nil
}

// RotateKey creates a new version of the key
func (km *LocalKeyManager) RotateKey(ctx context.Context, keyID string) error {
	km.mutex.Lock()
	defer km.mutex.Unlock()

	versions, exists := km.keys[keyID]
	if !exists {
		return fmt.Errorf("key %q not found", keyID)
	}

	oldMetadata := versions.versions[versions.latest]
	newVersion := versions.latest + 1

	newMetadata := &KeyMetadata{
		KeyID:     keyID,
		Version:   newVersion,
		CreatedAt: time.Now(),
		RotatedAt: time.Now(),
		IsActive:  true,
		Algorithm: oldMetadata.Algorithm,
		KeySize:   oldMetadata.KeySize,
	}

	versions.versions[newVersion] = newMetadata
	versions.latest = newVersion

	// Mark old version as no longer active but keep it for decryption
	oldMetadata.IsActive = false

	// Cleanup old versions based on retention policy
	if policy, exists := km.policies[keyID]; exists && policy.KeepOldVersions > 0 {
		km.cleanupOldVersions(versions, policy.KeepOldVersions)
	}

	return nil
}

// cleanupOldVersions removes versions older than the retention limit
func (km *LocalKeyManager) cleanupOldVersions(versions *keyVersions, keepCount int) {
	if len(versions.versions) <= keepCount {
		return
	}

	// Find oldest versions to delete
	var toDelete []int
	for version := range versions.versions {
		// Don't delete the latest version
		if version != versions.latest && len(versions.versions)-len(toDelete) > keepCount {
			toDelete = append(toDelete, version)
		}
	}

	for _, version := range toDelete {
		delete(versions.versions, version)
	}
}

// SetKeyPolicy sets the rotation policy for a key
func (km *LocalKeyManager) SetKeyPolicy(ctx context.Context, keyID string, policy *KeyRotationPolicy) error {
	km.mutex.Lock()
	defer km.mutex.Unlock()

	if _, exists := km.keys[keyID]; !exists {
		return fmt.Errorf("key %q not found", keyID)
	}

	km.policies[keyID] = policy
	return nil
}

// GetKeyPolicy retrieves the rotation policy for a key
func (km *LocalKeyManager) GetKeyPolicy(ctx context.Context, keyID string) (*KeyRotationPolicy, error) {
	km.mutex.RLock()
	defer km.mutex.RUnlock()

	policy, exists := km.policies[keyID]
	if !exists {
		return nil, fmt.Errorf("no policy found for key %q", keyID)
	}

	return policy, nil
}

// RevokeKey marks a key version as revoked
func (km *LocalKeyManager) RevokeKey(ctx context.Context, keyID string, version int) error {
	km.mutex.Lock()
	defer km.mutex.Unlock()

	versions, exists := km.keys[keyID]
	if !exists {
		return fmt.Errorf("key %q not found", keyID)
	}

	metadata, exists := versions.versions[version]
	if !exists {
		return fmt.Errorf("version %d of key %q not found", version, keyID)
	}

	if version == versions.latest {
		return fmt.Errorf("cannot revoke the active key version")
	}

	metadata.IsActive = false
	return nil
}

// CheckRotationDue checks if any keys are due for rotation
func (km *LocalKeyManager) CheckRotationDue(ctx context.Context) ([]string, error) {
	km.mutex.RLock()
	defer km.mutex.RUnlock()

	var dueForRotation []string
	now := time.Now()

	for keyID, versions := range km.keys {
		activeMetadata := versions.versions[versions.latest]
		policy := km.policies[keyID]

		// Check against MaxKeyAgeDays
		if policy.MaxKeyAgeDays > 0 {
			maxAge := time.Duration(policy.MaxKeyAgeDays) * 24 * time.Hour
			if now.Sub(activeMetadata.CreatedAt) > maxAge {
				dueForRotation = append(dueForRotation, keyID)
				continue
			}
		}

		// Check against RotationIntervalDays
		if policy.RotationIntervalDays > 0 && !activeMetadata.RotatedAt.IsZero() {
			interval := time.Duration(policy.RotationIntervalDays) * 24 * time.Hour
			if now.Sub(activeMetadata.RotatedAt) > interval {
				dueForRotation = append(dueForRotation, keyID)
				continue
			}
		}

		// Check expiration
		if !activeMetadata.ExpiresAt.IsZero() && now.After(activeMetadata.ExpiresAt) {
			dueForRotation = append(dueForRotation, keyID)
		}
	}

	return dueForRotation, nil
}
