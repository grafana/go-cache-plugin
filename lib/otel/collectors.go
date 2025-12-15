package otel

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"log"
	"os"
)

var tp *trace.TracerProvider

func SetupLoggingProvider(ctx context.Context, file string) (func(context.Context) error, error) {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	exporter, err := stdouttrace.New(stdouttrace.WithWriter(f))

	tp = trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.Default()),
	)

	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

func SetupOtelTraceProvider(ctx context.Context, address string) (func(context.Context) error, error) {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(address),
		otlptracegrpc.WithInsecure(),
	)

	if err != nil {
		return nil, err
	}

	tp = trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.Default()),
	)

	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}
