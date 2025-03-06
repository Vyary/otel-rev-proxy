package middleware

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter = otel.Meter("reverse-proxy")

	requestCounter   metric.Int64Counter
	requestDuration  metric.Float64Histogram
	errorCounter     metric.Int64Counter
	inFlightRequests metric.Int64UpDownCounter
)

// WithTracing wraps the provided HTTP handler with OpenTelemetry tracing instrumentation.
// It returns a new handler that assigns each request a span named using the HTTP method and URL path.
func WithTracing(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "proxy_request",
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		}),
	)
}
