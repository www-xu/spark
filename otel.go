package spark

import (
	"context"
	"os"

	"github.com/www-xu/spark/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTracer sets up the OpenTelemetry SDK. It accepts optional custom exporters.
// If no exporters are provided, it defaults to a console exporter writing to os.Stderr.
// It configures and sets the global tracer provider and propagator.
func InitTracer(serviceName string, exporters ...sdktrace.SpanExporter) (func(), error) {
	var usedExporters []sdktrace.SpanExporter
	if len(exporters) > 0 {
		usedExporters = exporters
	} else {
		// Default to a stderr exporter if none are provided.
		defaultExporter, err := stdouttrace.New(stdouttrace.WithWriter(os.Stderr))
		if err != nil {
			return nil, err
		}
		usedExporters = []sdktrace.SpanExporter{defaultExporter}
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create a slice of options for the TracerProvider.
	// We start with the resource...
	tpOptions := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
	}

	// ...and add a batcher for each exporter.
	for _, exporter := range usedExporters {
		tpOptions = append(tpOptions, sdktrace.WithBatcher(exporter))
	}

	tp := sdktrace.NewTracerProvider(tpOptions...)

	// Set the global TracerProvider.
	otel.SetTracerProvider(tp)

	// Set the global TextMapPropagator to use the W3C Trace Context format.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	shutdown := func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Error(context.Background(), "error shutting down tracer provider: ", err)
		}
	}
	return shutdown, nil
}
