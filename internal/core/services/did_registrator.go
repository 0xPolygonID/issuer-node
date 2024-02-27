package services

import (
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/polygonid/sh-id-platform/internal/config"
)

// RegisterCustomDIDMethods registers custom DID methods
func RegisterCustomDIDMethods(customdDids []config.CustomDIDMethods) {
	for _, cdid := range customdDids {
		params := core.DIDMethodNetworkParams{
			Method:      core.DIDMethodPolygonID,
			Blockchain:  core.Blockchain(cdid.Blockchain),
			Network:     core.NetworkID(cdid.Network),
			NetworkFlag: cdid.NetworkFlag,
		}
		core.RegisterDIDMethodNetwork(params, core.WithChainID(cdid.ChainID))
	}
}
