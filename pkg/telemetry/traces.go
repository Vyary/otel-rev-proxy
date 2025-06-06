package telemetry

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer(service)

func WithTraces(next *httputil.ReverseProxy) http.Handler {
	next.Transport = otelhttp.NewTransport(next.Transport)

	if next.Transport == nil {
		next.Transport = otelhttp.NewTransport(http.DefaultTransport)
	}

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

		originalDirector := next.Director
		next.Director = func(req *http.Request) {
			originalDirector(req)
			otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
		}

		rw := &ResponseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r.WithContext(ctx))

		span.SetAttributes(attribute.Int("http.status_code", rw.statusCode))

		if rw.statusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", rw.statusCode))
			return
		}

		span.SetStatus(codes.Ok, "successfully completed the request")
	})
}
