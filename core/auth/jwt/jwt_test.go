package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"
	"time"

	"github.com/cshekharsharma/photon/utils/types"
	gjwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

var (
	privateKey string = "LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1JSGNBZ0VCQkVJQjBwRTR1RmFXUng3dDAzQnNZbFl2RjFZdkthQkd5dm9ha3hub2RtOW91MFI5d0Mrc0pBakgKUVpaSmlrT2c0U3dOcWdRL2h5ck91REsyb0FWSGhnVkdjWW1nQndZRks0RUVBQ09oZ1lrRGdZWUFCQUFKWEl1dwoxMk1VenBIZ2dpYTlQT0JGWVhTeGFPR0tHYk1qSXlESSs2cTd3aTdMTXczSGdiYU9tZ0lxRkc3Mm84SkJRd1lOCjRJYlhIZitmODZDUlkxQUEyd0h6Ykh2dDZJaGtDWFROeEJFZmZhMXlNVWd1OG45Y0tLRjJpTGd5UUtjS3FXMzMKOGZHT3cvbjNSbTJZZC9FQjU2dTJybkQyOXFTK25PTTllR1MrZ3kzOU9RPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQ=="
	publicKey  string = "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUlHYk1CQUdCeXFHU000OUFnRUdCU3VCQkFBakE0R0dBQVFBQ1Z5THNOZGpGTTZSNElJbXZUemdSV0Ywc1dqaAppaG16SXlNZ3lQdXF1OEl1eXpNTng0RzJqcG9DS2hSdTlxUENRVU1HRGVDRzF4My9uL09na1dOUUFOc0I4Mng3CjdlaUlaQWwwemNRUkgzMnRjakZJTHZKL1hDaWhkb2k0TWtDbkNxbHQ5L0h4anNQNTkwWnRtSGZ4QWVlcnRxNXcKOXZha3ZwempQWGhrdm9NdC9Uaz0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0t"
	hmacKey    string = "super-secret-hmac-key"
)

func testFeed() JwtFeed {
	return JwtFeed{
		Subject:    "test",
		Audience:   "testing",
		Issuer:     "omega",
		IssuedAt:   time.Now().Unix(),
		Expiration: time.Now().Unix() + 100,
		NotBefore:  time.Now().Unix() - 100,
		CustomClaims: map[string]interface{}{
			"Foo": "Bar",
		},
	}
}

func TestSignAndVerify_ECDSA_Success(t *testing.T) {
	feed := testFeed()
	signature, err := SignECDSA(feed, privateKey, "KID-1", SigningMethodES512)
	assert.Nil(t, err)

	valid, token, verifyErr := Verify(signature, VerifyConfig{
		ECDSAPublicKeys:  map[string]string{"KID-1": publicKey},
		ExpectedIssuer:   "omega",
		ExpectedAudience: "testing",
	})

	assert.True(t, valid)
	assert.Nil(t, verifyErr)
	v, ok := GetClaim(token, "sub")
	assert.True(t, ok)
	assert.Equal(t, "test", v)
}

func TestSignAndVerify_ECDSA_ES384Success(t *testing.T) {
	es384PrivateKey, es384PublicKey := generateECDSAKeyPairForTest(t, elliptic.P384())

	signature, err := SignECDSA(testFeed(), es384PrivateKey, "KID-ES384", SigningMethodES384)
	assert.Nil(t, err)

	valid, token, verifyErr := Verify(signature, VerifyConfig{
		ECDSAPublicKeys:  map[string]string{"KID-ES384": es384PublicKey},
		ExpectedIssuer:   "omega",
		ExpectedAudience: "testing",
	})

	assert.True(t, valid)
	assert.Nil(t, verifyErr)
	assert.NotNil(t, token)
	assert.Equal(t, SigningMethodES384.Alg(), token.Method.Alg())
}

func generateECDSAKeyPairForTest(t *testing.T, curve elliptic.Curve) (string, string) {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ECDSA key: %v", err)
	}

	privateDER, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		t.Fatalf("failed to marshal private key: %v", err)
	}
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}

	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privateDER})
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})

	return base64.StdEncoding.EncodeToString(privatePEM), base64.StdEncoding.EncodeToString(publicPEM)
}

func TestSignAndVerify_HMAC_Success(t *testing.T) {
	feed := testFeed()
	feed.Subject = "hmac-user"
	signature, err := SignHMAC(feed, hmacKey, "KID-HS-1", SigningMethodHS256)
	assert.Nil(t, err)

	valid, token, verifyErr := Verify(signature, VerifyConfig{
		HMACSecrets:      map[string]string{"KID-HS-1": hmacKey},
		ExpectedIssuer:   "omega",
		ExpectedAudience: "testing",
	})
	assert.True(t, valid)
	assert.Nil(t, verifyErr)
	v, ok := GetClaim(token, "sub")
	assert.True(t, ok)
	assert.Equal(t, "hmac-user", v)
}

