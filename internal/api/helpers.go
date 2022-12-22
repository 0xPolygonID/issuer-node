package api

func ToPointer[T any](p T) *T {
	return &p
}
