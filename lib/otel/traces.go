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

type TracingContext struct {
	RunId      string
	RunAttempt string
	JobName    string
	StepName   string
	StepNumber string
}

func NewTracedFromString(traceParams string) *TracingContext {
	parts := strings.Split(traceParams, ":")
	runId, runAttempt, jobName, stepName, stepNumber := parts[0], parts[1], parts[2], parts[3], parts[4]

	return NewTracer(runId, runAttempt, jobName, stepName, stepNumber)
}

func NewTracer(runId, runAttempt, jobName, stepName, stepNumber string) *TracingContext {
	return &TracingContext{RunId: runId, RunAttempt: runAttempt, JobName: jobName, StepName: stepName, StepNumber: stepNumber}
}

func (t *TracingContext) SpanWithContext(context context.Context, name string, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	traceId, _ := trace.TraceIDFromHex(t.GenerateTraceID())
	parentSpan, _ := trace.SpanIDFromHex(t.GenerateStepSpanID_Number())

	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceId,
		SpanID:     parentSpan,
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

func (t *TracingContext) GenerateTraceID() string {
	return GenerateTraceID(t.RunId, t.RunAttempt)
}

func (t *TracingContext) GenerateJobSpanID() string {
	return GenerateJobSpanID(t.RunId, t.RunAttempt, t.JobName)
}

func (t *TracingContext) GenerateStepSpanID() string {
	return GenerateStepSpanID(t.RunId, t.RunAttempt, t.JobName, t.StepName)
}

func (t *TracingContext) GenerateStepSpanID_Number() string {
	return GenerateStepSpanID_Number(t.RunId, t.RunAttempt, t.JobName, t.StepNumber)
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
