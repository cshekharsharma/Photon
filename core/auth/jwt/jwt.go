// Package jwt provides an easy to use capability to work with JWT
// auth mechanism.
//
// This package internally uses [github.com/golang-jwt/jwt/v5] as raw
// implementation of the jwt specifications.
package jwt

import (
	"fmt"
	"strings"

	"github.com/cshekharsharma/photon/utils/types"

	"github.com/golang-jwt/jwt/v5"
)

// Standard JWT claims
const (
	JwtClaimAudience   string = "aud"
	JwtClaimIssuer     string = "iss"
	JwtClaimSubject    string = "sub"
	JwtClaimIssuedAt   string = "iat"
	JwtClaimExpiration string = "exp"
	JwtClaimNotBefore  string = "nbf"
	JwtClaimData       string = "data"
)

// SIGNING METHODS
var (
	SigningMethodES256 *jwt.SigningMethodECDSA = jwt.SigningMethodES256
	SigningMethodES384 *jwt.SigningMethodECDSA = jwt.SigningMethodES384
	SigningMethodES512 *jwt.SigningMethodECDSA = jwt.SigningMethodES512

	SigningMethodHS256 *jwt.SigningMethodHMAC = jwt.SigningMethodHS256
	SigningMethodHS384 *jwt.SigningMethodHMAC = jwt.SigningMethodHS384
	SigningMethodHS512 *jwt.SigningMethodHMAC = jwt.SigningMethodHS512
)

// VerifyConfig defines verifier behavior and accepted key material.
type VerifyConfig struct {
	ECDSAPublicKeys map[string]string // Base64-encoded PEM ECDSA public keys by kid.

	HMACSecrets map[string]string // Raw HMAC shared secrets by kid.

	ExpectedIssuer   string
	ExpectedAudience string
}

type keyResolver = jwt.Keyfunc

// Creates new JWT token with claims and returns an instance of `jwt.Token`.
// Feed JwtFeed : Struct of all standard and custom claims for jwt token.
func NewToken(feed JwtFeed, signingMethod jwt.SigningMethod) *jwt.Token {
	jwtClaims := jwt.MapClaims{
		JwtClaimAudience:   feed.Audience,
		JwtClaimIssuer:     feed.Issuer,
		JwtClaimSubject:    feed.Subject,
		JwtClaimIssuedAt:   feed.IssuedAt,
		JwtClaimExpiration: feed.Expiration,
		JwtClaimNotBefore:  feed.NotBefore,
		JwtClaimData:       feed.CustomClaims,
	}

	return jwt.NewWithClaims(signingMethod, jwtClaims)
}

// SignECDSA signs the token with an ECDSA private key.
// privateKey should be a base64-encoded PEM key.
func SignECDSA(feed JwtFeed, privateKey string, kid string, signingMethod *jwt.SigningMethodECDSA) (string, error) {
	if signingMethod == nil {
		return "", fmt.Errorf("ecdsa signing method is required")
	}
	if strings.TrimSpace(kid) == "" {
		return "", fmt.Errorf("kid is required")
	}
	if strings.TrimSpace(privateKey) == "" {
		return "", fmt.Errorf("ecdsa private key is required")
	}

	token := NewToken(feed, signingMethod)
	token.Header["kid"] = kid

	decodedPrivateKey, err := types.Base64Decode(privateKey)
	if err != nil {
		return "", err
	}
	parsedECDSA, err := jwt.ParseECPrivateKeyFromPEM([]byte(decodedPrivateKey))
	if err != nil {
		return "", err
	}
	return token.SignedString(parsedECDSA)
}

// SignHMAC signs the token with an HMAC shared secret.
func SignHMAC(feed JwtFeed, secret string, kid string, signingMethod *jwt.SigningMethodHMAC) (string, error) {
	if signingMethod == nil {
		return "", fmt.Errorf("hmac signing method is required")
	}
	if strings.TrimSpace(kid) == "" {
		return "", fmt.Errorf("kid is required")
	}
	if strings.TrimSpace(secret) == "" {
		return "", fmt.Errorf("hmac secret is required")
	}

	token := NewToken(feed, signingMethod)
	token.Header["kid"] = kid
	return token.SignedString([]byte(secret))
}

