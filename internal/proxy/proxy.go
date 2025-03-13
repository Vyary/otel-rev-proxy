package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/Vyary/otel-rev-proxy/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"gopkg.in/yaml.v3"
)

var PORT = os.Getenv("PORT")

type Config struct {
	Routes       map[string]string `yaml:"routes"`
	RoutesNoOtel map[string]string `yaml:"routes_no_otel"`
}

func New() (*http.Server, error) {
	configPath := os.Getenv("PROXY_CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/routes.yaml"
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	proxies := make(map[string]http.Handler)

	for host, target := range config.Routes {
		t, err := url.Parse(target)
		if err != nil {
			return nil, fmt.Errorf("Ivalid URL for host %s %v", host, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(t)
		proxy.Transport = otelhttp.NewTransport(http.DefaultTransport)

		otelHandler := otelhttp.NewHandler(proxy, "reverse_proxy",
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			}),
		)

		otelHandler = telemetry.WithMetrics(otelHandler)

		proxies[host] = otelHandler
	}

	for host, targetURL := range config.RoutesNoOtel {
		target, err := url.Parse(targetURL)
		if err != nil {
			return nil, fmt.Errorf("invalid URL for host %s: %v", host, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(target)

		proxies[host] = proxy
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy, exists := proxies[r.Host]
		if !exists {
			http.Error(w, "Host not found", http.StatusNotFound)
		}

		proxy.ServeHTTP(w, r)
	})

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", PORT),
		Handler: handler,
	}, nil
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
