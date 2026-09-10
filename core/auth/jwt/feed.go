package jwt

// JwtFeed represents the payload of a JWT (JSON Web Token) used for authentication
// and information exchange in applications. It encapsulates standard JWT claims
// as well as custom claims that can be provided by the application for specific
// use cases.
//
// Fields:
//   - Subject: Identifies the principal that is the subject of the JWT.
//   - Audience: Intended recipients for which the JWT is intended.
//   - Issuer: Identifies the principal that issued the JWT.
//   - IssuedAt: Timestamp indicating when the JWT was issued. It is represented
//     as seconds since Unix epoch.
//   - Expiration: Timestamp indicating the time after which the JWT is no longer valid.
//     It is represented as seconds since Unix epoch.
//   - NotBefore: Timestamp indicating the time before which the JWT should not be accepted
//     for processing. It is represented as seconds since Unix epoch.
//   - CustomClaims: A map containing custom claims provided by the application. These
//     can be used to convey additional information or attributes about the subject
//     or other aspects related to the token's use.
type JwtFeed struct {
	Subject      string
	Audience     string
	Issuer       string
	IssuedAt     int64
	Expiration   int64
	NotBefore    int64
	CustomClaims map[string]interface{}
}
