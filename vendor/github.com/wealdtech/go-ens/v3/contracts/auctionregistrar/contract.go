// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package auctionregistrar

import (
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
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// ContractABI is the input ABI used to generate the binding from.
const ContractABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_hash\",\"type\":\"bytes32\"}],\"name\":\"releaseDeed\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_hash\",\"type\":\"bytes32\"}],\"name\":\"getAllowedTime\",\"outputs\":[{\"name\":\"timestamp\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"unhashedName\",\"type\":\"string\"}],\"name\":\"invalidateName\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"},{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"bytes32\"}],\"name\":\"shaBid\",\"outputs\":[{\"name\":\"sealedBid\",\"type\":\"bytes32\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"bidder\",\"type\":\"address\"},{\"name\":\"seal\",\"type\":\"bytes32\"}],\"name\":\"cancelBid\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_hash\",\"type\":\"bytes32\"}],\"name\":\"entries\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"},{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"ens\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_hash\",\"type\":\"bytes32\"},{\"name\":\"_value\",\"type\":\"uint256\"},{\"name\":\"_salt\",\"type\":\"bytes32\"}],\"name\":\"unsealBid\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_hash\",\"type\":\"bytes32\"}],\"name\":\"transferRegistrars\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"sealedBids\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_hash\",\"type\":\"bytes32\"}],\"name\":\"state\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_hash\",\"type\":\"bytes32\"},{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transfer\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_hash\",\"type\":\"bytes32\"},{\"name\":\"_timestamp\",\"type\":\"uint256\"}],\"name\":\"isAllowed\",\"outputs\":[{\"name\":\"allowed\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_hash\",\"type\":\"bytes32\"}],\"name\":\"finalizeAuction\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"registryStarted\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"launchLength\",\"outputs\":[{\"name\":\"\",\"type\":\"uint32\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"sealedBid\",\"type\":\"bytes32\"}],\"name\":\"newBid\",\"outputs\":[],\"payable\":true,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"labels\",\"type\":\"bytes32[]\"}],\"name\":\"eraseNode\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_hashes\",\"type\":\"bytes32[]\"}],\"name\":\"startAuctions\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"},{\"name\":\"deed\",\"type\":\"address\"},{\"name\":\"registrationDate\",\"type\":\"uint256\"}],\"name\":\"acceptRegistrarTransfer\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_hash\",\"type\":\"bytes32\"}],\"name\":\"startAuction\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"rootNode\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"hashes\",\"type\":\"bytes32[]\"},{\"name\":\"sealedBid\",\"type\":\"bytes32\"}],\"name\":\"startAuctionsAndBid\",\"outputs\":[],\"payable\":true,\"type\":\"function\"},{\"inputs\":[{\"name\":\"_ens\",\"type\":\"address\"},{\"name\":\"_rootNode\",\"type\":\"bytes32\"},{\"name\":\"_startDate\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"hash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"registrationDate\",\"type\":\"uint256\"}],\"name\":\"AuctionStarted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"hash\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"bidder\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"deposit\",\"type\":\"uint256\"}],\"name\":\"NewBid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"hash\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"status\",\"type\":\"uint8\"}],\"name\":\"BidRevealed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"hash\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"registrationDate\",\"type\":\"uint256\"}],\"name\":\"HashRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"hash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"HashReleased\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"hash\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"name\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"registrationDate\",\"type\":\"uint256\"}],\"name\":\"HashInvalidated\",\"type\":\"event\"}]"

// Contract is an auto generated Go binding around an Ethereum contract.
type Contract struct {
	ContractCaller     // Read-only binding to the contract
	ContractTransactor // Write-only binding to the contract
	ContractFilterer   // Log filterer for contract events
}

// ContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractSession struct {
	Contract     *Contract         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractCallerSession struct {
	Contract *ContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractTransactorSession struct {
	Contract     *ContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractRaw struct {
	Contract *Contract // Generic contract binding to access the raw methods on
}

// ContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractCallerRaw struct {
	Contract *ContractCaller // Generic read-only contract binding to access the raw methods on
}

// ContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractTransactorRaw struct {
	Contract *ContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContract creates a new instance of Contract, bound to a specific deployed contract.
func NewContract(address common.Address, backend bind.ContractBackend) (*Contract, error) {
	contract, err := bindContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// NewContractCaller creates a new read-only instance of Contract, bound to a specific deployed contract.
func NewContractCaller(address common.Address, caller bind.ContractCaller) (*ContractCaller, error) {
	contract, err := bindContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractCaller{contract: contract}, nil
}

// NewContractTransactor creates a new write-only instance of Contract, bound to a specific deployed contract.
func NewContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractTransactor, error) {
	contract, err := bindContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractTransactor{contract: contract}, nil
}

// NewContractFilterer creates a new log filterer instance of Contract, bound to a specific deployed contract.
func NewContractFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractFilterer, error) {
	contract, err := bindContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractFilterer{contract: contract}, nil
}

// bindContract binds a generic wrapper to an already deployed contract.
func bindContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.ContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transact(opts, method, params...)
}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() returns(address)
func (_Contract *ContractCaller) Ens(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "ens")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() returns(address)
func (_Contract *ContractSession) Ens() (common.Address, error) {
	return _Contract.Contract.Ens(&_Contract.CallOpts)
}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() returns(address)
func (_Contract *ContractCallerSession) Ens() (common.Address, error) {
	return _Contract.Contract.Ens(&_Contract.CallOpts)
}

