package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestVerifier_VerifyToken(t *testing.T) {
	feed := JwtFeed{
		Subject:    "user",
		Audience:   "aud",
		Issuer:     "iss",
		IssuedAt:   time.Now().Unix(),
		Expiration: time.Now().Add(time.Hour).Unix(),
	}

	token, err := SignHMAC(feed, "secret", "kid-1", SigningMethodHS256)
	assert.NoError(t, err)

	v := Verifier{
		Config: VerifyConfig{
			HMACSecrets:      map[string]string{"kid-1": "secret"},
			ExpectedIssuer:   "iss",
			ExpectedAudience: "aud",
		},
	}

	parsed, err := v.VerifyToken(context.Background(), token)
	assert.NoError(t, err)
	assert.NotNil(t, parsed)

	_, err = v.VerifyToken(context.Background(), "")
	assert.Error(t, err)
}
