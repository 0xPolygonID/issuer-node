package repositories

import (
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type identity struct {
	db *db.Sqlx
}

// NewIdentity returns a ports.IndentityRepository database implementation
func NewIdentity(dbConn *db.Sqlx) ports.IndentityRepository {
	return &identity{
		db: dbConn,
	}
}

// Save saves something.
func (i *identity) Save() error {
	return nil
}
