package services

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
)

func TestSaveCredential(t *testing.T) {
	ctx := t.Context()
	identity, err := identityService.Create(ctx, "http://localhost", &ports.DIDCreationOptions{
		Blockchain: blockchain,
		Network:    net,
		Method:     method,
	})
	assert.NoError(t, err)
	assert.NotNil(t, identity)

	identifier := identity.Identifier
	did, err := w3c.ParseDID(identifier)
	assert.NoError(t, err)

	t.Run("save encrypted credential", func(t *testing.T) {
		claimId := uuid.New()
		encryptedKey := map[string]interface{}{
			"kty": "EC",
			"crv": "P-256",
			"alg": "ECDH-ES+A256KW",
			"use": "enc",
			"x":   "nw7Ag_FszrDu1uPi2lX3TtbF7FMZoysXZXUzrKxBwiQ",
			"y":   "l1I0EONJmEHMz7Nc4WQULDllKdPdjbTgHS5hCbqv0UQ",
			"kid": "tu-kid",
		}
		req := &ports.CreateClaimRequest{
			ClaimID: &claimId,
			DID:     did,
			Schema:  "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
			Type:    "KYCAgeCredential",
			CredentialSubject: map[string]any{
				"id":           "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
				"birthday":     19960425,
				"documentType": 2,
			},
			Expiration:     common.ToPointer(time.Now().Add(365 * 24 * time.Hour)),
			SignatureProof: true,
			Version:        0,
			RevNonce:       common.ToPointer[uint64](100),
			EncryptionKey:  encryptedKey,
		}

		credential, err := claimsService.Save(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, credential)

		credential, err = claimsService.GetByID(ctx, did, claimId)
		assert.NoError(t, err)
		assert.NotNil(t, credential)
		assert.NotNil(t, credential.EncryptedData)
		assert.Equal(t, pgtype.Null, credential.Data.Status)
		assert.NotEqual(t, pgtype.Null, credential.SignatureProof.Status)
	})

	t.Run("save non encrypted credential", func(t *testing.T) {
		claimId := uuid.New()
		req := &ports.CreateClaimRequest{
			ClaimID: &claimId,
			DID:     did,
			Schema:  "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
			Type:    "KYCAgeCredential",
			CredentialSubject: map[string]any{
				"id":           "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
				"birthday":     19960425,
				"documentType": 2,
			},
			Expiration:     common.ToPointer(time.Now().Add(365 * 24 * time.Hour)),
			SignatureProof: true,
			Version:        0,
			RevNonce:       common.ToPointer[uint64](100),
			EncryptionKey:  nil,
		}

		credential, err := claimsService.Save(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, credential)

		credential, err = claimsService.GetByID(ctx, did, claimId)
		assert.NoError(t, err)
		assert.NotNil(t, credential)
		assert.Nil(t, credential.EncryptedData)
		assert.NotEqual(t, pgtype.Null, credential.Data.Status)
		assert.NotEqual(t, pgtype.Null, credential.SignatureProof.Status)
	})
}

