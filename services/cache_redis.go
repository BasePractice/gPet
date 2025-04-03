package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	c *redis.Client
}

func (c *redisCache) Set(ctx context.Context, key string, value string) error {
	return c.SetTtl(ctx, key, value, 0)
}

func (c *redisCache) SetTtl(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.c.Set(ctx, key, value, ttl).Err()
}

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	return c.c.Get(ctx, key).Result()
}

func NewRedisCache(ctx context.Context) (Cache, error) {
	options, err := redis.ParseURL(RedisUrl)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(options)
	if err = errors.Join(redisotel.InstrumentTracing(client), redisotel.InstrumentMetrics(client)); err != nil {
		log.Println(err)
	}
	if err = client.Ping(ctx).Err(); err != nil {
		fmt.Printf("failed to connect to redis server: %s\n", err.Error())
		return nil, err
	}
	return &redisCache{client}, nil
}

func NewDefaultCache(ctx context.Context) (Cache, error) {
	return NewRedisCache(ctx)
}
