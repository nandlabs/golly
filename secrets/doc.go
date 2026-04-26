// Package secrets provides a comprehensive secret management system for Go applications.
// It offers pluggable storage backends, encryption, and credential management.
//
// # Core Features
//
// - Encryption: AES-CTR encryption for protecting secrets at rest
// - Storage: Pluggable Store interface supporting local file-based and cloud backends
// - Credentials: Type-safe credential management with metadata support
// - Registry: StoreManager for managing multiple store instances
//
// # Encryption Design
//
// The package uses AES in CTR (Counter) mode for encryption, which provides:
//   - Semantic security: The same plaintext encrypts to different ciphertexts
//   - IV handling: A random 16-byte IV is generated for each encryption and prepended to ciphertext
//   - Key support: 128-bit (16 bytes), 192-bit (24 bytes), and 256-bit (32 bytes) keys
//
// ## Important Security Assumptions
//
//  1. Master Key Storage: The application is responsible for securely storing the master key
//     used by LocalStore. This should be provisioned via:
//     - Environment variables (development)
//     - AWS Secrets Manager, GCP Secret Manager, or HashiCorp Vault (production)
//     - Key management systems (KMS) for high-security deployments
//
//  2. IV Randomness: Each encryption operation generates a cryptographically random IV.
//     This ensures that identical plaintexts produce different ciphertexts, preventing
//     patterns from leaking information. Tests verify IV randomization.
//
//  3. Authentication: CTR mode does not provide authentication. Use AES-GCM for
//     authenticated encryption when available (planned for future versions).
//
//  4. IV Prepending: The IV is prepended to the ciphertext (first 16 bytes). Both
//     encryption and decryption rely on this layout. Corrupted or truncated ciphertexts
//     will fail decryption with "encrypted block size is too short" error.
//
// # Usage
//
// Basic encryption:
//
//	key := "thisisamasterkeyABC" // Should be 16, 24, or 32 bytes
//	plaintext := "secret data"
//
//	ciphertext, err := AesEncryptStr(key, plaintext)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	decrypted, err := AesDecryptStr(ciphertext, key)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Using LocalStore:
//
//	store, err := NewLocalStore("./credentials.dat", "thisisamasterkeyABC")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	cred := &Credential{
//	    Value:       []byte("api-secret-value"),
//	    LastUpdated: time.Now(),
//	    Version:     "1.0",
//	}
//
//	err = store.Write("api-key", cred, context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	retrieved, err := store.Get("api-key", context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Thread Safety
//
// LocalStore uses sync.RWMutex to protect concurrent access to the credentials map.
// Multiple goroutines can safely read and write credentials simultaneously.
//
// # Future Enhancements
//
// Planned additions include:
//   - Additional encryption algorithms (AES-GCM, ChaCha20-Poly1305)
//   - Key rotation and versioning
//   - Cloud secret store integrations (AWS Secrets Manager, GCP Secret Manager, Vault)
//   - Audit logging for credential access and modifications
//   - Specialized credential types (API keys, passwords, certificates, tokens)
//   - Integration with clients package for dynamic credential resolution
package secrets
