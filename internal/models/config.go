package models

type Route struct {
	URL          string   `yaml:"url"`
	Otel         bool     `yaml:"otel"`
	AllowedPaths []string `yaml:"allowed_paths"`
}

type Config struct {
	BlockAllRequests bool             `yaml:"block_all_requests"`
	Routes           map[string]Route `yaml:"routes"`
}
