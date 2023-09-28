package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/google/uuid"
	"github.com/iden3/contracts-abi/state/go/abi"
	"github.com/iden3/go-circuits/v2"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree-sql/v2"
	jsonSuite "github.com/iden3/go-schema-processor/v2/json"
	"github.com/iden3/go-schema-processor/v2/merklize"
	"github.com/iden3/go-schema-processor/v2/processor"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/jackc/pgx/v4"
	"github.com/piprate/json-gold/ld"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/jsonschema"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/credentials/signature/circuit/signer"
)

// ErrAllClaimsRevoked all claims are revoked.
var (
	ErrAllClaimsRevoked = errors.New("all claims are revoked")
)

// Proof service generates and validates ZK zk
type Proof struct {
	claimSrv         ports.ClaimsService
	revocationSrv    ports.RevocationService
	identitySrv      ports.IdentityService
	mtService        ports.MtService
	claimsRepository ports.ClaimsRepository
	keyProvider      *kms.KMS
	storage          *db.Storage
	stateContract    *abi.State
	schemaLoader     loader.DocumentLoader
}

// NewProofService init proof service
func NewProofService(claimSrv ports.ClaimsService, revocationSrv ports.RevocationService, identitySrv ports.IdentityService, mtService ports.MtService, claimsRepository ports.ClaimsRepository, keyProvider *kms.KMS, storage *db.Storage, stateContract *abi.State, ld ld.DocumentLoader) ports.ProofService {
	return &Proof{
		claimSrv:         claimSrv,
		revocationSrv:    revocationSrv,
		identitySrv:      identitySrv,
		mtService:        mtService,
		claimsRepository: claimsRepository,
		keyProvider:      keyProvider,
		storage:          storage,
		stateContract:    stateContract,
		schemaLoader:     ld,
	}
}

// PrepareInputs prepare inputs for circuit.
//
//nolint:gocyclo // refactor later to avoid big PR.
func (p *Proof) PrepareInputs(ctx context.Context, identifier *w3c.DID, query ports.Query) (inputs []byte, vp map[string]any, err error) {
	var circuitInputs circuits.InputsMarshaller
	switch circuits.CircuitID(query.CircuitID) {
	case circuits.AtomicQuerySigV2CircuitID:
		circuitInputs, vp, err = p.prepareAtomicQuerySigV2Circuit(ctx,
			identifier, query)
		if err != nil {
			return nil, nil, err
		}

	case circuits.AtomicQueryMTPV2CircuitID:
		circuitInputs, vp, err = p.prepareAtomicQueryMTPV2Circuit(ctx,
			identifier, query)
		if err != nil {
			return nil, nil, err
		}
	//nolint:all
	//case circuits.AtomicQueryV3CircuitID:
	//	circuitInputs, vp, err = p.prepareAtomicQueryV3Circuit(ctx,
	//		identifier, query)
	//	if err != nil {
	//		return nil, nil, err
	//	}

	case circuits.AuthV2CircuitID:
		circuitInputs, err = p.prepareAuthV2Circuit(ctx, identifier, query.Challenge)
		if err != nil {
			return nil, nil, err
		}

	default:
		log.Error(ctx, "circuit with id is not supported", "circuitID", query.CircuitID)
		return nil, nil, fmt.Errorf("circuit with id %s is not supported", query.CircuitID)
	}

	inputs, err = circuitInputs.InputsMarshal()
	if err != nil {
		return nil, nil, err
	}

	return inputs, vp, nil
}

