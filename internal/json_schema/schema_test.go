package json_schema

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/loader"
)

func TestValidateCredentialSubject(t *testing.T) {
	const ipfsGatewayURL = "http://127.0.0.1:8080"
	ctx := context.Background()
	ld := loader.NewDocumentLoader(ipfsGatewayURL, false)

	type config struct {
		name              string
		credentialSubject map[string]interface{}
		expectedError     bool
		schemaURL         string
		schemaType        string
	}

	for _, tc := range []config{
		{
			name:       "invalid date format",
			schemaURL:  "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCEmployee-v101.json",
			schemaType: "KYCEmployee",
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
			name:       "invalid bool format",
			schemaURL:  "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCEmployee-v101.json",
			schemaType: "KYCEmployee",
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
			name:       "invalid integer format",
			schemaURL:  "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCEmployee-v101.json",
			schemaType: "KYCEmployee",
			credentialSubject: map[string]interface{}{
				"ZKPexperiance": true,
				"documentType":  "4",
				"hireDate":      "2022-10-10",
				"position":      "p",
				"salary":        123,
			},
			expectedError: true,
		},
		{
			name:       "invalid number format",
			schemaURL:  "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCEmployee-v101.json",
			schemaType: "KYCEmployee",
			credentialSubject: map[string]interface{}{
				"ZKPexperiance": true,
				"documentType":  4,
				"hireDate":      "2022-10-10",
				"position":      "p",
				"salary":        "123",
			},
			expectedError: true,
		},
		{
			name:       "invalid string format",
			schemaURL:  "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCEmployee-v101.json",
			schemaType: "KYCEmployee",
			credentialSubject: map[string]interface{}{
				"ZKPexperiance": true,
				"documentType":  4,
				"hireDate":      "2022-10-10",
				"position":      123,
				"salary":        123,
			},
			expectedError: true,
		},
		{
			name:       "happy path, ok",
			schemaURL:  "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCEmployee-v101.json",
			schemaType: "KYCEmployee",
			credentialSubject: map[string]interface{}{
				"ZKPexperiance": true,
				"documentType":  4,
				"hireDate":      "2022-10-10",
				"position":      "p",
				"salary":        123,
			},
			expectedError: false,
		},
		{
			name:       "should fail, invalid dateType",
			schemaURL:  "http://localhost:8080/json/exampleMultidepth.json",
			schemaType: "Employee",
			credentialSubject: map[string]interface{}{
				"ZKPexperiance": true,
				"hireDate":      "2023-05-11T09:49:16.335Z",
				"position":      "p",
				"salary":        123,
				"vegan":         true,
				"passportInfo": map[string]interface{}{
					"birthyear":        1950,
					"numberOfBrothers": 2,
					"name":             "John",
					"parents": map[string]interface{}{
						"fatherBirthyear": 1910,
						"motherBirthyear": 1914,
					},
				},
			},
			expectedError: true,
		},
		{
			name:       "invalid position should be one of the enum",
			schemaURL:  "http://localhost:8080/json/exampleMultidepth.json",
			schemaType: "Employee",
			credentialSubject: map[string]interface{}{
				"ZKPexperiance": true,
				"hireDate":      "2023-05-11",
				"position":      "p",
				"salary":        123,
				"vegan":         true,
				"passportInfo": map[string]interface{}{
					"birthyear":        1950,
					"numberOfBrothers": 2,
					"name":             "John",
					"parents": map[string]interface{}{
						"fatherBirthyear": 1910,
						"motherBirthyear": 1914,
					},
				},
			},
			expectedError: true,
		},
		{
			name:       "happy path multi depth with valid date",
			schemaURL:  "http://localhost:8080/json/exampleMultidepth.json",
			schemaType: "Employee",
			credentialSubject: map[string]interface{}{
				"ZKPexperiance": true,
				"hireDate":      "2023-05-03",
				"position":      "Account Executive",
				"salary":        123,
				"vegan":         true,
				"passportInfo": map[string]interface{}{
					"birthyear":        1950,
					"numberOfBrothers": 2,
					"name":             "John",
					"parents": map[string]interface{}{
						"fatherBirthyear": 1910,
						"motherBirthyear": 1914,
					},
				},
			},
			expectedError: false,
		},
		{
			name:       "ipfs file",
			schemaURL:  "ipfs://QmcoHYA1K1xiFPHTUbDtVkV6HoKRCentZetfv4yLbrjwfY",
			schemaType: "TestDefaultValues",
			credentialSubject: map[string]interface{}{
				"id":      "did:polygonid:polygon:mumbai:2qLPX9XnujT2xhuiPMHrqXTUD96UCV87CtThRUZFQm",
				"integer": 100,
				"boolean": false,
				"string":  "foo",
				"number":  99.99,
				"object": map[string]interface{}{
					"string":    "bar",
					"date":      "2023-06-29",
					"time":      "16:55:34.769+02:00",
					"boolean":   true,
					"number":    100.99,
					"date-time": "2023-06-29T16:54:23.920+02:00",
					"integer":   101,
				},
			},
			expectedError: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateCredentialSubject(ctx, ld, tc.schemaURL, tc.schemaType, tc.credentialSubject)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
