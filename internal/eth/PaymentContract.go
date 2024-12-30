// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package eth

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// MCPaymentIden3PaymentRailsERC20RequestV1 is an auto generated low-level Go binding around an user-defined struct.
type MCPaymentIden3PaymentRailsERC20RequestV1 struct {
	TokenAddress   common.Address
	Recipient      common.Address
	Amount         *big.Int
	ExpirationDate *big.Int
	Nonce          *big.Int
	Metadata       []byte
}

// MCPaymentIden3PaymentRailsRequestV1 is an auto generated low-level Go binding around an user-defined struct.
type MCPaymentIden3PaymentRailsRequestV1 struct {
	Recipient      common.Address
	Amount         *big.Int
	ExpirationDate *big.Int
	Nonce          *big.Int
	Metadata       []byte
}

// PaymentContractMetaData contains all meta data concerning the PaymentContract contract.
var PaymentContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"message\",\"type\":\"string\"}],\"name\":\"ECDSAInvalidSignatureLength\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"message\",\"type\":\"string\"}],\"name\":\"InvalidOwnerPercentage\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"message\",\"type\":\"string\"}],\"name\":\"InvalidSignature\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"message\",\"type\":\"string\"}],\"name\":\"PaymentError\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"message\",\"type\":\"string\"}],\"name\":\"WithdrawError\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EIP712DomainChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferStarted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"Payment\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"ERC_20_PAYMENT_DATA_TYPE_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PAYMENT_DATA_TYPE_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VERSION\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"acceptOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"eip712Domain\",\"outputs\":[{\"internalType\":\"bytes1\",\"name\":\"fields\",\"type\":\"bytes1\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"verifyingContract\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"salt\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"extensions\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"}],\"name\":\"getBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwnerBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwnerPercentage\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint8\",\"name\":\"ownerPercentage\",\"type\":\"uint8\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"isPaymentDone\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"issuerWithdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ownerWithdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"expirationDate\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"metadata\",\"type\":\"bytes\"}],\"internalType\":\"structMCPayment.Iden3PaymentRailsRequestV1\",\"name\":\"paymentData\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"pay\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"expirationDate\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"metadata\",\"type\":\"bytes\"}],\"internalType\":\"structMCPayment.Iden3PaymentRailsERC20RequestV1\",\"name\":\"paymentData\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"payERC20\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"permitSignature\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"expirationDate\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"metadata\",\"type\":\"bytes\"}],\"internalType\":\"structMCPayment.Iden3PaymentRailsERC20RequestV1\",\"name\":\"paymentData\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"payERC20Permit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"ownerPercentage\",\"type\":\"uint8\"}],\"name\":\"updateOwnerPercentage\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"expirationDate\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"metadata\",\"type\":\"bytes\"}],\"internalType\":\"structMCPayment.Iden3PaymentRailsERC20RequestV1\",\"name\":\"paymentData\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"verifyIden3PaymentRailsERC20RequestV1Signature\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"expirationDate\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"metadata\",\"type\":\"bytes\"}],\"internalType\":\"structMCPayment.Iden3PaymentRailsRequestV1\",\"name\":\"paymentData\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"verifyIden3PaymentRailsRequestV1Signature\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// PaymentContractABI is the input ABI used to generate the binding from.
// Deprecated: Use PaymentContractMetaData.ABI instead.
var PaymentContractABI = PaymentContractMetaData.ABI

// PaymentContract is an auto generated Go binding around an Ethereum contract.
type PaymentContract struct {
	PaymentContractCaller     // Read-only binding to the contract
	PaymentContractTransactor // Write-only binding to the contract
	PaymentContractFilterer   // Log filterer for contract events
}

// PaymentContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type PaymentContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PaymentContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PaymentContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PaymentContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PaymentContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PaymentContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PaymentContractSession struct {
	Contract     *PaymentContract  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PaymentContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PaymentContractCallerSession struct {
	Contract *PaymentContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// PaymentContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PaymentContractTransactorSession struct {
	Contract     *PaymentContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// PaymentContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type PaymentContractRaw struct {
	Contract *PaymentContract // Generic contract binding to access the raw methods on
}

// PaymentContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PaymentContractCallerRaw struct {
	Contract *PaymentContractCaller // Generic read-only contract binding to access the raw methods on
}

// PaymentContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PaymentContractTransactorRaw struct {
	Contract *PaymentContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPaymentContract creates a new instance of PaymentContract, bound to a specific deployed contract.
