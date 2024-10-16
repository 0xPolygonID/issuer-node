package common

import (
	"io"
	"testing"
)

// MyYAMLReader - Helper for testing
type MyYAMLReader struct {
	data   []byte
	offset int
}

// NewMyYAMLReader - Helper for testing
func NewMyYAMLReader(yamlData []byte) *MyYAMLReader {
	return &MyYAMLReader{data: yamlData}
}

func (r *MyYAMLReader) Read(p []byte) (int, error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}

	n := copy(p, r.data[r.offset:])
	r.offset += n

	return n, nil
}

// CreateFile - Helper for testing
func CreateFile(t *testing.T) *MyYAMLReader {
	t.Helper()
	yamlData := []byte(`polygon:
  amoy:
    contractAddress: 0x1a4cC30f2aA0377b0c3bc9848766D90cb4404124
    networkURL: https://polygon-amoy.g.alchemy.com/v2/123
    defaultGasLimit: 600000
    confirmationTimeout: 10s
    confirmationBlockCount: 5
    receiptTimeout: 600s
    minGasPrice: 0
    maxGasPrice: 1000000
    rpcResponseTimeout: 5s
    waitReceiptCycleTime: 30s
    waitBlockCycleTime: 30s
    gasLess: false
    transferAmountWei: 1000000000000000000
    rhsSettings:
      mode: None
      rhsUrl: https://rhs-staging.polygonid.me
      contractAddress: 0x3d3763eC0a50CE1AdF83d0b5D99FBE0e3fEB43fb
      chainID: 80002
      publishingKey: pbkey
`)

	reader := NewMyYAMLReader(yamlData)
	return reader
}

// CreateFileWithRHSOffChain - Helper for testing
func CreateFileWithRHSOffChain(t *testing.T) *MyYAMLReader {
	t.Helper()
	yamlData := []byte(`polygon:
  amoy:
    contractAddress: 0x1a4cC30f2aA0377b0c3bc9848766D90cb4404124
    networkURL: https://polygon-amoy.g.alchemy.com/v2/123
    defaultGasLimit: 600000
    confirmationTimeout: 10s
    confirmationBlockCount: 5
    receiptTimeout: 600s
    minGasPrice: 0
    maxGasPrice: 1000000
    rpcResponseTimeout: 5s
    waitReceiptCycleTime: 30s
    waitBlockCycleTime: 30s
    gasLess: false
    transferAmountWei: 1000000000000000000
    rhsSettings:
      mode: OffChain
      rhsUrl: https://rhs-staging.polygonid.me
      contractAddress: 0x3d3763eC0a50CE1AdF83d0b5D99FBE0e3fEB43fb
      chainID: 80002
      publishingKey: pbkey
`)

	reader := NewMyYAMLReader(yamlData)
	return reader
}

// CreateFileWithRHSOnChain - Helper for testing
func CreateFileWithRHSOnChain(t *testing.T) *MyYAMLReader {
	t.Helper()
	yamlData := []byte(`polygon:
  amoy:
    contractAddress: 0x1a4cC30f2aA0377b0c3bc9848766D90cb4404124
    networkURL: https://polygon-amoy.g.alchemy.com/v2/123
    defaultGasLimit: 600000
    confirmationTimeout: 10s
    confirmationBlockCount: 5
    receiptTimeout: 600s
    minGasPrice: 0
    maxGasPrice: 1000000
    rpcResponseTimeout: 5s
    waitReceiptCycleTime: 30s
    waitBlockCycleTime: 30s
    gasLess: false
    transferAmountWei: 1000000000000000000
    rhsSettings:
      mode: OnChain
      rhsUrl: https://rhs-staging.polygonid.me
      contractAddress: 0x3d3763eC0a50CE1AdF83d0b5D99FBE0e3fEB43fb
      chainID: 80002
      publishingKey: pbkey
`)

	reader := NewMyYAMLReader(yamlData)
	return reader
}

// CreateFileWithRHSAll - Helper for testing
func CreateFileWithRHSAll(t *testing.T) *MyYAMLReader {
	t.Helper()
	yamlData := []byte(`polygon:
  amoy:
    contractAddress: 0x1a4cC30f2aA0377b0c3bc9848766D90cb4404124
    networkURL: https://polygon-amoy.g.alchemy.com/v2/123
    defaultGasLimit: 600000
    confirmationTimeout: 10s
    confirmationBlockCount: 5
    receiptTimeout: 600s
    minGasPrice: 0
    maxGasPrice: 1000000
    rpcResponseTimeout: 5s
    waitReceiptCycleTime: 30s
    waitBlockCycleTime: 30s
    gasLess: false
    transferAmountWei: 1000000000000000000
    rhsSettings:
      mode: All
      rhsUrl: https://rhs-staging.polygonid.me
      contractAddress: 0x3d3763eC0a50CE1AdF83d0b5D99FBE0e3fEB43fb
      chainID: 80002
      publishingKey: pbkey
`)

	reader := NewMyYAMLReader(yamlData)
	return reader
}