func TestParseUnverifiedToken(t *testing.T) {
	signedToken, err := SignECDSA(testFeed(), privateKey, "KID-1", SigningMethodES512)
	assert.Nil(t, err)

	jwtToken, err := ParseUnverifiedToken(signedToken)
	assert.Nil(t, err)
	iss, ok := GetClaim(jwtToken, "iss")
	assert.True(t, ok)
	assert.Equal(t, "omega", iss)
}

func TestGetClaim_SafeAccess(t *testing.T) {
	v, ok := GetClaim(nil, "sub")
	assert.False(t, ok)
	assert.Nil(t, v)

	tok := &gjwt.Token{Claims: &gjwt.RegisteredClaims{Issuer: "x"}}
	v, ok = GetClaim(tok, "iss")
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestSign_Errors(t *testing.T) {
	feed := testFeed()

	_, err := SignECDSA(feed, "k", "kid", nil)
	assert.NotNil(t, err)

	_, err = SignHMAC(feed, "k", "", SigningMethodHS256)
	assert.NotNil(t, err)

	_, err = SignHMAC(feed, "", "kid", SigningMethodHS256)
	assert.NotNil(t, err)

	_, err = SignECDSA(feed, "", "kid", SigningMethodES512)
	assert.NotNil(t, err)

	_, err = SignECDSA(feed, "%%%not-base64%%%", "KID-1", SigningMethodES512)
	assert.NotNil(t, err)

	invalidPEMBase64 := base64.StdEncoding.EncodeToString([]byte("not-a-pem-private-key"))
	_, err = SignECDSA(feed, invalidPEMBase64, "KID-1", SigningMethodES512)
	assert.NotNil(t, err)
}

func TestSignSpecific_Errors(t *testing.T) {
	feed := testFeed()

	_, err := SignECDSA(feed, privateKey, "KID-1", nil)
	assert.NotNil(t, err)

	_, err = SignECDSA(feed, privateKey, "", SigningMethodES512)
	assert.NotNil(t, err)

	_, err = SignHMAC(feed, hmacKey, "KID-HS-1", nil)
	assert.NotNil(t, err)
}

func TestVerify_ErrorsAndPolicy(t *testing.T) {
	feed := testFeed()
	hmacToken, err := SignHMAC(feed, hmacKey, "KID-HS-1", SigningMethodHS256)
	assert.Nil(t, err)
	ecdsaToken, err := SignECDSA(feed, privateKey, "KID-EC-1", SigningMethodES512)
	assert.Nil(t, err)

	valid, _, err := Verify("", VerifyConfig{HMACSecrets: map[string]string{"KID-HS-1": hmacKey}})
	assert.False(t, valid)
	assert.NotNil(t, err)

	valid, _, err = Verify(hmacToken, VerifyConfig{})
	assert.False(t, valid)
	assert.NotNil(t, err)

	valid, _, err = Verify(hmacToken, VerifyConfig{HMACSecrets: map[string]string{}})
	assert.False(t, valid)
	assert.NotNil(t, err)

	valid, _, err = Verify(hmacToken, VerifyConfig{HMACSecrets: map[string]string{"KID-HS-1": "wrong"}})
	assert.False(t, valid)
	assert.NotNil(t, err)

	valid, _, err = Verify(hmacToken, VerifyConfig{HMACSecrets: map[string]string{"KID-HS-1": hmacKey}, ExpectedIssuer: "wrong"})
	assert.False(t, valid)
	assert.NotNil(t, err)

	valid, _, err = Verify(hmacToken, VerifyConfig{HMACSecrets: map[string]string{"KID-HS-1": hmacKey}, ExpectedAudience: "wrong"})
	assert.False(t, valid)
	assert.NotNil(t, err)

	valid, _, err = Verify(ecdsaToken, VerifyConfig{ECDSAPublicKeys: map[string]string{"KID-EC-1": "%%%not-base64%%%"}})
	assert.False(t, valid)
	assert.NotNil(t, err)

	valid, _, err = Verify(ecdsaToken, VerifyConfig{ECDSAPublicKeys: map[string]string{}})
	assert.False(t, valid)
	assert.NotNil(t, err)
}

func TestVerify_RejectsAlgConfusionAttack(t *testing.T) {
	// Attacker signs HS256 token using publicly known ECDSA public-key string as HMAC secret.
	feed := testFeed()
	attackToken, err := SignHMAC(feed, publicKey, "KID-1", SigningMethodHS256)
	assert.Nil(t, err)

	valid, _, verifyErr := Verify(attackToken, VerifyConfig{
		ECDSAPublicKeys: map[string]string{"KID-1": publicKey},
	})
	assert.False(t, valid)
	assert.NotNil(t, verifyErr)
}

func TestVerify_UnsupportedSigningMethodToken(t *testing.T) {
	feed := testFeed()
	feed.Subject = "rs-user"

	token := gjwt.NewWithClaims(gjwt.SigningMethodRS256, gjwt.MapClaims{
		JwtClaimAudience:   feed.Audience,
		JwtClaimIssuer:     feed.Issuer,
		JwtClaimSubject:    feed.Subject,
		JwtClaimIssuedAt:   feed.IssuedAt,
		JwtClaimExpiration: feed.Expiration,
		JwtClaimNotBefore:  feed.NotBefore,
	})
	token.Header["kid"] = "KID-RS-1"

	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.Nil(t, err)

	signedToken, err := token.SignedString(rsaPrivateKey)
	assert.Nil(t, err)

	valid, _, verifyErr := Verify(signedToken, VerifyConfig{
		ECDSAPublicKeys: map[string]string{"KID-1": publicKey},
		HMACSecrets:     map[string]string{"KID-HS-1": hmacKey},
	})
	assert.False(t, valid)
	assert.NotNil(t, verifyErr)
}

func TestVerify_InvalidOrMissingKID(t *testing.T) {
	baseClaims := gjwt.MapClaims{
		JwtClaimAudience:   "testing",
		JwtClaimIssuer:     "omega",
		JwtClaimSubject:    "test",
		JwtClaimIssuedAt:   time.Now().Unix(),
		JwtClaimExpiration: time.Now().Unix() + 100,
		JwtClaimNotBefore:  time.Now().Unix() - 100,
	}

	t1 := gjwt.NewWithClaims(gjwt.SigningMethodHS256, baseClaims)
	signedNoKid, err := t1.SignedString([]byte(hmacKey))
	assert.Nil(t, err)
	valid, _, err := Verify(signedNoKid, VerifyConfig{HMACSecrets: map[string]string{"KID": hmacKey}})
	assert.False(t, valid)
	assert.NotNil(t, err)

	t2 := gjwt.NewWithClaims(gjwt.SigningMethodHS256, baseClaims)
	t2.Header["kid"] = 12345
	signedBadKid, err := t2.SignedString([]byte(hmacKey))
	assert.Nil(t, err)
	valid, _, err = Verify(signedBadKid, VerifyConfig{HMACSecrets: map[string]string{"KID": hmacKey}})
	assert.False(t, valid)
	assert.NotNil(t, err)
}

func TestVerify_InvalidTokenString(t *testing.T) {
	valid, token, err := Verify("not-a-jwt", VerifyConfig{
		HMACSecrets: map[string]string{"KID-HS-1": hmacKey},
	})
	assert.False(t, valid)
	assert.Nil(t, token)
	assert.NotNil(t, err)
}

func TestVerifyECDSA_MissingKeyForKID(t *testing.T) {
	signedToken, err := SignECDSA(testFeed(), privateKey, "KID-EC-1", SigningMethodES512)
	assert.Nil(t, err)

	valid, _, err := VerifyECDSA(signedToken, VerifyConfig{
		ECDSAPublicKeys: map[string]string{"OTHER-KID": publicKey},
	})
	assert.False(t, valid)
	assert.NotNil(t, err)
}

func TestVerifyHMAC_MissingSecretForKID(t *testing.T) {
	signedToken, err := SignHMAC(testFeed(), hmacKey, "KID-HS-1", SigningMethodHS256)
	assert.Nil(t, err)

	valid, _, err := VerifyHMAC(signedToken, VerifyConfig{
		HMACSecrets: map[string]string{"OTHER-KID": hmacKey},
	})
	assert.False(t, valid)
	assert.NotNil(t, err)
}

func TestVerifyECDSA_InvalidOrMissingKID(t *testing.T) {
	decodedPrivateKey, err := types.Base64Decode(privateKey)
	assert.Nil(t, err)
	ecPriv, err := gjwt.ParseECPrivateKeyFromPEM([]byte(decodedPrivateKey))
	assert.Nil(t, err)

	baseClaims := gjwt.MapClaims{
		JwtClaimAudience:   "testing",
		JwtClaimIssuer:     "omega",
		JwtClaimSubject:    "test",
		JwtClaimIssuedAt:   time.Now().Unix(),
		JwtClaimExpiration: time.Now().Unix() + 100,
		JwtClaimNotBefore:  time.Now().Unix() - 100,
	}

	t1 := gjwt.NewWithClaims(gjwt.SigningMethodES512, baseClaims)
	signedNoKid, err := t1.SignedString(ecPriv)
	assert.Nil(t, err)
	valid, _, err := VerifyECDSA(signedNoKid, VerifyConfig{ECDSAPublicKeys: map[string]string{"KID-1": publicKey}})
	assert.False(t, valid)
	assert.NotNil(t, err)

	t2 := gjwt.NewWithClaims(gjwt.SigningMethodES512, baseClaims)
	t2.Header["kid"] = 12345
	signedBadKid, err := t2.SignedString(ecPriv)
	assert.Nil(t, err)
	valid, _, err = VerifyECDSA(signedBadKid, VerifyConfig{ECDSAPublicKeys: map[string]string{"KID-1": publicKey}})
	assert.False(t, valid)
	assert.NotNil(t, err)
}
