package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/Vyary/otel-rev-proxy/internal/proxy"
	"github.com/Vyary/otel-rev-proxy/internal/telemetry"
)

// main is the application's entry point. It calls run to start the reverse proxy server and terminates the program with a fatal error if run returns an error.
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run initializes signal handling, sets up the OpenTelemetry SDK, and starts the reverse proxy server. It listens for OS interrupt signals or server errors and, upon an interrupt, initiates a graceful shutdown of the proxy using a 20-second timeout. It returns an error if any initialization step fails or if the shutdown process encounters an error.
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	otelShutdown, err := telemetry.SetupOTelSDK(ctx)
	if err != nil {
		return err
	}
	defer otelShutdown(context.Background())

	srv, err := proxy.New()
	if err != nil {
		return err
	}

	srvErr := make(chan error, 1)
	go func() {
		slog.Info("Starting Reverse Proxy...")
		srvErr <- srv.ListenAndServe()
	}()

	select {
	case err = <-srvErr:
		return err
	case <-ctx.Done():
		stop()
	}

	ctxTO, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return srv.Shutdown(ctxTO)
}
