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

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

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
