package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/kesopeso/cryptography-exercise/internal/bitset"
	"github.com/kesopeso/cryptography-exercise/internal/cryptography"
)

func TestGetJWSTokenAndReturnStatusValue(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "jws_test_key.pem")
	if err := cryptography.GenerateAndSaveECDSAKey(keyPath); err != nil {
		t.Fatalf("failed generating/saving key: %v", err)
	}

	var jws string
	issuerPath := "/api/status/01961234-5678-7abc-8def-0123456789ab"
	apiPath := issuerPath + "/1"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != apiPath {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		response := struct {
			JWS string `json:"jws"`
		}{
			JWS: jws,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("failed to encode json: %v", err)
		}
	}))
	defer server.Close()

	values := []bool{true, false}
	for _, value := range values {
		bs := bitset.NewBitset()
		idx := bs.Add(value)

		encodedBs, err := bs.Encode()
		if err != nil {
			t.Fatalf("failed encoding bitset: %v", err)
		}

		jws, err = cryptography.SignStatusJWS(keyPath, server.URL+issuerPath, encodedBs, idx)
		if err != nil {
			t.Fatalf("failed signing status JWS: %v", err)
		}

		resultValue, err := GetJWSTokenAndReturnStatusValue(http.DefaultClient, server.URL+apiPath)
		if err != nil {
			t.Fatalf("function failed: %v", err)
		}

		if resultValue != value {
			t.Errorf("result value missmatch, got: %t, want: %t", resultValue, value)
		}
	}
}
