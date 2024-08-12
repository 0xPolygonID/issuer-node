package protocol

import (
	"context"
	"encoding/json"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/iden3/contracts-abi/state/go/abi"
	"github.com/iden3/go-circuits/v2"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/pkg/errors"
	"github.com/polygonid/sh-id-platform/internal/log"
)

// ErrStateNotFound issuer state is genesis state.
var (
	ErrStateNotFound = errors.New("Identity does not exist")
)

func stateVerificationHandler(ethStateContracts map[string]*abi.State, contracts, rpcs map[string]string) packers.VerificationHandlerFunc {
	return func(id circuits.CircuitID, pubsignals []string) error {
		switch id {
		case circuits.AuthV2CircuitID:
			return authV2CircuitStateVerification(ethStateContracts, pubsignals, contracts, rpcs)
		default:
			return errors.Errorf("'%s' unknow circuit ID", id)
		}
	}
}

// authV2CircuitStateVerification `authV2` circuit state verification
func authV2CircuitStateVerification(contracts map[string]*abi.State, pubsignals []string, rawContracts, rpcs map[string]string) error {
	bytePubsig, err := json.Marshal(pubsignals)
	if err != nil {
		return err
	}

	authPubSignals := circuits.AuthV2PubSignals{}
	err = authPubSignals.PubSignalsUnmarshal(bytePubsig)
	if err != nil {
		return err
	}

	did, err := core.ParseDIDFromID(*authPubSignals.UserID)
	if err != nil {
		return err
	}

	chainID, err := core.ChainIDfromDID(*did)
	if err != nil {
		return errors.Errorf("state verification error: %v", err)
	}

	chainIDStr := strconv.Itoa(int(chainID))
	contract, ok := contracts[chainIDStr]
	if !ok {
		return errors.Errorf("not supported blockchain for chainID '%s'", chainIDStr)
	}

	log.Info(context.Background(), "verifying authV2 circuit state", "chainID", chainIDStr, "did", did.String(), "contract_address", rawContracts[chainIDStr], "rpc", rpcs[chainIDStr])

	globalState := authPubSignals.GISTRoot.BigInt()
	globalStateInfo, err := contract.GetGISTRootInfo(&bind.CallOpts{}, globalState)
	if err != nil {
		return err
	}

	if globalState.Cmp(globalStateInfo.Root) != 0 {
		return errors.Errorf("invalid global state info in the smart contract, expected root %s, got %s", globalState.String(), globalStateInfo.Root.String())
	}

	if (big.NewInt(0)).Cmp(globalStateInfo.ReplacedByRoot) != 0 && time.Since(time.Unix(globalStateInfo.ReplacedAtTimestamp.Int64(), 0)) > time.Minute*15 {
		return errors.Errorf("global state is too old, replaced timestamp is %v", globalStateInfo.ReplacedAtTimestamp.Int64())
	}

	return nil
}
