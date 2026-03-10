package cryptography_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/kesopeso/cryptography-exercise/internal/cryptography"
)

func TestGenerateAndSaveECDSAKey(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "test.pem")

	err := cryptography.GenerateAndSaveECDSAKey(keyPath)
	if err != nil {
		t.Fatalf("GenerateAndSaveECDSAKey() error = %v", err)
	}

	data, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("failed to read generated key file: %v", err)
	}

	block, rest := pem.Decode(data)
	if block == nil {
		t.Fatal("failed to decode PEM block")
	}
	if len(rest) != 0 {
		t.Error("unexpected trailing data after PEM block")
	}
	if block.Type != "EC PRIVATE KEY" {
		t.Errorf("PEM type = %q, want %q", block.Type, "EC PRIVATE KEY")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse EC private key: %v", err)
	}
	if key.Curve != elliptic.P256() {
		t.Errorf("curve = %v, want P-256", key.Curve.Params().Name)
	}
}

func TestGenerateAndSaveECDSAKey_FileAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "test.pem")

	if err := os.WriteFile(keyPath, []byte("existing"), 0600); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	err := cryptography.GenerateAndSaveECDSAKey(keyPath)
	if err == nil {
		t.Fatal("expected error when file exists, got nil")
	}

	want := "file already exists: " + keyPath
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestGenerateAndSaveECDSAKey_InvalidPath(t *testing.T) {
	err := cryptography.GenerateAndSaveECDSAKey("/nonexistent/dir/key.pem")
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}

func TestGenerateAndSaveECDSAKey_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "test.pem")

	if err := cryptography.GenerateAndSaveECDSAKey(keyPath); err != nil {
		t.Fatalf("GenerateAndSaveECDSAKey() error = %v", err)
	}

	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("failed to stat key file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}

func TestGenerateAndSaveECDSAKey_ValidSignature(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "test.pem")

	if err := cryptography.GenerateAndSaveECDSAKey(keyPath); err != nil {
		t.Fatalf("GenerateAndSaveECDSAKey() error = %v", err)
	}

	data, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("failed to read key file: %v", err)
	}

	block, _ := pem.Decode(data)
	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse key: %v", err)
	}

	msg := []byte("test message")
	sig, err := ecdsa.SignASN1(nil, key, msg)
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	if !ecdsa.VerifyASN1(&key.PublicKey, msg, sig) {
		t.Error("signature verification failed")
	}
}
