package api

import (
	"context"
)

// GetSupportedNetworks is the controller to get supported networks
func (s *Server) GetSupportedNetworks(ctx context.Context, _ GetSupportedNetworksRequestObject) (GetSupportedNetworksResponseObject, error) {
	supportedNetworks := s.networkResolver.GetSupportedNetworks()

	var supportedNetworksResponse []SupportedNetworks
	for _, sn := range supportedNetworks {
		supportedNetworksResponse = append(supportedNetworksResponse, SupportedNetworks{
			Blockchain: sn.Blockchain,
			Networks:   s.supportedNetworksData(ctx, sn.Blockchain, sn.Networks),
		})
	}
	return GetSupportedNetworks200JSONResponse(supportedNetworksResponse), nil
}

func (s *Server) supportedNetworksData(ctx context.Context, blockchain string, networks []string) []NetworkData {
	var nd []NetworkData
	for _, network := range networks {
		rhsSettings, err := s.networkResolver.GetRhsSettings(ctx, blockchain+":"+network)
		if err != nil {
			continue
		}
		types := s.networkResolver.SupportedCredentialStatusTypes(rhsSettings.Mode)
		credentialStatusTypes := make([]string, 0, len(types))
		for _, t := range types {
			credentialStatusTypes = append(credentialStatusTypes, string(t))
		}
		nd = append(nd, NetworkData{
			Name:             network,
			CredentialStatus: credentialStatusTypes,
		})
	}
	return nd
}
