package cryptography

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
)

// deriveKey returns a 32-byte AES-256 key derived from the password via SHA-256.
func deriveKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

func getGCM(password string) (cipher.AEAD, error) {
	block, err := aes.NewCipher(deriveKey(password))
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return gcm, nil
}

// AESEncrypt encrypts plaintext using AES-256-GCM with a key derived from password.
// Returns the nonce prepended to the ciphertext as a byte slice.
func AESEncrypt(plaintext string, password string) ([]byte, error) {
	gcm, err := getGCM(password)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertext, nil
}

// AESDecrypt decrypts cypherBytes using AES-256-GCM with a key derived from password.
// cypherBytes must be in the format produced by AESEncrypt (nonce + ciphertext).
func AESDecrypt(cypherBytes []byte, password string) (string, error) {
	gcm, err := getGCM(password)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cypherBytes) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := cypherBytes[:nonceSize], cypherBytes[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}
