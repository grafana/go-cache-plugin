package otel

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"log"
	"os"
	"time"
)

type Mode string

const (
	ModeCollector Mode = "collector"
	ModeStdout    Mode = "stdout"
)

type Config struct {
	Mode    Mode
	Address string // collector address
	LogFile string
}

var tp *trace.TracerProvider

func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {

	var exporter trace.SpanExporter
	var err error

	switch cfg.Mode {
	case ModeCollector:
		exporter, err = otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(cfg.Address),
			otlptracegrpc.WithInsecure(),
		)
	default:
		f, err := os.OpenFile(cfg.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		exporter, err = stdouttrace.New(stdouttrace.WithWriter(f))
	}

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

func main() {
	ctx := context.Background()

	shutdown, err := Init(ctx, Config{
		Mode:    ModeStdout,
		Address: "localhost:4317",
	})
	if err != nil {
		panic(err)
	}
	defer shutdown(ctx)

	meta := TraceMeta{
		RunId:      "github-run-id",
		RunAttempt: "github-run-attempt",
		JobName:    "github-job-name",
		StepName:   "github-step-name",
	}
	runCI(ctx, meta)
}

func runCI(ctx context.Context, meta TraceMeta) {

	ctx, span1 := meta.SpanWithContext(ctx, "DOWNLOAD",
		attribute.String("package", "pkg.foo"),
	)
	time.Sleep(50 * time.Millisecond)
	defer span1.End()

	_, span2 := meta.SpanWithContext(ctx, "DOWNLOAD",
		attribute.String("package", "pkg.foo"),
	)
	time.Sleep(50 * time.Millisecond)
	defer span2.End()

	_, span3 := meta.SpanWithContext(ctx, "DOWNLOAD",
		attribute.String("package", "pkg.foo"),
	)
	time.Sleep(50 * time.Millisecond)
	defer span3.End()
}
