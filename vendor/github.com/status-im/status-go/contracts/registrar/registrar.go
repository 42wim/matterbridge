// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package registrar

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

// ApproveAndCallFallBackABI is the input ABI used to generate the binding from.
const ApproveAndCallFallBackABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"},{\"name\":\"_token\",\"type\":\"address\"},{\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"receiveApproval\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// ApproveAndCallFallBackFuncSigs maps the 4-byte function signature to its string representation.
var ApproveAndCallFallBackFuncSigs = map[string]string{
	"8f4ffcb1": "receiveApproval(address,uint256,address,bytes)",
}

// ApproveAndCallFallBack is an auto generated Go binding around an Ethereum contract.
type ApproveAndCallFallBack struct {
	ApproveAndCallFallBackCaller     // Read-only binding to the contract
	ApproveAndCallFallBackTransactor // Write-only binding to the contract
	ApproveAndCallFallBackFilterer   // Log filterer for contract events
}

// ApproveAndCallFallBackCaller is an auto generated read-only Go binding around an Ethereum contract.
type ApproveAndCallFallBackCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ApproveAndCallFallBackTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ApproveAndCallFallBackTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ApproveAndCallFallBackFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ApproveAndCallFallBackFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ApproveAndCallFallBackSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ApproveAndCallFallBackSession struct {
	Contract     *ApproveAndCallFallBack // Generic contract binding to set the session for
	CallOpts     bind.CallOpts           // Call options to use throughout this session
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// ApproveAndCallFallBackCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ApproveAndCallFallBackCallerSession struct {
	Contract *ApproveAndCallFallBackCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                 // Call options to use throughout this session
}

// ApproveAndCallFallBackTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ApproveAndCallFallBackTransactorSession struct {
	Contract     *ApproveAndCallFallBackTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                 // Transaction auth options to use throughout this session
}

// ApproveAndCallFallBackRaw is an auto generated low-level Go binding around an Ethereum contract.
type ApproveAndCallFallBackRaw struct {
	Contract *ApproveAndCallFallBack // Generic contract binding to access the raw methods on
}

// ApproveAndCallFallBackCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ApproveAndCallFallBackCallerRaw struct {
	Contract *ApproveAndCallFallBackCaller // Generic read-only contract binding to access the raw methods on
}

// ApproveAndCallFallBackTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ApproveAndCallFallBackTransactorRaw struct {
	Contract *ApproveAndCallFallBackTransactor // Generic write-only contract binding to access the raw methods on
}

