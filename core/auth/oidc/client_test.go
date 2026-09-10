package oidc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cshekharsharma/photon/core/session"
	gjwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type oidcTestServer struct {
	t             *testing.T
	srv           *httptest.Server
	issuer        string
	key           *rsa.PrivateKey
	kid           string
	idToken       string
	mu            sync.Mutex
	lastTokenForm url.Values
	omitAuthToken bool
}

func newOIDCTestServer(t *testing.T) *oidcTestServer {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := &oidcTestServer{
		t:   t,
		key: key,
		kid: "kid-1",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", s.handleDiscovery)
	mux.HandleFunc("/oauth/token", s.handleToken)
	mux.HandleFunc("/oauth/auth", s.handleAuth)
	mux.HandleFunc("/oauth/userinfo", s.handleUserInfo)
	mux.HandleFunc("/oauth/jwks", s.handleJWKS)

	s.srv = httptest.NewServer(mux)
	s.issuer = s.srv.URL
	return s
}

func newOIDCTestServerWithMissingEndpoints(t *testing.T) *oidcTestServer {
	t.Helper()
	s := newOIDCTestServer(t)
	s.omitAuthToken = true
	return s
}

func (s *oidcTestServer) close() {
	s.srv.Close()
}

func (s *oidcTestServer) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"issuer":                 s.issuer,
		"authorization_endpoint": "",
		"token_endpoint":         "",
		"jwks_uri":               s.issuer + "/oauth/jwks",
		"userinfo_endpoint":      s.issuer + "/oauth/userinfo",
		"end_session_endpoint":   s.issuer + "/oauth/logout",
	}
	if !s.omitAuthToken {
		resp["authorization_endpoint"] = s.issuer + "/oauth/auth"
		resp["token_endpoint"] = s.issuer + "/oauth/token"
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *oidcTestServer) handleToken(w http.ResponseWriter, r *http.Request) {
	require.NoError(s.t, r.ParseForm())

	s.mu.Lock()
	s.lastTokenForm = r.PostForm
	idToken := s.idToken
	s.mu.Unlock()

	if r.PostForm.Get("code") == "bad" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
		return
	}

	resp := map[string]interface{}{
		"access_token": "access",
		"token_type":   "Bearer",
		"expires_in":   3600,
	}
	if idToken != "" {
		resp["id_token"] = idToken
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *oidcTestServer) handleAuth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *oidcTestServer) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"sub":   "user-1",
		"email": "user@example.com",
	})
}

