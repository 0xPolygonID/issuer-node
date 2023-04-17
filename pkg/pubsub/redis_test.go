package pubsub

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/redis"
)

type MyEvent struct {
	Field1 string
	Field2 int
	Field3 int
	Field4 bool
}

func (e *MyEvent) Unmarshal(data Message) error {
	return json.Unmarshal(data, &e)
}

func (e *MyEvent) Marshal() (data Message, err error) {
	return json.Marshal(e)
}

func TestRedisHappyPath(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := miniredis.RunT(t)
	client, err := redis.Open("redis://" + s.Addr())
	require.NoError(t, err)

	wg := sync.WaitGroup{}

	ps := NewRedis(client)
	ps.WithLogger(log.Debug)
	ps.Subscribe(ctx, "topic", func(ctx context.Context, payload Message) error {
		defer wg.Done()
		var ev MyEvent
		assert.NoError(t, ev.Unmarshal(payload))
		assert.Equal(t, "field1", ev.Field1)
		assert.Equal(t, 33, ev.Field2)
		assert.Equal(t, -15, ev.Field3)
		assert.Equal(t, true, ev.Field4)
		return nil
	})

	wg.Add(1)
	require.NoError(t, ps.Publish(ctx, "topic", &MyEvent{
		Field1: "field1",
		Field2: 33,
		Field3: -15,
		Field4: true,
	}))

	wg.Wait()

	cancel()
}

func TestRedisRecover(t *testing.T) {
	const nEvents = 100
	ctx, cancel := context.WithCancel(context.Background())
	s := miniredis.RunT(t)
	client, err := redis.Open("redis://" + s.Addr())
	require.NoError(t, err)

	wg := sync.WaitGroup{}

	ps := NewRedis(client)
	// This method panics ...
	ps.Subscribe(ctx, "topic", func(ctx context.Context, payload Message) error {
		defer wg.Done()
		panic("Simulating a panic")
		return nil
	})
	var count atomic.Int64
	// ... but this other methods still run without problems
	ps.Subscribe(ctx, "topic", func(ctx context.Context, payload Message) error {
		defer wg.Done()
		count.Add(1)
		return nil
	})

	for i := 0; i < nEvents; i++ {
		wg.Add(2)
		require.NoError(t, ps.Publish(ctx, "topic", &MyEvent{}))
	}

	wg.Wait()

	assert.Equal(t, nEvents, int(count.Load()))

	cancel()
}
