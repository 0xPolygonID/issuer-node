package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupVaultTokenFromFile(t *testing.T) {
	token, err := lookupVaultTokenFromFile("file does not exist")
	assert.Empty(t, token)
	assert.Error(t, err)

	token, err = lookupVaultTokenFromFile("internal/config/testdata/init.out.bad")
	assert.Empty(t, token)
	assert.Error(t, err)

	token, err = lookupVaultTokenFromFile("internal/config/testdata/init.out.good")
	assert.NoError(t, err)
	assert.Equal(t, "hvs.xAIi0RxVOTfwSNYivBhb3Gfp", token)
}

func TestConfiguration_validateServerUrl(t *testing.T) {
	type expected struct {
		url   string
		error bool
	}
	type testConfig struct {
		name     string
		url      string
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name: "Empty url",
			url:  "",
			expected: expected{
				url:   "",
				error: true,
			},
		},
		{
			name: "wrong url",
			url:  "wrong",
			expected: expected{
				url:   "wrong",
				error: true,
			},
		},
		{
			name: "Relative url",
			url:  "/relative/url",
			expected: expected{
				url:   "/relative/url",
				error: true,
			},
		},
		{
			name: "Simple url",
			url:  "schema://site.org",
			expected: expected{
				url:   "schema://site.org",
				error: false,
			},
		},
		{
			name: "Url ending with a slash. Slash will be removed",
			url:  "schema://site.org/",
			expected: expected{
				url:   "schema://site.org",
				error: false,
			},
		},
		{
			name: "Url ending with multiple slashes. Slashes will be removed",
			url:  "schema://site.org///////",
			expected: expected{
				url:   "schema://site.org",
				error: false,
			},
		},
		{
			name: "Url with ?",
			url:  "schema://site.org?",
			expected: expected{
				url:   "schema://site.org",
				error: false,
			},
		},
		{
			name: "Url with query params",
			url:  "schema://site.org?p=1&q=2",
			expected: expected{
				url:   "schema://site.org",
				error: false,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Configuration{
				ServerUrl: tc.url,
			}
			sURL, err := cfg.validateServerUrl()
			if tc.expected.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expected.url, sURL)
		})
	}
}

func TestVerifierStateContracts_Parse(t *testing.T) {
	type testCase struct {
		name      string
		input     VerifierStateContracts
		addresses map[string]string
		rpcs      map[string]string
		err       error
	}

	tests := []testCase{
		{
			name: "Valid input",
			input: VerifierStateContracts{
				Addresses: "chain1=address1;chain2=address2",
				RPCs:      "chain1=rpc1;chain2=rpc2",
			},
			addresses: map[string]string{
				"chain1": "address1",
				"chain2": "address2",
			},
			rpcs: map[string]string{
				"chain1": "rpc1",
				"chain2": "rpc2",
			},
			err: nil,
		},
		{
			name: "Mismatched lengths",
			input: VerifierStateContracts{
				Addresses: "chain1=address1",
				RPCs:      "chain1=rpc1;chain2=rpc2",
			},
			addresses: nil,
			rpcs:      nil,
			err:       fmt.Errorf("addresses and rpcs must have the same length"),
		},
		{
			name: "Invalid address format",
			input: VerifierStateContracts{
				Addresses: "chain1address1;chain2=address2",
				RPCs:      "chain1=rpc1;chain2=rpc2",
			},
			addresses: nil,
			rpcs:      nil,
			err:       fmt.Errorf("error parsing addresses: pair must have the format chain=resource"),
		},
		{
			name: "Invalid rpc format",
			input: VerifierStateContracts{
				Addresses: "chain1=address1;chain2=address2",
				RPCs:      "chain1rpc1;chain2=rpc2",
			},
			addresses: nil,
			rpcs:      nil,
			err:       fmt.Errorf("error parsing rpcs: pair must have the format chain=resource"),
		},
		{
			name: "Single pair of addresses and rpcs",
			input: VerifierStateContracts{
				Addresses: "chain1=address1",
				RPCs:      "chain1=rpc1",
			},
			addresses: map[string]string{
				"chain1": "address1",
			},
			rpcs: map[string]string{
				"chain1": "rpc1",
			},
			err: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			addresses, rpcs, err := tc.input.Parse()
			if tc.err != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.err.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.addresses, addresses)
				assert.Equal(t, tc.rpcs, rpcs)
			}
		})
	}
}
