package sh_id_platform

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChorra(t *testing.T) {
	from := map[string]interface{}{
		"@context": []interface{}{"https://www.w3.org/2018/credentials/v1", "https://schema.iden3.io/core/jsonld/iden3proofs.jsonld", "schemaContext"},
		"credentialSchema": map[string]interface{}{
			"id":   "schemaDB.Type",
			"type": "JsonSchemaValidator2018",
		},
		"credentialStatus": map[string]interface{}{
			"id": "testStatus",
		},
		"credentialSubject": "credentialSubject",
		"id":                "testID",
		"issuanceDate":      "2023-04-27T23:45:29.498555+02:00",
		"issuer":            "did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5",
		"type": []interface{}{
			"VerifiableCredential", "schemaDB.Type",
		},
	}
	var to map[string]interface{}
	raw, err := json.Marshal(from)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &to))
	assert.Equal(t, from, to)
}
