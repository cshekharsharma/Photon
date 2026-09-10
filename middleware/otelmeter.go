package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cshekharsharma/photon/telemetry"
	"github.com/cshekharsharma/photon/telemetry/meter"
)

const (
	MetricApiRequestsActive     = "%s_api_requests_active"
	MetricApiRequestsTotal      = "%s_api_requests_total"
	MetricApiRequestErrorsTotal = "%s_api_request_errors_total"
	MetricApiRequestLatencyMs   = "%s_api_request_latency_ms"
)

// OTelMeterMiddleware tracks total requests, errors, latency, and active request count.
func OTelMeterMiddleware(service string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ctx := r.Context()

			resp := NewWrapResponseWriter(w)
			path := normalizePath(r.URL.Path)

			attr := map[string]interface{}{
				telemetry.AttributeHttpMethod: r.Method,
				telemetry.AttributeHttpPath:   path,
			}

			_ = meter.IncrementGaugeFunc(fmt.Sprintf(MetricApiRequestsActive, service), attr)

			defer func() {
				attr[telemetry.AttributeHttpStatusCode] = resp.Status()
				attr[telemetry.AttributeHttpStatusClass] = fmt.Sprintf("%dxx", resp.Status()/100)

				if rec := recover(); rec != nil {
					attr[telemetry.AttributeExceptionMessage] = fmt.Sprintf("%v", rec)

					meter.RecordLatencyFunc(t(MetricApiRequestLatencyMs, service), start, attr)
					meter.RecordCounterFunc(t(MetricApiRequestsTotal, service), 1, attr)
					meter.RecordCounterFunc(t(MetricApiRequestErrorsTotal, service), 1, attr)
					meter.DecrementGaugeFunc(t(MetricApiRequestsActive, service))

					panic(rec) // re-throw
				}

				meter.RecordLatencyFunc(t(MetricApiRequestLatencyMs, service), start, attr)
				meter.RecordCounterFunc(t(MetricApiRequestsTotal, service), 1, attr)

				if resp.Status() >= http.StatusBadRequest { // Error counter for 4xx and 5xx
					meter.RecordCounterFunc(t(MetricApiRequestErrorsTotal, service), 1, attr)
				}

				meter.DecrementGaugeFunc(t(MetricApiRequestsActive, service))
			}()

			next.ServeHTTP(resp, r.WithContext(ctx))
		})
	}
}

// normalizePath converts path to a label-safe format
func normalizePath(path string) string {
	segments := strings.Split(path, "/")
	for i, seg := range segments {
		if _, err := strconv.Atoi(seg); err == nil {
			segments[i] = ":id"
		}
	}
	return strings.Join(segments, "/")
}

// Convert the arguments to a slice of interface{}
// and pass them to fmt.Sprintf
func t(placeholder string, args ...interface{}) string {
	return fmt.Sprintf(placeholder, args...)
}
