package loaders

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/iden3/go-circuits"
)

const (
	wasmFile            = "circuit.wasm"
	proofingKeyFile     = "circuit_final.zkey"
	verificationKeyFile = "authV2.json"
)

// CircuitFilesSet set circuits files.
type CircuitFilesSet struct {
	Wasm            []byte
	ProofKey        []byte
	VerificationKey []byte
}

// Circuits load circuits key.
type Circuits struct {
	basePath string
}

// NewCircuits create loader that returns circuits files.
func NewCircuits(basePath string) *Circuits {
	return &Circuits{basePath: basePath}
}

// Load circuits files by circuitID.
func (l *Circuits) Load(circuitID circuits.CircuitID) (*CircuitFilesSet, error) {
	rawWasmFile, err := l.LoadWasm(circuitID)
	if err != nil {
		return nil, err
	}
	rawProofKeyFile, err := l.LoadProvingKey(circuitID)
	if err != nil {
		return nil, err
	}
	rawVerificationKeyFile, err := l.LoadVerificationKey(circuitID)
	if err != nil {
		return nil, err
	}

	return &CircuitFilesSet{
		Wasm:            rawWasmFile,
		ProofKey:        rawProofKeyFile,
		VerificationKey: rawVerificationKeyFile,
	}, nil
}

// LoadVerificationKey load verification key by circuit ID.
func (l *Circuits) LoadVerificationKey(circuitID circuits.CircuitID) ([]byte, error) {
	return l.getPathToFile(circuitID, verificationKeyFile)
}

// LoadProvingKey load proof key by circuit ID.
func (l *Circuits) LoadProvingKey(circuitID circuits.CircuitID) ([]byte, error) {
	return l.getPathToFile(circuitID, proofingKeyFile)
}

// LoadWasm load wasm file by circuit ID.
func (l *Circuits) LoadWasm(circuitID circuits.CircuitID) ([]byte, error) {
	return l.getPathToFile(circuitID, wasmFile)
}

func (l *Circuits) getPathToFile(circuitID circuits.CircuitID, fileName string) ([]byte, error) {
	path := filepath.Join(l.basePath, string(circuitID), fileName)
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed open file '%s' by path '%s': %v", fileName, path, err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed read file '%s' by path '%s': %v", fileName, path, err)
	}
	return data, nil
}
