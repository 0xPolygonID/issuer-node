package package_manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/iden3/iden3comm/v2/packers"
)

// NewUniversalDIDResolverHandler creates a new universal DID resolver handler
func NewUniversalDIDResolverHandler(universalResolverURL string) packers.DIDResolverHandlerFunc {
	return func(did string) (*verifiable.DIDDocument, error) {
		didDoc := &verifiable.DIDDocument{}

		resp, err := http.Get(fmt.Sprintf("%s/%s", universalResolverURL, did))
		if err != nil {
			return nil, err
		}

		defer func() {
			err := resp.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var didMetadata map[string]interface{}

		err = json.Unmarshal(body, &didMetadata)
		if err != nil {
			return nil, err
		}

		doc, ok := didMetadata["didDocument"]

		if !ok {
			return nil, errors.New("did document not found")
		}

		docBts, err := json.Marshal(doc)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(docBts, &didDoc)
		if err != nil {
			return nil, err
		}

		return didDoc, nil
	}
}
