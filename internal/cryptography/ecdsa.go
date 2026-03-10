package cryptography

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

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
		Type:  "EC PRIVATE KEY",
		Bytes: der,
	}

	err = pem.Encode(file, pemBlock)
	if err != nil {
		return fmt.Errorf("failed to write key to file: %w", err)
	}

	return nil
}
