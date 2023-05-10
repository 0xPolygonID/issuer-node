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

func TestTokenizeQuery(t *testing.T) {
	type testConfig struct {
		name   string
		input  string
		expect []string
	}
	for _, tc := range []testConfig{
		{
			name:   "empty string",
			input:  "",
			expect: []string{},
		},
		{
			name:   "one word",
			input:  "word",
			expect: []string{"word"},
		},
		{
			name:   "one word with spaces",
			input:  "    word   ",
			expect: []string{"word"},
		},
		{
			name:   "some  words with spaces",
			input:  "    word1   word2 word3",
			expect: []string{"word1", "word2", "word3"},
		},
		{
			name:   "some  words with spaces and commas",
			input:  "word1,    word2, word3",
			expect: []string{"word1", "word2", "word3"},
		},
		{
			name:   "repeated words are filtered",
			input:  "one two three one two three",
			expect: []string{"one", "two", "three"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, tokenizeQuery(tc.input))
		})
	}
}

func TestBuildPartialQueryDidLikes(t *testing.T) {
	type testConfig struct {
		name   string
		field  string
		input  []string
		cond   string
		expect string
	}
	for _, tc := range []testConfig{
		{
			name:   "nil input",
			field:  "did",
			input:  nil,
			cond:   "OR",
			expect: "",
		},
		{
			name:   "Empty list",
			field:  "did",
			input:  []string{},
			cond:   "OR",
			expect: "",
		},
		{
			name:   "One item",
			field:  "did",
			input:  []string{"lala"},
			cond:   "OR",
			expect: "did ILIKE '%lala%'",
		},
		{
			name:   "One item, with chars not valid in a did",
			field:  "did",
			input:  []string{"valid$^;#:@()notvalid"},
			cond:   "OR",
			expect: "did ILIKE '%valid:notvalid%'",
		},
		{
			name:   "sql injection",
			field:  "did",
			input:  []string{";DROP TABLE users;"},
			cond:   "OR",
			expect: "did ILIKE '%DROPTABLEusers%'",
		},
		{
			name:   "Some items",
			field:  "did",
			input:  []string{"search1", "search2", "search3"},
			cond:   "OR",
			expect: "did ILIKE '%search1%' OR did ILIKE '%search2%' OR did ILIKE '%search3%'",
		},
		{
			name:   "Some items. Empty words are filtered",
			field:  "did",
			input:  []string{"", "search1", "", "", "search2", "", "search3", ""},
			cond:   "OR",
			expect: "did ILIKE '%search1%' OR did ILIKE '%search2%' OR did ILIKE '%search3%'",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, buildPartialQueryDidLikes(tc.field, tc.input, tc.cond))
		})
	}
}

func TestBuildPartialQueryLikes(t *testing.T) {
	type testConfig struct {
		name   string
		field  string
		cond   string
		first  int
		n      int
		expect string
	}
	for _, tc := range []testConfig{
		{
			name:   "empty",
			field:  "field",
			cond:   "OR",
			first:  1,
			n:      0,
			expect: "",
		},
		{
			name:   "one field",
			field:  "field",
			cond:   "OR",
			first:  1,
			n:      1,
			expect: "field ILIKE '%' || $1 || '%'",
		},
		{
			name:   "2 fields",
			field:  "field",
			cond:   "OR",
			first:  1,
			n:      2,
			expect: "field ILIKE '%' || $1 || '%' OR field ILIKE '%' || $2 || '%'",
		},
		{
			name:   "2 fields, starting at 3",
			field:  "field",
			cond:   "OR",
			first:  3,
			n:      2,
			expect: "field ILIKE '%' || $3 || '%' OR field ILIKE '%' || $4 || '%'",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, buildPartialQueryLikes(tc.field, tc.cond, tc.first, tc.n))
		})
	}
}
