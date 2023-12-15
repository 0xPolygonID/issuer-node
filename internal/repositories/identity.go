package repositories

import (
	"context"
	"fmt"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/issuer-node/internal/core/domain"
	"github.com/polygonid/issuer-node/internal/core/ports"
	"github.com/polygonid/issuer-node/internal/db"
)

type identity struct{}

// NewIdentity TODO
func NewIdentity() ports.IndentityRepository {
	return &identity{}
}

// Save - Create new identity
func (i *identity) Save(ctx context.Context, conn db.Querier, identity *domain.Identity) error {
	_, err := conn.Exec(ctx, `INSERT INTO identities (identifier, address, keyType) VALUES ($1, $2, $3)`, identity.Identifier, identity.Address, identity.KeyType)
	return err
}

func (i *identity) GetByID(ctx context.Context, conn db.Querier, identifier w3c.DID) (*domain.Identity, error) {
	identity := domain.Identity{
		State: domain.IdentityState{},
	}
	row := conn.QueryRow(ctx,
		`SELECT  identities.identifier,
						identities.keyType,
						identities.address,
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
		&identity.KeyType,
		&identity.Address,
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

func (i *identity) Get(ctx context.Context, conn db.Querier) (identities []string, err error) {
	rows, err := conn.Query(ctx, `SELECT identifier FROM identities`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var identifier string
		err = rows.Scan(&identifier)
		if err != nil {
			return nil, err
		}
		identities = append(identities, identifier)
	}

	return identities, err
}

func (i *identity) GetUnprocessedIssuersIDs(ctx context.Context, conn db.Querier) (issuersIDs []*w3c.DID, err error) {
	rows, err := conn.Query(ctx,
		`WITH issuers_to_process AS
(
    SELECT  issuer 
		FROM claims
		WHERE identity_state ISNULL AND identifier = issuer
			UNION
		SELECT identifier FROM revocation where status = 0
), transacted_issuers AS
(
    SELECT identifier from identity_states WHERE status = 'transacted'
)

SELECT issuer FROM issuers_to_process WHERE issuer NOT IN (SELECT identifier FROM transacted_issuers);
`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	issuersIDs = make([]*w3c.DID, 0)

	for rows.Next() {
		var issuerIDStr string
		err := rows.Scan(&issuerIDStr)
		if err != nil {
			return nil, err
		}
		var did *w3c.DID
		did, err = w3c.ParseDID(issuerIDStr)
		if err != nil {
			return []*w3c.DID{}, err
		}
		issuersIDs = append(issuersIDs, did)

	}

	if rows.Err() != nil {
		return nil, err
	}

	return issuersIDs, nil
}

func (i *identity) HasUnprocessedStatesByID(ctx context.Context, conn db.Querier, identifier *w3c.DID) (bool, error) {
	row := conn.QueryRow(ctx,
		`WITH issuers_to_process AS
				(
					SELECT  issuer 
						FROM claims
						WHERE identity_state ISNULL AND identifier = issuer
							UNION
						SELECT identifier FROM revocation where status = 0
				), transacted_issuers AS
				(
					SELECT identifier from identity_states WHERE status = 'transacted' or status = 'failed'
				)
				
				SELECT COUNT(*) FROM issuers_to_process WHERE issuer NOT IN (SELECT identifier FROM transacted_issuers) AND issuer = $1`, identifier.String())

	var res int64
	if err := row.Scan(&res); err != nil {
		return false, fmt.Errorf("error getting unprocessed rows %w", err)
	}

	return res > 0, nil
}

func (i *identity) HasUnprocessedAndFailedStatesByID(ctx context.Context, conn db.Querier, identifier *w3c.DID) (bool, error) {
	row := conn.QueryRow(ctx,
		`WITH issuers_to_process AS
         (
             SELECT  issuer
             FROM claims
             WHERE identity_state ISNULL AND identifier = issuer
             UNION
             SELECT identifier FROM revocation where status = 0
         ), transacted_issuers AS
         (
             SELECT identifier from identity_states WHERE status = 'transacted'
         )

		SELECT COUNT(*) FROM (SELECT issuers_to_process.issuer
                    FROM issuers_to_process
                    WHERE issuer NOT IN (SELECT identifier FROM transacted_issuers)
                      AND issuer = $1
                    UNION
                    SELECT identity_states.tx_id
                    FROM identity_states
                    WHERE identifier = $1
                      AND status = 'failed') as res;`, identifier.String())

	var res int64
	if err := row.Scan(&res); err != nil {
		return false, fmt.Errorf("error getting unprocessed and failed rows %w", err)
	}

	return res > 0, nil
}
