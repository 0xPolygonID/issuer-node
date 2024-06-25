package urn

import (
	"errors"

	"github.com/google/uuid"
)

// URN represents a Uniform Resource Name.
type URN string

// FromUUID creates a URN from a UUID.
func FromUUID(uuid uuid.UUID) URN {
	return URN("urn:uuid:" + uuid.String())
}

// Parse extracts a UUID from a URN.
func Parse(u URN) (uuid.UUID, error) {
	if len(u) < len("urn:uuid:") {
		return uuid.UUID{}, errors.New("invalid uuid URN length")
	}
	if u[:9] != "urn:uuid:" {
		return uuid.UUID{}, errors.New("invalid uuid URN prefix")
	}
	return uuid.Parse(string(u[9:]))
}
