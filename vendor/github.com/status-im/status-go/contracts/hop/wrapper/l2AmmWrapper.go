// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package hopWrapper

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

// HopWrapperABI is the input ABI used to generate the binding from.
const HopWrapperABI = "[{\"inputs\":[{\"internalType\":\"contractL2_Bridge\",\"name\":\"_bridge\",\"type\":\"address\"},{\"internalType\":\"contractIERC20\",\"name\":\"_l2CanonicalToken\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"_l2CanonicalTokenIsEth\",\"type\":\"bool\"},{\"internalType\":\"contractIERC20\",\"name\":\"_hToken\",\"type\":\"address\"},{\"internalType\":\"contractSwap\",\"name\":\"_exchangeAddress\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amountOutMin\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"name\":\"attemptSwap\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"bridge\",\"outputs\":[{\"internalType\":\"contractL2_Bridge\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"exchangeAddress\",\"outputs\":[{\"internalType\":\"contractSwap\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"hToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"l2CanonicalToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"l2CanonicalTokenIsEth\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"bonderFee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amountOutMin\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"destinationAmountOutMin\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"destinationDeadline\",\"type\":\"uint256\"}],\"name\":\"swapAndSend\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]"

// HopWrapper is an auto generated Go binding around an Ethereum contract.
type HopWrapper struct {
	HopWrapperCaller     // Read-only binding to the contract
	HopWrapperTransactor // Write-only binding to the contract
	HopWrapperFilterer   // Log filterer for contract events
}

// HopWrapperCaller is an auto generated read-only Go binding around an Ethereum contract.
type HopWrapperCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HopWrapperTransactor is an auto generated write-only Go binding around an Ethereum contract.
type HopWrapperTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HopWrapperFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type HopWrapperFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HopWrapperSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type HopWrapperSession struct {
	Contract     *HopWrapper       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HopWrapperCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type HopWrapperCallerSession struct {
	Contract *HopWrapperCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// HopWrapperTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type HopWrapperTransactorSession struct {
	Contract     *HopWrapperTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// HopWrapperRaw is an auto generated low-level Go binding around an Ethereum contract.
type HopWrapperRaw struct {
	Contract *HopWrapper // Generic contract binding to access the raw methods on
}

// HopWrapperCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type HopWrapperCallerRaw struct {
	Contract *HopWrapperCaller // Generic read-only contract binding to access the raw methods on
}

// HopWrapperTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type HopWrapperTransactorRaw struct {
	Contract *HopWrapperTransactor // Generic write-only contract binding to access the raw methods on
}

