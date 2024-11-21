package schema

import "fmt"

// ParsingClaimError - Error type for parsing claims
type ParsingClaimError struct {
	Message string
}

// Error - Implements the error interface
func (e *ParsingClaimError) Error() string {
	return fmt.Sprintf("ParsingClaimError: %s", e.Message)
}

// NewParsingClaimError - Creates a new ParsingClaimError
func NewParsingClaimError(message string) *ParsingClaimError {
	return &ParsingClaimError{Message: message}
}

// Is - Implements the Is method of the error interface
func (e *ParsingClaimError) Is(target error) bool {
	_, ok := target.(*ParsingClaimError)
	return ok
}
