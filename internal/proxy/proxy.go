package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// New creates and returns an HTTP server configured as a reverse proxy to http://localhost:8080 with OpenTelemetry tracing.
//
// It sets up a reverse proxy using httputil.NewSingleHostReverseProxy and customizes its transport with tracingTransport to integrate tracing via OpenTelemetry.
// The proxy is wrapped with tracing middleware, and the resulting server listens on port 7000.
// An error is returned if the target URL cannot be parsed.
func New() (*http.Server, error) {
	target, err := url.Parse("http://localhost:8080")
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Transport = &tracingTransport{
		base: http.DefaultTransport,
	}

	handler := otelhttp.NewHandler(proxy, "proxy_request",
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		}),
	)

	return &http.Server{
		Addr:    ":7000",
		Handler: handler,
	}, nil
}

type tracingTransport struct {
	base http.RoundTripper
}

func (t *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return otelhttp.NewTransport(t.base).RoundTrip(req)
}
