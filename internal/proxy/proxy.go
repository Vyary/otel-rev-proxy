package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/Vyary/otel-rev-proxy/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func New() (*http.Server, error) {
	target, err := url.Parse("http://localhost:8080")
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = otelhttp.NewTransport(http.DefaultTransport)

	handler := otelhttp.NewHandler(proxy, "reverse_proxy",
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		}),
	)

	handler = telemetry.WithMetrics(handler)

	return &http.Server{
		Addr:    ":7000",
		Handler: handler,
	}, nil
}
