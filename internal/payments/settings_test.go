package payments

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
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
			PaymentRails: ethCommon.HexToAddress("0xF8E49b922D5Fb00d3EdD12bd14064f275726D339"),
			PaymentOption: PaymentOptionConfig{
				Name:            "AmoyNative",
				Type:            protocol.Iden3PaymentRailsRequestV1Type,
				ContractAddress: ethCommon.HexToAddress(""),
				Features:        nil,
			},
		},
		2: ChainConfig{
			ChainID:      80002,
			PaymentRails: ethCommon.HexToAddress("0xF8E49b922D5Fb00d3EdD12bd14064f275726D339"),
			PaymentOption: PaymentOptionConfig{
				Name:            "Amoy USDT",
				Type:            protocol.Iden3PaymentRailsERC20RequestV1Type,
				ContractAddress: ethCommon.HexToAddress("0x2FE40749812FAC39a0F380649eF59E01bccf3a1A"),
				Features:        nil,
			},
		},
		3: ChainConfig{
			ChainID:      80002,
			PaymentRails: ethCommon.HexToAddress("0xF8E49b922D5Fb00d3EdD12bd14064f275726D339"),
			PaymentOption: PaymentOptionConfig{
				Name:            "Amoy USDC",
				Type:            protocol.Iden3PaymentRailsERC20RequestV1Type,
				ContractAddress: ethCommon.HexToAddress("0x2FE40749812FAC39a0F380649eF59E01bccf3a1A"),
				Features:        []protocol.PaymentFeatures{"EIP-2612"},
			},
		},
		4: ChainConfig{
			ChainID:      59141,
			PaymentRails: ethCommon.HexToAddress("0x40E3EF221AA93F6Fe997c9b0393322823Bb207d3"),
			PaymentOption: PaymentOptionConfig{
				Name:            "LineaSepoliaNative",
				Type:            protocol.Iden3PaymentRailsRequestV1Type,
				ContractAddress: ethCommon.HexToAddress(""),
				Features:        nil,
			},
		},
		5: ChainConfig{
			ChainID:      59141,
			PaymentRails: ethCommon.HexToAddress("0x40E3EF221AA93F6Fe997c9b0393322823Bb207d3"),
			PaymentOption: PaymentOptionConfig{
				Name:            "Linea Sepolia USDT",
				Type:            protocol.Iden3PaymentRailsERC20RequestV1Type,
				ContractAddress: ethCommon.HexToAddress("0xb0101c1Ffdd1213B886FebeF6F07442e48990c9C"),
				Features:        nil,
			},
		},
		6: ChainConfig{
			ChainID:      59141,
			PaymentRails: ethCommon.HexToAddress("0x40E3EF221AA93F6Fe997c9b0393322823Bb207d3"),
			PaymentOption: PaymentOptionConfig{
				Name:            "Linea Sepolia USDC",
				Type:            protocol.Iden3PaymentRailsERC20RequestV1Type,
				ContractAddress: ethCommon.HexToAddress("0xb0101c1Ffdd1213B886FebeF6F07442e48990c9C"),
				Features:        []protocol.PaymentFeatures{"EIP-2612"},
			},
		},
		7: ChainConfig{
			ChainID:      2442,
			PaymentRails: ethCommon.HexToAddress("0x09c269e74d8B47c98537Acd6CbEe8056806F4c70"),
			PaymentOption: PaymentOptionConfig{
				Name:            "ZkEvmNative",
				Type:            protocol.Iden3PaymentRailsRequestV1Type,
				ContractAddress: ethCommon.HexToAddress(""),
				Features:        nil,
			},
		},
		8: ChainConfig{
			ChainID:      2442,
			PaymentRails: ethCommon.HexToAddress("0x09c269e74d8B47c98537Acd6CbEe8056806F4c70"),
			PaymentOption: PaymentOptionConfig{
				Name:            "ZkEvm USDT",
				Type:            protocol.Iden3PaymentRailsERC20RequestV1Type,
				ContractAddress: ethCommon.HexToAddress("0x986caE6ADcF5da2a1514afc7317FBdeE0B4048Db"),
				Features:        nil,
			},
		},
		9: ChainConfig{
			ChainID:      2442,
			PaymentRails: ethCommon.HexToAddress("0x09c269e74d8B47c98537Acd6CbEe8056806F4c70"),
			PaymentOption: PaymentOptionConfig{
				Name:            "ZkEvm USDC",
				Type:            protocol.Iden3PaymentRailsERC20RequestV1Type,
				ContractAddress: ethCommon.HexToAddress("0x986caE6ADcF5da2a1514afc7317FBdeE0B4048Db"),
				Features:        []protocol.PaymentFeatures{"EIP-2612"},
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
