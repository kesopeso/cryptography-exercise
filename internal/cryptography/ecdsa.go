package cryptography

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
)

// pemKeyType is the expected header for SEC 1 ECDSA private keys.
const PemKeyType = "EC PRIVATE KEY"

// GenerateAndSaveECDSAKey generates a new ECDSA private key using the P-256 (secp256r1) curve
// and saves it to the specified filepath in PEM-encoded format.
//
// The file is created with 0600 permissions (read/write for owner only).
// To prevent accidental data loss, the function returns an error if the file already exists.
func GenerateAndSaveECDSAKey(filepath string) error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate ecdsa key: %w", err)
	}

	der, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("file already exists: %s", filepath)
		}
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer file.Close()

	pemBlock := &pem.Block{
		Type:  PemKeyType,
		Bytes: der,
	}

	err = pem.Encode(file, pemBlock)
	if err != nil {
		return fmt.Errorf("failed to write key to file: %w", err)
	}

	return nil
}

// SignMessage reads an ECDSA private key from a PEM file and signs the given message.
// The message is hashed with SHA-256 before signing.
// The signature is returned as a Base64URL-encoded string.
func SignMessage(keyPath string, message string) (string, error) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}
	if block.Type != PemKeyType {
		return "", fmt.Errorf("invalid key type: %s", block.Type)
	}

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse EC private key: %w", err)
	}

	hash := sha256.Sum256([]byte(message))

	sig, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(sig), nil
}
