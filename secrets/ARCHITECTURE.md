# Secrets Package Architecture

This document describes the architectural design of the Golly `secrets` package.

## Overview

The secrets package provides a comprehensive credential management system with the following design principles:

- **Pluggable**: Multiple implementations of the `Store` interface
- **Type-Safe**: Dedicated credential types with automatic conversion
- **Secure**: Multiple encryption algorithms with key versioning
- **Multi-Cloud**: Support for AWS, GCP, HashiCorp Vault, and local storage
- **Observable**: Metadata tracking for audit and compliance
- **Performant**: Optional caching with TTL support

## Layer Architecture

```
┌────────────────────────────────────────────────────────────┐
│                   Application Code                         │
├────────────────────────────────────────────────────────────┤
│              Store Manager Registry                        │
│     (Get/Register credential store implementations)        │
├──────────────┬──────────────┬──────────────┬──────────────┤
│  LocalStore  │ AWSStore     │ GCPStore     │ VaultStore   │
│  (File-based)│ (AWS SecMgr)  │ (GCP SecMgr) │ (Vault KV)   │
├────────────────────────────────────────────────────────────┤
│          Store Interface (Get/Write/Delete/List)          │
├────────────────────────────────────────────────────────────┤
│                    Credential System                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Credential (core struct)                           │   │
│  │  - Value: []byte                                    │   │
│  │  - Version: string                                  │   │
│  │  - LastUpdated: time.Time                           │   │
│  │  - MetaData: map[string]interface{}                 │   │
│  └─────────────────────────────────────────────────────┘   │
│                         ▲                                   │
│  ┌──────────┬──────────┼──────────┬────────────┐            │
│  │          │          │          │            │            │
│ Impl    APIKey    Password    Certificate   Token          │
│         Credential  Credential   Credential  Credential    │
├────────────────────────────────────────────────────────────┤
│          Encryption System (Pluggable)                     │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Encryptor Interface                                │   │
│  │  - Encrypt(key, plaintext) -> ciphertext            │   │
│  │  - Decrypt(key, ciphertext) -> plaintext            │   │
│  │  - Algorithm() -> string                            │   │
│  │  - KeySize() -> int                                 │   │
│  └─────────────────────────────────────────────────────┘   │
│       ▲              ▲              ▲                       │
│   AES_CTR        AES_GCM       ChaCha20Poly1305           │
├────────────────────────────────────────────────────────────┤
│          Key Management System                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  KeyManager Interface                               │   │
│  │  - CreateKey(), RotateKey(), GetActiveKey()         │   │
│  │  - SetKeyPolicy(), CheckRotationDue()               │   │
│  │  - RevokeKey(), ListKeys()                          │   │
│  └─────────────────────────────────────────────────────┘   │
│                       ▲                                     │
│                LocalKeyManager (in-process)                │
│                                                             │
│  Key Rotation Policy:                                       │
│  - AutoRotate: bool                                         │
│  - RotationIntervalDays: int                                │
│  - MaxKeyAgeDays: int                                       │
│  - NotifyBeforeDays: int                                    │
│  - KeepOldVersions: int                                     │
├────────────────────────────────────────────────────────────┤
│         Standard Library & External Dependencies           │
│  crypto/aes, crypto/cipher, crypto/rand                    │
│  golang.org/x/crypto/chacha20poly1305                      │
│  AWS SDK v2, GCP Client Libraries, Vault API               │
└────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Store Interface

**Location**: `store.go`

**Purpose**: Defines the contract for credential storage backends

```go
type Store interface {
    Get(key string, ctx context.Context) (*Credential, error)
    Write(key string, credential *Credential, ctx context.Context) error
    Delete(key string, ctx context.Context) error
    List(ctx context.Context) ([]string, error)
    Provider() string
}
```

**Implementations**:
- **LocalStore**: File-based storage with local encryption
- **AWSSecretsStore**: AWS Secrets Manager backend
- **GCPSecretStore**: Google Cloud Secret Manager backend
- **VaultStore**: HashiCorp Vault KV engine backend

### 2. Credential System

**Core Types**:

#### Credential (Generic)
```go
type Credential struct {
    Value                 []byte
    LastUpdated          time.Time
    Version              string
    EncryptionKeyID      string
    EncryptionKeyVersion int
    MetaData             map[string]interface{}
}
```

**Specialized Types** (implement `CredentialType` interface):

1. **APIKeyCredential**
   - Use Case: API tokens, access keys
   - Fields: Provider, KeyID, KeySecret, Permissions, ExpiresAt
   - Expiration: Yes

2. **PasswordCredential**
   - Use Case: User passwords, database credentials
   - Fields: Username, Password, ExpiresAt
   - Expiration: Yes

3. **CertificateCredential**
   - Use Case: TLS certificates, client certs
   - Fields: Certificate (x509), PrivateKey, CertChain
   - Expiration: Extracted from certificate

4. **TokenCredential**
   - Use Case: JWT, OAuth tokens
   - Fields: Token, TokenType, Scopes, ExpiresAt
   - Expiration: Yes

**Type Conversion Flow**:
```
Typed Credential → Generic Credential → Storage
        ↓                   ↓
   ToCredential()     Store.Write()
        
