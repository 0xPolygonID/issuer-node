package cache

import (
	"context"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

type redisCache struct {
	redis *cache.Cache
}

// NewRedisCache returns a new cache based on Redis
func NewRedisCache(client *redis.Client) Cache {
	myc := cache.New(&cache.Options{Redis: client})
	return &redisCache{redis: myc}
}

// Set sets a new entry in redis cache
func (c *redisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	item := &cache.Item{
		Ctx:            ctx,
		Key:            key,
		Value:          value,
		TTL:            ttl,
		SkipLocalCache: true,
	}
	return c.redis.Set(item)
}

// Get returns an entry from redis and a boolean telling if the key has been found in redis
// value must be passed as reference as the cached value will be stored there
func (c *redisCache) Get(ctx context.Context, key string, value any) bool {
	if err := c.redis.Get(ctx, key, &value); err != nil {
		return false
	}

	return true
}

// Exists returns true if the key exists in redis
func (c *redisCache) Exists(ctx context.Context, key string) bool {
	return c.redis.Exists(ctx, key)
}

// Delete removes an entry from redis
func (c *redisCache) Delete(ctx context.Context, key string) error {
	return c.redis.Delete(ctx, key)
}
