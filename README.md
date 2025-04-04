# OpenTelemetry Reverse Proxy (otel-rev-proxy)

OpenTelemetry is a high-performance reverse proxy designed to seamlessly integrate with OpenTelemetry for distributed tracing, metrics, and logging. It acts as an intermediary between clients and backend services, enhancing observability by automatically capturing telemetry data from incoming and outgoing requests.

## Key Features

- **Built-in OpenTelemetry support** – Automatically collects traces, metrics, and logs.
- **Efficient request routing** – Routes and traffic to backend service.
- **Extensible architecture** – Easily integrates with existing observability stacks.
- **Minimal overhead** – Optimized for performance and low latency.
- **Security & reliability** – Supports TLS, rate limiting, and request validation.

## Use Cases

- Observability-first microservices infrastructure.
- Debugging and performance monitoring of distributed systems.
- Centralized telemetry collection for API gateways and service meshes.

## Configuration

The proxy is configured using a YAML file that specifies routing rules and OpenTelemetry settings.

### Configuration Structure

```go
// Route defines the configuration for a single host route
type Route struct {
	URL          string   `yaml:"url"`         // Target backend URL
	Otel         bool     `yaml:"otel"`        // Enable OpenTelemetry for this route
	AllowedPaths []string `yaml:"allowed_paths"` // List of paths allowed for this route if block all requests is enabled
}

// Config defines the overall proxy configuration
type Config struct {
	BlockAllRequests bool             `yaml:"block_all_requests"` // Block requests by default unless they match AllowedPaths
	Routes           map[string]Route `yaml:"routes"`            // Map of host to route configuration
}
```

### Example Configuration

```yaml
block_all_requests: true
routes:
  api.example.com:
    url: http://api-service:8080
    otel: true
    allowed_paths:
      - /v1/*
      - /health
  dashboard.example.com:
    url: http://dashboard:3000
    otel: false
    allowed_paths:
      - /*
```

### Path Matching

The proxy supports several types of path patterns in the `allowed_paths` section:

- **Exact match**: `/health`, `/api/v1/users`
- **Wildcard suffix**: `/api/*` matches `/api/v1`, `/api/users`, etc.
- **Global wildcard**: `/*` matches any path

### OpenTelemetry Integration

Set `otel: true` for any route where you want to enable OpenTelemetry instrumentation. This will automatically:

1. Create spans for all incoming requests
2. Propagate trace context to backend services
3. Add HTTP request/response metrics

### Environment Variables

- `PROXY_CONFIG_PATH`: Path to the configuration YAML file (default: `/etc/configs/otel-rev-proxy/routes.yaml`)

## Sequence Diagram(s)

```mermaid
sequenceDiagram
    participant Main as main.go
    participant Proxy as proxy.New(PORT)
    participant Config as Route Config Loader
    participant Tele as Telemetry Wrapper
    participant Backend as Backend Server
    Main->>Proxy: Call New(PORT)
    Proxy->>Config: Load YAML config & structured routes
    alt Route has Otel enabled
        Config->>Tele: Apply WithTraces wrapping
        Tele->>Proxy: Return instrumented handler
    end
    Proxy->>Backend: Forward incoming HTTP request
    Backend-->>Proxy: Return response
    Proxy-->>Main: Return server instance with routing and telemetry handling
```
