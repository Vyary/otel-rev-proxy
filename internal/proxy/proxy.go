package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/Vyary/otel-rev-proxy/internal/models"
	"github.com/Vyary/otel-rev-proxy/pkg/telemetry"
)

type proxies map[string]http.Handler

type proxyServer struct {
	config  models.Config
	proxies proxies
}

type sseRoundTripper struct {
	http.RoundTripper
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

		proxy.Transport = &sseRoundTripper{
			RoundTripper: &http.Transport{
				MaxIdleConns:       100,
				IdleConnTimeout:    90 * time.Second,
				DisableCompression: true,
				DisableKeepAlives:  false,
			},
		}

		proxy.ModifyResponse = func(r *http.Response) error {
			if r.Header.Get("Content-Type") == "text/event-stream" {
				r.Header.Del("Content-Length")
			}

			return nil
		}

		originalDirector := proxy.Director
		proxy.Director = func(r *http.Request) {
			originalDirector(r)

			if r.Header.Get("Accept") == "text/event-stream" {
				r.Header.Set("Connection", "keep-alive")
				r.Header.Set("Cache-Control", "no-cache")
			}
		}

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

func (s *sseRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := s.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// For SSE responses, ensure no buffering
	if resp.Header.Get("Content-Type") == "text/event-stream" {
		resp.Header.Del("Content-Length")
		resp.ContentLength = -1
	}

	return resp, nil
}