func (p *Proof) prepareAtomicQuerySigV2Circuit(ctx context.Context, did *w3c.DID, query ports.Query) (circuits.InputsMarshaller, map[string]any, error) {
	claim, claimNonRevProof, err := p.getClaimDataForAtomicQueryCircuit(ctx,
		did, query)
	if err != nil {
		return nil, nil, err
	}

	signatureProof, err := p.prepareBJJSignatureProof(ctx, claim)
	if err != nil {
		return nil, nil, err
	}

	circuitQuery, vp, err := p.toCircuitsQuery(ctx, *claim, query)
	if err != nil {
		return nil, nil, err
	}

	issuerDID, err := w3c.ParseDID(claim.Issuer)
	if err != nil {
		log.Error(ctx, "failed to parse issuer did", "err", err)
		return nil, nil, err
	}

	issuerID, err := core.IDFromDID(*issuerDID)
	if err != nil {
		log.Error(ctx, "failed to get issuer id", "err", err)
		return nil, nil, err
	}

	id, err := core.IDFromDID(*did)
	if err != nil {
		log.Error(ctx, "failed to get id", "err", err)
		return nil, nil, err
	}

	inputs := circuits.AtomicQuerySigV2Inputs{
		RequestID:                query.Challenge,
		ID:                       &id,
		ProfileNonce:             big.NewInt(0),
		ClaimSubjectProfileNonce: big.NewInt(0),
		Claim: circuits.ClaimWithSigProof{
			IssuerID:       &issuerID,
			Claim:          claim.CoreClaim.Get(),
			NonRevProof:    *claimNonRevProof,
			SignatureProof: signatureProof,
		},
		Query:                    circuitQuery,
		CurrentTimeStamp:         time.Now().Unix(),
		SkipClaimRevocationCheck: query.SkipClaimRevocationCheck,
	}

	return inputs, vp, nil

	//claim, claimNonRevProof, err := p.getClaimDataForAtomicQueryCircuit(ctx, did, query)
	//if err != nil {
	//	return nil, nil, err
	//}
	//
	//sigProof, err := claim.GetBJJSignatureProof2021()
	//if err != nil {
	//	return nil, nil, err
	//}
	//
	//sig, err := signer.BJJSignatureFromHexString(sigProof.Signature)
	//if err != nil {
	//	return nil, nil, err
	//}
	//
	//issuerDID, err := w3c.ParseDID(claim.Issuer)
	//if err != nil {
	//	return nil, nil, err
	//}
	//
	//issuerAuthNonRevProof, err := p.callNonRevProof(ctx, sigProof.IssuerData, issuerDID)
	//if err != nil {
	//	return nil, nil, err
	//}
	//
	//circuitQuery, err := p.toCircuitsQuery(ctx, *claim, query)
	//if err != nil {
	//	return nil, nil, err
	//}
	//
	//authClaim := &core.Claim{}
	//err = authClaim.FromHex(sigProof.IssuerData.AuthCoreClaim)
	//if err != nil {
	//	return nil, nil, err
	//}
	//
	//sig1 := circuits.BJJSignatureProof{
	//	Signature:       sig,
	//	IssuerAuthClaim: authClaim,
	//	IssuerAuthIncProof: circuits.MTProof{
	//		Proof: sigProof.IssuerData.MTP,
	//		TreeState: circuits.TreeState{
	//			State:          common.StrMTHex(sigProof.IssuerData.State.Value),
	//			ClaimsRoot:     common.StrMTHex(sigProof.IssuerData.State.ClaimsTreeRoot),
	//			RevocationRoot: common.StrMTHex(sigProof.IssuerData.State.RevocationTreeRoot),
	//			RootOfRoots:    common.StrMTHex(sigProof.IssuerData.State.RootOfRoots),
	//		},
	//	},
	//	IssuerAuthNonRevProof: issuerAuthNonRevProof,
	//}
	//
	//id, err := core.IDFromDID(*did)
	//if err != nil {
	//	return nil, nil, err
	//}
	//
	//inputs := circuits.AtomicQuerySigV2Inputs{
	//	RequestID:                big.NewInt(defaultAtomicCircuitsID),
	//	ID:                       &id,
	//	ProfileNonce:             big.NewInt(0),
	//	ClaimSubjectProfileNonce: big.NewInt(0),
	//	Claim: circuits.ClaimWithSigProof{
	//		IssuerID:       &id,
	//		Claim:          claim.CoreClaim.Get(),
	//		NonRevProof:    *claimNonRevProof,
	//		SignatureProof: sig1,
	//	},
	//	Query:                    circuitQuery,
	//	CurrentTimeStamp:         time.Now().Unix(),
	//	SkipClaimRevocationCheck: query.SkipClaimRevocationCheck,
	//}
	//
	//return inputs, claim, nil
}

