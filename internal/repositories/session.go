package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/issuer-node/internal/core/ports"
	"github.com/polygonid/issuer-node/pkg/cache"
	link_state "github.com/polygonid/issuer-node/pkg/link"
)

const (
	defaultTTL = 5 * time.Minute
)

type cached struct {
	cache cache.Cache
}

// NewSessionCached returns a new cached manager
func NewSessionCached(c cache.Cache) ports.SessionRepository {
	return &cached{cache: c}
}

// Get returns the cached session
func (c *cached) Get(ctx context.Context, key string) (protocol.AuthorizationRequestMessage, error) {
	var message protocol.AuthorizationRequestMessage
	found := c.cache.Get(ctx, key, &message)
	if !found {
		return message, fmt.Errorf("authorization request not found")
	}

	return message, nil
}

// Set stores the given session information
func (c *cached) Set(ctx context.Context, key string, value protocol.AuthorizationRequestMessage) error {
	return c.cache.Set(ctx, key, value, defaultTTL)
}

// SetLink - stores the given session information
func (c *cached) SetLink(ctx context.Context, key string, value link_state.State) error {
	return c.cache.Set(ctx, key, value, defaultTTL)
}

func (c *cached) GetLink(ctx context.Context, key string) (link_state.State, error) {
	var message link_state.State
	found := c.cache.Get(ctx, key, &message)
	if !found {
		return message, fmt.Errorf("link state not found")
	}
	return message, nil
}
