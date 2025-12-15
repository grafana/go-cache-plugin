package gobuild

import (
	"github.com/creachadair/gocache"
	"golang.org/x/net/context"
)

type TheCache interface {
	Get(ctx context.Context, actionID string) (outputID, diskPath string, _ error)
	Put(ctx context.Context, obj gocache.Object) (diskPath string, _ error)
	Close(ctx context.Context) error
}

type NoopCache struct {
}

func (l *NoopCache) Get(ctx context.Context, actionID string) (outputID, diskPath string, _ error) {
	return "", "", nil
}

func (l *NoopCache) Put(ctx context.Context, obj gocache.Object) (diskPath string, _ error) {
	return "", nil
}

func (l *NoopCache) Close(ctx context.Context) error {
	return nil
}
