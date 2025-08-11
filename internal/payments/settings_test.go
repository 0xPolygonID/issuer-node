package payments

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"testing"

	"github.com/iden3/iden3comm/v2/protocol"
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

	expectedSettings := Config{
		1: ChainConfig{
			ChainID:      80002,
			PaymentRails: "0xF8E49b922D5Fb00d3EdD12bd14064f275726D339",
			PaymentOption: PaymentOptionConfig{
				Name:            "AmoyNative",
				Type:            protocol.Iden3PaymentRailsRequestV1Type,
				ContractAddress: "",
				Features:        nil,
				Decimals:        18,
			},
		},
		2: ChainConfig{
			ChainID:      80002,
			PaymentRails: "0xF8E49b922D5Fb00d3EdD12bd14064f275726D339",
			PaymentOption: PaymentOptionConfig{
				Name:            "Amoy USDT",
				Type:            protocol.Iden3PaymentRailsERC20RequestV1Type,
				ContractAddress: "0x71dcc8Dc5Eb138003d3571255458Bc5692a60eD4",
				Features:        nil,
				Decimals:        6,
			},
		},
		3: ChainConfig{
			ChainID:      80002,
			PaymentRails: "0xF8E49b922D5Fb00d3EdD12bd14064f275726D339",
			PaymentOption: PaymentOptionConfig{
				Name:            "Amoy USDC",
				Type:            protocol.Iden3PaymentRailsERC20RequestV1Type,
				ContractAddress: "0x71dcc8Dc5Eb138003d3571255458Bc5692a60eD4",
				Features:        []protocol.PaymentFeatures{"EIP-2612"},
				Decimals:        6,
			},
		},
		4: ChainConfig{
			ChainID:      59141,
			PaymentRails: "0x40E3EF221AA93F6Fe997c9b0393322823Bb207d3",
			PaymentOption: PaymentOptionConfig{
				Name:            "LineaSepoliaNative",
				Type:            protocol.Iden3PaymentRailsRequestV1Type,
				ContractAddress: "",
				Features:        nil,
				Decimals:        18,
			},
		},
		5: ChainConfig{
			ChainID:      59141,
			PaymentRails: "0x40E3EF221AA93F6Fe997c9b0393322823Bb207d3",
			PaymentOption: PaymentOptionConfig{
				Name:            "Linea Sepolia USDT",
				Type:            protocol.Iden3PaymentRailsERC20RequestV1Type,
				ContractAddress: "0xb0101c1Ffdd1213B886FebeF6F07442e48990c9C",
				Features:        nil,
				Decimals:        18,
			},
		},
		6: ChainConfig{
			ChainID:      59141,
			PaymentRails: "0x40E3EF221AA93F6Fe997c9b0393322823Bb207d3",
			PaymentOption: PaymentOptionConfig{
				Name:            "Linea Sepolia USDC",
				Type:            protocol.Iden3PaymentRailsERC20RequestV1Type,
				ContractAddress: "0xb0101c1Ffdd1213B886FebeF6F07442e48990c9C",
				Features:        []protocol.PaymentFeatures{"EIP-2612"},
				Decimals:        18,
			},
		},
		7: ChainConfig{
			ChainID:      2442,
			PaymentRails: "0x09c269e74d8B47c98537Acd6CbEe8056806F4c70",
			PaymentOption: PaymentOptionConfig{
				Name:            "ZkEvmNative",
				Type:            protocol.Iden3PaymentRailsRequestV1Type,
				ContractAddress: "",
				Features:        nil,
				Decimals:        18,
			},
		},
		8: ChainConfig{
			ChainID:      2442,
			PaymentRails: "0x09c269e74d8B47c98537Acd6CbEe8056806F4c70",
			PaymentOption: PaymentOptionConfig{
				Name:            "ZkEvm USDT",
				Type:            protocol.Iden3PaymentRailsERC20RequestV1Type,
				ContractAddress: "0x986caE6ADcF5da2a1514afc7317FBdeE0B4048Db",
				Features:        nil,
				Decimals:        18,
			},
		},
		9: ChainConfig{
			ChainID:      2442,
			PaymentRails: "0x09c269e74d8B47c98537Acd6CbEe8056806F4c70",
			PaymentOption: PaymentOptionConfig{
				Name:            "ZkEvm USDC",
				Type:            protocol.Iden3PaymentRailsERC20RequestV1Type,
				ContractAddress: "0x986caE6ADcF5da2a1514afc7317FBdeE0B4048Db",
				Features:        []protocol.PaymentFeatures{"EIP-2612"},
				Decimals:        18,
			},
		},
		10: ChainConfig{
			ChainID:      103,
			PaymentRails: "Hys6CpX8McHbPBaPKbRYGVdXVxor1M5pSZUDMMwakGmM",
			PaymentOption: PaymentOptionConfig{
				Name:     "Solana devnet",
				Type:     protocol.Iden3PaymentRailsSolanaRequestV1Type,
				Decimals: 9,
			},
		},
		11: ChainConfig{
			ChainID:      103,
			PaymentRails: "Hys6CpX8McHbPBaPKbRYGVdXVxor1M5pSZUDMMwakGmM",
			PaymentOption: PaymentOptionConfig{
				Name:            "Solana devnet test SPL",
				Type:            protocol.Iden3PaymentRailsSolanaSPLRequestV1Type,
				ContractAddress: "4MjRhSkDaXmgdAL9d9UM7kmgJrWYGJH66oocUN2f3VUp",
				Decimals:        9,
			},
		},
	}

	type expected struct {
		err      error
		settings Config
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
				err: errors.New("illegal base64 data at input byte 5"),
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
				require.NoError(t, err)
				require.EqualValues(t, tc.expected.settings, *settings)
			} else {
				require.Equal(t, tc.expected.err.Error(), err.Error())
			}
		})
	}
}
