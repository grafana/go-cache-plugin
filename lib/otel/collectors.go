package otel

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

var tp *trace.TracerProvider

func SetupLoggingProvider(ctx context.Context, file string) (func(context.Context) error, error) {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	exporter, err := stdouttrace.New(stdouttrace.WithWriter(f))
	if err != nil {
		return nil, err
	}
	shutdownHook := setupTraceProvider(exporter)
	return shutdownHook, nil
}

func SetupOtelTraceProvider(ctx context.Context) (func(context.Context) error, error) {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" && os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") == "" {
		return nil, fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT or OTEL_EXPORTER_OTLP_TRACES_ENDPOINT env variable is not set")
	}

	auth := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")
	sep := strings.Index(auth, "=")
	name, value := auth[:sep], auth[sep+1:]

	println(name, value[:5])
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithHeaders(map[string]string{name: value}),
	)

	if err != nil {
		return nil, err
	}

	shutdownHook := setupTraceProvider(exporter)

	return shutdownHook, nil
}

func setupTraceProvider(exporter trace.SpanExporter) func(ctx context.Context) error {
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		attribute.String("env", "ci"),
	)

	tp = trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	return tp.Shutdown
}
