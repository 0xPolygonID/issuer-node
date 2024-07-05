package protocol

import (
	"context"
	"crypto"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/iden3/contracts-abi/state/go/abi"
	"github.com/iden3/go-circuits/v2"
	auth "github.com/iden3/go-iden3-auth/v2"
	"github.com/iden3/go-jwz/v2"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"

	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/loaders"
)

// InitPackageManager initializes the iden3comm package manager
func InitPackageManager(ctx context.Context, ethStateContracts map[string]*abi.State, circuitsPath string, keyStore *kms.KMS) (*iden3comm.PackageManager, error) {
	circuitsLoaderService := loaders.NewCircuits(circuitsPath)
	authV2Set, err := circuitsLoaderService.Load(circuits.AuthV2CircuitID)
	if err != nil {
		return nil, fmt.Errorf("failed upload circuits files: %w", err)
	}

	verifications := make(map[jwz.ProvingMethodAlg]packers.VerificationParams)
	verifications[jwz.AuthV2Groth16Alg] = packers.NewVerificationParams(authV2Set.VerificationKey, stateVerificationHandler(ethStateContracts))

	zkpPackerV2 := packers.NewZKPPacker(nil, verifications)

	// right now we can sign with BJJ key
	jwsPacker := packers.NewJWSPacker(auth.UniversalDIDResolver, kMSBJJAdapterSigner(keyStore))

	packageManager := iden3comm.NewPackageManager()
	err = packageManager.RegisterPackers(zkpPackerV2, &packers.PlainMessagePacker{}, jwsPacker)
	if err != nil {
		log.Error(ctx, "failed to register packers", "error", err)
		return nil, err
	}

	return packageManager, err
}

func kMSBJJAdapterSigner(keyStore *kms.KMS) packers.SignerResolverHandlerFunc {
	return func(kid string) (crypto.Signer, error) {
		parts := strings.Split(kid, "#")
		idBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, err
		}
		id := string(idBytes)

		kmsAdapter := KMSBJJJWSAdapter{
			kms:   keyStore,
			keyID: kms.KeyID{Type: kms.KeyTypeBabyJubJub, ID: id},
		}

		return kmsAdapter, nil
	}
}
