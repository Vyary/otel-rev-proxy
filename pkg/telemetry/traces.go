package telemetry

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func WithTraces(next *httputil.ReverseProxy) http.Handler {
	next.Transport = otelhttp.NewTransport(&http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
		ResponseHeaderTimeout: 15 * time.Second,
		ExpectContinueTimeout: 2 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		IdleConnTimeout:       90 * time.Second,
	})

	return otelhttp.NewHandler(next, "reverse_proxy",
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		}),
	)
}