// NewApproveAndCallFallBack creates a new instance of ApproveAndCallFallBack, bound to a specific deployed contract.
func NewApproveAndCallFallBack(address common.Address, backend bind.ContractBackend) (*ApproveAndCallFallBack, error) {
	contract, err := bindApproveAndCallFallBack(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ApproveAndCallFallBack{ApproveAndCallFallBackCaller: ApproveAndCallFallBackCaller{contract: contract}, ApproveAndCallFallBackTransactor: ApproveAndCallFallBackTransactor{contract: contract}, ApproveAndCallFallBackFilterer: ApproveAndCallFallBackFilterer{contract: contract}}, nil
}

// NewApproveAndCallFallBackCaller creates a new read-only instance of ApproveAndCallFallBack, bound to a specific deployed contract.
func NewApproveAndCallFallBackCaller(address common.Address, caller bind.ContractCaller) (*ApproveAndCallFallBackCaller, error) {
	contract, err := bindApproveAndCallFallBack(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ApproveAndCallFallBackCaller{contract: contract}, nil
}

// NewApproveAndCallFallBackTransactor creates a new write-only instance of ApproveAndCallFallBack, bound to a specific deployed contract.
func NewApproveAndCallFallBackTransactor(address common.Address, transactor bind.ContractTransactor) (*ApproveAndCallFallBackTransactor, error) {
	contract, err := bindApproveAndCallFallBack(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ApproveAndCallFallBackTransactor{contract: contract}, nil
}

// NewApproveAndCallFallBackFilterer creates a new log filterer instance of ApproveAndCallFallBack, bound to a specific deployed contract.
func NewApproveAndCallFallBackFilterer(address common.Address, filterer bind.ContractFilterer) (*ApproveAndCallFallBackFilterer, error) {
	contract, err := bindApproveAndCallFallBack(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ApproveAndCallFallBackFilterer{contract: contract}, nil
}

// bindApproveAndCallFallBack binds a generic wrapper to an already deployed contract.
func bindApproveAndCallFallBack(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ApproveAndCallFallBackABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ApproveAndCallFallBack *ApproveAndCallFallBackRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ApproveAndCallFallBack.Contract.ApproveAndCallFallBackCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ApproveAndCallFallBack *ApproveAndCallFallBackRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ApproveAndCallFallBack.Contract.ApproveAndCallFallBackTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ApproveAndCallFallBack *ApproveAndCallFallBackRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ApproveAndCallFallBack.Contract.ApproveAndCallFallBackTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ApproveAndCallFallBack *ApproveAndCallFallBackCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ApproveAndCallFallBack.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ApproveAndCallFallBack *ApproveAndCallFallBackTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ApproveAndCallFallBack.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ApproveAndCallFallBack *ApproveAndCallFallBackTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ApproveAndCallFallBack.Contract.contract.Transact(opts, method, params...)
}

// ReceiveApproval is a paid mutator transaction binding the contract method 0x8f4ffcb1.
//
// Solidity: function receiveApproval(address from, uint256 _amount, address _token, bytes _data) returns()
func (_ApproveAndCallFallBack *ApproveAndCallFallBackTransactor) ReceiveApproval(opts *bind.TransactOpts, from common.Address, _amount *big.Int, _token common.Address, _data []byte) (*types.Transaction, error) {
	return _ApproveAndCallFallBack.contract.Transact(opts, "receiveApproval", from, _amount, _token, _data)
}

// ReceiveApproval is a paid mutator transaction binding the contract method 0x8f4ffcb1.
//
// Solidity: function receiveApproval(address from, uint256 _amount, address _token, bytes _data) returns()
func (_ApproveAndCallFallBack *ApproveAndCallFallBackSession) ReceiveApproval(from common.Address, _amount *big.Int, _token common.Address, _data []byte) (*types.Transaction, error) {
	return _ApproveAndCallFallBack.Contract.ReceiveApproval(&_ApproveAndCallFallBack.TransactOpts, from, _amount, _token, _data)
}

// ReceiveApproval is a paid mutator transaction binding the contract method 0x8f4ffcb1.
//
// Solidity: function receiveApproval(address from, uint256 _amount, address _token, bytes _data) returns()
func (_ApproveAndCallFallBack *ApproveAndCallFallBackTransactorSession) ReceiveApproval(from common.Address, _amount *big.Int, _token common.Address, _data []byte) (*types.Transaction, error) {
	return _ApproveAndCallFallBack.Contract.ReceiveApproval(&_ApproveAndCallFallBack.TransactOpts, from, _amount, _token, _data)
}

// ControlledABI is the input ABI used to generate the binding from.
const ControlledABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_newController\",\"type\":\"address\"}],\"name\":\"changeController\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"controller\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// ControlledFuncSigs maps the 4-byte function signature to its string representation.
var ControlledFuncSigs = map[string]string{
	"3cebb823": "changeController(address)",
	"f77c4791": "controller()",
}

// Controlled is an auto generated Go binding around an Ethereum contract.
type Controlled struct {
	ControlledCaller     // Read-only binding to the contract
	ControlledTransactor // Write-only binding to the contract
	ControlledFilterer   // Log filterer for contract events
}

// ControlledCaller is an auto generated read-only Go binding around an Ethereum contract.
type ControlledCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ControlledTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ControlledTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ControlledFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ControlledFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ControlledSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ControlledSession struct {
	Contract     *Controlled       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ControlledCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ControlledCallerSession struct {
	Contract *ControlledCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// ControlledTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ControlledTransactorSession struct {
	Contract     *ControlledTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// ControlledRaw is an auto generated low-level Go binding around an Ethereum contract.
type ControlledRaw struct {
	Contract *Controlled // Generic contract binding to access the raw methods on
}

// ControlledCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ControlledCallerRaw struct {
	Contract *ControlledCaller // Generic read-only contract binding to access the raw methods on
}

// ControlledTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ControlledTransactorRaw struct {
	Contract *ControlledTransactor // Generic write-only contract binding to access the raw methods on
}

// NewControlled creates a new instance of Controlled, bound to a specific deployed contract.
func NewControlled(address common.Address, backend bind.ContractBackend) (*Controlled, error) {
	contract, err := bindControlled(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Controlled{ControlledCaller: ControlledCaller{contract: contract}, ControlledTransactor: ControlledTransactor{contract: contract}, ControlledFilterer: ControlledFilterer{contract: contract}}, nil
}

// NewControlledCaller creates a new read-only instance of Controlled, bound to a specific deployed contract.
func NewControlledCaller(address common.Address, caller bind.ContractCaller) (*ControlledCaller, error) {
	contract, err := bindControlled(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ControlledCaller{contract: contract}, nil
}

// NewControlledTransactor creates a new write-only instance of Controlled, bound to a specific deployed contract.
func NewControlledTransactor(address common.Address, transactor bind.ContractTransactor) (*ControlledTransactor, error) {
	contract, err := bindControlled(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ControlledTransactor{contract: contract}, nil
}

// NewControlledFilterer creates a new log filterer instance of Controlled, bound to a specific deployed contract.
func NewControlledFilterer(address common.Address, filterer bind.ContractFilterer) (*ControlledFilterer, error) {
	contract, err := bindControlled(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ControlledFilterer{contract: contract}, nil
}

// bindControlled binds a generic wrapper to an already deployed contract.
func bindControlled(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ControlledABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Controlled *ControlledRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Controlled.Contract.ControlledCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Controlled *ControlledRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Controlled.Contract.ControlledTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Controlled *ControlledRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Controlled.Contract.ControlledTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Controlled *ControlledCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Controlled.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Controlled *ControlledTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Controlled.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Controlled *ControlledTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Controlled.Contract.contract.Transact(opts, method, params...)
}

// Controller is a free data retrieval call binding the contract method 0xf77c4791.
//
// Solidity: function controller() view returns(address)
func (_Controlled *ControlledCaller) Controller(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Controlled.contract.Call(opts, &out, "controller")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Controller is a free data retrieval call binding the contract method 0xf77c4791.
//
// Solidity: function controller() view returns(address)
func (_Controlled *ControlledSession) Controller() (common.Address, error) {
	return _Controlled.Contract.Controller(&_Controlled.CallOpts)
}

// Controller is a free data retrieval call binding the contract method 0xf77c4791.
//
// Solidity: function controller() view returns(address)
func (_Controlled *ControlledCallerSession) Controller() (common.Address, error) {
	return _Controlled.Contract.Controller(&_Controlled.CallOpts)
}

// ChangeController is a paid mutator transaction binding the contract method 0x3cebb823.
//
// Solidity: function changeController(address _newController) returns()
func (_Controlled *ControlledTransactor) ChangeController(opts *bind.TransactOpts, _newController common.Address) (*types.Transaction, error) {
	return _Controlled.contract.Transact(opts, "changeController", _newController)
}

// ChangeController is a paid mutator transaction binding the contract method 0x3cebb823.
//
// Solidity: function changeController(address _newController) returns()
func (_Controlled *ControlledSession) ChangeController(_newController common.Address) (*types.Transaction, error) {
	return _Controlled.Contract.ChangeController(&_Controlled.TransactOpts, _newController)
}

// ChangeController is a paid mutator transaction binding the contract method 0x3cebb823.
//
// Solidity: function changeController(address _newController) returns()
func (_Controlled *ControlledTransactorSession) ChangeController(_newController common.Address) (*types.Transaction, error) {
	return _Controlled.Contract.ChangeController(&_Controlled.TransactOpts, _newController)
}

// UsernameRegistrarABI is the input ABI used to generate the binding from.
const UsernameRegistrarABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"resolver\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_secret\",\"type\":\"bytes32\"}],\"name\":\"reserveSlash\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"reservedUsernamesMerkleRoot\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_token\",\"type\":\"address\"},{\"name\":\"_beneficiary\",\"type\":\"address\"}],\"name\":\"withdrawExcessBalance\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_label\",\"type\":\"bytes32\"}],\"name\":\"updateAccountOwner\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newController\",\"type\":\"address\"}],\"name\":\"changeController\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_username\",\"type\":\"string\"},{\"name\":\"_offendingPos\",\"type\":\"uint256\"},{\"name\":\"_reserveSecret\",\"type\":\"uint256\"}],\"name\":\"slashInvalidUsername\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_username\",\"type\":\"string\"},{\"name\":\"_proof\",\"type\":\"bytes32[]\"},{\"name\":\"_reserveSecret\",\"type\":\"uint256\"}],\"name\":\"slashReservedUsername\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"reserveAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_resolver\",\"type\":\"address\"}],\"name\":\"setResolver\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"usernameMinLength\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_label\",\"type\":\"bytes32\"}],\"name\":\"release\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_label\",\"type\":\"bytes32\"}],\"name\":\"getCreationTime\",\"outputs\":[{\"name\":\"creationTime\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"releaseDelay\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"ensRegistry\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_label\",\"type\":\"bytes32\"},{\"name\":\"_tokenBalance\",\"type\":\"uint256\"},{\"name\":\"_creationTime\",\"type\":\"uint256\"},{\"name\":\"_accountOwner\",\"type\":\"address\"}],\"name\":\"migrateUsername\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_label\",\"type\":\"bytes32\"}],\"name\":\"getSlashRewardPart\",\"outputs\":[{\"name\":\"partReward\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_price\",\"type\":\"uint256\"}],\"name\":\"updateRegistryPrice\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_username\",\"type\":\"string\"},{\"name\":\"_reserveSecret\",\"type\":\"uint256\"}],\"name\":\"slashAddressLikeUsername\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"},{\"name\":\"_token\",\"type\":\"address\"},{\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"receiveApproval\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_username\",\"type\":\"string\"},{\"name\":\"_reserveSecret\",\"type\":\"uint256\"}],\"name\":\"slashSmallUsername\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getPrice\",\"outputs\":[{\"name\":\"registryPrice\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_price\",\"type\":\"uint256\"}],\"name\":\"migrateRegistry\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"price\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_label\",\"type\":\"bytes32\"}],\"name\":\"getExpirationTime\",\"outputs\":[{\"name\":\"releaseTime\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_label\",\"type\":\"bytes32\"}],\"name\":\"getAccountOwner\",\"outputs\":[{\"name\":\"owner\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_domainHash\",\"type\":\"bytes32\"},{\"name\":\"_beneficiary\",\"type\":\"address\"}],\"name\":\"withdrawWrongNode\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_price\",\"type\":\"uint256\"}],\"name\":\"activate\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_label\",\"type\":\"bytes32\"},{\"name\":\"_account\",\"type\":\"address\"},{\"name\":\"_pubkeyA\",\"type\":\"bytes32\"},{\"name\":\"_pubkeyB\",\"type\":\"bytes32\"}],\"name\":\"register\",\"outputs\":[{\"name\":\"namehash\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"accounts\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"},{\"name\":\"creationTime\",\"type\":\"uint256\"},{\"name\":\"owner\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"state\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_label\",\"type\":\"bytes32\"},{\"name\":\"_newRegistry\",\"type\":\"address\"}],\"name\":\"moveAccount\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"parentRegistry\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"ensNode\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_labels\",\"type\":\"bytes32[]\"}],\"name\":\"eraseNode\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newRegistry\",\"type\":\"address\"}],\"name\":\"moveRegistry\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_label\",\"type\":\"bytes32\"}],\"name\":\"getAccountBalance\",\"outputs\":[{\"name\":\"accountBalance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"controller\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_label\",\"type\":\"bytes32\"}],\"name\":\"dropUsername\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"state\",\"type\":\"uint8\"}],\"name\":\"RegistryState\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"price\",\"type\":\"uint256\"}],\"name\":\"RegistryPrice\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"newRegistry\",\"type\":\"address\"}],\"name\":\"RegistryMoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"nameHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"UsernameOwner\",\"type\":\"event\"}]"

// UsernameRegistrarFuncSigs maps the 4-byte function signature to its string representation.
var UsernameRegistrarFuncSigs = map[string]string{
	"bc529c43": "accounts(bytes32)",
	"b260c42a": "activate(uint256)",
	"3cebb823": "changeController(address)",
	"f77c4791": "controller()",
	"f9e54282": "dropUsername(bytes32)",
	"ddbcf3a1": "ensNode()",
	"7d73b231": "ensRegistry()",
	"de10f04b": "eraseNode(bytes32[])",
	"ebf701e0": "getAccountBalance(bytes32)",
	"aacffccf": "getAccountOwner(bytes32)",
	"6f79301d": "getCreationTime(bytes32)",
	"a1454830": "getExpirationTime(bytes32)",
	"98d5fdca": "getPrice()",
	"8382b460": "getSlashRewardPart(bytes32)",
	"98f038ff": "migrateRegistry(uint256)",
	"80cd0015": "migrateUsername(bytes32,uint256,uint256,address)",
	"c23e61b9": "moveAccount(bytes32,address)",
	"e882c3ce": "moveRegistry(address)",
	"c9b84d4d": "parentRegistry()",
	"a035b1fe": "price()",
	"8f4ffcb1": "receiveApproval(address,uint256,address,bytes)",
	"b82fedbb": "register(bytes32,address,bytes32,bytes32)",
	"67d42a8b": "release(bytes32)",
	"7195bf23": "releaseDelay()",
	"4b09b72a": "reserveAmount()",
	"05c24481": "reserveSlash(bytes32)",
	"07f908cb": "reservedUsernamesMerkleRoot()",
	"04f3bcec": "resolver()",
	"4e543b26": "setResolver(address)",
	"8cf7b7a4": "slashAddressLikeUsername(string,uint256)",
	"40784ebd": "slashInvalidUsername(string,uint256,uint256)",
	"40b1ad52": "slashReservedUsername(string,bytes32[],uint256)",
	"96bba9a8": "slashSmallUsername(string,uint256)",
	"c19d93fb": "state()",
	"fc0c546a": "token()",
	"32e1ed24": "updateAccountOwner(bytes32)",
	"860e9b0f": "updateRegistryPrice(uint256)",
	"59ad0209": "usernameMinLength()",
	"307c7a0d": "withdrawExcessBalance(address,address)",
	"afe12e77": "withdrawWrongNode(bytes32,address)",
}

// UsernameRegistrar is an auto generated Go binding around an Ethereum contract.
type UsernameRegistrar struct {
	UsernameRegistrarCaller     // Read-only binding to the contract
	UsernameRegistrarTransactor // Write-only binding to the contract
	UsernameRegistrarFilterer   // Log filterer for contract events
}

// UsernameRegistrarCaller is an auto generated read-only Go binding around an Ethereum contract.
type UsernameRegistrarCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UsernameRegistrarTransactor is an auto generated write-only Go binding around an Ethereum contract.
type UsernameRegistrarTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UsernameRegistrarFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type UsernameRegistrarFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UsernameRegistrarSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type UsernameRegistrarSession struct {
	Contract     *UsernameRegistrar // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// UsernameRegistrarCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type UsernameRegistrarCallerSession struct {
	Contract *UsernameRegistrarCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// UsernameRegistrarTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type UsernameRegistrarTransactorSession struct {
	Contract     *UsernameRegistrarTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// UsernameRegistrarRaw is an auto generated low-level Go binding around an Ethereum contract.
type UsernameRegistrarRaw struct {
	Contract *UsernameRegistrar // Generic contract binding to access the raw methods on
}

// UsernameRegistrarCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type UsernameRegistrarCallerRaw struct {
	Contract *UsernameRegistrarCaller // Generic read-only contract binding to access the raw methods on
}

// UsernameRegistrarTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type UsernameRegistrarTransactorRaw struct {
	Contract *UsernameRegistrarTransactor // Generic write-only contract binding to access the raw methods on
}

// NewUsernameRegistrar creates a new instance of UsernameRegistrar, bound to a specific deployed contract.
func NewUsernameRegistrar(address common.Address, backend bind.ContractBackend) (*UsernameRegistrar, error) {
	contract, err := bindUsernameRegistrar(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &UsernameRegistrar{UsernameRegistrarCaller: UsernameRegistrarCaller{contract: contract}, UsernameRegistrarTransactor: UsernameRegistrarTransactor{contract: contract}, UsernameRegistrarFilterer: UsernameRegistrarFilterer{contract: contract}}, nil
}

// NewUsernameRegistrarCaller creates a new read-only instance of UsernameRegistrar, bound to a specific deployed contract.
func NewUsernameRegistrarCaller(address common.Address, caller bind.ContractCaller) (*UsernameRegistrarCaller, error) {
	contract, err := bindUsernameRegistrar(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &UsernameRegistrarCaller{contract: contract}, nil
}

// NewUsernameRegistrarTransactor creates a new write-only instance of UsernameRegistrar, bound to a specific deployed contract.
func NewUsernameRegistrarTransactor(address common.Address, transactor bind.ContractTransactor) (*UsernameRegistrarTransactor, error) {
	contract, err := bindUsernameRegistrar(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &UsernameRegistrarTransactor{contract: contract}, nil
}

// NewUsernameRegistrarFilterer creates a new log filterer instance of UsernameRegistrar, bound to a specific deployed contract.
func NewUsernameRegistrarFilterer(address common.Address, filterer bind.ContractFilterer) (*UsernameRegistrarFilterer, error) {
	contract, err := bindUsernameRegistrar(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &UsernameRegistrarFilterer{contract: contract}, nil
}

// bindUsernameRegistrar binds a generic wrapper to an already deployed contract.
func bindUsernameRegistrar(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(UsernameRegistrarABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UsernameRegistrar *UsernameRegistrarRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _UsernameRegistrar.Contract.UsernameRegistrarCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UsernameRegistrar *UsernameRegistrarRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.UsernameRegistrarTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UsernameRegistrar *UsernameRegistrarRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.UsernameRegistrarTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UsernameRegistrar *UsernameRegistrarCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _UsernameRegistrar.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UsernameRegistrar *UsernameRegistrarTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UsernameRegistrar *UsernameRegistrarTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.contract.Transact(opts, method, params...)
}

// Accounts is a free data retrieval call binding the contract method 0xbc529c43.
//
// Solidity: function accounts(bytes32 ) view returns(uint256 balance, uint256 creationTime, address owner)
func (_UsernameRegistrar *UsernameRegistrarCaller) Accounts(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Balance      *big.Int
	CreationTime *big.Int
	Owner        common.Address
}, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "accounts", arg0)

	outstruct := new(struct {
		Balance      *big.Int
		CreationTime *big.Int
		Owner        common.Address
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Balance = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.CreationTime = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Owner = *abi.ConvertType(out[2], new(common.Address)).(*common.Address)

	return *outstruct, err

}

// Accounts is a free data retrieval call binding the contract method 0xbc529c43.
//
// Solidity: function accounts(bytes32 ) view returns(uint256 balance, uint256 creationTime, address owner)
func (_UsernameRegistrar *UsernameRegistrarSession) Accounts(arg0 [32]byte) (struct {
	Balance      *big.Int
	CreationTime *big.Int
	Owner        common.Address
}, error) {
	return _UsernameRegistrar.Contract.Accounts(&_UsernameRegistrar.CallOpts, arg0)
}

// Accounts is a free data retrieval call binding the contract method 0xbc529c43.
//
// Solidity: function accounts(bytes32 ) view returns(uint256 balance, uint256 creationTime, address owner)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) Accounts(arg0 [32]byte) (struct {
	Balance      *big.Int
	CreationTime *big.Int
	Owner        common.Address
}, error) {
	return _UsernameRegistrar.Contract.Accounts(&_UsernameRegistrar.CallOpts, arg0)
}

// Controller is a free data retrieval call binding the contract method 0xf77c4791.
//
// Solidity: function controller() view returns(address)
func (_UsernameRegistrar *UsernameRegistrarCaller) Controller(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "controller")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Controller is a free data retrieval call binding the contract method 0xf77c4791.
//
// Solidity: function controller() view returns(address)
func (_UsernameRegistrar *UsernameRegistrarSession) Controller() (common.Address, error) {
	return _UsernameRegistrar.Contract.Controller(&_UsernameRegistrar.CallOpts)
}

// Controller is a free data retrieval call binding the contract method 0xf77c4791.
//
// Solidity: function controller() view returns(address)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) Controller() (common.Address, error) {
	return _UsernameRegistrar.Contract.Controller(&_UsernameRegistrar.CallOpts)
}

// EnsNode is a free data retrieval call binding the contract method 0xddbcf3a1.
//
// Solidity: function ensNode() view returns(bytes32)
func (_UsernameRegistrar *UsernameRegistrarCaller) EnsNode(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "ensNode")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// EnsNode is a free data retrieval call binding the contract method 0xddbcf3a1.
//
// Solidity: function ensNode() view returns(bytes32)
func (_UsernameRegistrar *UsernameRegistrarSession) EnsNode() ([32]byte, error) {
	return _UsernameRegistrar.Contract.EnsNode(&_UsernameRegistrar.CallOpts)
}

// EnsNode is a free data retrieval call binding the contract method 0xddbcf3a1.
//
// Solidity: function ensNode() view returns(bytes32)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) EnsNode() ([32]byte, error) {
	return _UsernameRegistrar.Contract.EnsNode(&_UsernameRegistrar.CallOpts)
}

// EnsRegistry is a free data retrieval call binding the contract method 0x7d73b231.
//
// Solidity: function ensRegistry() view returns(address)
func (_UsernameRegistrar *UsernameRegistrarCaller) EnsRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "ensRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EnsRegistry is a free data retrieval call binding the contract method 0x7d73b231.
//
// Solidity: function ensRegistry() view returns(address)
func (_UsernameRegistrar *UsernameRegistrarSession) EnsRegistry() (common.Address, error) {
	return _UsernameRegistrar.Contract.EnsRegistry(&_UsernameRegistrar.CallOpts)
}

// EnsRegistry is a free data retrieval call binding the contract method 0x7d73b231.
//
// Solidity: function ensRegistry() view returns(address)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) EnsRegistry() (common.Address, error) {
	return _UsernameRegistrar.Contract.EnsRegistry(&_UsernameRegistrar.CallOpts)
}

// GetAccountBalance is a free data retrieval call binding the contract method 0xebf701e0.
//
// Solidity: function getAccountBalance(bytes32 _label) view returns(uint256 accountBalance)
func (_UsernameRegistrar *UsernameRegistrarCaller) GetAccountBalance(opts *bind.CallOpts, _label [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "getAccountBalance", _label)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAccountBalance is a free data retrieval call binding the contract method 0xebf701e0.
//
// Solidity: function getAccountBalance(bytes32 _label) view returns(uint256 accountBalance)
func (_UsernameRegistrar *UsernameRegistrarSession) GetAccountBalance(_label [32]byte) (*big.Int, error) {
	return _UsernameRegistrar.Contract.GetAccountBalance(&_UsernameRegistrar.CallOpts, _label)
}

// GetAccountBalance is a free data retrieval call binding the contract method 0xebf701e0.
//
// Solidity: function getAccountBalance(bytes32 _label) view returns(uint256 accountBalance)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) GetAccountBalance(_label [32]byte) (*big.Int, error) {
	return _UsernameRegistrar.Contract.GetAccountBalance(&_UsernameRegistrar.CallOpts, _label)
}

// GetAccountOwner is a free data retrieval call binding the contract method 0xaacffccf.
//
// Solidity: function getAccountOwner(bytes32 _label) view returns(address owner)
func (_UsernameRegistrar *UsernameRegistrarCaller) GetAccountOwner(opts *bind.CallOpts, _label [32]byte) (common.Address, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "getAccountOwner", _label)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetAccountOwner is a free data retrieval call binding the contract method 0xaacffccf.
//
// Solidity: function getAccountOwner(bytes32 _label) view returns(address owner)
func (_UsernameRegistrar *UsernameRegistrarSession) GetAccountOwner(_label [32]byte) (common.Address, error) {
	return _UsernameRegistrar.Contract.GetAccountOwner(&_UsernameRegistrar.CallOpts, _label)
}

// GetAccountOwner is a free data retrieval call binding the contract method 0xaacffccf.
//
// Solidity: function getAccountOwner(bytes32 _label) view returns(address owner)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) GetAccountOwner(_label [32]byte) (common.Address, error) {
	return _UsernameRegistrar.Contract.GetAccountOwner(&_UsernameRegistrar.CallOpts, _label)
}

// GetCreationTime is a free data retrieval call binding the contract method 0x6f79301d.
//
// Solidity: function getCreationTime(bytes32 _label) view returns(uint256 creationTime)
func (_UsernameRegistrar *UsernameRegistrarCaller) GetCreationTime(opts *bind.CallOpts, _label [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "getCreationTime", _label)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCreationTime is a free data retrieval call binding the contract method 0x6f79301d.
//
// Solidity: function getCreationTime(bytes32 _label) view returns(uint256 creationTime)
func (_UsernameRegistrar *UsernameRegistrarSession) GetCreationTime(_label [32]byte) (*big.Int, error) {
	return _UsernameRegistrar.Contract.GetCreationTime(&_UsernameRegistrar.CallOpts, _label)
}

// GetCreationTime is a free data retrieval call binding the contract method 0x6f79301d.
//
// Solidity: function getCreationTime(bytes32 _label) view returns(uint256 creationTime)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) GetCreationTime(_label [32]byte) (*big.Int, error) {
	return _UsernameRegistrar.Contract.GetCreationTime(&_UsernameRegistrar.CallOpts, _label)
}

// GetExpirationTime is a free data retrieval call binding the contract method 0xa1454830.
//
// Solidity: function getExpirationTime(bytes32 _label) view returns(uint256 releaseTime)
func (_UsernameRegistrar *UsernameRegistrarCaller) GetExpirationTime(opts *bind.CallOpts, _label [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "getExpirationTime", _label)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetExpirationTime is a free data retrieval call binding the contract method 0xa1454830.
//
// Solidity: function getExpirationTime(bytes32 _label) view returns(uint256 releaseTime)
func (_UsernameRegistrar *UsernameRegistrarSession) GetExpirationTime(_label [32]byte) (*big.Int, error) {
	return _UsernameRegistrar.Contract.GetExpirationTime(&_UsernameRegistrar.CallOpts, _label)
}

// GetExpirationTime is a free data retrieval call binding the contract method 0xa1454830.
//
// Solidity: function getExpirationTime(bytes32 _label) view returns(uint256 releaseTime)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) GetExpirationTime(_label [32]byte) (*big.Int, error) {
	return _UsernameRegistrar.Contract.GetExpirationTime(&_UsernameRegistrar.CallOpts, _label)
}

// GetPrice is a free data retrieval call binding the contract method 0x98d5fdca.
//
// Solidity: function getPrice() view returns(uint256 registryPrice)
func (_UsernameRegistrar *UsernameRegistrarCaller) GetPrice(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "getPrice")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetPrice is a free data retrieval call binding the contract method 0x98d5fdca.
//
// Solidity: function getPrice() view returns(uint256 registryPrice)
func (_UsernameRegistrar *UsernameRegistrarSession) GetPrice() (*big.Int, error) {
	return _UsernameRegistrar.Contract.GetPrice(&_UsernameRegistrar.CallOpts)
}

// GetPrice is a free data retrieval call binding the contract method 0x98d5fdca.
//
// Solidity: function getPrice() view returns(uint256 registryPrice)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) GetPrice() (*big.Int, error) {
	return _UsernameRegistrar.Contract.GetPrice(&_UsernameRegistrar.CallOpts)
}

// GetSlashRewardPart is a free data retrieval call binding the contract method 0x8382b460.
//
// Solidity: function getSlashRewardPart(bytes32 _label) view returns(uint256 partReward)
func (_UsernameRegistrar *UsernameRegistrarCaller) GetSlashRewardPart(opts *bind.CallOpts, _label [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "getSlashRewardPart", _label)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSlashRewardPart is a free data retrieval call binding the contract method 0x8382b460.
//
// Solidity: function getSlashRewardPart(bytes32 _label) view returns(uint256 partReward)
func (_UsernameRegistrar *UsernameRegistrarSession) GetSlashRewardPart(_label [32]byte) (*big.Int, error) {
	return _UsernameRegistrar.Contract.GetSlashRewardPart(&_UsernameRegistrar.CallOpts, _label)
}

// GetSlashRewardPart is a free data retrieval call binding the contract method 0x8382b460.
//
// Solidity: function getSlashRewardPart(bytes32 _label) view returns(uint256 partReward)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) GetSlashRewardPart(_label [32]byte) (*big.Int, error) {
	return _UsernameRegistrar.Contract.GetSlashRewardPart(&_UsernameRegistrar.CallOpts, _label)
}

// ParentRegistry is a free data retrieval call binding the contract method 0xc9b84d4d.
//
// Solidity: function parentRegistry() view returns(address)
func (_UsernameRegistrar *UsernameRegistrarCaller) ParentRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "parentRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ParentRegistry is a free data retrieval call binding the contract method 0xc9b84d4d.
//
// Solidity: function parentRegistry() view returns(address)
func (_UsernameRegistrar *UsernameRegistrarSession) ParentRegistry() (common.Address, error) {
	return _UsernameRegistrar.Contract.ParentRegistry(&_UsernameRegistrar.CallOpts)
}

// ParentRegistry is a free data retrieval call binding the contract method 0xc9b84d4d.
//
// Solidity: function parentRegistry() view returns(address)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) ParentRegistry() (common.Address, error) {
	return _UsernameRegistrar.Contract.ParentRegistry(&_UsernameRegistrar.CallOpts)
}

// Price is a free data retrieval call binding the contract method 0xa035b1fe.
//
// Solidity: function price() view returns(uint256)
func (_UsernameRegistrar *UsernameRegistrarCaller) Price(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "price")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Price is a free data retrieval call binding the contract method 0xa035b1fe.
//
// Solidity: function price() view returns(uint256)
func (_UsernameRegistrar *UsernameRegistrarSession) Price() (*big.Int, error) {
	return _UsernameRegistrar.Contract.Price(&_UsernameRegistrar.CallOpts)
}

// Price is a free data retrieval call binding the contract method 0xa035b1fe.
//
// Solidity: function price() view returns(uint256)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) Price() (*big.Int, error) {
	return _UsernameRegistrar.Contract.Price(&_UsernameRegistrar.CallOpts)
}

// ReleaseDelay is a free data retrieval call binding the contract method 0x7195bf23.
//
// Solidity: function releaseDelay() view returns(uint256)
func (_UsernameRegistrar *UsernameRegistrarCaller) ReleaseDelay(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "releaseDelay")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ReleaseDelay is a free data retrieval call binding the contract method 0x7195bf23.
//
// Solidity: function releaseDelay() view returns(uint256)
func (_UsernameRegistrar *UsernameRegistrarSession) ReleaseDelay() (*big.Int, error) {
	return _UsernameRegistrar.Contract.ReleaseDelay(&_UsernameRegistrar.CallOpts)
}

// ReleaseDelay is a free data retrieval call binding the contract method 0x7195bf23.
//
// Solidity: function releaseDelay() view returns(uint256)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) ReleaseDelay() (*big.Int, error) {
	return _UsernameRegistrar.Contract.ReleaseDelay(&_UsernameRegistrar.CallOpts)
}

// ReserveAmount is a free data retrieval call binding the contract method 0x4b09b72a.
//
// Solidity: function reserveAmount() view returns(uint256)
func (_UsernameRegistrar *UsernameRegistrarCaller) ReserveAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "reserveAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ReserveAmount is a free data retrieval call binding the contract method 0x4b09b72a.
//
// Solidity: function reserveAmount() view returns(uint256)
func (_UsernameRegistrar *UsernameRegistrarSession) ReserveAmount() (*big.Int, error) {
	return _UsernameRegistrar.Contract.ReserveAmount(&_UsernameRegistrar.CallOpts)
}

// ReserveAmount is a free data retrieval call binding the contract method 0x4b09b72a.
//
// Solidity: function reserveAmount() view returns(uint256)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) ReserveAmount() (*big.Int, error) {
	return _UsernameRegistrar.Contract.ReserveAmount(&_UsernameRegistrar.CallOpts)
}

// ReservedUsernamesMerkleRoot is a free data retrieval call binding the contract method 0x07f908cb.
//
// Solidity: function reservedUsernamesMerkleRoot() view returns(bytes32)
func (_UsernameRegistrar *UsernameRegistrarCaller) ReservedUsernamesMerkleRoot(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "reservedUsernamesMerkleRoot")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ReservedUsernamesMerkleRoot is a free data retrieval call binding the contract method 0x07f908cb.
//
// Solidity: function reservedUsernamesMerkleRoot() view returns(bytes32)
func (_UsernameRegistrar *UsernameRegistrarSession) ReservedUsernamesMerkleRoot() ([32]byte, error) {
	return _UsernameRegistrar.Contract.ReservedUsernamesMerkleRoot(&_UsernameRegistrar.CallOpts)
}

// ReservedUsernamesMerkleRoot is a free data retrieval call binding the contract method 0x07f908cb.
//
// Solidity: function reservedUsernamesMerkleRoot() view returns(bytes32)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) ReservedUsernamesMerkleRoot() ([32]byte, error) {
	return _UsernameRegistrar.Contract.ReservedUsernamesMerkleRoot(&_UsernameRegistrar.CallOpts)
}

// Resolver is a free data retrieval call binding the contract method 0x04f3bcec.
//
// Solidity: function resolver() view returns(address)
func (_UsernameRegistrar *UsernameRegistrarCaller) Resolver(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "resolver")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Resolver is a free data retrieval call binding the contract method 0x04f3bcec.
//
// Solidity: function resolver() view returns(address)
func (_UsernameRegistrar *UsernameRegistrarSession) Resolver() (common.Address, error) {
	return _UsernameRegistrar.Contract.Resolver(&_UsernameRegistrar.CallOpts)
}

// Resolver is a free data retrieval call binding the contract method 0x04f3bcec.
//
// Solidity: function resolver() view returns(address)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) Resolver() (common.Address, error) {
	return _UsernameRegistrar.Contract.Resolver(&_UsernameRegistrar.CallOpts)
}

