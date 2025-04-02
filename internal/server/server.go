package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Vyary/otel-rev-proxy/internal/models"
	"github.com/Vyary/otel-rev-proxy/internal/proxy"
	"gopkg.in/yaml.v3"
)

type ServerOptions struct {
	ConfigPath string
	Port       string
}

func New() (*http.Server, error) {
	return NewWithOptions(DefaultOptions())
}

func NewWithOptions(opts ServerOptions) (*http.Server, error) {
	config, err := loadConfig(opts.ConfigPath)
	if err != nil {
		return nil, err
	}

	proxy, err := proxy.NewProxy(config)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", opts.Port),
		Handler: proxy.Handler(),
	}, nil
}

func DefaultOptions() ServerOptions {
	configPath := os.Getenv("PROXY_CONFIG_PATH")
	if configPath == "" {
		configPath = "/etc/configs/otel-rev-proxy/routes.yaml"
	}

	return ServerOptions{
		ConfigPath: configPath,
		Port:       "443",
	}
}

func loadConfig(filename string) (*models.Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config models.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
