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

// UUID returns the UUID from a URN. It can throw an error to prevent bad constructor calls or urns without uuids
func (u URN) UUID() (uuid.UUID, error) {
	if err := u.valid(); err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(string(u[9:]))
}

func (u URN) valid() error {
	if len(u) < len("urn:uuid:") {
		return errors.New("invalid uuid URN length")
	}
	if u[:9] != "urn:uuid:" {
		return errors.New("invalid uuid URN prefix")
	}
	return nil
}

// Parse creates a URN from a string.
func Parse(u string) (URN, error) {
	if err := URN(u).valid(); err != nil {
		return "", err
	}
	return URN(u), nil
}

// UUIDFromURNString returns the UUID from a URN string.
func UUIDFromURNString(s string) (uuid.UUID, error) {
	urn, err := Parse(s)
	if err != nil {
		return uuid.Nil, err
	}
	return urn.UUID()
}
