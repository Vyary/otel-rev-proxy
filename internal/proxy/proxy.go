package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/Vyary/otel-rev-proxy/internal/models"
	"github.com/Vyary/otel-rev-proxy/pkg/telemetry"
)

type proxies map[string]http.Handler

type proxyServer struct {
	config  models.Config
	proxies proxies
}

func NewProxy(config *models.Config) (*proxyServer, error) {
	proxyServer := &proxyServer{
		config: *config,
	}

	err := proxyServer.createProxies()
	if err != nil {
		return nil, err
	}

	return proxyServer, nil
}

func (p *proxyServer) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, exists := p.config.Routes[r.Host]
		if !exists {
			http.Error(w, "Host Not Found", http.StatusNotFound)
			return
		}

		proxy, exists := p.proxies[r.Host]
		if !exists {
			http.Error(w, "Proxy is not configured", http.StatusInternalServerError)
			return
		}

		if p.config.BlockAllRequests && !p.isPathAllowed(r.URL.Path, route.AllowedPaths) {
			http.Error(w, "Request blocked by policy", http.StatusForbidden)
			return

		}

		proxy.ServeHTTP(w, r)
	})
}

func (p *proxyServer) isPathAllowed(path string, paths []string) bool {
	for _, pattern := range paths {
		if pattern == "/*" {
			return true
		}

		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "/*")

			if strings.HasPrefix(path, prefix) {
				return true
			}
		}

		if path == pattern {
			return true
		}
	}

	return false
}

func (p *proxyServer) createProxies() error {
	p.proxies = make(proxies, len(p.config.Routes))

	for host, route := range p.config.Routes {
		target, err := url.Parse(route.URL)
		if err != nil {
			return fmt.Errorf("Ivalid URL for host %s %v", host, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(target)

		if route.Otel {
			otelHandler := telemetry.WithTraces(proxy)
			otelHandler = telemetry.WithMetrics(otelHandler)

			p.proxies[host] = otelHandler
			continue
		}

		p.proxies[host] = proxy
	}

	return nil
}
