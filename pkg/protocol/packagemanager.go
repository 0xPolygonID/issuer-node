package protocol

import (
	"context"
	"fmt"

	"github.com/iden3/contracts-abi/state/go/abi"
	"github.com/iden3/go-circuits/v2"
	"github.com/iden3/go-jwz/v2"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"

	"github.com/polygonid/sh-id-platform/pkg/loaders"
)

// InitPackageManager initializes the iden3comm package manager
func InitPackageManager(_ context.Context, ethStateContracts map[string]*abi.State, circuitsPath string) (*iden3comm.PackageManager, error) {
	circuitsLoaderService := loaders.NewCircuits(circuitsPath)
	authV2Set, err := circuitsLoaderService.Load(circuits.AuthV2CircuitID)
	if err != nil {
		return nil, fmt.Errorf("failed upload circuits files: %w", err)
	}

	verifications := make(map[jwz.ProvingMethodAlg]packers.VerificationParams)
	verifications[jwz.AuthV2Groth16Alg] = packers.NewVerificationParams(authV2Set.VerificationKey, stateVerificationHandler(ethStateContracts))

	zkpPackerV2 := packers.NewZKPPacker(nil, verifications)
	packageManager := iden3comm.NewPackageManager()
	err = packageManager.RegisterPackers(zkpPackerV2, &packers.PlainMessagePacker{})
	if err != nil {
		return nil, err
	}

	return packageManager, err
}
