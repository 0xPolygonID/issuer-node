package repositories

import (
	"context"

	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type identity struct {
}

func NewIdentity() ports.IndentityRepository {
	return &identity{}
}

func (i *identity) Save(ctx context.Context, conn db.Querier, identity *domain.Identity) error {
	_, err := conn.Exec(ctx, `INSERT INTO identities (identifier, relay, immutable) VALUES ($1, $2, $3)`,
		identity.Identifier, identity.Relay, identity.Immutable)
	return err
}

func (i *identity) GetByID(ctx context.Context, conn db.Querier, identifier *core.DID) (*domain.Identity, error) {
	identity := domain.Identity{
		State: domain.IdentityState{},
	}
	row := conn.QueryRow(ctx,
		`SELECT  identities.identifier,
       					relay,
       					identities.immutable,
       					state_id,
   						state,           
    					root_of_roots,
    					revocation_tree_root,
    					claims_tree_root,
    					block_timestamp,
    					block_number,
    					tx_id,
    					previous_state,
    					status,               
    					modified_at,
    					created_at          
			   FROM identities
			   LEFT JOIN identity_states ON identities.identifier = identity_states.identifier
			   WHERE identities.identifier=$1 
			        AND ( status = 'transacted' OR status = 'confirmed')
    				OR (identities.identifier=$1 AND status = 'created' AND previous_state is null
				AND NOT EXISTS (SELECT 1 FROM identity_states WHERE identifier=$1 AND (status = 'transacted' OR status = 'confirmed')))
				ORDER BY state_id DESC LIMIT 1`, identifier.String())

	err := row.Scan(&identity.Identifier,
		&identity.Relay,
		&identity.Immutable,
		&identity.State.StateID,
		&identity.State.State,
		&identity.State.RootOfRoots,
		&identity.State.RevocationTreeRoot,
		&identity.State.ClaimsTreeRoot,
		&identity.State.BlockTimestamp,
		&identity.State.BlockNumber,
		&identity.State.TxID,
		&identity.State.PreviousState,
		&identity.State.Status,
		&identity.State.ModifiedAt,
		&identity.State.CreatedAt)

	return &identity, err
}
