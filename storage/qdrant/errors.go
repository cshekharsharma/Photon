package qdrant

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrInvalidPointID = errors.New("qdrant: invalid point id")

	ErrUnauthorized = errors.New("qdrant: unauthorized")
	ErrForbidden    = errors.New("qdrant: forbidden")
	ErrNotFound     = errors.New("qdrant: not found")
	ErrBadRequest   = errors.New("qdrant: bad request")
	ErrConflict     = errors.New("qdrant: conflict")
	ErrRateLimited  = errors.New("qdrant: rate limited")
	ErrServer       = errors.New("qdrant: server error")
)

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("qdrant: http %d: %s", e.StatusCode, e.Body)
}

func mapHTTPStatus(code int) error {
	switch code {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusBadRequest:
		return ErrBadRequest
	case http.StatusConflict:
		return ErrConflict
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		if code >= 500 {
			return ErrServer
		}
		return nil
	}
}
