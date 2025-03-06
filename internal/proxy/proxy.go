package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/Vyary/otel-rev-proxy/internal/middleware"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

const name = "reverse-proxy"

var (
	meter  = otel.Meter(name)
	logger = otelslog.NewLogger(name)
)

func New() (*http.Server, error) {
	target, err := url.Parse("http://localhost:8080")
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Transport = &tracingTransport{
		base: http.DefaultTransport,
	}

	handler := middleware.WithTracing(proxy)

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
