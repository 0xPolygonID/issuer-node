package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type envVarsT map[string]string

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

func TestLoad(t *testing.T) {
	envVars := initVariables(t)
	loadEnvironmentVariables(t, envVars)
	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "https://issuer-node.privado.id/issuer", cfg.ServerUrl)
	assert.Equal(t, 3001, cfg.ServerPort)
	assert.Equal(t, "pbkey", cfg.PublishingKeyPath)
	assert.Equal(t, time.Duration(60000000000), cfg.OnChainCheckStatusFrequency)
	assert.Equal(t, "postgres://polygonid:polygonid@localhost:5432/platformid?sslmode=disable", cfg.Database.URL)
	assert.Equal(t, -4, cfg.Log.Level)
	assert.Equal(t, 1, cfg.Log.Mode)
	assert.Equal(t, "user-issuer", cfg.HTTPBasicAuth.User)
	assert.Equal(t, "password-issuer", cfg.HTTPBasicAuth.Password)
	assert.Equal(t, "https://gateway.pinata.cloud", cfg.IPFS.GatewayURL)
	assert.Equal(t, "https://vault.privado.id", cfg.KeyStore.Address)
	assert.Equal(t, "iden3", cfg.KeyStore.PluginIden3MountPath)
	assert.Equal(t, "/localstorage", cfg.KeyStore.ProviderLocalStorageFilePath)
	assert.True(t, true, cfg.KeyStore.VaultUserPassAuthEnabled)
	assert.Equal(t, "issuernodepwd", cfg.KeyStore.VaultUserPassAuthPassword)
	assert.Equal(t, "localstorage", cfg.KeyStore.BJJProvider)
	assert.Equal(t, "localstorage", cfg.KeyStore.ETHProvider)
	assert.Equal(t, "XYZ", cfg.KeyStore.AWSAccessKey)
	assert.Equal(t, "123HHUBUuO5", cfg.KeyStore.AWSSecretKey)
	assert.Equal(t, "eu-west-1", cfg.KeyStore.AWSRegion)
	assert.Equal(t, "./resolvers_settings.yaml", cfg.NetworkResolverPath)
	assert.Equal(t, "hvs.NK8jrOU4XNY", cfg.KeyStore.Token)
	assert.Equal(t, "123", *cfg.NetworkResolverFile)
	assert.Equal(t, "./pkg/credentials/circuits", cfg.Circuit.Path)
	assert.Equal(t, "redis://@localhost:6379/1", cfg.Cache.Url)
	assert.Equal(t, "redis", cfg.Cache.Provider)
	assert.True(t, *cfg.MediaTypeManager.Enabled)
}

func TestLoadKmsProviders(t *testing.T) {
	envVars := initVariables(t)
	envVars["ISSUER_KMS_BJJ_PROVIDER"] = ""
	envVars["ISSUER_KMS_ETH_PROVIDER"] = ""
	loadEnvironmentVariables(t, envVars)
	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "localstorage", cfg.KeyStore.BJJProvider)
	assert.Equal(t, "localstorage", cfg.KeyStore.ETHProvider)

	envVars["ISSUER_KMS_ETH_PROVIDER"] = "aws"
	envVars["ISSUER_KMS_ETH_PLUGIN_AWS_ACCESS_KEY"] = ""
	loadEnvironmentVariables(t, envVars)
	_, err = Load()
	assert.Error(t, err)

	envVars["ISSUER_KMS_ETH_PROVIDER"] = "aws"
	envVars["ISSUER_KMS_ETH_PLUGIN_AWS_ACCESS_KEY"] = "123"
	envVars["ISSUER_KMS_ETH_PLUGIN_AWS_SECRET_KEY"] = ""
	loadEnvironmentVariables(t, envVars)
	_, err = Load()
	assert.Error(t, err)

	envVars["ISSUER_KMS_ETH_PROVIDER"] = "aws"
	envVars["ISSUER_KMS_ETH_PLUGIN_AWS_ACCESS_KEY"] = "123"
	envVars["ISSUER_KMS_ETH_PLUGIN_AWS_SECRET_KEY"] = "456"
	envVars["ISSUER_KMS_ETH_PLUGIN_AWS_REGION"] = ""
	loadEnvironmentVariables(t, envVars)
	_, err = Load()
	assert.Error(t, err)
}

