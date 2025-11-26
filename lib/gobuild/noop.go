package gobuild

import (
	"log"
	"time"

	"github.com/creachadair/gocache"
	"github.com/creachadair/gocache/cachedir"
	"golang.org/x/net/context"
)

type TheCache interface {
	Get(ctx context.Context, actionID string) (outputID, diskPath string, _ error)
	Put(ctx context.Context, obj gocache.Object) (diskPath string, _ error)
	Close(ctx context.Context) error
}

type LocalCache struct {
	Local  *cachedir.Dir
	Logger *log.Logger
}

func (l *LocalCache) Get(ctx context.Context, actionID string) (outputID, diskPath string, _ error) {
	defer l.logOperation("Get", actionID, time.Now())
	return l.Local.Get(ctx, actionID)
}

func (l *LocalCache) Put(ctx context.Context, obj gocache.Object) (diskPath string, _ error) {
	defer l.logOperation("Put", obj.ActionID, time.Now())
	return l.Local.Put(ctx, obj)
}

func (l *LocalCache) Close(ctx context.Context) error {
	return nil
}

func (l *LocalCache) logOperation(op, name string, start time.Time) {
	l.Logger.Printf("GOBUILD %s -> finished operation for %s in %v", op, name, time.Since(start))
}
