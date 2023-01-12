package rand

import (
	"crypto/rand"
	"encoding/binary"
)

// Int64 returns a new random uint64
func Int64() (uint64, error) {
	var buf [8]byte
	_, err := rand.Read(buf[:4]) // was rand.Read(buf[:])

	return binary.LittleEndian.Uint64(buf[:]), err
}