func TestAgent(t *testing.T) {
	ctx := t.Context()
	identity, err := identityService.Create(ctx, "http://localhost", &ports.DIDCreationOptions{
		Blockchain: blockchain,
		Network:    net,
		Method:     method,
	})
	assert.NoError(t, err)
	assert.NotNil(t, identity)

	identifier := identity.Identifier
	did, err := w3c.ParseDID(identifier)
	assert.NoError(t, err)

	t.Run("fetch encrypted credential", func(t *testing.T) {
		claimId := uuid.New()

		userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi")
		assert.NoError(t, err)

		encryptedKey := map[string]interface{}{
			"kty": "EC",
			"crv": "P-256",
			"alg": "ECDH-ES+A256KW",
			"use": "enc",
			"x":   "nw7Ag_FszrDu1uPi2lX3TtbF7FMZoysXZXUzrKxBwiQ",
			"y":   "l1I0EONJmEHMz7Nc4WQULDllKdPdjbTgHS5hCbqv0UQ",
			"kid": "tu-kid",
		}
		req := &ports.CreateClaimRequest{
			ClaimID: &claimId,
			DID:     did,
			Schema:  "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
			Type:    "KYCAgeCredential",
			CredentialSubject: map[string]any{
				"id":           userDID.String(),
				"birthday":     19960425,
				"documentType": 2,
			},
			Expiration:     common.ToPointer(time.Now().Add(365 * 24 * time.Hour)),
			SignatureProof: true,
			Version:        0,
			RevNonce:       common.ToPointer[uint64](100),
			EncryptionKey:  encryptedKey,
		}

		credential, err := claimsService.Save(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, credential)

		credential, err = claimsService.GetByID(ctx, did, claimId)
		assert.NoError(t, err)
		assert.NotNil(t, credential)
		assert.NotNil(t, credential.EncryptedData)
		assert.Equal(t, pgtype.Null, credential.Data.Status)
		assert.NotEqual(t, pgtype.Null, credential.SignatureProof.Status)

		mediaType := protocol.CredentialFetchRequestMessageType
		fetchCredentialBody := protocol.CredentialFetchRequestMessageBody{
			ID: claimId.String(),
		}

		fetchCredentialBodyBytes, err := json.Marshal(fetchCredentialBody)
		assert.NoError(t, err)
		agentRequest := &ports.AgentRequest{
			Body:      fetchCredentialBodyBytes,
			IssuerDID: did,
			UserDID:   userDID,
			Type:      protocol.CredentialFetchRequestMessageType,
			ThreadID:  uuid.New().String(),
		}
		basicMessage, err := claimsService.Agent(ctx, agentRequest, iden3comm.MediaType(mediaType))
		assert.NoError(t, err)
		assert.NotNil(t, basicMessage)
		assert.Equal(t, userDID.String(), basicMessage.To)
		assert.Equal(t, did.String(), basicMessage.From)
		var body map[string]any
		var jwe protocol.JWEJSONEncryption
		assert.NoError(t, json.Unmarshal(basicMessage.Body, &body))
		jwes, ok := body["data"].(map[string]interface{})
		assert.True(t, ok)
		jweb, err := json.Marshal(jwes)
		assert.NoError(t, err)
		assert.NoError(t, json.Unmarshal(jweb, &jwe))
		assert.NotNil(t, jwe.Iv)
		assert.NotNil(t, jwe.Protected)
		assert.NotNil(t, jwe.Ciphertext)
		assert.NotNil(t, jwe.Tag)
		assert.NotNil(t, jwe.EncryptedKey)
	})

	t.Run("fetch non encrypted credential", func(t *testing.T) {
		claimId := uuid.New()

		userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi")
		assert.NoError(t, err)

		req := &ports.CreateClaimRequest{
			ClaimID: &claimId,
			DID:     did,
			Schema:  "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
			Type:    "KYCAgeCredential",
			CredentialSubject: map[string]any{
				"id":           userDID.String(),
				"birthday":     19960425,
				"documentType": 2,
			},
			Expiration:     common.ToPointer(time.Now().Add(365 * 24 * time.Hour)),
			SignatureProof: true,
			Version:        0,
			RevNonce:       common.ToPointer[uint64](100),
			EncryptionKey:  nil,
		}

		credential, err := claimsService.Save(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, credential)

		credential, err = claimsService.GetByID(ctx, did, claimId)
		assert.NoError(t, err)
		assert.NotNil(t, credential)
		assert.NotEqual(t, pgtype.Null, credential.Data.Status)
		assert.NotEqual(t, pgtype.Null, credential.SignatureProof.Status)

		mediaType := protocol.CredentialFetchRequestMessageType
		fetchCredentialBody := protocol.CredentialFetchRequestMessageBody{
			ID: claimId.String(),
		}

		fetchCredentialBodyBytes, err := json.Marshal(fetchCredentialBody)
		assert.NoError(t, err)
		agentRequest := &ports.AgentRequest{
			Body:      fetchCredentialBodyBytes,
			IssuerDID: did,
			UserDID:   userDID,
			Type:      protocol.CredentialFetchRequestMessageType,
			ThreadID:  uuid.New().String(),
		}
		basicMessage, err := claimsService.Agent(ctx, agentRequest, iden3comm.MediaType(mediaType))
		assert.NoError(t, err)
		assert.NotNil(t, basicMessage)
		assert.Equal(t, userDID.String(), basicMessage.To)
		assert.Equal(t, did.String(), basicMessage.From)
		var body map[string]any
		var vc verifiable.W3CCredential
		assert.NoError(t, json.Unmarshal(basicMessage.Body, &body))
		vcs, ok := body["credential"].(map[string]interface{})
		assert.True(t, ok)
		vcb, err := json.Marshal(vcs)
		assert.NoError(t, err)
		assert.NoError(t, json.Unmarshal(vcb, &vc))
		assert.NotNil(t, vc.Context)
		assert.NotNil(t, vc.ID)
		assert.NotNil(t, vc.Type)
		assert.NotNil(t, vc.Issuer)
		assert.NotNil(t, vc.IssuanceDate)
		assert.NotNil(t, vc.CredentialSubject)
	})
}

