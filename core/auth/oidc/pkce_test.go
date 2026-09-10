package oidc

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateVerifier(t *testing.T) {
	orig := randReader
	defer func() { randReader = orig }()

	v, err := generateVerifier()
	require.NoError(t, err)
	assert.NotEmpty(t, v)

	decoded, err := base64.RawURLEncoding.DecodeString(v)
	require.NoError(t, err)
	assert.Len(t, decoded, pkceVerifierLen)

	randReader = failingReader{}
	_, err = generateVerifier()
	assert.Error(t, err)
}

func TestChallengeFromVerifier(t *testing.T) {
	ch1 := challengeFromVerifier("verifier-1")
	ch2 := challengeFromVerifier("verifier-1")
	ch3 := challengeFromVerifier("verifier-2")

	assert.NotEmpty(t, ch1)
	assert.Equal(t, ch1, ch2)
	assert.NotEqual(t, ch1, ch3)
}

type failingReader struct{}

func (failingReader) Read(_ []byte) (int, error) {
	return 0, errors.New("boom")
}
