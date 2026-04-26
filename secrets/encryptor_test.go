package secrets

import (
	"testing"
)

func TestAES_CTR_Encryptor(t *testing.T) {
	key16 := "0123456789abcdef"                 // 16 bytes for AES-128
	key24 := "0123456789abcdef01234567"         // 24 bytes for AES-192
	key32 := "0123456789abcdef0123456789abcdef" // 32 bytes for AES-256

	tests := []struct {
		name      string
		encryptor *AES_CTR_Encryptor
		key       string
		plaintext string
		wantErr   bool
	}{
		{
			name:      "AES-128-CTR",
			encryptor: NewAES_CTR_Encryptor(16),
			key:       key16,
			plaintext: "Hello, World!",
			wantErr:   false,
		},
		{
			name:      "AES-192-CTR",
			encryptor: NewAES_CTR_Encryptor(24),
			key:       key24,
			plaintext: "This is a test message",
			wantErr:   false,
		},
		{
			name:      "AES-256-CTR",
			encryptor: NewAES_CTR_Encryptor(32),
			key:       key32,
			plaintext: "Longer test message with more content",
			wantErr:   false,
		},
		{
			name:      "invalid-key-size",
			encryptor: NewAES_CTR_Encryptor(16),
			key:       "short", // 5 bytes
			plaintext: "test",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := tt.encryptor.Encrypt([]byte(tt.key), []byte(tt.plaintext))
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				plaintext, err := tt.encryptor.Decrypt([]byte(tt.key), ciphertext)
				if err != nil {
					t.Errorf("Decrypt() failed: %v", err)
					return
				}

				if string(plaintext) != tt.plaintext {
					t.Errorf("Decrypt() = %q, want %q", string(plaintext), tt.plaintext)
				}
			}
		})
	}
}

func TestAES_GCM_Encryptor(t *testing.T) {
	key16 := "0123456789abcdef"
	key24 := "0123456789abcdef01234567"
	key32 := "0123456789abcdef0123456789abcdef"

	tests := []struct {
		name      string
		encryptor *AES_GCM_Encryptor
		key       string
		plaintext string
		wantErr   bool
	}{
		{
			name:      "AES-128-GCM",
			encryptor: NewAES_GCM_Encryptor(16),
			key:       key16,
			plaintext: "Hello, World!",
			wantErr:   false,
		},
		{
			name:      "AES-192-GCM",
			encryptor: NewAES_GCM_Encryptor(24),
			key:       key24,
			plaintext: "This is a test message",
			wantErr:   false,
		},
		{
			name:      "AES-256-GCM",
			encryptor: NewAES_GCM_Encryptor(32),
			key:       key32,
			plaintext: "Longer test message with more content",
			wantErr:   false,
		},
		{
			name:      "invalid-key-size",
			encryptor: NewAES_GCM_Encryptor(16),
			key:       "short",
			plaintext: "test",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := tt.encryptor.Encrypt([]byte(tt.key), []byte(tt.plaintext))
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				plaintext, err := tt.encryptor.Decrypt([]byte(tt.key), ciphertext)
				if err != nil {
					t.Errorf("Decrypt() failed: %v", err)
					return
				}

				if string(plaintext) != tt.plaintext {
					t.Errorf("Decrypt() = %q, want %q", string(plaintext), tt.plaintext)
				}

				// Test authentication: modifying ciphertext should fail
				if len(ciphertext) > 16 {
					modifiedCiphertext := make([]byte, len(ciphertext))
					copy(modifiedCiphertext, ciphertext)
					modifiedCiphertext[16] ^= 0xFF // Flip some bits
					_, err := tt.encryptor.Decrypt([]byte(tt.key), modifiedCiphertext)
					if err == nil {
						t.Error("Expected authentication failure when ciphertext is modified")
					}
				}
			}
		})
	}
}

func TestChaCha20_Poly1305_Encryptor(t *testing.T) {
	key32 := "0123456789abcdef0123456789abcdef" // 32 bytes

	tests := []struct {
		name      string
		key       string
		plaintext string
		wantErr   bool
	}{
		{
			name:      "basic-encryption",
			key:       key32,
			plaintext: "Hello, ChaCha20!",
			wantErr:   false,
		},
		{
			name:      "empty-plaintext",
			key:       key32,
			plaintext: "",
			wantErr:   false,
		},
		{
			name:      "long-plaintext",
			key:       key32,
			plaintext: "This is a much longer test message to verify that ChaCha20-Poly1305 works correctly with larger payloads",
			wantErr:   false,
		},
		{
			name:      "invalid-key-size",
			key:       "short",
			plaintext: "test",
			wantErr:   true,
		},
	}

	encryptor := NewChaCha20_Poly1305_Encryptor()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := encryptor.Encrypt([]byte(tt.key), []byte(tt.plaintext))
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				plaintext, err := encryptor.Decrypt([]byte(tt.key), ciphertext)
				if err != nil {
					t.Errorf("Decrypt() failed: %v", err)
					return
				}

				if string(plaintext) != tt.plaintext {
					t.Errorf("Decrypt() = %q, want %q", string(plaintext), tt.plaintext)
				}

				// Test authentication
				if len(ciphertext) > 20 {
					modifiedCiphertext := make([]byte, len(ciphertext))
					copy(modifiedCiphertext, ciphertext)
					modifiedCiphertext[20] ^= 0xFF
					_, err := encryptor.Decrypt([]byte(tt.key), modifiedCiphertext)
					if err == nil {
						t.Error("Expected authentication failure")
					}
				}
			}
		})
	}
}

