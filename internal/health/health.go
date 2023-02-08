package health

import (
	"context"
)

// Health struct
type Health struct {
	db    storage
	cache cache
}

type storage interface {
	Ping(ctx context.Context) error
}

type cache interface {
	Ping(ctx context.Context) error
}

// Status struct
type Status struct {
	DB    bool
	Cache bool
}

// New returns a Health instance
func New(db storage, cache cache) *Health {
	return &Health{db: db, cache: cache}
}

// Status returns the whether the cache and the db is active or not
func (h *Health) Status(ctx context.Context) *Status {
	status := Status{
		DB:    true,
		Cache: true,
	}
	if err := h.db.Ping(ctx); err != nil {
		status.DB = false
	}

	if err := h.cache.Ping(ctx); err != nil {
		status.Cache = false
	}

	return &status
}