Storage → Generic Credential → Typed Credential
  ↓              ↓
Store.Get()  FromCredential()
```

**Type Detection**:
- Automatic detection from metadata
- Fallback to generic Credential if type unknown
- Enables polymorphic handling

### 3. Encryption System

**Design Pattern**: Pluggable Factory

**Interface**:
```go
type Encryptor interface {
    Encrypt(key, plaintext []byte) ([]byte, error)
    Decrypt(key, ciphertext []byte) ([]byte, error)
    Algorithm() string
    KeySize() int
}
```

**Algorithms**:

#### AES-CTR
- **Mode**: Counter (CTR)
- **Key Sizes**: 16, 24, 32 bytes (128, 192, 256 bits)
- **IV Size**: 16 bytes (random per encryption)
- **Authentication**: No (no AEAD tag)
- **Pros**: Fast, semantic security
- **Cons**: No authentication, requires separate integrity check

#### AES-GCM
- **Mode**: Galois/Counter Mode (authenticated)
- **Key Sizes**: 16, 24, 32 bytes (128, 192, 256 bits)
- **IV Size**: 12 bytes (recommended)
- **Authentication**: Yes (16-byte AEAD tag)
- **Pros**: NIST approved, authenticated encryption
- **Cons**: Slower than CTR

#### ChaCha20-Poly1305
- **Cipher**: ChaCha20 (modern stream cipher)
- **AEAD**: Poly1305 (universal hash)
- **Key Size**: 32 bytes (256 bits only)
- **IV Size**: 12 bytes (nonce)
- **Authentication**: Yes (16-byte AEAD tag)
- **Pros**: Timing-attack resistant, fast on all platforms
- **Cons**: 256-bit key only

**Factory Pattern**:
```go
func NewEncryptor(algo EncryptionAlgorithm, keySize int) (Encryptor, error)
// Returns appropriate implementation based on algorithm
```

**IV Randomization**:
- Each encryption generates random IV
- Different ciphertext for same plaintext
- IV prepended to ciphertext for decryption
- Provides semantic security

### 4. Key Management System

**Design Pattern**: Manager with policies

**Components**:

#### KeyManager Interface
```go
type KeyManager interface {
    CreateKey(keyID string, algo string, keySize int) (*KeyMetadata, error)
    GetActiveKey(keyID string) (*KeyMetadata, error)
    GetKey(keyID string, version string) (*KeyMetadata, error)
    ListKeys(keyID string) ([]*KeyMetadata, error)
    ListActiveKeys() ([]*KeyMetadata, error)
    RotateKey(keyID string, keySize int) (*KeyMetadata, error)
    SetKeyPolicy(keyID string, policy *KeyRotationPolicy) error
    GetKeyPolicy(keyID string) (*KeyRotationPolicy, error)
    RevokeKey(keyID string, version string) error
    CheckRotationDue() ([]*KeyMetadata, error)
}
```

#### LocalKeyManager
**In-Memory Storage**:
```go
type LocalKeyManager struct {
    keys     map[string]*keyVersions      // keyID -> versions
    policies map[string]*KeyRotationPolicy
    mutex    sync.RWMutex                 // Thread-safe
}
```

**Versioning Strategy**:
- Each key maintains multiple versions
- One active version at a time
- Old versions retained for decryption
- Thread-safe operations with RWMutex

**Rotation Policy**:
```go
type KeyRotationPolicy struct {
    AutoRotate           bool
    RotationIntervalDays int
    RotationSchedule     string     // cron-like
    MaxKeyAgeDays        int
    NotifyBeforeDays     int
    KeepOldVersions      int
}
```

**Rotation Triggers**:
1. Manual: `RotateKey()` call
2. Scheduled: `RotationIntervalDays` elapsed
3. Age-based: `MaxKeyAgeDays` exceeded
4. Emergency: `RevokeKey()` for compromised keys

### 5. Store Manager

**Purpose**: Registry for credential stores

**Pattern**: Manager with factory registration

```go
type StoreManager interface {
    Register(provider string, store Store) error
    Get(provider string) Store
    // ... other methods
}
```

**Usage**:
- Register multiple stores
- Select store by provider name
- Enable provider-agnostic code

## Data Flows

### Writing a Credential

```
Application
    ↓
Create Credential/CredentialType
    ↓
[Optional] Encrypt Value
    ↓
Store.Write()
    ↓
Provider-Specific Serialization
    ↓
Cloud Provider / Local File
```

### Reading a Credential

```
Application
    ↓
Store.Get()
    ↓
Provider-Specific Deserialization
    ↓
[Optional] Decrypt Value
    ↓
[Optional] Convert to CredentialType
    ↓
Return Credential
```

### Key Rotation

```
Time elapsed / Manual trigger
    ↓
KeyManager.CheckRotationDue() / RotateKey()
    ↓
Create new key version
    ↓
