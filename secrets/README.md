# Secrets Package

The `secrets` package provides comprehensive credential management for Go applications with enterprise-grade encryption, key management, and pluggable storage backends for multiple clouds.

---

## Table of Contents

- [Installation](#installation)
- [Features](#features)
- [Architecture](#architecture)
- [Architecture Details](ARCHITECTURE.md) 📘
- [Quick Start](#quick-start)
- [Encryption](#encryption)
  - [Algorithms](#algorithms)
  - [Examples](#encryption-examples)
- [Credentials](#credentials)
  - [Credential Types](#credential-types)
  - [Type Detection](#type-detection)
  - [Conversion](#conversion)
- [Key Management](#key-management)
  - [Key Versioning](#key-versioning)
  - [Key Rotation](#key-rotation)
- [Credential Stores](#credential-stores)
  - [Local Store](#local-store)
  - [Cloud Stores](#cloud-stores)
  - [Store Manager](#store-manager)
- [Best Practices](#best-practices)

---

## Installation

```sh
go get oss.nandlabs.io/golly
```

---

## Features

### 🔐 Encryption & Cryptography
- **Multiple Algorithms**: AES-CTR, AES-GCM, ChaCha20-Poly1305
- **Key Sizes**: Support for 128, 192, 256-bit keys
- **Randomized IVs**: Each encryption produces unique ciphertext
- **Authentication**: GCM and ChaCha20-Poly1305 provide authenticated encryption
- **String & Bytes**: Helper methods for common use cases

### 🔑 Credential Management
- **Typed Credentials**: APIKey, Password, Certificate, Token
- **Automatic Detection**: Type detection from metadata
- **Expiration Tracking**: Built-in expiration support
- **Round-trip Serialization**: Credential ↔ Storage conversion

### ⚙️ Key Lifecycle
- **Versioning**: Track multiple key versions
- **Rotation Policies**: Automatic rotation with configurable intervals
- **Active Key Tracking**: Know which key is currently active
- **Old Version Retention**: Maintain configurable number of old versions
- **Thread-Safe**: Concurrent access with built-in locking

### 🏢 Multi-Cloud Support
- **AWS Secrets Manager**: `golly-aws/secrets`
- **GCP Secret Manager**: `golly-gcp/secrets`
- **HashiCorp Vault**: `golly-vault/secrets`
- **Local Storage**: File-based with encryption

### 💾 Pluggable Stores
- **Unified Interface**: All stores implement `Store` interface
- **Caching**: Optional TTL-based caching
- **Metadata**: Provider-specific metadata tracking
- **Version Management**: Cloud-native versioning

---

## Architecture

The secrets package uses a layered architecture:

```
┌─────────────────────────────────────┐
│     Your Application Code           │
├─────────────────────────────────────┤
│  Store Manager (registration)       │
├─────────────────────────────────────┤
│  Store Interface (Get/Write/Delete) │
├──────────┬────────┬────────┬────────┤
│ Local    │  AWS   │  GCP   │ Vault  │
│ Store    │ Store  │ Store  │ Store  │
├──────────┴────────┴────────┴────────┤
│  Encryptor (encryption algorithms)  │
│  KeyManager (versioning/rotation)   │
│  Credential Types (type system)     │
└─────────────────────────────────────┘
```

**Key Components:**
- **Store**: Interface for reading/writing credentials
- **Encryptor**: Pluggable encryption with multiple algorithms
- **KeyManager**: Handles key versioning and rotation
- **Credential**: Unified representation of secrets
- **CredentialType**: Specialized credential types with metadata

---

## Quick Start

### Basic Credential Storage & Retrieval

```go
import "oss.nandlabs.io/golly/secrets"

// Create a local store with encryption
store, err := secrets.NewLocalStore("/path/to/store.json", "my-master-key-16bytes")
if err != nil {
    log.Fatal(err)
}

// Create and store a credential
cred := &secrets.Credential{
    Value:       []byte("my-secret-value"),
    LastUpdated: time.Now(),
    Version:     "1.0",
}

err = store.Write("my-secret", cred, context.Background())

// Retrieve the credential
retrieved, err := store.Get("my-secret", context.Background())
fmt.Println(string(retrieved.Value)) // my-secret-value
```

### Using Credential Types

```go
// Create an API Key credential
apiKeyCred := &secrets.APIKeyCredential{
    Provider:    "AWS",
    KeyID:       "AKIA...",
    KeySecret:   "secret-key",
    Permissions: []string{"s3:*", "ec2:DescribeInstances"},
    CreatedAt:   time.Now(),
    ExpiresAt:   time.Now().AddDate(0, 0, 90),
}

// Convert to generic credential for storage
cred := apiKeyCred.ToCredential("1.0")
store.Write("aws-api-key", cred, ctx)

// Later, retrieve and convert back
stored, _ := store.Get("aws-api-key", ctx)
retrieved := &secrets.APIKeyCredential{}
retrieved.FromCredential(stored)
fmt.Println(retrieved.Provider) // AWS
```

---

## Encryption

### Algorithms

The package supports three encryption algorithms with different trade-offs:

| Algorithm | Authentication | Performance | Use Case |
|-----------|----------------|-------------|----------|
| **AES-CTR** | ❌ No | 🚀 Fastest | Semantic security, requires separate auth |
| **AES-GCM** | ✅ Yes | ⚡ Fast | Authenticated encryption, NIST approved |
| **ChaCha20-Poly1305** | ✅ Yes | ⚡ Fast | Modern AEAD, resistant to timing attacks |

### Encryption Examples

```go
import "oss.nandlabs.io/golly/secrets"

// Using AES-GCM (recommended for most use cases)
encryptor, _ := secrets.NewEncryptor(secrets.AlgorithmAES_GCM, 32) // 256-bit key
key := make([]byte, 32)
rand.Read(key)

plaintext := []byte("sensitive data")
ciphertext, _ := encryptor.Encrypt(key, plaintext)

// Ciphertext is different each time (random IV)
ciphertext2, _ := encryptor.Encrypt(key, plaintext)
// ciphertext != ciphertext2, but both decrypt to plaintext

// Decrypt
decrypted, _ := encryptor.Decrypt(key, ciphertext)
fmt.Println(string(decrypted)) // sensitive data
```

**Encryption Algorithm Details:**

- **AES-CTR**: Counter mode provides semantic security but no authentication. Ideal for scenarios where integrity is verified separately.

- **AES-GCM**: Galois/Counter Mode provides both confidentiality and authentication. Recommended for most use cases. NIST approved.

- **ChaCha20-Poly1305**: Modern AEAD cipher resistant to timing attacks. Faster than AES on systems without hardware acceleration.

---

## Credentials

### Credential Types

The package provides four built-in credential types:

#### 1. APIKeyCredential
```go
cred := &secrets.APIKeyCredential{
    Provider:    "GitHub",
    KeyID:       "ghp_...",
    KeySecret:   "secret",
    Permissions: []string{"repo", "workflow"},
    CreatedAt:   time.Now(),
    ExpiresAt:   time.Now().AddDate(1, 0, 0),
}
```

#### 2. PasswordCredential
```go
cred := &secrets.PasswordCredential{
    Username:  "user@example.com",
    Password:  "encrypted-password",
    CreatedAt: time.Now(),
    ExpiresAt: time.Now().AddDate(0, 3, 0), // 3 months
}
```

#### 3. CertificateCredential
```go
cred := &secrets.CertificateCredential{
    Certificate: x509Cert,
    PrivateKey:  privateKeyPEM,
    CertChain:   []string{intermediateCert},
    CreatedAt:   time.Now(),
}
```

#### 4. TokenCredential
```go
cred := &secrets.TokenCredential{
    Token:     "jwt-token",
    TokenType: "Bearer",
    Scopes:    []string{"openid", "profile"},
    CreatedAt: time.Now(),
    ExpiresAt: time.Now().Add(24 * time.Hour),
}
```

### Type Detection

```go
// Automatically detect credential type from metadata
metadata := map[string]interface{}{
    "type": "api_key",
    "provider": "AWS",
}

credType := secrets.DetectCredentialType(metadata)
fmt.Println(credType) // APIKeyType
```

### Conversion

```go
// Create typed credential
apiKey := &secrets.APIKeyCredential{
    Provider:  "Stripe",
    KeyID:     "pk_live_...",
    KeySecret: "secret",
    CreatedAt: time.Now(),
}

// Convert to generic credential for storage
cred := apiKey.ToCredential("1.0")
store.Write("stripe-key", cred, ctx)

// Retrieve and convert back
stored, _ := store.Get("stripe-key", ctx)
retrieved := &secrets.APIKeyCredential{}
retrieved.FromCredential(stored)

// Check expiration
if retrieved.IsExpired() {
    fmt.Println("Credential has expired")
}
```

---

## Key Management

### Key Versioning

```go
import "oss.nandlabs.io/golly/secrets"

keyManager := secrets.NewLocalKeyManager()

// Create new encryption key
key1, _ := keyManager.CreateKey("master-key-v1", "AES-256", 32)
fmt.Println(key1.KeyID)      // master-key-v1
fmt.Println(key1.Version)    // 1
fmt.Println(key1.IsActive)   // true

// Get the active key for encryption
activeKey, _ := keyManager.GetActiveKey("master-key-v1")
fmt.Println(activeKey.Version) // 1
```

### Key Rotation

```go
// Set rotation policy
policy := &secrets.KeyRotationPolicy{
    AutoRotate:        true,
    RotationIntervalDays: 90,
    MaxKeyAgeDays:     180,
    NotifyBeforeDays:  7,
    KeepOldVersions:   3,
}

keyManager.SetKeyPolicy("master-key-v1", policy)

// Rotate key when needed
newKeyMeta, _ := keyManager.RotateKey("master-key-v1", 32)
fmt.Println(newKeyMeta.Version) // 2

// Check if rotation is due
keysForRotation, _ := keyManager.CheckRotationDue()
for _, key := range keysForRotation {
    fmt.Printf("Key %s needs rotation\n", key.KeyID)
}

// List all keys and their versions
allKeys, _ := keyManager.ListKeys("master-key-v1")
for _, key := range allKeys {
    fmt.Printf("Version %s: %v (Active: %v)\n", 
        key.Version, key.CreatedAt, key.IsActive)
}
```

---

## Credential Stores

### Local Store

Store credentials encrypted locally in JSON files:

```go
// Create local store
store, err := secrets.NewLocalStore("/tmp/secrets.json", "master-key-16bytes")
if err != nil {
    log.Fatal(err)
}

// Write credential
cred := &secrets.Credential{
    Value:       []byte("secret"),
    LastUpdated: time.Now(),
    Version:     "1.0",
}
store.Write("db-password", cred, ctx)

// Read credential
retrieved, _ := store.Get("db-password", ctx)

// Delete credential
store.Delete("db-password", ctx)

// List all credentials
keys, _ := store.List(ctx)
```

**File Format:**
```json
{
  "credentials": {
    "db-password": {
      "value": "base64-encoded-encrypted-data",
      "last_updated": 1682505600,
      "version": "1.0"
    }
  }
}
```

### Cloud Stores

#### AWS Secrets Manager

```go
import "oss.nandlabs.io/golly-aws/secrets"

store, _ := secrets.NewAWSSecretsStore(ctx, &secrets.AWSSecretsStoreConfig{
    Region: "us-east-1",
    TagFilter: map[string]string{
        "app": "myapp",
    },
    CacheTTL: 5 * time.Minute,
})

// Use same Store interface
cred := &secrets.Credential{Value: []byte("secret")}
store.Write("api-key", cred, ctx)
retrieved, _ := store.Get("api-key", ctx)
```

#### GCP Secret Manager

```go
import "oss.nandlabs.io/golly-gcp/secrets"

store, _ := secrets.NewGCPSecretStore(ctx, &secrets.GCPSecretStoreConfig{
    ProjectID: "my-gcp-project",
    Labels: map[string]string{
        "app": "myapp",
        "env": "production",
    },
    CacheTTL: 5 * time.Minute,
})

cred := &secrets.Credential{Value: []byte("secret")}
store.Write("api-key", cred, ctx)
retrieved, _ := store.Get("api-key", ctx)
```

#### HashiCorp Vault

```go
import "oss.nandlabs.io/golly-vault/secrets"

store, _ := secrets.NewVaultStore(&secrets.VaultStoreConfig{
    Address:   "https://vault.example.com:8200",
    Token:     "s.your-vault-token",
    Version:   "v2",
    BasePath:  "secret",
    CacheTTL:  5 * time.Minute,
})

cred := &secrets.Credential{Value: []byte("secret")}
store.Write("api-key", cred, ctx)
retrieved, _ := store.Get("api-key", ctx)
```

### Store Manager

Register and retrieve stores by provider name:

```go
// Get the store manager
manager := secrets.GetManager()

// Register stores
manager.Register("local", localStore)
manager.Register("aws", awsStore)
manager.Register("gcp", gcpStore)
manager.Register("vault", vaultStore)

// Retrieve by provider
store := manager.Get("aws")
cred, _ := store.Get("my-secret", ctx)
```

---

## Best Practices

### 🔒 Security

1. **Master Key Management**
   - Store master keys in environment variables or secure vaults
   - Rotate master keys periodically
   - Never commit master keys to version control

2. **Key Rotation**
   - Enable automatic key rotation with appropriate intervals
   - Maintain old keys for decrypting existing credentials
   - Monitor for keys exceeding maximum age

3. **Encryption Algorithm Selection**
   - Use AES-GCM or ChaCha20-Poly1305 for authentication
   - Use AES-CTR only when integrity is verified separately
   - 256-bit keys for maximum security

4. **Credential Expiration**
   - Set expiration dates on credentials
   - Monitor and refresh expiring credentials
   - Use `IsExpired()` to check before use

### 🚀 Performance

1. **Caching**
   - Enable TTL-based caching for frequently accessed credentials
   - Adjust TTL based on security requirements
   - Clear cache when credentials are updated

2. **Cloud Stores**
   - Batch operations where possible
   - Use regional endpoints for lower latency
   - Consider VPC endpoints for reduced costs

3. **Local Storage**
   - Use SSD storage for better performance
   - Consider in-memory caching for hot credentials
   - Monitor file size as it grows

### 🛠️ Operational

1. **Error Handling**
   ```go
   cred, err := store.Get("my-secret", ctx)
   if err != nil {
       // Handle not found
       // Handle permission denied
       // Handle network errors
       log.Printf("Failed to retrieve credential: %v", err)
   }
   ```

2. **Audit Logging**
   - Log all credential access (who, what, when)
   - Monitor for unusual access patterns
   - Retain logs for compliance

3. **Monitoring**
   - Track key rotation frequency
   - Monitor credential expiration
   - Alert on failed credential access

### 🔄 Migration

When switching between stores:

```go
// Read from old store
oldStore := secrets.NewLocalStore("/path/old.json", oldKey)
keys, _ := oldStore.List(ctx)

// Write to new store
newStore := secrets.NewAWSSecretsStore(ctx, &config)

for _, key := range keys {
    cred, _ := oldStore.Get(key, ctx)
    newStore.Write(key, cred, ctx)
}
```

---

## See Also

- [Architecture Details](ARCHITECTURE.md) - Deep dive into system design
- [Cloud Integration Guide](INTEGRATION_GUIDE.md) - Multi-cloud setup and strategies
- [AWS Secrets Manager](../golly-aws/secrets/README.md) - AWS integration
- [GCP Secret Manager](../golly-gcp/secrets/README.md) - GCP integration
- [HashiCorp Vault](../golly-vault/secrets/README.md) - Vault integration
- [Source Code](.) - Implementation details

---

## Contributing

Report issues or contribute improvements to the secrets package via [GitHub](https://github.com/nandlabs/golly/issues).
