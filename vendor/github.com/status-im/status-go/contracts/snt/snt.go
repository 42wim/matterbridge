// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package snt

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
const ControlledABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_newController\",\"type\":\"address\"}],\"name\":\"changeController\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"controller\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]"

// ControlledFuncSigs maps the 4-byte function signature to its string representation.
var ControlledFuncSigs = map[string]string{
	"3cebb823": "changeController(address)",
	"f77c4791": "controller()",
}

// ControlledBin is the compiled bytecode used for deploying new contracts.
var ControlledBin = "0x608060405234801561001057600080fd5b50610100806100206000396000f30060806040526004361060485763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416633cebb8238114604d578063f77c479114607a575b600080fd5b348015605857600080fd5b50607873ffffffffffffffffffffffffffffffffffffffff6004351660b5565b005b348015608557600080fd5b50608c60b8565b6040805173ffffffffffffffffffffffffffffffffffffffff9092168252519081900360200190f35b50565b60005473ffffffffffffffffffffffffffffffffffffffff16815600a165627a7a72305820fe8e3aeccba397cd6ed181c46c148a5ce1f4c75e6b15416b622abe4b1cf81a3d0029"

// DeployControlled deploys a new Ethereum contract, binding an instance of Controlled to it.
func DeployControlled(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Controlled, error) {
	parsed, err := abi.JSON(strings.NewReader(ControlledABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ControlledBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Controlled{ControlledCaller: ControlledCaller{contract: contract}, ControlledTransactor: ControlledTransactor{contract: contract}, ControlledFilterer: ControlledFilterer{contract: contract}}, nil
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

// ERC20TokenABI is the input ABI used to generate the binding from.
const ERC20TokenABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"remaining\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"

// ERC20TokenFuncSigs maps the 4-byte function signature to its string representation.
var ERC20TokenFuncSigs = map[string]string{
	"dd62ed3e": "allowance(address,address)",
	"095ea7b3": "approve(address,uint256)",
	"70a08231": "balanceOf(address)",
	"18160ddd": "totalSupply()",
	"a9059cbb": "transfer(address,uint256)",
	"23b872dd": "transferFrom(address,address,uint256)",
}

// ERC20Token is an auto generated Go binding around an Ethereum contract.
type ERC20Token struct {
	ERC20TokenCaller     // Read-only binding to the contract
	ERC20TokenTransactor // Write-only binding to the contract
	ERC20TokenFilterer   // Log filterer for contract events
}

// ERC20TokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type ERC20TokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20TokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ERC20TokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20TokenFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ERC20TokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20TokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ERC20TokenSession struct {
	Contract     *ERC20Token       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ERC20TokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ERC20TokenCallerSession struct {
	Contract *ERC20TokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// ERC20TokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ERC20TokenTransactorSession struct {
	Contract     *ERC20TokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// ERC20TokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type ERC20TokenRaw struct {
	Contract *ERC20Token // Generic contract binding to access the raw methods on
}

// ERC20TokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ERC20TokenCallerRaw struct {
	Contract *ERC20TokenCaller // Generic read-only contract binding to access the raw methods on
}

// ERC20TokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ERC20TokenTransactorRaw struct {
	Contract *ERC20TokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewERC20Token creates a new instance of ERC20Token, bound to a specific deployed contract.
func NewERC20Token(address common.Address, backend bind.ContractBackend) (*ERC20Token, error) {
	contract, err := bindERC20Token(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ERC20Token{ERC20TokenCaller: ERC20TokenCaller{contract: contract}, ERC20TokenTransactor: ERC20TokenTransactor{contract: contract}, ERC20TokenFilterer: ERC20TokenFilterer{contract: contract}}, nil
}

// NewERC20TokenCaller creates a new read-only instance of ERC20Token, bound to a specific deployed contract.
func NewERC20TokenCaller(address common.Address, caller bind.ContractCaller) (*ERC20TokenCaller, error) {
	contract, err := bindERC20Token(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ERC20TokenCaller{contract: contract}, nil
}

// NewERC20TokenTransactor creates a new write-only instance of ERC20Token, bound to a specific deployed contract.
func NewERC20TokenTransactor(address common.Address, transactor bind.ContractTransactor) (*ERC20TokenTransactor, error) {
	contract, err := bindERC20Token(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ERC20TokenTransactor{contract: contract}, nil
}

// NewERC20TokenFilterer creates a new log filterer instance of ERC20Token, bound to a specific deployed contract.
func NewERC20TokenFilterer(address common.Address, filterer bind.ContractFilterer) (*ERC20TokenFilterer, error) {
	contract, err := bindERC20Token(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ERC20TokenFilterer{contract: contract}, nil
}

// bindERC20Token binds a generic wrapper to an already deployed contract.
func bindERC20Token(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ERC20TokenABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ERC20Token *ERC20TokenRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ERC20Token.Contract.ERC20TokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ERC20Token *ERC20TokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20Token.Contract.ERC20TokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ERC20Token *ERC20TokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ERC20Token.Contract.ERC20TokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ERC20Token *ERC20TokenCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ERC20Token.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ERC20Token *ERC20TokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20Token.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ERC20Token *ERC20TokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ERC20Token.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address _spender) view returns(uint256 remaining)
func (_ERC20Token *ERC20TokenCaller) Allowance(opts *bind.CallOpts, _owner common.Address, _spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ERC20Token.contract.Call(opts, &out, "allowance", _owner, _spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address _spender) view returns(uint256 remaining)
func (_ERC20Token *ERC20TokenSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _ERC20Token.Contract.Allowance(&_ERC20Token.CallOpts, _owner, _spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address _spender) view returns(uint256 remaining)
func (_ERC20Token *ERC20TokenCallerSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _ERC20Token.Contract.Allowance(&_ERC20Token.CallOpts, _owner, _spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256 balance)
func (_ERC20Token *ERC20TokenCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ERC20Token.contract.Call(opts, &out, "balanceOf", _owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256 balance)
func (_ERC20Token *ERC20TokenSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _ERC20Token.Contract.BalanceOf(&_ERC20Token.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256 balance)
func (_ERC20Token *ERC20TokenCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _ERC20Token.Contract.BalanceOf(&_ERC20Token.CallOpts, _owner)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_ERC20Token *ERC20TokenCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ERC20Token.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_ERC20Token *ERC20TokenSession) TotalSupply() (*big.Int, error) {
	return _ERC20Token.Contract.TotalSupply(&_ERC20Token.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_ERC20Token *ERC20TokenCallerSession) TotalSupply() (*big.Int, error) {
	return _ERC20Token.Contract.TotalSupply(&_ERC20Token.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _value) returns(bool success)
func (_ERC20Token *ERC20TokenTransactor) Approve(opts *bind.TransactOpts, _spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _ERC20Token.contract.Transact(opts, "approve", _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _value) returns(bool success)
func (_ERC20Token *ERC20TokenSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _ERC20Token.Contract.Approve(&_ERC20Token.TransactOpts, _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _value) returns(bool success)
func (_ERC20Token *ERC20TokenTransactorSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _ERC20Token.Contract.Approve(&_ERC20Token.TransactOpts, _spender, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _value) returns(bool success)
func (_ERC20Token *ERC20TokenTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _ERC20Token.contract.Transact(opts, "transfer", _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _value) returns(bool success)
func (_ERC20Token *ERC20TokenSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _ERC20Token.Contract.Transfer(&_ERC20Token.TransactOpts, _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _value) returns(bool success)
func (_ERC20Token *ERC20TokenTransactorSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _ERC20Token.Contract.Transfer(&_ERC20Token.TransactOpts, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _value) returns(bool success)
func (_ERC20Token *ERC20TokenTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _ERC20Token.contract.Transact(opts, "transferFrom", _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _value) returns(bool success)
func (_ERC20Token *ERC20TokenSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _ERC20Token.Contract.TransferFrom(&_ERC20Token.TransactOpts, _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _value) returns(bool success)
func (_ERC20Token *ERC20TokenTransactorSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _ERC20Token.Contract.TransferFrom(&_ERC20Token.TransactOpts, _from, _to, _value)
}

// ERC20TokenApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the ERC20Token contract.
type ERC20TokenApprovalIterator struct {
	Event *ERC20TokenApproval // Event containing the contract specifics and raw log

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
func (it *ERC20TokenApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20TokenApproval)
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
		it.Event = new(ERC20TokenApproval)
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
func (it *ERC20TokenApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20TokenApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20TokenApproval represents a Approval event raised by the ERC20Token contract.
type ERC20TokenApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed _owner, address indexed _spender, uint256 _value)
func (_ERC20Token *ERC20TokenFilterer) FilterApproval(opts *bind.FilterOpts, _owner []common.Address, _spender []common.Address) (*ERC20TokenApprovalIterator, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _spenderRule []interface{}
	for _, _spenderItem := range _spender {
		_spenderRule = append(_spenderRule, _spenderItem)
	}

	logs, sub, err := _ERC20Token.contract.FilterLogs(opts, "Approval", _ownerRule, _spenderRule)
	if err != nil {
		return nil, err
	}
	return &ERC20TokenApprovalIterator{contract: _ERC20Token.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed _owner, address indexed _spender, uint256 _value)
func (_ERC20Token *ERC20TokenFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *ERC20TokenApproval, _owner []common.Address, _spender []common.Address) (event.Subscription, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _spenderRule []interface{}
	for _, _spenderItem := range _spender {
		_spenderRule = append(_spenderRule, _spenderItem)
	}

	logs, sub, err := _ERC20Token.contract.WatchLogs(opts, "Approval", _ownerRule, _spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20TokenApproval)
				if err := _ERC20Token.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed _owner, address indexed _spender, uint256 _value)
func (_ERC20Token *ERC20TokenFilterer) ParseApproval(log types.Log) (*ERC20TokenApproval, error) {
	event := new(ERC20TokenApproval)
	if err := _ERC20Token.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ERC20TokenTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the ERC20Token contract.
type ERC20TokenTransferIterator struct {
	Event *ERC20TokenTransfer // Event containing the contract specifics and raw log

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
func (it *ERC20TokenTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20TokenTransfer)
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
		it.Event = new(ERC20TokenTransfer)
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
func (it *ERC20TokenTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20TokenTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20TokenTransfer represents a Transfer event raised by the ERC20Token contract.
type ERC20TokenTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 _value)
func (_ERC20Token *ERC20TokenFilterer) FilterTransfer(opts *bind.FilterOpts, _from []common.Address, _to []common.Address) (*ERC20TokenTransferIterator, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}

	logs, sub, err := _ERC20Token.contract.FilterLogs(opts, "Transfer", _fromRule, _toRule)
	if err != nil {
		return nil, err
	}
	return &ERC20TokenTransferIterator{contract: _ERC20Token.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 _value)
func (_ERC20Token *ERC20TokenFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *ERC20TokenTransfer, _from []common.Address, _to []common.Address) (event.Subscription, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}

	logs, sub, err := _ERC20Token.contract.WatchLogs(opts, "Transfer", _fromRule, _toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20TokenTransfer)
				if err := _ERC20Token.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 _value)
func (_ERC20Token *ERC20TokenFilterer) ParseTransfer(log types.Log) (*ERC20TokenTransfer, error) {
	event := new(ERC20TokenTransfer)
	if err := _ERC20Token.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MiniMeTokenABI is the input ABI used to generate the binding from.
const MiniMeTokenABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"creationBlock\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newController\",\"type\":\"address\"}],\"name\":\"changeController\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_blockNumber\",\"type\":\"uint256\"}],\"name\":\"balanceOfAt\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_cloneTokenName\",\"type\":\"string\"},{\"name\":\"_cloneDecimalUnits\",\"type\":\"uint8\"},{\"name\":\"_cloneTokenSymbol\",\"type\":\"string\"},{\"name\":\"_snapshotBlock\",\"type\":\"uint256\"},{\"name\":\"_transfersEnabled\",\"type\":\"bool\"}],\"name\":\"createCloneToken\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"parentToken\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"generateTokens\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_blockNumber\",\"type\":\"uint256\"}],\"name\":\"totalSupplyAt\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"transfersEnabled\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"parentSnapShotBlock\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"},{\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"approveAndCall\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"destroyTokens\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"remaining\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"claimTokens\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"tokenFactory\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_transfersEnabled\",\"type\":\"bool\"}],\"name\":\"enableTransfers\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"controller\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_tokenFactory\",\"type\":\"address\"},{\"name\":\"_parentToken\",\"type\":\"address\"},{\"name\":\"_parentSnapShotBlock\",\"type\":\"uint256\"},{\"name\":\"_tokenName\",\"type\":\"string\"},{\"name\":\"_decimalUnits\",\"type\":\"uint8\"},{\"name\":\"_tokenSymbol\",\"type\":\"string\"},{\"name\":\"_transfersEnabled\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_controller\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"ClaimedTokens\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_cloneToken\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_snapshotBlock\",\"type\":\"uint256\"}],\"name\":\"NewCloneToken\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"

// MiniMeTokenFuncSigs maps the 4-byte function signature to its string representation.
var MiniMeTokenFuncSigs = map[string]string{
	"dd62ed3e": "allowance(address,address)",
	"095ea7b3": "approve(address,uint256)",
	"cae9ca51": "approveAndCall(address,uint256,bytes)",
	"70a08231": "balanceOf(address)",
	"4ee2cd7e": "balanceOfAt(address,uint256)",
	"3cebb823": "changeController(address)",
	"df8de3e7": "claimTokens(address)",
	"f77c4791": "controller()",
	"6638c087": "createCloneToken(string,uint8,string,uint256,bool)",
	"17634514": "creationBlock()",
	"313ce567": "decimals()",
	"d3ce77fe": "destroyTokens(address,uint256)",
	"f41e60c5": "enableTransfers(bool)",
	"827f32c0": "generateTokens(address,uint256)",
	"06fdde03": "name()",
	"c5bcc4f1": "parentSnapShotBlock()",
	"80a54001": "parentToken()",
	"95d89b41": "symbol()",
	"e77772fe": "tokenFactory()",
	"18160ddd": "totalSupply()",
	"981b24d0": "totalSupplyAt(uint256)",
	"a9059cbb": "transfer(address,uint256)",
	"23b872dd": "transferFrom(address,address,uint256)",
	"bef97c87": "transfersEnabled()",
	"54fd4d50": "version()",
}

// MiniMeTokenBin is the compiled bytecode used for deploying new contracts.
var MiniMeTokenBin = "0x60c0604052600760808190527f4d4d545f302e310000000000000000000000000000000000000000000000000060a090815261003e9160049190610063565b5034801561004b57600080fd5b506040516108113803806108118339016040526100fe565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106100a457805160ff19168380011785556100d1565b828001600101855582156100d1579182015b828111156100d15782518255916020019190600101906100b6565b506100dd9291506100e1565b5090565b6100fb91905b808211156100dd57600081556001016100e7565b90565b6107048061010d6000396000f3006080604052600436106101485763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306fdde03811461014a578063095ea7b3146101d4578063176345141461020c57806318160ddd1461023357806323b872dd14610248578063313ce567146102725780633cebb8231461029d5780634ee2cd7e146102be57806354fd4d50146102e25780636638c087146102f757806370a08231146103ba57806380a54001146103db578063827f32c0146101d457806395d89b41146103f0578063981b24d014610405578063a9059cbb146101d4578063bef97c871461041d578063c5bcc4f114610432578063cae9ca5114610447578063d3ce77fe146101d4578063dd62ed3e146104b0578063df8de3e71461029d578063e77772fe146104d7578063f41e60c5146104ec578063f77c479114610506575b005b34801561015657600080fd5b5061015f61051b565b6040805160208082528351818301528351919283929083019185019080838360005b83811015610199578181015183820152602001610181565b50505050905090810190601f1680156101c65780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b3480156101e057600080fd5b506101f8600160a060020a03600435166024356105a8565b604080519115158252519081900360200190f35b34801561021857600080fd5b506102216105b0565b60408051918252519081900360200190f35b34801561023f57600080fd5b506102216105b6565b34801561025457600080fd5b506101f8600160a060020a03600435811690602435166044356105bb565b34801561027e57600080fd5b506102876105c4565b6040805160ff9092168252519081900360200190f35b3480156102a957600080fd5b50610148600160a060020a03600435166105cd565b3480156102ca57600080fd5b50610221600160a060020a03600435166024356105a8565b3480156102ee57600080fd5b5061015f6105d0565b34801561030357600080fd5b506040805160206004803580820135601f810184900484028501840190955284845261039e94369492936024939284019190819084018382808284375050604080516020601f818a01358b0180359182018390048302840183018552818452989b60ff8b35169b909a90999401975091955091820193509150819084018382808284375094975050843595505050505060200135151561062b565b60408051600160a060020a039092168252519081900360200190f35b3480156103c657600080fd5b50610221600160a060020a0360043516610636565b3480156103e757600080fd5b5061039e61063c565b3480156103fc57600080fd5b5061015f61064b565b34801561041157600080fd5b50610221600435610636565b34801561042957600080fd5b506101f86106a6565b34801561043e57600080fd5b506102216106af565b34801561045357600080fd5b50604080516020600460443581810135601f81018490048402850184019095528484526101f8948235600160a060020a03169460248035953695946064949201919081908401838280828437509497506105bb9650505050505050565b3480156104bc57600080fd5b50610221600160a060020a03600435811690602435166105a8565b3480156104e357600080fd5b5061039e6106b5565b3480156104f857600080fd5b5061014860043515156105cd565b34801561051257600080fd5b5061039e6106c9565b60018054604080516020600284861615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156105a05780601f10610575576101008083540402835291602001916105a0565b820191906000526020600020905b81548152906001019060200180831161058357829003601f168201915b505050505081565b600092915050565b60075481565b600090565b60009392505050565b60025460ff1681565b50565b6004805460408051602060026001851615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156105a05780601f10610575576101008083540402835291602001916105a0565b600095945050505050565b50600090565b600554600160a060020a031681565b6003805460408051602060026001851615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156105a05780601f10610575576101008083540402835291602001916105a0565b600b5460ff1681565b60065481565b600b546101009004600160a060020a031681565b600054600160a060020a0316815600a165627a7a723058207f52ec28ddd26f36564d70817f98768042227de89c606c567120b8e1562a11fb0029"

// DeployMiniMeToken deploys a new Ethereum contract, binding an instance of MiniMeToken to it.
func DeployMiniMeToken(auth *bind.TransactOpts, backend bind.ContractBackend, _tokenFactory common.Address, _parentToken common.Address, _parentSnapShotBlock *big.Int, _tokenName string, _decimalUnits uint8, _tokenSymbol string, _transfersEnabled bool) (common.Address, *types.Transaction, *MiniMeToken, error) {
	parsed, err := abi.JSON(strings.NewReader(MiniMeTokenABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(MiniMeTokenBin), backend, _tokenFactory, _parentToken, _parentSnapShotBlock, _tokenName, _decimalUnits, _tokenSymbol, _transfersEnabled)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &MiniMeToken{MiniMeTokenCaller: MiniMeTokenCaller{contract: contract}, MiniMeTokenTransactor: MiniMeTokenTransactor{contract: contract}, MiniMeTokenFilterer: MiniMeTokenFilterer{contract: contract}}, nil
}

// MiniMeToken is an auto generated Go binding around an Ethereum contract.
type MiniMeToken struct {
	MiniMeTokenCaller     // Read-only binding to the contract
	MiniMeTokenTransactor // Write-only binding to the contract
	MiniMeTokenFilterer   // Log filterer for contract events
}

// MiniMeTokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type MiniMeTokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MiniMeTokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MiniMeTokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MiniMeTokenFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MiniMeTokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MiniMeTokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MiniMeTokenSession struct {
	Contract     *MiniMeToken      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MiniMeTokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MiniMeTokenCallerSession struct {
	Contract *MiniMeTokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// MiniMeTokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MiniMeTokenTransactorSession struct {
	Contract     *MiniMeTokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// MiniMeTokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type MiniMeTokenRaw struct {
	Contract *MiniMeToken // Generic contract binding to access the raw methods on
}

// MiniMeTokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MiniMeTokenCallerRaw struct {
	Contract *MiniMeTokenCaller // Generic read-only contract binding to access the raw methods on
}

// MiniMeTokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MiniMeTokenTransactorRaw struct {
	Contract *MiniMeTokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMiniMeToken creates a new instance of MiniMeToken, bound to a specific deployed contract.
func NewMiniMeToken(address common.Address, backend bind.ContractBackend) (*MiniMeToken, error) {
	contract, err := bindMiniMeToken(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MiniMeToken{MiniMeTokenCaller: MiniMeTokenCaller{contract: contract}, MiniMeTokenTransactor: MiniMeTokenTransactor{contract: contract}, MiniMeTokenFilterer: MiniMeTokenFilterer{contract: contract}}, nil
}

// NewMiniMeTokenCaller creates a new read-only instance of MiniMeToken, bound to a specific deployed contract.
func NewMiniMeTokenCaller(address common.Address, caller bind.ContractCaller) (*MiniMeTokenCaller, error) {
	contract, err := bindMiniMeToken(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MiniMeTokenCaller{contract: contract}, nil
}

// NewMiniMeTokenTransactor creates a new write-only instance of MiniMeToken, bound to a specific deployed contract.
func NewMiniMeTokenTransactor(address common.Address, transactor bind.ContractTransactor) (*MiniMeTokenTransactor, error) {
	contract, err := bindMiniMeToken(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MiniMeTokenTransactor{contract: contract}, nil
}

// NewMiniMeTokenFilterer creates a new log filterer instance of MiniMeToken, bound to a specific deployed contract.
func NewMiniMeTokenFilterer(address common.Address, filterer bind.ContractFilterer) (*MiniMeTokenFilterer, error) {
	contract, err := bindMiniMeToken(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MiniMeTokenFilterer{contract: contract}, nil
}

// bindMiniMeToken binds a generic wrapper to an already deployed contract.
func bindMiniMeToken(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MiniMeTokenABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MiniMeToken *MiniMeTokenRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MiniMeToken.Contract.MiniMeTokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MiniMeToken *MiniMeTokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MiniMeToken.Contract.MiniMeTokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MiniMeToken *MiniMeTokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MiniMeToken.Contract.MiniMeTokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MiniMeToken *MiniMeTokenCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MiniMeToken.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MiniMeToken *MiniMeTokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MiniMeToken.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MiniMeToken *MiniMeTokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MiniMeToken.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address _spender) view returns(uint256 remaining)
func (_MiniMeToken *MiniMeTokenCaller) Allowance(opts *bind.CallOpts, _owner common.Address, _spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _MiniMeToken.contract.Call(opts, &out, "allowance", _owner, _spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address _spender) view returns(uint256 remaining)
func (_MiniMeToken *MiniMeTokenSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _MiniMeToken.Contract.Allowance(&_MiniMeToken.CallOpts, _owner, _spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address _spender) view returns(uint256 remaining)
func (_MiniMeToken *MiniMeTokenCallerSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _MiniMeToken.Contract.Allowance(&_MiniMeToken.CallOpts, _owner, _spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256 balance)
func (_MiniMeToken *MiniMeTokenCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _MiniMeToken.contract.Call(opts, &out, "balanceOf", _owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256 balance)
func (_MiniMeToken *MiniMeTokenSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _MiniMeToken.Contract.BalanceOf(&_MiniMeToken.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256 balance)
func (_MiniMeToken *MiniMeTokenCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _MiniMeToken.Contract.BalanceOf(&_MiniMeToken.CallOpts, _owner)
}

// BalanceOfAt is a free data retrieval call binding the contract method 0x4ee2cd7e.
//
// Solidity: function balanceOfAt(address _owner, uint256 _blockNumber) view returns(uint256)
func (_MiniMeToken *MiniMeTokenCaller) BalanceOfAt(opts *bind.CallOpts, _owner common.Address, _blockNumber *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _MiniMeToken.contract.Call(opts, &out, "balanceOfAt", _owner, _blockNumber)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOfAt is a free data retrieval call binding the contract method 0x4ee2cd7e.
//
// Solidity: function balanceOfAt(address _owner, uint256 _blockNumber) view returns(uint256)
func (_MiniMeToken *MiniMeTokenSession) BalanceOfAt(_owner common.Address, _blockNumber *big.Int) (*big.Int, error) {
	return _MiniMeToken.Contract.BalanceOfAt(&_MiniMeToken.CallOpts, _owner, _blockNumber)
}

// BalanceOfAt is a free data retrieval call binding the contract method 0x4ee2cd7e.
//
// Solidity: function balanceOfAt(address _owner, uint256 _blockNumber) view returns(uint256)
func (_MiniMeToken *MiniMeTokenCallerSession) BalanceOfAt(_owner common.Address, _blockNumber *big.Int) (*big.Int, error) {
	return _MiniMeToken.Contract.BalanceOfAt(&_MiniMeToken.CallOpts, _owner, _blockNumber)
}

// Controller is a free data retrieval call binding the contract method 0xf77c4791.
//
// Solidity: function controller() view returns(address)
func (_MiniMeToken *MiniMeTokenCaller) Controller(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MiniMeToken.contract.Call(opts, &out, "controller")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Controller is a free data retrieval call binding the contract method 0xf77c4791.
//
// Solidity: function controller() view returns(address)
func (_MiniMeToken *MiniMeTokenSession) Controller() (common.Address, error) {
	return _MiniMeToken.Contract.Controller(&_MiniMeToken.CallOpts)
}

// Controller is a free data retrieval call binding the contract method 0xf77c4791.
//
// Solidity: function controller() view returns(address)
func (_MiniMeToken *MiniMeTokenCallerSession) Controller() (common.Address, error) {
	return _MiniMeToken.Contract.Controller(&_MiniMeToken.CallOpts)
}

// CreationBlock is a free data retrieval call binding the contract method 0x17634514.
//
// Solidity: function creationBlock() view returns(uint256)
func (_MiniMeToken *MiniMeTokenCaller) CreationBlock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MiniMeToken.contract.Call(opts, &out, "creationBlock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CreationBlock is a free data retrieval call binding the contract method 0x17634514.
//
// Solidity: function creationBlock() view returns(uint256)
func (_MiniMeToken *MiniMeTokenSession) CreationBlock() (*big.Int, error) {
	return _MiniMeToken.Contract.CreationBlock(&_MiniMeToken.CallOpts)
}

// CreationBlock is a free data retrieval call binding the contract method 0x17634514.
//
// Solidity: function creationBlock() view returns(uint256)
func (_MiniMeToken *MiniMeTokenCallerSession) CreationBlock() (*big.Int, error) {
	return _MiniMeToken.Contract.CreationBlock(&_MiniMeToken.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_MiniMeToken *MiniMeTokenCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _MiniMeToken.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_MiniMeToken *MiniMeTokenSession) Decimals() (uint8, error) {
	return _MiniMeToken.Contract.Decimals(&_MiniMeToken.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_MiniMeToken *MiniMeTokenCallerSession) Decimals() (uint8, error) {
	return _MiniMeToken.Contract.Decimals(&_MiniMeToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_MiniMeToken *MiniMeTokenCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _MiniMeToken.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_MiniMeToken *MiniMeTokenSession) Name() (string, error) {
	return _MiniMeToken.Contract.Name(&_MiniMeToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_MiniMeToken *MiniMeTokenCallerSession) Name() (string, error) {
	return _MiniMeToken.Contract.Name(&_MiniMeToken.CallOpts)
}

// ParentSnapShotBlock is a free data retrieval call binding the contract method 0xc5bcc4f1.
//
// Solidity: function parentSnapShotBlock() view returns(uint256)
func (_MiniMeToken *MiniMeTokenCaller) ParentSnapShotBlock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MiniMeToken.contract.Call(opts, &out, "parentSnapShotBlock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ParentSnapShotBlock is a free data retrieval call binding the contract method 0xc5bcc4f1.
//
// Solidity: function parentSnapShotBlock() view returns(uint256)
func (_MiniMeToken *MiniMeTokenSession) ParentSnapShotBlock() (*big.Int, error) {
	return _MiniMeToken.Contract.ParentSnapShotBlock(&_MiniMeToken.CallOpts)
}

// ParentSnapShotBlock is a free data retrieval call binding the contract method 0xc5bcc4f1.
//
// Solidity: function parentSnapShotBlock() view returns(uint256)
func (_MiniMeToken *MiniMeTokenCallerSession) ParentSnapShotBlock() (*big.Int, error) {
	return _MiniMeToken.Contract.ParentSnapShotBlock(&_MiniMeToken.CallOpts)
}

// ParentToken is a free data retrieval call binding the contract method 0x80a54001.
//
// Solidity: function parentToken() view returns(address)
func (_MiniMeToken *MiniMeTokenCaller) ParentToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MiniMeToken.contract.Call(opts, &out, "parentToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ParentToken is a free data retrieval call binding the contract method 0x80a54001.
//
// Solidity: function parentToken() view returns(address)
func (_MiniMeToken *MiniMeTokenSession) ParentToken() (common.Address, error) {
	return _MiniMeToken.Contract.ParentToken(&_MiniMeToken.CallOpts)
}

// ParentToken is a free data retrieval call binding the contract method 0x80a54001.
//
// Solidity: function parentToken() view returns(address)
func (_MiniMeToken *MiniMeTokenCallerSession) ParentToken() (common.Address, error) {
	return _MiniMeToken.Contract.ParentToken(&_MiniMeToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_MiniMeToken *MiniMeTokenCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _MiniMeToken.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_MiniMeToken *MiniMeTokenSession) Symbol() (string, error) {
	return _MiniMeToken.Contract.Symbol(&_MiniMeToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_MiniMeToken *MiniMeTokenCallerSession) Symbol() (string, error) {
	return _MiniMeToken.Contract.Symbol(&_MiniMeToken.CallOpts)
}

// TokenFactory is a free data retrieval call binding the contract method 0xe77772fe.
//
// Solidity: function tokenFactory() view returns(address)
func (_MiniMeToken *MiniMeTokenCaller) TokenFactory(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MiniMeToken.contract.Call(opts, &out, "tokenFactory")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TokenFactory is a free data retrieval call binding the contract method 0xe77772fe.
//
// Solidity: function tokenFactory() view returns(address)
func (_MiniMeToken *MiniMeTokenSession) TokenFactory() (common.Address, error) {
	return _MiniMeToken.Contract.TokenFactory(&_MiniMeToken.CallOpts)
}

// TokenFactory is a free data retrieval call binding the contract method 0xe77772fe.
//
// Solidity: function tokenFactory() view returns(address)
func (_MiniMeToken *MiniMeTokenCallerSession) TokenFactory() (common.Address, error) {
	return _MiniMeToken.Contract.TokenFactory(&_MiniMeToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_MiniMeToken *MiniMeTokenCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MiniMeToken.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_MiniMeToken *MiniMeTokenSession) TotalSupply() (*big.Int, error) {
	return _MiniMeToken.Contract.TotalSupply(&_MiniMeToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_MiniMeToken *MiniMeTokenCallerSession) TotalSupply() (*big.Int, error) {
	return _MiniMeToken.Contract.TotalSupply(&_MiniMeToken.CallOpts)
}

// TotalSupplyAt is a free data retrieval call binding the contract method 0x981b24d0.
//
// Solidity: function totalSupplyAt(uint256 _blockNumber) view returns(uint256)
func (_MiniMeToken *MiniMeTokenCaller) TotalSupplyAt(opts *bind.CallOpts, _blockNumber *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _MiniMeToken.contract.Call(opts, &out, "totalSupplyAt", _blockNumber)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupplyAt is a free data retrieval call binding the contract method 0x981b24d0.
//
// Solidity: function totalSupplyAt(uint256 _blockNumber) view returns(uint256)
func (_MiniMeToken *MiniMeTokenSession) TotalSupplyAt(_blockNumber *big.Int) (*big.Int, error) {
	return _MiniMeToken.Contract.TotalSupplyAt(&_MiniMeToken.CallOpts, _blockNumber)
}

// TotalSupplyAt is a free data retrieval call binding the contract method 0x981b24d0.
//
// Solidity: function totalSupplyAt(uint256 _blockNumber) view returns(uint256)
func (_MiniMeToken *MiniMeTokenCallerSession) TotalSupplyAt(_blockNumber *big.Int) (*big.Int, error) {
	return _MiniMeToken.Contract.TotalSupplyAt(&_MiniMeToken.CallOpts, _blockNumber)
}

// TransfersEnabled is a free data retrieval call binding the contract method 0xbef97c87.
//
// Solidity: function transfersEnabled() view returns(bool)
func (_MiniMeToken *MiniMeTokenCaller) TransfersEnabled(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _MiniMeToken.contract.Call(opts, &out, "transfersEnabled")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// TransfersEnabled is a free data retrieval call binding the contract method 0xbef97c87.
//
// Solidity: function transfersEnabled() view returns(bool)
func (_MiniMeToken *MiniMeTokenSession) TransfersEnabled() (bool, error) {
	return _MiniMeToken.Contract.TransfersEnabled(&_MiniMeToken.CallOpts)
}

// TransfersEnabled is a free data retrieval call binding the contract method 0xbef97c87.
//
// Solidity: function transfersEnabled() view returns(bool)
func (_MiniMeToken *MiniMeTokenCallerSession) TransfersEnabled() (bool, error) {
	return _MiniMeToken.Contract.TransfersEnabled(&_MiniMeToken.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_MiniMeToken *MiniMeTokenCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _MiniMeToken.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_MiniMeToken *MiniMeTokenSession) Version() (string, error) {
	return _MiniMeToken.Contract.Version(&_MiniMeToken.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_MiniMeToken *MiniMeTokenCallerSession) Version() (string, error) {
	return _MiniMeToken.Contract.Version(&_MiniMeToken.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _amount) returns(bool success)
func (_MiniMeToken *MiniMeTokenTransactor) Approve(opts *bind.TransactOpts, _spender common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MiniMeToken.contract.Transact(opts, "approve", _spender, _amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _amount) returns(bool success)
func (_MiniMeToken *MiniMeTokenSession) Approve(_spender common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MiniMeToken.Contract.Approve(&_MiniMeToken.TransactOpts, _spender, _amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _amount) returns(bool success)
func (_MiniMeToken *MiniMeTokenTransactorSession) Approve(_spender common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MiniMeToken.Contract.Approve(&_MiniMeToken.TransactOpts, _spender, _amount)
}

// ApproveAndCall is a paid mutator transaction binding the contract method 0xcae9ca51.
//
// Solidity: function approveAndCall(address _spender, uint256 _amount, bytes _extraData) returns(bool success)
func (_MiniMeToken *MiniMeTokenTransactor) ApproveAndCall(opts *bind.TransactOpts, _spender common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _MiniMeToken.contract.Transact(opts, "approveAndCall", _spender, _amount, _extraData)
}

// ApproveAndCall is a paid mutator transaction binding the contract method 0xcae9ca51.
//
// Solidity: function approveAndCall(address _spender, uint256 _amount, bytes _extraData) returns(bool success)
func (_MiniMeToken *MiniMeTokenSession) ApproveAndCall(_spender common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _MiniMeToken.Contract.ApproveAndCall(&_MiniMeToken.TransactOpts, _spender, _amount, _extraData)
}

// ApproveAndCall is a paid mutator transaction binding the contract method 0xcae9ca51.
//
// Solidity: function approveAndCall(address _spender, uint256 _amount, bytes _extraData) returns(bool success)
func (_MiniMeToken *MiniMeTokenTransactorSession) ApproveAndCall(_spender common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _MiniMeToken.Contract.ApproveAndCall(&_MiniMeToken.TransactOpts, _spender, _amount, _extraData)
}

// ChangeController is a paid mutator transaction binding the contract method 0x3cebb823.
//
// Solidity: function changeController(address _newController) returns()
func (_MiniMeToken *MiniMeTokenTransactor) ChangeController(opts *bind.TransactOpts, _newController common.Address) (*types.Transaction, error) {
	return _MiniMeToken.contract.Transact(opts, "changeController", _newController)
}

// ChangeController is a paid mutator transaction binding the contract method 0x3cebb823.
//
// Solidity: function changeController(address _newController) returns()
func (_MiniMeToken *MiniMeTokenSession) ChangeController(_newController common.Address) (*types.Transaction, error) {
	return _MiniMeToken.Contract.ChangeController(&_MiniMeToken.TransactOpts, _newController)
}

// ChangeController is a paid mutator transaction binding the contract method 0x3cebb823.
//
// Solidity: function changeController(address _newController) returns()
func (_MiniMeToken *MiniMeTokenTransactorSession) ChangeController(_newController common.Address) (*types.Transaction, error) {
	return _MiniMeToken.Contract.ChangeController(&_MiniMeToken.TransactOpts, _newController)
}

// ClaimTokens is a paid mutator transaction binding the contract method 0xdf8de3e7.
//
// Solidity: function claimTokens(address _token) returns()
func (_MiniMeToken *MiniMeTokenTransactor) ClaimTokens(opts *bind.TransactOpts, _token common.Address) (*types.Transaction, error) {
	return _MiniMeToken.contract.Transact(opts, "claimTokens", _token)
}

// ClaimTokens is a paid mutator transaction binding the contract method 0xdf8de3e7.
//
// Solidity: function claimTokens(address _token) returns()
func (_MiniMeToken *MiniMeTokenSession) ClaimTokens(_token common.Address) (*types.Transaction, error) {
	return _MiniMeToken.Contract.ClaimTokens(&_MiniMeToken.TransactOpts, _token)
}

// ClaimTokens is a paid mutator transaction binding the contract method 0xdf8de3e7.
//
// Solidity: function claimTokens(address _token) returns()
func (_MiniMeToken *MiniMeTokenTransactorSession) ClaimTokens(_token common.Address) (*types.Transaction, error) {
	return _MiniMeToken.Contract.ClaimTokens(&_MiniMeToken.TransactOpts, _token)
}

// CreateCloneToken is a paid mutator transaction binding the contract method 0x6638c087.
//
// Solidity: function createCloneToken(string _cloneTokenName, uint8 _cloneDecimalUnits, string _cloneTokenSymbol, uint256 _snapshotBlock, bool _transfersEnabled) returns(address)
func (_MiniMeToken *MiniMeTokenTransactor) CreateCloneToken(opts *bind.TransactOpts, _cloneTokenName string, _cloneDecimalUnits uint8, _cloneTokenSymbol string, _snapshotBlock *big.Int, _transfersEnabled bool) (*types.Transaction, error) {
	return _MiniMeToken.contract.Transact(opts, "createCloneToken", _cloneTokenName, _cloneDecimalUnits, _cloneTokenSymbol, _snapshotBlock, _transfersEnabled)
}

// CreateCloneToken is a paid mutator transaction binding the contract method 0x6638c087.
//
// Solidity: function createCloneToken(string _cloneTokenName, uint8 _cloneDecimalUnits, string _cloneTokenSymbol, uint256 _snapshotBlock, bool _transfersEnabled) returns(address)
func (_MiniMeToken *MiniMeTokenSession) CreateCloneToken(_cloneTokenName string, _cloneDecimalUnits uint8, _cloneTokenSymbol string, _snapshotBlock *big.Int, _transfersEnabled bool) (*types.Transaction, error) {
	return _MiniMeToken.Contract.CreateCloneToken(&_MiniMeToken.TransactOpts, _cloneTokenName, _cloneDecimalUnits, _cloneTokenSymbol, _snapshotBlock, _transfersEnabled)
}

// CreateCloneToken is a paid mutator transaction binding the contract method 0x6638c087.
//
// Solidity: function createCloneToken(string _cloneTokenName, uint8 _cloneDecimalUnits, string _cloneTokenSymbol, uint256 _snapshotBlock, bool _transfersEnabled) returns(address)
func (_MiniMeToken *MiniMeTokenTransactorSession) CreateCloneToken(_cloneTokenName string, _cloneDecimalUnits uint8, _cloneTokenSymbol string, _snapshotBlock *big.Int, _transfersEnabled bool) (*types.Transaction, error) {
	return _MiniMeToken.Contract.CreateCloneToken(&_MiniMeToken.TransactOpts, _cloneTokenName, _cloneDecimalUnits, _cloneTokenSymbol, _snapshotBlock, _transfersEnabled)
}

// DestroyTokens is a paid mutator transaction binding the contract method 0xd3ce77fe.
//
// Solidity: function destroyTokens(address _owner, uint256 _amount) returns(bool)
func (_MiniMeToken *MiniMeTokenTransactor) DestroyTokens(opts *bind.TransactOpts, _owner common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MiniMeToken.contract.Transact(opts, "destroyTokens", _owner, _amount)
}

// DestroyTokens is a paid mutator transaction binding the contract method 0xd3ce77fe.
//
// Solidity: function destroyTokens(address _owner, uint256 _amount) returns(bool)
func (_MiniMeToken *MiniMeTokenSession) DestroyTokens(_owner common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MiniMeToken.Contract.DestroyTokens(&_MiniMeToken.TransactOpts, _owner, _amount)
}

// DestroyTokens is a paid mutator transaction binding the contract method 0xd3ce77fe.
//
// Solidity: function destroyTokens(address _owner, uint256 _amount) returns(bool)
func (_MiniMeToken *MiniMeTokenTransactorSession) DestroyTokens(_owner common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MiniMeToken.Contract.DestroyTokens(&_MiniMeToken.TransactOpts, _owner, _amount)
}

// EnableTransfers is a paid mutator transaction binding the contract method 0xf41e60c5.
//
// Solidity: function enableTransfers(bool _transfersEnabled) returns()
func (_MiniMeToken *MiniMeTokenTransactor) EnableTransfers(opts *bind.TransactOpts, _transfersEnabled bool) (*types.Transaction, error) {
	return _MiniMeToken.contract.Transact(opts, "enableTransfers", _transfersEnabled)
}

// EnableTransfers is a paid mutator transaction binding the contract method 0xf41e60c5.
//
// Solidity: function enableTransfers(bool _transfersEnabled) returns()
func (_MiniMeToken *MiniMeTokenSession) EnableTransfers(_transfersEnabled bool) (*types.Transaction, error) {
	return _MiniMeToken.Contract.EnableTransfers(&_MiniMeToken.TransactOpts, _transfersEnabled)
}

// EnableTransfers is a paid mutator transaction binding the contract method 0xf41e60c5.
//
// Solidity: function enableTransfers(bool _transfersEnabled) returns()
func (_MiniMeToken *MiniMeTokenTransactorSession) EnableTransfers(_transfersEnabled bool) (*types.Transaction, error) {
	return _MiniMeToken.Contract.EnableTransfers(&_MiniMeToken.TransactOpts, _transfersEnabled)
}

// GenerateTokens is a paid mutator transaction binding the contract method 0x827f32c0.
//
// Solidity: function generateTokens(address _owner, uint256 _amount) returns(bool)
func (_MiniMeToken *MiniMeTokenTransactor) GenerateTokens(opts *bind.TransactOpts, _owner common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MiniMeToken.contract.Transact(opts, "generateTokens", _owner, _amount)
}

// GenerateTokens is a paid mutator transaction binding the contract method 0x827f32c0.
//
// Solidity: function generateTokens(address _owner, uint256 _amount) returns(bool)
func (_MiniMeToken *MiniMeTokenSession) GenerateTokens(_owner common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MiniMeToken.Contract.GenerateTokens(&_MiniMeToken.TransactOpts, _owner, _amount)
}

// GenerateTokens is a paid mutator transaction binding the contract method 0x827f32c0.
//
// Solidity: function generateTokens(address _owner, uint256 _amount) returns(bool)
func (_MiniMeToken *MiniMeTokenTransactorSession) GenerateTokens(_owner common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MiniMeToken.Contract.GenerateTokens(&_MiniMeToken.TransactOpts, _owner, _amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _amount) returns(bool success)
func (_MiniMeToken *MiniMeTokenTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MiniMeToken.contract.Transact(opts, "transfer", _to, _amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _amount) returns(bool success)
func (_MiniMeToken *MiniMeTokenSession) Transfer(_to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MiniMeToken.Contract.Transfer(&_MiniMeToken.TransactOpts, _to, _amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _amount) returns(bool success)
func (_MiniMeToken *MiniMeTokenTransactorSession) Transfer(_to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MiniMeToken.Contract.Transfer(&_MiniMeToken.TransactOpts, _to, _amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _amount) returns(bool success)
func (_MiniMeToken *MiniMeTokenTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MiniMeToken.contract.Transact(opts, "transferFrom", _from, _to, _amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _amount) returns(bool success)
func (_MiniMeToken *MiniMeTokenSession) TransferFrom(_from common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MiniMeToken.Contract.TransferFrom(&_MiniMeToken.TransactOpts, _from, _to, _amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _amount) returns(bool success)
func (_MiniMeToken *MiniMeTokenTransactorSession) TransferFrom(_from common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MiniMeToken.Contract.TransferFrom(&_MiniMeToken.TransactOpts, _from, _to, _amount)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_MiniMeToken *MiniMeTokenTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _MiniMeToken.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_MiniMeToken *MiniMeTokenSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _MiniMeToken.Contract.Fallback(&_MiniMeToken.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_MiniMeToken *MiniMeTokenTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _MiniMeToken.Contract.Fallback(&_MiniMeToken.TransactOpts, calldata)
}

// MiniMeTokenApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the MiniMeToken contract.
type MiniMeTokenApprovalIterator struct {
	Event *MiniMeTokenApproval // Event containing the contract specifics and raw log

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
func (it *MiniMeTokenApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MiniMeTokenApproval)
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
		it.Event = new(MiniMeTokenApproval)
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
func (it *MiniMeTokenApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MiniMeTokenApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MiniMeTokenApproval represents a Approval event raised by the MiniMeToken contract.
type MiniMeTokenApproval struct {
	Owner   common.Address
	Spender common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed _owner, address indexed _spender, uint256 _amount)
func (_MiniMeToken *MiniMeTokenFilterer) FilterApproval(opts *bind.FilterOpts, _owner []common.Address, _spender []common.Address) (*MiniMeTokenApprovalIterator, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _spenderRule []interface{}
	for _, _spenderItem := range _spender {
		_spenderRule = append(_spenderRule, _spenderItem)
	}

	logs, sub, err := _MiniMeToken.contract.FilterLogs(opts, "Approval", _ownerRule, _spenderRule)
	if err != nil {
		return nil, err
	}
	return &MiniMeTokenApprovalIterator{contract: _MiniMeToken.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed _owner, address indexed _spender, uint256 _amount)
func (_MiniMeToken *MiniMeTokenFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *MiniMeTokenApproval, _owner []common.Address, _spender []common.Address) (event.Subscription, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _spenderRule []interface{}
	for _, _spenderItem := range _spender {
		_spenderRule = append(_spenderRule, _spenderItem)
	}

	logs, sub, err := _MiniMeToken.contract.WatchLogs(opts, "Approval", _ownerRule, _spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MiniMeTokenApproval)
				if err := _MiniMeToken.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed _owner, address indexed _spender, uint256 _amount)
func (_MiniMeToken *MiniMeTokenFilterer) ParseApproval(log types.Log) (*MiniMeTokenApproval, error) {
	event := new(MiniMeTokenApproval)
	if err := _MiniMeToken.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MiniMeTokenClaimedTokensIterator is returned from FilterClaimedTokens and is used to iterate over the raw logs and unpacked data for ClaimedTokens events raised by the MiniMeToken contract.
type MiniMeTokenClaimedTokensIterator struct {
	Event *MiniMeTokenClaimedTokens // Event containing the contract specifics and raw log

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
func (it *MiniMeTokenClaimedTokensIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MiniMeTokenClaimedTokens)
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
		it.Event = new(MiniMeTokenClaimedTokens)
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
func (it *MiniMeTokenClaimedTokensIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MiniMeTokenClaimedTokensIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MiniMeTokenClaimedTokens represents a ClaimedTokens event raised by the MiniMeToken contract.
type MiniMeTokenClaimedTokens struct {
	Token      common.Address
	Controller common.Address
	Amount     *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterClaimedTokens is a free log retrieval operation binding the contract event 0xf931edb47c50b4b4104c187b5814a9aef5f709e17e2ecf9617e860cacade929c.
//
// Solidity: event ClaimedTokens(address indexed _token, address indexed _controller, uint256 _amount)
func (_MiniMeToken *MiniMeTokenFilterer) FilterClaimedTokens(opts *bind.FilterOpts, _token []common.Address, _controller []common.Address) (*MiniMeTokenClaimedTokensIterator, error) {

	var _tokenRule []interface{}
	for _, _tokenItem := range _token {
		_tokenRule = append(_tokenRule, _tokenItem)
	}
	var _controllerRule []interface{}
	for _, _controllerItem := range _controller {
		_controllerRule = append(_controllerRule, _controllerItem)
	}

	logs, sub, err := _MiniMeToken.contract.FilterLogs(opts, "ClaimedTokens", _tokenRule, _controllerRule)
	if err != nil {
		return nil, err
	}
	return &MiniMeTokenClaimedTokensIterator{contract: _MiniMeToken.contract, event: "ClaimedTokens", logs: logs, sub: sub}, nil
}

// WatchClaimedTokens is a free log subscription operation binding the contract event 0xf931edb47c50b4b4104c187b5814a9aef5f709e17e2ecf9617e860cacade929c.
//
// Solidity: event ClaimedTokens(address indexed _token, address indexed _controller, uint256 _amount)
func (_MiniMeToken *MiniMeTokenFilterer) WatchClaimedTokens(opts *bind.WatchOpts, sink chan<- *MiniMeTokenClaimedTokens, _token []common.Address, _controller []common.Address) (event.Subscription, error) {

	var _tokenRule []interface{}
	for _, _tokenItem := range _token {
		_tokenRule = append(_tokenRule, _tokenItem)
	}
	var _controllerRule []interface{}
	for _, _controllerItem := range _controller {
		_controllerRule = append(_controllerRule, _controllerItem)
	}

	logs, sub, err := _MiniMeToken.contract.WatchLogs(opts, "ClaimedTokens", _tokenRule, _controllerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MiniMeTokenClaimedTokens)
				if err := _MiniMeToken.contract.UnpackLog(event, "ClaimedTokens", log); err != nil {
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

// ParseClaimedTokens is a log parse operation binding the contract event 0xf931edb47c50b4b4104c187b5814a9aef5f709e17e2ecf9617e860cacade929c.
//
// Solidity: event ClaimedTokens(address indexed _token, address indexed _controller, uint256 _amount)
func (_MiniMeToken *MiniMeTokenFilterer) ParseClaimedTokens(log types.Log) (*MiniMeTokenClaimedTokens, error) {
	event := new(MiniMeTokenClaimedTokens)
	if err := _MiniMeToken.contract.UnpackLog(event, "ClaimedTokens", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MiniMeTokenNewCloneTokenIterator is returned from FilterNewCloneToken and is used to iterate over the raw logs and unpacked data for NewCloneToken events raised by the MiniMeToken contract.
type MiniMeTokenNewCloneTokenIterator struct {
	Event *MiniMeTokenNewCloneToken // Event containing the contract specifics and raw log

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
func (it *MiniMeTokenNewCloneTokenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MiniMeTokenNewCloneToken)
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
		it.Event = new(MiniMeTokenNewCloneToken)
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
func (it *MiniMeTokenNewCloneTokenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MiniMeTokenNewCloneTokenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MiniMeTokenNewCloneToken represents a NewCloneToken event raised by the MiniMeToken contract.
type MiniMeTokenNewCloneToken struct {
	CloneToken    common.Address
	SnapshotBlock *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterNewCloneToken is a free log retrieval operation binding the contract event 0x086c875b377f900b07ce03575813022f05dd10ed7640b5282cf6d3c3fc352ade.
//
// Solidity: event NewCloneToken(address indexed _cloneToken, uint256 _snapshotBlock)
func (_MiniMeToken *MiniMeTokenFilterer) FilterNewCloneToken(opts *bind.FilterOpts, _cloneToken []common.Address) (*MiniMeTokenNewCloneTokenIterator, error) {

	var _cloneTokenRule []interface{}
	for _, _cloneTokenItem := range _cloneToken {
		_cloneTokenRule = append(_cloneTokenRule, _cloneTokenItem)
	}

	logs, sub, err := _MiniMeToken.contract.FilterLogs(opts, "NewCloneToken", _cloneTokenRule)
	if err != nil {
		return nil, err
	}
	return &MiniMeTokenNewCloneTokenIterator{contract: _MiniMeToken.contract, event: "NewCloneToken", logs: logs, sub: sub}, nil
}

// WatchNewCloneToken is a free log subscription operation binding the contract event 0x086c875b377f900b07ce03575813022f05dd10ed7640b5282cf6d3c3fc352ade.
//
// Solidity: event NewCloneToken(address indexed _cloneToken, uint256 _snapshotBlock)
func (_MiniMeToken *MiniMeTokenFilterer) WatchNewCloneToken(opts *bind.WatchOpts, sink chan<- *MiniMeTokenNewCloneToken, _cloneToken []common.Address) (event.Subscription, error) {

	var _cloneTokenRule []interface{}
	for _, _cloneTokenItem := range _cloneToken {
		_cloneTokenRule = append(_cloneTokenRule, _cloneTokenItem)
	}

	logs, sub, err := _MiniMeToken.contract.WatchLogs(opts, "NewCloneToken", _cloneTokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MiniMeTokenNewCloneToken)
				if err := _MiniMeToken.contract.UnpackLog(event, "NewCloneToken", log); err != nil {
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

// ParseNewCloneToken is a log parse operation binding the contract event 0x086c875b377f900b07ce03575813022f05dd10ed7640b5282cf6d3c3fc352ade.
//
// Solidity: event NewCloneToken(address indexed _cloneToken, uint256 _snapshotBlock)
func (_MiniMeToken *MiniMeTokenFilterer) ParseNewCloneToken(log types.Log) (*MiniMeTokenNewCloneToken, error) {
	event := new(MiniMeTokenNewCloneToken)
	if err := _MiniMeToken.contract.UnpackLog(event, "NewCloneToken", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MiniMeTokenTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the MiniMeToken contract.
type MiniMeTokenTransferIterator struct {
	Event *MiniMeTokenTransfer // Event containing the contract specifics and raw log

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
func (it *MiniMeTokenTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MiniMeTokenTransfer)
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
		it.Event = new(MiniMeTokenTransfer)
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
func (it *MiniMeTokenTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MiniMeTokenTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MiniMeTokenTransfer represents a Transfer event raised by the MiniMeToken contract.
type MiniMeTokenTransfer struct {
	From   common.Address
	To     common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 _amount)
func (_MiniMeToken *MiniMeTokenFilterer) FilterTransfer(opts *bind.FilterOpts, _from []common.Address, _to []common.Address) (*MiniMeTokenTransferIterator, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}

	logs, sub, err := _MiniMeToken.contract.FilterLogs(opts, "Transfer", _fromRule, _toRule)
	if err != nil {
		return nil, err
	}
	return &MiniMeTokenTransferIterator{contract: _MiniMeToken.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 _amount)
func (_MiniMeToken *MiniMeTokenFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *MiniMeTokenTransfer, _from []common.Address, _to []common.Address) (event.Subscription, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}

	logs, sub, err := _MiniMeToken.contract.WatchLogs(opts, "Transfer", _fromRule, _toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MiniMeTokenTransfer)
				if err := _MiniMeToken.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 _amount)
func (_MiniMeToken *MiniMeTokenFilterer) ParseTransfer(log types.Log) (*MiniMeTokenTransfer, error) {
	event := new(MiniMeTokenTransfer)
	if err := _MiniMeToken.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MiniMeTokenFactoryABI is the input ABI used to generate the binding from.
const MiniMeTokenFactoryABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_parentToken\",\"type\":\"address\"},{\"name\":\"_snapshotBlock\",\"type\":\"uint256\"},{\"name\":\"_tokenName\",\"type\":\"string\"},{\"name\":\"_decimalUnits\",\"type\":\"uint8\"},{\"name\":\"_tokenSymbol\",\"type\":\"string\"},{\"name\":\"_transfersEnabled\",\"type\":\"bool\"}],\"name\":\"createCloneToken\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// MiniMeTokenFactoryFuncSigs maps the 4-byte function signature to its string representation.
var MiniMeTokenFactoryFuncSigs = map[string]string{
	"5b7b72c1": "createCloneToken(address,uint256,string,uint8,string,bool)",
}

// MiniMeTokenFactory is an auto generated Go binding around an Ethereum contract.
type MiniMeTokenFactory struct {
	MiniMeTokenFactoryCaller     // Read-only binding to the contract
	MiniMeTokenFactoryTransactor // Write-only binding to the contract
	MiniMeTokenFactoryFilterer   // Log filterer for contract events
}

// MiniMeTokenFactoryCaller is an auto generated read-only Go binding around an Ethereum contract.
type MiniMeTokenFactoryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MiniMeTokenFactoryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MiniMeTokenFactoryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MiniMeTokenFactoryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MiniMeTokenFactoryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MiniMeTokenFactorySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MiniMeTokenFactorySession struct {
	Contract     *MiniMeTokenFactory // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// MiniMeTokenFactoryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MiniMeTokenFactoryCallerSession struct {
	Contract *MiniMeTokenFactoryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// MiniMeTokenFactoryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MiniMeTokenFactoryTransactorSession struct {
	Contract     *MiniMeTokenFactoryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// MiniMeTokenFactoryRaw is an auto generated low-level Go binding around an Ethereum contract.
type MiniMeTokenFactoryRaw struct {
	Contract *MiniMeTokenFactory // Generic contract binding to access the raw methods on
}

// MiniMeTokenFactoryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MiniMeTokenFactoryCallerRaw struct {
	Contract *MiniMeTokenFactoryCaller // Generic read-only contract binding to access the raw methods on
}

// MiniMeTokenFactoryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MiniMeTokenFactoryTransactorRaw struct {
	Contract *MiniMeTokenFactoryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMiniMeTokenFactory creates a new instance of MiniMeTokenFactory, bound to a specific deployed contract.
func NewMiniMeTokenFactory(address common.Address, backend bind.ContractBackend) (*MiniMeTokenFactory, error) {
	contract, err := bindMiniMeTokenFactory(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MiniMeTokenFactory{MiniMeTokenFactoryCaller: MiniMeTokenFactoryCaller{contract: contract}, MiniMeTokenFactoryTransactor: MiniMeTokenFactoryTransactor{contract: contract}, MiniMeTokenFactoryFilterer: MiniMeTokenFactoryFilterer{contract: contract}}, nil
}

// NewMiniMeTokenFactoryCaller creates a new read-only instance of MiniMeTokenFactory, bound to a specific deployed contract.
func NewMiniMeTokenFactoryCaller(address common.Address, caller bind.ContractCaller) (*MiniMeTokenFactoryCaller, error) {
	contract, err := bindMiniMeTokenFactory(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MiniMeTokenFactoryCaller{contract: contract}, nil
}

// NewMiniMeTokenFactoryTransactor creates a new write-only instance of MiniMeTokenFactory, bound to a specific deployed contract.
func NewMiniMeTokenFactoryTransactor(address common.Address, transactor bind.ContractTransactor) (*MiniMeTokenFactoryTransactor, error) {
	contract, err := bindMiniMeTokenFactory(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MiniMeTokenFactoryTransactor{contract: contract}, nil
}

// NewMiniMeTokenFactoryFilterer creates a new log filterer instance of MiniMeTokenFactory, bound to a specific deployed contract.
func NewMiniMeTokenFactoryFilterer(address common.Address, filterer bind.ContractFilterer) (*MiniMeTokenFactoryFilterer, error) {
	contract, err := bindMiniMeTokenFactory(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MiniMeTokenFactoryFilterer{contract: contract}, nil
}

// bindMiniMeTokenFactory binds a generic wrapper to an already deployed contract.
func bindMiniMeTokenFactory(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MiniMeTokenFactoryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MiniMeTokenFactory *MiniMeTokenFactoryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MiniMeTokenFactory.Contract.MiniMeTokenFactoryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MiniMeTokenFactory *MiniMeTokenFactoryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MiniMeTokenFactory.Contract.MiniMeTokenFactoryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MiniMeTokenFactory *MiniMeTokenFactoryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MiniMeTokenFactory.Contract.MiniMeTokenFactoryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MiniMeTokenFactory *MiniMeTokenFactoryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MiniMeTokenFactory.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MiniMeTokenFactory *MiniMeTokenFactoryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MiniMeTokenFactory.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MiniMeTokenFactory *MiniMeTokenFactoryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MiniMeTokenFactory.Contract.contract.Transact(opts, method, params...)
}

// CreateCloneToken is a paid mutator transaction binding the contract method 0x5b7b72c1.
//
// Solidity: function createCloneToken(address _parentToken, uint256 _snapshotBlock, string _tokenName, uint8 _decimalUnits, string _tokenSymbol, bool _transfersEnabled) returns(address)
func (_MiniMeTokenFactory *MiniMeTokenFactoryTransactor) CreateCloneToken(opts *bind.TransactOpts, _parentToken common.Address, _snapshotBlock *big.Int, _tokenName string, _decimalUnits uint8, _tokenSymbol string, _transfersEnabled bool) (*types.Transaction, error) {
	return _MiniMeTokenFactory.contract.Transact(opts, "createCloneToken", _parentToken, _snapshotBlock, _tokenName, _decimalUnits, _tokenSymbol, _transfersEnabled)
}

// CreateCloneToken is a paid mutator transaction binding the contract method 0x5b7b72c1.
//
// Solidity: function createCloneToken(address _parentToken, uint256 _snapshotBlock, string _tokenName, uint8 _decimalUnits, string _tokenSymbol, bool _transfersEnabled) returns(address)
func (_MiniMeTokenFactory *MiniMeTokenFactorySession) CreateCloneToken(_parentToken common.Address, _snapshotBlock *big.Int, _tokenName string, _decimalUnits uint8, _tokenSymbol string, _transfersEnabled bool) (*types.Transaction, error) {
	return _MiniMeTokenFactory.Contract.CreateCloneToken(&_MiniMeTokenFactory.TransactOpts, _parentToken, _snapshotBlock, _tokenName, _decimalUnits, _tokenSymbol, _transfersEnabled)
}

// CreateCloneToken is a paid mutator transaction binding the contract method 0x5b7b72c1.
//
// Solidity: function createCloneToken(address _parentToken, uint256 _snapshotBlock, string _tokenName, uint8 _decimalUnits, string _tokenSymbol, bool _transfersEnabled) returns(address)
func (_MiniMeTokenFactory *MiniMeTokenFactoryTransactorSession) CreateCloneToken(_parentToken common.Address, _snapshotBlock *big.Int, _tokenName string, _decimalUnits uint8, _tokenSymbol string, _transfersEnabled bool) (*types.Transaction, error) {
	return _MiniMeTokenFactory.Contract.CreateCloneToken(&_MiniMeTokenFactory.TransactOpts, _parentToken, _snapshotBlock, _tokenName, _decimalUnits, _tokenSymbol, _transfersEnabled)
}

// SNTABI is the input ABI used to generate the binding from.
const SNTABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"creationBlock\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newController\",\"type\":\"address\"}],\"name\":\"changeController\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_blockNumber\",\"type\":\"uint256\"}],\"name\":\"balanceOfAt\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_cloneTokenName\",\"type\":\"string\"},{\"name\":\"_cloneDecimalUnits\",\"type\":\"uint8\"},{\"name\":\"_cloneTokenSymbol\",\"type\":\"string\"},{\"name\":\"_snapshotBlock\",\"type\":\"uint256\"},{\"name\":\"_transfersEnabled\",\"type\":\"bool\"}],\"name\":\"createCloneToken\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"parentToken\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"generateTokens\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_blockNumber\",\"type\":\"uint256\"}],\"name\":\"totalSupplyAt\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"transfersEnabled\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"parentSnapShotBlock\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"},{\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"approveAndCall\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"destroyTokens\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"remaining\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"claimTokens\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"tokenFactory\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_transfersEnabled\",\"type\":\"bool\"}],\"name\":\"enableTransfers\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"controller\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_controller\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"ClaimedTokens\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_cloneToken\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_snapshotBlock\",\"type\":\"uint256\"}],\"name\":\"NewCloneToken\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"

// SNTFuncSigs maps the 4-byte function signature to its string representation.
var SNTFuncSigs = map[string]string{
	"dd62ed3e": "allowance(address,address)",
	"095ea7b3": "approve(address,uint256)",
	"cae9ca51": "approveAndCall(address,uint256,bytes)",
	"70a08231": "balanceOf(address)",
	"4ee2cd7e": "balanceOfAt(address,uint256)",
	"3cebb823": "changeController(address)",
	"df8de3e7": "claimTokens(address)",
	"f77c4791": "controller()",
	"6638c087": "createCloneToken(string,uint8,string,uint256,bool)",
	"17634514": "creationBlock()",
	"313ce567": "decimals()",
	"d3ce77fe": "destroyTokens(address,uint256)",
	"f41e60c5": "enableTransfers(bool)",
	"827f32c0": "generateTokens(address,uint256)",
	"06fdde03": "name()",
	"c5bcc4f1": "parentSnapShotBlock()",
	"80a54001": "parentToken()",
	"95d89b41": "symbol()",
	"e77772fe": "tokenFactory()",
	"18160ddd": "totalSupply()",
	"981b24d0": "totalSupplyAt(uint256)",
	"a9059cbb": "transfer(address,uint256)",
	"23b872dd": "transferFrom(address,address,uint256)",
	"bef97c87": "transfersEnabled()",
	"54fd4d50": "version()",
}

// SNT is an auto generated Go binding around an Ethereum contract.
type SNT struct {
	SNTCaller     // Read-only binding to the contract
	SNTTransactor // Write-only binding to the contract
	SNTFilterer   // Log filterer for contract events
}

// SNTCaller is an auto generated read-only Go binding around an Ethereum contract.
type SNTCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SNTTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SNTTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SNTFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SNTFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SNTSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SNTSession struct {
	Contract     *SNT              // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SNTCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SNTCallerSession struct {
	Contract *SNTCaller    // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// SNTTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SNTTransactorSession struct {
	Contract     *SNTTransactor    // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SNTRaw is an auto generated low-level Go binding around an Ethereum contract.
type SNTRaw struct {
	Contract *SNT // Generic contract binding to access the raw methods on
}

// SNTCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SNTCallerRaw struct {
	Contract *SNTCaller // Generic read-only contract binding to access the raw methods on
}

// SNTTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SNTTransactorRaw struct {
	Contract *SNTTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSNT creates a new instance of SNT, bound to a specific deployed contract.
func NewSNT(address common.Address, backend bind.ContractBackend) (*SNT, error) {
	contract, err := bindSNT(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SNT{SNTCaller: SNTCaller{contract: contract}, SNTTransactor: SNTTransactor{contract: contract}, SNTFilterer: SNTFilterer{contract: contract}}, nil
}

// NewSNTCaller creates a new read-only instance of SNT, bound to a specific deployed contract.
func NewSNTCaller(address common.Address, caller bind.ContractCaller) (*SNTCaller, error) {
	contract, err := bindSNT(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SNTCaller{contract: contract}, nil
}

// NewSNTTransactor creates a new write-only instance of SNT, bound to a specific deployed contract.
func NewSNTTransactor(address common.Address, transactor bind.ContractTransactor) (*SNTTransactor, error) {
	contract, err := bindSNT(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SNTTransactor{contract: contract}, nil
}

// NewSNTFilterer creates a new log filterer instance of SNT, bound to a specific deployed contract.
func NewSNTFilterer(address common.Address, filterer bind.ContractFilterer) (*SNTFilterer, error) {
	contract, err := bindSNT(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SNTFilterer{contract: contract}, nil
}

// bindSNT binds a generic wrapper to an already deployed contract.
func bindSNT(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SNTABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SNT *SNTRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SNT.Contract.SNTCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SNT *SNTRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SNT.Contract.SNTTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SNT *SNTRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SNT.Contract.SNTTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SNT *SNTCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SNT.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SNT *SNTTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SNT.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SNT *SNTTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SNT.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address _spender) view returns(uint256 remaining)
func (_SNT *SNTCaller) Allowance(opts *bind.CallOpts, _owner common.Address, _spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SNT.contract.Call(opts, &out, "allowance", _owner, _spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address _spender) view returns(uint256 remaining)
func (_SNT *SNTSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _SNT.Contract.Allowance(&_SNT.CallOpts, _owner, _spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address _spender) view returns(uint256 remaining)
func (_SNT *SNTCallerSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _SNT.Contract.Allowance(&_SNT.CallOpts, _owner, _spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256 balance)
func (_SNT *SNTCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SNT.contract.Call(opts, &out, "balanceOf", _owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256 balance)
func (_SNT *SNTSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _SNT.Contract.BalanceOf(&_SNT.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256 balance)
func (_SNT *SNTCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _SNT.Contract.BalanceOf(&_SNT.CallOpts, _owner)
}

// BalanceOfAt is a free data retrieval call binding the contract method 0x4ee2cd7e.
//
// Solidity: function balanceOfAt(address _owner, uint256 _blockNumber) view returns(uint256)
func (_SNT *SNTCaller) BalanceOfAt(opts *bind.CallOpts, _owner common.Address, _blockNumber *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SNT.contract.Call(opts, &out, "balanceOfAt", _owner, _blockNumber)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOfAt is a free data retrieval call binding the contract method 0x4ee2cd7e.
//
// Solidity: function balanceOfAt(address _owner, uint256 _blockNumber) view returns(uint256)
func (_SNT *SNTSession) BalanceOfAt(_owner common.Address, _blockNumber *big.Int) (*big.Int, error) {
	return _SNT.Contract.BalanceOfAt(&_SNT.CallOpts, _owner, _blockNumber)
}

// BalanceOfAt is a free data retrieval call binding the contract method 0x4ee2cd7e.
//
// Solidity: function balanceOfAt(address _owner, uint256 _blockNumber) view returns(uint256)
func (_SNT *SNTCallerSession) BalanceOfAt(_owner common.Address, _blockNumber *big.Int) (*big.Int, error) {
	return _SNT.Contract.BalanceOfAt(&_SNT.CallOpts, _owner, _blockNumber)
}

// Controller is a free data retrieval call binding the contract method 0xf77c4791.
//
// Solidity: function controller() view returns(address)
func (_SNT *SNTCaller) Controller(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SNT.contract.Call(opts, &out, "controller")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Controller is a free data retrieval call binding the contract method 0xf77c4791.
//
// Solidity: function controller() view returns(address)
func (_SNT *SNTSession) Controller() (common.Address, error) {
	return _SNT.Contract.Controller(&_SNT.CallOpts)
}

// Controller is a free data retrieval call binding the contract method 0xf77c4791.
//
// Solidity: function controller() view returns(address)
func (_SNT *SNTCallerSession) Controller() (common.Address, error) {
	return _SNT.Contract.Controller(&_SNT.CallOpts)
}

// CreationBlock is a free data retrieval call binding the contract method 0x17634514.
//
// Solidity: function creationBlock() view returns(uint256)
func (_SNT *SNTCaller) CreationBlock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SNT.contract.Call(opts, &out, "creationBlock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CreationBlock is a free data retrieval call binding the contract method 0x17634514.
//
// Solidity: function creationBlock() view returns(uint256)
func (_SNT *SNTSession) CreationBlock() (*big.Int, error) {
	return _SNT.Contract.CreationBlock(&_SNT.CallOpts)
}

// CreationBlock is a free data retrieval call binding the contract method 0x17634514.
//
// Solidity: function creationBlock() view returns(uint256)
func (_SNT *SNTCallerSession) CreationBlock() (*big.Int, error) {
	return _SNT.Contract.CreationBlock(&_SNT.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_SNT *SNTCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _SNT.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_SNT *SNTSession) Decimals() (uint8, error) {
	return _SNT.Contract.Decimals(&_SNT.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_SNT *SNTCallerSession) Decimals() (uint8, error) {
	return _SNT.Contract.Decimals(&_SNT.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_SNT *SNTCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _SNT.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_SNT *SNTSession) Name() (string, error) {
	return _SNT.Contract.Name(&_SNT.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_SNT *SNTCallerSession) Name() (string, error) {
	return _SNT.Contract.Name(&_SNT.CallOpts)
}

// ParentSnapShotBlock is a free data retrieval call binding the contract method 0xc5bcc4f1.
//
// Solidity: function parentSnapShotBlock() view returns(uint256)
func (_SNT *SNTCaller) ParentSnapShotBlock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SNT.contract.Call(opts, &out, "parentSnapShotBlock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ParentSnapShotBlock is a free data retrieval call binding the contract method 0xc5bcc4f1.
//
// Solidity: function parentSnapShotBlock() view returns(uint256)
func (_SNT *SNTSession) ParentSnapShotBlock() (*big.Int, error) {
	return _SNT.Contract.ParentSnapShotBlock(&_SNT.CallOpts)
}

// ParentSnapShotBlock is a free data retrieval call binding the contract method 0xc5bcc4f1.
//
// Solidity: function parentSnapShotBlock() view returns(uint256)
func (_SNT *SNTCallerSession) ParentSnapShotBlock() (*big.Int, error) {
	return _SNT.Contract.ParentSnapShotBlock(&_SNT.CallOpts)
}

// ParentToken is a free data retrieval call binding the contract method 0x80a54001.
//
// Solidity: function parentToken() view returns(address)
func (_SNT *SNTCaller) ParentToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SNT.contract.Call(opts, &out, "parentToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ParentToken is a free data retrieval call binding the contract method 0x80a54001.
//
// Solidity: function parentToken() view returns(address)
func (_SNT *SNTSession) ParentToken() (common.Address, error) {
	return _SNT.Contract.ParentToken(&_SNT.CallOpts)
}

// ParentToken is a free data retrieval call binding the contract method 0x80a54001.
//
// Solidity: function parentToken() view returns(address)
func (_SNT *SNTCallerSession) ParentToken() (common.Address, error) {
	return _SNT.Contract.ParentToken(&_SNT.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_SNT *SNTCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _SNT.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_SNT *SNTSession) Symbol() (string, error) {
	return _SNT.Contract.Symbol(&_SNT.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_SNT *SNTCallerSession) Symbol() (string, error) {
	return _SNT.Contract.Symbol(&_SNT.CallOpts)
}

// TokenFactory is a free data retrieval call binding the contract method 0xe77772fe.
//
// Solidity: function tokenFactory() view returns(address)
func (_SNT *SNTCaller) TokenFactory(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SNT.contract.Call(opts, &out, "tokenFactory")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TokenFactory is a free data retrieval call binding the contract method 0xe77772fe.
//
// Solidity: function tokenFactory() view returns(address)
func (_SNT *SNTSession) TokenFactory() (common.Address, error) {
	return _SNT.Contract.TokenFactory(&_SNT.CallOpts)
}

// TokenFactory is a free data retrieval call binding the contract method 0xe77772fe.
//
// Solidity: function tokenFactory() view returns(address)
func (_SNT *SNTCallerSession) TokenFactory() (common.Address, error) {
	return _SNT.Contract.TokenFactory(&_SNT.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_SNT *SNTCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SNT.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_SNT *SNTSession) TotalSupply() (*big.Int, error) {
	return _SNT.Contract.TotalSupply(&_SNT.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_SNT *SNTCallerSession) TotalSupply() (*big.Int, error) {
	return _SNT.Contract.TotalSupply(&_SNT.CallOpts)
}

// TotalSupplyAt is a free data retrieval call binding the contract method 0x981b24d0.
//
// Solidity: function totalSupplyAt(uint256 _blockNumber) view returns(uint256)
func (_SNT *SNTCaller) TotalSupplyAt(opts *bind.CallOpts, _blockNumber *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SNT.contract.Call(opts, &out, "totalSupplyAt", _blockNumber)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupplyAt is a free data retrieval call binding the contract method 0x981b24d0.
//
// Solidity: function totalSupplyAt(uint256 _blockNumber) view returns(uint256)
func (_SNT *SNTSession) TotalSupplyAt(_blockNumber *big.Int) (*big.Int, error) {
	return _SNT.Contract.TotalSupplyAt(&_SNT.CallOpts, _blockNumber)
}

// TotalSupplyAt is a free data retrieval call binding the contract method 0x981b24d0.
//
// Solidity: function totalSupplyAt(uint256 _blockNumber) view returns(uint256)
func (_SNT *SNTCallerSession) TotalSupplyAt(_blockNumber *big.Int) (*big.Int, error) {
	return _SNT.Contract.TotalSupplyAt(&_SNT.CallOpts, _blockNumber)
}

// TransfersEnabled is a free data retrieval call binding the contract method 0xbef97c87.
//
// Solidity: function transfersEnabled() view returns(bool)
func (_SNT *SNTCaller) TransfersEnabled(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _SNT.contract.Call(opts, &out, "transfersEnabled")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// TransfersEnabled is a free data retrieval call binding the contract method 0xbef97c87.
//
// Solidity: function transfersEnabled() view returns(bool)
func (_SNT *SNTSession) TransfersEnabled() (bool, error) {
	return _SNT.Contract.TransfersEnabled(&_SNT.CallOpts)
}

// TransfersEnabled is a free data retrieval call binding the contract method 0xbef97c87.
//
// Solidity: function transfersEnabled() view returns(bool)
func (_SNT *SNTCallerSession) TransfersEnabled() (bool, error) {
	return _SNT.Contract.TransfersEnabled(&_SNT.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_SNT *SNTCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _SNT.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_SNT *SNTSession) Version() (string, error) {
	return _SNT.Contract.Version(&_SNT.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_SNT *SNTCallerSession) Version() (string, error) {
	return _SNT.Contract.Version(&_SNT.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _amount) returns(bool success)
func (_SNT *SNTTransactor) Approve(opts *bind.TransactOpts, _spender common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SNT.contract.Transact(opts, "approve", _spender, _amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _amount) returns(bool success)
func (_SNT *SNTSession) Approve(_spender common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SNT.Contract.Approve(&_SNT.TransactOpts, _spender, _amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _amount) returns(bool success)
func (_SNT *SNTTransactorSession) Approve(_spender common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SNT.Contract.Approve(&_SNT.TransactOpts, _spender, _amount)
}

// ApproveAndCall is a paid mutator transaction binding the contract method 0xcae9ca51.
//
// Solidity: function approveAndCall(address _spender, uint256 _amount, bytes _extraData) returns(bool success)
func (_SNT *SNTTransactor) ApproveAndCall(opts *bind.TransactOpts, _spender common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _SNT.contract.Transact(opts, "approveAndCall", _spender, _amount, _extraData)
}

// ApproveAndCall is a paid mutator transaction binding the contract method 0xcae9ca51.
//
// Solidity: function approveAndCall(address _spender, uint256 _amount, bytes _extraData) returns(bool success)
func (_SNT *SNTSession) ApproveAndCall(_spender common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _SNT.Contract.ApproveAndCall(&_SNT.TransactOpts, _spender, _amount, _extraData)
}

// ApproveAndCall is a paid mutator transaction binding the contract method 0xcae9ca51.
//
// Solidity: function approveAndCall(address _spender, uint256 _amount, bytes _extraData) returns(bool success)
func (_SNT *SNTTransactorSession) ApproveAndCall(_spender common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _SNT.Contract.ApproveAndCall(&_SNT.TransactOpts, _spender, _amount, _extraData)
}

// ChangeController is a paid mutator transaction binding the contract method 0x3cebb823.
//
// Solidity: function changeController(address _newController) returns()
func (_SNT *SNTTransactor) ChangeController(opts *bind.TransactOpts, _newController common.Address) (*types.Transaction, error) {
	return _SNT.contract.Transact(opts, "changeController", _newController)
}

// ChangeController is a paid mutator transaction binding the contract method 0x3cebb823.
//
// Solidity: function changeController(address _newController) returns()
func (_SNT *SNTSession) ChangeController(_newController common.Address) (*types.Transaction, error) {
	return _SNT.Contract.ChangeController(&_SNT.TransactOpts, _newController)
}

// ChangeController is a paid mutator transaction binding the contract method 0x3cebb823.
//
// Solidity: function changeController(address _newController) returns()
func (_SNT *SNTTransactorSession) ChangeController(_newController common.Address) (*types.Transaction, error) {
	return _SNT.Contract.ChangeController(&_SNT.TransactOpts, _newController)
}

// ClaimTokens is a paid mutator transaction binding the contract method 0xdf8de3e7.
//
// Solidity: function claimTokens(address _token) returns()
func (_SNT *SNTTransactor) ClaimTokens(opts *bind.TransactOpts, _token common.Address) (*types.Transaction, error) {
	return _SNT.contract.Transact(opts, "claimTokens", _token)
}

// ClaimTokens is a paid mutator transaction binding the contract method 0xdf8de3e7.
//
// Solidity: function claimTokens(address _token) returns()
func (_SNT *SNTSession) ClaimTokens(_token common.Address) (*types.Transaction, error) {
	return _SNT.Contract.ClaimTokens(&_SNT.TransactOpts, _token)
}

// ClaimTokens is a paid mutator transaction binding the contract method 0xdf8de3e7.
//
// Solidity: function claimTokens(address _token) returns()
func (_SNT *SNTTransactorSession) ClaimTokens(_token common.Address) (*types.Transaction, error) {
	return _SNT.Contract.ClaimTokens(&_SNT.TransactOpts, _token)
}

// CreateCloneToken is a paid mutator transaction binding the contract method 0x6638c087.
//
// Solidity: function createCloneToken(string _cloneTokenName, uint8 _cloneDecimalUnits, string _cloneTokenSymbol, uint256 _snapshotBlock, bool _transfersEnabled) returns(address)
func (_SNT *SNTTransactor) CreateCloneToken(opts *bind.TransactOpts, _cloneTokenName string, _cloneDecimalUnits uint8, _cloneTokenSymbol string, _snapshotBlock *big.Int, _transfersEnabled bool) (*types.Transaction, error) {
	return _SNT.contract.Transact(opts, "createCloneToken", _cloneTokenName, _cloneDecimalUnits, _cloneTokenSymbol, _snapshotBlock, _transfersEnabled)
}

// CreateCloneToken is a paid mutator transaction binding the contract method 0x6638c087.
//
// Solidity: function createCloneToken(string _cloneTokenName, uint8 _cloneDecimalUnits, string _cloneTokenSymbol, uint256 _snapshotBlock, bool _transfersEnabled) returns(address)
func (_SNT *SNTSession) CreateCloneToken(_cloneTokenName string, _cloneDecimalUnits uint8, _cloneTokenSymbol string, _snapshotBlock *big.Int, _transfersEnabled bool) (*types.Transaction, error) {
	return _SNT.Contract.CreateCloneToken(&_SNT.TransactOpts, _cloneTokenName, _cloneDecimalUnits, _cloneTokenSymbol, _snapshotBlock, _transfersEnabled)
}

// CreateCloneToken is a paid mutator transaction binding the contract method 0x6638c087.
//
// Solidity: function createCloneToken(string _cloneTokenName, uint8 _cloneDecimalUnits, string _cloneTokenSymbol, uint256 _snapshotBlock, bool _transfersEnabled) returns(address)
func (_SNT *SNTTransactorSession) CreateCloneToken(_cloneTokenName string, _cloneDecimalUnits uint8, _cloneTokenSymbol string, _snapshotBlock *big.Int, _transfersEnabled bool) (*types.Transaction, error) {
	return _SNT.Contract.CreateCloneToken(&_SNT.TransactOpts, _cloneTokenName, _cloneDecimalUnits, _cloneTokenSymbol, _snapshotBlock, _transfersEnabled)
}

// DestroyTokens is a paid mutator transaction binding the contract method 0xd3ce77fe.
//
// Solidity: function destroyTokens(address _owner, uint256 _amount) returns(bool)
func (_SNT *SNTTransactor) DestroyTokens(opts *bind.TransactOpts, _owner common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SNT.contract.Transact(opts, "destroyTokens", _owner, _amount)
}

// DestroyTokens is a paid mutator transaction binding the contract method 0xd3ce77fe.
//
// Solidity: function destroyTokens(address _owner, uint256 _amount) returns(bool)
func (_SNT *SNTSession) DestroyTokens(_owner common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SNT.Contract.DestroyTokens(&_SNT.TransactOpts, _owner, _amount)
}

// DestroyTokens is a paid mutator transaction binding the contract method 0xd3ce77fe.
//
// Solidity: function destroyTokens(address _owner, uint256 _amount) returns(bool)
func (_SNT *SNTTransactorSession) DestroyTokens(_owner common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SNT.Contract.DestroyTokens(&_SNT.TransactOpts, _owner, _amount)
}

// EnableTransfers is a paid mutator transaction binding the contract method 0xf41e60c5.
//
// Solidity: function enableTransfers(bool _transfersEnabled) returns()
func (_SNT *SNTTransactor) EnableTransfers(opts *bind.TransactOpts, _transfersEnabled bool) (*types.Transaction, error) {
	return _SNT.contract.Transact(opts, "enableTransfers", _transfersEnabled)
}

// EnableTransfers is a paid mutator transaction binding the contract method 0xf41e60c5.
//
// Solidity: function enableTransfers(bool _transfersEnabled) returns()
func (_SNT *SNTSession) EnableTransfers(_transfersEnabled bool) (*types.Transaction, error) {
	return _SNT.Contract.EnableTransfers(&_SNT.TransactOpts, _transfersEnabled)
}

// EnableTransfers is a paid mutator transaction binding the contract method 0xf41e60c5.
//
// Solidity: function enableTransfers(bool _transfersEnabled) returns()
func (_SNT *SNTTransactorSession) EnableTransfers(_transfersEnabled bool) (*types.Transaction, error) {
	return _SNT.Contract.EnableTransfers(&_SNT.TransactOpts, _transfersEnabled)
}

// GenerateTokens is a paid mutator transaction binding the contract method 0x827f32c0.
//
// Solidity: function generateTokens(address _owner, uint256 _amount) returns(bool)
func (_SNT *SNTTransactor) GenerateTokens(opts *bind.TransactOpts, _owner common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SNT.contract.Transact(opts, "generateTokens", _owner, _amount)
}

// GenerateTokens is a paid mutator transaction binding the contract method 0x827f32c0.
//
// Solidity: function generateTokens(address _owner, uint256 _amount) returns(bool)
func (_SNT *SNTSession) GenerateTokens(_owner common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SNT.Contract.GenerateTokens(&_SNT.TransactOpts, _owner, _amount)
}

// GenerateTokens is a paid mutator transaction binding the contract method 0x827f32c0.
//
// Solidity: function generateTokens(address _owner, uint256 _amount) returns(bool)
func (_SNT *SNTTransactorSession) GenerateTokens(_owner common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SNT.Contract.GenerateTokens(&_SNT.TransactOpts, _owner, _amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _amount) returns(bool success)
func (_SNT *SNTTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SNT.contract.Transact(opts, "transfer", _to, _amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _amount) returns(bool success)
func (_SNT *SNTSession) Transfer(_to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SNT.Contract.Transfer(&_SNT.TransactOpts, _to, _amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _amount) returns(bool success)
func (_SNT *SNTTransactorSession) Transfer(_to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SNT.Contract.Transfer(&_SNT.TransactOpts, _to, _amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _amount) returns(bool success)
func (_SNT *SNTTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SNT.contract.Transact(opts, "transferFrom", _from, _to, _amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _amount) returns(bool success)
func (_SNT *SNTSession) TransferFrom(_from common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SNT.Contract.TransferFrom(&_SNT.TransactOpts, _from, _to, _amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _amount) returns(bool success)
func (_SNT *SNTTransactorSession) TransferFrom(_from common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _SNT.Contract.TransferFrom(&_SNT.TransactOpts, _from, _to, _amount)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_SNT *SNTTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _SNT.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_SNT *SNTSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _SNT.Contract.Fallback(&_SNT.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_SNT *SNTTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _SNT.Contract.Fallback(&_SNT.TransactOpts, calldata)
}

// SNTApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the SNT contract.
type SNTApprovalIterator struct {
	Event *SNTApproval // Event containing the contract specifics and raw log

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
func (it *SNTApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SNTApproval)
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
		it.Event = new(SNTApproval)
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
func (it *SNTApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SNTApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SNTApproval represents a Approval event raised by the SNT contract.
type SNTApproval struct {
	Owner   common.Address
	Spender common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed _owner, address indexed _spender, uint256 _amount)
func (_SNT *SNTFilterer) FilterApproval(opts *bind.FilterOpts, _owner []common.Address, _spender []common.Address) (*SNTApprovalIterator, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _spenderRule []interface{}
	for _, _spenderItem := range _spender {
		_spenderRule = append(_spenderRule, _spenderItem)
	}

	logs, sub, err := _SNT.contract.FilterLogs(opts, "Approval", _ownerRule, _spenderRule)
	if err != nil {
		return nil, err
	}
	return &SNTApprovalIterator{contract: _SNT.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed _owner, address indexed _spender, uint256 _amount)
func (_SNT *SNTFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *SNTApproval, _owner []common.Address, _spender []common.Address) (event.Subscription, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _spenderRule []interface{}
	for _, _spenderItem := range _spender {
		_spenderRule = append(_spenderRule, _spenderItem)
	}

	logs, sub, err := _SNT.contract.WatchLogs(opts, "Approval", _ownerRule, _spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SNTApproval)
				if err := _SNT.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed _owner, address indexed _spender, uint256 _amount)
func (_SNT *SNTFilterer) ParseApproval(log types.Log) (*SNTApproval, error) {
	event := new(SNTApproval)
	if err := _SNT.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SNTClaimedTokensIterator is returned from FilterClaimedTokens and is used to iterate over the raw logs and unpacked data for ClaimedTokens events raised by the SNT contract.
type SNTClaimedTokensIterator struct {
	Event *SNTClaimedTokens // Event containing the contract specifics and raw log

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
func (it *SNTClaimedTokensIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SNTClaimedTokens)
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
		it.Event = new(SNTClaimedTokens)
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
func (it *SNTClaimedTokensIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SNTClaimedTokensIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SNTClaimedTokens represents a ClaimedTokens event raised by the SNT contract.
type SNTClaimedTokens struct {
	Token      common.Address
	Controller common.Address
	Amount     *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterClaimedTokens is a free log retrieval operation binding the contract event 0xf931edb47c50b4b4104c187b5814a9aef5f709e17e2ecf9617e860cacade929c.
//
// Solidity: event ClaimedTokens(address indexed _token, address indexed _controller, uint256 _amount)
func (_SNT *SNTFilterer) FilterClaimedTokens(opts *bind.FilterOpts, _token []common.Address, _controller []common.Address) (*SNTClaimedTokensIterator, error) {

	var _tokenRule []interface{}
	for _, _tokenItem := range _token {
		_tokenRule = append(_tokenRule, _tokenItem)
	}
	var _controllerRule []interface{}
	for _, _controllerItem := range _controller {
		_controllerRule = append(_controllerRule, _controllerItem)
	}

	logs, sub, err := _SNT.contract.FilterLogs(opts, "ClaimedTokens", _tokenRule, _controllerRule)
	if err != nil {
		return nil, err
	}
	return &SNTClaimedTokensIterator{contract: _SNT.contract, event: "ClaimedTokens", logs: logs, sub: sub}, nil
}

// WatchClaimedTokens is a free log subscription operation binding the contract event 0xf931edb47c50b4b4104c187b5814a9aef5f709e17e2ecf9617e860cacade929c.
//
// Solidity: event ClaimedTokens(address indexed _token, address indexed _controller, uint256 _amount)
func (_SNT *SNTFilterer) WatchClaimedTokens(opts *bind.WatchOpts, sink chan<- *SNTClaimedTokens, _token []common.Address, _controller []common.Address) (event.Subscription, error) {

	var _tokenRule []interface{}
	for _, _tokenItem := range _token {
		_tokenRule = append(_tokenRule, _tokenItem)
	}
	var _controllerRule []interface{}
	for _, _controllerItem := range _controller {
		_controllerRule = append(_controllerRule, _controllerItem)
	}

	logs, sub, err := _SNT.contract.WatchLogs(opts, "ClaimedTokens", _tokenRule, _controllerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SNTClaimedTokens)
				if err := _SNT.contract.UnpackLog(event, "ClaimedTokens", log); err != nil {
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

// ParseClaimedTokens is a log parse operation binding the contract event 0xf931edb47c50b4b4104c187b5814a9aef5f709e17e2ecf9617e860cacade929c.
//
// Solidity: event ClaimedTokens(address indexed _token, address indexed _controller, uint256 _amount)
func (_SNT *SNTFilterer) ParseClaimedTokens(log types.Log) (*SNTClaimedTokens, error) {
	event := new(SNTClaimedTokens)
	if err := _SNT.contract.UnpackLog(event, "ClaimedTokens", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SNTNewCloneTokenIterator is returned from FilterNewCloneToken and is used to iterate over the raw logs and unpacked data for NewCloneToken events raised by the SNT contract.
type SNTNewCloneTokenIterator struct {
	Event *SNTNewCloneToken // Event containing the contract specifics and raw log

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
func (it *SNTNewCloneTokenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SNTNewCloneToken)
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
		it.Event = new(SNTNewCloneToken)
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
func (it *SNTNewCloneTokenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SNTNewCloneTokenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SNTNewCloneToken represents a NewCloneToken event raised by the SNT contract.
type SNTNewCloneToken struct {
	CloneToken    common.Address
	SnapshotBlock *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterNewCloneToken is a free log retrieval operation binding the contract event 0x086c875b377f900b07ce03575813022f05dd10ed7640b5282cf6d3c3fc352ade.
//
// Solidity: event NewCloneToken(address indexed _cloneToken, uint256 _snapshotBlock)
func (_SNT *SNTFilterer) FilterNewCloneToken(opts *bind.FilterOpts, _cloneToken []common.Address) (*SNTNewCloneTokenIterator, error) {

	var _cloneTokenRule []interface{}
	for _, _cloneTokenItem := range _cloneToken {
		_cloneTokenRule = append(_cloneTokenRule, _cloneTokenItem)
	}

	logs, sub, err := _SNT.contract.FilterLogs(opts, "NewCloneToken", _cloneTokenRule)
	if err != nil {
		return nil, err
	}
	return &SNTNewCloneTokenIterator{contract: _SNT.contract, event: "NewCloneToken", logs: logs, sub: sub}, nil
}

// WatchNewCloneToken is a free log subscription operation binding the contract event 0x086c875b377f900b07ce03575813022f05dd10ed7640b5282cf6d3c3fc352ade.
//
// Solidity: event NewCloneToken(address indexed _cloneToken, uint256 _snapshotBlock)
func (_SNT *SNTFilterer) WatchNewCloneToken(opts *bind.WatchOpts, sink chan<- *SNTNewCloneToken, _cloneToken []common.Address) (event.Subscription, error) {

	var _cloneTokenRule []interface{}
	for _, _cloneTokenItem := range _cloneToken {
		_cloneTokenRule = append(_cloneTokenRule, _cloneTokenItem)
	}

	logs, sub, err := _SNT.contract.WatchLogs(opts, "NewCloneToken", _cloneTokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SNTNewCloneToken)
				if err := _SNT.contract.UnpackLog(event, "NewCloneToken", log); err != nil {
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

// ParseNewCloneToken is a log parse operation binding the contract event 0x086c875b377f900b07ce03575813022f05dd10ed7640b5282cf6d3c3fc352ade.
//
// Solidity: event NewCloneToken(address indexed _cloneToken, uint256 _snapshotBlock)
func (_SNT *SNTFilterer) ParseNewCloneToken(log types.Log) (*SNTNewCloneToken, error) {
	event := new(SNTNewCloneToken)
	if err := _SNT.contract.UnpackLog(event, "NewCloneToken", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SNTTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the SNT contract.
type SNTTransferIterator struct {
	Event *SNTTransfer // Event containing the contract specifics and raw log

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
func (it *SNTTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SNTTransfer)
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
		it.Event = new(SNTTransfer)
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
func (it *SNTTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SNTTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SNTTransfer represents a Transfer event raised by the SNT contract.
type SNTTransfer struct {
	From   common.Address
	To     common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 _amount)
func (_SNT *SNTFilterer) FilterTransfer(opts *bind.FilterOpts, _from []common.Address, _to []common.Address) (*SNTTransferIterator, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}

	logs, sub, err := _SNT.contract.FilterLogs(opts, "Transfer", _fromRule, _toRule)
	if err != nil {
		return nil, err
	}
	return &SNTTransferIterator{contract: _SNT.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 _amount)
func (_SNT *SNTFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *SNTTransfer, _from []common.Address, _to []common.Address) (event.Subscription, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}

	logs, sub, err := _SNT.contract.WatchLogs(opts, "Transfer", _fromRule, _toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SNTTransfer)
				if err := _SNT.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 _amount)
func (_SNT *SNTFilterer) ParseTransfer(log types.Log) (*SNTTransfer, error) {
	event := new(SNTTransfer)
	if err := _SNT.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TokenControllerABI is the input ABI used to generate the binding from.
const TokenControllerABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"onTransfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"onApprove\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"proxyPayment\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"}]"

// TokenControllerFuncSigs maps the 4-byte function signature to its string representation.
var TokenControllerFuncSigs = map[string]string{
	"da682aeb": "onApprove(address,address,uint256)",
	"4a393149": "onTransfer(address,address,uint256)",
	"f48c3054": "proxyPayment(address)",
}

// TokenController is an auto generated Go binding around an Ethereum contract.
type TokenController struct {
	TokenControllerCaller     // Read-only binding to the contract
	TokenControllerTransactor // Write-only binding to the contract
	TokenControllerFilterer   // Log filterer for contract events
}

// TokenControllerCaller is an auto generated read-only Go binding around an Ethereum contract.
type TokenControllerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TokenControllerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TokenControllerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TokenControllerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TokenControllerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TokenControllerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TokenControllerSession struct {
	Contract     *TokenController  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TokenControllerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TokenControllerCallerSession struct {
	Contract *TokenControllerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// TokenControllerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TokenControllerTransactorSession struct {
	Contract     *TokenControllerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// TokenControllerRaw is an auto generated low-level Go binding around an Ethereum contract.
type TokenControllerRaw struct {
	Contract *TokenController // Generic contract binding to access the raw methods on
}

// TokenControllerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TokenControllerCallerRaw struct {
	Contract *TokenControllerCaller // Generic read-only contract binding to access the raw methods on
}

// TokenControllerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TokenControllerTransactorRaw struct {
	Contract *TokenControllerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTokenController creates a new instance of TokenController, bound to a specific deployed contract.
func NewTokenController(address common.Address, backend bind.ContractBackend) (*TokenController, error) {
	contract, err := bindTokenController(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TokenController{TokenControllerCaller: TokenControllerCaller{contract: contract}, TokenControllerTransactor: TokenControllerTransactor{contract: contract}, TokenControllerFilterer: TokenControllerFilterer{contract: contract}}, nil
}

// NewTokenControllerCaller creates a new read-only instance of TokenController, bound to a specific deployed contract.
func NewTokenControllerCaller(address common.Address, caller bind.ContractCaller) (*TokenControllerCaller, error) {
	contract, err := bindTokenController(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TokenControllerCaller{contract: contract}, nil
}

// NewTokenControllerTransactor creates a new write-only instance of TokenController, bound to a specific deployed contract.
func NewTokenControllerTransactor(address common.Address, transactor bind.ContractTransactor) (*TokenControllerTransactor, error) {
	contract, err := bindTokenController(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TokenControllerTransactor{contract: contract}, nil
}

// NewTokenControllerFilterer creates a new log filterer instance of TokenController, bound to a specific deployed contract.
func NewTokenControllerFilterer(address common.Address, filterer bind.ContractFilterer) (*TokenControllerFilterer, error) {
	contract, err := bindTokenController(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TokenControllerFilterer{contract: contract}, nil
}

// bindTokenController binds a generic wrapper to an already deployed contract.
func bindTokenController(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(TokenControllerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TokenController *TokenControllerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TokenController.Contract.TokenControllerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TokenController *TokenControllerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TokenController.Contract.TokenControllerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TokenController *TokenControllerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TokenController.Contract.TokenControllerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TokenController *TokenControllerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TokenController.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TokenController *TokenControllerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TokenController.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TokenController *TokenControllerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TokenController.Contract.contract.Transact(opts, method, params...)
}

// OnApprove is a paid mutator transaction binding the contract method 0xda682aeb.
//
// Solidity: function onApprove(address _owner, address _spender, uint256 _amount) returns(bool)
func (_TokenController *TokenControllerTransactor) OnApprove(opts *bind.TransactOpts, _owner common.Address, _spender common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _TokenController.contract.Transact(opts, "onApprove", _owner, _spender, _amount)
}

// OnApprove is a paid mutator transaction binding the contract method 0xda682aeb.
//
// Solidity: function onApprove(address _owner, address _spender, uint256 _amount) returns(bool)
func (_TokenController *TokenControllerSession) OnApprove(_owner common.Address, _spender common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _TokenController.Contract.OnApprove(&_TokenController.TransactOpts, _owner, _spender, _amount)
}

// OnApprove is a paid mutator transaction binding the contract method 0xda682aeb.
//
// Solidity: function onApprove(address _owner, address _spender, uint256 _amount) returns(bool)
func (_TokenController *TokenControllerTransactorSession) OnApprove(_owner common.Address, _spender common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _TokenController.Contract.OnApprove(&_TokenController.TransactOpts, _owner, _spender, _amount)
}

// OnTransfer is a paid mutator transaction binding the contract method 0x4a393149.
//
// Solidity: function onTransfer(address _from, address _to, uint256 _amount) returns(bool)
func (_TokenController *TokenControllerTransactor) OnTransfer(opts *bind.TransactOpts, _from common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _TokenController.contract.Transact(opts, "onTransfer", _from, _to, _amount)
}

// OnTransfer is a paid mutator transaction binding the contract method 0x4a393149.
//
// Solidity: function onTransfer(address _from, address _to, uint256 _amount) returns(bool)
func (_TokenController *TokenControllerSession) OnTransfer(_from common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _TokenController.Contract.OnTransfer(&_TokenController.TransactOpts, _from, _to, _amount)
}

// OnTransfer is a paid mutator transaction binding the contract method 0x4a393149.
//
// Solidity: function onTransfer(address _from, address _to, uint256 _amount) returns(bool)
func (_TokenController *TokenControllerTransactorSession) OnTransfer(_from common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _TokenController.Contract.OnTransfer(&_TokenController.TransactOpts, _from, _to, _amount)
}

// ProxyPayment is a paid mutator transaction binding the contract method 0xf48c3054.
//
// Solidity: function proxyPayment(address _owner) payable returns(bool)
func (_TokenController *TokenControllerTransactor) ProxyPayment(opts *bind.TransactOpts, _owner common.Address) (*types.Transaction, error) {
	return _TokenController.contract.Transact(opts, "proxyPayment", _owner)
}

// ProxyPayment is a paid mutator transaction binding the contract method 0xf48c3054.
//
// Solidity: function proxyPayment(address _owner) payable returns(bool)
func (_TokenController *TokenControllerSession) ProxyPayment(_owner common.Address) (*types.Transaction, error) {
	return _TokenController.Contract.ProxyPayment(&_TokenController.TransactOpts, _owner)
}

// ProxyPayment is a paid mutator transaction binding the contract method 0xf48c3054.
//
// Solidity: function proxyPayment(address _owner) payable returns(bool)
func (_TokenController *TokenControllerTransactorSession) ProxyPayment(_owner common.Address) (*types.Transaction, error) {
	return _TokenController.Contract.ProxyPayment(&_TokenController.TransactOpts, _owner)
}
