package oidc

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

const (
	pkceVerifierLen = 32
)

var randReader = rand.Reader

func generateVerifier() (string, error) {
	b := make([]byte, pkceVerifierLen)
	if _, err := randReader.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func challengeFromVerifier(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
