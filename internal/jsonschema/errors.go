package jsonschema

// CredentialLinkAttributeError - Something wrong happened when we try to convert a credential attribute value.
type CredentialLinkAttributeError struct {
	message string
}

func newCredentialLinkAttributeError(message string) error {
	return &CredentialLinkAttributeError{
		message: message,
	}
}

func (e *CredentialLinkAttributeError) Error() string {
	return e.message
}
