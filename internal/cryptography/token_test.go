package cryptography

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/go-jose/go-jose/v4"
)

func TestSignStatusJWS(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "jws-test-key.pem")
	if err := GenerateAndSaveECDSAKey(keyPath); err != nil {
		t.Fatalf("failed generating/saving key: %v", err)
	}

	issuer := "https://example.com"
	encodedList := "uY29udGVudA"
	index := 42

	jws, err := SignStatusJWS(keyPath, issuer, encodedList, index)
	if err != nil {
		t.Fatalf("SignStatusJWS failed: %v", err)
	}

	object, err := jose.ParseSigned(jws, []jose.SignatureAlgorithm{jose.ES256})
	if err != nil {
		t.Fatalf("failed to parse resulting JWS: %v", err)
	}

	publicKey := object.Signatures[0].Header.JSONWebKey.Key
	payload, err := object.Verify(publicKey)
	if err != nil {
		t.Fatalf("cryptographic verification failed: %v", err)
	}

	var claims statusClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatalf("failed to unmarshal verified payload: %v", err)
	}

	if claims.Iss != issuer {
		t.Errorf("expected issuer %s, got %s", issuer, claims.Iss)
	}

	if claims.Status.Index != index {
		t.Errorf("Expected index %d, got %d", index, claims.Status.Index)
	}

	if claims.Status.EncodedList != encodedList {
		t.Errorf("Expected list %s, got %s", encodedList, claims.Status.EncodedList)
	}

	if len(object.Signatures) == 0 || object.Signatures[0].Header.JSONWebKey == nil {
		t.Error("JWS header is missing the public JWK")
	}
}
