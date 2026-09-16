package middleware

import (
	"fmt"
	"net/http"

	"github.com/cshekharsharma/photon/core/logger"
	"github.com/cshekharsharma/photon/utils/rest"
	"github.com/cshekharsharma/photon/utils/rest/apiresponse"
	"github.com/go-stack/stack"
)

// Recoverer middleware catches the panic that have occurred during program
// execution, and recovers from them. It logs the panic details and stack trace
// into log files, and also sends an error http response to the client if it is
// in middle of a web request.
func Recoverer(logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if p := recover(); p != nil {
					stackTrace := make([]string, 0)
					// Get the current stacktrace but trim the runtime
					traces := stack.Trace().TrimRuntime()

					// Format the stack trace removing the clutter from it
					for i := 0; i < len(traces); i++ {
						t := traces[i]
						tFunc := t.Frame().Function

						// This call is made before the code reaching our handlers,
						// we don't want to log things that are coming before
						// our own code, just from our handlers and downwards.
						if tFunc == "net/http.HandlerFunc.ServeHTTP" {
							break
						}

						stackTrace = append(stackTrace, fmt.Sprintf("%+v", t))
					}

					fields := map[string]interface{}{
						"request_id":  GetRequestID(r.Context()),
						"http_method": r.Method,
						"http_path":   r.URL.Path,
						"http_status": http.StatusInternalServerError,
						"module":      "http",
						"operation":   "middleware.recoverer",
						"error":       fmt.Sprintf("panic: %v", p),
						"stack":       stackTrace,
					}

					logger.ErrorWithFields(fields, "panic recovered")

					// Send JSON response to API clients only if it is a web request.
					if rest.IsHttpRequest(r) {
						errMsg := "Error: An internal server error occurred."
						apiresponse.New(false, apiresponse.InternalServerError, new(any), errMsg).Send(w, http.StatusInternalServerError)
						return
					}
				}

			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
