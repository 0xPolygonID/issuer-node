package api

import "context"

// GetSupportedNetworks is the controller to get supported networks
func (s *Server) GetSupportedNetworks(ctx context.Context, request GetSupportedNetworksRequestObject) (GetSupportedNetworksResponseObject, error) {
	supportedNetworks := s.networkResolver.GetSupportedNetworks()

	var supportedNetworksResponse []SupportedNetworks
	for _, sn := range supportedNetworks {
		supportedNetworksResponse = append(supportedNetworksResponse, SupportedNetworks{
			Blockchain: sn.Blockchain,
			Networks:   sn.Networks,
		})
	}
	return GetSupportedNetworks200JSONResponse(supportedNetworksResponse), nil
}
