package jsonschema

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/loader"
)

func Test_Attributes(t *testing.T) {
	ctx := context.Background()
	const url = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	jsonSchema, err := Load(ctx, loader.HTTPFactory(url))
	assert.NoError(t, err)
	attributes, err := jsonSchema.Attributes()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(attributes))
}

func Test_ValidateCredentialSubject(t *testing.T) {
	ctx := context.Background()
	const url = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	jsonSchema, err := Load(ctx, loader.HTTPFactory(url))
	require.NoError(t, err)

	type testConfig struct {
		name  string
		input domain.CredentialSubject
		error bool
	}

	for _, tc := range []testConfig{
		{
			name:  "should get an error, empty credentialSubject",
			input: domain.CredentialSubject{},
			error: true,
		},
		{
			name:  "should get an error, birthday and documentType required, 1 provided",
			input: domain.CredentialSubject{"birthday": 19960424},
			error: true,
		},
		{
			name:  "should get an error, birthday and documentType required, 1 provided",
			input: domain.CredentialSubject{"documentType": 2},
			error: true,
		},
		{
			name:  "should get an error, both required provided but invalid type of documentType",
			input: domain.CredentialSubject{"documentType": "2", "birthday": 19960424},
			error: true,
		},
		{
			name:  "happy path",
			input: domain.CredentialSubject{"documentType": 2, "birthday": 19960424},
			error: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := jsonSchema.ValidateCredentialSubject(tc.input)
			if tc.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
