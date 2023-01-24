package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/iden3/go-circuits"
	"github.com/iden3/go-rapidsnark/prover"
	"github.com/iden3/go-rapidsnark/witness"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/loaders"
)

// NativeProverConfig represents native prover config
type NativeProverConfig struct {
	CircuitsLoader *loaders.Circuits
}

// NativeProverService service responsible for native zk generation
type NativeProverService struct {
	config *NativeProverConfig
}

// NewNativeProverService new prover service that works with zero knowledge proofs
func NewNativeProverService(config *NativeProverConfig) *NativeProverService {
	return &NativeProverService{config: config}
}

// Generate calls prover-server for proof generation
func (s *NativeProverService) Generate(ctx context.Context, inputs json.RawMessage, circuitName string) (*domain.FullProof, error) {
	wasm, err := s.config.CircuitsLoader.LoadWasm(circuits.CircuitID(circuitName))
	if err != nil {
		return nil, err
	}

	calc, err := witness.NewCircom2WitnessCalculator(wasm, true)
	if err != nil {
		log.Error(ctx, "can't create witness calculator", err)
		return nil, fmt.Errorf("can't create witness calculator: %w", err)
	}

	parsedInputs, err := witness.ParseInputs(inputs)
	if err != nil {
		return nil, err
	}

	wtnsBytes, err := calc.CalculateWTNSBin(parsedInputs, true)
	if err != nil {
		log.Error(ctx, "can't generate witnesses", err)
		return nil, fmt.Errorf("can't generate witnesses: %w", err)
	}

	provingKey, err := s.config.CircuitsLoader.LoadProvingKey(circuits.CircuitID(circuitName))
	if err != nil {
		return nil, err
	}
	p, err := prover.Groth16Prover(provingKey, wtnsBytes)
	if err != nil {
		log.Error(ctx, "can't generate proof", err)
		return nil, fmt.Errorf("can't generate proof: %w", err)
	}
	// TODO: get rid of models.Proof structure
	return &domain.FullProof{
		Proof: &domain.ZKProof{
			A:        p.Proof.A,
			B:        p.Proof.B,
			C:        p.Proof.C,
			Protocol: p.Proof.Protocol,
		},
		PubSignals: p.PubSignals,
	}, nil
}