func TestEncryptorFactory(t *testing.T) {
	tests := []struct {
		name      string
		algorithm EncryptionAlgorithm
		keySize   int
		wantErr   bool
	}{
		{name: "AES-CTR-128", algorithm: AlgorithmAES_CTR, keySize: 16, wantErr: false},
		{name: "AES-CTR-192", algorithm: AlgorithmAES_CTR, keySize: 24, wantErr: false},
		{name: "AES-CTR-256", algorithm: AlgorithmAES_CTR, keySize: 32, wantErr: false},
		{name: "AES-CTR-invalid", algorithm: AlgorithmAES_CTR, keySize: 12, wantErr: true},
		{name: "AES-GCM-128", algorithm: AlgorithmAES_GCM, keySize: 16, wantErr: false},
		{name: "AES-GCM-192", algorithm: AlgorithmAES_GCM, keySize: 24, wantErr: false},
		{name: "AES-GCM-256", algorithm: AlgorithmAES_GCM, keySize: 32, wantErr: false},
		{name: "ChaCha20", algorithm: AlgorithmChaCha20_Poly1305, keySize: 32, wantErr: false},
		{name: "ChaCha20-wrong-key", algorithm: AlgorithmChaCha20_Poly1305, keySize: 16, wantErr: false}, // Factory doesn't validate ChaCha20 key size
		{name: "unknown-algorithm", algorithm: "UNKNOWN", keySize: 16, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := NewEncryptor(tt.algorithm, tt.keySize)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEncryptor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && enc == nil {
				t.Error("NewEncryptor() returned nil for non-error case")
			}
		})
	}
}

func TestEncryptor_Interface_Properties(t *testing.T) {
	tests := []struct {
		name              string
		encryptor         Encryptor
		expectedAlgorithm string
		expectedKeySize   int
	}{
		{"AES-CTR-128", NewAES_CTR_Encryptor(16), string(AlgorithmAES_CTR), 16},
		{"AES-CTR-256", NewAES_CTR_Encryptor(32), string(AlgorithmAES_CTR), 32},
		{"AES-GCM-128", NewAES_GCM_Encryptor(16), string(AlgorithmAES_GCM), 16},
		{"AES-GCM-256", NewAES_GCM_Encryptor(32), string(AlgorithmAES_GCM), 32},
		{"ChaCha20", NewChaCha20_Poly1305_Encryptor(), string(AlgorithmChaCha20_Poly1305), 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.encryptor.Algorithm() != tt.expectedAlgorithm {
				t.Errorf("Algorithm() = %q, want %q", tt.encryptor.Algorithm(), tt.expectedAlgorithm)
			}

			if tt.encryptor.KeySize() != tt.expectedKeySize {
				t.Errorf("KeySize() = %d, want %d", tt.encryptor.KeySize(), tt.expectedKeySize)
			}
		})
	}
}

func TestEncryptStr_DecryptStr(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	plaintext := "Secret message"

	encryptors := []Encryptor{
		NewAES_CTR_Encryptor(32),
		NewAES_GCM_Encryptor(32),
		NewChaCha20_Poly1305_Encryptor(),
	}

	for _, enc := range encryptors {
		t.Run(enc.Algorithm(), func(t *testing.T) {
			ciphertext, err := EncryptStr(enc, key, plaintext)
			if err != nil {
				t.Fatalf("EncryptStr() failed: %v", err)
			}

			decrypted, err := DecryptStr(enc, ciphertext, key)
			if err != nil {
				t.Fatalf("DecryptStr() failed: %v", err)
			}

			if decrypted != plaintext {
				t.Errorf("DecryptStr() = %q, want %q", decrypted, plaintext)
			}
		})
	}
}

func TestEncryptor_IV_Randomization(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	plaintext := []byte("Test message for IV randomization")

	encryptors := []Encryptor{
		NewAES_CTR_Encryptor(32),
		NewAES_GCM_Encryptor(32),
		NewChaCha20_Poly1305_Encryptor(),
	}

	for _, enc := range encryptors {
		t.Run(enc.Algorithm(), func(t *testing.T) {
			cipher1, _ := enc.Encrypt([]byte(key), plaintext)
			cipher2, _ := enc.Encrypt([]byte(key), plaintext)
			cipher3, _ := enc.Encrypt([]byte(key), plaintext)

			// All should be different due to random nonce/IV
			if len(cipher1) == 0 || len(cipher2) == 0 || len(cipher3) == 0 {
				t.Fatal("Empty ciphertext returned")
			}

			if string(cipher1) == string(cipher2) {
				t.Errorf("Multiple encryptions produced same result (non-unique IVs)")
			}

			if string(cipher1) == string(cipher3) {
				t.Errorf("Multiple encryptions produced same result (non-unique IVs)")
			}

			if string(cipher2) == string(cipher3) {
				t.Errorf("Multiple encryptions produced same result (non-unique IVs)")
			}

			// But all should decrypt to the same plaintext
			plain1, _ := enc.Decrypt([]byte(key), cipher1)
			plain2, _ := enc.Decrypt([]byte(key), cipher2)
			plain3, _ := enc.Decrypt([]byte(key), cipher3)

			if string(plain1) != string(plaintext) || string(plain2) != string(plaintext) || string(plain3) != string(plaintext) {
				t.Error("Decryption did not produce original plaintext")
			}
		})
	}
}
