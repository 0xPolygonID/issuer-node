package protocol

import (
	"context"
	"fmt"
	"math/big"

	"github.com/iden3/contracts-abi/state/go/abi"
	"github.com/iden3/go-circuits/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-jwz/v2"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"

	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/pkg/loaders"
)

// InitPackageManager initializes the iden3comm package manager
func InitPackageManager(ctx context.Context, ethStateContracts map[string]*abi.State, zkProofService ports.ProofService, circuitsPath string) (*iden3comm.PackageManager, error) {
	circuitsLoaderService := loaders.NewCircuits(circuitsPath)

	authV2Set, err := circuitsLoaderService.Load(circuits.AuthV2CircuitID)
	if err != nil {
		return nil, fmt.Errorf("failed upload circuits files: %w", err)
	}

	provers := make(map[jwz.ProvingMethodAlg]packers.ProvingParams)
	pParams := packers.ProvingParams{
		DataPreparer: prepareAuthInputs(ctx, zkProofService),
		ProvingKey:   authV2Set.ProofKey,
		Wasm:         authV2Set.Wasm,
	}
	provers[jwz.AuthV2Groth16Alg] = pParams

	verifications := make(map[jwz.ProvingMethodAlg]packers.VerificationParams)
	verifications[jwz.AuthV2Groth16Alg] = packers.NewVerificationParams(authV2Set.VerificationKey,
		stateVerificationHandler(ethStateContracts))

	zkpPackerV2 := packers.NewZKPPacker(
		provers,
		verifications,
	)

	// TODO: Why jwsPacker is not defined here?

	packageManager := iden3comm.NewPackageManager()

	err = packageManager.RegisterPackers(zkpPackerV2, &packers.PlainMessagePacker{})
	if err != nil {
		return nil, err
	}

	return packageManager, err
}

func prepareAuthInputs(ctx context.Context, proofService ports.ProofService) packers.DataPreparerHandlerFunc {
	return func(hash []byte, id *w3c.DID, circuitID circuits.CircuitID) ([]byte, error) {
		q := ports.Query{}
		q.CircuitID = string(circuitID)
		q.Challenge = new(big.Int).SetBytes(hash)
		circuitInputs, _, err := proofService.PrepareInputs(ctx, id, q)
		if err != nil {
			return nil, err
		}
		return circuitInputs, nil
	}
}
