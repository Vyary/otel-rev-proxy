package telemetry

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func WithTraces(next *httputil.ReverseProxy) http.Handler {
	next.Transport = otelhttp.NewTransport(http.DefaultTransport)

	return otelhttp.NewHandler(next, "reverse_proxy",
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		}),
	)
}
