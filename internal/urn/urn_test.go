package urn

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFromUUID(t *testing.T) {
	id := uuid.New()
	urn := FromUUID(id)
	assert.Equal(t, "urn:uuid:"+id.String(), string(urn))
}

func TestUUIDFromURNString(t *testing.T) {
	for _, ts := range []struct {
		urn string
		err error
	}{
		{"", errors.New("invalid uuid URN length")},
		{"urn:uuid:", errors.New("invalid UUID length: 0")},
		{"urn:uuid:1234", errors.New("invalid UUID length: 4")},
		{"urn:uuid:123e4567-e89b-12d3-a456-426614174000", nil},
	} {
		t.Run(ts.urn, func(t *testing.T) {
			u, err := UUIDFromURNString(ts.urn)
			if err == nil {
				assert.Equal(t, ts.urn[9:], u.String())
			} else {
				assert.Equal(t, ts.err.Error(), err.Error())
			}
		})
	}
}
