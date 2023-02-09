package health

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	iRedis "github.com/polygonid/sh-id-platform/internal/redis"
)

const (
	redis  = "redis"
	db     = "db"
	memory = "memory"
)

// Status struct
type Status struct {
	pingers map[string]Ping
}

// Ping interface
type Ping interface {
	Ping(ctx context.Context) error
}

// New returns a Health instance
func New(pingers ...Ping) *Status {
	m := make(map[string]Ping)

	for _, p := range pingers {
		switch t := p.(type) {
		case *pgxpool.Pool:
			m[db] = t
		case iRedis.Wrapper:
			m[redis] = t
		}
	}

	return &Status{m}
}

// Status returns the whether the cache and the db is active or not
func (h *Status) Status(ctx context.Context) map[string]bool {
	m := make(map[string]bool)

	for key, val := range h.pingers {
		m[key] = true
		if err := val.Ping(ctx); err != nil {
			m[key] = false
		}
	}

	return m
}
