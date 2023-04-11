package pubsub

import (
	"context"
)

// MockClient is a mock pubsub client
type MockClient struct{}

// NewMock returns a new mock pubsub client
func NewMock() Client {
	return &MockClient{}
}

// Publish mock
func (rdb *MockClient) Publish(ctx context.Context, topic string, payload Event) error {
	return nil
}

// Subscribe mock
func (rdb *MockClient) Subscribe(ctx context.Context, topic string, callback EventHandler) {}

// Unsubscribe mock
func (rdb *MockClient) Unsubscribe(ctx context.Context, topic string) error {
	return nil
}