func (s *oidcTestServer) handleJWKS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	n := base64.RawURLEncoding.EncodeToString(s.key.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(bigIntToBytes(int64(s.key.E)))
	resp := map[string]interface{}{
		"keys": []map[string]string{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"kid": s.kid,
				"n":   n,
				"e":   e,
			},
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *oidcTestServer) signIDToken(t *testing.T, clientID, nonce string, exp time.Time) string {
	t.Helper()
	claims := gjwt.MapClaims{
		"iss":   s.issuer,
		"sub":   "user-1",
		"aud":   []string{clientID},
		"iat":   time.Now().Unix(),
		"exp":   exp.Unix(),
		"nonce": nonce,
		"email": "user@example.com",
	}
	token := gjwt.NewWithClaims(gjwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.kid
	signed, err := token.SignedString(s.key)
	require.NoError(t, err)
	return signed
}

func (s *oidcTestServer) signIDTokenWithClaims(t *testing.T, claims gjwt.MapClaims) string {
	t.Helper()
	token := gjwt.NewWithClaims(gjwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.kid
	signed, err := token.SignedString(s.key)
	require.NoError(t, err)
	return signed
}

func bigIntToBytes(v int64) []byte {
	if v == 0 {
		return []byte{0}
	}
	out := []byte{}
	for v > 0 {
		out = append([]byte{byte(v & 0xff)}, out...)
		v >>= 8
	}
	return out
}

func newSessionManager(t *testing.T) *session.Manager {
	t.Helper()
	mgr, err := session.New(&session.Config{Store: session.StoreConfig{Type: session.StoreMemory}})
	require.NoError(t, err)
	return mgr
}

func sessionCtx(t *testing.T, mgr *session.Manager) context.Context {
	t.Helper()
	ctx, err := mgr.Load(context.Background(), "test-token")
	require.NoError(t, err)
	return ctx
}

func TestNew_ConfigValidation(t *testing.T) {
	_, err := New(context.Background(), nil)
	assert.ErrorIs(t, err, ErrNilConfig)

	_, err = New(context.Background(), &Config{})
	assert.ErrorIs(t, err, ErrMissingClientID)

	_, err = New(context.Background(), &Config{ClientID: "c"})
	assert.ErrorIs(t, err, ErrMissingRedirectURL)

	mgr := newSessionManager(t)
	_, err = New(context.Background(), &Config{ClientID: "c", RedirectURL: "http://cb", Session: mgr})
	assert.ErrorIs(t, err, ErrMissingIssuer)
}

func TestNew_MissingSession(t *testing.T) {
	_, err := New(context.Background(), &Config{
		Issuer:      "http://invalid-issuer",
		ClientID:    "c",
		RedirectURL: "http://cb",
		Session:     nil,
	})
	assert.ErrorIs(t, err, ErrMissingSession)
}

func TestNew_ProviderError(t *testing.T) {
	mgr := newSessionManager(t)
	_, err := New(context.Background(), &Config{
		Issuer:      "http://127.0.0.1:0",
		ClientID:    "c",
		RedirectURL: "http://cb",
		Session:     mgr,
	})
	assert.Error(t, err)
}

func TestNew_MissingEndpoints(t *testing.T) {
	srv := newOIDCTestServerWithMissingEndpoints(t)
	defer srv.close()

	mgr := newSessionManager(t)
	_, err := New(context.Background(), &Config{
		Issuer:      srv.issuer,
		ClientID:    "c",
		RedirectURL: "http://cb",
		Session:     mgr,
	})
	assert.Error(t, err)
}

func TestNew_EndpointsOverrideAndAuthCodeURL(t *testing.T) {
	srv := newOIDCTestServer(t)
	defer srv.close()

	mgr := newSessionManager(t)
	ctx := sessionCtx(t, mgr)

	overrideAuth := srv.issuer + "/custom/auth"
	cfg := &Config{
		Issuer:      srv.issuer,
		ClientID:    "client-1",
		RedirectURL: "http://app/cb",
		Session:     mgr,
		Endpoints: &Endpoints{
			AuthURL:  overrideAuth,
			TokenURL: srv.issuer + "/oauth/token",
		},
		Scopes: []string{"email"},
	}

	client, err := New(context.Background(), cfg)
	require.NoError(t, err)

	authURL, err := client.AuthCodeURL(ctx, oauth2.SetAuthURLParam("prompt", "login"))
	require.NoError(t, err)

	parsed, err := url.Parse(authURL)
	require.NoError(t, err)
	q := parsed.Query()

	assert.Equal(t, overrideAuth, parsed.Scheme+"://"+parsed.Host+parsed.Path)

	state, err := mgr.GetString(ctx, sessionStateKey)
	require.NoError(t, err)
	nonce, err := mgr.GetString(ctx, sessionNonceKey)
	require.NoError(t, err)
	verifier, err := mgr.GetString(ctx, sessionVerifierKey)
	require.NoError(t, err)

	assert.Equal(t, state, q.Get("state"))
	assert.Equal(t, nonce, q.Get("nonce"))
	assert.Equal(t, "S256", q.Get("code_challenge_method"))
	assert.Equal(t, challengeFromVerifier(verifier), q.Get("code_challenge"))
	assert.Equal(t, "login", q.Get("prompt"))
	assert.True(t, strings.Contains(q.Get("scope"), "openid"))
	assert.True(t, strings.Contains(q.Get("scope"), "email"))
}

func TestHandleCallback_SuccessAndCleanup(t *testing.T) {
	srv := newOIDCTestServer(t)
	defer srv.close()

	mgr := newSessionManager(t)
	ctx := sessionCtx(t, mgr)

	cfg := &Config{
		Issuer:      srv.issuer,
		ClientID:    "client-1",
		RedirectURL: "http://app/cb",
		Session:     mgr,
	}

	client, err := New(context.Background(), cfg)
	require.NoError(t, err)

	state := "state-1"
	nonce := "nonce-1"
	verifier := "verifier-1"
	require.NoError(t, mgr.Put(ctx, sessionStateKey, state))
	require.NoError(t, mgr.Put(ctx, sessionNonceKey, nonce))
	require.NoError(t, mgr.Put(ctx, sessionVerifierKey, verifier))

	srv.mu.Lock()
	srv.idToken = srv.signIDToken(t, cfg.ClientID, nonce, time.Now().Add(time.Hour))
	srv.mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "http://app/cb?code=good&state="+state, nil)
	tokens, claims, err := client.HandleCallback(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, tokens)
	require.NotNil(t, claims)
	assert.Equal(t, "user-1", claims.Subject)

	// session keys should be cleaned up
	gotState, _ := mgr.GetString(ctx, sessionStateKey)
	gotNonce, _ := mgr.GetString(ctx, sessionNonceKey)
	gotVerifier, _ := mgr.GetString(ctx, sessionVerifierKey)
	assert.Equal(t, "", gotState)
	assert.Equal(t, "", gotNonce)
	assert.Equal(t, "", gotVerifier)

	srv.mu.Lock()
	form := srv.lastTokenForm
	srv.mu.Unlock()
	assert.Equal(t, verifier, form.Get("code_verifier"))
}

func TestHandleCallback_Errors(t *testing.T) {
	srv := newOIDCTestServer(t)
	defer srv.close()

	mgr := newSessionManager(t)
	ctx := sessionCtx(t, mgr)

	cfg := &Config{
		Issuer:      srv.issuer,
		ClientID:    "client-1",
		RedirectURL: "http://app/cb",
		Session:     mgr,
	}
	client, err := New(context.Background(), cfg)
	require.NoError(t, err)

	_, _, err = (*Client)(nil).HandleCallback(ctx, nil)
	assert.ErrorIs(t, err, ErrNilConfig)

	_, _, err = client.HandleCallback(ctx, nil)
	assert.ErrorIs(t, err, ErrMissingCode)

	req := httptest.NewRequest(http.MethodGet, "http://app/cb", nil)
	_, _, err = client.HandleCallback(ctx, req)
	assert.ErrorIs(t, err, ErrMissingCode)

	req = httptest.NewRequest(http.MethodGet, "http://app/cb?code=ok&state=state", nil)
	_, _, err = client.HandleCallback(ctx, req)
	assert.ErrorIs(t, err, ErrMissingState)

	req = httptest.NewRequest(http.MethodGet, "http://app/cb?code=ok&state=state", nil)
	_, _, err = client.HandleCallback(context.Background(), req)
	assert.Error(t, err)

	require.NoError(t, mgr.Put(ctx, sessionStateKey, "state"))
	req = httptest.NewRequest(http.MethodGet, "http://app/cb?code=ok&state=bad", nil)
	_, _, err = client.HandleCallback(ctx, req)
	assert.ErrorIs(t, err, ErrInvalidState)

	req = httptest.NewRequest(http.MethodGet, "http://app/cb?code=ok&state=state", nil)
	_, _, err = client.HandleCallback(ctx, req)
	assert.ErrorIs(t, err, ErrMissingVerifier)

	require.NoError(t, mgr.Put(ctx, sessionVerifierKey, "ver-1"))
	req = httptest.NewRequest(http.MethodGet, "http://app/cb?code=bad&state=state", nil)
	_, _, err = client.HandleCallback(ctx, req)
	assert.Error(t, err)

	srv.mu.Lock()
	srv.idToken = ""
	srv.mu.Unlock()
	req = httptest.NewRequest(http.MethodGet, "http://app/cb?code=ok&state=state", nil)
	_, _, err = client.HandleCallback(ctx, req)
	assert.ErrorIs(t, err, ErrMissingIDToken)

	srv.mu.Lock()
	srv.idToken = "not-a-jwt"
	srv.mu.Unlock()
	req = httptest.NewRequest(http.MethodGet, "http://app/cb?code=ok&state=state", nil)
	_, _, err = client.HandleCallback(ctx, req)
	assert.Error(t, err)

	srv.mu.Lock()
	srv.idToken = srv.signIDToken(t, cfg.ClientID, "nonce-x", time.Now().Add(time.Hour))
	srv.mu.Unlock()
	req = httptest.NewRequest(http.MethodGet, "http://app/cb?code=ok&state=state", nil)
	_, _, err = client.HandleCallback(ctx, req)
	assert.ErrorIs(t, err, ErrMissingNonce)

	require.NoError(t, mgr.Put(ctx, sessionNonceKey, "nonce-good"))
	req = httptest.NewRequest(http.MethodGet, "http://app/cb?code=ok&state=state", nil)
	_, _, err = client.HandleCallback(ctx, req)
	assert.ErrorIs(t, err, ErrInvalidState)

	badClaims := gjwt.MapClaims{
		"iss":   srv.issuer,
		"sub":   "x",
		"aud":   []string{cfg.ClientID},
		"exp":   "bad",
		"iat":   1,
		"nonce": "nonce-good",
	}
	srv.mu.Lock()
	srv.idToken = srv.signIDTokenWithClaims(t, badClaims)
	srv.mu.Unlock()

	req = httptest.NewRequest(http.MethodGet, "http://app/cb?code=ok&state=state", nil)
	_, _, err = client.HandleCallback(ctx, req)
	assert.Error(t, err)

	validButBadClaims := gjwt.MapClaims{
		"iss":   srv.issuer,
		"sub":   "x",
		"aud":   cfg.ClientID,
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"nonce": "nonce-good",
	}
	srv.mu.Lock()
	srv.idToken = srv.signIDTokenWithClaims(t, validButBadClaims)
	srv.mu.Unlock()

	req = httptest.NewRequest(http.MethodGet, "http://app/cb?code=ok&state=state", nil)
	_, _, err = client.HandleCallback(ctx, req)
	assert.Error(t, err)
}

func TestUserInfoAndLogoutAndVerify(t *testing.T) {
	srv := newOIDCTestServer(t)
	defer srv.close()

	mgr := newSessionManager(t)
	ctx := sessionCtx(t, mgr)

	cfg := &Config{
		Issuer:      srv.issuer,
		ClientID:    "client-1",
		RedirectURL: "http://app/cb",
		Session:     mgr,
	}
	client, err := New(context.Background(), cfg)
	require.NoError(t, err)

	_, err = client.UserInfo(ctx, nil)
	assert.Error(t, err)

	tok := &oauth2.Token{AccessToken: "access", TokenType: "Bearer"}
	clientNoProvider := &Client{}
	_, err = clientNoProvider.UserInfo(ctx, tok)
	assert.Error(t, err)

	ui, err := client.UserInfo(ctx, tok)
	require.NoError(t, err)
	assert.Equal(t, "user-1", ui.Subject)

	_, err = (*Client)(nil).LogoutURL("", "")
	assert.ErrorIs(t, err, ErrNilConfig)

	_, err = client.LogoutURL("", "")
	assert.Error(t, err)

	client.endpoints.LogoutURL = "http://[::1"
	_, err = client.LogoutURL("", "")
	assert.Error(t, err)

	client.endpoints.LogoutURL = srv.issuer + "/oauth/logout"
	logoutURL, err := client.LogoutURL("id", "http://post")
	require.NoError(t, err)
	assert.Contains(t, logoutURL, "id_token_hint=id")
	assert.Contains(t, logoutURL, "post_logout_redirect_uri=http%3A%2F%2Fpost")

	_, err = (*Client)(nil).VerifyToken(ctx, "x")
	assert.ErrorIs(t, err, ErrNilConfig)

	_, err = client.VerifyToken(ctx, "")
	assert.ErrorIs(t, err, ErrMissingIDToken)

	_, err = client.VerifyToken(ctx, "not-a-jwt")
	assert.Error(t, err)

	raw := srv.signIDToken(t, cfg.ClientID, "n", time.Now().Add(time.Hour))
	claims, err := client.VerifyToken(ctx, raw)
	require.NoError(t, err)
	assert.Equal(t, "user-1", claims.Subject)
}

func TestHelpers(t *testing.T) {
	verifier, err := generateVerifier()
	require.NoError(t, err)
	assert.NotEmpty(t, verifier)

	ch := challengeFromVerifier("verifier")
	assert.NotEmpty(t, ch)

	scopes := normalizeScopes([]string{"", "openid", "email"})
	assert.Equal(t, []string{"openid", "email"}, scopes)

	exp := TokenExpiry(nil)
	assert.True(t, exp.IsZero())

	claims := IDTokenClaims{Expiry: 0}
	assert.True(t, claims.ExpiryTime().IsZero())

	now := time.Now().Add(time.Hour)
	claims.Expiry = now.Unix()
	assert.Equal(t, now.Unix(), claims.ExpiryTime().Unix())
}

func TestAuthCodeURL_NoSessionInContext(t *testing.T) {
	srv := newOIDCTestServer(t)
	defer srv.close()

	mgr := newSessionManager(t)
	client, err := New(context.Background(), &Config{
		Issuer:      srv.issuer,
		ClientID:    "client-1",
		RedirectURL: "http://app/cb",
		Session:     mgr,
	})
	require.NoError(t, err)

	_, err = client.AuthCodeURL(context.Background())
	assert.Error(t, err)
}

func TestAuthCodeURL_NilClientAndRandFailure(t *testing.T) {
	_, err := (*Client)(nil).AuthCodeURL(context.Background())
	assert.ErrorIs(t, err, ErrNilConfig)

	orig := randReader
	defer func() { randReader = orig }()
	randReader = failingReader{}

	srv := newOIDCTestServer(t)
	defer srv.close()

	mgr := newSessionManager(t)
	ctx := sessionCtx(t, mgr)
	client, err := New(context.Background(), &Config{
		Issuer:      srv.issuer,
		ClientID:    "client-1",
		RedirectURL: "http://app/cb",
		Session:     mgr,
	})
	require.NoError(t, err)

	_, err = client.AuthCodeURL(ctx)
	assert.Error(t, err)
}

func TestAuthCodeURL_PutErrors(t *testing.T) {
	client := &Client{
		sessions: &stubSessions{values: map[string]string{}, putErrAt: 2},
		oauth2Cfg: oauth2.Config{
			ClientID:    "c",
			RedirectURL: "http://cb",
			Endpoint: oauth2.Endpoint{
				AuthURL: "http://auth",
			},
		},
	}

	_, err := client.AuthCodeURL(context.Background())
	assert.Error(t, err)

	client.sessions = &stubSessions{values: map[string]string{}, putErrAt: 3}
	_, err = client.AuthCodeURL(context.Background())
	assert.Error(t, err)
}

func TestVerifyToken_ClaimsError(t *testing.T) {
	srv := newOIDCTestServer(t)
	defer srv.close()

	mgr := newSessionManager(t)
	ctx := sessionCtx(t, mgr)

	cfg := &Config{
		Issuer:      srv.issuer,
		ClientID:    "client-1",
		RedirectURL: "http://app/cb",
		Session:     mgr,
	}
	client, err := New(context.Background(), cfg)
	require.NoError(t, err)

	badClaims := gjwt.MapClaims{
		"iss":   srv.issuer,
		"sub":   "x",
		"aud":   cfg.ClientID,
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"nonce": "n",
	}
	raw := srv.signIDTokenWithClaims(t, badClaims)
	_, err = client.VerifyToken(ctx, raw)
	assert.Error(t, err)
}

func TestHandleCallback_GetStringErrors(t *testing.T) {
	srv := newOIDCTestServer(t)
	defer srv.close()

	mgr := newSessionManager(t)
	cfg := &Config{
		Issuer:      srv.issuer,
		ClientID:    "client-1",
		RedirectURL: "http://app/cb",
		Session:     mgr,
	}
	client, err := New(context.Background(), cfg)
	require.NoError(t, err)

	client.sessions = &stubSessions{
		values: map[string]string{
			sessionStateKey:    "state",
			sessionVerifierKey: "verifier",
		},
		getErrAt: 2,
	}
	req := httptest.NewRequest(http.MethodGet, "http://app/cb?code=ok&state=state", nil)
	_, _, err = client.HandleCallback(context.Background(), req)
	assert.Error(t, err)

	srv.mu.Lock()
	srv.idToken = srv.signIDToken(t, cfg.ClientID, "nonce-ok", time.Now().Add(time.Hour))
	srv.mu.Unlock()

	client.sessions = &stubSessions{
		values: map[string]string{
			sessionStateKey:    "state",
			sessionVerifierKey: "verifier",
		},
		getErrAt: 3,
	}
	req = httptest.NewRequest(http.MethodGet, "http://app/cb?code=ok&state=state", nil)
	_, _, err = client.HandleCallback(context.Background(), req)
	assert.Error(t, err)
}

type stubSessions struct {
	values   map[string]string
	putErrAt int
	getErrAt int
	putCalls int
	getCalls int
}

func (s *stubSessions) Put(_ context.Context, key string, val interface{}) error {
	s.putCalls++
	if s.putErrAt == s.putCalls {
		return errors.New("put fail")
	}
	if v, ok := val.(string); ok {
		s.values[key] = v
	}
	return nil
}

func (s *stubSessions) GetString(_ context.Context, key string) (string, error) {
	s.getCalls++
	if s.getErrAt == s.getCalls {
		return "", errors.New("get fail")
	}
	return s.values[key], nil
}

func (s *stubSessions) Remove(_ context.Context, _ string) error {
	return nil
}

func TestAuthCodeURL_GenerateVerifierStages(t *testing.T) {
	srv := newOIDCTestServer(t)
	defer srv.close()

	mgr := newSessionManager(t)
	ctx := sessionCtx(t, mgr)
	client, err := New(context.Background(), &Config{
		Issuer:      srv.issuer,
		ClientID:    "client-1",
		RedirectURL: "http://app/cb",
		Session:     mgr,
	})
	require.NoError(t, err)

	orig := randReader
	defer func() { randReader = orig }()

	randReader = &seqReader{failAt: 2}
	_, err = client.AuthCodeURL(ctx)
	assert.Error(t, err)

	randReader = &seqReader{failAt: 3}
	_, err = client.AuthCodeURL(ctx)
	assert.Error(t, err)
}

type seqReader struct {
	reads  int
	failAt int
}

func (s *seqReader) Read(p []byte) (int, error) {
	s.reads++
	if s.reads == s.failAt {
		return 0, errors.New("boom")
	}
	for i := range p {
		p[i] = byte(i)
	}
	return len(p), nil
}
