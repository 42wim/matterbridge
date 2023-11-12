// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package directory

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

// DirectoryMetaData contains all meta data concerning the Directory contract.
var DirectoryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_votingContract\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_featuredVotingContract\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"community\",\"type\":\"bytes\"}],\"name\":\"addCommunity\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"featuredVotingContract\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCommunities\",\"outputs\":[{\"internalType\":\"bytes[]\",\"name\":\"\",\"type\":\"bytes[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getFeaturedCommunities\",\"outputs\":[{\"internalType\":\"bytes[]\",\"name\":\"\",\"type\":\"bytes[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"community\",\"type\":\"bytes\"}],\"name\":\"isCommunityFeatured\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"community\",\"type\":\"bytes\"}],\"name\":\"isCommunityInDirectory\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"community\",\"type\":\"bytes\"}],\"name\":\"removeCommunity\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes[]\",\"name\":\"_featuredCommunities\",\"type\":\"bytes[]\"}],\"name\":\"setFeaturedCommunities\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"votingContract\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// DirectoryABI is the input ABI used to generate the binding from.
// Deprecated: Use DirectoryMetaData.ABI instead.
var DirectoryABI = DirectoryMetaData.ABI

// Directory is an auto generated Go binding around an Ethereum contract.
type Directory struct {
	DirectoryCaller     // Read-only binding to the contract
	DirectoryTransactor // Write-only binding to the contract
	DirectoryFilterer   // Log filterer for contract events
}

