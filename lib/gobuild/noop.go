package gobuild

import (
	"github.com/creachadair/gocache"
	"github.com/creachadair/gocache/cachedir"
	"github.com/grafana/go-cache-plugin/lib/otel"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/net/context"
)

type TheCache interface {
	Get(ctx context.Context, actionID string) (outputID, diskPath string, _ error)
	Put(ctx context.Context, obj gocache.Object) (diskPath string, _ error)
	Close(ctx context.Context) error
}

type LocalCache struct {
	Local  *cachedir.Dir
	Tracer *otel.TraceMeta
}

func (l *LocalCache) Get(ctx context.Context, actionID string) (outputID, diskPath string, _ error) {
	_, span := l.Tracer.SpanWithContext(ctx, "BUILD-GET", attribute.KeyValue{Key: "action_id", Value: attribute.StringValue(actionID)})
	defer func() {
		span.End()
	}()

	return l.Local.Get(ctx, actionID)
}

func (l *LocalCache) Put(ctx context.Context, obj gocache.Object) (diskPath string, _ error) {
	_, span := l.Tracer.SpanWithContext(ctx, "BUILD-PUT", attribute.KeyValue{Key: "action_id", Value: attribute.StringValue(obj.ActionID)})
	defer func() {
		span.End()
	}()
	return l.Local.Put(ctx, obj)
}

func (l *LocalCache) Close(ctx context.Context) error {
	return nil
}
