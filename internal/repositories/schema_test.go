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
