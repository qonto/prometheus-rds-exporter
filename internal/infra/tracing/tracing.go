// Package tracing provides OpenTelemetry trace
package tracing

import (
	"context"
	"fmt"
	"time"

	"github.com/qonto/prometheus-rds-exporter/internal/infra/build"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

// SetupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context) (*trace.TracerProvider, error) {
	tracerProvider, err := newTraceProvider()
	if err != nil {
		return tracerProvider, err
	}

	otel.SetTracerProvider(tracerProvider)

	return tracerProvider, err
}

func newTraceProvider() (*trace.TracerProvider, error) {
	grpcExporter, err := otlptracegrpc.New(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("could not initialized gPRC exporter: %w", err)
	}

	resources, err := resource.New(context.Background(),
		resource.WithAttributes(semconv.ServiceName("prometheus-rds-exporter"), semconv.ServiceVersion(build.Version)),
		resource.WithFromEnv(),   // pull attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables
		resource.WithOS(),        // This option configures a set of Detectors that discover OS information
		resource.WithContainer(), // This option configures a set of Detectors that discover container information
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialized otel resources: %w", err)
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithResource(resources),
		trace.WithBatcher(grpcExporter, trace.WithBatchTimeout(time.Second)),
	)

	return traceProvider, nil
}
