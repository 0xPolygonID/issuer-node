package cache

import (
	"context"
	"time"
)

// NullCache is a null cache that does nothing.
type NullCache struct{}

// Set does nothing
func (n *NullCache) Set(_ context.Context, _ string, _ any, _ time.Duration) error {
	return nil
}

// Get returns not found
func (n *NullCache) Get(_ context.Context, _ string) (any, bool) {
	return nil, false
}

// Exists returns it doesn't exists
func (n *NullCache) Exists(_ context.Context, _ string) bool {
	return false
}

// Delete does nothing
func (n *NullCache) Delete(_ context.Context, _ string) error {
	return nil
}
