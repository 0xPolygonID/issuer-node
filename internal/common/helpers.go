package common

import (
	"github.com/mr-tron/base58"
)

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

// IsSolanaAddress checks if the given address is a valid Solana address.
func IsSolanaAddress(addr string) bool {
	decoded, err := base58.Decode(addr)
	if err != nil {
		return false
	}
	const expectedLength = 32 // Solana addresses are 32 bytes long
	return len(decoded) == expectedLength
}
