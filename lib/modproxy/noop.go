package modproxy

import (
	"expvar"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/goproxy/goproxy"
	"golang.org/x/net/context"
)

type LocalCache struct {
	Local  goproxy.DirCacher
	Logger *log.Logger
}

func NowLocalModCacher(path string, logger *log.Logger) *LocalCache {
	return &LocalCache{Local: goproxy.DirCacher(path), Logger: logger}
}

func (l *LocalCache) Get(ctx context.Context, name string) (io.ReadCloser, error) {
	defer l.logOperation("Get", name, time.Now())
	return l.Local.Get(ctx, name)
}

// Put puts a cache for the name with the content.
func (l *LocalCache) Put(ctx context.Context, name string, content io.ReadSeeker) error {
	defer l.logOperation("Put", name, time.Now())
	return l.Local.Put(ctx, name, content)
}

func (l *LocalCache) Metrics() *expvar.Map {
	return &expvar.Map{}
}

func (l *LocalCache) logOperation(op, name string, start time.Time) {
	l.Logger.Printf("GOMOD %s -> finished operation for %s in %v", op, name, time.Since(start))
}

type LoggingGoFetcher struct {
	Delegate goproxy.Fetcher
	Logger   *log.Logger
}

func (l *LoggingGoFetcher) Query(ctx context.Context, path, query string) (string, time.Time, error) {
	defer l.logOperation("QUERY", query, time.Now())
	return l.Delegate.Query(ctx, path, query)
}

func (l *LoggingGoFetcher) List(ctx context.Context, path string) (versions []string, err error) {
	defer l.logOperation("LIST", path, time.Now())
	return l.Delegate.List(ctx, path)
}
func (l *LoggingGoFetcher) Download(ctx context.Context, path, version string) (info, mod, zip io.ReadSeekCloser, err error) {
	defer l.logOperation("DOWNLOAD", fmt.Sprintf("%s/%s", path, version), time.Now())
	return l.Delegate.Download(ctx, path, version)
}

func (l *LoggingGoFetcher) logOperation(op, name string, start time.Time) {
	l.Logger.Printf("GOMOD %s -> finished operation for %s in %v", op, name, time.Since(start))
}