// State is a free data retrieval call binding the contract method 0xc19d93fb.
//
// Solidity: function state() view returns(uint8)
func (_UsernameRegistrar *UsernameRegistrarCaller) State(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "state")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// State is a free data retrieval call binding the contract method 0xc19d93fb.
//
// Solidity: function state() view returns(uint8)
func (_UsernameRegistrar *UsernameRegistrarSession) State() (uint8, error) {
	return _UsernameRegistrar.Contract.State(&_UsernameRegistrar.CallOpts)
}

// State is a free data retrieval call binding the contract method 0xc19d93fb.
//
// Solidity: function state() view returns(uint8)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) State() (uint8, error) {
	return _UsernameRegistrar.Contract.State(&_UsernameRegistrar.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_UsernameRegistrar *UsernameRegistrarCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_UsernameRegistrar *UsernameRegistrarSession) Token() (common.Address, error) {
	return _UsernameRegistrar.Contract.Token(&_UsernameRegistrar.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) Token() (common.Address, error) {
	return _UsernameRegistrar.Contract.Token(&_UsernameRegistrar.CallOpts)
}

// UsernameMinLength is a free data retrieval call binding the contract method 0x59ad0209.
//
// Solidity: function usernameMinLength() view returns(uint256)
func (_UsernameRegistrar *UsernameRegistrarCaller) UsernameMinLength(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _UsernameRegistrar.contract.Call(opts, &out, "usernameMinLength")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UsernameMinLength is a free data retrieval call binding the contract method 0x59ad0209.
//
// Solidity: function usernameMinLength() view returns(uint256)
func (_UsernameRegistrar *UsernameRegistrarSession) UsernameMinLength() (*big.Int, error) {
	return _UsernameRegistrar.Contract.UsernameMinLength(&_UsernameRegistrar.CallOpts)
}

// UsernameMinLength is a free data retrieval call binding the contract method 0x59ad0209.
//
// Solidity: function usernameMinLength() view returns(uint256)
func (_UsernameRegistrar *UsernameRegistrarCallerSession) UsernameMinLength() (*big.Int, error) {
	return _UsernameRegistrar.Contract.UsernameMinLength(&_UsernameRegistrar.CallOpts)
}

// Activate is a paid mutator transaction binding the contract method 0xb260c42a.
//
// Solidity: function activate(uint256 _price) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) Activate(opts *bind.TransactOpts, _price *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "activate", _price)
}

// Activate is a paid mutator transaction binding the contract method 0xb260c42a.
//
// Solidity: function activate(uint256 _price) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) Activate(_price *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.Activate(&_UsernameRegistrar.TransactOpts, _price)
}

