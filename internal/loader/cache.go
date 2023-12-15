package loader

import (
	"context"
	"fmt"

	"github.com/polygonid/issuer-node/internal/log"
	"github.com/polygonid/issuer-node/pkg/cache"
)

type schemaData struct {
	Schema    []byte
	Extension string
}

type cached struct {
	url    string
	loader Loader
	cache  cache.Cache
}

// Load returns a schema. It uses an internal cache and a loader. This caches can, and probably is, shared with
// other loaders. If the file is found in the cache it returns it. If not, loads the file using the internal loader
// and caches it.
// TTL for cached items is forever
func (c *cached) Load(ctx context.Context) (schema []byte, extension string, err error) {
	ctx = log.With(ctx, "key", c.key(c.url))
	d := schemaData{}
	if found := c.cache.Get(ctx, c.key(c.url), &d); found {
		log.Debug(ctx, "schema found in cache")
		return d.Schema, d.Extension, nil
	}

	if d.Schema, d.Extension, err = c.loader.Load(ctx); err != nil {
		return nil, "", err
	}

	if err := c.cache.Set(ctx, c.key(c.url), d, cache.ForEver); err != nil {
		log.Warn(ctx, "adding schema to Redis. Bypassing cache")
	}

	return d.Schema, d.Extension, nil
}

func (c *cached) key(url string) string {
	return fmt.Sprintf("schema-%s", url)
}

// Cached is a file loader that uses a cache. That cache can be shared by multiple loaders.
func Cached(l Loader, c cache.Cache, url string) Loader {
	return &cached{
		url:    url,
		loader: l,
		cache:  c,
	}
}

// CachedFactory returns a function factory able to create Cached Loaders. That means, file loaders that
// looks on a cache for a file before trying to fetch it
func CachedFactory(f Factory, c cache.Cache) Factory {
	return func(url string) Loader {
		return Cached(f(url), c, url)
	}
}
