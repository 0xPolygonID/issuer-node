package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iden3/go-circuits/v2"
	"github.com/iden3/go-rapidsnark/prover"
	"github.com/iden3/go-rapidsnark/types"
	"github.com/iden3/go-rapidsnark/witness/v2"
	"github.com/iden3/go-rapidsnark/witness/wazero"

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
func (s *NativeProverService) Generate(ctx context.Context, inputs json.RawMessage, circuitName string) (*types.ZKProof, error) {
	wasm, err := s.config.CircuitsLoader.LoadWasm(circuits.CircuitID(circuitName))
	if err != nil {
		return nil, err
	}

	calc, err := witness.NewCalculator(wasm, witness.WithWasmEngine(wazero.NewCircom2WZWitnessCalculator))
	if err != nil {
		log.Error(ctx, "can't create witness calculator", "err", err)
		return nil, errors.New("can't create witness calculator")
	}

	parsedInputs, err := witness.ParseInputs(inputs)
	if err != nil {
		log.Error(ctx, "can't parse inputs", "err", err)
		return nil, errors.Join(err, errors.New("can't parse inputs"))
	}

	wtnsBytes, err := calc.CalculateWTNSBin(parsedInputs, true)
	if err != nil {
		log.Error(ctx, "can't generate witnesses", "err", err)
		return nil, errors.New("can't generate witnesses")
	}

	provingKey, err := s.config.CircuitsLoader.LoadProvingKey(circuits.CircuitID(circuitName))
	if err != nil {
		return nil, err
	}
	p, err := prover.Groth16Prover(provingKey, wtnsBytes)
	if err != nil {
		log.Error(ctx, "can't generate proof", "err", err)
		return nil, fmt.Errorf("can't generate proof: %w", err)
	}

	return p, nil
}
