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

type Cache interface {
	Get(key string) (string, error)
	SetTtl(key string, value string, expiration time.Duration) error
	Set(key string, value string) error
}

type redisCache struct {
	c   *redis.Client
	ctx context.Context
}

func (c *redisCache) Set(key string, value string) error {
	return c.SetTtl(key, value, 0)
}

func (c *redisCache) SetTtl(key string, value string, ttl time.Duration) error {
	return c.c.Set(c.ctx, key, value, ttl).Err()
}

func (c *redisCache) Get(key string) (string, error) {
	return c.c.Get(c.ctx, key).Result()
}

func NewRedisCache() (Cache, error) {
	options, err := redis.ParseURL(RedisUrl)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(options)
	if err = errors.Join(redisotel.InstrumentTracing(client), redisotel.InstrumentMetrics(client)); err != nil {
		log.Println(err)
	}
	ctx := context.Background()
	if err = client.Ping(ctx).Err(); err != nil {
		fmt.Printf("failed to connect to redis server: %s\n", err.Error())
		return nil, err
	}
	return &redisCache{client, ctx}, nil
}

func NewDefaultCache() (Cache, error) {
	return NewRedisCache()
}
