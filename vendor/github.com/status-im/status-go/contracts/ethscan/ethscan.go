// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ethscan

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

// BalanceScannerResult is an auto generated low-level Go binding around an user-defined struct.
type BalanceScannerResult struct {
	Success bool
	Data    []byte
}

// BalanceScannerABI is the input ABI used to generate the binding from.
const BalanceScannerABI = "[{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"contracts\",\"type\":\"address[]\"},{\"internalType\":\"bytes[]\",\"name\":\"data\",\"type\":\"bytes[]\"},{\"internalType\":\"uint256\",\"name\":\"gas\",\"type\":\"uint256\"}],\"name\":\"call\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"internalType\":\"structBalanceScanner.Result[]\",\"name\":\"results\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"contracts\",\"type\":\"address[]\"},{\"internalType\":\"bytes[]\",\"name\":\"data\",\"type\":\"bytes[]\"}],\"name\":\"call\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"internalType\":\"structBalanceScanner.Result[]\",\"name\":\"results\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"addresses\",\"type\":\"address[]\"}],\"name\":\"etherBalances\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"internalType\":\"structBalanceScanner.Result[]\",\"name\":\"results\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"addresses\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"tokenBalances\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"internalType\":\"structBalanceScanner.Result[]\",\"name\":\"results\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"contracts\",\"type\":\"address[]\"}],\"name\":\"tokensBalance\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"internalType\":\"structBalanceScanner.Result[]\",\"name\":\"results\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// BalanceScannerFuncSigs maps the 4-byte function signature to its string representation.
var BalanceScannerFuncSigs = map[string]string{
	"458b3a7c": "call(address[],bytes[])",
	"36738374": "call(address[],bytes[],uint256)",
	"dbdbb51b": "etherBalances(address[])",
	"aad33091": "tokenBalances(address[],address)",
	"e5da1b68": "tokensBalance(address,address[])",
}

// BalanceScanner is an auto generated Go binding around an Ethereum contract.
type BalanceScanner struct {
	BalanceScannerCaller     // Read-only binding to the contract
	BalanceScannerTransactor // Write-only binding to the contract
	BalanceScannerFilterer   // Log filterer for contract events
}

