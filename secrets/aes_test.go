package secrets

import (
	"testing"
)

func TestAes(t *testing.T) {
	type args struct {
		key     string
		message string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "simple-message-16bit",
			args: args{
				key:     "This12BitKey0001",
				message: "This is a simple message",
			},
			wantErr: false,
		}, {
			name: "simple-message-24bit",
			args: args{
				key:     "This24BitKeyWillBeUsed01",
				message: "This is a simple message",
			},
			wantErr: false,
		},
		{
			name: "simple-message-32bit",
			args: args{
				key:     "thisisa32bitkeythisisa32bitkey02",
				message: "This is a simple message",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		var err error
		var encryptedMsg string
		var decryptedMsg string

		t.Run(tt.name, func(t *testing.T) {
			encryptedMsg, err = AesEncryptStr(tt.args.key, tt.args.message)

			if (err != nil) != tt.wantErr {
				t.Errorf("AesEncryptStr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			decryptedMsg, err = AesDecryptStr(encryptedMsg, tt.args.key)

			if (err != nil) != tt.wantErr {
				t.Errorf("AesDecryptStr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if decryptedMsg != tt.args.message {
				t.Errorf("AesEncryptStr() gotEncrypted = %v, want %v", decryptedMsg, tt.args.message)
			}
		})
	}
}

// TestAesIVRandomization verifies that multiple encryptions of the same plaintext
// with the same key produce different ciphertexts due to random IV generation.
// This is critical for semantic security in CTR mode.
func TestAesIVRandomization(t *testing.T) {
	key := "This12BitKey0001"
	message := "This is a test message for IV randomization"

	// Encrypt the same message multiple times
	encrypted1, err := AesEncryptStr(key, message)
	if err != nil {
		t.Fatalf("First encryption failed: %v", err)
	}

	encrypted2, err := AesEncryptStr(key, message)
	if err != nil {
		t.Fatalf("Second encryption failed: %v", err)
	}

	encrypted3, err := AesEncryptStr(key, message)
	if err != nil {
		t.Fatalf("Third encryption failed: %v", err)
	}

	// All three ciphertexts should be different due to random IV
	if encrypted1 == encrypted2 {
		t.Error("Expected different ciphertexts for same plaintext (encrypted1 == encrypted2), but got same result")
	}

	if encrypted1 == encrypted3 {
		t.Error("Expected different ciphertexts for same plaintext (encrypted1 == encrypted3), but got same result")
	}

	if encrypted2 == encrypted3 {
		t.Error("Expected different ciphertexts for same plaintext (encrypted2 == encrypted3), but got same result")
	}

	// But all should decrypt to the same original message
	decrypted1, _ := AesDecryptStr(encrypted1, key)
	decrypted2, _ := AesDecryptStr(encrypted2, key)
	decrypted3, _ := AesDecryptStr(encrypted3, key)

	if decrypted1 != message {
		t.Errorf("Decryption 1 failed: got %q, want %q", decrypted1, message)
	}
	if decrypted2 != message {
		t.Errorf("Decryption 2 failed: got %q, want %q", decrypted2, message)
	}
	if decrypted3 != message {
		t.Errorf("Decryption 3 failed: got %q, want %q", decrypted3, message)
	}
}

// TestAesDecryptInvalidBlockSize tests that decryption properly handles
// blocks that are too short to contain a valid IV.
func TestAesDecryptInvalidBlockSize(t *testing.T) {
	key := "This12BitKey0001"

	// Create a ciphertext that's shorter than one AES block (16 bytes)
	tooShortCiphertext := []byte("short")

	_, err := AesDecrypt(tooShortCiphertext, []byte(key))
	if err == nil {
		t.Error("Expected an error for too-short ciphertext, but got nil")
	}

	if err.Error() != "encrypted block size is too short" {
		t.Errorf("Expected error message 'encrypted block size is too short', got %q", err.Error())
	}
}
