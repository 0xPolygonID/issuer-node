package repositories

import (
	"context"
	"fmt"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/jackc/pgtype"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type identityMerkleTreeRepository struct{}

// NewIdentityMerkleTreeRepository returns a new identityMerkleTreeRepository
func NewIdentityMerkleTreeRepository() ports.IdentityMerkleTreeRepository {
	return &identityMerkleTreeRepository{}
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

func (mt *identityMerkleTreeRepository) GetByID(ctx context.Context, conn db.Querier, mtID uint64) (*domain.IdentityMerkleTree, error) {
	var res domain.IdentityMerkleTree
	row := conn.QueryRow(ctx, "SELECT id, identifier, type FROM identity_mts WHERE id = $1", mtID)
	if err := row.Scan(&res.ID, &res.Identifier, &res.Type); err != nil {
		return nil, fmt.Errorf("error getting merkle tree by id %w", err)
	}
	return &res, nil
}

func (mt *identityMerkleTreeRepository) GetByIdentifierAndTypes(ctx context.Context, conn db.Querier, identifier *w3c.DID, mtTypes []uint16) ([]domain.IdentityMerkleTree, error) {
	var typesSQL pgtype.Int2Array
	if err := typesSQL.Set(mtTypes); err != nil {
		return nil, err
	}

	rows, err := conn.Query(ctx,
		`SELECT id, identifier, type FROM identity_mts WHERE identifier = $1 AND type = ANY($2)`,
		identifier.String(), typesSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trees := make([]domain.IdentityMerkleTree, 0, len(mtTypes))
	for rows.Next() {
		var tree domain.IdentityMerkleTree
		if err = rows.Scan(&tree.ID, &tree.Identifier, &tree.Type); err != nil {
			return nil, err
		}
		trees = append(trees, tree)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return trees, nil
}
