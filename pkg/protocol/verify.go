package protocol

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/iden3/contracts-abi/state/go/abi"
	"github.com/iden3/go-circuits"
	"github.com/iden3/iden3comm/packers"
	"github.com/pkg/errors"
)

// ErrStateNotFound issuer state is genesis state.
var (
	ErrStateNotFound = errors.New("Identity does not exist")
)

func stateVerificationHandler(ethStateContract *abi.State) packers.VerificationHandlerFunc {
	return func(id circuits.CircuitID, pubsignals []string) error {
		switch id {
		case circuits.AuthV2CircuitID:
			return authV2CircuitStateVerification(ethStateContract, pubsignals)
		default:
			return errors.Errorf("'%s' unknow circuit ID", id)
		}
	}
}

// authV2CircuitStateVerification `authV2` circuit state verification
func authV2CircuitStateVerification(contract *abi.State, pubsignals []string) error {
	bytePubsig, err := json.Marshal(pubsignals)
	if err != nil {
		return err
	}

	authPubSignals := circuits.AuthV2PubSignals{}
	err = authPubSignals.PubSignalsUnmarshal(bytePubsig)
	if err != nil {
		return err
	}

	globalState := authPubSignals.GISTRoot.BigInt()
	globalStateInfo, err := contract.GetGISTRootInfo(&bind.CallOpts{}, globalState)
	if err != nil {
		return err
	}
	if (big.NewInt(0)).Cmp(globalStateInfo.CreatedAtTimestamp) == 0 {
		return errors.Errorf("root %s doesn't exist in smart contract", globalState.String())
	}
	if globalState.Cmp(globalStateInfo.Root) != 0 {
		return errors.Errorf("invalid global state info in the smart contract, expected root %s, got %s", globalState.String(), globalStateInfo.Root.String())
	}

	if (big.NewInt(0)).Cmp(globalStateInfo.ReplacedByRoot) != 0 && time.Since(time.Unix(globalStateInfo.ReplacedAtTimestamp.Int64(), 0)) > time.Minute*15 {
		return errors.Errorf("global state is too old, replaced timestamp is %v", globalStateInfo.ReplacedAtTimestamp.Int64())
	}

	return nil
}
