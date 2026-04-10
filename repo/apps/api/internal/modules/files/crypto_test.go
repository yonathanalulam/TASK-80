package files

import (
	"bytes"
	"testing"
)

func TestGenerateDEK(t *testing.T) {
	dek, err := GenerateDEK()
	if err != nil {
		t.Fatalf("GenerateDEK failed: %v", err)
	}
	if len(dek) != 32 {
		t.Errorf("DEK should be 32 bytes, got %d", len(dek))
	}

	// Two DEKs should be different (random)
	dek2, _ := GenerateDEK()
	if bytes.Equal(dek, dek2) {
		t.Error("two generated DEKs should not be identical")
	}
}

func TestEncryptDecryptFile(t *testing.T) {
	dek, err := GenerateDEK()
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("This is sensitive contract data that must be encrypted at rest.")

	ciphertext, err := EncryptFile(plaintext, dek)
	if err != nil {
		t.Fatalf("EncryptFile failed: %v", err)
	}

	// Ciphertext should not equal plaintext
	if bytes.Equal(ciphertext, plaintext) {
		t.Error("ciphertext should not equal plaintext")
	}

	// Ciphertext should be longer than plaintext (nonce + tag overhead)
	if len(ciphertext) <= len(plaintext) {
		t.Error("ciphertext should be longer than plaintext due to nonce and auth tag")
	}

	// Decrypt should recover original plaintext
	recovered, err := DecryptFile(ciphertext, dek)
	if err != nil {
		t.Fatalf("DecryptFile failed: %v", err)
	}

	if !bytes.Equal(recovered, plaintext) {
		t.Errorf("decrypted data does not match original.\ngot:  %q\nwant: %q", recovered, plaintext)
	}
}

func TestDecryptFile_WrongKey(t *testing.T) {
	dek1, _ := GenerateDEK()
	dek2, _ := GenerateDEK()

	plaintext := []byte("secret data")
	ciphertext, _ := EncryptFile(plaintext, dek1)

	// Decrypting with wrong key should fail
	_, err := DecryptFile(ciphertext, dek2)
	if err == nil {
		t.Error("decryption with wrong key should fail")
	}
}

func TestDecryptFile_TamperedCiphertext(t *testing.T) {
	dek, _ := GenerateDEK()
	plaintext := []byte("integrity check data")
	ciphertext, _ := EncryptFile(plaintext, dek)

	// Tamper with ciphertext
	if len(ciphertext) > 20 {
		ciphertext[20] ^= 0xFF
	}

	_, err := DecryptFile(ciphertext, dek)
	if err == nil {
		t.Error("decryption of tampered ciphertext should fail")
	}
}

func TestWrapUnwrapDEK(t *testing.T) {
	masterKey, _ := GenerateDEK() // 32 bytes
	dek, _ := GenerateDEK()

	wrapped, err := WrapDEK(dek, masterKey)
	if err != nil {
		t.Fatalf("WrapDEK failed: %v", err)
	}
	if wrapped == "" {
		t.Fatal("wrapped DEK should not be empty")
	}

	unwrapped, err := UnwrapDEK(wrapped, masterKey)
	if err != nil {
		t.Fatalf("UnwrapDEK failed: %v", err)
	}

	if !bytes.Equal(unwrapped, dek) {
		t.Error("unwrapped DEK does not match original")
	}
}

func TestUnwrapDEK_WrongMasterKey(t *testing.T) {
	masterKey1, _ := GenerateDEK()
	masterKey2, _ := GenerateDEK()
	dek, _ := GenerateDEK()

	wrapped, _ := WrapDEK(dek, masterKey1)
	_, err := UnwrapDEK(wrapped, masterKey2)
	if err == nil {
		t.Error("unwrap with wrong master key should fail")
	}
}

func TestEncryptFile_EmptyData(t *testing.T) {
	dek, _ := GenerateDEK()
	ciphertext, err := EncryptFile([]byte{}, dek)
	if err != nil {
		t.Fatalf("encrypting empty data should not fail: %v", err)
	}

	recovered, err := DecryptFile(ciphertext, dek)
	if err != nil {
		t.Fatalf("decrypting empty data should not fail: %v", err)
	}
	if len(recovered) != 0 {
		t.Errorf("recovered data should be empty, got %d bytes", len(recovered))
	}
}
