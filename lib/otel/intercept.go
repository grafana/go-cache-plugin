package otel

import (
	"context"
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

func (m *Spanner) ProcessId(ctx context.Context, id string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	span, ok := m.data[id]
	if ok {
		span.End()
		delete(m.data, id)
		return
	}

	_, span = m.tracingContext.SpanWithContext(ctx, id)
	m.data[id] = span
}
