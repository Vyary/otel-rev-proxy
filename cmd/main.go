package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Vyary/otel-rev-proxy/internal/server"
	"github.com/Vyary/otel-rev-proxy/pkg/telemetry"
	"go.opentelemetry.io/contrib/bridges/otelslog"
)

var (
	service = os.Getenv("SERVICE_NAME")
	logger  = otelslog.NewLogger(service)
)

func main() {
	if err := run(); err != nil {
		slog.Error("failed to start the reverse proxy", "error", err)
		os.Exit(1)
	}
}

func run() error {
	certFile := os.Getenv("CERT_FILE")
	keyFile := os.Getenv("KEY_FILE")

	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	otelShutdown, err := telemetry.SetupOTelSDK(ctx)
	if err != nil {
		return err
	}
	defer otelShutdown(context.Background())

	srv, err := server.New()
	if err != nil {
		return err
	}

	srvErr := make(chan error, 1)
	go func() {
		slog.Info("Starting Redirect Handler...")
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
		})
		srvErr <- http.ListenAndServe(":80", nil)
	}()

	go func() {
		slog.Info("Starting Secure Reverse Proxy...")
		srvErr <- srv.ListenAndServeTLS(certFile, keyFile)
	}()

	select {
	case err = <-srvErr:
		return err
	case <-ctx.Done():
		stop()
	}

	ctxTO, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	slog.Info("Shutting Down the Reverse Proxy...")

	return srv.Shutdown(ctxTO)
}