Set as active
    ↓
Retain old versions
    ↓
[Optional] Re-encrypt with new key
```

## Thread Safety

**RWMutex Usage**:
- LocalKeyManager: RWMutex protects key map and policies
- Store Implementations: Optional locking for cache
- Concurrent reads: Multiple threads OK
- Concurrent writes: Serialized access

**Example**:
```go
type LocalKeyManager struct {
    mutex sync.RWMutex
    // ...
}

// Read operation
func (km *LocalKeyManager) GetActiveKey(keyID string) (*KeyMetadata, error) {
    km.mutex.RLock()      // Multiple readers OK
    defer km.mutex.RUnlock()
    // ...
}

// Write operation
func (km *LocalKeyManager) RotateKey(keyID string, keySize int) (*KeyMetadata, error) {
    km.mutex.Lock()       // Exclusive access
    defer km.mutex.Unlock()
    // ...
}
```

## Error Handling

**Error Categories**:

1. **Input Validation Errors**
   - Invalid key sizes
   - Malformed credentials
   - Missing required fields

2. **Encryption Errors**
   - Invalid cipher text
   - Authentication tag mismatch
   - Key derivation failures

3. **Storage Errors**
   - Credential not found
   - Permission denied
   - Network/connection errors
   - Corrupted data

4. **Key Management Errors**
   - Key not found
   - Duplicate key ID
   - Invalid rotation policy
   - Key revocation errors

**Error Wrapping**:
- All errors wrapped with context
- Original error preserved with `fmt.Errorf("%w")`
- Enables programmatic error handling

## Performance Considerations

### Caching Strategy

**TTL-Based Cache**:
- Optional per-store
- Reduces cloud API calls
- Configurable TTL
- Manual cache invalidation

**Trade-offs**:
- **Faster**: Reduces latency and costs
- **Stale Data Risk**: May serve outdated credentials
- **Mitigation**: Appropriate TTL based on rotation frequency

### Encryption Performance

**Algorithm Speed** (approximate):
1. AES-CTR: Fastest
2. AES-GCM: ~same as CTR
3. ChaCha20-Poly1305: Fast, platform-dependent

**Hardware Acceleration**:
- AES: Hardware accelerated on most platforms
- ChaCha20: No hardware acceleration (OK with SW)

### Network Optimization

**For Cloud Stores**:
- Use regional endpoints
- Consider VPC endpoints
- Batch operations when possible
- Connection pooling (built-in to SDKs)

## Security Considerations

### Master Key Management

**Local Store**:
- Master key required for all operations
- Separate from data (environment variable recommended)
- Never log or serialize

**Cloud Stores**:
- Use cloud provider's key management (KMS)
- Leverage cloud-native encryption
- Automatic key rotation support

### Credential Lifecycle

1. **Creation**: Metadata set, expiration configured
2. **Storage**: Encrypted at rest by default
3. **Retrieval**: Expiration checked before use
4. **Rotation**: Periodic or manual refresh
5. **Revocation**: Immediate invalidation
6. **Deletion**: Permanent removal

### Threat Model

**Threats Mitigated**:
- Plaintext credential exposure: Encryption
- Credential reuse: Version tracking, expiration
- Key compromise: Versioning, rotation
- Unauthorized access: Cloud provider IAM
- Data tampering: Authentication tags (GCM, ChaCha20)

**Threats Not Mitigated**:
- Memory leaks: Application responsibility
- Timing attacks: Reduced but not eliminated
- Compromised machine: Out of scope
- Insider access: Cloud provider controls

## Testing Strategy

**Unit Tests** (~77 tests):
- Encryption correctness
- Key versioning
- Type conversion
- Type detection
- Credential expiration

**Integration Tests**:
- Local store read/write
- Encryption round-trips
- Key rotation workflows

**Mock Tests**:
- Cloud store operations (simulated)
- Error conditions
- Concurrent access

## Extension Points

### Adding New Encryption Algorithm

1. Implement `Encryptor` interface
2. Register in `NewEncryptor()` factory
3. Add tests
4. Update documentation

### Adding New Store Implementation

1. Implement `Store` interface
2. Handle provider-specific serialization
3. Implement caching if needed
4. Add configuration struct
5. Register with manager
6. Add tests and documentation

### Adding New Credential Type

1. Define struct with metadata fields
2. Implement `CredentialType` interface:
   - `ToCredential(version string) *Credential`
   - `FromCredential(cred *Credential) error`
   - `IsExpired() bool`
3. Add to `DetectCredentialType()` factory
4. Add tests
5. Update documentation

## Future Enhancements

1. **Audit Logging**: Track all credential access
2. **Credential Delegation**: Support delegation patterns
3. **Hardware Security Modules**: HSM key storage
4. **Certificate Management**: Auto-renewal support
5. **Multi-Tenancy**: Tenant isolation
6. **Metrics & Observability**: Prometheus metrics
7. **Compliance Reporting**: Audit trail exports

---

For implementation details, see the source code. For usage examples, see [README.md](README.md).