func (p Proof) prepareBJJSignatureProof(ctx context.Context, claim *domain.Claim) (circuits.BJJSignatureProof, error) {
	sigProof, err := claim.GetBJJSignatureProof2021()
	if err != nil {
		return circuits.BJJSignatureProof{}, err
	}
	sig, err := signer.BJJSignatureFromHexString(sigProof.Signature)
	if err != nil {
		return circuits.BJJSignatureProof{}, err
	}

	issuerDID, err := w3c.ParseDID(claim.Issuer)
	if err != nil {
		return circuits.BJJSignatureProof{}, err
	}
	userDID, err := w3c.ParseDID(claim.OtherIdentifier)
	if err != nil {
		return circuits.BJJSignatureProof{}, err
	}

	issuerAuthNonRevProof, err := p.callNonRevProof(ctx, sigProof.IssuerData, issuerDID, userDID)
	if err != nil {
		return circuits.BJJSignatureProof{}, err
	}

	authClaim := &core.Claim{}
	err = authClaim.FromHex(sigProof.IssuerData.AuthCoreClaim)
	if err != nil {
		return circuits.BJJSignatureProof{}, err
	}

	return circuits.BJJSignatureProof{
		Signature:       sig,
		IssuerAuthClaim: authClaim,
		IssuerAuthIncProof: circuits.MTProof{
			Proof: sigProof.IssuerData.MTP,
			TreeState: circuits.TreeState{
				State:          common.StrMTHex(sigProof.IssuerData.State.Value),
				ClaimsRoot:     common.StrMTHex(sigProof.IssuerData.State.ClaimsTreeRoot),
				RevocationRoot: common.StrMTHex(sigProof.IssuerData.State.RevocationTreeRoot),
				RootOfRoots:    common.StrMTHex(sigProof.IssuerData.State.RootOfRoots),
			},
		},
		IssuerAuthNonRevProof: issuerAuthNonRevProof,
	}, nil
}

func (p *Proof) prepareAtomicQueryMTPV2Circuit(ctx context.Context, did *w3c.DID, query ports.Query) (circuits.InputsMarshaller, map[string]any, error) {
	claim, claimNonRevProof, err := p.getClaimDataForAtomicQueryCircuit(ctx, did, query)
	if err != nil {
		return nil, nil, err
	}

	claimInc, err := claim.GetCircuitIncProof()
	if err != nil {
		return nil, nil, err
	}

	circuitQuery, vp, err := p.toCircuitsQuery(ctx, *claim, query)
	if err != nil {
		return nil, nil, err
	}

	issuerDID, err := w3c.ParseDID(claim.Issuer)
	if err != nil {
		return nil, nil, err
	}

	issueID, err := core.IDFromDID(*issuerDID)
	if err != nil {
		return nil, nil, err
	}

	id, err := core.IDFromDID(*did)
	if err != nil {
		return nil, nil, err
	}

	inputs := circuits.AtomicQueryMTPV2Inputs{
		RequestID:                query.Challenge,
		ID:                       &id,
		ProfileNonce:             big.NewInt(0),
		ClaimSubjectProfileNonce: big.NewInt(0),
		Claim: circuits.ClaimWithMTPProof{
			IssuerID:    &issueID, // claim.Issuer,
			Claim:       claim.CoreClaim.Get(),
			NonRevProof: *claimNonRevProof,
			IncProof:    claimInc,
		},
		Query:                    circuitQuery,
		CurrentTimeStamp:         time.Now().Unix(),
		SkipClaimRevocationCheck: query.SkipClaimRevocationCheck,
	}

	return inputs, vp, nil
}

func (p *Proof) getClaimDataForAtomicQueryCircuit(ctx context.Context, identifier *w3c.DID, query ports.Query) (claim *domain.Claim, revStatus *circuits.MTProof, err error) {
	var claims []*domain.Claim

	if query.ClaimID != "" {
		// if claimID exist. Search by claimID.
		claimUUID, err := uuid.Parse(query.ClaimID)
		if err != nil {
			return nil, nil, err
		}
		var c *domain.Claim
		c, err = p.claimSrv.GetByID(ctx, identifier, claimUUID)
		if err != nil {
			return nil, nil, err
		}
		// we need to be sure that the hallmark selected by ID matches circuitQuery.
		claims = append(claims, c)
	} else {
		// if claimID NOT exist in request select all claims and filter it.
		claims, err = p.findClaimForQuery(ctx, identifier, query)
		if err != nil {
			return claim, nil, err
		}
	}

	var claimRs circuits.MTProof
	if query.SkipClaimRevocationCheck {
		claim = claims[0]
		rsClaim, err := p.checkRevocationStatus(ctx, claim)
		if err != nil {
			return claim, nil, err
		}
		claimRs = circuits.MTProof{
			TreeState: domain.RevocationStatusToTreeState(*rsClaim),
			Proof:     &rsClaim.MTP,
		}
	} else {
		claim, claimRs, err = p.findNonRevokedClaim(ctx, claims)
		if err != nil {
			return claim, nil, err
		}
	}
	return claim, &claimRs, nil
}

