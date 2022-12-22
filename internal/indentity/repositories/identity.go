package repositories

import (
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/indentity/core/ports"
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
