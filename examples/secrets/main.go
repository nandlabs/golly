// Package main demonstrates the secrets package for encryption and secret management.
package main

import (
	"fmt"
	"log"

	"oss.nandlabs.io/golly/secrets"
)

func main() {
	// --- AES Encryption / Decryption ---
	// Key must be 16, 24, or 32 bytes for AES-128, AES-192, AES-256
	key := "this-is-a-32-byte-key-for-aes!!" // 32 bytes = AES-256
	message := "Hello, this is a secret message!"

	// Encrypt
	encrypted, err := secrets.AesEncryptStr(key, message)
	if err != nil {
		log.Fatal("Encrypt error:", err)
	}
	fmt.Println("Original:", message)
	fmt.Println("Encrypted:", encrypted)

	// Decrypt
	decrypted, err := secrets.AesDecryptStr(encrypted, key)
	if err != nil {
		log.Fatal("Decrypt error:", err)
	}
	fmt.Println("Decrypted:", decrypted)

	// Byte-level encryption
	data := []byte("binary secret data")
	encBytes, err := secrets.AesEncrypt([]byte(key), data)
	if err != nil {
		log.Fatal(err)
	}
	decBytes, err := secrets.AesDecrypt(encBytes, []byte(key))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Byte decrypt:", string(decBytes))
}
