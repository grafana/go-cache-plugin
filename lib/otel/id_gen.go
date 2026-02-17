package otel

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type TracingContext struct {
	TraceID      trace.TraceID
	ParentSpanID trace.SpanID
}

func NewTracingContextFromRunData(repo, runId, runAttempt, jobName, stepName string) *TracingContext {
	traceId, _ := trace.TraceIDFromHex(GenerateTraceID(repo, runId, runAttempt))
	spanId, _ := trace.SpanIDFromHex(GenerateStepSpanID(repo, runId, runAttempt, jobName, stepName))
	return &TracingContext{
		TraceID:      traceId,
		ParentSpanID: spanId,
	}
}

func (t *TracingContext) SpanWithContext(context context.Context, name string, attributes ...attribute.KeyValue) (context.Context, trace.Span) {

	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    t.TraceID,
		SpanID:     t.ParentSpanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})

	ctx := trace.ContextWithSpanContext(context, spanContext)
	tracer := otel.Tracer("github.com/grafana/go-cache-plugin")

	start, span := tracer.Start(ctx, name)
	if len(attributes) > 0 {
		span.SetAttributes(attributes...)
	}

	return start, span
}

func GenerateTraceID(repo, runID, runAttempt string) string {
	input := fmt.Sprintf("%s%s%st", repo, runID, runAttempt)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[:32]
}

func GenerateParentSpanID(repo, runID, runAttempt string) string {
	input := fmt.Sprintf("%s%s%ss", repo, runID, runAttempt)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[16:32]
}

func GenerateJobSpanID(repo, runID, runAttempt, jobName string) string {
	input := fmt.Sprintf("%s%s%s%s", repo, runID, runAttempt, jobName)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[16:32]
}

func GenerateStepSpanID(repo, runID, runAttempt, jobName, stepName string) string {
	input := fmt.Sprintf("%s%s%s%s%s", repo, runID, runAttempt, jobName, stepName)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[16:32]
}

func GenerateStepSpanID_Number(repo, runID, runAttempt, jobName, stepNumber string) string {
	input := fmt.Sprintf("%s%s%s%s%s", repo, runID, runAttempt, jobName, stepNumber)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[16:32]
}
