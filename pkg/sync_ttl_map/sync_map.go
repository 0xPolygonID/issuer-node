package sync_ttl_map

import (
	"sync"
	"time"
)

type TTLMap struct {
	TTL  time.Duration
	data sync.Map
}

type expireEntry struct {
	ExpiresAt time.Time
	Value     interface{}
}

func (t *TTLMap) Store(key string, val interface{}) {
	t.data.Store(key, expireEntry{
		ExpiresAt: time.Now().Add(t.TTL),
		Value:     val,
	})
}

func (t *TTLMap) Delete(key string) {
	t.data.Delete(key)
}

func (t *TTLMap) Load(key string) (val interface{}) {
	entry, ok := t.data.Load(key)
	if !ok {
		return nil
	}

	expireEntry := entry.(expireEntry)
	if time.Now().After(expireEntry.ExpiresAt) {
		return nil
	}

	return expireEntry.Value
}

func (t *TTLMap) CleaningBackground(cleaning time.Duration) {
	go func() {
		for now := range time.Tick(cleaning) {
			t.data.Range(func(k, v interface{}) bool {
				if v.(expireEntry).ExpiresAt.After(now) {
					t.data.Delete(k)
				}
				return true
			})
		}
	}()
}

func New(ttl time.Duration) *TTLMap {
	return &TTLMap{TTL: ttl}
}
