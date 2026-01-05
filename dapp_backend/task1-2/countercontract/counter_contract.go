// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package countercontract

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

// CountercontractMetaData contains all meta data concerning the Countercontract contract.
var CountercontractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"count\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"increment\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// CountercontractABI is the input ABI used to generate the binding from.
// Deprecated: Use CountercontractMetaData.ABI instead.
var CountercontractABI = CountercontractMetaData.ABI

// Countercontract is an auto generated Go binding around an Ethereum contract.
type Countercontract struct {
	CountercontractCaller     // Read-only binding to the contract
	CountercontractTransactor // Write-only binding to the contract
	CountercontractFilterer   // Log filterer for contract events
}

// CountercontractCaller is an auto generated read-only Go binding around an Ethereum contract.
type CountercontractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CountercontractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type CountercontractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CountercontractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type CountercontractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CountercontractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type CountercontractSession struct {
	Contract     *Countercontract  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// CountercontractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type CountercontractCallerSession struct {
	Contract *CountercontractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// CountercontractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type CountercontractTransactorSession struct {
	Contract     *CountercontractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// CountercontractRaw is an auto generated low-level Go binding around an Ethereum contract.
type CountercontractRaw struct {
	Contract *Countercontract // Generic contract binding to access the raw methods on
}

// CountercontractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type CountercontractCallerRaw struct {
	Contract *CountercontractCaller // Generic read-only contract binding to access the raw methods on
}

// CountercontractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type CountercontractTransactorRaw struct {
	Contract *CountercontractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCountercontract creates a new instance of Countercontract, bound to a specific deployed contract.
func NewCountercontract(address common.Address, backend bind.ContractBackend) (*Countercontract, error) {
	contract, err := bindCountercontract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Countercontract{CountercontractCaller: CountercontractCaller{contract: contract}, CountercontractTransactor: CountercontractTransactor{contract: contract}, CountercontractFilterer: CountercontractFilterer{contract: contract}}, nil
}

// NewCountercontractCaller creates a new read-only instance of Countercontract, bound to a specific deployed contract.
func NewCountercontractCaller(address common.Address, caller bind.ContractCaller) (*CountercontractCaller, error) {
	contract, err := bindCountercontract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CountercontractCaller{contract: contract}, nil
}

// NewCountercontractTransactor creates a new write-only instance of Countercontract, bound to a specific deployed contract.
func NewCountercontractTransactor(address common.Address, transactor bind.ContractTransactor) (*CountercontractTransactor, error) {
	contract, err := bindCountercontract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CountercontractTransactor{contract: contract}, nil
}

// NewCountercontractFilterer creates a new log filterer instance of Countercontract, bound to a specific deployed contract.
func NewCountercontractFilterer(address common.Address, filterer bind.ContractFilterer) (*CountercontractFilterer, error) {
	contract, err := bindCountercontract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CountercontractFilterer{contract: contract}, nil
}

// bindCountercontract binds a generic wrapper to an already deployed contract.
func bindCountercontract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := CountercontractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Countercontract *CountercontractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Countercontract.Contract.CountercontractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Countercontract *CountercontractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Countercontract.Contract.CountercontractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Countercontract *CountercontractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Countercontract.Contract.CountercontractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Countercontract *CountercontractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Countercontract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Countercontract *CountercontractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Countercontract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Countercontract *CountercontractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Countercontract.Contract.contract.Transact(opts, method, params...)
}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() view returns(uint256)
func (_Countercontract *CountercontractCaller) Count(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Countercontract.contract.Call(opts, &out, "count")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() view returns(uint256)
func (_Countercontract *CountercontractSession) Count() (*big.Int, error) {
	return _Countercontract.Contract.Count(&_Countercontract.CallOpts)
}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() view returns(uint256)
func (_Countercontract *CountercontractCallerSession) Count() (*big.Int, error) {
	return _Countercontract.Contract.Count(&_Countercontract.CallOpts)
}

// Increment is a paid mutator transaction binding the contract method 0xd09de08a.
//
// Solidity: function increment() returns()
func (_Countercontract *CountercontractTransactor) Increment(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Countercontract.contract.Transact(opts, "increment")
}

// Increment is a paid mutator transaction binding the contract method 0xd09de08a.
//
// Solidity: function increment() returns()
func (_Countercontract *CountercontractSession) Increment() (*types.Transaction, error) {
	return _Countercontract.Contract.Increment(&_Countercontract.TransactOpts)
}

// Increment is a paid mutator transaction binding the contract method 0xd09de08a.
//
// Solidity: function increment() returns()
func (_Countercontract *CountercontractTransactorSession) Increment() (*types.Transaction, error) {
	return _Countercontract.Contract.Increment(&_Countercontract.TransactOpts)
}
