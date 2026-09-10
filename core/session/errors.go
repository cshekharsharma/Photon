package session

import "errors"

var (
	ErrNotInitialized     = errors.New("session manager not initialized")
	ErrNilConfig          = errors.New("session config cannot be nil")
	ErrInvalidStore       = errors.New("invalid session store type")
	ErrInvalidEncoding    = errors.New("invalid session encoding")
	ErrInvalidCookie      = errors.New("invalid cookie configuration")
	ErrNoSessionInContext = errors.New("no session data in context")
)
