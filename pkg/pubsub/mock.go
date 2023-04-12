package pubsub

import (
	"context"
)

// Mock is a mock pubsub client
type Mock struct{}

// NewMock returns a new mock pubsub client
func NewMock() Client {
	return &Mock{}
}

// Publish mock
func (rdb *Mock) Publish(ctx context.Context, topic string, payload Event) error {
	return nil
}

// Subscribe mock
func (rdb *Mock) Subscribe(ctx context.Context, topic string, callback EventHandler) {}