func TestConvertDataToJWEJsonEncryption(t *testing.T) {
	data := "eyJjaXBoZXJ0ZXh0IjoiaFFWQkNFRlRDUm1JNE04c3I0NUNQV0NKcFFsdnYtNWVrTjN5bXRQWExTU01fWVB0c1BwbzUxcmVtRG5XT0Utd2RxcDZvaXpqdUdJQ2dTWnVSVnQxMlpTQllOek5XcDhwaERKcm1NeUVlUTNtWWE4MUtBYjJ1c0pyaDFfNzZ2dVRSazBJaHB0TGNFNnlkS0N6cTJfMUllRVhDSVNsb2xiQ1VKUndrMXBGV09RN2tpOG55Q0lpUjFLWWpIQTV6Y0RvRVV1c1NNUk9yaC1vTzBmTTZocUI2Ykw3Z1BfcENiTVQ2MzRoTndiVUV6UHROaFB2TEdzU0xLWEdybjhzTEY1QzlkVkxUS1lYNS01aXY3ZWhPQnFBdW5LV21BUUI1eDVTSVRSQ1VGR2tOZ0NWTFFnbFQwRTAwdDNqXzltZWlLZjBvdWUyc2pHOWxNT0R1aXF0Q09PODZocFQ5Zi1RUTBNcEdTVDBuU3k1cmVlZlZXdUpja3pYamh0S0NHVVpDdFBZRFRzUGJLQlk4VEllTHBVWWVhMzRVSTNFNTBNMUhETHNrN2pEajBmVzVYbzZlOWpvbXRaTlVyMk1UU2lEbno2LXBFUy1takNWUm1OcmQ5QzJZd3JGWFI2eTBibDljMmVNQWd1a1piRUtWbWZwUG9RU0tvdFdiQjRnRVR5emp6UjRqN1NDaVZIZ0NHZHR4akV2bXlqUms0blhBVkxnUjN3SmdfdURTZm9mVkVJNld4c3RReHNaRThoM2wwN1NjaGM3Z1FIcU53YTh4dkhVOFhhR0VtTlVyN1VsUHk2TlNxMGlTQm96a3hhZGVRdDFmZWVjd3o2YlBtMWktNE9Ya0hZWXl4NnJBUUFlNzR6Z1B3WVM5YWM1LXhVVEZscWNIdGpKWnRuckJoWUNvS18wcTNBREhQaWFjZDUtRk1QWHZ4WUpjTnA5aHctYXZxeGFRVXFfVlpNenNlU2VfM1hmWWhLTUdLVk5IRHNKMk5iTTc2SjNteVFhWlBod1pxZkVMY200N2NaalZWMzktVWJpdVo0bU5fSnNsNDZWdklSbDRhZGtEUkVvTG16OC1FSExZYTFIZ1ltc2dkdnVWTzZ2OWV0Ukdpb185UjFXa1lfNmlFbmRyNEMyNGpYT0JacmxLUWpxcUlIbEZqV1BPYmZrTm1yMEkxM1hxY3F4T2w0TVRrWW40RTNmMEFaY25WdC1oZlNWUmtVZm9oZlVOM1Rfak9pTEdfUTVqSjhPZ0ZleWpjWVpKZk8wRmVVQWFVNmJJOTA3M3NfdFZHTXFxUk4yQ3ZHTEVKeERQVkUteTd4eVNhUXVkUWJZX3h1cTZtVXM4c3EwWE5YVGZSOEJuUS01QVl3SWJDMjVkelpmQUtkSXFwUDAzTXZKbjUxbXVjVDhHUEtWTWxQaWNib0VKbDlIM2VQeUU2TUx5a283b3RTT3o5SHEwZ2UxNlBlUVVKSWlWeHN6SWpCQlZZUXduQ2U0dTRoSzhlUnMyajNNUC1PN0E4eE9VY1JDUnhrRzY0WkRuSHlqRjA5bE9jTFU3dUtjYlc4eTZtcC1ZLTNEUVZIdWpBajA0c3F2RzU2T1NseF9wX0JJMHF3M2t5eW1kNnZMS2pKNUxHNXB5TnNncGtIVFU2ZHZrVTRtTmh0STZCb3lURjBNYUU2ZG5NRXNXOTQiLCJlbmNyeXB0ZWRfa2V5IjoiTDV4ZXpNMXQweGcxVW1GcW9DNUhEbmZQbHJPbFVvNHhqTm1fSGZ2LXdfakIzaUJ3ZzhHam9BIiwiaGVhZGVyIjp7ImFsZyI6IkVDREgtRVMrQTI1NktXIiwiZXBrIjp7ImNydiI6IlgyNTUxOSIsImt0eSI6Ik9LUCIsIngiOiJBV1hScWdTSmp6MkxuM21kNGtPdGRUOG8weGd0QlJSSnA0YzVwbUJwY0VFIn0sImtpZCI6InR1LWtpZCJ9LCJpdiI6IklOWmphR1ZGVzM1OVVWbzMiLCJwcm90ZWN0ZWQiOiJleUpoYkdjaU9pSkZRMFJJTFVWVEswRXlOVFpMVnlJc0ltVnVZeUk2SWtFeU5UWkhRMDBpTENKbGNHc2lPbnNpWTNKMklqb2lXREkxTlRFNUlpd2lhM1I1SWpvaVQwdFFJaXdpZUNJNklrRlhXRkp4WjFOS2Fub3lURzR6YldRMGEwOTBaRlE0YnpCNFozUkNVbEpLY0RSak5YQnRRbkJqUlVVaWZTd2lhMmxrSWpvaWRIVXRhMmxrSWl3aWRIbHdJam9pWVhCd2JHbGpZWFJwYjI0dmFXUmxiak5qYjIxdExXVnVZM0o1Y0hSbFpDMXFjMjl1SW4wIiwidGFnIjoiRVV5eUNaYTZIdzA0NmhtSVJBLWJBQSJ9"
	JWE, err := convertDataToJWEJsonEncryption(&data)
	assert.NoError(t, err)
	assert.NotNil(t, JWE)

	epk, ok := JWE.Header["epk"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "ECDH-ES+A256KW", JWE.Header["alg"])
	assert.Equal(t, "X25519", epk["crv"])
	assert.Equal(t, "OKP", epk["kty"])
	assert.Equal(t, "AWXRqgSJjz2Ln3md4kOtdT8o0xgtBRRJp4c5pmBpcEE", epk["x"])
	assert.Equal(t, "L5xezM1t0xg1UmFqoC5HDnfPlrOlUo4xjNm_Hfv-w_jB3iBwg8GjoA", JWE.EncryptedKey)
	assert.Equal(t, "INZjaGVFW359UVo3", JWE.Iv)
	assert.Equal(t, "eyJhbGciOiJFQ0RILUVTK0EyNTZLVyIsImVuYyI6IkEyNTZHQ00iLCJlcGsiOnsiY3J2IjoiWDI1NTE5Iiwia3R5IjoiT0tQIiwieCI6IkFXWFJxZ1NKanoyTG4zbWQ0a090ZFQ4bzB4Z3RCUlJKcDRjNXBtQnBjRUUifSwia2lkIjoidHUta2lkIiwidHlwIjoiYXBwbGljYXRpb24vaWRlbjNjb21tLWVuY3J5cHRlZC1qc29uIn0", JWE.Protected)
	assert.Equal(t, "EUyyCZa6Hw046hmIRA-bAA", JWE.Tag)
}

