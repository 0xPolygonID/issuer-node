package services

import "fmt"

// PublishingStateError is a special error type used to signal an error when publishing a state
type PublishingStateError struct {
	Message string
}

// Error satisfies error interface for PublishingStateError
func (e *PublishingStateError) Error() string {
	return fmt.Sprintf("Error: %s", e.Message)
}
