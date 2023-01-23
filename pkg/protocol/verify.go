package protocol

import (
	"encoding/json"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/iden3/go-circuits"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/iden3comm/packers"
	"github.com/pkg/errors"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/pkg/blockchain/eth"
)

// ErrStateNotFound issuer state is genesis state.
var (
	ErrStateNotFound = errors.New("Identity does not exist")
)

func stateVerificationHandler(ethStateContract *eth.State) packers.VerificationHandlerFunc {
	return func(id circuits.CircuitID, pubsignals []string) error {
		switch id {
		case circuits.AuthCircuitID:
			return authCircuitStateVerification(ethStateContract, pubsignals)
		case circuits.AuthV2CircuitID:
			return authV2CircuitStateVerification(ethStateContract, pubsignals)
		default:
			return errors.Errorf("'%s' unknow circuit ID", id)
		}
	}
}

// authCircuitStateVerification `auth` circuit state verification
func authCircuitStateVerification(contract *eth.State, pubsignals []string) error {
	bytePubsig, err := json.Marshal(pubsignals)
	if err != nil {
		return err
	}

	authPubSignals := circuits.AuthPubSignals{}
	err = authPubSignals.PubSignalsUnmarshal(bytePubsig)
	if err != nil {
		return err
	}

	userID := authPubSignals.UserID
	userState := authPubSignals.UserState.BigInt()

	userDID, err := core.ParseDIDFromID(*userID)
	if err != nil {
		return err
	}

	stateInfo, err := contract.GetStateInfoById(&bind.CallOpts{}, userID.BigInt())
	if err != nil {
		return err
	}

	if err != nil && strings.Contains(err.Error(), ErrStateNotFound.Error()) {
		err = common.CheckGenesisStateDID(userDID, userState)
		if err != nil {
			return err
		}
		return nil
	}

	if stateInfo.State.String() != userState.String() {
		return errors.New("not latest state")
	}

	return nil
}

// authV2CircuitStateVerification `authV2` circuit state verification
func authV2CircuitStateVerification(contract *eth.State, pubsignals []string) error {
	bytePubsig, err := json.Marshal(pubsignals)
	if err != nil {
		return err
	}

	authPubSignals := circuits.AuthV2PubSignals{}
	err = authPubSignals.PubSignalsUnmarshal(bytePubsig)
	if err != nil {
		return err
	}

	globalState := authPubSignals.GlobalRoot.BigInt()
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