func TestBuildEncryptedCredentialBody(t *testing.T) {
	ctx := t.Context()
	identity, err := identityService.Create(ctx, "http://localhost", &ports.DIDCreationOptions{
		Blockchain: blockchain,
		Network:    net,
		Method:     method,
	})
	assert.NoError(t, err)
	assert.NotNil(t, identity)

	identifier := identity.Identifier
	did, err := w3c.ParseDID(identifier)
	assert.NoError(t, err)
	claimId := uuid.New()
	encryptedKey := map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"alg": "ECDH-ES+A256KW",
		"use": "enc",
		"x":   "nw7Ag_FszrDu1uPi2lX3TtbF7FMZoysXZXUzrKxBwiQ",
		"y":   "l1I0EONJmEHMz7Nc4WQULDllKdPdjbTgHS5hCbqv0UQ",
		"kid": "tu-kid",
	}
	req := &ports.CreateClaimRequest{
		ClaimID: &claimId,
		DID:     did,
		Schema:  "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
		Type:    "KYCAgeCredential",
		CredentialSubject: map[string]any{
			"id":           "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
			"birthday":     19960425,
			"documentType": 2,
		},
		Expiration:     common.ToPointer(time.Now().Add(365 * 24 * time.Hour)),
		SignatureProof: true,
		Version:        0,
		RevNonce:       common.ToPointer[uint64](100),
		EncryptionKey:  encryptedKey,
	}

	credential, err := claimsService.Save(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, credential)

	credential, err = claimsService.GetByID(ctx, did, claimId)
	assert.NoError(t, err)
	assert.NotNil(t, credential)

	body, err := buildEncryptedCredentialBody(ctx, credential)
	assert.NoError(t, err)
	assert.NotNil(t, body)

	var encryptedIssuanceMessageBody protocol.EncryptedIssuanceMessageBody
	assert.NoError(t, json.Unmarshal(body, &encryptedIssuanceMessageBody))
	assert.NotNil(t, encryptedIssuanceMessageBody.ID)
	assert.NotNil(t, encryptedIssuanceMessageBody.Data)
	assert.NotNil(t, encryptedIssuanceMessageBody.Type)
	epk, ok := encryptedIssuanceMessageBody.Data.Header["epk"].(map[string]interface{})
	assert.True(t, ok)
	assert.NotNil(t, epk)
	assert.NotEmpty(t, encryptedIssuanceMessageBody.Data.EncryptedKey)
	assert.NotEmpty(t, encryptedIssuanceMessageBody.Data.Ciphertext)
}
