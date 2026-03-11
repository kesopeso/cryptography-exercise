package cryptography

import (
	"encoding/json"
	"fmt"
	"time"

	jose "github.com/go-jose/go-jose/v4"
)

type statusField struct {
	EncodedList string `json:"encodedList"`
	Index       int    `json:"index"`
}

type statusClaims struct {
	Iat    int64       `json:"iat"`
	Exp    int64       `json:"exp"`
	Iss    string      `json:"iss"`
	Status statusField `json:"status"`
}

func SignStatusJWS(keyPath string, issuer string, encodedList string, index int) (string, error) {
	now := time.Now().Unix()

	payload := statusClaims{
		Iat: now,
		Exp: now + 24*60*60,
		Iss: issuer,
		Status: statusField{
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
