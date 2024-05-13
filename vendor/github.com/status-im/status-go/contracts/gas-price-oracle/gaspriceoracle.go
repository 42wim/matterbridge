// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package gaspriceoracle

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

// GaspriceoracleMetaData contains all meta data concerning the Gaspriceoracle contract.
var GaspriceoracleMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"DecimalsUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"GasPriceUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"L1BaseFeeUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"OverheadUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"ScalarUpdated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"adminCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"gasPrice\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAdmin\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"adminAddress\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"getL1Fee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"getL1GasUsed\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"initPayload\",\"type\":\"bytes\"}],\"name\":\"init\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"l1BaseFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"overhead\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"scalar\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"adminAddress\",\"type\":\"address\"}],\"name\":\"setAdmin\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_decimals\",\"type\":\"uint256\"}],\"name\":\"setDecimals\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_gasPrice\",\"type\":\"uint256\"}],\"name\":\"setGasPrice\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_baseFee\",\"type\":\"uint256\"}],\"name\":\"setL1BaseFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_overhead\",\"type\":\"uint256\"}],\"name\":\"setOverhead\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_scalar\",\"type\":\"uint256\"}],\"name\":\"setScalar\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// GaspriceoracleABI is the input ABI used to generate the binding from.
// Deprecated: Use GaspriceoracleMetaData.ABI instead.
var GaspriceoracleABI = GaspriceoracleMetaData.ABI

// Gaspriceoracle is an auto generated Go binding around an Ethereum contract.
type Gaspriceoracle struct {
	GaspriceoracleCaller     // Read-only binding to the contract
	GaspriceoracleTransactor // Write-only binding to the contract
	GaspriceoracleFilterer   // Log filterer for contract events
}

