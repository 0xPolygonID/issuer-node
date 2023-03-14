package schema

import (
	"encoding/json"
	"fmt"

	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/jackc/pgtype"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// FromClaimModelToW3CCredential JSON-LD response base on claim
func FromClaimModelToW3CCredential(claim domain.Claim) (*verifiable.W3CCredential, error) {
	var cred verifiable.W3CCredential

	err := json.Unmarshal(claim.Data.Bytes, &cred)
	if err != nil {
		return nil, err
	}
	if claim.CredentialStatus.Status == pgtype.Null {
		return nil, fmt.Errorf("credential status is not set")
	}

	proofs := make(verifiable.CredentialProofs, 0)

	var signatureProof *verifiable.BJJSignatureProof2021
	if claim.SignatureProof.Status != pgtype.Null {
		err = claim.SignatureProof.AssignTo(&signatureProof)
		if err != nil {
			return nil, err
		}
		proofs = append(proofs, signatureProof)
	}

	var mtpProof *verifiable.Iden3SparseMerkleTreeProof

	if claim.MTPProof.Status != pgtype.Null {
		err = claim.MTPProof.AssignTo(&mtpProof)
		if err != nil {
			return nil, err
		}
		proofs = append(proofs, mtpProof)

	}
	cred.Proof = proofs

	return &cred, nil
}