// DirectoryCaller is an auto generated read-only Go binding around an Ethereum contract.
type DirectoryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DirectoryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DirectoryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DirectoryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DirectoryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DirectorySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DirectorySession struct {
	Contract     *Directory        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DirectoryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DirectoryCallerSession struct {
	Contract *DirectoryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// DirectoryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DirectoryTransactorSession struct {
	Contract     *DirectoryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// DirectoryRaw is an auto generated low-level Go binding around an Ethereum contract.
type DirectoryRaw struct {
	Contract *Directory // Generic contract binding to access the raw methods on
}

// DirectoryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DirectoryCallerRaw struct {
	Contract *DirectoryCaller // Generic read-only contract binding to access the raw methods on
}

// DirectoryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DirectoryTransactorRaw struct {
	Contract *DirectoryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDirectory creates a new instance of Directory, bound to a specific deployed contract.
func NewDirectory(address common.Address, backend bind.ContractBackend) (*Directory, error) {
	contract, err := bindDirectory(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Directory{DirectoryCaller: DirectoryCaller{contract: contract}, DirectoryTransactor: DirectoryTransactor{contract: contract}, DirectoryFilterer: DirectoryFilterer{contract: contract}}, nil
}

// NewDirectoryCaller creates a new read-only instance of Directory, bound to a specific deployed contract.
func NewDirectoryCaller(address common.Address, caller bind.ContractCaller) (*DirectoryCaller, error) {
	contract, err := bindDirectory(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DirectoryCaller{contract: contract}, nil
}

// NewDirectoryTransactor creates a new write-only instance of Directory, bound to a specific deployed contract.
func NewDirectoryTransactor(address common.Address, transactor bind.ContractTransactor) (*DirectoryTransactor, error) {
	contract, err := bindDirectory(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DirectoryTransactor{contract: contract}, nil
}

// NewDirectoryFilterer creates a new log filterer instance of Directory, bound to a specific deployed contract.
func NewDirectoryFilterer(address common.Address, filterer bind.ContractFilterer) (*DirectoryFilterer, error) {
	contract, err := bindDirectory(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DirectoryFilterer{contract: contract}, nil
}

// bindDirectory binds a generic wrapper to an already deployed contract.
func bindDirectory(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := DirectoryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Directory *DirectoryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Directory.Contract.DirectoryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Directory *DirectoryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Directory.Contract.DirectoryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Directory *DirectoryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Directory.Contract.DirectoryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Directory *DirectoryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Directory.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Directory *DirectoryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Directory.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Directory *DirectoryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Directory.Contract.contract.Transact(opts, method, params...)
}

// FeaturedVotingContract is a free data retrieval call binding the contract method 0x7475fe93.
//
// Solidity: function featuredVotingContract() view returns(address)
func (_Directory *DirectoryCaller) FeaturedVotingContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Directory.contract.Call(opts, &out, "featuredVotingContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// FeaturedVotingContract is a free data retrieval call binding the contract method 0x7475fe93.
//
// Solidity: function featuredVotingContract() view returns(address)
func (_Directory *DirectorySession) FeaturedVotingContract() (common.Address, error) {
	return _Directory.Contract.FeaturedVotingContract(&_Directory.CallOpts)
}

// FeaturedVotingContract is a free data retrieval call binding the contract method 0x7475fe93.
//
// Solidity: function featuredVotingContract() view returns(address)
func (_Directory *DirectoryCallerSession) FeaturedVotingContract() (common.Address, error) {
	return _Directory.Contract.FeaturedVotingContract(&_Directory.CallOpts)
}

// GetCommunities is a free data retrieval call binding the contract method 0xc251b565.
//
// Solidity: function getCommunities() view returns(bytes[])
func (_Directory *DirectoryCaller) GetCommunities(opts *bind.CallOpts) ([][]byte, error) {
	var out []interface{}
	err := _Directory.contract.Call(opts, &out, "getCommunities")

	if err != nil {
		return *new([][]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][]byte)).(*[][]byte)

	return out0, err

}

// GetCommunities is a free data retrieval call binding the contract method 0xc251b565.
//
// Solidity: function getCommunities() view returns(bytes[])
func (_Directory *DirectorySession) GetCommunities() ([][]byte, error) {
	return _Directory.Contract.GetCommunities(&_Directory.CallOpts)
}

// GetCommunities is a free data retrieval call binding the contract method 0xc251b565.
//
// Solidity: function getCommunities() view returns(bytes[])
func (_Directory *DirectoryCallerSession) GetCommunities() ([][]byte, error) {
	return _Directory.Contract.GetCommunities(&_Directory.CallOpts)
}

// GetFeaturedCommunities is a free data retrieval call binding the contract method 0x967961c6.
//
// Solidity: function getFeaturedCommunities() view returns(bytes[])
func (_Directory *DirectoryCaller) GetFeaturedCommunities(opts *bind.CallOpts) ([][]byte, error) {
	var out []interface{}
	err := _Directory.contract.Call(opts, &out, "getFeaturedCommunities")

	if err != nil {
		return *new([][]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][]byte)).(*[][]byte)

	return out0, err

}

// GetFeaturedCommunities is a free data retrieval call binding the contract method 0x967961c6.
//
// Solidity: function getFeaturedCommunities() view returns(bytes[])
func (_Directory *DirectorySession) GetFeaturedCommunities() ([][]byte, error) {
	return _Directory.Contract.GetFeaturedCommunities(&_Directory.CallOpts)
}

// GetFeaturedCommunities is a free data retrieval call binding the contract method 0x967961c6.
//
// Solidity: function getFeaturedCommunities() view returns(bytes[])
func (_Directory *DirectoryCallerSession) GetFeaturedCommunities() ([][]byte, error) {
	return _Directory.Contract.GetFeaturedCommunities(&_Directory.CallOpts)
}

// IsCommunityFeatured is a free data retrieval call binding the contract method 0xf6a18e62.
//
// Solidity: function isCommunityFeatured(bytes community) view returns(bool)
func (_Directory *DirectoryCaller) IsCommunityFeatured(opts *bind.CallOpts, community []byte) (bool, error) {
	var out []interface{}
	err := _Directory.contract.Call(opts, &out, "isCommunityFeatured", community)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsCommunityFeatured is a free data retrieval call binding the contract method 0xf6a18e62.
//
// Solidity: function isCommunityFeatured(bytes community) view returns(bool)
func (_Directory *DirectorySession) IsCommunityFeatured(community []byte) (bool, error) {
	return _Directory.Contract.IsCommunityFeatured(&_Directory.CallOpts, community)
}

// IsCommunityFeatured is a free data retrieval call binding the contract method 0xf6a18e62.
//
// Solidity: function isCommunityFeatured(bytes community) view returns(bool)
func (_Directory *DirectoryCallerSession) IsCommunityFeatured(community []byte) (bool, error) {
	return _Directory.Contract.IsCommunityFeatured(&_Directory.CallOpts, community)
}

// IsCommunityInDirectory is a free data retrieval call binding the contract method 0xb3dbb52a.
//
// Solidity: function isCommunityInDirectory(bytes community) view returns(bool)
func (_Directory *DirectoryCaller) IsCommunityInDirectory(opts *bind.CallOpts, community []byte) (bool, error) {
	var out []interface{}
	err := _Directory.contract.Call(opts, &out, "isCommunityInDirectory", community)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsCommunityInDirectory is a free data retrieval call binding the contract method 0xb3dbb52a.
//
// Solidity: function isCommunityInDirectory(bytes community) view returns(bool)
func (_Directory *DirectorySession) IsCommunityInDirectory(community []byte) (bool, error) {
	return _Directory.Contract.IsCommunityInDirectory(&_Directory.CallOpts, community)
}

// IsCommunityInDirectory is a free data retrieval call binding the contract method 0xb3dbb52a.
//
// Solidity: function isCommunityInDirectory(bytes community) view returns(bool)
func (_Directory *DirectoryCallerSession) IsCommunityInDirectory(community []byte) (bool, error) {
	return _Directory.Contract.IsCommunityInDirectory(&_Directory.CallOpts, community)
}

// VotingContract is a free data retrieval call binding the contract method 0xc1fc006a.
//
// Solidity: function votingContract() view returns(address)
func (_Directory *DirectoryCaller) VotingContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Directory.contract.Call(opts, &out, "votingContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// VotingContract is a free data retrieval call binding the contract method 0xc1fc006a.
//
// Solidity: function votingContract() view returns(address)
func (_Directory *DirectorySession) VotingContract() (common.Address, error) {
	return _Directory.Contract.VotingContract(&_Directory.CallOpts)
}

// VotingContract is a free data retrieval call binding the contract method 0xc1fc006a.
//
// Solidity: function votingContract() view returns(address)
func (_Directory *DirectoryCallerSession) VotingContract() (common.Address, error) {
	return _Directory.Contract.VotingContract(&_Directory.CallOpts)
}

// AddCommunity is a paid mutator transaction binding the contract method 0x74837935.
//
// Solidity: function addCommunity(bytes community) returns()
func (_Directory *DirectoryTransactor) AddCommunity(opts *bind.TransactOpts, community []byte) (*types.Transaction, error) {
	return _Directory.contract.Transact(opts, "addCommunity", community)
}

// AddCommunity is a paid mutator transaction binding the contract method 0x74837935.
//
// Solidity: function addCommunity(bytes community) returns()
func (_Directory *DirectorySession) AddCommunity(community []byte) (*types.Transaction, error) {
	return _Directory.Contract.AddCommunity(&_Directory.TransactOpts, community)
}

// AddCommunity is a paid mutator transaction binding the contract method 0x74837935.
//
// Solidity: function addCommunity(bytes community) returns()
func (_Directory *DirectoryTransactorSession) AddCommunity(community []byte) (*types.Transaction, error) {
	return _Directory.Contract.AddCommunity(&_Directory.TransactOpts, community)
}

// RemoveCommunity is a paid mutator transaction binding the contract method 0x3c01b93c.
//
// Solidity: function removeCommunity(bytes community) returns()
func (_Directory *DirectoryTransactor) RemoveCommunity(opts *bind.TransactOpts, community []byte) (*types.Transaction, error) {
	return _Directory.contract.Transact(opts, "removeCommunity", community)
}

// RemoveCommunity is a paid mutator transaction binding the contract method 0x3c01b93c.
//
// Solidity: function removeCommunity(bytes community) returns()
func (_Directory *DirectorySession) RemoveCommunity(community []byte) (*types.Transaction, error) {
	return _Directory.Contract.RemoveCommunity(&_Directory.TransactOpts, community)
}

// RemoveCommunity is a paid mutator transaction binding the contract method 0x3c01b93c.
//
// Solidity: function removeCommunity(bytes community) returns()
func (_Directory *DirectoryTransactorSession) RemoveCommunity(community []byte) (*types.Transaction, error) {
	return _Directory.Contract.RemoveCommunity(&_Directory.TransactOpts, community)
}

// SetFeaturedCommunities is a paid mutator transaction binding the contract method 0xd62879f1.
//
// Solidity: function setFeaturedCommunities(bytes[] _featuredCommunities) returns()
func (_Directory *DirectoryTransactor) SetFeaturedCommunities(opts *bind.TransactOpts, _featuredCommunities [][]byte) (*types.Transaction, error) {
	return _Directory.contract.Transact(opts, "setFeaturedCommunities", _featuredCommunities)
}

// SetFeaturedCommunities is a paid mutator transaction binding the contract method 0xd62879f1.
//
// Solidity: function setFeaturedCommunities(bytes[] _featuredCommunities) returns()
func (_Directory *DirectorySession) SetFeaturedCommunities(_featuredCommunities [][]byte) (*types.Transaction, error) {
	return _Directory.Contract.SetFeaturedCommunities(&_Directory.TransactOpts, _featuredCommunities)
}

// SetFeaturedCommunities is a paid mutator transaction binding the contract method 0xd62879f1.
//
// Solidity: function setFeaturedCommunities(bytes[] _featuredCommunities) returns()
func (_Directory *DirectoryTransactorSession) SetFeaturedCommunities(_featuredCommunities [][]byte) (*types.Transaction, error) {
	return _Directory.Contract.SetFeaturedCommunities(&_Directory.TransactOpts, _featuredCommunities)
}
