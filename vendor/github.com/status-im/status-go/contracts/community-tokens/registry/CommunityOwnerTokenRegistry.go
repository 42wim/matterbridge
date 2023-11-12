// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package communityownertokenregistry

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

// CommunityOwnerTokenRegistryMetaData contains all meta data concerning the CommunityOwnerTokenRegistry contract.
var CommunityOwnerTokenRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"CommunityOwnerTokenRegistry_EntryAlreadyExists\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CommunityOwnerTokenRegistry_InvalidAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CommunityOwnerTokenRegistry_NotAuthorized\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"AddEntry\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferStarted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"TokenDeployerAddressChange\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"acceptOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_communityAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_tokenAddress\",\"type\":\"address\"}],\"name\":\"addEntry\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"communityAddressToTokenAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_communityAddress\",\"type\":\"address\"}],\"name\":\"getEntry\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_tokenDeployer\",\"type\":\"address\"}],\"name\":\"setCommunityTokenDeployerAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tokenDeployer\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b5061001a3361001f565b610096565b600180546001600160a01b031916905561004381610046602090811b6105de17901c565b50565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b610790806100a56000396000f3fe608060405234801561001057600080fd5b50600436106100be5760003560e01c8063a7a9584011610076578063d1f7c48a1161005b578063d1f7c48a146101be578063e30c3978146101d1578063f2fde38b146101ef57600080fd5b8063a7a9584014610175578063b97e6ab91461018857600080fd5b806379ba5097116100a757806379ba5097146101165780637db6a4e41461011e5780638da5cb5b1461015757600080fd5b80632a2dae0a146100c3578063715018a61461010c575b600080fd5b6002546100e39073ffffffffffffffffffffffffffffffffffffffff1681565b60405173ffffffffffffffffffffffffffffffffffffffff909116815260200160405180910390f35b610114610202565b005b610114610216565b6100e361012c36600461072e565b73ffffffffffffffffffffffffffffffffffffffff9081166000908152600360205260409020541690565b60005473ffffffffffffffffffffffffffffffffffffffff166100e3565b610114610183366004610750565b6102d0565b6100e361019636600461072e565b60036020526000908152604090205473ffffffffffffffffffffffffffffffffffffffff1681565b6101146101cc36600461072e565b61046a565b60015473ffffffffffffffffffffffffffffffffffffffff166100e3565b6101146101fd36600461072e565b61052e565b61020a610653565b61021460006106d4565b565b600154339073ffffffffffffffffffffffffffffffffffffffff1681146102c4576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602960248201527f4f776e61626c6532537465703a2063616c6c6572206973206e6f74207468652060448201527f6e6577206f776e6572000000000000000000000000000000000000000000000060648201526084015b60405180910390fd5b6102cd816106d4565b50565b60025473ffffffffffffffffffffffffffffffffffffffff163314610321576040517f6a60770200000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b73ffffffffffffffffffffffffffffffffffffffff8281166000908152600360205260409020541615610380576040517fec22bbb900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b73ffffffffffffffffffffffffffffffffffffffff821615806103b7575073ffffffffffffffffffffffffffffffffffffffff8116155b156103ee576040517f911f6bec00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b73ffffffffffffffffffffffffffffffffffffffff82811660008181526003602052604080822080547fffffffffffffffffffffffff0000000000000000000000000000000000000000169486169485179055517f4bc4774424bf8749f142d7c1df17ee73cf36394616f38dd6799e99ea3bb4763a9190a35050565b610472610653565b73ffffffffffffffffffffffffffffffffffffffff81166104bf576040517f911f6bec00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b600280547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff83169081179091556040517f057829294de8b35baa4c034fc338afc6ecb2eb9b3035615c44100d80ecc93db790600090a250565b610536610653565b6001805473ffffffffffffffffffffffffffffffffffffffff83167fffffffffffffffffffffffff0000000000000000000000000000000000000000909116811790915561059960005473ffffffffffffffffffffffffffffffffffffffff1690565b73ffffffffffffffffffffffffffffffffffffffff167f38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e2270060405160405180910390a350565b6000805473ffffffffffffffffffffffffffffffffffffffff8381167fffffffffffffffffffffffff0000000000000000000000000000000000000000831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b60005473ffffffffffffffffffffffffffffffffffffffff163314610214576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e657260448201526064016102bb565b600180547fffffffffffffffffffffffff00000000000000000000000000000000000000001690556102cd816105de565b803573ffffffffffffffffffffffffffffffffffffffff8116811461072957600080fd5b919050565b60006020828403121561074057600080fd5b61074982610705565b9392505050565b6000806040838503121561076357600080fd5b61076c83610705565b915061077a60208401610705565b9050925092905056fea164736f6c6343000811000a",
}

// CommunityOwnerTokenRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use CommunityOwnerTokenRegistryMetaData.ABI instead.
var CommunityOwnerTokenRegistryABI = CommunityOwnerTokenRegistryMetaData.ABI

// CommunityOwnerTokenRegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use CommunityOwnerTokenRegistryMetaData.Bin instead.
var CommunityOwnerTokenRegistryBin = CommunityOwnerTokenRegistryMetaData.Bin

// DeployCommunityOwnerTokenRegistry deploys a new Ethereum contract, binding an instance of CommunityOwnerTokenRegistry to it.
func DeployCommunityOwnerTokenRegistry(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *CommunityOwnerTokenRegistry, error) {
	parsed, err := CommunityOwnerTokenRegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(CommunityOwnerTokenRegistryBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &CommunityOwnerTokenRegistry{CommunityOwnerTokenRegistryCaller: CommunityOwnerTokenRegistryCaller{contract: contract}, CommunityOwnerTokenRegistryTransactor: CommunityOwnerTokenRegistryTransactor{contract: contract}, CommunityOwnerTokenRegistryFilterer: CommunityOwnerTokenRegistryFilterer{contract: contract}}, nil
}

// CommunityOwnerTokenRegistry is an auto generated Go binding around an Ethereum contract.
type CommunityOwnerTokenRegistry struct {
	CommunityOwnerTokenRegistryCaller     // Read-only binding to the contract
	CommunityOwnerTokenRegistryTransactor // Write-only binding to the contract
	CommunityOwnerTokenRegistryFilterer   // Log filterer for contract events
}

// CommunityOwnerTokenRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type CommunityOwnerTokenRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CommunityOwnerTokenRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type CommunityOwnerTokenRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CommunityOwnerTokenRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type CommunityOwnerTokenRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CommunityOwnerTokenRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type CommunityOwnerTokenRegistrySession struct {
	Contract     *CommunityOwnerTokenRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                // Call options to use throughout this session
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// CommunityOwnerTokenRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type CommunityOwnerTokenRegistryCallerSession struct {
	Contract *CommunityOwnerTokenRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                      // Call options to use throughout this session
}

// CommunityOwnerTokenRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type CommunityOwnerTokenRegistryTransactorSession struct {
	Contract     *CommunityOwnerTokenRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                      // Transaction auth options to use throughout this session
}

// CommunityOwnerTokenRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type CommunityOwnerTokenRegistryRaw struct {
	Contract *CommunityOwnerTokenRegistry // Generic contract binding to access the raw methods on
}

// CommunityOwnerTokenRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type CommunityOwnerTokenRegistryCallerRaw struct {
	Contract *CommunityOwnerTokenRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// CommunityOwnerTokenRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type CommunityOwnerTokenRegistryTransactorRaw struct {
	Contract *CommunityOwnerTokenRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCommunityOwnerTokenRegistry creates a new instance of CommunityOwnerTokenRegistry, bound to a specific deployed contract.
