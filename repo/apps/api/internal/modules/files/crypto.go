package files

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

func GenerateDEK() ([]byte, error) {
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, fmt.Errorf("generate DEK: %w", err)
	}
	return dek, nil
}

func EncryptFile(plaintext, dek []byte) ([]byte, error) {
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func DecryptFile(ciphertext, dek []byte) ([]byte, error) {
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, data := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plaintext, nil
}

func WrapDEK(dek, masterKey []byte) (string, error) {
	wrapped, err := EncryptFile(dek, masterKey)
	if err != nil {
		return "", fmt.Errorf("wrap DEK: %w", err)
	}
	return hex.EncodeToString(wrapped), nil
}

func UnwrapDEK(wrapped string, masterKey []byte) ([]byte, error) {
	data, err := hex.DecodeString(wrapped)
	if err != nil {
		return nil, fmt.Errorf("decode wrapped DEK: %w", err)
	}
	dek, err := DecryptFile(data, masterKey)
	if err != nil {
		return nil, fmt.Errorf("unwrap DEK: %w", err)
	}
	return dek, nil
}
