package services

import (
	"context"
	"fmt"

	core "github.com/iden3/go-iden3-core"
	"github.com/jackc/pgx/v4"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type identityState struct {
	identityStateRepository ports.IdentityStateRepository
	claimsRepository        ports.ClaimsRepository
	revocationRepository    ports.RevocationRepository
	mtService               ports.MtService
	storage                 *db.Storage
}

// NewIdentityState creates a new identity
func NewIdentityState(identityStateRepository ports.IdentityStateRepository, mtservice ports.MtService, claimsRepository ports.ClaimsRepository, revocationRepository ports.RevocationRepository, storage *db.Storage) ports.IdentityStateService {
	return &identityState{
		identityStateRepository: identityStateRepository,
		claimsRepository:        claimsRepository,
		revocationRepository:    revocationRepository,
		mtService:               mtservice,
		storage:                 storage,
	}
}

func (is *identityState) UpdateIdentityClaims(ctx context.Context, did *core.DID) (*domain.IdentityState, error) {
	newState := &domain.IdentityState{
		Identifier: did.String(),
		Status:     domain.StatusCreated,
	}

	err := is.storage.Pgx.BeginFunc(ctx,
		func(tx pgx.Tx) error {
			iTrees, err := is.mtService.GetIdentityMerkleTrees(ctx, tx, did)
			if err != nil {
				return err
			}

			previousState, err := is.identityStateRepository.GetLatestStateByIdentifier(ctx, tx, did)
			if err != nil {
				return fmt.Errorf("error getting the identifier last state: %w", err)
			}

			lc, err := is.claimsRepository.ListByState(ctx, tx, did, nil)
			if err != nil {
				return fmt.Errorf("error getting the states: %w", err)
			}

			for i := range lc {
				err = iTrees.AddClaim(ctx, &lc[i])
				if err != nil {
					return err
				}
			}

			err = populateIdentityState(ctx, iTrees, newState, previousState)
			if err != nil {
				return err
			}

			err = is.update(ctx, tx, did, *newState)
			if err != nil {
				return err
			}

			_, err = is.revocationRepository.UpdateStatus(ctx, tx, did)
			if err != nil {
				return err
			}

			err = is.identityStateRepository.Save(ctx, tx, *newState)
			if err != nil {
				return fmt.Errorf("error saving new identity state: %w", err)
			}

			return err
		},
	)

	return newState, err
}

func (is *identityState) update(ctx context.Context, conn db.Querier, id *core.DID, currentState domain.IdentityState) error {
	claims, err := is.claimsRepository.ListByState(ctx, conn, id, nil)
	if err != nil {
		return err
	}

	// do not have claims to process
	if len(claims) == 0 {
		return nil
	}

	for i := range claims {
		var err error
		claims[i].IdentityState = currentState.State

		affected, err := is.claimsRepository.UpdateState(ctx, is.storage.Pgx, &claims[i])
		if err != nil {
			return fmt.Errorf("can't update claim: %w", err)
		}
		if affected == 0 {
			return fmt.Errorf("claim has not been updated %v", claims[i])
		}
	}

	return nil
}
