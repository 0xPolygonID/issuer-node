package config

import (
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
