package cache

import (
	"context"
	"time"
)

const (
	ForEver = 0 * time.Second // ForEver It can be cached forever
)

// Cache interface propose an interface that any cache should adhere
type Cache interface {
	// Set sets a value in the caches accessible by the key. The ttl param is the maximum time to live in the cache
	// a ttl=0 means that the entry could be cached forever
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	// Get searches for a non expired entry in the cache and returns the result in the value variable sent as reference and a found paramenter. You should only trust the returned value if f is true
	Get(ctx context.Context, key string, value any) bool
	// Exists tells whether a key exists in the cache with a valid ttl
	Exists(ctx context.Context, key string) bool
	// Delete removes an entry from the cache.
	Delete(ctx context.Context, key string) error
}