func NewCommunityOwnerTokenRegistry(address common.Address, backend bind.ContractBackend) (*CommunityOwnerTokenRegistry, error) {
	contract, err := bindCommunityOwnerTokenRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &CommunityOwnerTokenRegistry{CommunityOwnerTokenRegistryCaller: CommunityOwnerTokenRegistryCaller{contract: contract}, CommunityOwnerTokenRegistryTransactor: CommunityOwnerTokenRegistryTransactor{contract: contract}, CommunityOwnerTokenRegistryFilterer: CommunityOwnerTokenRegistryFilterer{contract: contract}}, nil
}

// NewCommunityOwnerTokenRegistryCaller creates a new read-only instance of CommunityOwnerTokenRegistry, bound to a specific deployed contract.
func NewCommunityOwnerTokenRegistryCaller(address common.Address, caller bind.ContractCaller) (*CommunityOwnerTokenRegistryCaller, error) {
	contract, err := bindCommunityOwnerTokenRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CommunityOwnerTokenRegistryCaller{contract: contract}, nil
}

// NewCommunityOwnerTokenRegistryTransactor creates a new write-only instance of CommunityOwnerTokenRegistry, bound to a specific deployed contract.
func NewCommunityOwnerTokenRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*CommunityOwnerTokenRegistryTransactor, error) {
	contract, err := bindCommunityOwnerTokenRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CommunityOwnerTokenRegistryTransactor{contract: contract}, nil
}

// NewCommunityOwnerTokenRegistryFilterer creates a new log filterer instance of CommunityOwnerTokenRegistry, bound to a specific deployed contract.
func NewCommunityOwnerTokenRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*CommunityOwnerTokenRegistryFilterer, error) {
	contract, err := bindCommunityOwnerTokenRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CommunityOwnerTokenRegistryFilterer{contract: contract}, nil
}

// bindCommunityOwnerTokenRegistry binds a generic wrapper to an already deployed contract.
func bindCommunityOwnerTokenRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := CommunityOwnerTokenRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CommunityOwnerTokenRegistry.Contract.CommunityOwnerTokenRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.Contract.CommunityOwnerTokenRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.Contract.CommunityOwnerTokenRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CommunityOwnerTokenRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.Contract.contract.Transact(opts, method, params...)
}

// CommunityAddressToTokenAddress is a free data retrieval call binding the contract method 0xb97e6ab9.
//
// Solidity: function communityAddressToTokenAddress(address ) view returns(address)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryCaller) CommunityAddressToTokenAddress(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var out []interface{}
	err := _CommunityOwnerTokenRegistry.contract.Call(opts, &out, "communityAddressToTokenAddress", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CommunityAddressToTokenAddress is a free data retrieval call binding the contract method 0xb97e6ab9.
//
// Solidity: function communityAddressToTokenAddress(address ) view returns(address)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistrySession) CommunityAddressToTokenAddress(arg0 common.Address) (common.Address, error) {
	return _CommunityOwnerTokenRegistry.Contract.CommunityAddressToTokenAddress(&_CommunityOwnerTokenRegistry.CallOpts, arg0)
}

// CommunityAddressToTokenAddress is a free data retrieval call binding the contract method 0xb97e6ab9.
//
// Solidity: function communityAddressToTokenAddress(address ) view returns(address)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryCallerSession) CommunityAddressToTokenAddress(arg0 common.Address) (common.Address, error) {
	return _CommunityOwnerTokenRegistry.Contract.CommunityAddressToTokenAddress(&_CommunityOwnerTokenRegistry.CallOpts, arg0)
}

// GetEntry is a free data retrieval call binding the contract method 0x7db6a4e4.
//
// Solidity: function getEntry(address _communityAddress) view returns(address)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryCaller) GetEntry(opts *bind.CallOpts, _communityAddress common.Address) (common.Address, error) {
	var out []interface{}
	err := _CommunityOwnerTokenRegistry.contract.Call(opts, &out, "getEntry", _communityAddress)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetEntry is a free data retrieval call binding the contract method 0x7db6a4e4.
//
// Solidity: function getEntry(address _communityAddress) view returns(address)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistrySession) GetEntry(_communityAddress common.Address) (common.Address, error) {
	return _CommunityOwnerTokenRegistry.Contract.GetEntry(&_CommunityOwnerTokenRegistry.CallOpts, _communityAddress)
}

