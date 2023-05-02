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
	t.Skip("timeout redis in ga")
	ctx, cancel := context.WithCancel(context.Background())
	s := miniredis.RunT(t)
	client, err := redis.Open("redis://" + s.Addr())
	require.NoError(t, err)
	defer func() { assert.NoError(t, client.Close()) }()

	wg := sync.WaitGroup{}

	ps := NewRedis(client)
	ps.WithLogger(log.Error)
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
	t.Skip("Skipped because it fails on github actions. Seems that last message is lost.")
	const nEvents = 100
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := miniredis.RunT(t)
	client, err := redis.Open("redis://" + s.Addr())
	require.NoError(t, err)
	defer func() { assert.NoError(t, client.Close()) }()

	ps := NewRedis(client)

	waitg := sync.WaitGroup{}

	// This method panics ...
	ps.Subscribe(ctx, "topic", func(ctx context.Context, payload Message) error {
		defer waitg.Done()
		panic("simulating a panic")
	})

	var count atomic.Int64
	// ... but this other methods still run without problems
	ps.Subscribe(ctx, "topic", func(ctx context.Context, payload Message) error {
		defer waitg.Done()
		return nil
	})
	for i := 0; i < nEvents; i++ {
		waitg.Add(1)
		waitg.Add(1)
		require.NoError(t, ps.Publish(ctx, "topic", &MyEvent{}))
	}

	waitg.Wait()

	assert.Equal(t, nEvents, int(count.Load()))
}
