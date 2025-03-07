package telemetry

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter           = otel.Meter("reverse-proxy")
	proxyUptime     metric.Float64ObservableCounter
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
)

func WithMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		ww := &statusCodeResponseWriter{ResponseWriter: w}

		next.ServeHTTP(ww, r)

		requestCounter.Add(r.Context(), 1, metric.WithAttributes(attribute.Key("method").String(r.Method)), metric.WithAttributes(attribute.Key("code").Int(ww.statusCode)))

		duration := time.Since(startTime).Seconds()
		requestDuration.Record(r.Context(), duration, metric.WithAttributes(attribute.Key("method").String(r.Method)), metric.WithAttributes(attribute.Key("code").Int(ww.statusCode)))
	})
}

func init() {
	var err error

	start := time.Now()

	proxyUptime, err = meter.Float64ObservableCounter(
		"uptime",
		metric.WithUnit("s"),
		metric.WithFloat64Callback(func(_ context.Context, o metric.Float64Observer) error {
			o.Observe(float64(time.Since(start).Seconds()))
			return nil
		}),
	)
	if err != nil {
		panic(err)
	}

	requestCounter, err = meter.Int64Counter(
		"requests_total",
		metric.WithDescription("Total number of requests handled by the reverse proxy."),
	)
	if err != nil {
		panic(err)
	}

	requestDuration, err = meter.Float64Histogram(
		"request_duration_seconds",
		metric.WithDescription("Duration of requests handled by the reverse proxy in seconds."),
		metric.WithUnit("s"),
	)
	if err != nil {
		panic(err)
	}
}

type statusCodeResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusCodeResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
