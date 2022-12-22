package repositories

import (
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type identity struct {
	db *db.Sqlx
}

func NewIdentity(db *db.Sqlx) ports.IndentityRepository {
	return &identity{
		db: db,
	}
}

func (i *identity) Save() error {
	return nil
}
