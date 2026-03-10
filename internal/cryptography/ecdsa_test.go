package cryptography_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
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
	if block.Type != cryptography.PemKeyType {
		t.Errorf("PEM type = %q, want %q", block.Type, cryptography.PemKeyType)
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

func generateTestKey(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "test.pem")
	if err := cryptography.GenerateAndSaveECDSAKey(keyPath); err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	return keyPath
}

func readPrivateKey(t *testing.T, keyPath string) *ecdsa.PrivateKey {
	t.Helper()
	data, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("failed to read key file: %v", err)
	}
	block, _ := pem.Decode(data)
	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse key: %v", err)
	}
	return key
}

func TestSignMessage(t *testing.T) {
	keyPath := generateTestKey(t)
	message := "hello world"

	sig, err := cryptography.SignMessage(keyPath, message)
	if err != nil {
		t.Fatalf("SignMessage() error = %v", err)
	}

	// Verify the signature is valid Base64URL
	sigBytes, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		t.Fatalf("signature is not valid Base64URL: %v", err)
	}

	// Verify the signature is cryptographically valid
	key := readPrivateKey(t, keyPath)
	hash := sha256.Sum256([]byte(message))
	if !ecdsa.VerifyASN1(&key.PublicKey, hash[:], sigBytes) {
		t.Error("signature verification failed")
	}
}

func TestSignMessage_DifferentMessagesProduceDifferentSignatures(t *testing.T) {
	keyPath := generateTestKey(t)

	sig1, err := cryptography.SignMessage(keyPath, "message one")
	if err != nil {
		t.Fatalf("SignMessage() error = %v", err)
	}

	sig2, err := cryptography.SignMessage(keyPath, "message two")
	if err != nil {
		t.Fatalf("SignMessage() error = %v", err)
	}

	if sig1 == sig2 {
		t.Error("different messages produced identical signatures")
	}
}

func TestSignMessage_NonexistentKeyFile(t *testing.T) {
	_, err := cryptography.SignMessage("/nonexistent/key.pem", "msg")
	if err == nil {
		t.Fatal("expected error for nonexistent key file, got nil")
	}
}

func TestSignMessage_InvalidPEMFile(t *testing.T) {
	dir := t.TempDir()
	badFile := filepath.Join(dir, "bad.pem")
	if err := os.WriteFile(badFile, []byte("not a pem file"), 0600); err != nil {
		t.Fatalf("failed to write bad file: %v", err)
	}

	_, err := cryptography.SignMessage(badFile, "msg")
	if err == nil {
		t.Fatal("expected error for invalid PEM, got nil")
	}
}

func TestSignMessage_InvalidKeyType(t *testing.T) {
	dir := t.TempDir()
	keyFile := filepath.Join(dir, "wrong_type.pem")

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: []byte("fake key data"),
	})
	if err := os.WriteFile(keyFile, pemData, 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := cryptography.SignMessage(keyFile, "msg")
	if err == nil {
		t.Fatal("expected error for invalid key type, got nil")
	}

	want := "invalid key type: RSA PRIVATE KEY"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestSignMessage_EmptyMessage(t *testing.T) {
	keyPath := generateTestKey(t)

	sig, err := cryptography.SignMessage(keyPath, "")
	if err != nil {
		t.Fatalf("SignMessage() error = %v", err)
	}

	sigBytes, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		t.Fatalf("signature is not valid Base64URL: %v", err)
	}

	key := readPrivateKey(t, keyPath)
	hash := sha256.Sum256([]byte{})
	if !ecdsa.VerifyASN1(&key.PublicKey, hash[:], sigBytes) {
		t.Error("signature verification failed for empty message")
	}
}