func (p *Proof) findClaimForQuery(ctx context.Context, identifier *w3c.DID, query ports.Query) ([]*domain.Claim, error) {
	var err error

	// TODO "query_value":    value,
	// TODO "query_operator": operator,
	filter := &ports.ClaimsFilter{SchemaType: query.SchemaType()}
	if !query.SkipClaimRevocationCheck {
		filter.Revoked = common.ToPointer(false)
	}

	claim, err := p.claimsRepository.GetAllByIssuerID(ctx, p.storage.Pgx, *identifier, filter)
	if errors.Is(err, repositories.ErrClaimDoesNotExist) {
		return nil, fmt.Errorf("claim with credential type %v was not found", query)
	}

	return claim, err
}

func (p *Proof) checkRevocationStatus(ctx context.Context, claim *domain.Claim) (*verifiable.RevocationStatus, error) {
	var (
		err     error
		claimRs *verifiable.RevocationStatus
	)

	var cs map[string]interface{}
	if err = json.Unmarshal(claim.CredentialStatus.Bytes, &cs); err != nil {
		return nil, fmt.Errorf("failed unmasrshal credentialStatus: %s", err)
	}
	issuerDID, err := w3c.ParseDID(claim.Issuer)
	if err != nil {
		return nil, err
	}

	userDID, err := w3c.ParseDID(claim.OtherIdentifier)
	if err != nil {
		return nil, err
	}

	sigProof, err := claim.GetBJJSignatureProof2021()
	if err != nil {
		return nil, err
	}

	claimRs, err = p.revocationSrv.Status(ctx, cs, userDID, issuerDID, &sigProof.IssuerData)
	if err != nil {
		return nil, err
	}
	if claimRs.MTP.Existence {
		// update revocation status
		err = p.storage.Pgx.BeginFunc(ctx, func(tx pgx.Tx) error {
			claim.Revoked = true
			_, err = p.claimsRepository.Save(ctx, p.storage.Pgx, claim)
			if err != nil {
				return errors.Join(err, errors.New("failed to save claim"))
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		// claim revoked
		return claimRs, nil
	}
	// claim not revoked
	return claimRs, nil
}

func (p *Proof) findNonRevokedClaim(ctx context.Context, claims []*domain.Claim) (*domain.Claim, circuits.MTProof, error) {
	for _, claim := range claims {
		rsClaim, err := p.checkRevocationStatus(ctx, claim)
		if err != nil {
			return nil, circuits.MTProof{}, err
		}
		// current claim revoked. To try next claim.
		if rsClaim == nil {
			continue
		}

		revStatus := circuits.MTProof{
			TreeState: domain.RevocationStatusToTreeState(*rsClaim),
			Proof:     &rsClaim.MTP,
		}

		return claim, revStatus, nil
	}
	return nil, circuits.MTProof{}, ErrAllClaimsRevoked
}

func (p *Proof) toCircuitsQuery(ctx context.Context, claim domain.Claim, query ports.Query) (circuits.Query, map[string]any, error) {
	// check if merklized
	coreClaim := claim.CoreClaim.Get()

	merklizePosition, err := coreClaim.GetMerklizedPosition()
	if err != nil {
		return circuits.Query{}, nil, err
	}
	if merklizePosition == core.MerklizedRootPositionNone {
		merklizedQuery, err := p.prepareNonMerklizedQuery(ctx, claim.SchemaURL, query)
		if err != nil {
			return circuits.Query{}, nil, err
		}
		// for selective disclosure we need to add slot value to the query value to be able to prove it
		if merklizedQuery.Values == nil {
			merklizedQuery.Values = []*big.Int{coreClaim.RawSlotsAsInts()[merklizedQuery.SlotIndex]}
		}
		return merklizedQuery, nil, err
	}

	return p.prepareMerklizedQuery(ctx, claim, query)
}

func (p *Proof) prepareMerklizedQuery(ctx context.Context, claim domain.Claim, query ports.Query) (circuits.Query, map[string]any, error) {
	var verifiablePresentation map[string]any
	vc, err := claim.GetVerifiableCredential()
	if err != nil {
		return circuits.Query{}, verifiablePresentation, err
	}

	// TODO: add merklize options to query
	//nolint:all
	//mk, err := vc.Merklize(ctx, p.schemaLoader.MerklizeOptions()...)
	mk, err := vc.Merklize(ctx)
	if err != nil {
		return circuits.Query{}, verifiablePresentation, err
	}

	remoteDocument, err := p.schemaLoader.LoadDocument(query.Context)
	if err != nil {
		return circuits.Query{}, verifiablePresentation, err
	}

	schema, err := json.Marshal(remoteDocument.Document)
	if err != nil {
		log.Error(ctx, "Failed to marshal schema", "err", err)
		return circuits.Query{}, verifiablePresentation, err
	}

	circuitQuery, field, err := convertToCircuitQuery(
		query.CredentialSubject,
		query.Type,
		schema,
		true,
	)
	if err != nil {
		return circuits.Query{}, verifiablePresentation, err
	}

	fieldPath := merklize.Path{}

	if field != "" {
		fieldPath, err = merklize.NewFieldPathFromContext(schema, query.Type, field)
		if err != nil {
			return circuits.Query{}, verifiablePresentation, err
		}
	}

	err = fieldPath.Prepend("https://www.w3.org/2018/credentials#credentialSubject")
	if err != nil {
		return circuits.Query{}, verifiablePresentation, err
	}

	jsonP, v, err := mk.Proof(ctx, fieldPath)
	if err != nil {
		return circuits.Query{}, verifiablePresentation, err
	}

	value, err := v.MtEntry()
	if err != nil {
		return circuits.Query{}, verifiablePresentation, err
	}

	path, err := fieldPath.MtEntry()
	if err != nil {
		return circuits.Query{}, verifiablePresentation, err
	}

	circuitQuery.ValueProof = &circuits.ValueProof{
		Path:  path,
		Value: value,
		MTP:   jsonP,
	}

	// selective disclosure. For selective disclosure we need to add value to query
	if circuitQuery.Values == nil {
		circuitQuery.Values = []*big.Int{value}
		val, err := mk.RawValue(fieldPath)
		if err != nil {
			return circuits.Query{}, verifiablePresentation, err
		}
		verifiablePresentation = fmtVerifiablePresentation(query.Context,
			query.Type, field, val)
	}

	return circuitQuery, verifiablePresentation, nil
}

func (p *Proof) prepareNonMerklizedQuery(ctx context.Context, jsonSchemaURL string, query ports.Query) (circuits.Query, error) {
	parser := jsonSuite.Parser{}
	pr := processor.InitProcessorOptions(&processor.Processor{},
		processor.WithParser(parser),
		processor.WithDocumentLoader(p.schemaLoader))

	schema, err := jsonschema.Load(ctx, jsonSchemaURL, p.schemaLoader)
	if err != nil {
		return circuits.Query{}, err
	}

	if len(query.Req) > 1 {
		return circuits.Query{}, errors.New("multiple requests are currently not supported")
	}

	circuitQuery, field, err := parseQueryWithoutSlot(query.Req)
	if err != nil {
		return circuits.Query{}, err
	}

	circuitQuery.SlotIndex, err = pr.GetFieldSlotIndex(field, query.Type, schema.BytesNoErr())
	if err != nil {
		return circuits.Query{}, err
	}

	return circuitQuery, nil
}

func (p *Proof) callNonRevProof(ctx context.Context, issuerData verifiable.IssuerData, issuerDID, userDID *w3c.DID) (circuits.MTProof, error) {
	nonRevProof, err := p.revocationSrv.Status(ctx, issuerData.CredentialStatus, userDID, issuerDID, &issuerData)

	if err != nil && errors.Is(err, ErrStateNotFound) {
		state, errIn := merkletree.NewHashFromHex(*issuerData.State.Value)
		if errIn != nil {
			return circuits.MTProof{}, err
		}
		if common.CheckGenesisStateDID(issuerDID, state.BigInt()) != nil {
			return circuits.MTProof{}, errors.New("issuer identity is not genesis and not published")
		}
		return circuits.MTProof{
			Proof: &merkletree.Proof{
				Existence: false,
				NodeAux:   nil,
			},
			TreeState: domain.RevocationStatusToTreeState(verifiable.RevocationStatus{
				Issuer: struct {
					State              *string `json:"state,omitempty"`
					RootOfRoots        *string `json:"rootOfRoots,omitempty"`
					ClaimsTreeRoot     *string `json:"claimsTreeRoot,omitempty"`
					RevocationTreeRoot *string `json:"revocationTreeRoot,omitempty"`
				}{
					State:              issuerData.State.Value,
					RootOfRoots:        issuerData.State.RootOfRoots,
					ClaimsTreeRoot:     issuerData.State.ClaimsTreeRoot,
					RevocationTreeRoot: issuerData.State.RevocationTreeRoot,
				},
				MTP: merkletree.Proof{},
			}),
		}, nil
	}

	return circuits.MTProof{
		Proof:     &nonRevProof.MTP,
		TreeState: domain.RevocationStatusToTreeState(*nonRevProof),
	}, nil
}

func (p *Proof) prepareAuthV2Circuit(ctx context.Context, identifier *w3c.DID, challenge *big.Int) (circuits.AuthV2Inputs, error) {
	authClaim, err := p.claimSrv.GetAuthClaim(ctx, identifier)
	if err != nil {
		return circuits.AuthV2Inputs{}, err
	}

	authClaimData, err := p.fillAuthClaimData(ctx, identifier, authClaim)
	if err != nil {
		return circuits.AuthV2Inputs{}, err
	}
	signature, err := p.signChallange(ctx, authClaim, challenge)
	if err != nil {
		return circuits.AuthV2Inputs{}, err
	}
	globalTree, err := populateGlobalTree(ctx, *identifier, p.stateContract)
	if err != nil {
		return circuits.AuthV2Inputs{}, err
	}
	id, err := core.IDFromDID(*identifier)
	if err != nil {
		return circuits.AuthV2Inputs{}, err
	}
	circuitInputs := prepareAuthV2CircuitInputs(id, authClaimData, challenge, signature, globalTree)
	return circuitInputs, nil
}

func (p *Proof) signChallange(ctx context.Context, authClaim *domain.Claim, challenge *big.Int) (*babyjub.Signature, error) {
	signingKeyID, err := p.identitySrv.GetKeyIDFromAuthClaim(ctx, authClaim)
	if err != nil {
		return nil, err
	}

	challengeDigest := kms.BJJDigest(challenge)

	var sigBytes []byte
	sigBytes, err = p.keyProvider.Sign(ctx, signingKeyID, challengeDigest)
	if err != nil {
		return nil, err
	}

	return kms.DecodeBJJSignature(sigBytes)
}

func (p *Proof) fillAuthClaimData(ctx context.Context, identifier *w3c.DID, authClaim *domain.Claim) (circuits.ClaimWithMTPProof, error) {
	var authClaimData circuits.ClaimWithMTPProof

	err := p.storage.Pgx.BeginFunc(
		ctx, func(tx pgx.Tx) error {
			var errIn error
			var idState *domain.IdentityState
			idState, errIn = p.identitySrv.GetLatestStateByID(ctx, *identifier)
			if errIn != nil {
				return errIn
			}

			identityTrees, errIn := p.mtService.GetIdentityMerkleTrees(ctx, tx, identifier)
			if errIn != nil {
				return errIn
			}

			claimsTree, errIn := identityTrees.ClaimsTree()
			if errIn != nil {
				return errIn
			}
			// get index hash of authClaim
			coreClaim := authClaim.CoreClaim.Get()
			hIndex, errIn := coreClaim.HIndex()
			if errIn != nil {
				return errIn
			}

			authClaimMTP, _, errIn := claimsTree.GenerateProof(ctx, hIndex, idState.TreeState().ClaimsRoot)
			if errIn != nil {
				return errIn
			}

			authClaimData = circuits.ClaimWithMTPProof{
				Claim: coreClaim,
			}

			authClaimData.IncProof = circuits.MTProof{
				Proof:     authClaimMTP,
				TreeState: idState.TreeState(),
			}

			// revocation / non revocation MTP for the latest identity state
			nonRevocationProof, errIn := identityTrees.
				GenerateRevocationProof(ctx, new(big.Int).SetUint64(uint64(authClaim.RevNonce)), idState.TreeState().RevocationRoot)

			authClaimData.NonRevProof = circuits.MTProof{
				TreeState: idState.TreeState(),
				Proof:     nonRevocationProof,
			}

			return errIn
		})
	if err != nil {
		return authClaimData, err
	}
	return authClaimData, nil
}

func prepareAuthV2CircuitInputs(id core.ID, authClaim circuits.ClaimWithMTPProof, challenge *big.Int, signature *babyjub.Signature, globalMTP circuits.GISTProof) circuits.AuthV2Inputs {
	return circuits.AuthV2Inputs{
		GenesisID:          &id,
		ProfileNonce:       big.NewInt(0),
		AuthClaim:          authClaim.Claim,
		AuthClaimIncMtp:    authClaim.IncProof.Proof,
		AuthClaimNonRevMtp: authClaim.NonRevProof.Proof,
		TreeState:          authClaim.IncProof.TreeState,
		Signature:          signature,
		Challenge:          challenge,
		GISTProof:          globalMTP,
	}
}

func populateGlobalTree(ctx context.Context, did w3c.DID, contract *abi.State) (circuits.GISTProof, error) {
	// get global root
	id, err := core.IDFromDID(did)
	if err != nil {
		return circuits.GISTProof{}, err
	}
	globalProof, err := contract.GetGISTProof(&bind.CallOpts{Context: ctx}, id.BigInt())
	if err != nil {
		return circuits.GISTProof{}, err
	}

	return toMerkleTreeProof(globalProof)
}

func toMerkleTreeProof(smtProof abi.IStateGistProof) (circuits.GISTProof, error) {
	var existence bool
	var nodeAux *merkletree.NodeAux
	var err error

	if smtProof.Existence {
		existence = true
	} else {
		existence = false
		if smtProof.AuxExistence {
			nodeAux = &merkletree.NodeAux{}
			nodeAux.Key, err = merkletree.NewHashFromBigInt(smtProof.AuxIndex)
			if err != nil {
				return circuits.GISTProof{}, err
			}
			nodeAux.Value, err = merkletree.NewHashFromBigInt(smtProof.AuxValue)
			if err != nil {
				return circuits.GISTProof{}, err
			}
		}
	}

	allSiblings := make([]*merkletree.Hash, len(smtProof.Siblings))
	for i, s := range smtProof.Siblings {
		sh, err2 := merkletree.NewHashFromBigInt(s)
		if err2 != nil {
			return circuits.GISTProof{}, err
		}
		allSiblings[i] = sh
	}

	proof, err := merkletree.NewProofFromData(existence, allSiblings, nodeAux)
	if err != nil {
		return circuits.GISTProof{}, err
	}

	root, err := merkletree.NewHashFromBigInt(smtProof.Root)
	if err != nil {
		return circuits.GISTProof{}, err
	}

	return circuits.GISTProof{
		Root:  root,
		Proof: proof,
	}, nil
}

func getValuesFromArray(v interface{}) ([]*big.Int, error) {
	values := []*big.Int{}

	switch value := v.(type) {
	case float64:
		values = []*big.Int{new(big.Int).SetInt64(int64(value))}
	case []interface{}:
		for _, item := range value {
			if itemFloat, ok := item.(float64); ok {
				values = append(values, new(big.Int).SetInt64(int64(itemFloat)))
			} else {
				return nil, fmt.Errorf("unsupported values type in value element %T, expected float64", item)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported values type %T", v)
	}

	return values, nil
}

func parseQueryWithoutSlot(req map[string]interface{}) (circuits.Query, string, error) {
	for field, body := range req {
		condition, ok := body.(map[string]interface{})

		if !ok {
			return circuits.Query{}, "", errors.New("failed cast type map[string]interface")
		}

		if len(condition) > 1 {
			return circuits.Query{}, "", errors.New("multiple predicates are currently not supported")
		}

		for op, v := range condition {

			intOp, ok := circuits.QueryOperators[op]
			if !ok {
				return circuits.Query{}, "", errors.New("query operator is not supported")
			}

			values, err := getValuesFromArray(v)
			if err != nil {
				return circuits.Query{}, "", err
			}

			return circuits.Query{
				Operator: intOp,
				Values:   values,
			}, field, nil
		}
	}
	return circuits.Query{
		Operator:  0,
		Values:    []*big.Int{},
		SlotIndex: 0,
	}, "", nil
}

func convertToCircuitQuery(req map[string]interface{}, credentialType string, schema []byte, merklized bool) (circuits.Query, string, error) {
	fieldName, op, values, err := parseCredentialSubject(req)
	if err != nil {
		return circuits.Query{}, "", err
	}

	var datatype string
	if fieldName != "" {
		datatype, err = merklize.TypeFromContext(
			schema,
			fmt.Sprintf("%s.%s", credentialType, fieldName),
		)
		if err != nil {
			return circuits.Query{}, "", err
		}
	}

	q, err := createQuery(op, values, datatype, merklized)
	if err != nil {
		return circuits.Query{}, "", err
	}

	return q, fieldName, nil
}

func parseCredentialSubject(req map[string]interface{}) (
	field string,
	intOp int,
	values any,
	err error,
) {
	for field, body := range req {
		condition, ok := body.(map[string]interface{})
		if !ok {
			err := errors.New("failed cast type map[string]interface")
			return "", 0, nil, err
		}
		if len(condition) > 1 {
			err := errors.New("multiple predicates are currently not supported")
			return "", 0, nil, err
		}
		if len(condition) == 0 {
			return field, circuits.EQ, nil, nil
		}

		for op, v := range condition {
			intOp, ok := circuits.QueryOperators[op]
			if !ok {
				err := errors.New("query operator is not supported")
				return "", 0, nil, err
			}

			return field, intOp, v, nil
		}
	}
	return "", 0, nil, nil
}

func createQuery(op int, values any, datatype string, merklized bool) (circuits.Query, error) {
	if op == circuits.NOOP {
		if merklized {
			op = circuits.EQ
		}
		return circuits.Query{
			Operator: op,
			Values:   nil,
		}, nil
	}

	v, err := getValuesAsArray(values, datatype)
	if err != nil {
		return circuits.Query{}, err
	}

	return circuits.Query{
		Operator: op,
		Values:   v,
	}, nil
}

func getValuesAsArray(v interface{}, valueType string) ([]*big.Int, error) {
	if v == nil {
		return nil, nil
	}
	var values []*big.Int
	listOfValues, ok := v.([]interface{})
	if ok {
		values = make([]*big.Int, len(listOfValues))
		for i, item := range listOfValues {
			hashedValue, err := merklize.HashValue(valueType, item)
			if err != nil {
				return nil, err
			}
			values[i] = hashedValue
		}
		return values, nil
	}
	hashedValue, err := merklize.HashValue(valueType, v)
	if err != nil {
		return nil, err
	}
	values = append(values, hashedValue)

	return values, nil
}

func fmtVerifiablePresentation(tpContext string, tp string, field string, value any) map[string]any {
	var ldContext any
	baseContext := []any{"https://www.w3.org/2018/credentials/v1"}
	if tpContext == baseContext[0] {
		ldContext = baseContext
	} else {
		contexts := baseContext
		contexts = append(contexts, tpContext)
		ldContext = contexts
	}

	vcTypes := []any{"VerifiableCredential"}
	if tp != "VerifiableCredential" {
		vcTypes = append(vcTypes, tp)
	}

	return map[string]any{
		"@context": baseContext,
		"@type":    "VerifiablePresentation",
		"verifiableCredential": map[string]any{
			"@context": ldContext,
			"@type":    vcTypes,
			"credentialSubject": map[string]any{
				"@type": tp,
				field:   value,
			},
		},
	}
}
