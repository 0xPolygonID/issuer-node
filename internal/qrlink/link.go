package qrlink

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
)

const (
	qRStoreUrl              = "iden3comm://?request_uri=%s/v2/qr-store?id=%s"
	qRStoreUrlWithDID       = "iden3comm://?request_uri=%s/v2/qr-store?id=%s&issuer=%s"
	universalLinkURL        = "%s#request_uri=%s/v2/qr-store?id=%s"
	universalLinkURLWithDID = "%s#request_uri=%s/v2/qr-store?id=%s&issuer=%s"
)

// NewDeepLink creates a deep link
// If issuerDID is nil, it will return a deep link without the issuer DID for backward compatibility
func NewDeepLink(hostURL string, id uuid.UUID, issuerDID *w3c.DID) string {
	if issuerDID != nil {
		return fmt.Sprintf(qRStoreUrlWithDID, hostURL, id.String(), issuerDID.String())
	}
	return fmt.Sprintf(qRStoreUrl, hostURL, id.String())
}

// NewUniversal creates a universal link
// If issuerDID is nil, it will return a universal link without the issuer DID for backward compatibility
func NewUniversal(uLinkBaseUrl string, hostURL string, id uuid.UUID, issuerDID *w3c.DID) string {
	if issuerDID != nil {
		return fmt.Sprintf(universalLinkURLWithDID, uLinkBaseUrl, hostURL, id.String(), issuerDID.String())
	}
	return fmt.Sprintf(universalLinkURL, uLinkBaseUrl, hostURL, id.String())
}