// GetEntry is a free data retrieval call binding the contract method 0x7db6a4e4.
//
// Solidity: function getEntry(address _communityAddress) view returns(address)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryCallerSession) GetEntry(_communityAddress common.Address) (common.Address, error) {
	return _CommunityOwnerTokenRegistry.Contract.GetEntry(&_CommunityOwnerTokenRegistry.CallOpts, _communityAddress)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CommunityOwnerTokenRegistry.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistrySession) Owner() (common.Address, error) {
	return _CommunityOwnerTokenRegistry.Contract.Owner(&_CommunityOwnerTokenRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryCallerSession) Owner() (common.Address, error) {
	return _CommunityOwnerTokenRegistry.Contract.Owner(&_CommunityOwnerTokenRegistry.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CommunityOwnerTokenRegistry.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistrySession) PendingOwner() (common.Address, error) {
	return _CommunityOwnerTokenRegistry.Contract.PendingOwner(&_CommunityOwnerTokenRegistry.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryCallerSession) PendingOwner() (common.Address, error) {
	return _CommunityOwnerTokenRegistry.Contract.PendingOwner(&_CommunityOwnerTokenRegistry.CallOpts)
}

// TokenDeployer is a free data retrieval call binding the contract method 0x2a2dae0a.
//
// Solidity: function tokenDeployer() view returns(address)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryCaller) TokenDeployer(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CommunityOwnerTokenRegistry.contract.Call(opts, &out, "tokenDeployer")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TokenDeployer is a free data retrieval call binding the contract method 0x2a2dae0a.
//
// Solidity: function tokenDeployer() view returns(address)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistrySession) TokenDeployer() (common.Address, error) {
	return _CommunityOwnerTokenRegistry.Contract.TokenDeployer(&_CommunityOwnerTokenRegistry.CallOpts)
}

// TokenDeployer is a free data retrieval call binding the contract method 0x2a2dae0a.
//
// Solidity: function tokenDeployer() view returns(address)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryCallerSession) TokenDeployer() (common.Address, error) {
	return _CommunityOwnerTokenRegistry.Contract.TokenDeployer(&_CommunityOwnerTokenRegistry.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistrySession) AcceptOwnership() (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.Contract.AcceptOwnership(&_CommunityOwnerTokenRegistry.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.Contract.AcceptOwnership(&_CommunityOwnerTokenRegistry.TransactOpts)
}

// AddEntry is a paid mutator transaction binding the contract method 0xa7a95840.
//
// Solidity: function addEntry(address _communityAddress, address _tokenAddress) returns()
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryTransactor) AddEntry(opts *bind.TransactOpts, _communityAddress common.Address, _tokenAddress common.Address) (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.contract.Transact(opts, "addEntry", _communityAddress, _tokenAddress)
}

// AddEntry is a paid mutator transaction binding the contract method 0xa7a95840.
//
// Solidity: function addEntry(address _communityAddress, address _tokenAddress) returns()
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistrySession) AddEntry(_communityAddress common.Address, _tokenAddress common.Address) (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.Contract.AddEntry(&_CommunityOwnerTokenRegistry.TransactOpts, _communityAddress, _tokenAddress)
}

// AddEntry is a paid mutator transaction binding the contract method 0xa7a95840.
//
// Solidity: function addEntry(address _communityAddress, address _tokenAddress) returns()
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryTransactorSession) AddEntry(_communityAddress common.Address, _tokenAddress common.Address) (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.Contract.AddEntry(&_CommunityOwnerTokenRegistry.TransactOpts, _communityAddress, _tokenAddress)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.Contract.RenounceOwnership(&_CommunityOwnerTokenRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.Contract.RenounceOwnership(&_CommunityOwnerTokenRegistry.TransactOpts)
}

// SetCommunityTokenDeployerAddress is a paid mutator transaction binding the contract method 0xd1f7c48a.
//
// Solidity: function setCommunityTokenDeployerAddress(address _tokenDeployer) returns()
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryTransactor) SetCommunityTokenDeployerAddress(opts *bind.TransactOpts, _tokenDeployer common.Address) (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.contract.Transact(opts, "setCommunityTokenDeployerAddress", _tokenDeployer)
}

