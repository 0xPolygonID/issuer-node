package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// Open opens a connection to redis and returns it
func Open(ctx context.Context, url string) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(opts)
	if err := Status(ctx, rdb); err != nil {
		return nil, err
	}
	return rdb, nil
}

// Status returns nil of redis status is ok. Otherwise a redis status err
func Status(ctx context.Context, rdb *redis.Client) error {
	if pingCmd := rdb.Ping(ctx); pingCmd.Err() != nil {
		return pingCmd.Err()
	}
	return nil
}
