package cache

import (
	"context"
	"reflect"
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

// NewMemoryCache returns a basic in memory cache
func NewMemoryCache() Cache {
	return &memory{
		c: cache.New(memoryDefTTL, memoryCleanUPPeriod),
	}
}

// Set sets an item in the in memory cache
func (m *memory) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	m.c.Set(key, value, ttl)
	return nil
}

// Get retrieves a cache entry and a boolean telling it is found or not
// value must be passed as reference as the cached value will be stored there
func (m *memory) Get(_ context.Context, key string, value any) bool {
	mVal, exists := m.c.Get(key)
	if exists && reflect.TypeOf(value) == reflect.TypeOf(&mVal) {
		reflect.ValueOf(value).Elem().Set(reflect.ValueOf(mVal))
		return true
	}

	return false
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
