package middleware

import (
	"context"
	"net/http"

	uuid "github.com/google/uuid"
)

type ctxKey int

const ridKey ctxKey = ctxKey(0)

// Retrive unique request id from the provided context object.
func GetRequestID(ctx context.Context) string {
	val := ctx.Value(ridKey)
	if val == nil {
		return ""
	}
	return val.(string)
}

// RequestId middleware creates a unique and random request id for each
// http request, and thats that into request context, so the same can be
// used by the application as an unique identifier.
// This same request id is also added to the http response header so the
// similar correlation can be created at client side as well.
func RequestId(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")

		if rid == "" {
			rid = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), ridKey, rid)
		w.Header().Add("X-Request-ID", rid)

		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}