// Activate is a paid mutator transaction binding the contract method 0xb260c42a.
//
// Solidity: function activate(uint256 _price) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) Activate(_price *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.Activate(&_UsernameRegistrar.TransactOpts, _price)
}

// ChangeController is a paid mutator transaction binding the contract method 0x3cebb823.
//
// Solidity: function changeController(address _newController) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) ChangeController(opts *bind.TransactOpts, _newController common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "changeController", _newController)
}

// ChangeController is a paid mutator transaction binding the contract method 0x3cebb823.
//
// Solidity: function changeController(address _newController) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) ChangeController(_newController common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.ChangeController(&_UsernameRegistrar.TransactOpts, _newController)
}

// ChangeController is a paid mutator transaction binding the contract method 0x3cebb823.
//
// Solidity: function changeController(address _newController) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) ChangeController(_newController common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.ChangeController(&_UsernameRegistrar.TransactOpts, _newController)
}

// DropUsername is a paid mutator transaction binding the contract method 0xf9e54282.
//
// Solidity: function dropUsername(bytes32 _label) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) DropUsername(opts *bind.TransactOpts, _label [32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "dropUsername", _label)
}

// DropUsername is a paid mutator transaction binding the contract method 0xf9e54282.
//
// Solidity: function dropUsername(bytes32 _label) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) DropUsername(_label [32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.DropUsername(&_UsernameRegistrar.TransactOpts, _label)
}