// SetCommunityTokenDeployerAddress is a paid mutator transaction binding the contract method 0xd1f7c48a.
//
// Solidity: function setCommunityTokenDeployerAddress(address _tokenDeployer) returns()
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistrySession) SetCommunityTokenDeployerAddress(_tokenDeployer common.Address) (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.Contract.SetCommunityTokenDeployerAddress(&_CommunityOwnerTokenRegistry.TransactOpts, _tokenDeployer)
}

// SetCommunityTokenDeployerAddress is a paid mutator transaction binding the contract method 0xd1f7c48a.
//
// Solidity: function setCommunityTokenDeployerAddress(address _tokenDeployer) returns()
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryTransactorSession) SetCommunityTokenDeployerAddress(_tokenDeployer common.Address) (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.Contract.SetCommunityTokenDeployerAddress(&_CommunityOwnerTokenRegistry.TransactOpts, _tokenDeployer)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistrySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.Contract.TransferOwnership(&_CommunityOwnerTokenRegistry.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _CommunityOwnerTokenRegistry.Contract.TransferOwnership(&_CommunityOwnerTokenRegistry.TransactOpts, newOwner)
}

// CommunityOwnerTokenRegistryAddEntryIterator is returned from FilterAddEntry and is used to iterate over the raw logs and unpacked data for AddEntry events raised by the CommunityOwnerTokenRegistry contract.
type CommunityOwnerTokenRegistryAddEntryIterator struct {
	Event *CommunityOwnerTokenRegistryAddEntry // Event containing the contract specifics and raw log

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
func (it *CommunityOwnerTokenRegistryAddEntryIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CommunityOwnerTokenRegistryAddEntry)
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
		it.Event = new(CommunityOwnerTokenRegistryAddEntry)
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
func (it *CommunityOwnerTokenRegistryAddEntryIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CommunityOwnerTokenRegistryAddEntryIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CommunityOwnerTokenRegistryAddEntry represents a AddEntry event raised by the CommunityOwnerTokenRegistry contract.
type CommunityOwnerTokenRegistryAddEntry struct {
	Arg0 common.Address
	Arg1 common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAddEntry is a free log retrieval operation binding the contract event 0x4bc4774424bf8749f142d7c1df17ee73cf36394616f38dd6799e99ea3bb4763a.
//
// Solidity: event AddEntry(address indexed arg0, address indexed arg1)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryFilterer) FilterAddEntry(opts *bind.FilterOpts, arg0 []common.Address, arg1 []common.Address) (*CommunityOwnerTokenRegistryAddEntryIterator, error) {

	var arg0Rule []interface{}
	for _, arg0Item := range arg0 {
		arg0Rule = append(arg0Rule, arg0Item)
	}
	var arg1Rule []interface{}
	for _, arg1Item := range arg1 {
		arg1Rule = append(arg1Rule, arg1Item)
	}

	logs, sub, err := _CommunityOwnerTokenRegistry.contract.FilterLogs(opts, "AddEntry", arg0Rule, arg1Rule)
	if err != nil {
		return nil, err
	}
	return &CommunityOwnerTokenRegistryAddEntryIterator{contract: _CommunityOwnerTokenRegistry.contract, event: "AddEntry", logs: logs, sub: sub}, nil
}

// WatchAddEntry is a free log subscription operation binding the contract event 0x4bc4774424bf8749f142d7c1df17ee73cf36394616f38dd6799e99ea3bb4763a.
//
// Solidity: event AddEntry(address indexed arg0, address indexed arg1)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryFilterer) WatchAddEntry(opts *bind.WatchOpts, sink chan<- *CommunityOwnerTokenRegistryAddEntry, arg0 []common.Address, arg1 []common.Address) (event.Subscription, error) {

	var arg0Rule []interface{}
	for _, arg0Item := range arg0 {
		arg0Rule = append(arg0Rule, arg0Item)
	}
	var arg1Rule []interface{}
	for _, arg1Item := range arg1 {
		arg1Rule = append(arg1Rule, arg1Item)
	}

	logs, sub, err := _CommunityOwnerTokenRegistry.contract.WatchLogs(opts, "AddEntry", arg0Rule, arg1Rule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CommunityOwnerTokenRegistryAddEntry)
				if err := _CommunityOwnerTokenRegistry.contract.UnpackLog(event, "AddEntry", log); err != nil {
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

// ParseAddEntry is a log parse operation binding the contract event 0x4bc4774424bf8749f142d7c1df17ee73cf36394616f38dd6799e99ea3bb4763a.
//
// Solidity: event AddEntry(address indexed arg0, address indexed arg1)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryFilterer) ParseAddEntry(log types.Log) (*CommunityOwnerTokenRegistryAddEntry, error) {
	event := new(CommunityOwnerTokenRegistryAddEntry)
	if err := _CommunityOwnerTokenRegistry.contract.UnpackLog(event, "AddEntry", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CommunityOwnerTokenRegistryOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the CommunityOwnerTokenRegistry contract.
type CommunityOwnerTokenRegistryOwnershipTransferStartedIterator struct {
	Event *CommunityOwnerTokenRegistryOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *CommunityOwnerTokenRegistryOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CommunityOwnerTokenRegistryOwnershipTransferStarted)
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
		it.Event = new(CommunityOwnerTokenRegistryOwnershipTransferStarted)
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
func (it *CommunityOwnerTokenRegistryOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CommunityOwnerTokenRegistryOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CommunityOwnerTokenRegistryOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the CommunityOwnerTokenRegistry contract.
type CommunityOwnerTokenRegistryOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*CommunityOwnerTokenRegistryOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _CommunityOwnerTokenRegistry.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &CommunityOwnerTokenRegistryOwnershipTransferStartedIterator{contract: _CommunityOwnerTokenRegistry.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *CommunityOwnerTokenRegistryOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _CommunityOwnerTokenRegistry.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CommunityOwnerTokenRegistryOwnershipTransferStarted)
				if err := _CommunityOwnerTokenRegistry.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryFilterer) ParseOwnershipTransferStarted(log types.Log) (*CommunityOwnerTokenRegistryOwnershipTransferStarted, error) {
	event := new(CommunityOwnerTokenRegistryOwnershipTransferStarted)
	if err := _CommunityOwnerTokenRegistry.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CommunityOwnerTokenRegistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the CommunityOwnerTokenRegistry contract.
type CommunityOwnerTokenRegistryOwnershipTransferredIterator struct {
	Event *CommunityOwnerTokenRegistryOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *CommunityOwnerTokenRegistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CommunityOwnerTokenRegistryOwnershipTransferred)
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
		it.Event = new(CommunityOwnerTokenRegistryOwnershipTransferred)
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
func (it *CommunityOwnerTokenRegistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CommunityOwnerTokenRegistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CommunityOwnerTokenRegistryOwnershipTransferred represents a OwnershipTransferred event raised by the CommunityOwnerTokenRegistry contract.
type CommunityOwnerTokenRegistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*CommunityOwnerTokenRegistryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _CommunityOwnerTokenRegistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &CommunityOwnerTokenRegistryOwnershipTransferredIterator{contract: _CommunityOwnerTokenRegistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *CommunityOwnerTokenRegistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _CommunityOwnerTokenRegistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CommunityOwnerTokenRegistryOwnershipTransferred)
				if err := _CommunityOwnerTokenRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryFilterer) ParseOwnershipTransferred(log types.Log) (*CommunityOwnerTokenRegistryOwnershipTransferred, error) {
	event := new(CommunityOwnerTokenRegistryOwnershipTransferred)
	if err := _CommunityOwnerTokenRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CommunityOwnerTokenRegistryTokenDeployerAddressChangeIterator is returned from FilterTokenDeployerAddressChange and is used to iterate over the raw logs and unpacked data for TokenDeployerAddressChange events raised by the CommunityOwnerTokenRegistry contract.
type CommunityOwnerTokenRegistryTokenDeployerAddressChangeIterator struct {
	Event *CommunityOwnerTokenRegistryTokenDeployerAddressChange // Event containing the contract specifics and raw log

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
func (it *CommunityOwnerTokenRegistryTokenDeployerAddressChangeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CommunityOwnerTokenRegistryTokenDeployerAddressChange)
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
		it.Event = new(CommunityOwnerTokenRegistryTokenDeployerAddressChange)
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
func (it *CommunityOwnerTokenRegistryTokenDeployerAddressChangeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CommunityOwnerTokenRegistryTokenDeployerAddressChangeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CommunityOwnerTokenRegistryTokenDeployerAddressChange represents a TokenDeployerAddressChange event raised by the CommunityOwnerTokenRegistry contract.
type CommunityOwnerTokenRegistryTokenDeployerAddressChange struct {
	Arg0 common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterTokenDeployerAddressChange is a free log retrieval operation binding the contract event 0x057829294de8b35baa4c034fc338afc6ecb2eb9b3035615c44100d80ecc93db7.
//
// Solidity: event TokenDeployerAddressChange(address indexed arg0)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryFilterer) FilterTokenDeployerAddressChange(opts *bind.FilterOpts, arg0 []common.Address) (*CommunityOwnerTokenRegistryTokenDeployerAddressChangeIterator, error) {

	var arg0Rule []interface{}
	for _, arg0Item := range arg0 {
		arg0Rule = append(arg0Rule, arg0Item)
	}

	logs, sub, err := _CommunityOwnerTokenRegistry.contract.FilterLogs(opts, "TokenDeployerAddressChange", arg0Rule)
	if err != nil {
		return nil, err
	}
	return &CommunityOwnerTokenRegistryTokenDeployerAddressChangeIterator{contract: _CommunityOwnerTokenRegistry.contract, event: "TokenDeployerAddressChange", logs: logs, sub: sub}, nil
}

// WatchTokenDeployerAddressChange is a free log subscription operation binding the contract event 0x057829294de8b35baa4c034fc338afc6ecb2eb9b3035615c44100d80ecc93db7.
//
// Solidity: event TokenDeployerAddressChange(address indexed arg0)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryFilterer) WatchTokenDeployerAddressChange(opts *bind.WatchOpts, sink chan<- *CommunityOwnerTokenRegistryTokenDeployerAddressChange, arg0 []common.Address) (event.Subscription, error) {

	var arg0Rule []interface{}
	for _, arg0Item := range arg0 {
		arg0Rule = append(arg0Rule, arg0Item)
	}

	logs, sub, err := _CommunityOwnerTokenRegistry.contract.WatchLogs(opts, "TokenDeployerAddressChange", arg0Rule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CommunityOwnerTokenRegistryTokenDeployerAddressChange)
				if err := _CommunityOwnerTokenRegistry.contract.UnpackLog(event, "TokenDeployerAddressChange", log); err != nil {
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

// ParseTokenDeployerAddressChange is a log parse operation binding the contract event 0x057829294de8b35baa4c034fc338afc6ecb2eb9b3035615c44100d80ecc93db7.
//
// Solidity: event TokenDeployerAddressChange(address indexed arg0)
func (_CommunityOwnerTokenRegistry *CommunityOwnerTokenRegistryFilterer) ParseTokenDeployerAddressChange(log types.Log) (*CommunityOwnerTokenRegistryTokenDeployerAddressChange, error) {
	event := new(CommunityOwnerTokenRegistryTokenDeployerAddressChange)
	if err := _CommunityOwnerTokenRegistry.contract.UnpackLog(event, "TokenDeployerAddressChange", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
