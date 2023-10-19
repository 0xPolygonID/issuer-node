package ports

import (
	// "context"
	"net/http"
)

// VerifierService is the interface implemented by the verifier service
type VerifierRepository interface {
	GetAuthRequest(w http.ResponseWriter, r *http.Request)

}