// DropUsername is a paid mutator transaction binding the contract method 0xf9e54282.
//
// Solidity: function dropUsername(bytes32 _label) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) DropUsername(_label [32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.DropUsername(&_UsernameRegistrar.TransactOpts, _label)
}

// EraseNode is a paid mutator transaction binding the contract method 0xde10f04b.
//
// Solidity: function eraseNode(bytes32[] _labels) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) EraseNode(opts *bind.TransactOpts, _labels [][32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "eraseNode", _labels)
}

// EraseNode is a paid mutator transaction binding the contract method 0xde10f04b.
//
// Solidity: function eraseNode(bytes32[] _labels) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) EraseNode(_labels [][32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.EraseNode(&_UsernameRegistrar.TransactOpts, _labels)
}

// EraseNode is a paid mutator transaction binding the contract method 0xde10f04b.
//
// Solidity: function eraseNode(bytes32[] _labels) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) EraseNode(_labels [][32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.EraseNode(&_UsernameRegistrar.TransactOpts, _labels)
}

// MigrateRegistry is a paid mutator transaction binding the contract method 0x98f038ff.
//
// Solidity: function migrateRegistry(uint256 _price) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) MigrateRegistry(opts *bind.TransactOpts, _price *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "migrateRegistry", _price)
}

// MigrateRegistry is a paid mutator transaction binding the contract method 0x98f038ff.
//
// Solidity: function migrateRegistry(uint256 _price) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) MigrateRegistry(_price *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.MigrateRegistry(&_UsernameRegistrar.TransactOpts, _price)
}

// MigrateRegistry is a paid mutator transaction binding the contract method 0x98f038ff.
//
// Solidity: function migrateRegistry(uint256 _price) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) MigrateRegistry(_price *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.MigrateRegistry(&_UsernameRegistrar.TransactOpts, _price)
}

// MigrateUsername is a paid mutator transaction binding the contract method 0x80cd0015.
//
// Solidity: function migrateUsername(bytes32 _label, uint256 _tokenBalance, uint256 _creationTime, address _accountOwner) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) MigrateUsername(opts *bind.TransactOpts, _label [32]byte, _tokenBalance *big.Int, _creationTime *big.Int, _accountOwner common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "migrateUsername", _label, _tokenBalance, _creationTime, _accountOwner)
}

// MigrateUsername is a paid mutator transaction binding the contract method 0x80cd0015.
//
// Solidity: function migrateUsername(bytes32 _label, uint256 _tokenBalance, uint256 _creationTime, address _accountOwner) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) MigrateUsername(_label [32]byte, _tokenBalance *big.Int, _creationTime *big.Int, _accountOwner common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.MigrateUsername(&_UsernameRegistrar.TransactOpts, _label, _tokenBalance, _creationTime, _accountOwner)
}