func NewPaymentContract(address common.Address, backend bind.ContractBackend) (*PaymentContract, error) {
	contract, err := bindPaymentContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PaymentContract{PaymentContractCaller: PaymentContractCaller{contract: contract}, PaymentContractTransactor: PaymentContractTransactor{contract: contract}, PaymentContractFilterer: PaymentContractFilterer{contract: contract}}, nil
}

// NewPaymentContractCaller creates a new read-only instance of PaymentContract, bound to a specific deployed contract.
func NewPaymentContractCaller(address common.Address, caller bind.ContractCaller) (*PaymentContractCaller, error) {
	contract, err := bindPaymentContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PaymentContractCaller{contract: contract}, nil
}

// NewPaymentContractTransactor creates a new write-only instance of PaymentContract, bound to a specific deployed contract.
func NewPaymentContractTransactor(address common.Address, transactor bind.ContractTransactor) (*PaymentContractTransactor, error) {
	contract, err := bindPaymentContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PaymentContractTransactor{contract: contract}, nil
}

// NewPaymentContractFilterer creates a new log filterer instance of PaymentContract, bound to a specific deployed contract.
func NewPaymentContractFilterer(address common.Address, filterer bind.ContractFilterer) (*PaymentContractFilterer, error) {
	contract, err := bindPaymentContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PaymentContractFilterer{contract: contract}, nil
}

// bindPaymentContract binds a generic wrapper to an already deployed contract.
func bindPaymentContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PaymentContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PaymentContract *PaymentContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PaymentContract.Contract.PaymentContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PaymentContract *PaymentContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PaymentContract.Contract.PaymentContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PaymentContract *PaymentContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PaymentContract.Contract.PaymentContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PaymentContract *PaymentContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PaymentContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PaymentContract *PaymentContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PaymentContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PaymentContract *PaymentContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PaymentContract.Contract.contract.Transact(opts, method, params...)
}