// GaspriceoracleCaller is an auto generated read-only Go binding around an Ethereum contract.
type GaspriceoracleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GaspriceoracleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GaspriceoracleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GaspriceoracleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GaspriceoracleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GaspriceoracleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GaspriceoracleSession struct {
	Contract     *Gaspriceoracle   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GaspriceoracleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GaspriceoracleCallerSession struct {
	Contract *GaspriceoracleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// GaspriceoracleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GaspriceoracleTransactorSession struct {
	Contract     *GaspriceoracleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// GaspriceoracleRaw is an auto generated low-level Go binding around an Ethereum contract.
type GaspriceoracleRaw struct {
	Contract *Gaspriceoracle // Generic contract binding to access the raw methods on
}

// GaspriceoracleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GaspriceoracleCallerRaw struct {
	Contract *GaspriceoracleCaller // Generic read-only contract binding to access the raw methods on
}

// GaspriceoracleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GaspriceoracleTransactorRaw struct {
	Contract *GaspriceoracleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGaspriceoracle creates a new instance of Gaspriceoracle, bound to a specific deployed contract.
func NewGaspriceoracle(address common.Address, backend bind.ContractBackend) (*Gaspriceoracle, error) {
	contract, err := bindGaspriceoracle(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Gaspriceoracle{GaspriceoracleCaller: GaspriceoracleCaller{contract: contract}, GaspriceoracleTransactor: GaspriceoracleTransactor{contract: contract}, GaspriceoracleFilterer: GaspriceoracleFilterer{contract: contract}}, nil
}

// NewGaspriceoracleCaller creates a new read-only instance of Gaspriceoracle, bound to a specific deployed contract.
func NewGaspriceoracleCaller(address common.Address, caller bind.ContractCaller) (*GaspriceoracleCaller, error) {
	contract, err := bindGaspriceoracle(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GaspriceoracleCaller{contract: contract}, nil
}

// NewGaspriceoracleTransactor creates a new write-only instance of Gaspriceoracle, bound to a specific deployed contract.
func NewGaspriceoracleTransactor(address common.Address, transactor bind.ContractTransactor) (*GaspriceoracleTransactor, error) {
	contract, err := bindGaspriceoracle(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GaspriceoracleTransactor{contract: contract}, nil
}

// NewGaspriceoracleFilterer creates a new log filterer instance of Gaspriceoracle, bound to a specific deployed contract.
func NewGaspriceoracleFilterer(address common.Address, filterer bind.ContractFilterer) (*GaspriceoracleFilterer, error) {
	contract, err := bindGaspriceoracle(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GaspriceoracleFilterer{contract: contract}, nil
}

// bindGaspriceoracle binds a generic wrapper to an already deployed contract.
func bindGaspriceoracle(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := GaspriceoracleMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Gaspriceoracle *GaspriceoracleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Gaspriceoracle.Contract.GaspriceoracleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Gaspriceoracle *GaspriceoracleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.GaspriceoracleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Gaspriceoracle *GaspriceoracleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.GaspriceoracleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Gaspriceoracle *GaspriceoracleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Gaspriceoracle.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Gaspriceoracle *GaspriceoracleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Gaspriceoracle *GaspriceoracleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.contract.Transact(opts, method, params...)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_Gaspriceoracle *GaspriceoracleCaller) Admin(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Gaspriceoracle.contract.Call(opts, &out, "admin")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_Gaspriceoracle *GaspriceoracleSession) Admin() (common.Address, error) {
	return _Gaspriceoracle.Contract.Admin(&_Gaspriceoracle.CallOpts)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_Gaspriceoracle *GaspriceoracleCallerSession) Admin() (common.Address, error) {
	return _Gaspriceoracle.Contract.Admin(&_Gaspriceoracle.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleCaller) Decimals(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Gaspriceoracle.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleSession) Decimals() (*big.Int, error) {
	return _Gaspriceoracle.Contract.Decimals(&_Gaspriceoracle.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleCallerSession) Decimals() (*big.Int, error) {
	return _Gaspriceoracle.Contract.Decimals(&_Gaspriceoracle.CallOpts)
}

// GasPrice is a free data retrieval call binding the contract method 0xfe173b97.
//
// Solidity: function gasPrice() view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleCaller) GasPrice(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Gaspriceoracle.contract.Call(opts, &out, "gasPrice")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GasPrice is a free data retrieval call binding the contract method 0xfe173b97.
//
// Solidity: function gasPrice() view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleSession) GasPrice() (*big.Int, error) {
	return _Gaspriceoracle.Contract.GasPrice(&_Gaspriceoracle.CallOpts)
}

// GasPrice is a free data retrieval call binding the contract method 0xfe173b97.
//
// Solidity: function gasPrice() view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleCallerSession) GasPrice() (*big.Int, error) {
	return _Gaspriceoracle.Contract.GasPrice(&_Gaspriceoracle.CallOpts)
}

// GetAdmin is a free data retrieval call binding the contract method 0x6e9960c3.
//
// Solidity: function getAdmin() view returns(address adminAddress)
func (_Gaspriceoracle *GaspriceoracleCaller) GetAdmin(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Gaspriceoracle.contract.Call(opts, &out, "getAdmin")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetAdmin is a free data retrieval call binding the contract method 0x6e9960c3.
//
// Solidity: function getAdmin() view returns(address adminAddress)
func (_Gaspriceoracle *GaspriceoracleSession) GetAdmin() (common.Address, error) {
	return _Gaspriceoracle.Contract.GetAdmin(&_Gaspriceoracle.CallOpts)
}

// GetAdmin is a free data retrieval call binding the contract method 0x6e9960c3.
//
// Solidity: function getAdmin() view returns(address adminAddress)
func (_Gaspriceoracle *GaspriceoracleCallerSession) GetAdmin() (common.Address, error) {
	return _Gaspriceoracle.Contract.GetAdmin(&_Gaspriceoracle.CallOpts)
}

// GetL1Fee is a free data retrieval call binding the contract method 0x49948e0e.
//
// Solidity: function getL1Fee(bytes _data) view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleCaller) GetL1Fee(opts *bind.CallOpts, _data []byte) (*big.Int, error) {
	var out []interface{}
	err := _Gaspriceoracle.contract.Call(opts, &out, "getL1Fee", _data)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetL1Fee is a free data retrieval call binding the contract method 0x49948e0e.
//
// Solidity: function getL1Fee(bytes _data) view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleSession) GetL1Fee(_data []byte) (*big.Int, error) {
	return _Gaspriceoracle.Contract.GetL1Fee(&_Gaspriceoracle.CallOpts, _data)
}

// GetL1Fee is a free data retrieval call binding the contract method 0x49948e0e.
//
// Solidity: function getL1Fee(bytes _data) view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleCallerSession) GetL1Fee(_data []byte) (*big.Int, error) {
	return _Gaspriceoracle.Contract.GetL1Fee(&_Gaspriceoracle.CallOpts, _data)
}

// GetL1GasUsed is a free data retrieval call binding the contract method 0xde26c4a1.
//
// Solidity: function getL1GasUsed(bytes _data) view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleCaller) GetL1GasUsed(opts *bind.CallOpts, _data []byte) (*big.Int, error) {
	var out []interface{}
	err := _Gaspriceoracle.contract.Call(opts, &out, "getL1GasUsed", _data)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetL1GasUsed is a free data retrieval call binding the contract method 0xde26c4a1.
//
// Solidity: function getL1GasUsed(bytes _data) view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleSession) GetL1GasUsed(_data []byte) (*big.Int, error) {
	return _Gaspriceoracle.Contract.GetL1GasUsed(&_Gaspriceoracle.CallOpts, _data)
}

// GetL1GasUsed is a free data retrieval call binding the contract method 0xde26c4a1.
//
// Solidity: function getL1GasUsed(bytes _data) view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleCallerSession) GetL1GasUsed(_data []byte) (*big.Int, error) {
	return _Gaspriceoracle.Contract.GetL1GasUsed(&_Gaspriceoracle.CallOpts, _data)
}

// L1BaseFee is a free data retrieval call binding the contract method 0x519b4bd3.
//
// Solidity: function l1BaseFee() view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleCaller) L1BaseFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Gaspriceoracle.contract.Call(opts, &out, "l1BaseFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// L1BaseFee is a free data retrieval call binding the contract method 0x519b4bd3.
//
// Solidity: function l1BaseFee() view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleSession) L1BaseFee() (*big.Int, error) {
	return _Gaspriceoracle.Contract.L1BaseFee(&_Gaspriceoracle.CallOpts)
}

// L1BaseFee is a free data retrieval call binding the contract method 0x519b4bd3.
//
// Solidity: function l1BaseFee() view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleCallerSession) L1BaseFee() (*big.Int, error) {
	return _Gaspriceoracle.Contract.L1BaseFee(&_Gaspriceoracle.CallOpts)
}

