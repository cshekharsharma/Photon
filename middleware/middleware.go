package middleware

import "net/http"

// Conscious call to build over net/http package to support public middleware
// packages built over and above net/http package
type MiddlewareFn func(http.HandlerFunc) http.HandlerFunc

// Convert `func(http.Handler) http.Handler` to MiddlewareFn
func ToMiddlewareFn(mw func(http.Handler) http.Handler) MiddlewareFn {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			adaptedHandler := mw(http.HandlerFunc(next))
			adaptedHandler.ServeHTTP(w, r)
		}
	}
}
