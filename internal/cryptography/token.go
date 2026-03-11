package cryptography

import (
	"encoding/json"
	"fmt"
	"time"

	jose "github.com/go-jose/go-jose/v4"
)

type StatusField struct {
	EncodedList string `json:"encodedList"`
	Index       int    `json:"index"`
}

type StatusClaims struct {
	Iat    int64       `json:"iat"`
	Exp    int64       `json:"exp"`
	Iss    string      `json:"iss"`
	Status StatusField `json:"status"`
}

// SignStatusJWS signs a status payload as a JWS compact serialization using the
// ECDSA private key at keyPath. The token includes iat (now), exp (now + 24h),
// iss (issuer), and a status field containing the encodedList and index. The
// public key is embedded in the JWS header as a JWK so verifiers can validate
// the signature.
func SignStatusJWS(keyPath string, issuer string, encodedList string, index int) (string, error) {
	now := time.Now().Unix()

	payload := StatusClaims{
		Iat: now,
		Exp: now + 24*60*60,
		Iss: issuer,
		Status: StatusField{
			EncodedList: encodedList,
			Index:       index,
		},
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed marshaling payload: %w", err)
	}

	privateKey, err := readPrivateKey(keyPath)
	if err != nil {
		return "", fmt.Errorf("failed reading private key: %w", err)
	}

	publicJWK := jose.JSONWebKey{
		Key:       &privateKey.PublicKey,
		Algorithm: string(jose.ES256),
		Use:       "sig",
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: jose.ES256,
			Key:       privateKey,
		},
		(&jose.SignerOptions{}).
			WithType("JWT").
			WithHeader("jwk", publicJWK),
	)
	if err != nil {
		return "", fmt.Errorf("failed create signer: %w", err)
	}

	jws, err := signer.Sign(payloadJSON)
	if err != nil {
		return "", fmt.Errorf("failed sign payload: %w", err)
	}

	compact, err := jws.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed compact serialize: %w", err)
	}

	return compact, nil
}

// ParseAndGetClaimsFromStatusJWS parses a JWS compact serialization, extracts
// the embedded JWK from the header, verifies the signature, and returns the
// decoded StatusClaims.
// Returns an error if the token is malformed, the JWK is missing, or the signature is invalid.
func ParseAndGetClaimsFromStatusJWS(jws string) (StatusClaims, error) {
	parsedJWS, err := jose.ParseSigned(jws, []jose.SignatureAlgorithm{jose.ES256})
	if err != nil {
		return StatusClaims{}, fmt.Errorf("failed to parse JWS: %w", err)
	}

	jwk := parsedJWS.Signatures[0].Header.JSONWebKey
	if jwk == nil {
		return StatusClaims{}, fmt.Errorf("missing jwk header")
	}

	payload, err := parsedJWS.Verify(jwk)
	if err != nil {
		return StatusClaims{}, fmt.Errorf("invalid signature: %w", err)
	}

	var claims StatusClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return StatusClaims{}, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	return claims, nil
}
