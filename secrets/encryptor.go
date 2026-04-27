package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

// EncryptionAlgorithm represents the name of an encryption algorithm
type EncryptionAlgorithm string

const (
	AlgorithmAES_CTR           EncryptionAlgorithm = "AES-CTR"
	AlgorithmAES_GCM           EncryptionAlgorithm = "AES-GCM"
	AlgorithmChaCha20_Poly1305 EncryptionAlgorithm = "ChaCha20-Poly1305"
)

// Encryptor interface defines the contract for encryption algorithms
type Encryptor interface {
	// Encrypt encrypts plaintext with the given key and returns ciphertext
	Encrypt(key []byte, plaintext []byte) (ciphertext []byte, err error)

	// Decrypt decrypts ciphertext with the given key and returns plaintext
	Decrypt(key []byte, ciphertext []byte) (plaintext []byte, err error)

	// Algorithm returns the name of this encryption algorithm
	Algorithm() string

	// KeySize returns the required key size in bytes
	KeySize() int
}

// AES_CTR_Encryptor implements Encryptor using AES in CTR mode
// Note: CTR mode provides no authentication. For authenticated encryption, use AES_GCM_Encryptor
type AES_CTR_Encryptor struct {
	keySize int
}

// NewAES_CTR_Encryptor creates a new AES-CTR encryptor
// keySize must be 16, 24, or 32 bytes (AES-128, AES-192, AES-256)
func NewAES_CTR_Encryptor(keySize int) *AES_CTR_Encryptor {
	return &AES_CTR_Encryptor{keySize: keySize}
}

// Encrypt encrypts plaintext using AES-CTR
func (e *AES_CTR_Encryptor) Encrypt(key []byte, plaintext []byte) (ciphertext []byte, err error) {
	if len(key) != e.keySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", e.keySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext = make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-CTR
func (e *AES_CTR_Encryptor) Decrypt(key []byte, ciphertext []byte) (plaintext []byte, err error) {
	if len(key) != e.keySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", e.keySize, len(key))
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCTR(block, iv)
	plaintext = make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

// Algorithm returns the algorithm name
func (e *AES_CTR_Encryptor) Algorithm() string {
	return string(AlgorithmAES_CTR)
}

// KeySize returns the key size in bytes
func (e *AES_CTR_Encryptor) KeySize() int {
	return e.keySize
}

// AES_GCM_Encryptor implements Encryptor using AES in GCM mode
// GCM mode provides both confidentiality and authenticity
type AES_GCM_Encryptor struct {
	keySize int
}

// NewAES_GCM_Encryptor creates a new AES-GCM encryptor
// keySize must be 16, 24, or 32 bytes (AES-128, AES-192, AES-256)
func NewAES_GCM_Encryptor(keySize int) *AES_GCM_Encryptor {
	return &AES_GCM_Encryptor{keySize: keySize}
}

// Encrypt encrypts plaintext using AES-GCM
func (e *AES_GCM_Encryptor) Encrypt(key []byte, plaintext []byte) (ciphertext []byte, err error) {
	if len(key) != e.keySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", e.keySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext = gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-GCM
func (e *AES_GCM_Encryptor) Decrypt(key []byte, ciphertext []byte) (plaintext []byte, err error) {
	if len(key) != e.keySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", e.keySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	plaintext, err = gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// Algorithm returns the algorithm name
func (e *AES_GCM_Encryptor) Algorithm() string {
	return string(AlgorithmAES_GCM)
}

// KeySize returns the key size in bytes
func (e *AES_GCM_Encryptor) KeySize() int {
	return e.keySize
}

// ChaCha20_Poly1305_Encryptor implements Encryptor using ChaCha20-Poly1305
// ChaCha20-Poly1305 is a modern AEAD cipher with excellent performance on platforms without AES-NI
type ChaCha20_Poly1305_Encryptor struct {
	keySize int
}

// NewChaCha20_Poly1305_Encryptor creates a new ChaCha20-Poly1305 encryptor
// ChaCha20-Poly1305 uses a fixed key size of 32 bytes
func NewChaCha20_Poly1305_Encryptor() *ChaCha20_Poly1305_Encryptor {
	return &ChaCha20_Poly1305_Encryptor{keySize: 32}
}

// Encrypt encrypts plaintext using ChaCha20-Poly1305
func (e *ChaCha20_Poly1305_Encryptor) Encrypt(key []byte, plaintext []byte) (ciphertext []byte, err error) {
	if len(key) != e.keySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", e.keySize, len(key))
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext = aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext using ChaCha20-Poly1305
func (e *ChaCha20_Poly1305_Encryptor) Decrypt(key []byte, ciphertext []byte) (plaintext []byte, err error) {
	if len(key) != e.keySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", e.keySize, len(key))
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	plaintext, err = aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// Algorithm returns the algorithm name
func (e *ChaCha20_Poly1305_Encryptor) Algorithm() string {
	return string(AlgorithmChaCha20_Poly1305)
}

// KeySize returns the key size in bytes
func (e *ChaCha20_Poly1305_Encryptor) KeySize() int {
	return e.keySize
}

// EncryptStr encrypts a string using the specified encryptor and returns base64-encoded result
func EncryptStr(enc Encryptor, key string, plaintext string) (ciphertext string, err error) {
	ciphertextBytes, err := enc.Encrypt([]byte(key), []byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(ciphertextBytes), nil
}

// DecryptStr decrypts a base64-encoded string using the specified encryptor
func DecryptStr(enc Encryptor, ciphertext string, key string) (plaintext string, err error) {
	ciphertextBytes, err := base64.RawStdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	plaintextBytes, err := enc.Decrypt([]byte(key), ciphertextBytes)
	if err != nil {
		return "", err
	}
	return string(plaintextBytes), nil
}

// NewEncryptor creates a new encryptor based on the algorithm name
func NewEncryptor(algorithm EncryptionAlgorithm, keySize int) (Encryptor, error) {
	switch algorithm {
	case AlgorithmAES_CTR:
		if keySize != 16 && keySize != 24 && keySize != 32 {
			return nil, fmt.Errorf("invalid AES key size: %d", keySize)
		}
		return NewAES_CTR_Encryptor(keySize), nil
	case AlgorithmAES_GCM:
		if keySize != 16 && keySize != 24 && keySize != 32 {
			return nil, fmt.Errorf("invalid AES key size: %d", keySize)
		}
		return NewAES_GCM_Encryptor(keySize), nil
	case AlgorithmChaCha20_Poly1305:
		return NewChaCha20_Poly1305_Encryptor(), nil
	default:
		return nil, fmt.Errorf("unknown encryption algorithm: %s", algorithm)
	}
}