// Entries is a free data retrieval call binding the contract method 0x267b6922.
//
// Solidity: function entries(bytes32 _hash) returns(uint8, address, uint256, uint256, uint256)
func (_Contract *ContractCaller) Entries(opts *bind.CallOpts, _hash [32]byte) (uint8, common.Address, *big.Int, *big.Int, *big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "entries", _hash)

	if err != nil {
		return *new(uint8), *new(common.Address), *new(*big.Int), *new(*big.Int), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)
	out1 := *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	out2 := *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	out3 := *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	out4 := *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)

	return out0, out1, out2, out3, out4, err

}

// Entries is a free data retrieval call binding the contract method 0x267b6922.
//
// Solidity: function entries(bytes32 _hash) returns(uint8, address, uint256, uint256, uint256)
func (_Contract *ContractSession) Entries(_hash [32]byte) (uint8, common.Address, *big.Int, *big.Int, *big.Int, error) {
	return _Contract.Contract.Entries(&_Contract.CallOpts, _hash)
}

// Entries is a free data retrieval call binding the contract method 0x267b6922.
//
// Solidity: function entries(bytes32 _hash) returns(uint8, address, uint256, uint256, uint256)
func (_Contract *ContractCallerSession) Entries(_hash [32]byte) (uint8, common.Address, *big.Int, *big.Int, *big.Int, error) {
	return _Contract.Contract.Entries(&_Contract.CallOpts, _hash)
}

