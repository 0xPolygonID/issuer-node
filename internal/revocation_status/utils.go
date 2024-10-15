package revocation_status

import (
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-core/v2/w3c"
)

func buildRHSRevocationURL(host, issuerState string) string {
	if issuerState == "" {
		return host
	}
	return fmt.Sprintf("%s/node?state=%s", host, issuerState)
}

func buildIden3OnchainSMTProofURL(issuerDID w3c.DID, nonce uint64, contractAddress ethcommon.Address, chainID string, stateHex string) string {
	return fmt.Sprintf("%s/credentialStatus?revocationNonce=%v&contractAddress=%s:%s&state=%s", issuerDID.String(), nonce, chainID, contractAddress.Hex(), stateHex)
}
