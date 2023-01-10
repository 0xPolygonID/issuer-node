package repositories

import (
	"context"
	"fmt"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type identityState struct{}

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
