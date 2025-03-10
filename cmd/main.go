package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Vyary/otel-rev-proxy/internal/proxy"
	"github.com/Vyary/otel-rev-proxy/internal/telemetry"
	"go.opentelemetry.io/contrib/bridges/otelslog"
)

var logger = otelslog.NewLogger("reverse-proxy")

func main() {
	if err := run(); err != nil {
		logger.Error("failed to start the reverse proxy", "error", err)
		slog.Error("failed to start the reverse proxy", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
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

	slog.Info("Shutting Down Reverse Proxy...")

	return srv.Shutdown(ctxTO)
}
