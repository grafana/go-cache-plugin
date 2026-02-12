package otel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// GoCacheTracer The sole purpose of GoCacheTracer is to intercept GOCACHEPROG commands and report traces
type GoCacheTracer struct {
	tracingContext *TracingContext
	mutex          sync.Mutex
	data           map[string]trace.Span
}

func NewGoCacheTracer(context *TracingContext) *GoCacheTracer {
	return &GoCacheTracer{tracingContext: context, data: make(map[string]trace.Span)}
}

type CacheRequest struct {
	Id       string `json:"ID"`
	Miss     bool   `json:"Miss"`
	ActionId string `json:"ActionID"`
	Command  string `json:"Command"`
}

func (m *GoCacheTracer) ProcessCacheRequest(ctx context.Context, data []byte) error {
	request, err := parseCacheRequest(data)
	if err != nil {
		return err
	}

	m.ProcessId(ctx, request)
	return nil
}

func (m *GoCacheTracer) ProcessId(ctx context.Context, request CacheRequest) {
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

func parseCacheRequest(buf []byte) (CacheRequest, error) {
	var data map[string]any

	err := json.Unmarshal(buf, &data)
	if err != nil {
		slice, _, found := bytes.Cut(buf, []byte{'\n'})
		if found {
			err2 := json.Unmarshal(slice, &data)
			if err2 != nil {
				return CacheRequest{}, err
			}
		}
	}

	var id string
	var miss bool
	var command string
	var actionId string

	id1, ok := data["ID"]
	if !ok {
		return CacheRequest{}, errors.New("id field not found in the request")
	} else {
		id = fmt.Sprint(id1)
	}

	command1, ok := data["Command"]
	if !ok {
		command = ""
	} else {
		command = fmt.Sprint(command1)
	}

	_, ok = data["Miss"]
	if !ok {
		miss = false
	} else {
		miss = true
	}

	actionId1, ok := data["ActionID"]
	if !ok {
		actionId = ""
	} else {
		actionId = fmt.Sprint(actionId1)
	}

	data2 := CacheRequest{
		Id:       id,
		ActionId: actionId,
		Miss:     miss,
		Command:  command,
	}
	return data2, nil
}
