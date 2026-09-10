package jwt

import "context"

// Verifier adapts VerifyConfig to the core/auth Verifier interface.
type Verifier struct {
	Config VerifyConfig
}

// VerifyToken verifies a JWT string and returns the parsed token on success.
func (v Verifier) VerifyToken(_ context.Context, token string) (interface{}, error) {
	_, parsed, err := Verify(token, v.Config)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}