// Verifies the provided JWT token string and returns
// token validity status, parsed token, and error.
func Verify(tokenString string, cfg VerifyConfig) (bool, *jwt.Token, error) {
	if strings.TrimSpace(tokenString) == "" {
		return false, nil, fmt.Errorf("token string is required")
	}

	unverifiedToken, err := ParseUnverifiedToken(tokenString)
	if err != nil || unverifiedToken == nil {
		return false, nil, err
	}

	switch unverifiedToken.Method.(type) {
	case *jwt.SigningMethodECDSA:
		return VerifyECDSA(tokenString, cfg)
	case *jwt.SigningMethodHMAC:
		return VerifyHMAC(tokenString, cfg)
	default:
		return false, nil, fmt.Errorf("unexpected signing method: %v", unverifiedToken.Header["alg"])
	}
}

// VerifyECDSA verifies tokens signed with ECDSA algorithms.
func VerifyECDSA(tokenString string, cfg VerifyConfig) (bool, *jwt.Token, error) {
	if len(cfg.ECDSAPublicKeys) == 0 {
		return false, nil, fmt.Errorf("at least one ecdsa public key must be provided")
	}

	return verifyWithConfig(
		tokenString,
		cfg,
		[]string{SigningMethodES256.Alg(), SigningMethodES384.Alg(), SigningMethodES512.Alg()},
		func(token *jwt.Token) (interface{}, error) {
			kidKey, err := tokenKID(token)
			if err != nil {
				return nil, err
			}
			key := cfg.ECDSAPublicKeys[kidKey]
			if strings.TrimSpace(key) == "" {
				return nil, fmt.Errorf("missing ecdsa public key for kid: %s", kidKey)
			}
			publicKey, err := types.Base64Decode(key)
			if err != nil {
				return nil, err
			}
			return jwt.ParseECPublicKeyFromPEM([]byte(publicKey))
		},
	)
}

// VerifyHMAC verifies tokens signed with HMAC algorithms.
func VerifyHMAC(tokenString string, cfg VerifyConfig) (bool, *jwt.Token, error) {
	if len(cfg.HMACSecrets) == 0 {
		return false, nil, fmt.Errorf("at least one hmac secret must be provided")
	}

	return verifyWithConfig(
		tokenString,
		cfg,
		[]string{SigningMethodHS256.Alg(), SigningMethodHS384.Alg(), SigningMethodHS512.Alg()},
		func(token *jwt.Token) (interface{}, error) {
			kidKey, err := tokenKID(token)
			if err != nil {
				return nil, err
			}
			secret := cfg.HMACSecrets[kidKey]
			if strings.TrimSpace(secret) == "" {
				return nil, fmt.Errorf("missing hmac secret for kid: %s", kidKey)
			}
			return []byte(secret), nil
		},
	)
}

func verifyWithConfig(tokenString string, cfg VerifyConfig, validMethods []string, resolver keyResolver) (bool, *jwt.Token, error) {
	claims := &jwt.MapClaims{}
	parserOptions := buildParserOptions(cfg, validMethods)

	token, err := jwt.ParseWithClaims(tokenString, claims, resolver, parserOptions...)
	if err != nil || token == nil {
		return false, token, err
	}

	return token.Valid, token, nil
}

func buildParserOptions(cfg VerifyConfig, validMethods []string) []jwt.ParserOption {
	parserOptions := []jwt.ParserOption{jwt.WithValidMethods(validMethods)}
	if strings.TrimSpace(cfg.ExpectedIssuer) != "" {
		parserOptions = append(parserOptions, jwt.WithIssuer(cfg.ExpectedIssuer))
	}
	if strings.TrimSpace(cfg.ExpectedAudience) != "" {
		parserOptions = append(parserOptions, jwt.WithAudience(cfg.ExpectedAudience))
	}
	return parserOptions
}

func tokenKID(token *jwt.Token) (string, error) {
	kidValue, ok := token.Header["kid"]
	if !ok {
		return "", fmt.Errorf("missing kid header")
	}
	kidKey, ok := kidValue.(string)
	if !ok || strings.TrimSpace(kidKey) == "" {
		return "", fmt.Errorf("invalid kid header")
	}
	return kidKey, nil
}

func ParseUnverifiedToken(tokenString string) (*jwt.Token, error) {
	claims := &jwt.MapClaims{}
	jwtToken, _, err := jwt.NewParser().ParseUnverified(tokenString, claims)
	return jwtToken, err
}

// Get claims by claim-key from provided JWT token.
func GetClaim(token *jwt.Token, key string) (interface{}, bool) {
	if token == nil || token.Claims == nil {
		return nil, false
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok || claims == nil {
		return nil, false
	}

	value, exists := (*claims)[key]
	return value, exists
}
