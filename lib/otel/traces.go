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

func NewTracingContext(traceId, parentSpanId string) (*TracingContext, error) {
	traceIDFromHex, err := trace.TraceIDFromHex(traceId)
	if err != nil {
		return nil, err
	}
	spanIDFromHex, err := trace.SpanIDFromHex(parentSpanId)
	if err != nil {
		return nil, err
	}

	return &TracingContext{
		TraceID:      traceIDFromHex,
		ParentSpanID: spanIDFromHex,
	}, nil
}

func NewTracingContextFromRunData(runId, runAttempt, jobName, stepName, stepNumber string) *TracingContext {
	traceId, _ := trace.TraceIDFromHex(GenerateTraceID(runId, runAttempt))
	spanId, _ := trace.SpanIDFromHex(GenerateStepSpanID_Number(runId, runAttempt, jobName, stepNumber))

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

func GenerateTraceID(runID, runAttempt string) string {
	input := fmt.Sprintf("%s%st", runID, runAttempt)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[:32]
}

func GenerateParentSpanID(runID, runAttempt string) string {
	input := fmt.Sprintf("%s%ss", runID, runAttempt)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[16:32]
}

func GenerateJobSpanID(runID, runAttempt, jobName string) string {
	input := fmt.Sprintf("%s%s%s", runID, runAttempt, jobName)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[16:32]
}

func GenerateStepSpanID(runID, runAttempt, jobName, stepName string) string {
	input := fmt.Sprintf("%s%s%s%s", runID, runAttempt, jobName, stepName)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[16:32]
}

func GenerateStepSpanID_Number(runID, runAttempt, jobName, stepNumber string) string {
	input := fmt.Sprintf("%s%s%s%s", runID, runAttempt, jobName, stepNumber)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[16:32]
}
