package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/cshekharsharma/photon/utils/rest"
)

// Context key type for locale
type ContextLocaleKey string

const LocaleKey ContextLocaleKey = "locale"

func parseForm(r *http.Request) {
	if err := r.ParseForm(); err != nil {
		return
	}
}

// Get locale value from request context
func GetLocale(ctx context.Context) string {
	return ctx.Value(LocaleKey).(string)
}

// Set locale value to the request context
func SetLocale(r *http.Request, locale string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), LocaleKey, locale))
}

// Application request initialiser middleware. All the init actions should be done
// here that are to be performed before the request execution starts.
//
// ParseForm reads r.Body for form-encoded POSTs which would leave downstream
// handlers with an empty body. Buffer the body first, reset it after ParseForm,
// so handlers that need the raw payload (e.g. webhook signature verification)
// can still read it.
func RequestInit(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		if r.Body != nil {
			buf, err := io.ReadAll(r.Body)
			if err == nil {
				r.Body = io.NopCloser(bytes.NewReader(buf))
				parseForm(r)
				r.Body = io.NopCloser(bytes.NewReader(buf))
			}
		} else {
			parseForm(r)
		}

		locale := r.Header.Get(rest.HeaderContentLanguage)

		if len(locale) == 5 {
			reqCtx = context.WithValue(reqCtx, LocaleKey, locale)
		} else {
			defaultLocale := ""
			reqCtx = context.WithValue(reqCtx, LocaleKey, defaultLocale)
		}

		next.ServeHTTP(w, r.WithContext(reqCtx))
	}

	return http.HandlerFunc(fn)
}
