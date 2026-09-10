package flashdb

import "errors"

var (
	ErrNotFound     = errors.New("flashdb: key not found")            // not in cache
	ErrBadType      = errors.New("flashdb: value type mismatch")      // type assertion failed
	ErrIllegalField = errors.New("flashdb: illegal or missing field") // e.g. empty key
	ErrInvalidKey   = errors.New("flashdb: invalid/empty key")        // e.g. malformed key
)
