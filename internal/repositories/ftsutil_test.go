package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullTextSearchQuery(t *testing.T) {
	type testConfig struct {
		input    string
		operator string
		expect   string
	}
	for _, tc := range []testConfig{
		{
			input:    "",
			operator: "&",
			expect:   "",
		},
		{
			input:    "word",
			operator: "&",
			expect:   "(word:* | word)",
		},
		{
			input:    "two words",
			operator: "&",
			expect:   "(two:* | two)&(words:* | words)",
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expect, fullTextSearchQuery(tc.input, tc.operator))
		})
	}
}

func TestGetDIDFromQuery(t *testing.T) {
	type testConfig struct {
		input  string
		expect string
	}
	for _, tc := range []testConfig{
		{
			input:  "",
			expect: "",
		},
		{
			input:  "word",
			expect: "",
		},
		{
			input:  "did",
			expect: "",
		},
		{
			input:  "did:",
			expect: "did:",
		},
		{
			input:  "did:polygonid:polygon:mumbai:2qFpPHotk6oyaX1fcrpQFT4BMnmg8YszUwxYtaoGoe",
			expect: "did:polygonid:polygon:mumbai:2qFpPHotk6oyaX1fcrpQFT4BMnmg8YszUwxYtaoGoe",
		},
		{
			input:  "peter tango did:polygonid:polygon:mumbai:2qFpPHotk6oyaX1fcrpQFT4BMnmg8YszUwxYtaoGoe cash",
			expect: "did:polygonid:polygon:mumbai:2qFpPHotk6oyaX1fcrpQFT4BMnmg8YszUwxYtaoGoe",
		},
		{
			input:  "two dids only first did:polygonid:polygon:mumbai:2qFpPHotk6oyaX1fcrpQFT4BMnmg8YszUwxYtaoGoe did:polygonid:polygon:mumbai:2qNevtQ3kDbgMuV4mLGnHM7nmeHRtACJaq8etV1mC1 cash",
			expect: "did:polygonid:polygon:mumbai:2qFpPHotk6oyaX1fcrpQFT4BMnmg8YszUwxYtaoGoe",
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expect, getDIDFromQuery(tc.input))
		})
	}
}
