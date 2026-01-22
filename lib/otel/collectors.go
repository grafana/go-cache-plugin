package otel

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

var tp *trace.TracerProvider

func SetupLoggingProvider(ctx context.Context, service, file string) (func(context.Context) error, error) {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	exporter, err := stdouttrace.New(stdouttrace.WithWriter(f))
	if err != nil {
		return nil, err
	}
	shutdownHook := setupTraceProvider(service, exporter)
	return shutdownHook, nil
}

func SetupOtelTraceProvider(ctx context.Context, service, address string) (func(context.Context) error, error) {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(address),
		//otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithHeaders(map[string]string{"Authorization": fmt.Sprintf("Basic %s", os.Getenv("OTEL_EXPORTER_OTLP_HEADERS_API_KEY"))}),
	)

	if err != nil {
		return nil, err
	}

	shutdownHook := setupTraceProvider(service, exporter)

	return shutdownHook, nil
}

func setupTraceProvider(service string, exporter trace.SpanExporter) func(ctx context.Context) error {
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(service),
		attribute.String("env", "ci"),
	)

	tp = trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	return tp.Shutdown
}
