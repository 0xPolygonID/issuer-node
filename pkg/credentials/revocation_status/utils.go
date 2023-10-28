package revocation_status

import (
	"fmt"
	"net/url"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-core/v2/w3c"
)

func buildRHSRevocationURL(host, issuerState string) string {
	if issuerState == "" {
		return host
	}
	return fmt.Sprintf("%s/node?state=%s", host, issuerState)
}

func buildDirectRevocationURL(host, issuerDID string, nonce uint64, singleIssuer bool) string {
	if singleIssuer {
		return fmt.Sprintf("%s/v1/credentials/revocation/status/%d",
			host, nonce)
	}
	return fmt.Sprintf("%s/v1/%s/claims/revocation/status/%d", host, url.QueryEscape(issuerDID), nonce)
}

func buildIden3OnchainSMTProofURL(issuerDID w3c.DID, nonce uint64, contractAddress ethcommon.Address, chainID string, stateHex string) string {
	return fmt.Sprintf("%s/credentialStatus?revocationNonce=%v&contractAddress=%s:%s&state=%s", issuerDID.String(), nonce, chainID, contractAddress.Hex(), stateHex)
}
