package otel

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"sync"
)

type Spanner struct {
	tracingContext *TracingContext
	mutex          sync.Mutex
	data           map[string]trace.Span
}

func NewAwesomeSpanner(context *TracingContext) *Spanner {
	return &Spanner{tracingContext: context, data: make(map[string]trace.Span)}
}

type CacheRequest struct {
	Id       string `json:"ID"`
	Miss     bool   `json:"Miss"`
	ActionId string `json:"ActionID"`
	Command  string `json:"Command"`
}

func (m *Spanner) ProcessId(ctx context.Context, request CacheRequest) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	span, ok := m.data[request.Id]
	if ok {
		if request.Miss {
			span.SetStatus(codes.Error, "Cache Miss")
		} else {
			span.SetStatus(codes.Ok, "Cache Hit")
		}

		span.End()
		delete(m.data, request.Id)
		return
	}

	_, span = m.tracingContext.SpanWithContext(
		ctx,
		fmt.Sprintf("Cache-%s", request.ActionId),
		attribute.KeyValue{Key: "command", Value: attribute.StringValue(request.Command)},
	)
	m.data[request.Id] = span
}
