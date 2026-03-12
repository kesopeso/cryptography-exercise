package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/kesopeso/cryptography-exercise/internal/bitset"
	"github.com/kesopeso/cryptography-exercise/pkg/cryptography"
)

// GetJWSTokenAndReturnStatusValue calls url defined in getJWSTokenURL parameter,
// parses the JWS token from the {"jws": "..."} JSON response and validates the claims.
// It then decodes the claims status and returns the boolean value at the given index.
func GetJWSTokenAndReturnStatusValue(httpClient *http.Client, getJWSTokenURL string) (bool, error) {
	resp, err := httpClient.Get(getJWSTokenURL)
	if err != nil {
		return false, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var body struct {
		JWS string `json:"jws"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	claims, err := cryptography.ParseAndGetClaimsFromStatusJWS(body.JWS)

	now := time.Now().Unix()
	if claims.Iat > now {
		return false, fmt.Errorf("token issued in the future")
	}

	if claims.Exp <= now {
		return false, fmt.Errorf("token expired")
	}

	u, err := url.Parse(getJWSTokenURL)
	if err != nil {
		return false, fmt.Errorf("failed to parse URL: %w", err)
	}
	u.Path = path.Dir(u.Path)
	if claims.Iss != u.String() {
		return false, fmt.Errorf("invalid issuer: got %q, want %q", claims.Iss, u.String())
	}

	bs, err := bitset.Decode(claims.Status.EncodedList)
	if err != nil {
		return false, fmt.Errorf("failed to decode status list: %w", err)
	}

	value, err := bs.Get(claims.Status.Index)
	if err != nil {
		return false, fmt.Errorf("failed to get status value: %w", err)
	}

	return value, nil
}
