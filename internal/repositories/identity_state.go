package repositories

import (
	"context"
	"fmt"

	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type identityState struct{}

// NewIdentityState returns a new identity state repository
func NewIdentityState() ports.IdentityStateRepository {
	return &identityState{}
}

func (isr *identityState) Save(ctx context.Context, conn db.Querier, state domain.IdentityState) error {
	query := `INSERT INTO identity_states (
		identifier,
		state,
		root_of_roots,
		claims_tree_root,
		revocation_tree_root,
		block_timestamp,
		block_number,
		tx_id,
		previous_state,
		status
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT DO NOTHING`
	_, err := conn.Exec(ctx, query,
		state.Identifier,
		state.State,
		state.RootOfRoots,
		state.ClaimsTreeRoot,
		state.RevocationTreeRoot,
		state.BlockTimestamp,
		state.BlockNumber,
		state.TxID,
		state.PreviousState,
		state.Status,
	)
	if err != nil {
		return fmt.Errorf("failed insert new state record: %w", err)
	}
	return nil
}

// GetLatestStateByIdentifier returns the latest confirmed state or genesis state.
// Firstly try to return a 'confirmed' and non-genesis state.
// If 'confirmed' and non-genesis state are not found. Return genesis state.
func (isr *identityState) GetLatestStateByIdentifier(ctx context.Context, conn db.Querier, identifier *core.DID) (*domain.IdentityState, error) {
	row := conn.QueryRow(ctx, `SELECT state_id, identifier, state, root_of_roots, claims_tree_root, 
       revocation_tree_root, block_timestamp, block_number, tx_id, previous_state, status, modified_at, created_at 
FROM identity_states
WHERE identifier=$1 AND status = 'confirmed' ORDER BY state_id DESC LIMIT 1`, identifier.String())
	state := domain.IdentityState{}
	if err := row.Scan(&state.StateID,
		&state.Identifier,
		&state.State,
		&state.RootOfRoots,
		&state.ClaimsTreeRoot,
		&state.RevocationTreeRoot,
		&state.BlockTimestamp,
		&state.BlockNumber,
		&state.TxID,
		&state.PreviousState,
		&state.Status,
		&state.ModifiedAt,
		&state.CreatedAt); err != nil {
		return nil, fmt.Errorf("error trying to get latest state:%w", err)
	}

	return &state, nil
}

// GetStatesByStatus returns states which are not transated
func (isr *identityState) GetStatesByStatus(ctx context.Context, conn db.Querier, status domain.IdentityStatus) ([]domain.IdentityState, error) {
	rows, err := conn.Query(ctx, `SELECT state_id, identifier, state, root_of_roots, claims_tree_root, revocation_tree_root, block_timestamp, block_number, 
       tx_id, previous_state, status, modified_at, created_at 
	FROM identity_states WHERE status = $1 and previous_state IS NOT NULL`, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states := []domain.IdentityState{}
	for rows.Next() {
		var state domain.IdentityState
		if err := rows.Scan(&state.StateID,
			&state.Identifier,
			&state.State,
			&state.RootOfRoots,
			&state.ClaimsTreeRoot,
			&state.RevocationTreeRoot,
			&state.BlockTimestamp,
			&state.BlockNumber,
			&state.TxID,
			&state.PreviousState,
			&state.Status,
			&state.ModifiedAt,
			&state.CreatedAt); err != nil {
			return nil, err
		}
		states = append(states, state)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return states, nil
}

func (isr *identityState) UpdateState(ctx context.Context, conn db.Querier, state *domain.IdentityState) (int64, error) {
	tag, err := conn.Exec(ctx, `UPDATE identity_states 
		SET block_timestamp=$1, block_number=$2, tx_id=$3, status=$4 WHERE state = $5 `,
		state.BlockTimestamp, state.BlockNumber, state.TxID, state.Status, state.State)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}
