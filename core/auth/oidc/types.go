package oidc

import (
	"time"

	"golang.org/x/oauth2"
)

type Tokens struct {
	OAuth2 *oauth2.Token
	RawID  string
}

type IDTokenClaims struct {
	Issuer        string   `json:"iss"`
	Subject       string   `json:"sub"`
	Audience      []string `json:"aud"`
	Expiry        int64    `json:"exp"`
	IssuedAt      int64    `json:"iat"`
	Nonce         string   `json:"nonce"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Name          string   `json:"name"`
	GivenName     string   `json:"given_name"`
	FamilyName    string   `json:"family_name"`
	Picture       string   `json:"picture"`
	Locale        string   `json:"locale"`
}

func (c IDTokenClaims) ExpiryTime() time.Time {
	if c.Expiry == 0 {
		return time.Time{}
	}
	return time.Unix(c.Expiry, 0)
}
