package oidc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const (
	sessionStateKey    = "oidc_state"
	sessionNonceKey    = "oidc_nonce"
	sessionVerifierKey = "oidc_verifier"
)

type Client struct {
	issuer    string
	sessions  sessionStore
	provider  *gooidc.Provider
	verifier  *gooidc.IDTokenVerifier
	oauth2Cfg oauth2.Config
	endpoints Endpoints
}

type sessionStore interface {
	Put(ctx context.Context, key string, val interface{}) error
	GetString(ctx context.Context, key string) (string, error)
	Remove(ctx context.Context, key string) error
}

func New(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}
	if cfg.ClientID == "" {
		return nil, ErrMissingClientID
	}
	if cfg.RedirectURL == "" {
		return nil, ErrMissingRedirectURL
	}

	sess := cfg.Session
	if sess == nil {
		return nil, ErrMissingSession
	}

	var provider *gooidc.Provider
	if cfg.Issuer == "" {
		return nil, ErrMissingIssuer
	}

	var err error
	provider, err = gooidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, err
	}

	endpoints := Endpoints{}
	if cfg.Endpoints != nil {
		endpoints = *cfg.Endpoints
	}

	if endpoints.AuthURL == "" || endpoints.TokenURL == "" {
		auth := provider.Endpoint()
		if endpoints.AuthURL == "" {
			endpoints.AuthURL = auth.AuthURL
		}
		if endpoints.TokenURL == "" {
			endpoints.TokenURL = auth.TokenURL
		}
	}
	if endpoints.AuthURL == "" || endpoints.TokenURL == "" {
		return nil, fmt.Errorf("oidc endpoints are required")
	}

	oauth2Cfg := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       normalizeScopes(cfg.Scopes),
		Endpoint: oauth2.Endpoint{
			AuthURL:  endpoints.AuthURL,
			TokenURL: endpoints.TokenURL,
		},
	}

	verifier := provider.Verifier(&gooidc.Config{ClientID: cfg.ClientID})

	return &Client{
		issuer:    cfg.Issuer,
		sessions:  sess,
		provider:  provider,
		verifier:  verifier,
		oauth2Cfg: oauth2Cfg,
		endpoints: endpoints,
	}, nil
}

func normalizeScopes(scopes []string) []string {
	out := []string{"openid"}
	for _, s := range scopes {
		if s == "" || s == "openid" {
			continue
		}
		out = append(out, s)
	}
	return out
}

// AuthCodeURL builds the authorization URL and stores state/nonce/verifier in session.
func (c *Client) AuthCodeURL(ctx context.Context, opts ...oauth2.AuthCodeOption) (string, error) {
	if c == nil {
		return "", ErrNilConfig
	}

	state, err := generateVerifier()
	if err != nil {
		return "", err
	}
	nonce, err := generateVerifier()
	if err != nil {
		return "", err
	}
	verifier, err := generateVerifier()
	if err != nil {
		return "", err
	}

	if err := c.sessions.Put(ctx, sessionStateKey, state); err != nil {
		return "", err
	}
	if err := c.sessions.Put(ctx, sessionNonceKey, nonce); err != nil {
		return "", err
	}
	if err := c.sessions.Put(ctx, sessionVerifierKey, verifier); err != nil {
		return "", err
	}

	challenge := challengeFromVerifier(verifier)
	baseOpts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	}
	baseOpts = append(baseOpts, opts...)

	return c.oauth2Cfg.AuthCodeURL(state, baseOpts...), nil
}

// HandleCallback validates state, exchanges code, and verifies the ID token.
func (c *Client) HandleCallback(ctx context.Context, r *http.Request) (*Tokens, *IDTokenClaims, error) {
	if c == nil {
		return nil, nil, ErrNilConfig
	}
	if r == nil {
		return nil, nil, ErrMissingCode
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, nil, ErrMissingCode
	}
	state := r.URL.Query().Get("state")

	expectedState, err := c.sessions.GetString(ctx, sessionStateKey)
	if err != nil {
		return nil, nil, err
	}
	if expectedState == "" {
		return nil, nil, ErrMissingState
	}
	if state != expectedState {
		return nil, nil, ErrInvalidState
	}

	verifier, err := c.sessions.GetString(ctx, sessionVerifierKey)
	if err != nil {
		return nil, nil, err
	}
	if verifier == "" {
		return nil, nil, ErrMissingVerifier
	}

	token, err := c.oauth2Cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", verifier))
	if err != nil {
		return nil, nil, err
	}

	rawID, ok := token.Extra("id_token").(string)
	if !ok || rawID == "" {
		return nil, nil, ErrMissingIDToken
	}

	idToken, err := c.verifier.Verify(ctx, rawID)
	if err != nil {
		return nil, nil, err
	}

	nonce, err := c.sessions.GetString(ctx, sessionNonceKey)
	if err != nil {
		return nil, nil, err
	}
	if nonce == "" {
		return nil, nil, ErrMissingNonce
	}
	if idToken.Nonce != nonce {
		return nil, nil, ErrInvalidState
	}

	var claims IDTokenClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, nil, err
	}

	_ = c.sessions.Remove(ctx, sessionStateKey)
	_ = c.sessions.Remove(ctx, sessionNonceKey)
	_ = c.sessions.Remove(ctx, sessionVerifierKey)

	return &Tokens{OAuth2: token, RawID: rawID}, &claims, nil
}

// UserInfo fetches userinfo for the given token when supported.
func (c *Client) UserInfo(ctx context.Context, token *oauth2.Token) (*gooidc.UserInfo, error) {
	if c == nil || c.provider == nil {
		return nil, fmt.Errorf("provider not configured")
	}
	if token == nil {
		return nil, fmt.Errorf("token is nil")
	}
	return c.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
}

// LogoutURL returns the provider logout URL if configured.
func (c *Client) LogoutURL(idTokenHint string, postLogoutRedirect string) (string, error) {
	if c == nil {
		return "", ErrNilConfig
	}
	if c.endpoints.LogoutURL == "" {
		return "", fmt.Errorf("logout endpoint not configured")
	}

	req, err := http.NewRequest(http.MethodGet, c.endpoints.LogoutURL, nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	if idTokenHint != "" {
		q.Set("id_token_hint", idTokenHint)
	}
	if postLogoutRedirect != "" {
		q.Set("post_logout_redirect_uri", postLogoutRedirect)
	}
	req.URL.RawQuery = q.Encode()
	return req.URL.String(), nil
}

// TokenExpiry returns the expiry time for the access token when present.
func TokenExpiry(tok *oauth2.Token) time.Time {
	if tok == nil {
		return time.Time{}
	}
	return tok.Expiry
}

// VerifyToken verifies a raw ID token and returns parsed claims.
func (c *Client) VerifyToken(ctx context.Context, raw string) (*IDTokenClaims, error) {
	if c == nil {
		return nil, ErrNilConfig
	}
	if raw == "" {
		return nil, ErrMissingIDToken
	}
	idToken, err := c.verifier.Verify(ctx, raw)
	if err != nil {
		return nil, err
	}
	var claims IDTokenClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}
	return &claims, nil
}
