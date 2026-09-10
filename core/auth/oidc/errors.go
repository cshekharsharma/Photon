package oidc

import "errors"

var (
	ErrNilConfig          = errors.New("oidc config cannot be nil")
	ErrMissingIssuer      = errors.New("oidc issuer is required for discovery or verification")
	ErrMissingClientID    = errors.New("oidc client id is required")
	ErrMissingRedirectURL = errors.New("oidc redirect url is required")
	ErrMissingSession     = errors.New("oidc session manager is required")
	ErrMissingState       = errors.New("oidc state not found in session")
	ErrInvalidState       = errors.New("oidc state mismatch")
	ErrMissingCode        = errors.New("oidc authorization code is missing")
	ErrMissingNonce       = errors.New("oidc nonce not found in session")
	ErrMissingVerifier    = errors.New("oidc pkce verifier not found in session")
	ErrMissingIDToken     = errors.New("oidc id_token is missing")
)