func TestLoadKmsProvidersFolder(t *testing.T) {
	envVars := initVariables(t)
	envVars["ISSUER_KMS_PROVIDER_LOCAL_STORAGE_FILE_PATH"] = "./newfolder"
	loadEnvironmentVariables(t, envVars)
	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "localstorage", cfg.KeyStore.BJJProvider)
	assert.Equal(t, "localstorage", cfg.KeyStore.ETHProvider)
	assert.Equal(t, "./newfolder", cfg.KeyStore.ProviderLocalStorageFilePath)
}

func TestLoadNetworkResolver(t *testing.T) {
	envVars := initVariables(t)
	envVars["ISSUER_RESOLVER_PATH"] = ""
	envVars["ISSUER_RESOLVER_FILE"] = "an encoded base 64 file content"

	loadEnvironmentVariables(t, envVars)
	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "", cfg.NetworkResolverPath)
	assert.Equal(t, "an encoded base 64 file content", *cfg.NetworkResolverFile)

	envVars["ISSUER_RESOLVER_PATH"] = ""
	envVars["ISSUER_RESOLVER_FILE"] = ""
	loadEnvironmentVariables(t, envVars)
	cfg, err = Load()
	assert.NoError(t, err)
	assert.Equal(t, "./resolvers_settings.yaml", cfg.NetworkResolverPath)
}

func TestLoadCacheProvider(t *testing.T) {
	envVars := initVariables(t)
	envVars["ISSUER_CACHE_PROVIDER"] = ""
	loadEnvironmentVariables(t, envVars)
	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "redis", cfg.Cache.Provider)

	envVars["ISSUER_CACHE_URL"] = ""
	loadEnvironmentVariables(t, envVars)
	_, err = Load()
	assert.Error(t, err)
}

func initVariables(t *testing.T) envVarsT {
	t.Helper()
	envVars := map[string]string{
		"ISSUER_SERVER_URL":                           "https://issuer-node.privado.id/issuer",
		"ISSUER_SERVER_PORT":                          "3001",
		"ISSUER_NATIVE_PROOF_GENERATION_ENABLED":      "true",
		"ISSUER_PUBLISH_KEY_PATH":                     "pbkey",
		"ISSUER_ONCHAIN_PUBLISH_STATE_FREQUENCY":      "1m",
		"ISSUER_ONCHAIN_CHECK_STATUS_FREQUENCY":       "1m",
		"ISSUER_DATABASE_URL":                         "postgres://polygonid:polygonid@localhost:5432/platformid?sslmode=disable",
		"ISSUER_LOG_LEVEL":                            "-4",
		"ISSUER_LOG_MODE":                             "1",
		"ISSUER_API_AUTH_USER":                        "user-issuer",
		"ISSUER_API_AUTH_PASSWORD":                    "password-issuer",
		"ISSUER_IPFS_GATEWAY_URL":                     "https://gateway.pinata.cloud",
		"ISSUER_KEY_STORE_ADDRESS":                    "https://vault.privado.id",
		"ISSUER_KEY_STORE_PLUGIN_IDEN3_MOUNT_PATH":    "iden3",
		"ISSUER_KMS_PROVIDER_LOCAL_STORAGE_FILE_PATH": "/localstorage",
		"ISSUER_VAULT_USERPASS_AUTH_ENABLED":          "true",
		"ISSUER_VAULT_USERPASS_AUTH_PASSWORD":         "issuernodepwd",
		"ISSUER_KMS_BJJ_PROVIDER":                     "localstorage",
		"ISSUER_KMS_ETH_PROVIDER":                     "localstorage",
		"ISSUER_KMS_ETH_PLUGIN_AWS_ACCESS_KEY":        "XYZ",
		"ISSUER_KMS_ETH_PLUGIN_AWS_SECRET_KEY":        "123HHUBUuO5",
		"ISSUER_KMS_ETH_PLUGIN_AWS_REGION":            "eu-west-1",
		"ISSUER_RESOLVER_PATH":                        "./resolvers_settings.yaml",
		"ISSUER_KEY_STORE_TOKEN":                      "hvs.NK8jrOU4XNY",
		"ISSUER_RESOLVER_FILE":                        "123",
		"ISSUER_CIRCUIT_PATH":                         "./pkg/credentials/circuits",
		"ISSUER_REDIS_URL":                            "redis://@localhost:6379/1",
		"ISSUER_MEDIA_TYPE_MANAGER_ENABLED":           "true",
		"ISSUER_CACHE_PROVIDER":                       "redis",
		"ISSUER_CACHE_URL":                            "redis://@localhost:6379/1",
	}
	return envVars
}

func loadEnvironmentVariables(t *testing.T, envVars envVarsT) {
	t.Helper()
	for key, value := range envVars {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}
}