// Overhead is a free data retrieval call binding the contract method 0x0c18c162.
//
// Solidity: function overhead() view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleCaller) Overhead(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Gaspriceoracle.contract.Call(opts, &out, "overhead")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Overhead is a free data retrieval call binding the contract method 0x0c18c162.
//
// Solidity: function overhead() view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleSession) Overhead() (*big.Int, error) {
	return _Gaspriceoracle.Contract.Overhead(&_Gaspriceoracle.CallOpts)
}

// Overhead is a free data retrieval call binding the contract method 0x0c18c162.
//
// Solidity: function overhead() view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleCallerSession) Overhead() (*big.Int, error) {
	return _Gaspriceoracle.Contract.Overhead(&_Gaspriceoracle.CallOpts)
}

// Scalar is a free data retrieval call binding the contract method 0xf45e65d8.
//
// Solidity: function scalar() view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleCaller) Scalar(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Gaspriceoracle.contract.Call(opts, &out, "scalar")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Scalar is a free data retrieval call binding the contract method 0xf45e65d8.
//
// Solidity: function scalar() view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleSession) Scalar() (*big.Int, error) {
	return _Gaspriceoracle.Contract.Scalar(&_Gaspriceoracle.CallOpts)
}

// Scalar is a free data retrieval call binding the contract method 0xf45e65d8.
//
// Solidity: function scalar() view returns(uint256)
func (_Gaspriceoracle *GaspriceoracleCallerSession) Scalar() (*big.Int, error) {
	return _Gaspriceoracle.Contract.Scalar(&_Gaspriceoracle.CallOpts)
}

// AdminCall is a paid mutator transaction binding the contract method 0xbf64a82d.
//
// Solidity: function adminCall(address target, bytes data) payable returns()
func (_Gaspriceoracle *GaspriceoracleTransactor) AdminCall(opts *bind.TransactOpts, target common.Address, data []byte) (*types.Transaction, error) {
	return _Gaspriceoracle.contract.Transact(opts, "adminCall", target, data)
}

// AdminCall is a paid mutator transaction binding the contract method 0xbf64a82d.
//
// Solidity: function adminCall(address target, bytes data) payable returns()
func (_Gaspriceoracle *GaspriceoracleSession) AdminCall(target common.Address, data []byte) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.AdminCall(&_Gaspriceoracle.TransactOpts, target, data)
}

// AdminCall is a paid mutator transaction binding the contract method 0xbf64a82d.
//
// Solidity: function adminCall(address target, bytes data) payable returns()
func (_Gaspriceoracle *GaspriceoracleTransactorSession) AdminCall(target common.Address, data []byte) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.AdminCall(&_Gaspriceoracle.TransactOpts, target, data)
}

