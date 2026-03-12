package cryptography

import (
	"bytes"
	"testing"
)

func TestAESEncryptDecrypt_RoundTrip(t *testing.T) {
	plaintext := "hello world"
	password := "secret-password"

	encrypted, err := AESEncrypt(plaintext, password)
	if err != nil {
		t.Fatalf("AESEncrypt() error = %v", err)
	}

	decrypted, err := AESDecrypt(encrypted, password)
	if err != nil {
		t.Fatalf("AESDecrypt() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("got %q, want %q", decrypted, plaintext)
	}
}

func TestAESEncrypt_ProducesRandomCiphertext(t *testing.T) {
	plaintext := "hello world"
	password := "secret-password"

	enc1, err := AESEncrypt(plaintext, password)
	if err != nil {
		t.Fatalf("AESEncrypt() error = %v", err)
	}

	enc2, err := AESEncrypt(plaintext, password)
	if err != nil {
		t.Fatalf("AESEncrypt() error = %v", err)
	}

	if bytes.Equal(enc1, enc2) {
		t.Error("two encryptions of the same plaintext produced identical ciphertext")
	}
}

func TestAESDecrypt_WrongPassword(t *testing.T) {
	encrypted, err := AESEncrypt("hello world", "correct-password")
	if err != nil {
		t.Fatalf("AESEncrypt() error = %v", err)
	}

	_, err = AESDecrypt(encrypted, "wrong-password")
	if err == nil {
		t.Fatal("expected error when decrypting with wrong password, got nil")
	}
}

func TestAESDecrypt_TamperedCiphertext(t *testing.T) {
	encrypted, err := AESEncrypt("hello world", "secret-password")
	if err != nil {
		t.Fatalf("AESEncrypt() error = %v", err)
	}

	encrypted[len(encrypted)-1] ^= 0xFF

	_, err = AESDecrypt(encrypted, "secret-password")
	if err == nil {
		t.Fatal("expected error for tampered ciphertext, got nil")
	}
}

func TestAESDecrypt_TooShort(t *testing.T) {
	_, err := AESDecrypt([]byte("short"), "password")
	if err == nil {
		t.Fatal("expected error for too-short ciphertext, got nil")
	}
}
