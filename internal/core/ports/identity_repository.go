package ports

// IndentityRepository is the interface implemented by the identity service
type IndentityRepository interface {
	Save() error
}
