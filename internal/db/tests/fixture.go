package tests

import (
	"github.com/polygonid/sh-id-platform/internal/db"
)

type Fixture struct {
	storage *db.Storage
}

func NewFixture(storage *db.Storage) *Fixture {
	return &Fixture{
		storage: storage,
	}
}
