package packagemanager

import (
	"context"
	"fmt"

	"github.com/iden3/contracts-abi/state/go/abi"
	"github.com/iden3/go-circuits/v2"
	"github.com/iden3/go-jwz/v2"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"

	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/loaders"
)

// New initializes the iden3comm package manager
func New(ctx context.Context, ethStateContracts map[string]*abi.State, circuitsPath string, didResolverHandler packers.DIDResolverHandlerFunc) (*iden3comm.PackageManager, error) {
	circuitsLoaderService := loaders.NewCircuits(circuitsPath)
	authV2Set, err := circuitsLoaderService.Load(circuits.AuthV2CircuitID)
	if err != nil {
		return nil, fmt.Errorf("failed upload circuits files: %w", err)
	}

	authV3Set, err := circuitsLoaderService.Load(circuits.AuthV3CircuitID)
	if err != nil {
		return nil, fmt.Errorf("failed upload circuits files: %w", err)
	}

	authV3_8_32Set, err := circuitsLoaderService.Load(circuits.AuthV3_8_32CircuitID)
	if err != nil {
		return nil, fmt.Errorf("failed upload circuits files: %w", err)
	}

	verifications := make(map[jwz.ProvingMethodAlg]packers.VerificationParams)
	verifications[jwz.AuthV2Groth16Alg] = packers.NewVerificationParams(authV2Set.VerificationKey, stateVerificationHandler(ethStateContracts))
	verifications[jwz.AuthV3Groth16Alg] = packers.NewVerificationParams(authV3Set.VerificationKey, stateVerificationHandler(ethStateContracts))
	verifications[jwz.AuthV3_8_32Groth16Alg] = packers.NewVerificationParams(authV3_8_32Set.VerificationKey, stateVerificationHandler(ethStateContracts))

	zkpPacker := packers.NewZKPPacker(nil, verifications)
	jwsPacker := packers.NewJWSPacker(didResolverHandler, nil)
	packageManager := iden3comm.NewPackageManager()
	err = packageManager.RegisterPackers(zkpPacker, &packers.PlainMessagePacker{}, jwsPacker)
	if err != nil {
		log.Error(ctx, "failed to register packers", "error", err)
		return nil, err
	}

	return packageManager, err
}
