package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/valkey-io/valkey-go"

	"github.com/polygonid/sh-id-platform/internal/log"
)

type valKeyCache struct {
	client valkey.Client
}

// NewValKeyCache returns a new cache based on Valkey
func NewValKeyCache(client valkey.Client) Cache {
	return &valKeyCache{client: client}
}

func (v valKeyCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	var val []byte
	var err error
	val, err = json.Marshal(value)
	if err != nil {
		log.Error(ctx, "error marshalling value", "err:", err)
		return err
	}
	return v.client.Do(ctx, v.client.B().Set().Key(key).Value(string(val)).Px(ttl).Build()).Error()
}

func (v valKeyCache) Get(ctx context.Context, key string, value any) bool {
	result := v.client.Do(ctx, v.client.B().Get().Key(key).Build())
	if result.Error() != nil {
		log.Error(ctx, "error getting value", "err:", result.Error())
		return false
	}
	value1, err := result.AsBytes()
	if err != nil {
		log.Error(ctx, "error converting value", "err:", err)
		return false
	}

	if err := json.Unmarshal(value1, value); err != nil {
		log.Error(ctx, "error unmarshalling value", "err:", err)
		return false
	}

	return true
}

func (v valKeyCache) Exists(ctx context.Context, key string) bool {
	result := v.client.Do(ctx, v.client.B().Exists().Key(key).Build())
	if result.Error() != nil {
		log.Error(ctx, "error checking if key exists", "err:", result.Error())
		return false
	}
	ok, err := result.AsInt64()
	if err != nil {
		log.Error(ctx, "error converting result to int64", "err:", err)
		return false
	}
	return ok == 1
}

func (v valKeyCache) Delete(ctx context.Context, key string) error {
	err := v.client.Do(ctx, v.client.B().Del().Key(key).Build()).Error()
	return err
}
