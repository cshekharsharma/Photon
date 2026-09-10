package oidc

import "github.com/cshekharsharma/photon/core/session"

// Endpoints overrides discovered endpoints when provided.
type Endpoints struct {
	AuthURL   string
	TokenURL  string
	LogoutURL string
	JWKsURL   string
}

// Config controls OIDC client initialization.
type Config struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	Endpoints    *Endpoints
	Session      *session.Manager
}