// Init is a paid mutator transaction binding the contract method 0x4ddf47d4.
//
// Solidity: function init(bytes initPayload) returns(bytes4)
func (_Gaspriceoracle *GaspriceoracleTransactor) Init(opts *bind.TransactOpts, initPayload []byte) (*types.Transaction, error) {
	return _Gaspriceoracle.contract.Transact(opts, "init", initPayload)
}

// Init is a paid mutator transaction binding the contract method 0x4ddf47d4.
//
// Solidity: function init(bytes initPayload) returns(bytes4)
func (_Gaspriceoracle *GaspriceoracleSession) Init(initPayload []byte) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.Init(&_Gaspriceoracle.TransactOpts, initPayload)
}

// Init is a paid mutator transaction binding the contract method 0x4ddf47d4.
//
// Solidity: function init(bytes initPayload) returns(bytes4)
func (_Gaspriceoracle *GaspriceoracleTransactorSession) Init(initPayload []byte) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.Init(&_Gaspriceoracle.TransactOpts, initPayload)
}

// SetAdmin is a paid mutator transaction binding the contract method 0x704b6c02.
//
// Solidity: function setAdmin(address adminAddress) returns()
func (_Gaspriceoracle *GaspriceoracleTransactor) SetAdmin(opts *bind.TransactOpts, adminAddress common.Address) (*types.Transaction, error) {
	return _Gaspriceoracle.contract.Transact(opts, "setAdmin", adminAddress)
}

// SetAdmin is a paid mutator transaction binding the contract method 0x704b6c02.
//
// Solidity: function setAdmin(address adminAddress) returns()
func (_Gaspriceoracle *GaspriceoracleSession) SetAdmin(adminAddress common.Address) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.SetAdmin(&_Gaspriceoracle.TransactOpts, adminAddress)
}

// SetAdmin is a paid mutator transaction binding the contract method 0x704b6c02.
//
// Solidity: function setAdmin(address adminAddress) returns()
func (_Gaspriceoracle *GaspriceoracleTransactorSession) SetAdmin(adminAddress common.Address) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.SetAdmin(&_Gaspriceoracle.TransactOpts, adminAddress)
}

// SetDecimals is a paid mutator transaction binding the contract method 0x8c8885c8.
//
// Solidity: function setDecimals(uint256 _decimals) returns()
func (_Gaspriceoracle *GaspriceoracleTransactor) SetDecimals(opts *bind.TransactOpts, _decimals *big.Int) (*types.Transaction, error) {
	return _Gaspriceoracle.contract.Transact(opts, "setDecimals", _decimals)
}

// SetDecimals is a paid mutator transaction binding the contract method 0x8c8885c8.
//
// Solidity: function setDecimals(uint256 _decimals) returns()
func (_Gaspriceoracle *GaspriceoracleSession) SetDecimals(_decimals *big.Int) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.SetDecimals(&_Gaspriceoracle.TransactOpts, _decimals)
}

// SetDecimals is a paid mutator transaction binding the contract method 0x8c8885c8.
//
// Solidity: function setDecimals(uint256 _decimals) returns()
func (_Gaspriceoracle *GaspriceoracleTransactorSession) SetDecimals(_decimals *big.Int) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.SetDecimals(&_Gaspriceoracle.TransactOpts, _decimals)
}

// SetGasPrice is a paid mutator transaction binding the contract method 0xbf1fe420.
//
// Solidity: function setGasPrice(uint256 _gasPrice) returns()
func (_Gaspriceoracle *GaspriceoracleTransactor) SetGasPrice(opts *bind.TransactOpts, _gasPrice *big.Int) (*types.Transaction, error) {
	return _Gaspriceoracle.contract.Transact(opts, "setGasPrice", _gasPrice)
}

// SetGasPrice is a paid mutator transaction binding the contract method 0xbf1fe420.
//
// Solidity: function setGasPrice(uint256 _gasPrice) returns()
func (_Gaspriceoracle *GaspriceoracleSession) SetGasPrice(_gasPrice *big.Int) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.SetGasPrice(&_Gaspriceoracle.TransactOpts, _gasPrice)
}

// SetGasPrice is a paid mutator transaction binding the contract method 0xbf1fe420.
//
// Solidity: function setGasPrice(uint256 _gasPrice) returns()
func (_Gaspriceoracle *GaspriceoracleTransactorSession) SetGasPrice(_gasPrice *big.Int) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.SetGasPrice(&_Gaspriceoracle.TransactOpts, _gasPrice)
}

