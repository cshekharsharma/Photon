package middleware

import (
	"net/http"
	"path"

	"github.com/cshekharsharma/photon/utils/rest"
	"github.com/cshekharsharma/photon/utils/rest/minifier"
)

// ApiMinifier is a middleware that sets a specific header to the response
// if the request URL path is in the provided map and the corresponding value is true.
//
// Parameters:
//   - httproutes: A map where the keys are URL paths and the values indicate
//     whether the response for that path should be minified.
//
// Returns:
//
//	A middleware function that can be used with an HTTP handler.
func ApiMinifier(httproutes map[string]bool) func(http.Handler) http.Handler {
	// Defensive copy to avoid races if caller mutates the map later.
	routes := make(map[string]bool, len(httproutes))
	for p, v := range httproutes {
		routes[path.Clean(p)] = v
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only signal if the feature is enabled.
			if minifier.IsApiKeyMinifierEnabled() {
				if shouldMinify := routes[path.Clean(r.URL.Path)]; shouldMinify {
					w.Header().Set(rest.HeaderXApiMinifier, rest.XAPIMinifierValue)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
