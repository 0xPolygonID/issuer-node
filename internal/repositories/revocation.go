package repositories

import (
	"context"

	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type revocation struct{}

// NewRevocation TODO
func NewRevocation() ports.RevocationRepository {
	return &revocation{}
}

func (r *revocation) UpdateStatus(ctx context.Context, conn db.Querier, did *core.DID) ([]*domain.Revocation, error) {
	rows, err := conn.Query(ctx, `UPDATE revocation SET status = $2 WHERE identifier = $1 AND status = $3
RETURNING identifier, nonce, version, status, description`,
		did.String(), domain.RevPublished, domain.RevPending)
	if err != nil {
		return nil, err
	}

	var revs []*domain.Revocation
	for rows.Next() {
		var revoke domain.Revocation
		if err = rows.Scan(&revoke.Identifier, &revoke.Nonce, &revoke.Version, &revoke.Status, &revoke.Description); err != nil {
			return nil, err
		}
		revs = append(revs, &revoke)
	}
	if rows.Err() != nil {
		return nil, err
	}

	return revs, nil
}
