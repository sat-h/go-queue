package tracing

import (
	"context"
	"os"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/sat-h/go-queue/internal/observability/logger"
)

var (
	tracer trace.Tracer
	once   sync.Once
)

// Init initializes the OpenTelemetry tracing system with Jaeger exporter
func Init(serviceName string) error {
	var err error

	once.Do(func() {
		// Default to localhost if not specified
		jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
		if jaegerEndpoint == "" {
			jaegerEndpoint = "http://localhost:14268/api/traces"
		}

		// Create Jaeger exporter
		exp, expErr := jaeger.New(
			jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)),
		)
		if expErr != nil {
			err = expErr
			logger.Error("Failed to create Jaeger exporter", zap.Error(expErr))
			return
		}

		// Create trace provider with Jaeger exporter
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exp),
			sdktrace.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
			)),
		)

		// Set global trace provider
		otel.SetTracerProvider(tp)

		// Create tracer
		tracer = tp.Tracer(serviceName)
		logger.Info("Tracing initialized", zap.String("service", serviceName))
	})

	return err
}

// GetTracer returns the global tracer instance
func GetTracer() trace.Tracer {
	if tracer == nil {
		// If tracer hasn't been initialized, initialize with a default service name
		err := Init("go-queue-default")
		if err != nil {
			logger.Error("Failed to initialize tracer with default service name", zap.Error(err))
		}
	}
	return tracer
}

// StartSpan starts a new span with the provided name and context
func StartSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return GetTracer().Start(ctx, spanName)
}

// Shutdown gracefully shuts down the tracer provider
func Shutdown(ctx context.Context) error {
	provider := otel.GetTracerProvider()
	if provider == nil {
		return nil
	}

	if shutdownProvider, ok := provider.(*sdktrace.TracerProvider); ok {
		return shutdownProvider.Shutdown(ctx)
	}

	return nil
}
