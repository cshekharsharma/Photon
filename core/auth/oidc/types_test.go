package oidc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestTokenExpiry(t *testing.T) {
	assert.True(t, TokenExpiry(nil).IsZero())

	exp := time.Now().Add(time.Hour)
	tok := &oauth2.Token{Expiry: exp}
	assert.Equal(t, exp, TokenExpiry(tok))
}

func TestIDTokenClaimsExpiryTime(t *testing.T) {
	var claims IDTokenClaims
	assert.True(t, claims.ExpiryTime().IsZero())

	exp := time.Now().Add(time.Hour)
	claims.Expiry = exp.Unix()
	assert.Equal(t, exp.Unix(), claims.ExpiryTime().Unix())
}