// SetL1BaseFee is a paid mutator transaction binding the contract method 0xbede39b5.
//
// Solidity: function setL1BaseFee(uint256 _baseFee) returns()
func (_Gaspriceoracle *GaspriceoracleTransactor) SetL1BaseFee(opts *bind.TransactOpts, _baseFee *big.Int) (*types.Transaction, error) {
	return _Gaspriceoracle.contract.Transact(opts, "setL1BaseFee", _baseFee)
}

// SetL1BaseFee is a paid mutator transaction binding the contract method 0xbede39b5.
//
// Solidity: function setL1BaseFee(uint256 _baseFee) returns()
func (_Gaspriceoracle *GaspriceoracleSession) SetL1BaseFee(_baseFee *big.Int) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.SetL1BaseFee(&_Gaspriceoracle.TransactOpts, _baseFee)
}

// SetL1BaseFee is a paid mutator transaction binding the contract method 0xbede39b5.
//
// Solidity: function setL1BaseFee(uint256 _baseFee) returns()
func (_Gaspriceoracle *GaspriceoracleTransactorSession) SetL1BaseFee(_baseFee *big.Int) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.SetL1BaseFee(&_Gaspriceoracle.TransactOpts, _baseFee)
}

// SetOverhead is a paid mutator transaction binding the contract method 0x3577afc5.
//
// Solidity: function setOverhead(uint256 _overhead) returns()
func (_Gaspriceoracle *GaspriceoracleTransactor) SetOverhead(opts *bind.TransactOpts, _overhead *big.Int) (*types.Transaction, error) {
	return _Gaspriceoracle.contract.Transact(opts, "setOverhead", _overhead)
}

// SetOverhead is a paid mutator transaction binding the contract method 0x3577afc5.
//
// Solidity: function setOverhead(uint256 _overhead) returns()
func (_Gaspriceoracle *GaspriceoracleSession) SetOverhead(_overhead *big.Int) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.SetOverhead(&_Gaspriceoracle.TransactOpts, _overhead)
}

// SetOverhead is a paid mutator transaction binding the contract method 0x3577afc5.
//
// Solidity: function setOverhead(uint256 _overhead) returns()
func (_Gaspriceoracle *GaspriceoracleTransactorSession) SetOverhead(_overhead *big.Int) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.SetOverhead(&_Gaspriceoracle.TransactOpts, _overhead)
}

// SetScalar is a paid mutator transaction binding the contract method 0x70465597.
//
// Solidity: function setScalar(uint256 _scalar) returns()
func (_Gaspriceoracle *GaspriceoracleTransactor) SetScalar(opts *bind.TransactOpts, _scalar *big.Int) (*types.Transaction, error) {
	return _Gaspriceoracle.contract.Transact(opts, "setScalar", _scalar)
}

// SetScalar is a paid mutator transaction binding the contract method 0x70465597.
//
// Solidity: function setScalar(uint256 _scalar) returns()
func (_Gaspriceoracle *GaspriceoracleSession) SetScalar(_scalar *big.Int) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.SetScalar(&_Gaspriceoracle.TransactOpts, _scalar)
}

// SetScalar is a paid mutator transaction binding the contract method 0x70465597.
//
// Solidity: function setScalar(uint256 _scalar) returns()
func (_Gaspriceoracle *GaspriceoracleTransactorSession) SetScalar(_scalar *big.Int) (*types.Transaction, error) {
	return _Gaspriceoracle.Contract.SetScalar(&_Gaspriceoracle.TransactOpts, _scalar)
}

