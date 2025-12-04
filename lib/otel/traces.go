package otel

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"strings"
)

type TraceMeta struct {
	RunId      string
	RunAttempt string
	JobName    string
	StepName   string
}

func NewTracedFromString(traceParams string) *TraceMeta {
	parts := strings.Split(traceParams, ":")
	runId, runAttempt, jobName, stepName := parts[0], parts[1], parts[2], parts[3]

	return NewTracer(runId, runAttempt, jobName, stepName)
}

func NewTracer(runId, runAttempt, jobName, stepName string) *TraceMeta {
	return &TraceMeta{RunId: runId, RunAttempt: runAttempt, JobName: jobName, StepName: stepName}
}

func (t *TraceMeta) SpanWithContext(context context.Context, name string, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	traceId, _ := trace.TraceIDFromHex(t.GenerateTraceID())
	parentSpan, _ := trace.SpanIDFromHex(t.GenerateStepSpanID())

	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceId,
		SpanID:     parentSpan,
		TraceFlags: trace.FlagsSampled,
	})

	ctx := trace.ContextWithSpanContext(context, spanContext)
	tracer := otel.Tracer("github.com/grafana/go-cache-plugin")

	start, span := tracer.Start(ctx, name)
	if len(attributes) > 0 {
		span.SetAttributes(attributes...)
	}

	return start, span
}

func (t *TraceMeta) GenerateTraceID() string {
	return GenerateTraceID(t.RunId, t.RunAttempt)
}

func (t *TraceMeta) GenerateJobSpanID() string {
	return GenerateJobSpanID(t.RunId, t.RunAttempt, t.JobName)
}

func (t *TraceMeta) GenerateStepSpanID() string {
	return GenerateStepSpanID(t.RunId, t.RunAttempt, t.JobName, t.StepName)
}

func GenerateTraceID(runID, runAttempt string) string {
	input := fmt.Sprintf("%s%st", runID, runAttempt)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[:32]
}

func GenerateJobSpanID(runID, runAttempt, jobName string) string {
	input := fmt.Sprintf("%s%s%s", runID, runAttempt, jobName)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[:16]
}

func GenerateStepSpanID(runID, runAttempt, jobName, stepName string) string {
	input := fmt.Sprintf("%s%s%s%s", runID, runAttempt, jobName, stepName)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[:16]
}
