package auth

import "context"

// Verifier validates a token and returns provider-specific claims.
type Verifier interface {
	VerifyToken(ctx context.Context, token string) (interface{}, error)
}
