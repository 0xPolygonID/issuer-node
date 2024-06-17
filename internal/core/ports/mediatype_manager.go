package ports

import "github.com/iden3/iden3comm/v2"

// MediatypeManager - Define the interface for media type manager
type MediatypeManager interface {
	AllowMediaType(protoclMessage iden3comm.ProtocolMessage, mediaType iden3comm.MediaType) bool
}