// MigrateUsername is a paid mutator transaction binding the contract method 0x80cd0015.
//
// Solidity: function migrateUsername(bytes32 _label, uint256 _tokenBalance, uint256 _creationTime, address _accountOwner) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) MigrateUsername(_label [32]byte, _tokenBalance *big.Int, _creationTime *big.Int, _accountOwner common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.MigrateUsername(&_UsernameRegistrar.TransactOpts, _label, _tokenBalance, _creationTime, _accountOwner)
}

// MoveAccount is a paid mutator transaction binding the contract method 0xc23e61b9.
//
// Solidity: function moveAccount(bytes32 _label, address _newRegistry) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) MoveAccount(opts *bind.TransactOpts, _label [32]byte, _newRegistry common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "moveAccount", _label, _newRegistry)
}

// MoveAccount is a paid mutator transaction binding the contract method 0xc23e61b9.
//
// Solidity: function moveAccount(bytes32 _label, address _newRegistry) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) MoveAccount(_label [32]byte, _newRegistry common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.MoveAccount(&_UsernameRegistrar.TransactOpts, _label, _newRegistry)
}

// MoveAccount is a paid mutator transaction binding the contract method 0xc23e61b9.
//
// Solidity: function moveAccount(bytes32 _label, address _newRegistry) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) MoveAccount(_label [32]byte, _newRegistry common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.MoveAccount(&_UsernameRegistrar.TransactOpts, _label, _newRegistry)
}

// MoveRegistry is a paid mutator transaction binding the contract method 0xe882c3ce.
//
// Solidity: function moveRegistry(address _newRegistry) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) MoveRegistry(opts *bind.TransactOpts, _newRegistry common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "moveRegistry", _newRegistry)
}

// MoveRegistry is a paid mutator transaction binding the contract method 0xe882c3ce.
//
// Solidity: function moveRegistry(address _newRegistry) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) MoveRegistry(_newRegistry common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.MoveRegistry(&_UsernameRegistrar.TransactOpts, _newRegistry)
}

// MoveRegistry is a paid mutator transaction binding the contract method 0xe882c3ce.
//
// Solidity: function moveRegistry(address _newRegistry) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) MoveRegistry(_newRegistry common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.MoveRegistry(&_UsernameRegistrar.TransactOpts, _newRegistry)
}

// ReceiveApproval is a paid mutator transaction binding the contract method 0x8f4ffcb1.
//
// Solidity: function receiveApproval(address _from, uint256 _amount, address _token, bytes _data) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) ReceiveApproval(opts *bind.TransactOpts, _from common.Address, _amount *big.Int, _token common.Address, _data []byte) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "receiveApproval", _from, _amount, _token, _data)
}

// ReceiveApproval is a paid mutator transaction binding the contract method 0x8f4ffcb1.
//
// Solidity: function receiveApproval(address _from, uint256 _amount, address _token, bytes _data) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) ReceiveApproval(_from common.Address, _amount *big.Int, _token common.Address, _data []byte) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.ReceiveApproval(&_UsernameRegistrar.TransactOpts, _from, _amount, _token, _data)
}

// ReceiveApproval is a paid mutator transaction binding the contract method 0x8f4ffcb1.
//
// Solidity: function receiveApproval(address _from, uint256 _amount, address _token, bytes _data) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) ReceiveApproval(_from common.Address, _amount *big.Int, _token common.Address, _data []byte) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.ReceiveApproval(&_UsernameRegistrar.TransactOpts, _from, _amount, _token, _data)
}

// Register is a paid mutator transaction binding the contract method 0xb82fedbb.
//
// Solidity: function register(bytes32 _label, address _account, bytes32 _pubkeyA, bytes32 _pubkeyB) returns(bytes32 namehash)
func (_UsernameRegistrar *UsernameRegistrarTransactor) Register(opts *bind.TransactOpts, _label [32]byte, _account common.Address, _pubkeyA [32]byte, _pubkeyB [32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "register", _label, _account, _pubkeyA, _pubkeyB)
}

// Register is a paid mutator transaction binding the contract method 0xb82fedbb.
//
// Solidity: function register(bytes32 _label, address _account, bytes32 _pubkeyA, bytes32 _pubkeyB) returns(bytes32 namehash)
func (_UsernameRegistrar *UsernameRegistrarSession) Register(_label [32]byte, _account common.Address, _pubkeyA [32]byte, _pubkeyB [32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.Register(&_UsernameRegistrar.TransactOpts, _label, _account, _pubkeyA, _pubkeyB)
}

// Register is a paid mutator transaction binding the contract method 0xb82fedbb.
//
// Solidity: function register(bytes32 _label, address _account, bytes32 _pubkeyA, bytes32 _pubkeyB) returns(bytes32 namehash)
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) Register(_label [32]byte, _account common.Address, _pubkeyA [32]byte, _pubkeyB [32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.Register(&_UsernameRegistrar.TransactOpts, _label, _account, _pubkeyA, _pubkeyB)
}

// Release is a paid mutator transaction binding the contract method 0x67d42a8b.
//
// Solidity: function release(bytes32 _label) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) Release(opts *bind.TransactOpts, _label [32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "release", _label)
}

// Release is a paid mutator transaction binding the contract method 0x67d42a8b.
//
// Solidity: function release(bytes32 _label) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) Release(_label [32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.Release(&_UsernameRegistrar.TransactOpts, _label)
}

// Release is a paid mutator transaction binding the contract method 0x67d42a8b.
//
// Solidity: function release(bytes32 _label) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) Release(_label [32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.Release(&_UsernameRegistrar.TransactOpts, _label)
}

// ReserveSlash is a paid mutator transaction binding the contract method 0x05c24481.
//
// Solidity: function reserveSlash(bytes32 _secret) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) ReserveSlash(opts *bind.TransactOpts, _secret [32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "reserveSlash", _secret)
}

// ReserveSlash is a paid mutator transaction binding the contract method 0x05c24481.
//
// Solidity: function reserveSlash(bytes32 _secret) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) ReserveSlash(_secret [32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.ReserveSlash(&_UsernameRegistrar.TransactOpts, _secret)
}

// ReserveSlash is a paid mutator transaction binding the contract method 0x05c24481.
//
// Solidity: function reserveSlash(bytes32 _secret) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) ReserveSlash(_secret [32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.ReserveSlash(&_UsernameRegistrar.TransactOpts, _secret)
}

// SetResolver is a paid mutator transaction binding the contract method 0x4e543b26.
//
// Solidity: function setResolver(address _resolver) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) SetResolver(opts *bind.TransactOpts, _resolver common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "setResolver", _resolver)
}

// SetResolver is a paid mutator transaction binding the contract method 0x4e543b26.
//
// Solidity: function setResolver(address _resolver) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) SetResolver(_resolver common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.SetResolver(&_UsernameRegistrar.TransactOpts, _resolver)
}

// SetResolver is a paid mutator transaction binding the contract method 0x4e543b26.
//
// Solidity: function setResolver(address _resolver) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) SetResolver(_resolver common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.SetResolver(&_UsernameRegistrar.TransactOpts, _resolver)
}

// SlashAddressLikeUsername is a paid mutator transaction binding the contract method 0x8cf7b7a4.
//
// Solidity: function slashAddressLikeUsername(string _username, uint256 _reserveSecret) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) SlashAddressLikeUsername(opts *bind.TransactOpts, _username string, _reserveSecret *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "slashAddressLikeUsername", _username, _reserveSecret)
}

// SlashAddressLikeUsername is a paid mutator transaction binding the contract method 0x8cf7b7a4.
//
// Solidity: function slashAddressLikeUsername(string _username, uint256 _reserveSecret) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) SlashAddressLikeUsername(_username string, _reserveSecret *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.SlashAddressLikeUsername(&_UsernameRegistrar.TransactOpts, _username, _reserveSecret)
}

// SlashAddressLikeUsername is a paid mutator transaction binding the contract method 0x8cf7b7a4.
//
// Solidity: function slashAddressLikeUsername(string _username, uint256 _reserveSecret) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) SlashAddressLikeUsername(_username string, _reserveSecret *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.SlashAddressLikeUsername(&_UsernameRegistrar.TransactOpts, _username, _reserveSecret)
}

// SlashInvalidUsername is a paid mutator transaction binding the contract method 0x40784ebd.
//
// Solidity: function slashInvalidUsername(string _username, uint256 _offendingPos, uint256 _reserveSecret) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) SlashInvalidUsername(opts *bind.TransactOpts, _username string, _offendingPos *big.Int, _reserveSecret *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "slashInvalidUsername", _username, _offendingPos, _reserveSecret)
}

// SlashInvalidUsername is a paid mutator transaction binding the contract method 0x40784ebd.
//
// Solidity: function slashInvalidUsername(string _username, uint256 _offendingPos, uint256 _reserveSecret) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) SlashInvalidUsername(_username string, _offendingPos *big.Int, _reserveSecret *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.SlashInvalidUsername(&_UsernameRegistrar.TransactOpts, _username, _offendingPos, _reserveSecret)
}

