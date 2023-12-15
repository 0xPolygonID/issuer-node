package gateways

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/iden3/go-rapidsnark/types"

	"github.com/polygonid/issuer-node/internal/config"
	"github.com/polygonid/issuer-node/internal/core/domain"
	"github.com/polygonid/issuer-node/internal/core/ports"
	"github.com/polygonid/issuer-node/internal/core/services"
	"github.com/polygonid/issuer-node/internal/log"
	client "github.com/polygonid/issuer-node/pkg/http"
	"github.com/polygonid/issuer-node/pkg/loaders"
)

// ProverConfig represents prover server config
type ProverConfig struct {
	ServerURL       string
	ResponseTimeout time.Duration
}

// ProverService service responsible for zk generation
type ProverService struct {
	proverConfig *ProverConfig
}

// NewProverService new prover service that works with zero knowledge proofs
func NewProverService(config *ProverConfig) *ProverService {
	return &ProverService{proverConfig: config}
}

// NewProver returns a new prover with the given configuration.
// If NativeProofGenerationEnabled is true it will return a NativeProverService
// If NativeProofGenerationEnabled is false it will return an external ProverService
func NewProver(ctx context.Context, config *config.Configuration, circuitLoaderService *loaders.Circuits) ports.ZKGenerator {
	log.Info(ctx, "native prover enabled", "enabled", config.NativeProofGenerationEnabled)
	if config.NativeProofGenerationEnabled {
		proverConfig := &services.NativeProverConfig{
			CircuitsLoader: circuitLoaderService,
		}
		return services.NewNativeProverService(proverConfig)
	}

	proverConfig := &ProverConfig{
		ServerURL:       config.Prover.ServerURL,
		ResponseTimeout: config.Prover.ResponseTimeout,
	}

	return NewProverService(proverConfig)
}

// Verify calls prover server for proof verification
func (s *ProverService) Verify(ctx context.Context, zkp *domain.FullProof, circuitName string) (bool, error) {
	r := struct {
		ZKP         *domain.FullProof `json:"zkp"`
		CircuitName string            `json:"circuit_name"`
	}{
		ZKP:         zkp,
		CircuitName: circuitName,
	}

	proverReq, err := json.Marshal(r)
	if err != nil {
		log.Error(ctx, "can't json encode request: ", "err", err)
		return false, err
	}

	url := s.proverConfig.ServerURL + "/api/v1/proof/verify"

	res, err := client.NewClient(http.Client{Timeout: s.proverConfig.ResponseTimeout}).Post(ctx, url, proverReq)
	if err != nil {
		return false, err
	}
	verifyResp := struct {
		Valid bool `json:"valid"`
	}{}
	err = json.Unmarshal(res, &verifyResp)
	if err != nil {
		log.Error(ctx, "failed to unmarshal proof verification result:", "err", err)
		return false, err
	}

	return verifyResp.Valid, nil
}

// Generate calls prover-server for proof generation
func (s *ProverService) Generate(ctx context.Context, inputs json.RawMessage, circuitName string) (*types.ZKProof, error) {
	var zkp types.ZKProof

	r := struct {
		Inputs      json.RawMessage `json:"inputs"`
		CircuitName string          `json:"circuit_name"`
	}{
		Inputs:      inputs,
		CircuitName: circuitName,
	}

	req, err := json.Marshal(r)
	if err != nil {
		log.Error(ctx, "can't json encode request:", "err", err)
		return nil, err
	}

	url := s.proverConfig.ServerURL + "/api/v1/proof/generate"

	res, err := client.NewClient(http.Client{Timeout: s.proverConfig.ResponseTimeout}).Post(ctx, url, req)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(res, &zkp)
	if err != nil {
		log.Error(ctx, "failed to unmarshal proof generation result: ", "err", err)
		return nil, err
	}

	return &zkp, nil
}
