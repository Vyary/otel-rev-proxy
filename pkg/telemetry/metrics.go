package telemetry

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	logger = otelslog.NewLogger(service)

	meter                  = otel.Meter(service)
	proxyUptime            metric.Float64ObservableCounter
	requestCounter         metric.Int64Counter
	requestDuration        metric.Float64Histogram
	requestDurationBuckets = []float64{
		0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10,
	}
)

func WithMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		rw := &ResponseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		requestCounter.Add(r.Context(), 1,
			metric.WithAttributes(
				attribute.Key("method").String(r.Method),
				attribute.Key("route").String(r.URL.Path),
				attribute.Key("code").Int(rw.statusCode)))

		duration := time.Since(startTime).Seconds()
		requestDuration.Record(r.Context(), duration,
			metric.WithAttributes(
				attribute.Key("method").String(r.Method),
				attribute.Key("route").String(r.URL.Path),
				attribute.Key("code").Int(rw.statusCode)))
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
		logger.Error("failed to create an uptime metric", "error", err)
		panic(err)
	}

	requestCounter, err = meter.Int64Counter(
		"requests_total",
		metric.WithDescription("Total number of requests handled by the reverse proxy."),
	)
	if err != nil {
		logger.Error("failed to create an requests total metric", "error", err)
		panic(err)
	}

	requestDuration, err = meter.Float64Histogram(
		"request_duration_seconds",
		metric.WithDescription("Duration of requests handled by the reverse proxy in seconds."),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(requestDurationBuckets...),
	)
	if err != nil {
		logger.Error("failed to create request duration metric")
		panic(err)
	}
}
