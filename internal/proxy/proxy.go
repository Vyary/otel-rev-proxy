package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/Vyary/otel-rev-proxy/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func New() (*http.Server, error) {
	proxyTarget := os.Getenv("PROXY_TARGET")
	target, err := url.Parse(proxyTarget)
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
		Addr:    ":80",
		Handler: handler,
	}, nil
}
