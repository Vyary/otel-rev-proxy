package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/Vyary/otel-rev-proxy/pkg/telemetry"
	"gopkg.in/yaml.v3"
)

type Route struct {
	URL  string `yaml:"url"`
	Otel bool   `yaml:"otel"`
}

type Config struct {
	Routes map[string]Route `yaml:"routes"`
}

func New(port string) (*http.Server, error) {
	configPath := os.Getenv("PROXY_CONFIG_PATH")
	if configPath == "" {
		configPath = "/etc/configs/otel-rev-proxy/routes.yaml"
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	proxies := make(map[string]http.Handler)

	for host, route := range config.Routes {
		target, err := url.Parse(route.URL)
		if err != nil {
			return nil, fmt.Errorf("Ivalid URL for host %s %v", host, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(target)

		if route.Otel {
			otelHandler := telemetry.WithTraces(proxy)
			otelHandler = telemetry.WithMetrics(otelHandler)

			proxies[host] = otelHandler
			continue
		}

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
		Addr:    fmt.Sprintf(":%s", port),
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