// GaspriceoracleDecimalsUpdatedIterator is returned from FilterDecimalsUpdated and is used to iterate over the raw logs and unpacked data for DecimalsUpdated events raised by the Gaspriceoracle contract.
type GaspriceoracleDecimalsUpdatedIterator struct {
	Event *GaspriceoracleDecimalsUpdated // Event containing the contract specifics and raw log

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
func (it *GaspriceoracleDecimalsUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GaspriceoracleDecimalsUpdated)
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
		it.Event = new(GaspriceoracleDecimalsUpdated)
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
func (it *GaspriceoracleDecimalsUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GaspriceoracleDecimalsUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GaspriceoracleDecimalsUpdated represents a DecimalsUpdated event raised by the Gaspriceoracle contract.
type GaspriceoracleDecimalsUpdated struct {
	Arg0 *big.Int
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterDecimalsUpdated is a free log retrieval operation binding the contract event 0xd68112a8707e326d08be3656b528c1bcc5bbbfc47f4177e2179b14d8640838c1.
//
// Solidity: event DecimalsUpdated(uint256 arg0)
func (_Gaspriceoracle *GaspriceoracleFilterer) FilterDecimalsUpdated(opts *bind.FilterOpts) (*GaspriceoracleDecimalsUpdatedIterator, error) {

	logs, sub, err := _Gaspriceoracle.contract.FilterLogs(opts, "DecimalsUpdated")
	if err != nil {
		return nil, err
	}
	return &GaspriceoracleDecimalsUpdatedIterator{contract: _Gaspriceoracle.contract, event: "DecimalsUpdated", logs: logs, sub: sub}, nil
}

// WatchDecimalsUpdated is a free log subscription operation binding the contract event 0xd68112a8707e326d08be3656b528c1bcc5bbbfc47f4177e2179b14d8640838c1.
//
// Solidity: event DecimalsUpdated(uint256 arg0)
func (_Gaspriceoracle *GaspriceoracleFilterer) WatchDecimalsUpdated(opts *bind.WatchOpts, sink chan<- *GaspriceoracleDecimalsUpdated) (event.Subscription, error) {

	logs, sub, err := _Gaspriceoracle.contract.WatchLogs(opts, "DecimalsUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GaspriceoracleDecimalsUpdated)
				if err := _Gaspriceoracle.contract.UnpackLog(event, "DecimalsUpdated", log); err != nil {
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

// ParseDecimalsUpdated is a log parse operation binding the contract event 0xd68112a8707e326d08be3656b528c1bcc5bbbfc47f4177e2179b14d8640838c1.
//
// Solidity: event DecimalsUpdated(uint256 arg0)
func (_Gaspriceoracle *GaspriceoracleFilterer) ParseDecimalsUpdated(log types.Log) (*GaspriceoracleDecimalsUpdated, error) {
	event := new(GaspriceoracleDecimalsUpdated)
	if err := _Gaspriceoracle.contract.UnpackLog(event, "DecimalsUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GaspriceoracleGasPriceUpdatedIterator is returned from FilterGasPriceUpdated and is used to iterate over the raw logs and unpacked data for GasPriceUpdated events raised by the Gaspriceoracle contract.
type GaspriceoracleGasPriceUpdatedIterator struct {
	Event *GaspriceoracleGasPriceUpdated // Event containing the contract specifics and raw log

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
func (it *GaspriceoracleGasPriceUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GaspriceoracleGasPriceUpdated)
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
		it.Event = new(GaspriceoracleGasPriceUpdated)
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
func (it *GaspriceoracleGasPriceUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GaspriceoracleGasPriceUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GaspriceoracleGasPriceUpdated represents a GasPriceUpdated event raised by the Gaspriceoracle contract.
type GaspriceoracleGasPriceUpdated struct {
	Arg0 *big.Int
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterGasPriceUpdated is a free log retrieval operation binding the contract event 0xfcdccc6074c6c42e4bd578aa9870c697dc976a270968452d2b8c8dc369fae396.
//
// Solidity: event GasPriceUpdated(uint256 arg0)
func (_Gaspriceoracle *GaspriceoracleFilterer) FilterGasPriceUpdated(opts *bind.FilterOpts) (*GaspriceoracleGasPriceUpdatedIterator, error) {

	logs, sub, err := _Gaspriceoracle.contract.FilterLogs(opts, "GasPriceUpdated")
	if err != nil {
		return nil, err
	}
	return &GaspriceoracleGasPriceUpdatedIterator{contract: _Gaspriceoracle.contract, event: "GasPriceUpdated", logs: logs, sub: sub}, nil
}

// WatchGasPriceUpdated is a free log subscription operation binding the contract event 0xfcdccc6074c6c42e4bd578aa9870c697dc976a270968452d2b8c8dc369fae396.
//
// Solidity: event GasPriceUpdated(uint256 arg0)
func (_Gaspriceoracle *GaspriceoracleFilterer) WatchGasPriceUpdated(opts *bind.WatchOpts, sink chan<- *GaspriceoracleGasPriceUpdated) (event.Subscription, error) {

	logs, sub, err := _Gaspriceoracle.contract.WatchLogs(opts, "GasPriceUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GaspriceoracleGasPriceUpdated)
				if err := _Gaspriceoracle.contract.UnpackLog(event, "GasPriceUpdated", log); err != nil {
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

// ParseGasPriceUpdated is a log parse operation binding the contract event 0xfcdccc6074c6c42e4bd578aa9870c697dc976a270968452d2b8c8dc369fae396.
//
// Solidity: event GasPriceUpdated(uint256 arg0)
func (_Gaspriceoracle *GaspriceoracleFilterer) ParseGasPriceUpdated(log types.Log) (*GaspriceoracleGasPriceUpdated, error) {
	event := new(GaspriceoracleGasPriceUpdated)
	if err := _Gaspriceoracle.contract.UnpackLog(event, "GasPriceUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GaspriceoracleL1BaseFeeUpdatedIterator is returned from FilterL1BaseFeeUpdated and is used to iterate over the raw logs and unpacked data for L1BaseFeeUpdated events raised by the Gaspriceoracle contract.
type GaspriceoracleL1BaseFeeUpdatedIterator struct {
	Event *GaspriceoracleL1BaseFeeUpdated // Event containing the contract specifics and raw log

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
func (it *GaspriceoracleL1BaseFeeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GaspriceoracleL1BaseFeeUpdated)
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
		it.Event = new(GaspriceoracleL1BaseFeeUpdated)
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
func (it *GaspriceoracleL1BaseFeeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GaspriceoracleL1BaseFeeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GaspriceoracleL1BaseFeeUpdated represents a L1BaseFeeUpdated event raised by the Gaspriceoracle contract.
type GaspriceoracleL1BaseFeeUpdated struct {
	Arg0 *big.Int
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterL1BaseFeeUpdated is a free log retrieval operation binding the contract event 0x351fb23757bb5ea0546c85b7996ddd7155f96b939ebaa5ff7bc49c75f27f2c44.
//
// Solidity: event L1BaseFeeUpdated(uint256 arg0)
func (_Gaspriceoracle *GaspriceoracleFilterer) FilterL1BaseFeeUpdated(opts *bind.FilterOpts) (*GaspriceoracleL1BaseFeeUpdatedIterator, error) {

	logs, sub, err := _Gaspriceoracle.contract.FilterLogs(opts, "L1BaseFeeUpdated")
	if err != nil {
		return nil, err
	}
	return &GaspriceoracleL1BaseFeeUpdatedIterator{contract: _Gaspriceoracle.contract, event: "L1BaseFeeUpdated", logs: logs, sub: sub}, nil
}

// WatchL1BaseFeeUpdated is a free log subscription operation binding the contract event 0x351fb23757bb5ea0546c85b7996ddd7155f96b939ebaa5ff7bc49c75f27f2c44.
//
// Solidity: event L1BaseFeeUpdated(uint256 arg0)
func (_Gaspriceoracle *GaspriceoracleFilterer) WatchL1BaseFeeUpdated(opts *bind.WatchOpts, sink chan<- *GaspriceoracleL1BaseFeeUpdated) (event.Subscription, error) {

	logs, sub, err := _Gaspriceoracle.contract.WatchLogs(opts, "L1BaseFeeUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GaspriceoracleL1BaseFeeUpdated)
				if err := _Gaspriceoracle.contract.UnpackLog(event, "L1BaseFeeUpdated", log); err != nil {
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

// ParseL1BaseFeeUpdated is a log parse operation binding the contract event 0x351fb23757bb5ea0546c85b7996ddd7155f96b939ebaa5ff7bc49c75f27f2c44.
//
// Solidity: event L1BaseFeeUpdated(uint256 arg0)
func (_Gaspriceoracle *GaspriceoracleFilterer) ParseL1BaseFeeUpdated(log types.Log) (*GaspriceoracleL1BaseFeeUpdated, error) {
	event := new(GaspriceoracleL1BaseFeeUpdated)
	if err := _Gaspriceoracle.contract.UnpackLog(event, "L1BaseFeeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GaspriceoracleOverheadUpdatedIterator is returned from FilterOverheadUpdated and is used to iterate over the raw logs and unpacked data for OverheadUpdated events raised by the Gaspriceoracle contract.
type GaspriceoracleOverheadUpdatedIterator struct {
	Event *GaspriceoracleOverheadUpdated // Event containing the contract specifics and raw log

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
func (it *GaspriceoracleOverheadUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GaspriceoracleOverheadUpdated)
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
		it.Event = new(GaspriceoracleOverheadUpdated)
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
func (it *GaspriceoracleOverheadUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GaspriceoracleOverheadUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GaspriceoracleOverheadUpdated represents a OverheadUpdated event raised by the Gaspriceoracle contract.
type GaspriceoracleOverheadUpdated struct {
	Arg0 *big.Int
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterOverheadUpdated is a free log retrieval operation binding the contract event 0x32740b35c0ea213650f60d44366b4fb211c9033b50714e4a1d34e65d5beb9bb4.
//
// Solidity: event OverheadUpdated(uint256 arg0)
func (_Gaspriceoracle *GaspriceoracleFilterer) FilterOverheadUpdated(opts *bind.FilterOpts) (*GaspriceoracleOverheadUpdatedIterator, error) {

	logs, sub, err := _Gaspriceoracle.contract.FilterLogs(opts, "OverheadUpdated")
	if err != nil {
		return nil, err
	}
	return &GaspriceoracleOverheadUpdatedIterator{contract: _Gaspriceoracle.contract, event: "OverheadUpdated", logs: logs, sub: sub}, nil
}

// WatchOverheadUpdated is a free log subscription operation binding the contract event 0x32740b35c0ea213650f60d44366b4fb211c9033b50714e4a1d34e65d5beb9bb4.
//
// Solidity: event OverheadUpdated(uint256 arg0)
func (_Gaspriceoracle *GaspriceoracleFilterer) WatchOverheadUpdated(opts *bind.WatchOpts, sink chan<- *GaspriceoracleOverheadUpdated) (event.Subscription, error) {

	logs, sub, err := _Gaspriceoracle.contract.WatchLogs(opts, "OverheadUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GaspriceoracleOverheadUpdated)
				if err := _Gaspriceoracle.contract.UnpackLog(event, "OverheadUpdated", log); err != nil {
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

// ParseOverheadUpdated is a log parse operation binding the contract event 0x32740b35c0ea213650f60d44366b4fb211c9033b50714e4a1d34e65d5beb9bb4.
//
// Solidity: event OverheadUpdated(uint256 arg0)
func (_Gaspriceoracle *GaspriceoracleFilterer) ParseOverheadUpdated(log types.Log) (*GaspriceoracleOverheadUpdated, error) {
	event := new(GaspriceoracleOverheadUpdated)
	if err := _Gaspriceoracle.contract.UnpackLog(event, "OverheadUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GaspriceoracleScalarUpdatedIterator is returned from FilterScalarUpdated and is used to iterate over the raw logs and unpacked data for ScalarUpdated events raised by the Gaspriceoracle contract.
type GaspriceoracleScalarUpdatedIterator struct {
	Event *GaspriceoracleScalarUpdated // Event containing the contract specifics and raw log

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
func (it *GaspriceoracleScalarUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GaspriceoracleScalarUpdated)
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
		it.Event = new(GaspriceoracleScalarUpdated)
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
func (it *GaspriceoracleScalarUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GaspriceoracleScalarUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GaspriceoracleScalarUpdated represents a ScalarUpdated event raised by the Gaspriceoracle contract.
type GaspriceoracleScalarUpdated struct {
	Arg0 *big.Int
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterScalarUpdated is a free log retrieval operation binding the contract event 0x3336cd9708eaf2769a0f0dc0679f30e80f15dcd88d1921b5a16858e8b85c591a.
//
// Solidity: event ScalarUpdated(uint256 arg0)
func (_Gaspriceoracle *GaspriceoracleFilterer) FilterScalarUpdated(opts *bind.FilterOpts) (*GaspriceoracleScalarUpdatedIterator, error) {

	logs, sub, err := _Gaspriceoracle.contract.FilterLogs(opts, "ScalarUpdated")
	if err != nil {
		return nil, err
	}
	return &GaspriceoracleScalarUpdatedIterator{contract: _Gaspriceoracle.contract, event: "ScalarUpdated", logs: logs, sub: sub}, nil
}

// WatchScalarUpdated is a free log subscription operation binding the contract event 0x3336cd9708eaf2769a0f0dc0679f30e80f15dcd88d1921b5a16858e8b85c591a.
//
// Solidity: event ScalarUpdated(uint256 arg0)
func (_Gaspriceoracle *GaspriceoracleFilterer) WatchScalarUpdated(opts *bind.WatchOpts, sink chan<- *GaspriceoracleScalarUpdated) (event.Subscription, error) {

	logs, sub, err := _Gaspriceoracle.contract.WatchLogs(opts, "ScalarUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GaspriceoracleScalarUpdated)
				if err := _Gaspriceoracle.contract.UnpackLog(event, "ScalarUpdated", log); err != nil {
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

// ParseScalarUpdated is a log parse operation binding the contract event 0x3336cd9708eaf2769a0f0dc0679f30e80f15dcd88d1921b5a16858e8b85c591a.
//
// Solidity: event ScalarUpdated(uint256 arg0)
func (_Gaspriceoracle *GaspriceoracleFilterer) ParseScalarUpdated(log types.Log) (*GaspriceoracleScalarUpdated, error) {
	event := new(GaspriceoracleScalarUpdated)
	if err := _Gaspriceoracle.contract.UnpackLog(event, "ScalarUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
