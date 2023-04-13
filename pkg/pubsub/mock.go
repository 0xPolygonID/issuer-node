package pubsub

import (
	"context"
)

// Mock is a mock pubsub client
type Mock struct {
	events map[string][]Event
}

// NewMock returns a new mock pubsub client
func NewMock() *Mock {
	return &Mock{
		events: make(map[string][]Event),
	}
}

// Publish a payload into a topic
func (ps *Mock) Publish(_ context.Context, topic string, payload Event) error {
	if _, found := ps.events[topic]; !found {
		ps.events[topic] = make([]Event, 0)
	}
	ps.events[topic] = append(ps.events[topic], payload)
	return nil
}

// Subscribe to a topic.
// Not implemented
func (ps *Mock) Subscribe(_ context.Context, topic string, callback EventHandler) {}

// AllPublishedEvents returns the list of all published events allowing test inspection
func (ps *Mock) AllPublishedEvents(topic string) []Event {
	collection, found := ps.events[topic]
	if !found {
		return nil
	}
	return collection
}

// Clear empties the list of events in this topic
func (ps *Mock) Clear(topic string) {
	if len(ps.AllPublishedEvents(topic)) != 0 {
		ps.events[topic] = make([]Event, 0)
	}
}