// ERC20PAYMENTDATATYPEHASH is a free data retrieval call binding the contract method 0xc6bfaa3f.
//
// Solidity: function ERC_20_PAYMENT_DATA_TYPE_HASH() view returns(bytes32)
func (_PaymentContract *PaymentContractCaller) ERC20PAYMENTDATATYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _PaymentContract.contract.Call(opts, &out, "ERC_20_PAYMENT_DATA_TYPE_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ERC20PAYMENTDATATYPEHASH is a free data retrieval call binding the contract method 0xc6bfaa3f.
//
// Solidity: function ERC_20_PAYMENT_DATA_TYPE_HASH() view returns(bytes32)
func (_PaymentContract *PaymentContractSession) ERC20PAYMENTDATATYPEHASH() ([32]byte, error) {
	return _PaymentContract.Contract.ERC20PAYMENTDATATYPEHASH(&_PaymentContract.CallOpts)
}

// ERC20PAYMENTDATATYPEHASH is a free data retrieval call binding the contract method 0xc6bfaa3f.
//
// Solidity: function ERC_20_PAYMENT_DATA_TYPE_HASH() view returns(bytes32)
func (_PaymentContract *PaymentContractCallerSession) ERC20PAYMENTDATATYPEHASH() ([32]byte, error) {
	return _PaymentContract.Contract.ERC20PAYMENTDATATYPEHASH(&_PaymentContract.CallOpts)
}

// PAYMENTDATATYPEHASH is a free data retrieval call binding the contract method 0xf0dd6899.
//
// Solidity: function PAYMENT_DATA_TYPE_HASH() view returns(bytes32)
func (_PaymentContract *PaymentContractCaller) PAYMENTDATATYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _PaymentContract.contract.Call(opts, &out, "PAYMENT_DATA_TYPE_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PAYMENTDATATYPEHASH is a free data retrieval call binding the contract method 0xf0dd6899.
//
// Solidity: function PAYMENT_DATA_TYPE_HASH() view returns(bytes32)
func (_PaymentContract *PaymentContractSession) PAYMENTDATATYPEHASH() ([32]byte, error) {
	return _PaymentContract.Contract.PAYMENTDATATYPEHASH(&_PaymentContract.CallOpts)
}

// PAYMENTDATATYPEHASH is a free data retrieval call binding the contract method 0xf0dd6899.
//
// Solidity: function PAYMENT_DATA_TYPE_HASH() view returns(bytes32)
func (_PaymentContract *PaymentContractCallerSession) PAYMENTDATATYPEHASH() ([32]byte, error) {
	return _PaymentContract.Contract.PAYMENTDATATYPEHASH(&_PaymentContract.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_PaymentContract *PaymentContractCaller) VERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _PaymentContract.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_PaymentContract *PaymentContractSession) VERSION() (string, error) {
	return _PaymentContract.Contract.VERSION(&_PaymentContract.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_PaymentContract *PaymentContractCallerSession) VERSION() (string, error) {
	return _PaymentContract.Contract.VERSION(&_PaymentContract.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_PaymentContract *PaymentContractCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _PaymentContract.contract.Call(opts, &out, "eip712Domain")

	outstruct := new(struct {
		Fields            [1]byte
		Name              string
		Version           string
		ChainId           *big.Int
		VerifyingContract common.Address
		Salt              [32]byte
		Extensions        []*big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Fields = *abi.ConvertType(out[0], new([1]byte)).(*[1]byte)
	outstruct.Name = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Version = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.ChainId = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.VerifyingContract = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.Salt = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.Extensions = *abi.ConvertType(out[6], new([]*big.Int)).(*[]*big.Int)

	return *outstruct, err

}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_PaymentContract *PaymentContractSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _PaymentContract.Contract.Eip712Domain(&_PaymentContract.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_PaymentContract *PaymentContractCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _PaymentContract.Contract.Eip712Domain(&_PaymentContract.CallOpts)
}

// GetBalance is a free data retrieval call binding the contract method 0xf8b2cb4f.
//
// Solidity: function getBalance(address recipient) view returns(uint256)
func (_PaymentContract *PaymentContractCaller) GetBalance(opts *bind.CallOpts, recipient common.Address) (*big.Int, error) {
	var out []interface{}
	err := _PaymentContract.contract.Call(opts, &out, "getBalance", recipient)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetBalance is a free data retrieval call binding the contract method 0xf8b2cb4f.
//
// Solidity: function getBalance(address recipient) view returns(uint256)
func (_PaymentContract *PaymentContractSession) GetBalance(recipient common.Address) (*big.Int, error) {
	return _PaymentContract.Contract.GetBalance(&_PaymentContract.CallOpts, recipient)
}

// GetBalance is a free data retrieval call binding the contract method 0xf8b2cb4f.
//
// Solidity: function getBalance(address recipient) view returns(uint256)
func (_PaymentContract *PaymentContractCallerSession) GetBalance(recipient common.Address) (*big.Int, error) {
	return _PaymentContract.Contract.GetBalance(&_PaymentContract.CallOpts, recipient)
}

// GetOwnerBalance is a free data retrieval call binding the contract method 0x590791f2.
//
// Solidity: function getOwnerBalance() view returns(uint256)
func (_PaymentContract *PaymentContractCaller) GetOwnerBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PaymentContract.contract.Call(opts, &out, "getOwnerBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOwnerBalance is a free data retrieval call binding the contract method 0x590791f2.
//
// Solidity: function getOwnerBalance() view returns(uint256)
func (_PaymentContract *PaymentContractSession) GetOwnerBalance() (*big.Int, error) {
	return _PaymentContract.Contract.GetOwnerBalance(&_PaymentContract.CallOpts)
}

// GetOwnerBalance is a free data retrieval call binding the contract method 0x590791f2.
//
// Solidity: function getOwnerBalance() view returns(uint256)
func (_PaymentContract *PaymentContractCallerSession) GetOwnerBalance() (*big.Int, error) {
	return _PaymentContract.Contract.GetOwnerBalance(&_PaymentContract.CallOpts)
}

// GetOwnerPercentage is a free data retrieval call binding the contract method 0x309a042c.
//
// Solidity: function getOwnerPercentage() view returns(uint8)
func (_PaymentContract *PaymentContractCaller) GetOwnerPercentage(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _PaymentContract.contract.Call(opts, &out, "getOwnerPercentage")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetOwnerPercentage is a free data retrieval call binding the contract method 0x309a042c.
//
// Solidity: function getOwnerPercentage() view returns(uint8)
func (_PaymentContract *PaymentContractSession) GetOwnerPercentage() (uint8, error) {
	return _PaymentContract.Contract.GetOwnerPercentage(&_PaymentContract.CallOpts)
}

// GetOwnerPercentage is a free data retrieval call binding the contract method 0x309a042c.
//
// Solidity: function getOwnerPercentage() view returns(uint8)
func (_PaymentContract *PaymentContractCallerSession) GetOwnerPercentage() (uint8, error) {
	return _PaymentContract.Contract.GetOwnerPercentage(&_PaymentContract.CallOpts)
}

// IsPaymentDone is a free data retrieval call binding the contract method 0x9d9c12b7.
//
// Solidity: function isPaymentDone(address recipient, uint256 nonce) view returns(bool)
func (_PaymentContract *PaymentContractCaller) IsPaymentDone(opts *bind.CallOpts, recipient common.Address, nonce *big.Int) (bool, error) {
	var out []interface{}
	err := _PaymentContract.contract.Call(opts, &out, "isPaymentDone", recipient, nonce)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsPaymentDone is a free data retrieval call binding the contract method 0x9d9c12b7.
//
// Solidity: function isPaymentDone(address recipient, uint256 nonce) view returns(bool)
func (_PaymentContract *PaymentContractSession) IsPaymentDone(recipient common.Address, nonce *big.Int) (bool, error) {
	return _PaymentContract.Contract.IsPaymentDone(&_PaymentContract.CallOpts, recipient, nonce)
}

// IsPaymentDone is a free data retrieval call binding the contract method 0x9d9c12b7.
//
// Solidity: function isPaymentDone(address recipient, uint256 nonce) view returns(bool)
func (_PaymentContract *PaymentContractCallerSession) IsPaymentDone(recipient common.Address, nonce *big.Int) (bool, error) {
	return _PaymentContract.Contract.IsPaymentDone(&_PaymentContract.CallOpts, recipient, nonce)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PaymentContract *PaymentContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PaymentContract.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PaymentContract *PaymentContractSession) Owner() (common.Address, error) {
	return _PaymentContract.Contract.Owner(&_PaymentContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PaymentContract *PaymentContractCallerSession) Owner() (common.Address, error) {
	return _PaymentContract.Contract.Owner(&_PaymentContract.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_PaymentContract *PaymentContractCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PaymentContract.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_PaymentContract *PaymentContractSession) PendingOwner() (common.Address, error) {
	return _PaymentContract.Contract.PendingOwner(&_PaymentContract.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_PaymentContract *PaymentContractCallerSession) PendingOwner() (common.Address, error) {
	return _PaymentContract.Contract.PendingOwner(&_PaymentContract.CallOpts)
}

// VerifyIden3PaymentRailsERC20RequestV1Signature is a free data retrieval call binding the contract method 0x3039009d.
//
// Solidity: function verifyIden3PaymentRailsERC20RequestV1Signature((address,address,uint256,uint256,uint256,bytes) paymentData, bytes signature) view returns()
func (_PaymentContract *PaymentContractCaller) VerifyIden3PaymentRailsERC20RequestV1Signature(opts *bind.CallOpts, paymentData MCPaymentIden3PaymentRailsERC20RequestV1, signature []byte) error {
	var out []interface{}
	err := _PaymentContract.contract.Call(opts, &out, "verifyIden3PaymentRailsERC20RequestV1Signature", paymentData, signature)

	if err != nil {
		return err
	}

	return err

}

// VerifyIden3PaymentRailsERC20RequestV1Signature is a free data retrieval call binding the contract method 0x3039009d.
//
// Solidity: function verifyIden3PaymentRailsERC20RequestV1Signature((address,address,uint256,uint256,uint256,bytes) paymentData, bytes signature) view returns()
func (_PaymentContract *PaymentContractSession) VerifyIden3PaymentRailsERC20RequestV1Signature(paymentData MCPaymentIden3PaymentRailsERC20RequestV1, signature []byte) error {
	return _PaymentContract.Contract.VerifyIden3PaymentRailsERC20RequestV1Signature(&_PaymentContract.CallOpts, paymentData, signature)
}

// VerifyIden3PaymentRailsERC20RequestV1Signature is a free data retrieval call binding the contract method 0x3039009d.
//
// Solidity: function verifyIden3PaymentRailsERC20RequestV1Signature((address,address,uint256,uint256,uint256,bytes) paymentData, bytes signature) view returns()
func (_PaymentContract *PaymentContractCallerSession) VerifyIden3PaymentRailsERC20RequestV1Signature(paymentData MCPaymentIden3PaymentRailsERC20RequestV1, signature []byte) error {
	return _PaymentContract.Contract.VerifyIden3PaymentRailsERC20RequestV1Signature(&_PaymentContract.CallOpts, paymentData, signature)
}

// VerifyIden3PaymentRailsRequestV1Signature is a free data retrieval call binding the contract method 0x955317f6.
//
// Solidity: function verifyIden3PaymentRailsRequestV1Signature((address,uint256,uint256,uint256,bytes) paymentData, bytes signature) view returns()
func (_PaymentContract *PaymentContractCaller) VerifyIden3PaymentRailsRequestV1Signature(opts *bind.CallOpts, paymentData MCPaymentIden3PaymentRailsRequestV1, signature []byte) error {
	var out []interface{}
	err := _PaymentContract.contract.Call(opts, &out, "verifyIden3PaymentRailsRequestV1Signature", paymentData, signature)

	if err != nil {
		return err
	}

	return err

}

// VerifyIden3PaymentRailsRequestV1Signature is a free data retrieval call binding the contract method 0x955317f6.
//
// Solidity: function verifyIden3PaymentRailsRequestV1Signature((address,uint256,uint256,uint256,bytes) paymentData, bytes signature) view returns()
func (_PaymentContract *PaymentContractSession) VerifyIden3PaymentRailsRequestV1Signature(paymentData MCPaymentIden3PaymentRailsRequestV1, signature []byte) error {
	return _PaymentContract.Contract.VerifyIden3PaymentRailsRequestV1Signature(&_PaymentContract.CallOpts, paymentData, signature)
}

// VerifyIden3PaymentRailsRequestV1Signature is a free data retrieval call binding the contract method 0x955317f6.
//
// Solidity: function verifyIden3PaymentRailsRequestV1Signature((address,uint256,uint256,uint256,bytes) paymentData, bytes signature) view returns()
func (_PaymentContract *PaymentContractCallerSession) VerifyIden3PaymentRailsRequestV1Signature(paymentData MCPaymentIden3PaymentRailsRequestV1, signature []byte) error {
	return _PaymentContract.Contract.VerifyIden3PaymentRailsRequestV1Signature(&_PaymentContract.CallOpts, paymentData, signature)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_PaymentContract *PaymentContractTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PaymentContract.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_PaymentContract *PaymentContractSession) AcceptOwnership() (*types.Transaction, error) {
	return _PaymentContract.Contract.AcceptOwnership(&_PaymentContract.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_PaymentContract *PaymentContractTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _PaymentContract.Contract.AcceptOwnership(&_PaymentContract.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x943b24b2.
//
// Solidity: function initialize(address owner, uint8 ownerPercentage) returns()
func (_PaymentContract *PaymentContractTransactor) Initialize(opts *bind.TransactOpts, owner common.Address, ownerPercentage uint8) (*types.Transaction, error) {
	return _PaymentContract.contract.Transact(opts, "initialize", owner, ownerPercentage)
}

// Initialize is a paid mutator transaction binding the contract method 0x943b24b2.
//
// Solidity: function initialize(address owner, uint8 ownerPercentage) returns()
func (_PaymentContract *PaymentContractSession) Initialize(owner common.Address, ownerPercentage uint8) (*types.Transaction, error) {
	return _PaymentContract.Contract.Initialize(&_PaymentContract.TransactOpts, owner, ownerPercentage)
}

// Initialize is a paid mutator transaction binding the contract method 0x943b24b2.
//
// Solidity: function initialize(address owner, uint8 ownerPercentage) returns()
func (_PaymentContract *PaymentContractTransactorSession) Initialize(owner common.Address, ownerPercentage uint8) (*types.Transaction, error) {
	return _PaymentContract.Contract.Initialize(&_PaymentContract.TransactOpts, owner, ownerPercentage)
}

// IssuerWithdraw is a paid mutator transaction binding the contract method 0xcc9cd961.
//
// Solidity: function issuerWithdraw() returns()
func (_PaymentContract *PaymentContractTransactor) IssuerWithdraw(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PaymentContract.contract.Transact(opts, "issuerWithdraw")
}

// IssuerWithdraw is a paid mutator transaction binding the contract method 0xcc9cd961.
//
// Solidity: function issuerWithdraw() returns()
func (_PaymentContract *PaymentContractSession) IssuerWithdraw() (*types.Transaction, error) {
	return _PaymentContract.Contract.IssuerWithdraw(&_PaymentContract.TransactOpts)
}

// IssuerWithdraw is a paid mutator transaction binding the contract method 0xcc9cd961.
//
// Solidity: function issuerWithdraw() returns()
func (_PaymentContract *PaymentContractTransactorSession) IssuerWithdraw() (*types.Transaction, error) {
	return _PaymentContract.Contract.IssuerWithdraw(&_PaymentContract.TransactOpts)
}

// OwnerWithdraw is a paid mutator transaction binding the contract method 0x4311de8f.
//
// Solidity: function ownerWithdraw() returns()
func (_PaymentContract *PaymentContractTransactor) OwnerWithdraw(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PaymentContract.contract.Transact(opts, "ownerWithdraw")
}

// OwnerWithdraw is a paid mutator transaction binding the contract method 0x4311de8f.
//
// Solidity: function ownerWithdraw() returns()
func (_PaymentContract *PaymentContractSession) OwnerWithdraw() (*types.Transaction, error) {
	return _PaymentContract.Contract.OwnerWithdraw(&_PaymentContract.TransactOpts)
}

// OwnerWithdraw is a paid mutator transaction binding the contract method 0x4311de8f.
//
// Solidity: function ownerWithdraw() returns()
func (_PaymentContract *PaymentContractTransactorSession) OwnerWithdraw() (*types.Transaction, error) {
	return _PaymentContract.Contract.OwnerWithdraw(&_PaymentContract.TransactOpts)
}

// Pay is a paid mutator transaction binding the contract method 0xaa021669.
//
// Solidity: function pay((address,uint256,uint256,uint256,bytes) paymentData, bytes signature) payable returns()
func (_PaymentContract *PaymentContractTransactor) Pay(opts *bind.TransactOpts, paymentData MCPaymentIden3PaymentRailsRequestV1, signature []byte) (*types.Transaction, error) {
	return _PaymentContract.contract.Transact(opts, "pay", paymentData, signature)
}

// Pay is a paid mutator transaction binding the contract method 0xaa021669.
//
// Solidity: function pay((address,uint256,uint256,uint256,bytes) paymentData, bytes signature) payable returns()
func (_PaymentContract *PaymentContractSession) Pay(paymentData MCPaymentIden3PaymentRailsRequestV1, signature []byte) (*types.Transaction, error) {
	return _PaymentContract.Contract.Pay(&_PaymentContract.TransactOpts, paymentData, signature)
}

// Pay is a paid mutator transaction binding the contract method 0xaa021669.
//
// Solidity: function pay((address,uint256,uint256,uint256,bytes) paymentData, bytes signature) payable returns()
func (_PaymentContract *PaymentContractTransactorSession) Pay(paymentData MCPaymentIden3PaymentRailsRequestV1, signature []byte) (*types.Transaction, error) {
	return _PaymentContract.Contract.Pay(&_PaymentContract.TransactOpts, paymentData, signature)
}

// PayERC20 is a paid mutator transaction binding the contract method 0x57615a3a.
//
// Solidity: function payERC20((address,address,uint256,uint256,uint256,bytes) paymentData, bytes signature) returns()
func (_PaymentContract *PaymentContractTransactor) PayERC20(opts *bind.TransactOpts, paymentData MCPaymentIden3PaymentRailsERC20RequestV1, signature []byte) (*types.Transaction, error) {
	return _PaymentContract.contract.Transact(opts, "payERC20", paymentData, signature)
}

// PayERC20 is a paid mutator transaction binding the contract method 0x57615a3a.
//
// Solidity: function payERC20((address,address,uint256,uint256,uint256,bytes) paymentData, bytes signature) returns()
func (_PaymentContract *PaymentContractSession) PayERC20(paymentData MCPaymentIden3PaymentRailsERC20RequestV1, signature []byte) (*types.Transaction, error) {
	return _PaymentContract.Contract.PayERC20(&_PaymentContract.TransactOpts, paymentData, signature)
}

// PayERC20 is a paid mutator transaction binding the contract method 0x57615a3a.
//
// Solidity: function payERC20((address,address,uint256,uint256,uint256,bytes) paymentData, bytes signature) returns()
func (_PaymentContract *PaymentContractTransactorSession) PayERC20(paymentData MCPaymentIden3PaymentRailsERC20RequestV1, signature []byte) (*types.Transaction, error) {
	return _PaymentContract.Contract.PayERC20(&_PaymentContract.TransactOpts, paymentData, signature)
}

// PayERC20Permit is a paid mutator transaction binding the contract method 0x3dbbdbe5.
//
// Solidity: function payERC20Permit(bytes permitSignature, (address,address,uint256,uint256,uint256,bytes) paymentData, bytes signature) returns()
func (_PaymentContract *PaymentContractTransactor) PayERC20Permit(opts *bind.TransactOpts, permitSignature []byte, paymentData MCPaymentIden3PaymentRailsERC20RequestV1, signature []byte) (*types.Transaction, error) {
	return _PaymentContract.contract.Transact(opts, "payERC20Permit", permitSignature, paymentData, signature)
}

// PayERC20Permit is a paid mutator transaction binding the contract method 0x3dbbdbe5.
//
// Solidity: function payERC20Permit(bytes permitSignature, (address,address,uint256,uint256,uint256,bytes) paymentData, bytes signature) returns()
func (_PaymentContract *PaymentContractSession) PayERC20Permit(permitSignature []byte, paymentData MCPaymentIden3PaymentRailsERC20RequestV1, signature []byte) (*types.Transaction, error) {
	return _PaymentContract.Contract.PayERC20Permit(&_PaymentContract.TransactOpts, permitSignature, paymentData, signature)
}

// PayERC20Permit is a paid mutator transaction binding the contract method 0x3dbbdbe5.
//
// Solidity: function payERC20Permit(bytes permitSignature, (address,address,uint256,uint256,uint256,bytes) paymentData, bytes signature) returns()
func (_PaymentContract *PaymentContractTransactorSession) PayERC20Permit(permitSignature []byte, paymentData MCPaymentIden3PaymentRailsERC20RequestV1, signature []byte) (*types.Transaction, error) {
	return _PaymentContract.Contract.PayERC20Permit(&_PaymentContract.TransactOpts, permitSignature, paymentData, signature)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_PaymentContract *PaymentContractTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PaymentContract.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_PaymentContract *PaymentContractSession) RenounceOwnership() (*types.Transaction, error) {
	return _PaymentContract.Contract.RenounceOwnership(&_PaymentContract.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_PaymentContract *PaymentContractTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _PaymentContract.Contract.RenounceOwnership(&_PaymentContract.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_PaymentContract *PaymentContractTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _PaymentContract.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_PaymentContract *PaymentContractSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _PaymentContract.Contract.TransferOwnership(&_PaymentContract.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_PaymentContract *PaymentContractTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _PaymentContract.Contract.TransferOwnership(&_PaymentContract.TransactOpts, newOwner)
}

// UpdateOwnerPercentage is a paid mutator transaction binding the contract method 0x0cea58aa.
//
// Solidity: function updateOwnerPercentage(uint8 ownerPercentage) returns()
func (_PaymentContract *PaymentContractTransactor) UpdateOwnerPercentage(opts *bind.TransactOpts, ownerPercentage uint8) (*types.Transaction, error) {
	return _PaymentContract.contract.Transact(opts, "updateOwnerPercentage", ownerPercentage)
}

// UpdateOwnerPercentage is a paid mutator transaction binding the contract method 0x0cea58aa.
//
// Solidity: function updateOwnerPercentage(uint8 ownerPercentage) returns()
func (_PaymentContract *PaymentContractSession) UpdateOwnerPercentage(ownerPercentage uint8) (*types.Transaction, error) {
	return _PaymentContract.Contract.UpdateOwnerPercentage(&_PaymentContract.TransactOpts, ownerPercentage)
}

// UpdateOwnerPercentage is a paid mutator transaction binding the contract method 0x0cea58aa.
//
// Solidity: function updateOwnerPercentage(uint8 ownerPercentage) returns()
func (_PaymentContract *PaymentContractTransactorSession) UpdateOwnerPercentage(ownerPercentage uint8) (*types.Transaction, error) {
	return _PaymentContract.Contract.UpdateOwnerPercentage(&_PaymentContract.TransactOpts, ownerPercentage)
}

// PaymentContractEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the PaymentContract contract.
type PaymentContractEIP712DomainChangedIterator struct {
	Event *PaymentContractEIP712DomainChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PaymentContractEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PaymentContractEIP712DomainChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PaymentContractEIP712DomainChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PaymentContractEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PaymentContractEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PaymentContractEIP712DomainChanged represents a EIP712DomainChanged event raised by the PaymentContract contract.
type PaymentContractEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_PaymentContract *PaymentContractFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*PaymentContractEIP712DomainChangedIterator, error) {

	logs, sub, err := _PaymentContract.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &PaymentContractEIP712DomainChangedIterator{contract: _PaymentContract.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_PaymentContract *PaymentContractFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *PaymentContractEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _PaymentContract.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PaymentContractEIP712DomainChanged)
				if err := _PaymentContract.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEIP712DomainChanged is a log parse operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_PaymentContract *PaymentContractFilterer) ParseEIP712DomainChanged(log types.Log) (*PaymentContractEIP712DomainChanged, error) {
	event := new(PaymentContractEIP712DomainChanged)
	if err := _PaymentContract.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PaymentContractInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the PaymentContract contract.
type PaymentContractInitializedIterator struct {
	Event *PaymentContractInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PaymentContractInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PaymentContractInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PaymentContractInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PaymentContractInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PaymentContractInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PaymentContractInitialized represents a Initialized event raised by the PaymentContract contract.
type PaymentContractInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_PaymentContract *PaymentContractFilterer) FilterInitialized(opts *bind.FilterOpts) (*PaymentContractInitializedIterator, error) {

	logs, sub, err := _PaymentContract.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &PaymentContractInitializedIterator{contract: _PaymentContract.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_PaymentContract *PaymentContractFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *PaymentContractInitialized) (event.Subscription, error) {

	logs, sub, err := _PaymentContract.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PaymentContractInitialized)
				if err := _PaymentContract.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_PaymentContract *PaymentContractFilterer) ParseInitialized(log types.Log) (*PaymentContractInitialized, error) {
	event := new(PaymentContractInitialized)
	if err := _PaymentContract.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PaymentContractOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the PaymentContract contract.
type PaymentContractOwnershipTransferStartedIterator struct {
	Event *PaymentContractOwnershipTransferStarted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PaymentContractOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PaymentContractOwnershipTransferStarted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PaymentContractOwnershipTransferStarted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PaymentContractOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PaymentContractOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PaymentContractOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the PaymentContract contract.
type PaymentContractOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_PaymentContract *PaymentContractFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*PaymentContractOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _PaymentContract.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &PaymentContractOwnershipTransferStartedIterator{contract: _PaymentContract.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_PaymentContract *PaymentContractFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *PaymentContractOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _PaymentContract.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PaymentContractOwnershipTransferStarted)
				if err := _PaymentContract.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferStarted is a log parse operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_PaymentContract *PaymentContractFilterer) ParseOwnershipTransferStarted(log types.Log) (*PaymentContractOwnershipTransferStarted, error) {
	event := new(PaymentContractOwnershipTransferStarted)
	if err := _PaymentContract.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PaymentContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the PaymentContract contract.
type PaymentContractOwnershipTransferredIterator struct {
	Event *PaymentContractOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PaymentContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PaymentContractOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PaymentContractOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PaymentContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PaymentContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PaymentContractOwnershipTransferred represents a OwnershipTransferred event raised by the PaymentContract contract.
type PaymentContractOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_PaymentContract *PaymentContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*PaymentContractOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _PaymentContract.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &PaymentContractOwnershipTransferredIterator{contract: _PaymentContract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_PaymentContract *PaymentContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *PaymentContractOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _PaymentContract.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PaymentContractOwnershipTransferred)
				if err := _PaymentContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_PaymentContract *PaymentContractFilterer) ParseOwnershipTransferred(log types.Log) (*PaymentContractOwnershipTransferred, error) {
	event := new(PaymentContractOwnershipTransferred)
	if err := _PaymentContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PaymentContractPaymentIterator is returned from FilterPayment and is used to iterate over the raw logs and unpacked data for Payment events raised by the PaymentContract contract.
type PaymentContractPaymentIterator struct {
	Event *PaymentContractPayment // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PaymentContractPaymentIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PaymentContractPayment)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PaymentContractPayment)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PaymentContractPaymentIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PaymentContractPaymentIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PaymentContractPayment represents a Payment event raised by the PaymentContract contract.
type PaymentContractPayment struct {
	Recipient common.Address
	Nonce     *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterPayment is a free log retrieval operation binding the contract event 0xd4f43975feb89f48dd30cabbb32011045be187d1e11c8ea9faa43efc35282519.
//
// Solidity: event Payment(address indexed recipient, uint256 indexed nonce)
func (_PaymentContract *PaymentContractFilterer) FilterPayment(opts *bind.FilterOpts, recipient []common.Address, nonce []*big.Int) (*PaymentContractPaymentIterator, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}
	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}

	logs, sub, err := _PaymentContract.contract.FilterLogs(opts, "Payment", recipientRule, nonceRule)
	if err != nil {
		return nil, err
	}
	return &PaymentContractPaymentIterator{contract: _PaymentContract.contract, event: "Payment", logs: logs, sub: sub}, nil
}

// WatchPayment is a free log subscription operation binding the contract event 0xd4f43975feb89f48dd30cabbb32011045be187d1e11c8ea9faa43efc35282519.
//
// Solidity: event Payment(address indexed recipient, uint256 indexed nonce)
func (_PaymentContract *PaymentContractFilterer) WatchPayment(opts *bind.WatchOpts, sink chan<- *PaymentContractPayment, recipient []common.Address, nonce []*big.Int) (event.Subscription, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}
	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}

	logs, sub, err := _PaymentContract.contract.WatchLogs(opts, "Payment", recipientRule, nonceRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PaymentContractPayment)
				if err := _PaymentContract.contract.UnpackLog(event, "Payment", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePayment is a log parse operation binding the contract event 0xd4f43975feb89f48dd30cabbb32011045be187d1e11c8ea9faa43efc35282519.
//
// Solidity: event Payment(address indexed recipient, uint256 indexed nonce)
func (_PaymentContract *PaymentContractFilterer) ParsePayment(log types.Log) (*PaymentContractPayment, error) {
	event := new(PaymentContractPayment)
	if err := _PaymentContract.contract.UnpackLog(event, "Payment", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
