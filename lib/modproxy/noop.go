package modproxy

import (
	"expvar"
	"github.com/grafana/go-cache-plugin/lib/otel"
	"go.opentelemetry.io/otel/attribute"
	"io"
	"io/fs"
	"time"

	"github.com/goproxy/goproxy"
	"golang.org/x/net/context"
)

type NoopCacher struct {
}

func (n *NoopCacher) Get(ctx context.Context, name string) (io.ReadCloser, error) {
	return nil, fs.ErrNotExist
}

func (n *NoopCacher) Put(ctx context.Context, name string, content io.ReadSeeker) error {
	return nil
}

type LocalCache struct {
	Local  goproxy.Cacher
	Tracer *otel.TraceMeta
}

func NewLocalModCacher(path string, tracer *otel.TraceMeta) *LocalCache {
	return &LocalCache{Local: goproxy.DirCacher(path), Tracer: tracer}
}

func NewNoopModCacher(tracer *otel.TraceMeta) *LocalCache {
	return &LocalCache{Local: &NoopCacher{}, Tracer: tracer}
}

func (l *LocalCache) Get(ctx context.Context, name string) (io.ReadCloser, error) {
	_, span := l.Tracer.SpanWithContext(ctx, "GOMOD-GET", attribute.KeyValue{Key: "package", Value: attribute.StringValue(name)})
	defer func() {
		span.End()
	}()
	return l.Local.Get(ctx, name)
}

// Put puts a cache for the name with the content.
func (l *LocalCache) Put(ctx context.Context, name string, content io.ReadSeeker) error {
	_, span := l.Tracer.SpanWithContext(ctx, "GOMOD-PUT", attribute.KeyValue{Key: "package", Value: attribute.StringValue(name)})
	defer func() {
		span.End()
	}()
	return l.Local.Put(ctx, name, content)
}

func (l *LocalCache) Metrics() *expvar.Map {
	return &expvar.Map{}
}

type LoggingGoFetcher struct {
	Delegate goproxy.Fetcher
	Tracer   *otel.TraceMeta
}

func (l *LoggingGoFetcher) Query(ctx context.Context, path, query string) (string, time.Time, error) {
	_, span := l.Tracer.SpanWithContext(ctx, "GOMOD-QUERY", attribute.KeyValue{Key: "package", Value: attribute.StringValue(path)})
	defer func() {
		span.End()
	}()
	return l.Delegate.Query(ctx, path, query)
}

func (l *LoggingGoFetcher) List(ctx context.Context, path string) (versions []string, err error) {
	_, span := l.Tracer.SpanWithContext(ctx, "GOMOD-LIST", attribute.KeyValue{Key: "package", Value: attribute.StringValue(path)})
	defer func() {
		span.End()
	}()
	return l.Delegate.List(ctx, path)
}
func (l *LoggingGoFetcher) Download(ctx context.Context, path, version string) (info, mod, zip io.ReadSeekCloser, err error) {
	_, span := l.Tracer.SpanWithContext(ctx, "GOMOD-DOWNLOAD", attribute.KeyValue{Key: "package", Value: attribute.StringValue(path)})
	defer func() {
		span.End()
	}()
	return l.Delegate.Download(ctx, path, version)
}
