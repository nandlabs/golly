package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

func AesEncryptStr(key, message string) (encrypted string, err error) {
	var encryptedBytes []byte
	encryptedBytes, err = AesEncrypt([]byte(key), []byte(message))
	if err == nil {
		encrypted = base64.RawStdEncoding.EncodeToString(encryptedBytes)
	}
	return
}

func AesEncrypt(key, message []byte) (encrypted []byte, err error) {
	var block cipher.Block
	block, err = aes.NewCipher(key)
	if err == nil {
		encrypted = make([]byte, aes.BlockSize+len(message))
		iv := encrypted[:aes.BlockSize]
		if _, err = io.ReadFull(rand.Reader, iv); err == nil {
			stream := cipher.NewCFBEncrypter(block, iv)
			stream.XORKeyStream(encrypted[aes.BlockSize:], message)
		}
	}
	return
}

func AesDecryptStr(encrypted, key string) (message string, err error) {
	var decryptedBytes, encryptedBytes []byte
	encryptedBytes, err = base64.RawStdEncoding.DecodeString(encrypted)
	if err != nil {
		return
	}
	decryptedBytes, err = AesDecrypt(encryptedBytes, []byte(key))
	if err == nil {
		message = string(decryptedBytes)
	}
	return
}

func AesDecrypt(encrypted, key []byte) (message []byte, err error) {
	var block cipher.Block
	block, err = aes.NewCipher(key)
	message = encrypted
	if err == nil {
		if len(encrypted) < aes.BlockSize {
			err = errors.New("encrypted block size is too short")
		} else {
			iv := message[:aes.BlockSize]
			message = message[aes.BlockSize:]
			stream := cipher.NewCFBDecrypter(block, iv)
			stream.XORKeyStream(message, message)
		}
	}

	return
}
