package syncttlmap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMtSave(t *testing.T) {
	ttl := 50 * time.Millisecond
	cleanup := 150 * time.Millisecond

	mMap := New(ttl)
	mMap.CleaningBackground(cleanup)

	assert.Equal(t, mMap.TTL, ttl)

	notExists := mMap.Load("notExistingKey")
	assert.Equal(t, notExists, nil)

	mMap.Store("hello", "world")
	exists := mMap.Load("hello")
	assert.Equal(t, exists, "world")

	time.Sleep(200 * time.Millisecond)
	shouldBeNil := mMap.Load("hello")
	assert.Equal(t, shouldBeNil, nil)
}
