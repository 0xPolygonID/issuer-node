package common

// ToPointer is a helper function to create a pointer to a value.
// x := &5 doesn't compile
// x := ToPointer(5) good.
func ToPointer[T any](p T) *T {
	return &p
}

// CopyMap returns a deep copy of the input map
func CopyMap(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})
	for k, v := range m {
		vm, ok := v.(map[string]interface{})
		if ok {
			cp[k] = CopyMap(vm)
		} else {
			cp[k] = v
		}
	}

	return cp
}
