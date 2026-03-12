package cryptography_test

import (
	"bytes"
	"testing"

	"github.com/kesopeso/cryptography-exercise/pkg/cryptography"
)

func TestAESEncryptDecrypt_RoundTrip(t *testing.T) {
	plaintext := "hello world"
	password := "secret-password"

	encrypted, err := cryptography.AESEncrypt(plaintext, password)
	if err != nil {
		t.Fatalf("AESEncrypt() error = %v", err)
	}

	decrypted, err := cryptography.AESDecrypt(encrypted, password)
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

	enc1, err := cryptography.AESEncrypt(plaintext, password)
	if err != nil {
		t.Fatalf("AESEncrypt() error = %v", err)
	}

	enc2, err := cryptography.AESEncrypt(plaintext, password)
	if err != nil {
		t.Fatalf("AESEncrypt() error = %v", err)
	}

	if bytes.Equal(enc1, enc2) {
		t.Error("two encryptions of the same plaintext produced identical ciphertext")
	}
}

func TestAESDecrypt_WrongPassword(t *testing.T) {
	encrypted, err := cryptography.AESEncrypt("hello world", "correct-password")
	if err != nil {
		t.Fatalf("AESEncrypt() error = %v", err)
	}

	_, err = cryptography.AESDecrypt(encrypted, "wrong-password")
	if err == nil {
		t.Fatal("expected error when decrypting with wrong password, got nil")
	}
}

func TestAESDecrypt_TamperedCiphertext(t *testing.T) {
	encrypted, err := cryptography.AESEncrypt("hello world", "secret-password")
	if err != nil {
		t.Fatalf("AESEncrypt() error = %v", err)
	}

	encrypted[len(encrypted)-1] ^= 0xFF

	_, err = cryptography.AESDecrypt(encrypted, "secret-password")
	if err == nil {
		t.Fatal("expected error for tampered ciphertext, got nil")
	}
}

func TestAESDecrypt_TooShort(t *testing.T) {
	_, err := cryptography.AESDecrypt([]byte("short"), "password")
	if err == nil {
		t.Fatal("expected error for too-short ciphertext, got nil")
	}
}
