package services

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	SetTtl(ctx context.Context, key string, value string, expiration time.Duration) error
	Set(ctx context.Context, key string, value string) error
}