// NewHopWrapper creates a new instance of HopWrapper, bound to a specific deployed contract.
func NewHopWrapper(address common.Address, backend bind.ContractBackend) (*HopWrapper, error) {
	contract, err := bindHopWrapper(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &HopWrapper{HopWrapperCaller: HopWrapperCaller{contract: contract}, HopWrapperTransactor: HopWrapperTransactor{contract: contract}, HopWrapperFilterer: HopWrapperFilterer{contract: contract}}, nil
}

// NewHopWrapperCaller creates a new read-only instance of HopWrapper, bound to a specific deployed contract.
func NewHopWrapperCaller(address common.Address, caller bind.ContractCaller) (*HopWrapperCaller, error) {
	contract, err := bindHopWrapper(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &HopWrapperCaller{contract: contract}, nil
}

// NewHopWrapperTransactor creates a new write-only instance of HopWrapper, bound to a specific deployed contract.
func NewHopWrapperTransactor(address common.Address, transactor bind.ContractTransactor) (*HopWrapperTransactor, error) {
	contract, err := bindHopWrapper(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &HopWrapperTransactor{contract: contract}, nil
}

// NewHopWrapperFilterer creates a new log filterer instance of HopWrapper, bound to a specific deployed contract.
func NewHopWrapperFilterer(address common.Address, filterer bind.ContractFilterer) (*HopWrapperFilterer, error) {
	contract, err := bindHopWrapper(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &HopWrapperFilterer{contract: contract}, nil
}

// bindHopWrapper binds a generic wrapper to an already deployed contract.
func bindHopWrapper(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(HopWrapperABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HopWrapper *HopWrapperRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _HopWrapper.Contract.HopWrapperCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HopWrapper *HopWrapperRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HopWrapper.Contract.HopWrapperTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HopWrapper *HopWrapperRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HopWrapper.Contract.HopWrapperTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HopWrapper *HopWrapperCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _HopWrapper.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HopWrapper *HopWrapperTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HopWrapper.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HopWrapper *HopWrapperTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HopWrapper.Contract.contract.Transact(opts, method, params...)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_HopWrapper *HopWrapperCaller) Bridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _HopWrapper.contract.Call(opts, &out, "bridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_HopWrapper *HopWrapperSession) Bridge() (common.Address, error) {
	return _HopWrapper.Contract.Bridge(&_HopWrapper.CallOpts)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_HopWrapper *HopWrapperCallerSession) Bridge() (common.Address, error) {
	return _HopWrapper.Contract.Bridge(&_HopWrapper.CallOpts)
}

// ExchangeAddress is a free data retrieval call binding the contract method 0x9cd01605.
//
// Solidity: function exchangeAddress() view returns(address)
func (_HopWrapper *HopWrapperCaller) ExchangeAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _HopWrapper.contract.Call(opts, &out, "exchangeAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ExchangeAddress is a free data retrieval call binding the contract method 0x9cd01605.
//
// Solidity: function exchangeAddress() view returns(address)
func (_HopWrapper *HopWrapperSession) ExchangeAddress() (common.Address, error) {
	return _HopWrapper.Contract.ExchangeAddress(&_HopWrapper.CallOpts)
}

// ExchangeAddress is a free data retrieval call binding the contract method 0x9cd01605.
//
// Solidity: function exchangeAddress() view returns(address)
func (_HopWrapper *HopWrapperCallerSession) ExchangeAddress() (common.Address, error) {
	return _HopWrapper.Contract.ExchangeAddress(&_HopWrapper.CallOpts)
}

// HToken is a free data retrieval call binding the contract method 0xfc6e3b3b.
//
// Solidity: function hToken() view returns(address)
func (_HopWrapper *HopWrapperCaller) HToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _HopWrapper.contract.Call(opts, &out, "hToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// HToken is a free data retrieval call binding the contract method 0xfc6e3b3b.
//
// Solidity: function hToken() view returns(address)
func (_HopWrapper *HopWrapperSession) HToken() (common.Address, error) {
	return _HopWrapper.Contract.HToken(&_HopWrapper.CallOpts)
}

// HToken is a free data retrieval call binding the contract method 0xfc6e3b3b.
//
// Solidity: function hToken() view returns(address)
func (_HopWrapper *HopWrapperCallerSession) HToken() (common.Address, error) {
	return _HopWrapper.Contract.HToken(&_HopWrapper.CallOpts)
}

// L2CanonicalToken is a free data retrieval call binding the contract method 0x1ee1bf67.
//
// Solidity: function l2CanonicalToken() view returns(address)
func (_HopWrapper *HopWrapperCaller) L2CanonicalToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _HopWrapper.contract.Call(opts, &out, "l2CanonicalToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L2CanonicalToken is a free data retrieval call binding the contract method 0x1ee1bf67.
//
// Solidity: function l2CanonicalToken() view returns(address)
func (_HopWrapper *HopWrapperSession) L2CanonicalToken() (common.Address, error) {
	return _HopWrapper.Contract.L2CanonicalToken(&_HopWrapper.CallOpts)
}

// L2CanonicalToken is a free data retrieval call binding the contract method 0x1ee1bf67.
//
// Solidity: function l2CanonicalToken() view returns(address)
func (_HopWrapper *HopWrapperCallerSession) L2CanonicalToken() (common.Address, error) {
	return _HopWrapper.Contract.L2CanonicalToken(&_HopWrapper.CallOpts)
}

// L2CanonicalTokenIsEth is a free data retrieval call binding the contract method 0x28555125.
//
// Solidity: function l2CanonicalTokenIsEth() view returns(bool)
func (_HopWrapper *HopWrapperCaller) L2CanonicalTokenIsEth(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _HopWrapper.contract.Call(opts, &out, "l2CanonicalTokenIsEth")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// L2CanonicalTokenIsEth is a free data retrieval call binding the contract method 0x28555125.
//
// Solidity: function l2CanonicalTokenIsEth() view returns(bool)
func (_HopWrapper *HopWrapperSession) L2CanonicalTokenIsEth() (bool, error) {
	return _HopWrapper.Contract.L2CanonicalTokenIsEth(&_HopWrapper.CallOpts)
}

// L2CanonicalTokenIsEth is a free data retrieval call binding the contract method 0x28555125.
//
// Solidity: function l2CanonicalTokenIsEth() view returns(bool)
func (_HopWrapper *HopWrapperCallerSession) L2CanonicalTokenIsEth() (bool, error) {
	return _HopWrapper.Contract.L2CanonicalTokenIsEth(&_HopWrapper.CallOpts)
}

// AttemptSwap is a paid mutator transaction binding the contract method 0x676c5ef6.
//
// Solidity: function attemptSwap(address recipient, uint256 amount, uint256 amountOutMin, uint256 deadline) returns()
func (_HopWrapper *HopWrapperTransactor) AttemptSwap(opts *bind.TransactOpts, recipient common.Address, amount *big.Int, amountOutMin *big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopWrapper.contract.Transact(opts, "attemptSwap", recipient, amount, amountOutMin, deadline)
}

// AttemptSwap is a paid mutator transaction binding the contract method 0x676c5ef6.
//
// Solidity: function attemptSwap(address recipient, uint256 amount, uint256 amountOutMin, uint256 deadline) returns()
func (_HopWrapper *HopWrapperSession) AttemptSwap(recipient common.Address, amount *big.Int, amountOutMin *big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopWrapper.Contract.AttemptSwap(&_HopWrapper.TransactOpts, recipient, amount, amountOutMin, deadline)
}

// AttemptSwap is a paid mutator transaction binding the contract method 0x676c5ef6.
//
// Solidity: function attemptSwap(address recipient, uint256 amount, uint256 amountOutMin, uint256 deadline) returns()
func (_HopWrapper *HopWrapperTransactorSession) AttemptSwap(recipient common.Address, amount *big.Int, amountOutMin *big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopWrapper.Contract.AttemptSwap(&_HopWrapper.TransactOpts, recipient, amount, amountOutMin, deadline)
}

// SwapAndSend is a paid mutator transaction binding the contract method 0xeea0d7b2.
//
// Solidity: function swapAndSend(uint256 chainId, address recipient, uint256 amount, uint256 bonderFee, uint256 amountOutMin, uint256 deadline, uint256 destinationAmountOutMin, uint256 destinationDeadline) payable returns()
func (_HopWrapper *HopWrapperTransactor) SwapAndSend(opts *bind.TransactOpts, chainId *big.Int, recipient common.Address, amount *big.Int, bonderFee *big.Int, amountOutMin *big.Int, deadline *big.Int, destinationAmountOutMin *big.Int, destinationDeadline *big.Int) (*types.Transaction, error) {
	return _HopWrapper.contract.Transact(opts, "swapAndSend", chainId, recipient, amount, bonderFee, amountOutMin, deadline, destinationAmountOutMin, destinationDeadline)
}

// SwapAndSend is a paid mutator transaction binding the contract method 0xeea0d7b2.
//
// Solidity: function swapAndSend(uint256 chainId, address recipient, uint256 amount, uint256 bonderFee, uint256 amountOutMin, uint256 deadline, uint256 destinationAmountOutMin, uint256 destinationDeadline) payable returns()
func (_HopWrapper *HopWrapperSession) SwapAndSend(chainId *big.Int, recipient common.Address, amount *big.Int, bonderFee *big.Int, amountOutMin *big.Int, deadline *big.Int, destinationAmountOutMin *big.Int, destinationDeadline *big.Int) (*types.Transaction, error) {
	return _HopWrapper.Contract.SwapAndSend(&_HopWrapper.TransactOpts, chainId, recipient, amount, bonderFee, amountOutMin, deadline, destinationAmountOutMin, destinationDeadline)
}

// SwapAndSend is a paid mutator transaction binding the contract method 0xeea0d7b2.
//
// Solidity: function swapAndSend(uint256 chainId, address recipient, uint256 amount, uint256 bonderFee, uint256 amountOutMin, uint256 deadline, uint256 destinationAmountOutMin, uint256 destinationDeadline) payable returns()
func (_HopWrapper *HopWrapperTransactorSession) SwapAndSend(chainId *big.Int, recipient common.Address, amount *big.Int, bonderFee *big.Int, amountOutMin *big.Int, deadline *big.Int, destinationAmountOutMin *big.Int, destinationDeadline *big.Int) (*types.Transaction, error) {
	return _HopWrapper.Contract.SwapAndSend(&_HopWrapper.TransactOpts, chainId, recipient, amount, bonderFee, amountOutMin, deadline, destinationAmountOutMin, destinationDeadline)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_HopWrapper *HopWrapperTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HopWrapper.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_HopWrapper *HopWrapperSession) Receive() (*types.Transaction, error) {
	return _HopWrapper.Contract.Receive(&_HopWrapper.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_HopWrapper *HopWrapperTransactorSession) Receive() (*types.Transaction, error) {
	return _HopWrapper.Contract.Receive(&_HopWrapper.TransactOpts)
}
