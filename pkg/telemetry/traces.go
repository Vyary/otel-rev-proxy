package telemetry

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("reverse-proxy")

func WithTraces(next *httputil.ReverseProxy) http.Handler {
	next.Transport = otelhttp.NewTransport(&http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
		ResponseHeaderTimeout: 15 * time.Second,
		ExpectContinueTimeout: 2 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          500,
		MaxIdleConnsPerHost:   100,
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		ctx, span := tracer.Start(r.Context(), spanName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.String("http.host", r.Host),
			attribute.String("http.user_agent", r.UserAgent()),
		)

		rw := &ResponseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r.WithContext(ctx))

		span.SetAttributes(attribute.Int("http.status_code", rw.statusCode))

		if rw.statusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", rw.statusCode))
		}
	})
}
