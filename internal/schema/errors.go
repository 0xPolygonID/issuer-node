package schema

import "fmt"

// ParseClaimError - Error type for parsing claims
type ParseClaimError struct {
	Message string
}

// Error - Implements the error interface
func (e *ParseClaimError) Error() string {
	return fmt.Sprintf("ParseClaimError: %s", e.Message)
}

// NewParseClaimError - Creates a new ParseClaimError
func NewParseClaimError(message string) *ParseClaimError {
	return &ParseClaimError{Message: message}
}

// Is - Implements the Is method of the error interface
func (e *ParseClaimError) Is(target error) bool {
	_, ok := target.(*ParseClaimError)
	return ok
}
