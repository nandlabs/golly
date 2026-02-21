# Secrets Package

The `secrets` package provides AES encryption/decryption utilities and a pluggable credential store for managing secrets in Go.

---

- [Installation](#installation)
- [Features](#features)
- [Usage](#usage)
  - [AES Encryption](#aes-encryption)
  - [Credential Store](#credential-store)

---

## Installation

```sh
go get oss.nandlabs.io/golly
```

## Features

- **AES Encryption/Decryption** for strings and byte slices
- **Credential Store** interface with a built-in local file-backed implementation
- **Manager** for registering and retrieving credential stores by provider name

## Usage

### AES Encryption

```go
import "oss.nandlabs.io/golly/secrets"

key := "my-secret-key-16" // 16, 24, or 32 bytes for AES-128/192/256

// Encrypt a string
encrypted, err := secrets.AesEncryptStr(key, "Hello, World!")

// Decrypt a string
decrypted, err := secrets.AesDecryptStr(encrypted, key)
fmt.Println(decrypted) // Hello, World!

// Encrypt/decrypt raw bytes
encBytes, _ := secrets.AesEncrypt([]byte(key), []byte("secret data"))
decBytes, _ := secrets.AesDecrypt(encBytes, []byte(key))
```

### Credential Store

```go
import "oss.nandlabs.io/golly/secrets"

// Create a local file-backed credential store
store, err := secrets.NewLocalStore("/path/to/secrets.json", "master-key-16bytes")

// Use the manager to register and retrieve stores
manager := secrets.GetManager()
```