// GetAllowedTime is a free data retrieval call binding the contract method 0x13c89a8f.
//
// Solidity: function getAllowedTime(bytes32 _hash) returns(uint256 timestamp)
func (_Contract *ContractCaller) GetAllowedTime(opts *bind.CallOpts, _hash [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getAllowedTime", _hash)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAllowedTime is a free data retrieval call binding the contract method 0x13c89a8f.
//
// Solidity: function getAllowedTime(bytes32 _hash) returns(uint256 timestamp)
func (_Contract *ContractSession) GetAllowedTime(_hash [32]byte) (*big.Int, error) {
	return _Contract.Contract.GetAllowedTime(&_Contract.CallOpts, _hash)
}

// GetAllowedTime is a free data retrieval call binding the contract method 0x13c89a8f.
//
// Solidity: function getAllowedTime(bytes32 _hash) returns(uint256 timestamp)
func (_Contract *ContractCallerSession) GetAllowedTime(_hash [32]byte) (*big.Int, error) {
	return _Contract.Contract.GetAllowedTime(&_Contract.CallOpts, _hash)
}

// IsAllowed is a free data retrieval call binding the contract method 0x93503337.
//
// Solidity: function isAllowed(bytes32 _hash, uint256 _timestamp) returns(bool allowed)
func (_Contract *ContractCaller) IsAllowed(opts *bind.CallOpts, _hash [32]byte, _timestamp *big.Int) (bool, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "isAllowed", _hash, _timestamp)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAllowed is a free data retrieval call binding the contract method 0x93503337.
//
// Solidity: function isAllowed(bytes32 _hash, uint256 _timestamp) returns(bool allowed)
func (_Contract *ContractSession) IsAllowed(_hash [32]byte, _timestamp *big.Int) (bool, error) {
	return _Contract.Contract.IsAllowed(&_Contract.CallOpts, _hash, _timestamp)
}

// IsAllowed is a free data retrieval call binding the contract method 0x93503337.
//
// Solidity: function isAllowed(bytes32 _hash, uint256 _timestamp) returns(bool allowed)
func (_Contract *ContractCallerSession) IsAllowed(_hash [32]byte, _timestamp *big.Int) (bool, error) {
	return _Contract.Contract.IsAllowed(&_Contract.CallOpts, _hash, _timestamp)
}

// LaunchLength is a free data retrieval call binding the contract method 0xae1a0b0c.
//
// Solidity: function launchLength() returns(uint32)
func (_Contract *ContractCaller) LaunchLength(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "launchLength")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// LaunchLength is a free data retrieval call binding the contract method 0xae1a0b0c.
//
// Solidity: function launchLength() returns(uint32)
func (_Contract *ContractSession) LaunchLength() (uint32, error) {
	return _Contract.Contract.LaunchLength(&_Contract.CallOpts)
}

// LaunchLength is a free data retrieval call binding the contract method 0xae1a0b0c.
//
// Solidity: function launchLength() returns(uint32)
func (_Contract *ContractCallerSession) LaunchLength() (uint32, error) {
	return _Contract.Contract.LaunchLength(&_Contract.CallOpts)
}

// RegistryStarted is a free data retrieval call binding the contract method 0x9c67f06f.
//
// Solidity: function registryStarted() returns(uint256)
func (_Contract *ContractCaller) RegistryStarted(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "registryStarted")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RegistryStarted is a free data retrieval call binding the contract method 0x9c67f06f.
//
// Solidity: function registryStarted() returns(uint256)
func (_Contract *ContractSession) RegistryStarted() (*big.Int, error) {
	return _Contract.Contract.RegistryStarted(&_Contract.CallOpts)
}

// RegistryStarted is a free data retrieval call binding the contract method 0x9c67f06f.
//
// Solidity: function registryStarted() returns(uint256)
func (_Contract *ContractCallerSession) RegistryStarted() (*big.Int, error) {
	return _Contract.Contract.RegistryStarted(&_Contract.CallOpts)
}

// RootNode is a free data retrieval call binding the contract method 0xfaff50a8.
//
// Solidity: function rootNode() returns(bytes32)
func (_Contract *ContractCaller) RootNode(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "rootNode")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// RootNode is a free data retrieval call binding the contract method 0xfaff50a8.
//
// Solidity: function rootNode() returns(bytes32)
func (_Contract *ContractSession) RootNode() ([32]byte, error) {
	return _Contract.Contract.RootNode(&_Contract.CallOpts)
}

// RootNode is a free data retrieval call binding the contract method 0xfaff50a8.
//
// Solidity: function rootNode() returns(bytes32)
func (_Contract *ContractCallerSession) RootNode() ([32]byte, error) {
	return _Contract.Contract.RootNode(&_Contract.CallOpts)
}

// SealedBids is a free data retrieval call binding the contract method 0x5e431709.
//
// Solidity: function sealedBids(address , bytes32 ) returns(address)
func (_Contract *ContractCaller) SealedBids(opts *bind.CallOpts, arg0 common.Address, arg1 [32]byte) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "sealedBids", arg0, arg1)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SealedBids is a free data retrieval call binding the contract method 0x5e431709.
//
// Solidity: function sealedBids(address , bytes32 ) returns(address)
func (_Contract *ContractSession) SealedBids(arg0 common.Address, arg1 [32]byte) (common.Address, error) {
	return _Contract.Contract.SealedBids(&_Contract.CallOpts, arg0, arg1)
}

// SealedBids is a free data retrieval call binding the contract method 0x5e431709.
//
// Solidity: function sealedBids(address , bytes32 ) returns(address)
func (_Contract *ContractCallerSession) SealedBids(arg0 common.Address, arg1 [32]byte) (common.Address, error) {
	return _Contract.Contract.SealedBids(&_Contract.CallOpts, arg0, arg1)
}

// ShaBid is a free data retrieval call binding the contract method 0x22ec1244.
//
// Solidity: function shaBid(bytes32 hash, address owner, uint256 value, bytes32 salt) returns(bytes32 sealedBid)
func (_Contract *ContractCaller) ShaBid(opts *bind.CallOpts, hash [32]byte, owner common.Address, value *big.Int, salt [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "shaBid", hash, owner, value, salt)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ShaBid is a free data retrieval call binding the contract method 0x22ec1244.
//
// Solidity: function shaBid(bytes32 hash, address owner, uint256 value, bytes32 salt) returns(bytes32 sealedBid)
func (_Contract *ContractSession) ShaBid(hash [32]byte, owner common.Address, value *big.Int, salt [32]byte) ([32]byte, error) {
	return _Contract.Contract.ShaBid(&_Contract.CallOpts, hash, owner, value, salt)
}

// ShaBid is a free data retrieval call binding the contract method 0x22ec1244.
//
// Solidity: function shaBid(bytes32 hash, address owner, uint256 value, bytes32 salt) returns(bytes32 sealedBid)
func (_Contract *ContractCallerSession) ShaBid(hash [32]byte, owner common.Address, value *big.Int, salt [32]byte) ([32]byte, error) {
	return _Contract.Contract.ShaBid(&_Contract.CallOpts, hash, owner, value, salt)
}

// State is a free data retrieval call binding the contract method 0x61d585da.
//
// Solidity: function state(bytes32 _hash) returns(uint8)
func (_Contract *ContractCaller) State(opts *bind.CallOpts, _hash [32]byte) (uint8, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "state", _hash)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// State is a free data retrieval call binding the contract method 0x61d585da.
//
// Solidity: function state(bytes32 _hash) returns(uint8)
func (_Contract *ContractSession) State(_hash [32]byte) (uint8, error) {
	return _Contract.Contract.State(&_Contract.CallOpts, _hash)
}

// State is a free data retrieval call binding the contract method 0x61d585da.
//
// Solidity: function state(bytes32 _hash) returns(uint8)
func (_Contract *ContractCallerSession) State(_hash [32]byte) (uint8, error) {
	return _Contract.Contract.State(&_Contract.CallOpts, _hash)
}

// AcceptRegistrarTransfer is a paid mutator transaction binding the contract method 0xea9e107a.
//
// Solidity: function acceptRegistrarTransfer(bytes32 hash, address deed, uint256 registrationDate) returns()
func (_Contract *ContractTransactor) AcceptRegistrarTransfer(opts *bind.TransactOpts, hash [32]byte, deed common.Address, registrationDate *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "acceptRegistrarTransfer", hash, deed, registrationDate)
}

// AcceptRegistrarTransfer is a paid mutator transaction binding the contract method 0xea9e107a.
//
// Solidity: function acceptRegistrarTransfer(bytes32 hash, address deed, uint256 registrationDate) returns()
func (_Contract *ContractSession) AcceptRegistrarTransfer(hash [32]byte, deed common.Address, registrationDate *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.AcceptRegistrarTransfer(&_Contract.TransactOpts, hash, deed, registrationDate)
}

// AcceptRegistrarTransfer is a paid mutator transaction binding the contract method 0xea9e107a.
//
// Solidity: function acceptRegistrarTransfer(bytes32 hash, address deed, uint256 registrationDate) returns()
func (_Contract *ContractTransactorSession) AcceptRegistrarTransfer(hash [32]byte, deed common.Address, registrationDate *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.AcceptRegistrarTransfer(&_Contract.TransactOpts, hash, deed, registrationDate)
}

// CancelBid is a paid mutator transaction binding the contract method 0x2525f5c1.
//
// Solidity: function cancelBid(address bidder, bytes32 seal) returns()
func (_Contract *ContractTransactor) CancelBid(opts *bind.TransactOpts, bidder common.Address, seal [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "cancelBid", bidder, seal)
}

// CancelBid is a paid mutator transaction binding the contract method 0x2525f5c1.
//
// Solidity: function cancelBid(address bidder, bytes32 seal) returns()
func (_Contract *ContractSession) CancelBid(bidder common.Address, seal [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.CancelBid(&_Contract.TransactOpts, bidder, seal)
}

// CancelBid is a paid mutator transaction binding the contract method 0x2525f5c1.
//
// Solidity: function cancelBid(address bidder, bytes32 seal) returns()
func (_Contract *ContractTransactorSession) CancelBid(bidder common.Address, seal [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.CancelBid(&_Contract.TransactOpts, bidder, seal)
}

// EraseNode is a paid mutator transaction binding the contract method 0xde10f04b.
//
// Solidity: function eraseNode(bytes32[] labels) returns()
func (_Contract *ContractTransactor) EraseNode(opts *bind.TransactOpts, labels [][32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "eraseNode", labels)
}

// EraseNode is a paid mutator transaction binding the contract method 0xde10f04b.
//
// Solidity: function eraseNode(bytes32[] labels) returns()
func (_Contract *ContractSession) EraseNode(labels [][32]byte) (*types.Transaction, error) {
	return _Contract.Contract.EraseNode(&_Contract.TransactOpts, labels)
}

// EraseNode is a paid mutator transaction binding the contract method 0xde10f04b.
//
// Solidity: function eraseNode(bytes32[] labels) returns()
func (_Contract *ContractTransactorSession) EraseNode(labels [][32]byte) (*types.Transaction, error) {
	return _Contract.Contract.EraseNode(&_Contract.TransactOpts, labels)
}

// FinalizeAuction is a paid mutator transaction binding the contract method 0x983b94fb.
//
// Solidity: function finalizeAuction(bytes32 _hash) returns()
func (_Contract *ContractTransactor) FinalizeAuction(opts *bind.TransactOpts, _hash [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "finalizeAuction", _hash)
}

// FinalizeAuction is a paid mutator transaction binding the contract method 0x983b94fb.
//
// Solidity: function finalizeAuction(bytes32 _hash) returns()
func (_Contract *ContractSession) FinalizeAuction(_hash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.FinalizeAuction(&_Contract.TransactOpts, _hash)
}

// FinalizeAuction is a paid mutator transaction binding the contract method 0x983b94fb.
//
// Solidity: function finalizeAuction(bytes32 _hash) returns()
func (_Contract *ContractTransactorSession) FinalizeAuction(_hash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.FinalizeAuction(&_Contract.TransactOpts, _hash)
}

// InvalidateName is a paid mutator transaction binding the contract method 0x15f73331.
//
// Solidity: function invalidateName(string unhashedName) returns()
func (_Contract *ContractTransactor) InvalidateName(opts *bind.TransactOpts, unhashedName string) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "invalidateName", unhashedName)
}

// InvalidateName is a paid mutator transaction binding the contract method 0x15f73331.
//
// Solidity: function invalidateName(string unhashedName) returns()
func (_Contract *ContractSession) InvalidateName(unhashedName string) (*types.Transaction, error) {
	return _Contract.Contract.InvalidateName(&_Contract.TransactOpts, unhashedName)
}

// InvalidateName is a paid mutator transaction binding the contract method 0x15f73331.
//
// Solidity: function invalidateName(string unhashedName) returns()
func (_Contract *ContractTransactorSession) InvalidateName(unhashedName string) (*types.Transaction, error) {
	return _Contract.Contract.InvalidateName(&_Contract.TransactOpts, unhashedName)
}

// NewBid is a paid mutator transaction binding the contract method 0xce92dced.
//
// Solidity: function newBid(bytes32 sealedBid) returns()
func (_Contract *ContractTransactor) NewBid(opts *bind.TransactOpts, sealedBid [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "newBid", sealedBid)
}

// NewBid is a paid mutator transaction binding the contract method 0xce92dced.
//
// Solidity: function newBid(bytes32 sealedBid) returns()
func (_Contract *ContractSession) NewBid(sealedBid [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.NewBid(&_Contract.TransactOpts, sealedBid)
}

// NewBid is a paid mutator transaction binding the contract method 0xce92dced.
//
// Solidity: function newBid(bytes32 sealedBid) returns()
func (_Contract *ContractTransactorSession) NewBid(sealedBid [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.NewBid(&_Contract.TransactOpts, sealedBid)
}

// ReleaseDeed is a paid mutator transaction binding the contract method 0x0230a07c.
//
// Solidity: function releaseDeed(bytes32 _hash) returns()
func (_Contract *ContractTransactor) ReleaseDeed(opts *bind.TransactOpts, _hash [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "releaseDeed", _hash)
}

// ReleaseDeed is a paid mutator transaction binding the contract method 0x0230a07c.
//
// Solidity: function releaseDeed(bytes32 _hash) returns()
func (_Contract *ContractSession) ReleaseDeed(_hash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.ReleaseDeed(&_Contract.TransactOpts, _hash)
}

// ReleaseDeed is a paid mutator transaction binding the contract method 0x0230a07c.
//
// Solidity: function releaseDeed(bytes32 _hash) returns()
func (_Contract *ContractTransactorSession) ReleaseDeed(_hash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.ReleaseDeed(&_Contract.TransactOpts, _hash)
}

// StartAuction is a paid mutator transaction binding the contract method 0xede8acdb.
//
// Solidity: function startAuction(bytes32 _hash) returns()
func (_Contract *ContractTransactor) StartAuction(opts *bind.TransactOpts, _hash [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "startAuction", _hash)
}

// StartAuction is a paid mutator transaction binding the contract method 0xede8acdb.
//
// Solidity: function startAuction(bytes32 _hash) returns()
func (_Contract *ContractSession) StartAuction(_hash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.StartAuction(&_Contract.TransactOpts, _hash)
}

// StartAuction is a paid mutator transaction binding the contract method 0xede8acdb.
//
// Solidity: function startAuction(bytes32 _hash) returns()
func (_Contract *ContractTransactorSession) StartAuction(_hash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.StartAuction(&_Contract.TransactOpts, _hash)
}

// StartAuctions is a paid mutator transaction binding the contract method 0xe27fe50f.
//
// Solidity: function startAuctions(bytes32[] _hashes) returns()
func (_Contract *ContractTransactor) StartAuctions(opts *bind.TransactOpts, _hashes [][32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "startAuctions", _hashes)
}

// StartAuctions is a paid mutator transaction binding the contract method 0xe27fe50f.
//
// Solidity: function startAuctions(bytes32[] _hashes) returns()
func (_Contract *ContractSession) StartAuctions(_hashes [][32]byte) (*types.Transaction, error) {
	return _Contract.Contract.StartAuctions(&_Contract.TransactOpts, _hashes)
}

// StartAuctions is a paid mutator transaction binding the contract method 0xe27fe50f.
//
// Solidity: function startAuctions(bytes32[] _hashes) returns()
func (_Contract *ContractTransactorSession) StartAuctions(_hashes [][32]byte) (*types.Transaction, error) {
	return _Contract.Contract.StartAuctions(&_Contract.TransactOpts, _hashes)
}

// StartAuctionsAndBid is a paid mutator transaction binding the contract method 0xfebefd61.
//
// Solidity: function startAuctionsAndBid(bytes32[] hashes, bytes32 sealedBid) returns()
func (_Contract *ContractTransactor) StartAuctionsAndBid(opts *bind.TransactOpts, hashes [][32]byte, sealedBid [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "startAuctionsAndBid", hashes, sealedBid)
}

// StartAuctionsAndBid is a paid mutator transaction binding the contract method 0xfebefd61.
//
// Solidity: function startAuctionsAndBid(bytes32[] hashes, bytes32 sealedBid) returns()
func (_Contract *ContractSession) StartAuctionsAndBid(hashes [][32]byte, sealedBid [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.StartAuctionsAndBid(&_Contract.TransactOpts, hashes, sealedBid)
}

// StartAuctionsAndBid is a paid mutator transaction binding the contract method 0xfebefd61.
//
// Solidity: function startAuctionsAndBid(bytes32[] hashes, bytes32 sealedBid) returns()
func (_Contract *ContractTransactorSession) StartAuctionsAndBid(hashes [][32]byte, sealedBid [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.StartAuctionsAndBid(&_Contract.TransactOpts, hashes, sealedBid)
}

// Transfer is a paid mutator transaction binding the contract method 0x79ce9fac.
//
// Solidity: function transfer(bytes32 _hash, address newOwner) returns()
func (_Contract *ContractTransactor) Transfer(opts *bind.TransactOpts, _hash [32]byte, newOwner common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "transfer", _hash, newOwner)
}

// Transfer is a paid mutator transaction binding the contract method 0x79ce9fac.
//
// Solidity: function transfer(bytes32 _hash, address newOwner) returns()
func (_Contract *ContractSession) Transfer(_hash [32]byte, newOwner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Transfer(&_Contract.TransactOpts, _hash, newOwner)
}

// Transfer is a paid mutator transaction binding the contract method 0x79ce9fac.
//
// Solidity: function transfer(bytes32 _hash, address newOwner) returns()
func (_Contract *ContractTransactorSession) Transfer(_hash [32]byte, newOwner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Transfer(&_Contract.TransactOpts, _hash, newOwner)
}

// TransferRegistrars is a paid mutator transaction binding the contract method 0x5ddae283.
//
// Solidity: function transferRegistrars(bytes32 _hash) returns()
func (_Contract *ContractTransactor) TransferRegistrars(opts *bind.TransactOpts, _hash [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "transferRegistrars", _hash)
}

// TransferRegistrars is a paid mutator transaction binding the contract method 0x5ddae283.
//
// Solidity: function transferRegistrars(bytes32 _hash) returns()
func (_Contract *ContractSession) TransferRegistrars(_hash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.TransferRegistrars(&_Contract.TransactOpts, _hash)
}

// TransferRegistrars is a paid mutator transaction binding the contract method 0x5ddae283.
//
// Solidity: function transferRegistrars(bytes32 _hash) returns()
func (_Contract *ContractTransactorSession) TransferRegistrars(_hash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.TransferRegistrars(&_Contract.TransactOpts, _hash)
}

// UnsealBid is a paid mutator transaction binding the contract method 0x47872b42.
//
// Solidity: function unsealBid(bytes32 _hash, uint256 _value, bytes32 _salt) returns()
func (_Contract *ContractTransactor) UnsealBid(opts *bind.TransactOpts, _hash [32]byte, _value *big.Int, _salt [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "unsealBid", _hash, _value, _salt)
}

// UnsealBid is a paid mutator transaction binding the contract method 0x47872b42.
//
// Solidity: function unsealBid(bytes32 _hash, uint256 _value, bytes32 _salt) returns()
func (_Contract *ContractSession) UnsealBid(_hash [32]byte, _value *big.Int, _salt [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.UnsealBid(&_Contract.TransactOpts, _hash, _value, _salt)
}

// UnsealBid is a paid mutator transaction binding the contract method 0x47872b42.
//
// Solidity: function unsealBid(bytes32 _hash, uint256 _value, bytes32 _salt) returns()
func (_Contract *ContractTransactorSession) UnsealBid(_hash [32]byte, _value *big.Int, _salt [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.UnsealBid(&_Contract.TransactOpts, _hash, _value, _salt)
}

// ContractAuctionStartedIterator is returned from FilterAuctionStarted and is used to iterate over the raw logs and unpacked data for AuctionStarted events raised by the Contract contract.
type ContractAuctionStartedIterator struct {
	Event *ContractAuctionStarted // Event containing the contract specifics and raw log

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
func (it *ContractAuctionStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractAuctionStarted)
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
		it.Event = new(ContractAuctionStarted)
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
func (it *ContractAuctionStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractAuctionStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractAuctionStarted represents a AuctionStarted event raised by the Contract contract.
type ContractAuctionStarted struct {
	Hash             [32]byte
	RegistrationDate *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterAuctionStarted is a free log retrieval operation binding the contract event 0x87e97e825a1d1fa0c54e1d36c7506c1dea8b1efd451fe68b000cf96f7cf40003.
//
// Solidity: event AuctionStarted(bytes32 indexed hash, uint256 registrationDate)
func (_Contract *ContractFilterer) FilterAuctionStarted(opts *bind.FilterOpts, hash [][32]byte) (*ContractAuctionStartedIterator, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "AuctionStarted", hashRule)
	if err != nil {
		return nil, err
	}
	return &ContractAuctionStartedIterator{contract: _Contract.contract, event: "AuctionStarted", logs: logs, sub: sub}, nil
}

// WatchAuctionStarted is a free log subscription operation binding the contract event 0x87e97e825a1d1fa0c54e1d36c7506c1dea8b1efd451fe68b000cf96f7cf40003.
//
// Solidity: event AuctionStarted(bytes32 indexed hash, uint256 registrationDate)
func (_Contract *ContractFilterer) WatchAuctionStarted(opts *bind.WatchOpts, sink chan<- *ContractAuctionStarted, hash [][32]byte) (event.Subscription, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "AuctionStarted", hashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractAuctionStarted)
				if err := _Contract.contract.UnpackLog(event, "AuctionStarted", log); err != nil {
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

// ParseAuctionStarted is a log parse operation binding the contract event 0x87e97e825a1d1fa0c54e1d36c7506c1dea8b1efd451fe68b000cf96f7cf40003.
//
// Solidity: event AuctionStarted(bytes32 indexed hash, uint256 registrationDate)
func (_Contract *ContractFilterer) ParseAuctionStarted(log types.Log) (*ContractAuctionStarted, error) {
	event := new(ContractAuctionStarted)
	if err := _Contract.contract.UnpackLog(event, "AuctionStarted", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractBidRevealedIterator is returned from FilterBidRevealed and is used to iterate over the raw logs and unpacked data for BidRevealed events raised by the Contract contract.
type ContractBidRevealedIterator struct {
	Event *ContractBidRevealed // Event containing the contract specifics and raw log

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
func (it *ContractBidRevealedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractBidRevealed)
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
		it.Event = new(ContractBidRevealed)
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
func (it *ContractBidRevealedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractBidRevealedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractBidRevealed represents a BidRevealed event raised by the Contract contract.
type ContractBidRevealed struct {
	Hash   [32]byte
	Owner  common.Address
	Value  *big.Int
	Status uint8
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBidRevealed is a free log retrieval operation binding the contract event 0x7b6c4b278d165a6b33958f8ea5dfb00c8c9d4d0acf1985bef5d10786898bc3e7.
//
// Solidity: event BidRevealed(bytes32 indexed hash, address indexed owner, uint256 value, uint8 status)
func (_Contract *ContractFilterer) FilterBidRevealed(opts *bind.FilterOpts, hash [][32]byte, owner []common.Address) (*ContractBidRevealedIterator, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "BidRevealed", hashRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &ContractBidRevealedIterator{contract: _Contract.contract, event: "BidRevealed", logs: logs, sub: sub}, nil
}

// WatchBidRevealed is a free log subscription operation binding the contract event 0x7b6c4b278d165a6b33958f8ea5dfb00c8c9d4d0acf1985bef5d10786898bc3e7.
//
// Solidity: event BidRevealed(bytes32 indexed hash, address indexed owner, uint256 value, uint8 status)
func (_Contract *ContractFilterer) WatchBidRevealed(opts *bind.WatchOpts, sink chan<- *ContractBidRevealed, hash [][32]byte, owner []common.Address) (event.Subscription, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "BidRevealed", hashRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractBidRevealed)
				if err := _Contract.contract.UnpackLog(event, "BidRevealed", log); err != nil {
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

// ParseBidRevealed is a log parse operation binding the contract event 0x7b6c4b278d165a6b33958f8ea5dfb00c8c9d4d0acf1985bef5d10786898bc3e7.
//
// Solidity: event BidRevealed(bytes32 indexed hash, address indexed owner, uint256 value, uint8 status)
func (_Contract *ContractFilterer) ParseBidRevealed(log types.Log) (*ContractBidRevealed, error) {
	event := new(ContractBidRevealed)
	if err := _Contract.contract.UnpackLog(event, "BidRevealed", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractHashInvalidatedIterator is returned from FilterHashInvalidated and is used to iterate over the raw logs and unpacked data for HashInvalidated events raised by the Contract contract.
type ContractHashInvalidatedIterator struct {
	Event *ContractHashInvalidated // Event containing the contract specifics and raw log

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
func (it *ContractHashInvalidatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractHashInvalidated)
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
		it.Event = new(ContractHashInvalidated)
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
func (it *ContractHashInvalidatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractHashInvalidatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractHashInvalidated represents a HashInvalidated event raised by the Contract contract.
type ContractHashInvalidated struct {
	Hash             [32]byte
	Name             common.Hash
	Value            *big.Int
	RegistrationDate *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterHashInvalidated is a free log retrieval operation binding the contract event 0x1f9c649fe47e58bb60f4e52f0d90e4c47a526c9f90c5113df842c025970b66ad.
//
// Solidity: event HashInvalidated(bytes32 indexed hash, string indexed name, uint256 value, uint256 registrationDate)
func (_Contract *ContractFilterer) FilterHashInvalidated(opts *bind.FilterOpts, hash [][32]byte, name []string) (*ContractHashInvalidatedIterator, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}
	var nameRule []interface{}
	for _, nameItem := range name {
		nameRule = append(nameRule, nameItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "HashInvalidated", hashRule, nameRule)
	if err != nil {
		return nil, err
	}
	return &ContractHashInvalidatedIterator{contract: _Contract.contract, event: "HashInvalidated", logs: logs, sub: sub}, nil
}

// WatchHashInvalidated is a free log subscription operation binding the contract event 0x1f9c649fe47e58bb60f4e52f0d90e4c47a526c9f90c5113df842c025970b66ad.
//
// Solidity: event HashInvalidated(bytes32 indexed hash, string indexed name, uint256 value, uint256 registrationDate)
func (_Contract *ContractFilterer) WatchHashInvalidated(opts *bind.WatchOpts, sink chan<- *ContractHashInvalidated, hash [][32]byte, name []string) (event.Subscription, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}
	var nameRule []interface{}
	for _, nameItem := range name {
		nameRule = append(nameRule, nameItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "HashInvalidated", hashRule, nameRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractHashInvalidated)
				if err := _Contract.contract.UnpackLog(event, "HashInvalidated", log); err != nil {
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

// ParseHashInvalidated is a log parse operation binding the contract event 0x1f9c649fe47e58bb60f4e52f0d90e4c47a526c9f90c5113df842c025970b66ad.
//
// Solidity: event HashInvalidated(bytes32 indexed hash, string indexed name, uint256 value, uint256 registrationDate)
func (_Contract *ContractFilterer) ParseHashInvalidated(log types.Log) (*ContractHashInvalidated, error) {
	event := new(ContractHashInvalidated)
	if err := _Contract.contract.UnpackLog(event, "HashInvalidated", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractHashRegisteredIterator is returned from FilterHashRegistered and is used to iterate over the raw logs and unpacked data for HashRegistered events raised by the Contract contract.
type ContractHashRegisteredIterator struct {
	Event *ContractHashRegistered // Event containing the contract specifics and raw log

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
func (it *ContractHashRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractHashRegistered)
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
		it.Event = new(ContractHashRegistered)
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
func (it *ContractHashRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractHashRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractHashRegistered represents a HashRegistered event raised by the Contract contract.
type ContractHashRegistered struct {
	Hash             [32]byte
	Owner            common.Address
	Value            *big.Int
	RegistrationDate *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterHashRegistered is a free log retrieval operation binding the contract event 0x0f0c27adfd84b60b6f456b0e87cdccb1e5fb9603991588d87fa99f5b6b61e670.
//
// Solidity: event HashRegistered(bytes32 indexed hash, address indexed owner, uint256 value, uint256 registrationDate)
func (_Contract *ContractFilterer) FilterHashRegistered(opts *bind.FilterOpts, hash [][32]byte, owner []common.Address) (*ContractHashRegisteredIterator, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "HashRegistered", hashRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &ContractHashRegisteredIterator{contract: _Contract.contract, event: "HashRegistered", logs: logs, sub: sub}, nil
}

// WatchHashRegistered is a free log subscription operation binding the contract event 0x0f0c27adfd84b60b6f456b0e87cdccb1e5fb9603991588d87fa99f5b6b61e670.
//
// Solidity: event HashRegistered(bytes32 indexed hash, address indexed owner, uint256 value, uint256 registrationDate)
func (_Contract *ContractFilterer) WatchHashRegistered(opts *bind.WatchOpts, sink chan<- *ContractHashRegistered, hash [][32]byte, owner []common.Address) (event.Subscription, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "HashRegistered", hashRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractHashRegistered)
				if err := _Contract.contract.UnpackLog(event, "HashRegistered", log); err != nil {
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

// ParseHashRegistered is a log parse operation binding the contract event 0x0f0c27adfd84b60b6f456b0e87cdccb1e5fb9603991588d87fa99f5b6b61e670.
//
// Solidity: event HashRegistered(bytes32 indexed hash, address indexed owner, uint256 value, uint256 registrationDate)
func (_Contract *ContractFilterer) ParseHashRegistered(log types.Log) (*ContractHashRegistered, error) {
	event := new(ContractHashRegistered)
	if err := _Contract.contract.UnpackLog(event, "HashRegistered", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractHashReleasedIterator is returned from FilterHashReleased and is used to iterate over the raw logs and unpacked data for HashReleased events raised by the Contract contract.
type ContractHashReleasedIterator struct {
	Event *ContractHashReleased // Event containing the contract specifics and raw log

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
func (it *ContractHashReleasedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractHashReleased)
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
		it.Event = new(ContractHashReleased)
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
func (it *ContractHashReleasedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractHashReleasedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractHashReleased represents a HashReleased event raised by the Contract contract.
type ContractHashReleased struct {
	Hash  [32]byte
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterHashReleased is a free log retrieval operation binding the contract event 0x292b79b9246fa2c8e77d3fe195b251f9cb839d7d038e667c069ee7708c631e16.
//
// Solidity: event HashReleased(bytes32 indexed hash, uint256 value)
func (_Contract *ContractFilterer) FilterHashReleased(opts *bind.FilterOpts, hash [][32]byte) (*ContractHashReleasedIterator, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "HashReleased", hashRule)
	if err != nil {
		return nil, err
	}
	return &ContractHashReleasedIterator{contract: _Contract.contract, event: "HashReleased", logs: logs, sub: sub}, nil
}

// WatchHashReleased is a free log subscription operation binding the contract event 0x292b79b9246fa2c8e77d3fe195b251f9cb839d7d038e667c069ee7708c631e16.
//
// Solidity: event HashReleased(bytes32 indexed hash, uint256 value)
func (_Contract *ContractFilterer) WatchHashReleased(opts *bind.WatchOpts, sink chan<- *ContractHashReleased, hash [][32]byte) (event.Subscription, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "HashReleased", hashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractHashReleased)
				if err := _Contract.contract.UnpackLog(event, "HashReleased", log); err != nil {
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

// ParseHashReleased is a log parse operation binding the contract event 0x292b79b9246fa2c8e77d3fe195b251f9cb839d7d038e667c069ee7708c631e16.
//
// Solidity: event HashReleased(bytes32 indexed hash, uint256 value)
func (_Contract *ContractFilterer) ParseHashReleased(log types.Log) (*ContractHashReleased, error) {
	event := new(ContractHashReleased)
	if err := _Contract.contract.UnpackLog(event, "HashReleased", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractNewBidIterator is returned from FilterNewBid and is used to iterate over the raw logs and unpacked data for NewBid events raised by the Contract contract.
type ContractNewBidIterator struct {
	Event *ContractNewBid // Event containing the contract specifics and raw log

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
func (it *ContractNewBidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractNewBid)
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
		it.Event = new(ContractNewBid)
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
func (it *ContractNewBidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractNewBidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractNewBid represents a NewBid event raised by the Contract contract.
type ContractNewBid struct {
	Hash    [32]byte
	Bidder  common.Address
	Deposit *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNewBid is a free log retrieval operation binding the contract event 0xb556ff269c1b6714f432c36431e2041d28436a73b6c3f19c021827bbdc6bfc29.
//
// Solidity: event NewBid(bytes32 indexed hash, address indexed bidder, uint256 deposit)
func (_Contract *ContractFilterer) FilterNewBid(opts *bind.FilterOpts, hash [][32]byte, bidder []common.Address) (*ContractNewBidIterator, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}
	var bidderRule []interface{}
	for _, bidderItem := range bidder {
		bidderRule = append(bidderRule, bidderItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "NewBid", hashRule, bidderRule)
	if err != nil {
		return nil, err
	}
	return &ContractNewBidIterator{contract: _Contract.contract, event: "NewBid", logs: logs, sub: sub}, nil
}

// WatchNewBid is a free log subscription operation binding the contract event 0xb556ff269c1b6714f432c36431e2041d28436a73b6c3f19c021827bbdc6bfc29.
//
// Solidity: event NewBid(bytes32 indexed hash, address indexed bidder, uint256 deposit)
func (_Contract *ContractFilterer) WatchNewBid(opts *bind.WatchOpts, sink chan<- *ContractNewBid, hash [][32]byte, bidder []common.Address) (event.Subscription, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}
	var bidderRule []interface{}
	for _, bidderItem := range bidder {
		bidderRule = append(bidderRule, bidderItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "NewBid", hashRule, bidderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractNewBid)
				if err := _Contract.contract.UnpackLog(event, "NewBid", log); err != nil {
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

// ParseNewBid is a log parse operation binding the contract event 0xb556ff269c1b6714f432c36431e2041d28436a73b6c3f19c021827bbdc6bfc29.
//
// Solidity: event NewBid(bytes32 indexed hash, address indexed bidder, uint256 deposit)
func (_Contract *ContractFilterer) ParseNewBid(log types.Log) (*ContractNewBid, error) {
	event := new(ContractNewBid)
	if err := _Contract.contract.UnpackLog(event, "NewBid", log); err != nil {
		return nil, err
	}
	return event, nil
}
