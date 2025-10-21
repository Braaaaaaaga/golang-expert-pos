package storage

import (
	"context"
	"time"
)

type Storage interface {
	Increment(ctx context.Context, key string, ttl time.Duration) (int64, error)
	Get(ctx context.Context, key string) (int64, error)
	Set(ctx context.Context, key string, value int64, ttl time.Duration) error
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Close() error
}
