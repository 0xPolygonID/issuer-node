package repositories

import (
	"context"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/issuer-node/internal/core/domain"
	"github.com/polygonid/issuer-node/internal/core/ports"
	"github.com/polygonid/issuer-node/internal/db"
)

type revocation struct{}

// NewRevocation TODO
func NewRevocation() ports.RevocationRepository {
	return &revocation{}
}

func (r *revocation) UpdateStatus(ctx context.Context, conn db.Querier, did *w3c.DID) ([]*domain.Revocation, error) {
	rows, err := conn.Query(ctx, `UPDATE revocation SET status = $2 WHERE identifier = $1 AND status = $3
RETURNING identifier, nonce, version, status, description`,
		did.String(), domain.RevPublished, domain.RevPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