// BalanceScannerCaller is an auto generated read-only Go binding around an Ethereum contract.
type BalanceScannerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BalanceScannerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BalanceScannerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BalanceScannerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BalanceScannerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BalanceScannerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BalanceScannerSession struct {
	Contract     *BalanceScanner   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BalanceScannerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BalanceScannerCallerSession struct {
	Contract *BalanceScannerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// BalanceScannerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BalanceScannerTransactorSession struct {
	Contract     *BalanceScannerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// BalanceScannerRaw is an auto generated low-level Go binding around an Ethereum contract.
type BalanceScannerRaw struct {
	Contract *BalanceScanner // Generic contract binding to access the raw methods on
}

// BalanceScannerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BalanceScannerCallerRaw struct {
	Contract *BalanceScannerCaller // Generic read-only contract binding to access the raw methods on
}

// BalanceScannerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BalanceScannerTransactorRaw struct {
	Contract *BalanceScannerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBalanceScanner creates a new instance of BalanceScanner, bound to a specific deployed contract.
func NewBalanceScanner(address common.Address, backend bind.ContractBackend) (*BalanceScanner, error) {
	contract, err := bindBalanceScanner(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BalanceScanner{BalanceScannerCaller: BalanceScannerCaller{contract: contract}, BalanceScannerTransactor: BalanceScannerTransactor{contract: contract}, BalanceScannerFilterer: BalanceScannerFilterer{contract: contract}}, nil
}

// NewBalanceScannerCaller creates a new read-only instance of BalanceScanner, bound to a specific deployed contract.
func NewBalanceScannerCaller(address common.Address, caller bind.ContractCaller) (*BalanceScannerCaller, error) {
	contract, err := bindBalanceScanner(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BalanceScannerCaller{contract: contract}, nil
}

// NewBalanceScannerTransactor creates a new write-only instance of BalanceScanner, bound to a specific deployed contract.
func NewBalanceScannerTransactor(address common.Address, transactor bind.ContractTransactor) (*BalanceScannerTransactor, error) {
	contract, err := bindBalanceScanner(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BalanceScannerTransactor{contract: contract}, nil
}

// NewBalanceScannerFilterer creates a new log filterer instance of BalanceScanner, bound to a specific deployed contract.
func NewBalanceScannerFilterer(address common.Address, filterer bind.ContractFilterer) (*BalanceScannerFilterer, error) {
	contract, err := bindBalanceScanner(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BalanceScannerFilterer{contract: contract}, nil
}

// bindBalanceScanner binds a generic wrapper to an already deployed contract.
func bindBalanceScanner(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BalanceScannerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BalanceScanner *BalanceScannerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BalanceScanner.Contract.BalanceScannerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BalanceScanner *BalanceScannerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BalanceScanner.Contract.BalanceScannerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BalanceScanner *BalanceScannerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BalanceScanner.Contract.BalanceScannerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BalanceScanner *BalanceScannerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BalanceScanner.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BalanceScanner *BalanceScannerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BalanceScanner.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BalanceScanner *BalanceScannerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BalanceScanner.Contract.contract.Transact(opts, method, params...)
}

// Call is a free data retrieval call binding the contract method 0x36738374.
//
// Solidity: function call(address[] contracts, bytes[] data, uint256 gas) view returns((bool,bytes)[] results)
func (_BalanceScanner *BalanceScannerCaller) Call(opts *bind.CallOpts, contracts []common.Address, data [][]byte, gas *big.Int) ([]BalanceScannerResult, error) {
	var out []interface{}
	err := _BalanceScanner.contract.Call(opts, &out, "call", contracts, data, gas)

	if err != nil {
		return *new([]BalanceScannerResult), err
	}

	out0 := *abi.ConvertType(out[0], new([]BalanceScannerResult)).(*[]BalanceScannerResult)

	return out0, err

}

// Call is a free data retrieval call binding the contract method 0x36738374.
//
// Solidity: function call(address[] contracts, bytes[] data, uint256 gas) view returns((bool,bytes)[] results)
func (_BalanceScanner *BalanceScannerSession) Call(contracts []common.Address, data [][]byte, gas *big.Int) ([]BalanceScannerResult, error) {
	return _BalanceScanner.Contract.Call(&_BalanceScanner.CallOpts, contracts, data, gas)
}

// Call is a free data retrieval call binding the contract method 0x36738374.
//
// Solidity: function call(address[] contracts, bytes[] data, uint256 gas) view returns((bool,bytes)[] results)
func (_BalanceScanner *BalanceScannerCallerSession) Call(contracts []common.Address, data [][]byte, gas *big.Int) ([]BalanceScannerResult, error) {
	return _BalanceScanner.Contract.Call(&_BalanceScanner.CallOpts, contracts, data, gas)
}

// Call0 is a free data retrieval call binding the contract method 0x458b3a7c.
//
// Solidity: function call(address[] contracts, bytes[] data) view returns((bool,bytes)[] results)
func (_BalanceScanner *BalanceScannerCaller) Call0(opts *bind.CallOpts, contracts []common.Address, data [][]byte) ([]BalanceScannerResult, error) {
	var out []interface{}
	err := _BalanceScanner.contract.Call(opts, &out, "call0", contracts, data)

	if err != nil {
		return *new([]BalanceScannerResult), err
	}

	out0 := *abi.ConvertType(out[0], new([]BalanceScannerResult)).(*[]BalanceScannerResult)

	return out0, err

}

// Call0 is a free data retrieval call binding the contract method 0x458b3a7c.
//
// Solidity: function call(address[] contracts, bytes[] data) view returns((bool,bytes)[] results)
func (_BalanceScanner *BalanceScannerSession) Call0(contracts []common.Address, data [][]byte) ([]BalanceScannerResult, error) {
	return _BalanceScanner.Contract.Call0(&_BalanceScanner.CallOpts, contracts, data)
}

// Call0 is a free data retrieval call binding the contract method 0x458b3a7c.
//
// Solidity: function call(address[] contracts, bytes[] data) view returns((bool,bytes)[] results)
func (_BalanceScanner *BalanceScannerCallerSession) Call0(contracts []common.Address, data [][]byte) ([]BalanceScannerResult, error) {
	return _BalanceScanner.Contract.Call0(&_BalanceScanner.CallOpts, contracts, data)
}

// EtherBalances is a free data retrieval call binding the contract method 0xdbdbb51b.
//
// Solidity: function etherBalances(address[] addresses) view returns((bool,bytes)[] results)
func (_BalanceScanner *BalanceScannerCaller) EtherBalances(opts *bind.CallOpts, addresses []common.Address) ([]BalanceScannerResult, error) {
	var out []interface{}
	err := _BalanceScanner.contract.Call(opts, &out, "etherBalances", addresses)

	if err != nil {
		return *new([]BalanceScannerResult), err
	}

	out0 := *abi.ConvertType(out[0], new([]BalanceScannerResult)).(*[]BalanceScannerResult)

	return out0, err

}

// EtherBalances is a free data retrieval call binding the contract method 0xdbdbb51b.
//
// Solidity: function etherBalances(address[] addresses) view returns((bool,bytes)[] results)
func (_BalanceScanner *BalanceScannerSession) EtherBalances(addresses []common.Address) ([]BalanceScannerResult, error) {
	return _BalanceScanner.Contract.EtherBalances(&_BalanceScanner.CallOpts, addresses)
}

// EtherBalances is a free data retrieval call binding the contract method 0xdbdbb51b.
//
// Solidity: function etherBalances(address[] addresses) view returns((bool,bytes)[] results)
func (_BalanceScanner *BalanceScannerCallerSession) EtherBalances(addresses []common.Address) ([]BalanceScannerResult, error) {
	return _BalanceScanner.Contract.EtherBalances(&_BalanceScanner.CallOpts, addresses)
}

// TokenBalances is a free data retrieval call binding the contract method 0xaad33091.
//
// Solidity: function tokenBalances(address[] addresses, address token) view returns((bool,bytes)[] results)
func (_BalanceScanner *BalanceScannerCaller) TokenBalances(opts *bind.CallOpts, addresses []common.Address, token common.Address) ([]BalanceScannerResult, error) {
	var out []interface{}
	err := _BalanceScanner.contract.Call(opts, &out, "tokenBalances", addresses, token)

	if err != nil {
		return *new([]BalanceScannerResult), err
	}

	out0 := *abi.ConvertType(out[0], new([]BalanceScannerResult)).(*[]BalanceScannerResult)

	return out0, err

}

// TokenBalances is a free data retrieval call binding the contract method 0xaad33091.
//
// Solidity: function tokenBalances(address[] addresses, address token) view returns((bool,bytes)[] results)
func (_BalanceScanner *BalanceScannerSession) TokenBalances(addresses []common.Address, token common.Address) ([]BalanceScannerResult, error) {
	return _BalanceScanner.Contract.TokenBalances(&_BalanceScanner.CallOpts, addresses, token)
}

// TokenBalances is a free data retrieval call binding the contract method 0xaad33091.
//
// Solidity: function tokenBalances(address[] addresses, address token) view returns((bool,bytes)[] results)
func (_BalanceScanner *BalanceScannerCallerSession) TokenBalances(addresses []common.Address, token common.Address) ([]BalanceScannerResult, error) {
	return _BalanceScanner.Contract.TokenBalances(&_BalanceScanner.CallOpts, addresses, token)
}

// TokensBalance is a free data retrieval call binding the contract method 0xe5da1b68.
//
// Solidity: function tokensBalance(address owner, address[] contracts) view returns((bool,bytes)[] results)
func (_BalanceScanner *BalanceScannerCaller) TokensBalance(opts *bind.CallOpts, owner common.Address, contracts []common.Address) ([]BalanceScannerResult, error) {
	var out []interface{}
	err := _BalanceScanner.contract.Call(opts, &out, "tokensBalance", owner, contracts)

	if err != nil {
		return *new([]BalanceScannerResult), err
	}

	out0 := *abi.ConvertType(out[0], new([]BalanceScannerResult)).(*[]BalanceScannerResult)

	return out0, err

}

// TokensBalance is a free data retrieval call binding the contract method 0xe5da1b68.
//
// Solidity: function tokensBalance(address owner, address[] contracts) view returns((bool,bytes)[] results)
func (_BalanceScanner *BalanceScannerSession) TokensBalance(owner common.Address, contracts []common.Address) ([]BalanceScannerResult, error) {
	return _BalanceScanner.Contract.TokensBalance(&_BalanceScanner.CallOpts, owner, contracts)
}

// TokensBalance is a free data retrieval call binding the contract method 0xe5da1b68.
//
// Solidity: function tokensBalance(address owner, address[] contracts) view returns((bool,bytes)[] results)
func (_BalanceScanner *BalanceScannerCallerSession) TokensBalance(owner common.Address, contracts []common.Address) ([]BalanceScannerResult, error) {
	return _BalanceScanner.Contract.TokensBalance(&_BalanceScanner.CallOpts, owner, contracts)
}
