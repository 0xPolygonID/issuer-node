package common

// ToPointer is a helper function to create a pointer to a value.
// x := &5 doesn't compile
// x := ToPointer(5) good.
func ToPointer[T any](p T) *T {
	return &p
}
