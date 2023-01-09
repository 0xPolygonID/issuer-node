package repositories

import (
	"context"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type identityMerkleTreeRepository struct {
	conn db.Querier
}

func NewIdentityMerkleTreeRepository(conn db.Querier) ports.IdentityMerkleTreeRepository {
	return &identityMerkleTreeRepository{
		conn: conn,
	}
}

func (mt *identityMerkleTreeRepository) Save(ctx context.Context, conn db.Querier, identifier string, mtType uint16) (*domain.IdentityMerkleTree, error) {
	var id uint64
	row := conn.QueryRow(ctx, `INSERT INTO identity_mts (identifier, type) VALUES ($1, $2) RETURNING id`, identifier, mtType)
	if err := row.Scan(&id); err != nil {
		return nil, err
	}

	imt := &domain.IdentityMerkleTree{
		ID:         id,
		Identifier: identifier,
		Type:       mtType,
	}
	return imt, nil
}

func (mt *identityMerkleTreeRepository) UpdateByID(ctx context.Context, conn db.Querier, imt *domain.IdentityMerkleTree) error {
	_, err := conn.Exec(ctx, `UPDATE identity_mts SET identifier = $1, type = $2 WHERE id = $3`,
		imt.Identifier, imt.Type, imt.ID)
	return err
}
