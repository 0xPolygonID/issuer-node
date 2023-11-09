package notifications

import (
	"encoding/json"
	"errors"

	"github.com/iden3/go-schema-processor/v2/verifiable"
)

var (
	// ErrNoService is an error when user did document doesn't contain push service
	ErrNoService = errors.New("no service in did document")
	// ErrNoPushService push service doesn't exist on did document
	ErrNoPushService = errors.New("no push service in did document")
)

// FindNotificationService returns the verifiable push service of a given didDoc
func FindNotificationService(didDoc verifiable.DIDDocument) (verifiable.PushService, error) {
	var service verifiable.PushService
	for _, s := range didDoc.Service {
		serviceBytes, err := json.Marshal(s)
		if err != nil {
			return verifiable.PushService{}, err
		}
		err = json.Unmarshal(serviceBytes, &service)
		if err != nil {
			return verifiable.PushService{}, err
		}
		if service.Type == verifiable.PushNotificationServiceType {
			return service, nil
		}
	}
	return verifiable.PushService{}, ErrNoPushService
}

// FindServiceByType returns the verifiable service of a given didDoc
func FindServiceByType(didDoc verifiable.DIDDocument, serviceType string) (verifiable.Service, error) {
	var service verifiable.Service
	for _, s := range didDoc.Service {
		serviceBytes, err := json.Marshal(s)
		if err != nil {
			return verifiable.Service{}, err
		}
		err = json.Unmarshal(serviceBytes, &service)
		if err != nil {
			return verifiable.Service{}, err
		}
		if service.Type == serviceType {
			return service, nil
		}
	}
	return verifiable.Service{}, ErrNoService
}
