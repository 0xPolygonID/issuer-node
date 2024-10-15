package sync_ttl_map

import (
	"sync"
	"time"
)

// TTLMap structure
type TTLMap struct {
	TTL  time.Duration
	data sync.Map
}

type expireEntry struct {
	ExpiresAt time.Time
	Value     interface{}
}

// Store saves a key/value pair into TTLMap
func (t *TTLMap) Store(key string, val interface{}) {
	t.data.Store(key, expireEntry{
		ExpiresAt: time.Now().Add(t.TTL),
		Value:     val,
	})
}

// Delete deletes the given key from TTLMap
func (t *TTLMap) Delete(key string) {
	t.data.Delete(key)
}

// Load retrieves the value of the given key from TTLMap
func (t *TTLMap) Load(key string) (val interface{}) {
	entry, ok := t.data.Load(key)
	if !ok {
		return nil
	}

	expireEntry, ok := entry.(expireEntry)
	if !ok {
		return nil
	}

	if time.Now().After(expireEntry.ExpiresAt) {
		return nil
	}

	return expireEntry.Value
}

// CleaningBackground starts a go routine for cleaning expired entries
func (t *TTLMap) CleaningBackground(cleaning time.Duration) {
	go func() {
		for now := range time.Tick(cleaning) {
			t.data.Range(func(k, v interface{}) bool {
				if expireEntry, ok := v.(expireEntry); ok && expireEntry.ExpiresAt.After(now) {
					t.data.Delete(k)
				}
				return true
			})
		}
	}()
}

// New returns a new TTLMap
func New(ttl time.Duration) *TTLMap {
	return &TTLMap{TTL: ttl}
}
