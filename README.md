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

## Sequence Diagram(s)

```mermaid
sequenceDiagram
    participant C as Client
    participant P as Proxy Server
    participant M as Metrics Middleware
    participant T as OTel Transport
    participant U as Upstream Server

    C->>P: HTTP Request
    P->>M: Wrap request with WithMetrics
    M->>T: Forward request using otelhttp.NewTransport
    T->>U: Send request to Upstream Server
    U-->>T: Return response
    T-->>M: Pass response data with timing/status
    M-->>P: Forward response with metrics logged
    P-->>C: HTTP Response
```
