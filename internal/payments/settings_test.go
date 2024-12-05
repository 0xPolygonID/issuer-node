package payments

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/config"
)

func TestSettingsFromConfig(t *testing.T) {
	ctx := context.Background()
	filePath := "testdata/payment_settings.test.yaml"
	fileContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	fileContentUrlBase64 := base64.StdEncoding.EncodeToString(fileContent)

	expectedSettings := Settings{
		137: {
			MCPayment: "0x380dd90852d3Fe75B4f08D0c47416D6c4E0dC774",
			ERC20: &ERC20{
				USDT: Token{
					ContractAddress: "0xc2132D05D31c914a87C6611C10748AEb04B58e8F",
					Features:        []string{},
				},
				USDC: Token{
					ContractAddress: "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359",
					Features:        []string{"EIP-2612"},
				},
			},
		},
		80002: {
			MCPayment: "0x380dd90852d3Fe75B4f08D0c47416D6c4E0dC774",
		},
		1101: {
			MCPayment: "0x380dd90852d3Fe75B4f08D0c47416D6c4E0dC774",
		},
	}

	type expected struct {
		err      error
		settings Settings
	}

	for _, tc := range []struct {
		name     string
		cfg      config.Payments
		expected expected
	}{
		{
			name: "Config from file content base64 encoded",
			cfg: config.Payments{
				SettingsPath: "",
				SettingsFile: common.ToPointer(fileContentUrlBase64),
			},
			expected: expected{
				settings: expectedSettings,
			},
		},
		{
			name: "Config from file path",
			cfg: config.Payments{
				SettingsPath: filePath,
				SettingsFile: nil,
			},
			expected: expected{
				settings: expectedSettings,
			},
		},
		{
			name: "Config from file has preference",
			cfg: config.Payments{
				SettingsPath: filePath,
				SettingsFile: common.ToPointer("irrelevant wrong content"),
			},
			expected: expected{
				settings: expectedSettings,
			},
		},
		{
			name: "No file or path configured",
			cfg: config.Payments{
				SettingsPath: "",
				SettingsFile: nil,
			},
			expected: expected{
				err: errors.New("no payment settings file or payment file path provided"),
			},
		},
		{
			name: "Wrong file content. Not base64 encoded",
			cfg: config.Payments{
				SettingsPath: "",
				SettingsFile: common.ToPointer(string(fileContent)),
			},
			expected: expected{
				err: errors.New("illegal base64 data at input byte 3"),
			},
		},
		{
			name: "Wrong file path",
			cfg: config.Payments{
				SettingsPath: "/wrong/path",
			},
			expected: expected{
				err: errors.New("payment settings file not found: stat /wrong/path: no such file or directory"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			settings, err := SettingsFromConfig(ctx, &tc.cfg)
			if tc.expected.err == nil {
				require.EqualValues(t, &tc.expected.settings, settings)
			} else {
				require.Equal(t, tc.expected.err.Error(), err.Error())
			}
		})
	}
}