// SlashInvalidUsername is a paid mutator transaction binding the contract method 0x40784ebd.
//
// Solidity: function slashInvalidUsername(string _username, uint256 _offendingPos, uint256 _reserveSecret) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) SlashInvalidUsername(_username string, _offendingPos *big.Int, _reserveSecret *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.SlashInvalidUsername(&_UsernameRegistrar.TransactOpts, _username, _offendingPos, _reserveSecret)
}

// SlashReservedUsername is a paid mutator transaction binding the contract method 0x40b1ad52.
//
// Solidity: function slashReservedUsername(string _username, bytes32[] _proof, uint256 _reserveSecret) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) SlashReservedUsername(opts *bind.TransactOpts, _username string, _proof [][32]byte, _reserveSecret *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "slashReservedUsername", _username, _proof, _reserveSecret)
}

// SlashReservedUsername is a paid mutator transaction binding the contract method 0x40b1ad52.
//
// Solidity: function slashReservedUsername(string _username, bytes32[] _proof, uint256 _reserveSecret) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) SlashReservedUsername(_username string, _proof [][32]byte, _reserveSecret *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.SlashReservedUsername(&_UsernameRegistrar.TransactOpts, _username, _proof, _reserveSecret)
}

// SlashReservedUsername is a paid mutator transaction binding the contract method 0x40b1ad52.
//
// Solidity: function slashReservedUsername(string _username, bytes32[] _proof, uint256 _reserveSecret) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) SlashReservedUsername(_username string, _proof [][32]byte, _reserveSecret *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.SlashReservedUsername(&_UsernameRegistrar.TransactOpts, _username, _proof, _reserveSecret)
}

// SlashSmallUsername is a paid mutator transaction binding the contract method 0x96bba9a8.
//
// Solidity: function slashSmallUsername(string _username, uint256 _reserveSecret) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) SlashSmallUsername(opts *bind.TransactOpts, _username string, _reserveSecret *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "slashSmallUsername", _username, _reserveSecret)
}

// SlashSmallUsername is a paid mutator transaction binding the contract method 0x96bba9a8.
//
// Solidity: function slashSmallUsername(string _username, uint256 _reserveSecret) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) SlashSmallUsername(_username string, _reserveSecret *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.SlashSmallUsername(&_UsernameRegistrar.TransactOpts, _username, _reserveSecret)
}

// SlashSmallUsername is a paid mutator transaction binding the contract method 0x96bba9a8.
//
// Solidity: function slashSmallUsername(string _username, uint256 _reserveSecret) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) SlashSmallUsername(_username string, _reserveSecret *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.SlashSmallUsername(&_UsernameRegistrar.TransactOpts, _username, _reserveSecret)
}

// UpdateAccountOwner is a paid mutator transaction binding the contract method 0x32e1ed24.
//
// Solidity: function updateAccountOwner(bytes32 _label) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) UpdateAccountOwner(opts *bind.TransactOpts, _label [32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "updateAccountOwner", _label)
}

// UpdateAccountOwner is a paid mutator transaction binding the contract method 0x32e1ed24.
//
// Solidity: function updateAccountOwner(bytes32 _label) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) UpdateAccountOwner(_label [32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.UpdateAccountOwner(&_UsernameRegistrar.TransactOpts, _label)
}

// UpdateAccountOwner is a paid mutator transaction binding the contract method 0x32e1ed24.
//
// Solidity: function updateAccountOwner(bytes32 _label) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) UpdateAccountOwner(_label [32]byte) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.UpdateAccountOwner(&_UsernameRegistrar.TransactOpts, _label)
}

// UpdateRegistryPrice is a paid mutator transaction binding the contract method 0x860e9b0f.
//
// Solidity: function updateRegistryPrice(uint256 _price) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) UpdateRegistryPrice(opts *bind.TransactOpts, _price *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "updateRegistryPrice", _price)
}

// UpdateRegistryPrice is a paid mutator transaction binding the contract method 0x860e9b0f.
//
// Solidity: function updateRegistryPrice(uint256 _price) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) UpdateRegistryPrice(_price *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.UpdateRegistryPrice(&_UsernameRegistrar.TransactOpts, _price)
}

// UpdateRegistryPrice is a paid mutator transaction binding the contract method 0x860e9b0f.
//
// Solidity: function updateRegistryPrice(uint256 _price) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) UpdateRegistryPrice(_price *big.Int) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.UpdateRegistryPrice(&_UsernameRegistrar.TransactOpts, _price)
}

// WithdrawExcessBalance is a paid mutator transaction binding the contract method 0x307c7a0d.
//
// Solidity: function withdrawExcessBalance(address _token, address _beneficiary) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) WithdrawExcessBalance(opts *bind.TransactOpts, _token common.Address, _beneficiary common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "withdrawExcessBalance", _token, _beneficiary)
}

// WithdrawExcessBalance is a paid mutator transaction binding the contract method 0x307c7a0d.
//
// Solidity: function withdrawExcessBalance(address _token, address _beneficiary) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) WithdrawExcessBalance(_token common.Address, _beneficiary common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.WithdrawExcessBalance(&_UsernameRegistrar.TransactOpts, _token, _beneficiary)
}

// WithdrawExcessBalance is a paid mutator transaction binding the contract method 0x307c7a0d.
//
// Solidity: function withdrawExcessBalance(address _token, address _beneficiary) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) WithdrawExcessBalance(_token common.Address, _beneficiary common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.WithdrawExcessBalance(&_UsernameRegistrar.TransactOpts, _token, _beneficiary)
}

// WithdrawWrongNode is a paid mutator transaction binding the contract method 0xafe12e77.
//
// Solidity: function withdrawWrongNode(bytes32 _domainHash, address _beneficiary) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactor) WithdrawWrongNode(opts *bind.TransactOpts, _domainHash [32]byte, _beneficiary common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.contract.Transact(opts, "withdrawWrongNode", _domainHash, _beneficiary)
}

// WithdrawWrongNode is a paid mutator transaction binding the contract method 0xafe12e77.
//
// Solidity: function withdrawWrongNode(bytes32 _domainHash, address _beneficiary) returns()
func (_UsernameRegistrar *UsernameRegistrarSession) WithdrawWrongNode(_domainHash [32]byte, _beneficiary common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.WithdrawWrongNode(&_UsernameRegistrar.TransactOpts, _domainHash, _beneficiary)
}

// WithdrawWrongNode is a paid mutator transaction binding the contract method 0xafe12e77.
//
// Solidity: function withdrawWrongNode(bytes32 _domainHash, address _beneficiary) returns()
func (_UsernameRegistrar *UsernameRegistrarTransactorSession) WithdrawWrongNode(_domainHash [32]byte, _beneficiary common.Address) (*types.Transaction, error) {
	return _UsernameRegistrar.Contract.WithdrawWrongNode(&_UsernameRegistrar.TransactOpts, _domainHash, _beneficiary)
}

