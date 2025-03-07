package telemetry

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// SetupOTelSDK initializes the OpenTelemetry SDK for tracing, metrics, and logging by configuring the text map propagator, trace provider, meter provider, and logger provider. It returns a shutdown function that aggregates cleanup routines for all configured components; the shutdown function must be called to release resources when the SDK is no longer needed.
//
// The provided context is used for managing deadlines and cancellation during both initialization and cleanup. If any initialization step fails, the function performs cleanup and returns the encountered error.
func SetupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTracerProvider()
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider()
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider()
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return
}

// newPropagator creates a composite text map propagator that supports both TraceContext and Baggage propagation.
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// newTracerProvider creates and configures a new tracer provider that batches and exports trace data via a gRPC exporter over an insecure connection, attaching a resource with the service name "reverse-proxy". It panics if the exporter cannot be initialized and returns an error if resource creation fails.
func newTracerProvider() (*trace.TracerProvider, error) {
	ctx := context.Background()
	exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("reverse-proxy"),
		),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(trace.WithBatcher(exp), trace.WithResource(res))
	return tracerProvider, nil
}

// newMeterProvider creates a MeterProvider that exports metrics via an OTLP gRPC exporter using a periodic reader. It associates the provider with a resource labeled using the service name "reverse-proxy". The function panics if the OTLP metric exporter initialization fails and returns an error if creating the resource does.
func newMeterProvider() (*metric.MeterProvider, error) {
	ctx := context.Background()
	exp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create meter exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("reverse-proxy"),
		),
	)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exp)), metric.WithResource(res))
	return meterProvider, nil
}

// newLoggerProvider creates a new OTLP logger provider configured with a batch processor.
// It initializes an OTLP gRPC log exporter using an insecure connection and sets up a batch processor
// to handle log entries. The function panics if exporter initialization fails.
func newLoggerProvider() (*log.LoggerProvider, error) {
	ctx := context.Background()
	exp, err := otlploggrpc.New(ctx, otlploggrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}

	processor := log.NewBatchProcessor(exp)
	loggerProvider := log.NewLoggerProvider(log.WithProcessor(processor))
	return loggerProvider, nil
}
