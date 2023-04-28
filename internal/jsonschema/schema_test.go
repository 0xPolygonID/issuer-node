package jsonschema

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/pkg/cache"
)

func TestValidateCredentialSubject(t *testing.T) {
	schemaLoader := loader.CachedFactory(loader.HTTPFactory, cache.NewMemoryCache())
	ctx := context.Background()
	schemaURL := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCEmployee-v101.json"
	schemaType := "KYCEmployee"

	type config struct {
		name              string
		credentialSubject map[string]interface{}
		expectedError     bool
	}

	for _, tc := range []config{
		{
			name: "invalid date format",
			credentialSubject: map[string]interface{}{
				"ZKPexperiance": true,
				"documentType":  4,
				"hireDate":      "asdads2022-10-10",
				"position":      "p",
				"salary":        123,
			},
			expectedError: true,
		},
		{
			name: "invalid bool format",
			credentialSubject: map[string]interface{}{
				"ZKPexperiance": "true",
				"documentType":  4,
				"hireDate":      "asdads2022-10-10",
				"position":      "p",
				"salary":        123,
			},
			expectedError: true,
		},
		{
			name: "invalid integer format",
			credentialSubject: map[string]interface{}{
				"ZKPexperiance": true,
				"documentType":  "4",
				"hireDate":      "2022-10-10",
				"position":      "p",
				"salary":        123,
			},
			expectedError: false,
		},
		{
			name: "invalid number format",
			credentialSubject: map[string]interface{}{
				"ZKPexperiance": true,
				"documentType":  4,
				"hireDate":      "2022-10-10",
				"position":      "p",
				"salary":        "123",
			},
			expectedError: false,
		},
		{
			name: "invalid string format",
			credentialSubject: map[string]interface{}{
				"ZKPexperiance": true,
				"documentType":  4,
				"hireDate":      "2022-10-10",
				"position":      123,
				"salary":        123,
			},
			expectedError: false,
		},
		{
			name: "happy path, ok",
			credentialSubject: map[string]interface{}{
				"ZKPexperiance": true,
				"documentType":  4,
				"hireDate":      "2022-10-10",
				"position":      "p",
				"salary":        123,
			},
			expectedError: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateCredentialSubject(ctx, schemaLoader(schemaURL), schemaType, tc.credentialSubject)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