// UsernameRegistrarRegistryMovedIterator is returned from FilterRegistryMoved and is used to iterate over the raw logs and unpacked data for RegistryMoved events raised by the UsernameRegistrar contract.
type UsernameRegistrarRegistryMovedIterator struct {
	Event *UsernameRegistrarRegistryMoved // Event containing the contract specifics and raw log

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
func (it *UsernameRegistrarRegistryMovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UsernameRegistrarRegistryMoved)
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
		it.Event = new(UsernameRegistrarRegistryMoved)
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
func (it *UsernameRegistrarRegistryMovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UsernameRegistrarRegistryMovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UsernameRegistrarRegistryMoved represents a RegistryMoved event raised by the UsernameRegistrar contract.
type UsernameRegistrarRegistryMoved struct {
	NewRegistry common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterRegistryMoved is a free log retrieval operation binding the contract event 0xce0afb4c27dbd57a3646e2d639557521bfb05a42dc0ec50f9c1fe13d92e3e6d6.
//
// Solidity: event RegistryMoved(address newRegistry)
func (_UsernameRegistrar *UsernameRegistrarFilterer) FilterRegistryMoved(opts *bind.FilterOpts) (*UsernameRegistrarRegistryMovedIterator, error) {

	logs, sub, err := _UsernameRegistrar.contract.FilterLogs(opts, "RegistryMoved")
	if err != nil {
		return nil, err
	}
	return &UsernameRegistrarRegistryMovedIterator{contract: _UsernameRegistrar.contract, event: "RegistryMoved", logs: logs, sub: sub}, nil
}

// WatchRegistryMoved is a free log subscription operation binding the contract event 0xce0afb4c27dbd57a3646e2d639557521bfb05a42dc0ec50f9c1fe13d92e3e6d6.
//
// Solidity: event RegistryMoved(address newRegistry)
func (_UsernameRegistrar *UsernameRegistrarFilterer) WatchRegistryMoved(opts *bind.WatchOpts, sink chan<- *UsernameRegistrarRegistryMoved) (event.Subscription, error) {

	logs, sub, err := _UsernameRegistrar.contract.WatchLogs(opts, "RegistryMoved")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UsernameRegistrarRegistryMoved)
				if err := _UsernameRegistrar.contract.UnpackLog(event, "RegistryMoved", log); err != nil {
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

// ParseRegistryMoved is a log parse operation binding the contract event 0xce0afb4c27dbd57a3646e2d639557521bfb05a42dc0ec50f9c1fe13d92e3e6d6.
//
// Solidity: event RegistryMoved(address newRegistry)
func (_UsernameRegistrar *UsernameRegistrarFilterer) ParseRegistryMoved(log types.Log) (*UsernameRegistrarRegistryMoved, error) {
	event := new(UsernameRegistrarRegistryMoved)
	if err := _UsernameRegistrar.contract.UnpackLog(event, "RegistryMoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// UsernameRegistrarRegistryPriceIterator is returned from FilterRegistryPrice and is used to iterate over the raw logs and unpacked data for RegistryPrice events raised by the UsernameRegistrar contract.
type UsernameRegistrarRegistryPriceIterator struct {
	Event *UsernameRegistrarRegistryPrice // Event containing the contract specifics and raw log

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
func (it *UsernameRegistrarRegistryPriceIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UsernameRegistrarRegistryPrice)
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
		it.Event = new(UsernameRegistrarRegistryPrice)
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
func (it *UsernameRegistrarRegistryPriceIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UsernameRegistrarRegistryPriceIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UsernameRegistrarRegistryPrice represents a RegistryPrice event raised by the UsernameRegistrar contract.
type UsernameRegistrarRegistryPrice struct {
	Price *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterRegistryPrice is a free log retrieval operation binding the contract event 0x45d3cd7c7bd7d211f00610f51660b2f114c7833e0c52ef3603c6d41ed07a7458.
//
// Solidity: event RegistryPrice(uint256 price)
func (_UsernameRegistrar *UsernameRegistrarFilterer) FilterRegistryPrice(opts *bind.FilterOpts) (*UsernameRegistrarRegistryPriceIterator, error) {

	logs, sub, err := _UsernameRegistrar.contract.FilterLogs(opts, "RegistryPrice")
	if err != nil {
		return nil, err
	}
	return &UsernameRegistrarRegistryPriceIterator{contract: _UsernameRegistrar.contract, event: "RegistryPrice", logs: logs, sub: sub}, nil
}

// WatchRegistryPrice is a free log subscription operation binding the contract event 0x45d3cd7c7bd7d211f00610f51660b2f114c7833e0c52ef3603c6d41ed07a7458.
//
// Solidity: event RegistryPrice(uint256 price)
func (_UsernameRegistrar *UsernameRegistrarFilterer) WatchRegistryPrice(opts *bind.WatchOpts, sink chan<- *UsernameRegistrarRegistryPrice) (event.Subscription, error) {

	logs, sub, err := _UsernameRegistrar.contract.WatchLogs(opts, "RegistryPrice")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UsernameRegistrarRegistryPrice)
				if err := _UsernameRegistrar.contract.UnpackLog(event, "RegistryPrice", log); err != nil {
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

// ParseRegistryPrice is a log parse operation binding the contract event 0x45d3cd7c7bd7d211f00610f51660b2f114c7833e0c52ef3603c6d41ed07a7458.
//
// Solidity: event RegistryPrice(uint256 price)
func (_UsernameRegistrar *UsernameRegistrarFilterer) ParseRegistryPrice(log types.Log) (*UsernameRegistrarRegistryPrice, error) {
	event := new(UsernameRegistrarRegistryPrice)
	if err := _UsernameRegistrar.contract.UnpackLog(event, "RegistryPrice", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// UsernameRegistrarRegistryStateIterator is returned from FilterRegistryState and is used to iterate over the raw logs and unpacked data for RegistryState events raised by the UsernameRegistrar contract.
type UsernameRegistrarRegistryStateIterator struct {
	Event *UsernameRegistrarRegistryState // Event containing the contract specifics and raw log

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
func (it *UsernameRegistrarRegistryStateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UsernameRegistrarRegistryState)
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
		it.Event = new(UsernameRegistrarRegistryState)
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
func (it *UsernameRegistrarRegistryStateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UsernameRegistrarRegistryStateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UsernameRegistrarRegistryState represents a RegistryState event raised by the UsernameRegistrar contract.
type UsernameRegistrarRegistryState struct {
	State uint8
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterRegistryState is a free log retrieval operation binding the contract event 0xee85d4d9a9722e814f07db07f29734cd5a97e0e58781ad41ae4572193b1caea0.
//
// Solidity: event RegistryState(uint8 state)
func (_UsernameRegistrar *UsernameRegistrarFilterer) FilterRegistryState(opts *bind.FilterOpts) (*UsernameRegistrarRegistryStateIterator, error) {

	logs, sub, err := _UsernameRegistrar.contract.FilterLogs(opts, "RegistryState")
	if err != nil {
		return nil, err
	}
	return &UsernameRegistrarRegistryStateIterator{contract: _UsernameRegistrar.contract, event: "RegistryState", logs: logs, sub: sub}, nil
}

// WatchRegistryState is a free log subscription operation binding the contract event 0xee85d4d9a9722e814f07db07f29734cd5a97e0e58781ad41ae4572193b1caea0.
//
// Solidity: event RegistryState(uint8 state)
func (_UsernameRegistrar *UsernameRegistrarFilterer) WatchRegistryState(opts *bind.WatchOpts, sink chan<- *UsernameRegistrarRegistryState) (event.Subscription, error) {

	logs, sub, err := _UsernameRegistrar.contract.WatchLogs(opts, "RegistryState")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UsernameRegistrarRegistryState)
				if err := _UsernameRegistrar.contract.UnpackLog(event, "RegistryState", log); err != nil {
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

// ParseRegistryState is a log parse operation binding the contract event 0xee85d4d9a9722e814f07db07f29734cd5a97e0e58781ad41ae4572193b1caea0.
//
// Solidity: event RegistryState(uint8 state)
func (_UsernameRegistrar *UsernameRegistrarFilterer) ParseRegistryState(log types.Log) (*UsernameRegistrarRegistryState, error) {
	event := new(UsernameRegistrarRegistryState)
	if err := _UsernameRegistrar.contract.UnpackLog(event, "RegistryState", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// UsernameRegistrarUsernameOwnerIterator is returned from FilterUsernameOwner and is used to iterate over the raw logs and unpacked data for UsernameOwner events raised by the UsernameRegistrar contract.
type UsernameRegistrarUsernameOwnerIterator struct {
	Event *UsernameRegistrarUsernameOwner // Event containing the contract specifics and raw log

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
func (it *UsernameRegistrarUsernameOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UsernameRegistrarUsernameOwner)
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
		it.Event = new(UsernameRegistrarUsernameOwner)
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
func (it *UsernameRegistrarUsernameOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UsernameRegistrarUsernameOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UsernameRegistrarUsernameOwner represents a UsernameOwner event raised by the UsernameRegistrar contract.
type UsernameRegistrarUsernameOwner struct {
	NameHash [32]byte
	Owner    common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterUsernameOwner is a free log retrieval operation binding the contract event 0xd2da4206c3fa95b8fc1ee48627023d322b59cc7218e14cb95cf0c0fe562f2e4d.
//
// Solidity: event UsernameOwner(bytes32 indexed nameHash, address owner)
func (_UsernameRegistrar *UsernameRegistrarFilterer) FilterUsernameOwner(opts *bind.FilterOpts, nameHash [][32]byte) (*UsernameRegistrarUsernameOwnerIterator, error) {

	var nameHashRule []interface{}
	for _, nameHashItem := range nameHash {
		nameHashRule = append(nameHashRule, nameHashItem)
	}

	logs, sub, err := _UsernameRegistrar.contract.FilterLogs(opts, "UsernameOwner", nameHashRule)
	if err != nil {
		return nil, err
	}
	return &UsernameRegistrarUsernameOwnerIterator{contract: _UsernameRegistrar.contract, event: "UsernameOwner", logs: logs, sub: sub}, nil
}

// WatchUsernameOwner is a free log subscription operation binding the contract event 0xd2da4206c3fa95b8fc1ee48627023d322b59cc7218e14cb95cf0c0fe562f2e4d.
//
// Solidity: event UsernameOwner(bytes32 indexed nameHash, address owner)
func (_UsernameRegistrar *UsernameRegistrarFilterer) WatchUsernameOwner(opts *bind.WatchOpts, sink chan<- *UsernameRegistrarUsernameOwner, nameHash [][32]byte) (event.Subscription, error) {

	var nameHashRule []interface{}
	for _, nameHashItem := range nameHash {
		nameHashRule = append(nameHashRule, nameHashItem)
	}

	logs, sub, err := _UsernameRegistrar.contract.WatchLogs(opts, "UsernameOwner", nameHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UsernameRegistrarUsernameOwner)
				if err := _UsernameRegistrar.contract.UnpackLog(event, "UsernameOwner", log); err != nil {
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

// ParseUsernameOwner is a log parse operation binding the contract event 0xd2da4206c3fa95b8fc1ee48627023d322b59cc7218e14cb95cf0c0fe562f2e4d.
//
// Solidity: event UsernameOwner(bytes32 indexed nameHash, address owner)
func (_UsernameRegistrar *UsernameRegistrarFilterer) ParseUsernameOwner(log types.Log) (*UsernameRegistrarUsernameOwner, error) {
	event := new(UsernameRegistrarUsernameOwner)
	if err := _UsernameRegistrar.contract.UnpackLog(event, "UsernameOwner", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
