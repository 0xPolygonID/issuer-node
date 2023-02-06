package cache

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	memoryDefTTL        = 60 * time.Minute
	memoryCleanUPPeriod = 1 * time.Minute
)

type memory struct {
	c *cache.Cache
}

// Set sets an item in the in memory cache
func (m *memory) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	m.c.Set(key, value, ttl)
	return nil
}

// Get retrieves a cache entry and a boolean telling it is found or not
func (m *memory) Get(_ context.Context, key string) (v any, f bool) {
	return m.c.Get(key)
}

// Exists returns true if the key exists in the cache
func (m *memory) Exists(_ context.Context, key string) bool {
	_, found := m.c.Get(key)
	return found
}

// Delete removes and entry from the cache
func (m *memory) Delete(_ context.Context, key string) error {
	m.c.Delete(key)
	return nil
}

// NewMemoryCache returns a basic in memory cache
func NewMemoryCache() Cache {
	return &memory{
		c: cache.New(memoryDefTTL, memoryCleanUPPeriod),
	}
}
