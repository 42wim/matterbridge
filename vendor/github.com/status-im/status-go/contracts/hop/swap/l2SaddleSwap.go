// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package hopSwap

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

// HopSwapABI is the input ABI used to generate the binding from.
const HopSwapABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"tokenAmounts\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"fees\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"invariant\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lpTokenSupply\",\"type\":\"uint256\"}],\"name\":\"AddLiquidity\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newAdminFee\",\"type\":\"uint256\"}],\"name\":\"NewAdminFee\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newSwapFee\",\"type\":\"uint256\"}],\"name\":\"NewSwapFee\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newWithdrawFee\",\"type\":\"uint256\"}],\"name\":\"NewWithdrawFee\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"oldA\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newA\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"initialTime\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"futureTime\",\"type\":\"uint256\"}],\"name\":\"RampA\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"tokenAmounts\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lpTokenSupply\",\"type\":\"uint256\"}],\"name\":\"RemoveLiquidity\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"tokenAmounts\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"fees\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"invariant\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lpTokenSupply\",\"type\":\"uint256\"}],\"name\":\"RemoveLiquidityImbalance\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lpTokenAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lpTokenSupply\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"boughtId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokensBought\",\"type\":\"uint256\"}],\"name\":\"RemoveLiquidityOne\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"currentA\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"}],\"name\":\"StopRampA\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"buyer\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokensSold\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokensBought\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"soldId\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"boughtId\",\"type\":\"uint128\"}],\"name\":\"TokenSwap\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"amounts\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256\",\"name\":\"minToMint\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"name\":\"addLiquidity\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"}],\"name\":\"calculateCurrentWithdrawFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"calculateRemoveLiquidity\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"tokenIndex\",\"type\":\"uint8\"}],\"name\":\"calculateRemoveLiquidityOneToken\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"availableTokenAmount\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"tokenIndexFrom\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"tokenIndexTo\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"dx\",\"type\":\"uint256\"}],\"name\":\"calculateSwap\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256[]\",\"name\":\"amounts\",\"type\":\"uint256[]\"},{\"internalType\":\"bool\",\"name\":\"deposit\",\"type\":\"bool\"}],\"name\":\"calculateTokenAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getA\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAPrecise\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getAdminBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"}],\"name\":\"getDepositTimestamp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"index\",\"type\":\"uint8\"}],\"name\":\"getToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"index\",\"type\":\"uint8\"}],\"name\":\"getTokenBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"}],\"name\":\"getTokenIndex\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getVirtualPrice\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIERC20[]\",\"name\":\"_pooledTokens\",\"type\":\"address[]\"},{\"internalType\":\"uint8[]\",\"name\":\"decimals\",\"type\":\"uint8[]\"},{\"internalType\":\"string\",\"name\":\"lpTokenName\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"lpTokenSymbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_a\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_fee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_adminFee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_withdrawFee\",\"type\":\"uint256\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256[]\",\"name\":\"minAmounts\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"name\":\"removeLiquidity\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"amounts\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256\",\"name\":\"maxBurnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"name\":\"removeLiquidityImbalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"tokenIndex\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"minAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"name\":\"removeLiquidityOneToken\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"tokenIndexFrom\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"tokenIndexTo\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"dx\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minDy\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"name\":\"swap\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"swapStorage\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"initialA\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"futureA\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"initialATime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"futureATime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"swapFee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"adminFee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"defaultWithdrawFee\",\"type\":\"uint256\"},{\"internalType\":\"contractLPToken\",\"name\":\"lpToken\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"transferAmount\",\"type\":\"uint256\"}],\"name\":\"updateUserWithdrawFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// HopSwap is an auto generated Go binding around an Ethereum contract.
type HopSwap struct {
	HopSwapCaller     // Read-only binding to the contract
	HopSwapTransactor // Write-only binding to the contract
	HopSwapFilterer   // Log filterer for contract events
}

// HopSwapCaller is an auto generated read-only Go binding around an Ethereum contract.
type HopSwapCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HopSwapTransactor is an auto generated write-only Go binding around an Ethereum contract.
type HopSwapTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HopSwapFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type HopSwapFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HopSwapSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type HopSwapSession struct {
	Contract     *HopSwap          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HopSwapCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type HopSwapCallerSession struct {
	Contract *HopSwapCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// HopSwapTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type HopSwapTransactorSession struct {
	Contract     *HopSwapTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// HopSwapRaw is an auto generated low-level Go binding around an Ethereum contract.
type HopSwapRaw struct {
	Contract *HopSwap // Generic contract binding to access the raw methods on
}

// HopSwapCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type HopSwapCallerRaw struct {
	Contract *HopSwapCaller // Generic read-only contract binding to access the raw methods on
}

// HopSwapTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type HopSwapTransactorRaw struct {
	Contract *HopSwapTransactor // Generic write-only contract binding to access the raw methods on
}

// NewHopSwap creates a new instance of HopSwap, bound to a specific deployed contract.
func NewHopSwap(address common.Address, backend bind.ContractBackend) (*HopSwap, error) {
	contract, err := bindHopSwap(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &HopSwap{HopSwapCaller: HopSwapCaller{contract: contract}, HopSwapTransactor: HopSwapTransactor{contract: contract}, HopSwapFilterer: HopSwapFilterer{contract: contract}}, nil
}

// NewHopSwapCaller creates a new read-only instance of HopSwap, bound to a specific deployed contract.
func NewHopSwapCaller(address common.Address, caller bind.ContractCaller) (*HopSwapCaller, error) {
	contract, err := bindHopSwap(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &HopSwapCaller{contract: contract}, nil
}

// NewHopSwapTransactor creates a new write-only instance of HopSwap, bound to a specific deployed contract.
func NewHopSwapTransactor(address common.Address, transactor bind.ContractTransactor) (*HopSwapTransactor, error) {
	contract, err := bindHopSwap(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &HopSwapTransactor{contract: contract}, nil
}

// NewHopSwapFilterer creates a new log filterer instance of HopSwap, bound to a specific deployed contract.
func NewHopSwapFilterer(address common.Address, filterer bind.ContractFilterer) (*HopSwapFilterer, error) {
	contract, err := bindHopSwap(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &HopSwapFilterer{contract: contract}, nil
}

// bindHopSwap binds a generic wrapper to an already deployed contract.
func bindHopSwap(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(HopSwapABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HopSwap *HopSwapRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _HopSwap.Contract.HopSwapCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HopSwap *HopSwapRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HopSwap.Contract.HopSwapTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HopSwap *HopSwapRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HopSwap.Contract.HopSwapTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HopSwap *HopSwapCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _HopSwap.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HopSwap *HopSwapTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HopSwap.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HopSwap *HopSwapTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HopSwap.Contract.contract.Transact(opts, method, params...)
}

// CalculateCurrentWithdrawFee is a free data retrieval call binding the contract method 0x4a1b0d57.
//
// Solidity: function calculateCurrentWithdrawFee(address user) view returns(uint256)
func (_HopSwap *HopSwapCaller) CalculateCurrentWithdrawFee(opts *bind.CallOpts, user common.Address) (*big.Int, error) {
	var out []interface{}
	err := _HopSwap.contract.Call(opts, &out, "calculateCurrentWithdrawFee", user)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculateCurrentWithdrawFee is a free data retrieval call binding the contract method 0x4a1b0d57.
//
// Solidity: function calculateCurrentWithdrawFee(address user) view returns(uint256)
func (_HopSwap *HopSwapSession) CalculateCurrentWithdrawFee(user common.Address) (*big.Int, error) {
	return _HopSwap.Contract.CalculateCurrentWithdrawFee(&_HopSwap.CallOpts, user)
}

// CalculateCurrentWithdrawFee is a free data retrieval call binding the contract method 0x4a1b0d57.
//
// Solidity: function calculateCurrentWithdrawFee(address user) view returns(uint256)
func (_HopSwap *HopSwapCallerSession) CalculateCurrentWithdrawFee(user common.Address) (*big.Int, error) {
	return _HopSwap.Contract.CalculateCurrentWithdrawFee(&_HopSwap.CallOpts, user)
}

// CalculateRemoveLiquidity is a free data retrieval call binding the contract method 0x7c61e561.
//
// Solidity: function calculateRemoveLiquidity(address account, uint256 amount) view returns(uint256[])
func (_HopSwap *HopSwapCaller) CalculateRemoveLiquidity(opts *bind.CallOpts, account common.Address, amount *big.Int) ([]*big.Int, error) {
	var out []interface{}
	err := _HopSwap.contract.Call(opts, &out, "calculateRemoveLiquidity", account, amount)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// CalculateRemoveLiquidity is a free data retrieval call binding the contract method 0x7c61e561.
//
// Solidity: function calculateRemoveLiquidity(address account, uint256 amount) view returns(uint256[])
func (_HopSwap *HopSwapSession) CalculateRemoveLiquidity(account common.Address, amount *big.Int) ([]*big.Int, error) {
	return _HopSwap.Contract.CalculateRemoveLiquidity(&_HopSwap.CallOpts, account, amount)
}

// CalculateRemoveLiquidity is a free data retrieval call binding the contract method 0x7c61e561.
//
// Solidity: function calculateRemoveLiquidity(address account, uint256 amount) view returns(uint256[])
func (_HopSwap *HopSwapCallerSession) CalculateRemoveLiquidity(account common.Address, amount *big.Int) ([]*big.Int, error) {
	return _HopSwap.Contract.CalculateRemoveLiquidity(&_HopSwap.CallOpts, account, amount)
}

// CalculateRemoveLiquidityOneToken is a free data retrieval call binding the contract method 0x98899f40.
//
// Solidity: function calculateRemoveLiquidityOneToken(address account, uint256 tokenAmount, uint8 tokenIndex) view returns(uint256 availableTokenAmount)
func (_HopSwap *HopSwapCaller) CalculateRemoveLiquidityOneToken(opts *bind.CallOpts, account common.Address, tokenAmount *big.Int, tokenIndex uint8) (*big.Int, error) {
	var out []interface{}
	err := _HopSwap.contract.Call(opts, &out, "calculateRemoveLiquidityOneToken", account, tokenAmount, tokenIndex)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculateRemoveLiquidityOneToken is a free data retrieval call binding the contract method 0x98899f40.
//
// Solidity: function calculateRemoveLiquidityOneToken(address account, uint256 tokenAmount, uint8 tokenIndex) view returns(uint256 availableTokenAmount)
func (_HopSwap *HopSwapSession) CalculateRemoveLiquidityOneToken(account common.Address, tokenAmount *big.Int, tokenIndex uint8) (*big.Int, error) {
	return _HopSwap.Contract.CalculateRemoveLiquidityOneToken(&_HopSwap.CallOpts, account, tokenAmount, tokenIndex)
}

// CalculateRemoveLiquidityOneToken is a free data retrieval call binding the contract method 0x98899f40.
//
// Solidity: function calculateRemoveLiquidityOneToken(address account, uint256 tokenAmount, uint8 tokenIndex) view returns(uint256 availableTokenAmount)
func (_HopSwap *HopSwapCallerSession) CalculateRemoveLiquidityOneToken(account common.Address, tokenAmount *big.Int, tokenIndex uint8) (*big.Int, error) {
	return _HopSwap.Contract.CalculateRemoveLiquidityOneToken(&_HopSwap.CallOpts, account, tokenAmount, tokenIndex)
}

// CalculateSwap is a free data retrieval call binding the contract method 0xa95b089f.
//
// Solidity: function calculateSwap(uint8 tokenIndexFrom, uint8 tokenIndexTo, uint256 dx) view returns(uint256)
func (_HopSwap *HopSwapCaller) CalculateSwap(opts *bind.CallOpts, tokenIndexFrom uint8, tokenIndexTo uint8, dx *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _HopSwap.contract.Call(opts, &out, "calculateSwap", tokenIndexFrom, tokenIndexTo, dx)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculateSwap is a free data retrieval call binding the contract method 0xa95b089f.
//
// Solidity: function calculateSwap(uint8 tokenIndexFrom, uint8 tokenIndexTo, uint256 dx) view returns(uint256)
func (_HopSwap *HopSwapSession) CalculateSwap(tokenIndexFrom uint8, tokenIndexTo uint8, dx *big.Int) (*big.Int, error) {
	return _HopSwap.Contract.CalculateSwap(&_HopSwap.CallOpts, tokenIndexFrom, tokenIndexTo, dx)
}

// CalculateSwap is a free data retrieval call binding the contract method 0xa95b089f.
//
// Solidity: function calculateSwap(uint8 tokenIndexFrom, uint8 tokenIndexTo, uint256 dx) view returns(uint256)
func (_HopSwap *HopSwapCallerSession) CalculateSwap(tokenIndexFrom uint8, tokenIndexTo uint8, dx *big.Int) (*big.Int, error) {
	return _HopSwap.Contract.CalculateSwap(&_HopSwap.CallOpts, tokenIndexFrom, tokenIndexTo, dx)
}

// CalculateTokenAmount is a free data retrieval call binding the contract method 0xf9273ffb.
//
// Solidity: function calculateTokenAmount(address account, uint256[] amounts, bool deposit) view returns(uint256)
func (_HopSwap *HopSwapCaller) CalculateTokenAmount(opts *bind.CallOpts, account common.Address, amounts []*big.Int, deposit bool) (*big.Int, error) {
	var out []interface{}
	err := _HopSwap.contract.Call(opts, &out, "calculateTokenAmount", account, amounts, deposit)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculateTokenAmount is a free data retrieval call binding the contract method 0xf9273ffb.
//
// Solidity: function calculateTokenAmount(address account, uint256[] amounts, bool deposit) view returns(uint256)
func (_HopSwap *HopSwapSession) CalculateTokenAmount(account common.Address, amounts []*big.Int, deposit bool) (*big.Int, error) {
	return _HopSwap.Contract.CalculateTokenAmount(&_HopSwap.CallOpts, account, amounts, deposit)
}

// CalculateTokenAmount is a free data retrieval call binding the contract method 0xf9273ffb.
//
// Solidity: function calculateTokenAmount(address account, uint256[] amounts, bool deposit) view returns(uint256)
func (_HopSwap *HopSwapCallerSession) CalculateTokenAmount(account common.Address, amounts []*big.Int, deposit bool) (*big.Int, error) {
	return _HopSwap.Contract.CalculateTokenAmount(&_HopSwap.CallOpts, account, amounts, deposit)
}

// GetA is a free data retrieval call binding the contract method 0xd46300fd.
//
// Solidity: function getA() view returns(uint256)
func (_HopSwap *HopSwapCaller) GetA(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HopSwap.contract.Call(opts, &out, "getA")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetA is a free data retrieval call binding the contract method 0xd46300fd.
//
// Solidity: function getA() view returns(uint256)
func (_HopSwap *HopSwapSession) GetA() (*big.Int, error) {
	return _HopSwap.Contract.GetA(&_HopSwap.CallOpts)
}

// GetA is a free data retrieval call binding the contract method 0xd46300fd.
//
// Solidity: function getA() view returns(uint256)
func (_HopSwap *HopSwapCallerSession) GetA() (*big.Int, error) {
	return _HopSwap.Contract.GetA(&_HopSwap.CallOpts)
}

// GetAPrecise is a free data retrieval call binding the contract method 0x0ba81959.
//
// Solidity: function getAPrecise() view returns(uint256)
func (_HopSwap *HopSwapCaller) GetAPrecise(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HopSwap.contract.Call(opts, &out, "getAPrecise")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAPrecise is a free data retrieval call binding the contract method 0x0ba81959.
//
// Solidity: function getAPrecise() view returns(uint256)
func (_HopSwap *HopSwapSession) GetAPrecise() (*big.Int, error) {
	return _HopSwap.Contract.GetAPrecise(&_HopSwap.CallOpts)
}

// GetAPrecise is a free data retrieval call binding the contract method 0x0ba81959.
//
// Solidity: function getAPrecise() view returns(uint256)
func (_HopSwap *HopSwapCallerSession) GetAPrecise() (*big.Int, error) {
	return _HopSwap.Contract.GetAPrecise(&_HopSwap.CallOpts)
}

// GetAdminBalance is a free data retrieval call binding the contract method 0xef0a712f.
//
// Solidity: function getAdminBalance(uint256 index) view returns(uint256)
func (_HopSwap *HopSwapCaller) GetAdminBalance(opts *bind.CallOpts, index *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _HopSwap.contract.Call(opts, &out, "getAdminBalance", index)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAdminBalance is a free data retrieval call binding the contract method 0xef0a712f.
//
// Solidity: function getAdminBalance(uint256 index) view returns(uint256)
func (_HopSwap *HopSwapSession) GetAdminBalance(index *big.Int) (*big.Int, error) {
	return _HopSwap.Contract.GetAdminBalance(&_HopSwap.CallOpts, index)
}

// GetAdminBalance is a free data retrieval call binding the contract method 0xef0a712f.
//
// Solidity: function getAdminBalance(uint256 index) view returns(uint256)
func (_HopSwap *HopSwapCallerSession) GetAdminBalance(index *big.Int) (*big.Int, error) {
	return _HopSwap.Contract.GetAdminBalance(&_HopSwap.CallOpts, index)
}

// GetDepositTimestamp is a free data retrieval call binding the contract method 0xda7a77be.
//
// Solidity: function getDepositTimestamp(address user) view returns(uint256)
func (_HopSwap *HopSwapCaller) GetDepositTimestamp(opts *bind.CallOpts, user common.Address) (*big.Int, error) {
	var out []interface{}
	err := _HopSwap.contract.Call(opts, &out, "getDepositTimestamp", user)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetDepositTimestamp is a free data retrieval call binding the contract method 0xda7a77be.
//
// Solidity: function getDepositTimestamp(address user) view returns(uint256)
func (_HopSwap *HopSwapSession) GetDepositTimestamp(user common.Address) (*big.Int, error) {
	return _HopSwap.Contract.GetDepositTimestamp(&_HopSwap.CallOpts, user)
}

// GetDepositTimestamp is a free data retrieval call binding the contract method 0xda7a77be.
//
// Solidity: function getDepositTimestamp(address user) view returns(uint256)
func (_HopSwap *HopSwapCallerSession) GetDepositTimestamp(user common.Address) (*big.Int, error) {
	return _HopSwap.Contract.GetDepositTimestamp(&_HopSwap.CallOpts, user)
}

// GetToken is a free data retrieval call binding the contract method 0x82b86600.
//
// Solidity: function getToken(uint8 index) view returns(address)
func (_HopSwap *HopSwapCaller) GetToken(opts *bind.CallOpts, index uint8) (common.Address, error) {
	var out []interface{}
	err := _HopSwap.contract.Call(opts, &out, "getToken", index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetToken is a free data retrieval call binding the contract method 0x82b86600.
//
// Solidity: function getToken(uint8 index) view returns(address)
func (_HopSwap *HopSwapSession) GetToken(index uint8) (common.Address, error) {
	return _HopSwap.Contract.GetToken(&_HopSwap.CallOpts, index)
}

// GetToken is a free data retrieval call binding the contract method 0x82b86600.
//
// Solidity: function getToken(uint8 index) view returns(address)
func (_HopSwap *HopSwapCallerSession) GetToken(index uint8) (common.Address, error) {
	return _HopSwap.Contract.GetToken(&_HopSwap.CallOpts, index)
}

// GetTokenBalance is a free data retrieval call binding the contract method 0x91ceb3eb.
//
// Solidity: function getTokenBalance(uint8 index) view returns(uint256)
func (_HopSwap *HopSwapCaller) GetTokenBalance(opts *bind.CallOpts, index uint8) (*big.Int, error) {
	var out []interface{}
	err := _HopSwap.contract.Call(opts, &out, "getTokenBalance", index)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTokenBalance is a free data retrieval call binding the contract method 0x91ceb3eb.
//
// Solidity: function getTokenBalance(uint8 index) view returns(uint256)
func (_HopSwap *HopSwapSession) GetTokenBalance(index uint8) (*big.Int, error) {
	return _HopSwap.Contract.GetTokenBalance(&_HopSwap.CallOpts, index)
}

// GetTokenBalance is a free data retrieval call binding the contract method 0x91ceb3eb.
//
// Solidity: function getTokenBalance(uint8 index) view returns(uint256)
func (_HopSwap *HopSwapCallerSession) GetTokenBalance(index uint8) (*big.Int, error) {
	return _HopSwap.Contract.GetTokenBalance(&_HopSwap.CallOpts, index)
}

// GetTokenIndex is a free data retrieval call binding the contract method 0x66c0bd24.
//
// Solidity: function getTokenIndex(address tokenAddress) view returns(uint8)
func (_HopSwap *HopSwapCaller) GetTokenIndex(opts *bind.CallOpts, tokenAddress common.Address) (uint8, error) {
	var out []interface{}
	err := _HopSwap.contract.Call(opts, &out, "getTokenIndex", tokenAddress)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetTokenIndex is a free data retrieval call binding the contract method 0x66c0bd24.
//
// Solidity: function getTokenIndex(address tokenAddress) view returns(uint8)
func (_HopSwap *HopSwapSession) GetTokenIndex(tokenAddress common.Address) (uint8, error) {
	return _HopSwap.Contract.GetTokenIndex(&_HopSwap.CallOpts, tokenAddress)
}

// GetTokenIndex is a free data retrieval call binding the contract method 0x66c0bd24.
//
// Solidity: function getTokenIndex(address tokenAddress) view returns(uint8)
func (_HopSwap *HopSwapCallerSession) GetTokenIndex(tokenAddress common.Address) (uint8, error) {
	return _HopSwap.Contract.GetTokenIndex(&_HopSwap.CallOpts, tokenAddress)
}

// GetVirtualPrice is a free data retrieval call binding the contract method 0xe25aa5fa.
//
// Solidity: function getVirtualPrice() view returns(uint256)
func (_HopSwap *HopSwapCaller) GetVirtualPrice(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HopSwap.contract.Call(opts, &out, "getVirtualPrice")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetVirtualPrice is a free data retrieval call binding the contract method 0xe25aa5fa.
//
// Solidity: function getVirtualPrice() view returns(uint256)
func (_HopSwap *HopSwapSession) GetVirtualPrice() (*big.Int, error) {
	return _HopSwap.Contract.GetVirtualPrice(&_HopSwap.CallOpts)
}

// GetVirtualPrice is a free data retrieval call binding the contract method 0xe25aa5fa.
//
// Solidity: function getVirtualPrice() view returns(uint256)
func (_HopSwap *HopSwapCallerSession) GetVirtualPrice() (*big.Int, error) {
	return _HopSwap.Contract.GetVirtualPrice(&_HopSwap.CallOpts)
}

// SwapStorage is a free data retrieval call binding the contract method 0x5fd65f0f.
//
// Solidity: function swapStorage() view returns(uint256 initialA, uint256 futureA, uint256 initialATime, uint256 futureATime, uint256 swapFee, uint256 adminFee, uint256 defaultWithdrawFee, address lpToken)
func (_HopSwap *HopSwapCaller) SwapStorage(opts *bind.CallOpts) (struct {
	InitialA           *big.Int
	FutureA            *big.Int
	InitialATime       *big.Int
	FutureATime        *big.Int
	SwapFee            *big.Int
	AdminFee           *big.Int
	DefaultWithdrawFee *big.Int
	LpToken            common.Address
}, error) {
	var out []interface{}
	err := _HopSwap.contract.Call(opts, &out, "swapStorage")

	outstruct := new(struct {
		InitialA           *big.Int
		FutureA            *big.Int
		InitialATime       *big.Int
		FutureATime        *big.Int
		SwapFee            *big.Int
		AdminFee           *big.Int
		DefaultWithdrawFee *big.Int
		LpToken            common.Address
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.InitialA = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.FutureA = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.InitialATime = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.FutureATime = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.SwapFee = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.AdminFee = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.DefaultWithdrawFee = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)
	outstruct.LpToken = *abi.ConvertType(out[7], new(common.Address)).(*common.Address)

	return *outstruct, err

}

// SwapStorage is a free data retrieval call binding the contract method 0x5fd65f0f.
//
// Solidity: function swapStorage() view returns(uint256 initialA, uint256 futureA, uint256 initialATime, uint256 futureATime, uint256 swapFee, uint256 adminFee, uint256 defaultWithdrawFee, address lpToken)
func (_HopSwap *HopSwapSession) SwapStorage() (struct {
	InitialA           *big.Int
	FutureA            *big.Int
	InitialATime       *big.Int
	FutureATime        *big.Int
	SwapFee            *big.Int
	AdminFee           *big.Int
	DefaultWithdrawFee *big.Int
	LpToken            common.Address
}, error) {
	return _HopSwap.Contract.SwapStorage(&_HopSwap.CallOpts)
}

// SwapStorage is a free data retrieval call binding the contract method 0x5fd65f0f.
//
// Solidity: function swapStorage() view returns(uint256 initialA, uint256 futureA, uint256 initialATime, uint256 futureATime, uint256 swapFee, uint256 adminFee, uint256 defaultWithdrawFee, address lpToken)
func (_HopSwap *HopSwapCallerSession) SwapStorage() (struct {
	InitialA           *big.Int
	FutureA            *big.Int
	InitialATime       *big.Int
	FutureATime        *big.Int
	SwapFee            *big.Int
	AdminFee           *big.Int
	DefaultWithdrawFee *big.Int
	LpToken            common.Address
}, error) {
	return _HopSwap.Contract.SwapStorage(&_HopSwap.CallOpts)
}

// AddLiquidity is a paid mutator transaction binding the contract method 0x4d49e87d.
//
// Solidity: function addLiquidity(uint256[] amounts, uint256 minToMint, uint256 deadline) returns(uint256)
func (_HopSwap *HopSwapTransactor) AddLiquidity(opts *bind.TransactOpts, amounts []*big.Int, minToMint *big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopSwap.contract.Transact(opts, "addLiquidity", amounts, minToMint, deadline)
}

// AddLiquidity is a paid mutator transaction binding the contract method 0x4d49e87d.
//
// Solidity: function addLiquidity(uint256[] amounts, uint256 minToMint, uint256 deadline) returns(uint256)
func (_HopSwap *HopSwapSession) AddLiquidity(amounts []*big.Int, minToMint *big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopSwap.Contract.AddLiquidity(&_HopSwap.TransactOpts, amounts, minToMint, deadline)
}

// AddLiquidity is a paid mutator transaction binding the contract method 0x4d49e87d.
//
// Solidity: function addLiquidity(uint256[] amounts, uint256 minToMint, uint256 deadline) returns(uint256)
func (_HopSwap *HopSwapTransactorSession) AddLiquidity(amounts []*big.Int, minToMint *big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopSwap.Contract.AddLiquidity(&_HopSwap.TransactOpts, amounts, minToMint, deadline)
}

// Initialize is a paid mutator transaction binding the contract method 0x6dd4480b.
//
// Solidity: function initialize(address[] _pooledTokens, uint8[] decimals, string lpTokenName, string lpTokenSymbol, uint256 _a, uint256 _fee, uint256 _adminFee, uint256 _withdrawFee) returns()
func (_HopSwap *HopSwapTransactor) Initialize(opts *bind.TransactOpts, _pooledTokens []common.Address, decimals []uint8, lpTokenName string, lpTokenSymbol string, _a *big.Int, _fee *big.Int, _adminFee *big.Int, _withdrawFee *big.Int) (*types.Transaction, error) {
	return _HopSwap.contract.Transact(opts, "initialize", _pooledTokens, decimals, lpTokenName, lpTokenSymbol, _a, _fee, _adminFee, _withdrawFee)
}

// Initialize is a paid mutator transaction binding the contract method 0x6dd4480b.
//
// Solidity: function initialize(address[] _pooledTokens, uint8[] decimals, string lpTokenName, string lpTokenSymbol, uint256 _a, uint256 _fee, uint256 _adminFee, uint256 _withdrawFee) returns()
func (_HopSwap *HopSwapSession) Initialize(_pooledTokens []common.Address, decimals []uint8, lpTokenName string, lpTokenSymbol string, _a *big.Int, _fee *big.Int, _adminFee *big.Int, _withdrawFee *big.Int) (*types.Transaction, error) {
	return _HopSwap.Contract.Initialize(&_HopSwap.TransactOpts, _pooledTokens, decimals, lpTokenName, lpTokenSymbol, _a, _fee, _adminFee, _withdrawFee)
}

// Initialize is a paid mutator transaction binding the contract method 0x6dd4480b.
//
// Solidity: function initialize(address[] _pooledTokens, uint8[] decimals, string lpTokenName, string lpTokenSymbol, uint256 _a, uint256 _fee, uint256 _adminFee, uint256 _withdrawFee) returns()
func (_HopSwap *HopSwapTransactorSession) Initialize(_pooledTokens []common.Address, decimals []uint8, lpTokenName string, lpTokenSymbol string, _a *big.Int, _fee *big.Int, _adminFee *big.Int, _withdrawFee *big.Int) (*types.Transaction, error) {
	return _HopSwap.Contract.Initialize(&_HopSwap.TransactOpts, _pooledTokens, decimals, lpTokenName, lpTokenSymbol, _a, _fee, _adminFee, _withdrawFee)
}

// RemoveLiquidity is a paid mutator transaction binding the contract method 0x31cd52b0.
//
// Solidity: function removeLiquidity(uint256 amount, uint256[] minAmounts, uint256 deadline) returns(uint256[])
func (_HopSwap *HopSwapTransactor) RemoveLiquidity(opts *bind.TransactOpts, amount *big.Int, minAmounts []*big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopSwap.contract.Transact(opts, "removeLiquidity", amount, minAmounts, deadline)
}

// RemoveLiquidity is a paid mutator transaction binding the contract method 0x31cd52b0.
//
// Solidity: function removeLiquidity(uint256 amount, uint256[] minAmounts, uint256 deadline) returns(uint256[])
func (_HopSwap *HopSwapSession) RemoveLiquidity(amount *big.Int, minAmounts []*big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopSwap.Contract.RemoveLiquidity(&_HopSwap.TransactOpts, amount, minAmounts, deadline)
}

// RemoveLiquidity is a paid mutator transaction binding the contract method 0x31cd52b0.
//
// Solidity: function removeLiquidity(uint256 amount, uint256[] minAmounts, uint256 deadline) returns(uint256[])
func (_HopSwap *HopSwapTransactorSession) RemoveLiquidity(amount *big.Int, minAmounts []*big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopSwap.Contract.RemoveLiquidity(&_HopSwap.TransactOpts, amount, minAmounts, deadline)
}

// RemoveLiquidityImbalance is a paid mutator transaction binding the contract method 0x84cdd9bc.
//
// Solidity: function removeLiquidityImbalance(uint256[] amounts, uint256 maxBurnAmount, uint256 deadline) returns(uint256)
func (_HopSwap *HopSwapTransactor) RemoveLiquidityImbalance(opts *bind.TransactOpts, amounts []*big.Int, maxBurnAmount *big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopSwap.contract.Transact(opts, "removeLiquidityImbalance", amounts, maxBurnAmount, deadline)
}

// RemoveLiquidityImbalance is a paid mutator transaction binding the contract method 0x84cdd9bc.
//
// Solidity: function removeLiquidityImbalance(uint256[] amounts, uint256 maxBurnAmount, uint256 deadline) returns(uint256)
func (_HopSwap *HopSwapSession) RemoveLiquidityImbalance(amounts []*big.Int, maxBurnAmount *big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopSwap.Contract.RemoveLiquidityImbalance(&_HopSwap.TransactOpts, amounts, maxBurnAmount, deadline)
}

// RemoveLiquidityImbalance is a paid mutator transaction binding the contract method 0x84cdd9bc.
//
// Solidity: function removeLiquidityImbalance(uint256[] amounts, uint256 maxBurnAmount, uint256 deadline) returns(uint256)
func (_HopSwap *HopSwapTransactorSession) RemoveLiquidityImbalance(amounts []*big.Int, maxBurnAmount *big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopSwap.Contract.RemoveLiquidityImbalance(&_HopSwap.TransactOpts, amounts, maxBurnAmount, deadline)
}

// RemoveLiquidityOneToken is a paid mutator transaction binding the contract method 0x3e3a1560.
//
// Solidity: function removeLiquidityOneToken(uint256 tokenAmount, uint8 tokenIndex, uint256 minAmount, uint256 deadline) returns(uint256)
func (_HopSwap *HopSwapTransactor) RemoveLiquidityOneToken(opts *bind.TransactOpts, tokenAmount *big.Int, tokenIndex uint8, minAmount *big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopSwap.contract.Transact(opts, "removeLiquidityOneToken", tokenAmount, tokenIndex, minAmount, deadline)
}

// RemoveLiquidityOneToken is a paid mutator transaction binding the contract method 0x3e3a1560.
//
// Solidity: function removeLiquidityOneToken(uint256 tokenAmount, uint8 tokenIndex, uint256 minAmount, uint256 deadline) returns(uint256)
func (_HopSwap *HopSwapSession) RemoveLiquidityOneToken(tokenAmount *big.Int, tokenIndex uint8, minAmount *big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopSwap.Contract.RemoveLiquidityOneToken(&_HopSwap.TransactOpts, tokenAmount, tokenIndex, minAmount, deadline)
}

// RemoveLiquidityOneToken is a paid mutator transaction binding the contract method 0x3e3a1560.
//
// Solidity: function removeLiquidityOneToken(uint256 tokenAmount, uint8 tokenIndex, uint256 minAmount, uint256 deadline) returns(uint256)
func (_HopSwap *HopSwapTransactorSession) RemoveLiquidityOneToken(tokenAmount *big.Int, tokenIndex uint8, minAmount *big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopSwap.Contract.RemoveLiquidityOneToken(&_HopSwap.TransactOpts, tokenAmount, tokenIndex, minAmount, deadline)
}

// Swap is a paid mutator transaction binding the contract method 0x91695586.
//
// Solidity: function swap(uint8 tokenIndexFrom, uint8 tokenIndexTo, uint256 dx, uint256 minDy, uint256 deadline) returns(uint256)
func (_HopSwap *HopSwapTransactor) Swap(opts *bind.TransactOpts, tokenIndexFrom uint8, tokenIndexTo uint8, dx *big.Int, minDy *big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopSwap.contract.Transact(opts, "swap", tokenIndexFrom, tokenIndexTo, dx, minDy, deadline)
}

// Swap is a paid mutator transaction binding the contract method 0x91695586.
//
// Solidity: function swap(uint8 tokenIndexFrom, uint8 tokenIndexTo, uint256 dx, uint256 minDy, uint256 deadline) returns(uint256)
func (_HopSwap *HopSwapSession) Swap(tokenIndexFrom uint8, tokenIndexTo uint8, dx *big.Int, minDy *big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopSwap.Contract.Swap(&_HopSwap.TransactOpts, tokenIndexFrom, tokenIndexTo, dx, minDy, deadline)
}

// Swap is a paid mutator transaction binding the contract method 0x91695586.
//
// Solidity: function swap(uint8 tokenIndexFrom, uint8 tokenIndexTo, uint256 dx, uint256 minDy, uint256 deadline) returns(uint256)
func (_HopSwap *HopSwapTransactorSession) Swap(tokenIndexFrom uint8, tokenIndexTo uint8, dx *big.Int, minDy *big.Int, deadline *big.Int) (*types.Transaction, error) {
	return _HopSwap.Contract.Swap(&_HopSwap.TransactOpts, tokenIndexFrom, tokenIndexTo, dx, minDy, deadline)
}

// UpdateUserWithdrawFee is a paid mutator transaction binding the contract method 0xc00c125c.
//
// Solidity: function updateUserWithdrawFee(address recipient, uint256 transferAmount) returns()
func (_HopSwap *HopSwapTransactor) UpdateUserWithdrawFee(opts *bind.TransactOpts, recipient common.Address, transferAmount *big.Int) (*types.Transaction, error) {
	return _HopSwap.contract.Transact(opts, "updateUserWithdrawFee", recipient, transferAmount)
}

// UpdateUserWithdrawFee is a paid mutator transaction binding the contract method 0xc00c125c.
//
// Solidity: function updateUserWithdrawFee(address recipient, uint256 transferAmount) returns()
func (_HopSwap *HopSwapSession) UpdateUserWithdrawFee(recipient common.Address, transferAmount *big.Int) (*types.Transaction, error) {
	return _HopSwap.Contract.UpdateUserWithdrawFee(&_HopSwap.TransactOpts, recipient, transferAmount)
}

// UpdateUserWithdrawFee is a paid mutator transaction binding the contract method 0xc00c125c.
//
// Solidity: function updateUserWithdrawFee(address recipient, uint256 transferAmount) returns()
func (_HopSwap *HopSwapTransactorSession) UpdateUserWithdrawFee(recipient common.Address, transferAmount *big.Int) (*types.Transaction, error) {
	return _HopSwap.Contract.UpdateUserWithdrawFee(&_HopSwap.TransactOpts, recipient, transferAmount)
}

// HopSwapAddLiquidityIterator is returned from FilterAddLiquidity and is used to iterate over the raw logs and unpacked data for AddLiquidity events raised by the HopSwap contract.
type HopSwapAddLiquidityIterator struct {
	Event *HopSwapAddLiquidity // Event containing the contract specifics and raw log

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
func (it *HopSwapAddLiquidityIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopSwapAddLiquidity)
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
		it.Event = new(HopSwapAddLiquidity)
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
func (it *HopSwapAddLiquidityIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopSwapAddLiquidityIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopSwapAddLiquidity represents a AddLiquidity event raised by the HopSwap contract.
type HopSwapAddLiquidity struct {
	Provider      common.Address
	TokenAmounts  []*big.Int
	Fees          []*big.Int
	Invariant     *big.Int
	LpTokenSupply *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAddLiquidity is a free log retrieval operation binding the contract event 0x189c623b666b1b45b83d7178f39b8c087cb09774317ca2f53c2d3c3726f222a2.
//
// Solidity: event AddLiquidity(address indexed provider, uint256[] tokenAmounts, uint256[] fees, uint256 invariant, uint256 lpTokenSupply)
func (_HopSwap *HopSwapFilterer) FilterAddLiquidity(opts *bind.FilterOpts, provider []common.Address) (*HopSwapAddLiquidityIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _HopSwap.contract.FilterLogs(opts, "AddLiquidity", providerRule)
	if err != nil {
		return nil, err
	}
	return &HopSwapAddLiquidityIterator{contract: _HopSwap.contract, event: "AddLiquidity", logs: logs, sub: sub}, nil
}

// WatchAddLiquidity is a free log subscription operation binding the contract event 0x189c623b666b1b45b83d7178f39b8c087cb09774317ca2f53c2d3c3726f222a2.
//
// Solidity: event AddLiquidity(address indexed provider, uint256[] tokenAmounts, uint256[] fees, uint256 invariant, uint256 lpTokenSupply)
func (_HopSwap *HopSwapFilterer) WatchAddLiquidity(opts *bind.WatchOpts, sink chan<- *HopSwapAddLiquidity, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _HopSwap.contract.WatchLogs(opts, "AddLiquidity", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopSwapAddLiquidity)
				if err := _HopSwap.contract.UnpackLog(event, "AddLiquidity", log); err != nil {
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

// ParseAddLiquidity is a log parse operation binding the contract event 0x189c623b666b1b45b83d7178f39b8c087cb09774317ca2f53c2d3c3726f222a2.
//
// Solidity: event AddLiquidity(address indexed provider, uint256[] tokenAmounts, uint256[] fees, uint256 invariant, uint256 lpTokenSupply)
func (_HopSwap *HopSwapFilterer) ParseAddLiquidity(log types.Log) (*HopSwapAddLiquidity, error) {
	event := new(HopSwapAddLiquidity)
	if err := _HopSwap.contract.UnpackLog(event, "AddLiquidity", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopSwapNewAdminFeeIterator is returned from FilterNewAdminFee and is used to iterate over the raw logs and unpacked data for NewAdminFee events raised by the HopSwap contract.
type HopSwapNewAdminFeeIterator struct {
	Event *HopSwapNewAdminFee // Event containing the contract specifics and raw log

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
func (it *HopSwapNewAdminFeeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopSwapNewAdminFee)
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
		it.Event = new(HopSwapNewAdminFee)
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
func (it *HopSwapNewAdminFeeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopSwapNewAdminFeeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopSwapNewAdminFee represents a NewAdminFee event raised by the HopSwap contract.
type HopSwapNewAdminFee struct {
	NewAdminFee *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterNewAdminFee is a free log retrieval operation binding the contract event 0xab599d640ca80cde2b09b128a4154a8dfe608cb80f4c9399c8b954b01fd35f38.
//
// Solidity: event NewAdminFee(uint256 newAdminFee)
func (_HopSwap *HopSwapFilterer) FilterNewAdminFee(opts *bind.FilterOpts) (*HopSwapNewAdminFeeIterator, error) {

	logs, sub, err := _HopSwap.contract.FilterLogs(opts, "NewAdminFee")
	if err != nil {
		return nil, err
	}
	return &HopSwapNewAdminFeeIterator{contract: _HopSwap.contract, event: "NewAdminFee", logs: logs, sub: sub}, nil
}

// WatchNewAdminFee is a free log subscription operation binding the contract event 0xab599d640ca80cde2b09b128a4154a8dfe608cb80f4c9399c8b954b01fd35f38.
//
// Solidity: event NewAdminFee(uint256 newAdminFee)
func (_HopSwap *HopSwapFilterer) WatchNewAdminFee(opts *bind.WatchOpts, sink chan<- *HopSwapNewAdminFee) (event.Subscription, error) {

	logs, sub, err := _HopSwap.contract.WatchLogs(opts, "NewAdminFee")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopSwapNewAdminFee)
				if err := _HopSwap.contract.UnpackLog(event, "NewAdminFee", log); err != nil {
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

// ParseNewAdminFee is a log parse operation binding the contract event 0xab599d640ca80cde2b09b128a4154a8dfe608cb80f4c9399c8b954b01fd35f38.
//
// Solidity: event NewAdminFee(uint256 newAdminFee)
func (_HopSwap *HopSwapFilterer) ParseNewAdminFee(log types.Log) (*HopSwapNewAdminFee, error) {
	event := new(HopSwapNewAdminFee)
	if err := _HopSwap.contract.UnpackLog(event, "NewAdminFee", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopSwapNewSwapFeeIterator is returned from FilterNewSwapFee and is used to iterate over the raw logs and unpacked data for NewSwapFee events raised by the HopSwap contract.
type HopSwapNewSwapFeeIterator struct {
	Event *HopSwapNewSwapFee // Event containing the contract specifics and raw log

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
func (it *HopSwapNewSwapFeeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopSwapNewSwapFee)
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
		it.Event = new(HopSwapNewSwapFee)
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
func (it *HopSwapNewSwapFeeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopSwapNewSwapFeeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopSwapNewSwapFee represents a NewSwapFee event raised by the HopSwap contract.
type HopSwapNewSwapFee struct {
	NewSwapFee *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterNewSwapFee is a free log retrieval operation binding the contract event 0xd88ea5155021c6f8dafa1a741e173f595cdf77ce7c17d43342131d7f06afdfe5.
//
// Solidity: event NewSwapFee(uint256 newSwapFee)
func (_HopSwap *HopSwapFilterer) FilterNewSwapFee(opts *bind.FilterOpts) (*HopSwapNewSwapFeeIterator, error) {

	logs, sub, err := _HopSwap.contract.FilterLogs(opts, "NewSwapFee")
	if err != nil {
		return nil, err
	}
	return &HopSwapNewSwapFeeIterator{contract: _HopSwap.contract, event: "NewSwapFee", logs: logs, sub: sub}, nil
}

// WatchNewSwapFee is a free log subscription operation binding the contract event 0xd88ea5155021c6f8dafa1a741e173f595cdf77ce7c17d43342131d7f06afdfe5.
//
// Solidity: event NewSwapFee(uint256 newSwapFee)
func (_HopSwap *HopSwapFilterer) WatchNewSwapFee(opts *bind.WatchOpts, sink chan<- *HopSwapNewSwapFee) (event.Subscription, error) {

	logs, sub, err := _HopSwap.contract.WatchLogs(opts, "NewSwapFee")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopSwapNewSwapFee)
				if err := _HopSwap.contract.UnpackLog(event, "NewSwapFee", log); err != nil {
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

// ParseNewSwapFee is a log parse operation binding the contract event 0xd88ea5155021c6f8dafa1a741e173f595cdf77ce7c17d43342131d7f06afdfe5.
//
// Solidity: event NewSwapFee(uint256 newSwapFee)
func (_HopSwap *HopSwapFilterer) ParseNewSwapFee(log types.Log) (*HopSwapNewSwapFee, error) {
	event := new(HopSwapNewSwapFee)
	if err := _HopSwap.contract.UnpackLog(event, "NewSwapFee", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopSwapNewWithdrawFeeIterator is returned from FilterNewWithdrawFee and is used to iterate over the raw logs and unpacked data for NewWithdrawFee events raised by the HopSwap contract.
type HopSwapNewWithdrawFeeIterator struct {
	Event *HopSwapNewWithdrawFee // Event containing the contract specifics and raw log

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
func (it *HopSwapNewWithdrawFeeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopSwapNewWithdrawFee)
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
		it.Event = new(HopSwapNewWithdrawFee)
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
func (it *HopSwapNewWithdrawFeeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopSwapNewWithdrawFeeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopSwapNewWithdrawFee represents a NewWithdrawFee event raised by the HopSwap contract.
type HopSwapNewWithdrawFee struct {
	NewWithdrawFee *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterNewWithdrawFee is a free log retrieval operation binding the contract event 0xd5fe46099fa396290a7f57e36c3c3c8774e2562c18ed5d1dcc0fa75071e03f1d.
//
// Solidity: event NewWithdrawFee(uint256 newWithdrawFee)
func (_HopSwap *HopSwapFilterer) FilterNewWithdrawFee(opts *bind.FilterOpts) (*HopSwapNewWithdrawFeeIterator, error) {

	logs, sub, err := _HopSwap.contract.FilterLogs(opts, "NewWithdrawFee")
	if err != nil {
		return nil, err
	}
	return &HopSwapNewWithdrawFeeIterator{contract: _HopSwap.contract, event: "NewWithdrawFee", logs: logs, sub: sub}, nil
}

// WatchNewWithdrawFee is a free log subscription operation binding the contract event 0xd5fe46099fa396290a7f57e36c3c3c8774e2562c18ed5d1dcc0fa75071e03f1d.
//
// Solidity: event NewWithdrawFee(uint256 newWithdrawFee)
func (_HopSwap *HopSwapFilterer) WatchNewWithdrawFee(opts *bind.WatchOpts, sink chan<- *HopSwapNewWithdrawFee) (event.Subscription, error) {

	logs, sub, err := _HopSwap.contract.WatchLogs(opts, "NewWithdrawFee")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopSwapNewWithdrawFee)
				if err := _HopSwap.contract.UnpackLog(event, "NewWithdrawFee", log); err != nil {
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

// ParseNewWithdrawFee is a log parse operation binding the contract event 0xd5fe46099fa396290a7f57e36c3c3c8774e2562c18ed5d1dcc0fa75071e03f1d.
//
// Solidity: event NewWithdrawFee(uint256 newWithdrawFee)
func (_HopSwap *HopSwapFilterer) ParseNewWithdrawFee(log types.Log) (*HopSwapNewWithdrawFee, error) {
	event := new(HopSwapNewWithdrawFee)
	if err := _HopSwap.contract.UnpackLog(event, "NewWithdrawFee", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopSwapRampAIterator is returned from FilterRampA and is used to iterate over the raw logs and unpacked data for RampA events raised by the HopSwap contract.
type HopSwapRampAIterator struct {
	Event *HopSwapRampA // Event containing the contract specifics and raw log

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
func (it *HopSwapRampAIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopSwapRampA)
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
		it.Event = new(HopSwapRampA)
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
func (it *HopSwapRampAIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopSwapRampAIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopSwapRampA represents a RampA event raised by the HopSwap contract.
type HopSwapRampA struct {
	OldA        *big.Int
	NewA        *big.Int
	InitialTime *big.Int
	FutureTime  *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterRampA is a free log retrieval operation binding the contract event 0xa2b71ec6df949300b59aab36b55e189697b750119dd349fcfa8c0f779e83c254.
//
// Solidity: event RampA(uint256 oldA, uint256 newA, uint256 initialTime, uint256 futureTime)
func (_HopSwap *HopSwapFilterer) FilterRampA(opts *bind.FilterOpts) (*HopSwapRampAIterator, error) {

	logs, sub, err := _HopSwap.contract.FilterLogs(opts, "RampA")
	if err != nil {
		return nil, err
	}
	return &HopSwapRampAIterator{contract: _HopSwap.contract, event: "RampA", logs: logs, sub: sub}, nil
}

// WatchRampA is a free log subscription operation binding the contract event 0xa2b71ec6df949300b59aab36b55e189697b750119dd349fcfa8c0f779e83c254.
//
// Solidity: event RampA(uint256 oldA, uint256 newA, uint256 initialTime, uint256 futureTime)
func (_HopSwap *HopSwapFilterer) WatchRampA(opts *bind.WatchOpts, sink chan<- *HopSwapRampA) (event.Subscription, error) {

	logs, sub, err := _HopSwap.contract.WatchLogs(opts, "RampA")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopSwapRampA)
				if err := _HopSwap.contract.UnpackLog(event, "RampA", log); err != nil {
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

// ParseRampA is a log parse operation binding the contract event 0xa2b71ec6df949300b59aab36b55e189697b750119dd349fcfa8c0f779e83c254.
//
// Solidity: event RampA(uint256 oldA, uint256 newA, uint256 initialTime, uint256 futureTime)
func (_HopSwap *HopSwapFilterer) ParseRampA(log types.Log) (*HopSwapRampA, error) {
	event := new(HopSwapRampA)
	if err := _HopSwap.contract.UnpackLog(event, "RampA", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopSwapRemoveLiquidityIterator is returned from FilterRemoveLiquidity and is used to iterate over the raw logs and unpacked data for RemoveLiquidity events raised by the HopSwap contract.
type HopSwapRemoveLiquidityIterator struct {
	Event *HopSwapRemoveLiquidity // Event containing the contract specifics and raw log

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
func (it *HopSwapRemoveLiquidityIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopSwapRemoveLiquidity)
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
		it.Event = new(HopSwapRemoveLiquidity)
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
func (it *HopSwapRemoveLiquidityIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopSwapRemoveLiquidityIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopSwapRemoveLiquidity represents a RemoveLiquidity event raised by the HopSwap contract.
type HopSwapRemoveLiquidity struct {
	Provider      common.Address
	TokenAmounts  []*big.Int
	LpTokenSupply *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterRemoveLiquidity is a free log retrieval operation binding the contract event 0x88d38ed598fdd809c2bf01ee49cd24b7fdabf379a83d29567952b60324d58cef.
//
// Solidity: event RemoveLiquidity(address indexed provider, uint256[] tokenAmounts, uint256 lpTokenSupply)
func (_HopSwap *HopSwapFilterer) FilterRemoveLiquidity(opts *bind.FilterOpts, provider []common.Address) (*HopSwapRemoveLiquidityIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _HopSwap.contract.FilterLogs(opts, "RemoveLiquidity", providerRule)
	if err != nil {
		return nil, err
	}
	return &HopSwapRemoveLiquidityIterator{contract: _HopSwap.contract, event: "RemoveLiquidity", logs: logs, sub: sub}, nil
}

// WatchRemoveLiquidity is a free log subscription operation binding the contract event 0x88d38ed598fdd809c2bf01ee49cd24b7fdabf379a83d29567952b60324d58cef.
//
// Solidity: event RemoveLiquidity(address indexed provider, uint256[] tokenAmounts, uint256 lpTokenSupply)
func (_HopSwap *HopSwapFilterer) WatchRemoveLiquidity(opts *bind.WatchOpts, sink chan<- *HopSwapRemoveLiquidity, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _HopSwap.contract.WatchLogs(opts, "RemoveLiquidity", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopSwapRemoveLiquidity)
				if err := _HopSwap.contract.UnpackLog(event, "RemoveLiquidity", log); err != nil {
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

// ParseRemoveLiquidity is a log parse operation binding the contract event 0x88d38ed598fdd809c2bf01ee49cd24b7fdabf379a83d29567952b60324d58cef.
//
// Solidity: event RemoveLiquidity(address indexed provider, uint256[] tokenAmounts, uint256 lpTokenSupply)
func (_HopSwap *HopSwapFilterer) ParseRemoveLiquidity(log types.Log) (*HopSwapRemoveLiquidity, error) {
	event := new(HopSwapRemoveLiquidity)
	if err := _HopSwap.contract.UnpackLog(event, "RemoveLiquidity", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopSwapRemoveLiquidityImbalanceIterator is returned from FilterRemoveLiquidityImbalance and is used to iterate over the raw logs and unpacked data for RemoveLiquidityImbalance events raised by the HopSwap contract.
type HopSwapRemoveLiquidityImbalanceIterator struct {
	Event *HopSwapRemoveLiquidityImbalance // Event containing the contract specifics and raw log

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
func (it *HopSwapRemoveLiquidityImbalanceIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopSwapRemoveLiquidityImbalance)
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
		it.Event = new(HopSwapRemoveLiquidityImbalance)
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
func (it *HopSwapRemoveLiquidityImbalanceIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopSwapRemoveLiquidityImbalanceIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopSwapRemoveLiquidityImbalance represents a RemoveLiquidityImbalance event raised by the HopSwap contract.
type HopSwapRemoveLiquidityImbalance struct {
	Provider      common.Address
	TokenAmounts  []*big.Int
	Fees          []*big.Int
	Invariant     *big.Int
	LpTokenSupply *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterRemoveLiquidityImbalance is a free log retrieval operation binding the contract event 0x3631c28b1f9dd213e0319fb167b554d76b6c283a41143eb400a0d1adb1af1755.
//
// Solidity: event RemoveLiquidityImbalance(address indexed provider, uint256[] tokenAmounts, uint256[] fees, uint256 invariant, uint256 lpTokenSupply)
func (_HopSwap *HopSwapFilterer) FilterRemoveLiquidityImbalance(opts *bind.FilterOpts, provider []common.Address) (*HopSwapRemoveLiquidityImbalanceIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _HopSwap.contract.FilterLogs(opts, "RemoveLiquidityImbalance", providerRule)
	if err != nil {
		return nil, err
	}
	return &HopSwapRemoveLiquidityImbalanceIterator{contract: _HopSwap.contract, event: "RemoveLiquidityImbalance", logs: logs, sub: sub}, nil
}

// WatchRemoveLiquidityImbalance is a free log subscription operation binding the contract event 0x3631c28b1f9dd213e0319fb167b554d76b6c283a41143eb400a0d1adb1af1755.
//
// Solidity: event RemoveLiquidityImbalance(address indexed provider, uint256[] tokenAmounts, uint256[] fees, uint256 invariant, uint256 lpTokenSupply)
func (_HopSwap *HopSwapFilterer) WatchRemoveLiquidityImbalance(opts *bind.WatchOpts, sink chan<- *HopSwapRemoveLiquidityImbalance, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _HopSwap.contract.WatchLogs(opts, "RemoveLiquidityImbalance", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopSwapRemoveLiquidityImbalance)
				if err := _HopSwap.contract.UnpackLog(event, "RemoveLiquidityImbalance", log); err != nil {
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

// ParseRemoveLiquidityImbalance is a log parse operation binding the contract event 0x3631c28b1f9dd213e0319fb167b554d76b6c283a41143eb400a0d1adb1af1755.
//
// Solidity: event RemoveLiquidityImbalance(address indexed provider, uint256[] tokenAmounts, uint256[] fees, uint256 invariant, uint256 lpTokenSupply)
func (_HopSwap *HopSwapFilterer) ParseRemoveLiquidityImbalance(log types.Log) (*HopSwapRemoveLiquidityImbalance, error) {
	event := new(HopSwapRemoveLiquidityImbalance)
	if err := _HopSwap.contract.UnpackLog(event, "RemoveLiquidityImbalance", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopSwapRemoveLiquidityOneIterator is returned from FilterRemoveLiquidityOne and is used to iterate over the raw logs and unpacked data for RemoveLiquidityOne events raised by the HopSwap contract.
type HopSwapRemoveLiquidityOneIterator struct {
	Event *HopSwapRemoveLiquidityOne // Event containing the contract specifics and raw log

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
func (it *HopSwapRemoveLiquidityOneIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopSwapRemoveLiquidityOne)
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
		it.Event = new(HopSwapRemoveLiquidityOne)
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
func (it *HopSwapRemoveLiquidityOneIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopSwapRemoveLiquidityOneIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopSwapRemoveLiquidityOne represents a RemoveLiquidityOne event raised by the HopSwap contract.
type HopSwapRemoveLiquidityOne struct {
	Provider      common.Address
	LpTokenAmount *big.Int
	LpTokenSupply *big.Int
	BoughtId      *big.Int
	TokensBought  *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterRemoveLiquidityOne is a free log retrieval operation binding the contract event 0x43fb02998f4e03da2e0e6fff53fdbf0c40a9f45f145dc377fc30615d7d7a8a64.
//
// Solidity: event RemoveLiquidityOne(address indexed provider, uint256 lpTokenAmount, uint256 lpTokenSupply, uint256 boughtId, uint256 tokensBought)
func (_HopSwap *HopSwapFilterer) FilterRemoveLiquidityOne(opts *bind.FilterOpts, provider []common.Address) (*HopSwapRemoveLiquidityOneIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _HopSwap.contract.FilterLogs(opts, "RemoveLiquidityOne", providerRule)
	if err != nil {
		return nil, err
	}
	return &HopSwapRemoveLiquidityOneIterator{contract: _HopSwap.contract, event: "RemoveLiquidityOne", logs: logs, sub: sub}, nil
}

// WatchRemoveLiquidityOne is a free log subscription operation binding the contract event 0x43fb02998f4e03da2e0e6fff53fdbf0c40a9f45f145dc377fc30615d7d7a8a64.
//
// Solidity: event RemoveLiquidityOne(address indexed provider, uint256 lpTokenAmount, uint256 lpTokenSupply, uint256 boughtId, uint256 tokensBought)
func (_HopSwap *HopSwapFilterer) WatchRemoveLiquidityOne(opts *bind.WatchOpts, sink chan<- *HopSwapRemoveLiquidityOne, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _HopSwap.contract.WatchLogs(opts, "RemoveLiquidityOne", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopSwapRemoveLiquidityOne)
				if err := _HopSwap.contract.UnpackLog(event, "RemoveLiquidityOne", log); err != nil {
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

// ParseRemoveLiquidityOne is a log parse operation binding the contract event 0x43fb02998f4e03da2e0e6fff53fdbf0c40a9f45f145dc377fc30615d7d7a8a64.
//
// Solidity: event RemoveLiquidityOne(address indexed provider, uint256 lpTokenAmount, uint256 lpTokenSupply, uint256 boughtId, uint256 tokensBought)
func (_HopSwap *HopSwapFilterer) ParseRemoveLiquidityOne(log types.Log) (*HopSwapRemoveLiquidityOne, error) {
	event := new(HopSwapRemoveLiquidityOne)
	if err := _HopSwap.contract.UnpackLog(event, "RemoveLiquidityOne", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopSwapStopRampAIterator is returned from FilterStopRampA and is used to iterate over the raw logs and unpacked data for StopRampA events raised by the HopSwap contract.
type HopSwapStopRampAIterator struct {
	Event *HopSwapStopRampA // Event containing the contract specifics and raw log

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
func (it *HopSwapStopRampAIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopSwapStopRampA)
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
		it.Event = new(HopSwapStopRampA)
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
func (it *HopSwapStopRampAIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopSwapStopRampAIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopSwapStopRampA represents a StopRampA event raised by the HopSwap contract.
type HopSwapStopRampA struct {
	CurrentA *big.Int
	Time     *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterStopRampA is a free log retrieval operation binding the contract event 0x46e22fb3709ad289f62ce63d469248536dbc78d82b84a3d7e74ad606dc201938.
//
// Solidity: event StopRampA(uint256 currentA, uint256 time)
func (_HopSwap *HopSwapFilterer) FilterStopRampA(opts *bind.FilterOpts) (*HopSwapStopRampAIterator, error) {

	logs, sub, err := _HopSwap.contract.FilterLogs(opts, "StopRampA")
	if err != nil {
		return nil, err
	}
	return &HopSwapStopRampAIterator{contract: _HopSwap.contract, event: "StopRampA", logs: logs, sub: sub}, nil
}

// WatchStopRampA is a free log subscription operation binding the contract event 0x46e22fb3709ad289f62ce63d469248536dbc78d82b84a3d7e74ad606dc201938.
//
// Solidity: event StopRampA(uint256 currentA, uint256 time)
func (_HopSwap *HopSwapFilterer) WatchStopRampA(opts *bind.WatchOpts, sink chan<- *HopSwapStopRampA) (event.Subscription, error) {

	logs, sub, err := _HopSwap.contract.WatchLogs(opts, "StopRampA")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopSwapStopRampA)
				if err := _HopSwap.contract.UnpackLog(event, "StopRampA", log); err != nil {
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

// ParseStopRampA is a log parse operation binding the contract event 0x46e22fb3709ad289f62ce63d469248536dbc78d82b84a3d7e74ad606dc201938.
//
// Solidity: event StopRampA(uint256 currentA, uint256 time)
func (_HopSwap *HopSwapFilterer) ParseStopRampA(log types.Log) (*HopSwapStopRampA, error) {
	event := new(HopSwapStopRampA)
	if err := _HopSwap.contract.UnpackLog(event, "StopRampA", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopSwapTokenSwapIterator is returned from FilterTokenSwap and is used to iterate over the raw logs and unpacked data for TokenSwap events raised by the HopSwap contract.
type HopSwapTokenSwapIterator struct {
	Event *HopSwapTokenSwap // Event containing the contract specifics and raw log

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
func (it *HopSwapTokenSwapIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopSwapTokenSwap)
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
		it.Event = new(HopSwapTokenSwap)
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
func (it *HopSwapTokenSwapIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopSwapTokenSwapIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopSwapTokenSwap represents a TokenSwap event raised by the HopSwap contract.
type HopSwapTokenSwap struct {
	Buyer        common.Address
	TokensSold   *big.Int
	TokensBought *big.Int
	SoldId       *big.Int
	BoughtId     *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterTokenSwap is a free log retrieval operation binding the contract event 0xc6c1e0630dbe9130cc068028486c0d118ddcea348550819defd5cb8c257f8a38.
//
// Solidity: event TokenSwap(address indexed buyer, uint256 tokensSold, uint256 tokensBought, uint128 soldId, uint128 boughtId)
func (_HopSwap *HopSwapFilterer) FilterTokenSwap(opts *bind.FilterOpts, buyer []common.Address) (*HopSwapTokenSwapIterator, error) {

	var buyerRule []interface{}
	for _, buyerItem := range buyer {
		buyerRule = append(buyerRule, buyerItem)
	}

	logs, sub, err := _HopSwap.contract.FilterLogs(opts, "TokenSwap", buyerRule)
	if err != nil {
		return nil, err
	}
	return &HopSwapTokenSwapIterator{contract: _HopSwap.contract, event: "TokenSwap", logs: logs, sub: sub}, nil
}

// WatchTokenSwap is a free log subscription operation binding the contract event 0xc6c1e0630dbe9130cc068028486c0d118ddcea348550819defd5cb8c257f8a38.
//
// Solidity: event TokenSwap(address indexed buyer, uint256 tokensSold, uint256 tokensBought, uint128 soldId, uint128 boughtId)
func (_HopSwap *HopSwapFilterer) WatchTokenSwap(opts *bind.WatchOpts, sink chan<- *HopSwapTokenSwap, buyer []common.Address) (event.Subscription, error) {

	var buyerRule []interface{}
	for _, buyerItem := range buyer {
		buyerRule = append(buyerRule, buyerItem)
	}

	logs, sub, err := _HopSwap.contract.WatchLogs(opts, "TokenSwap", buyerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopSwapTokenSwap)
				if err := _HopSwap.contract.UnpackLog(event, "TokenSwap", log); err != nil {
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

// ParseTokenSwap is a log parse operation binding the contract event 0xc6c1e0630dbe9130cc068028486c0d118ddcea348550819defd5cb8c257f8a38.
//
// Solidity: event TokenSwap(address indexed buyer, uint256 tokensSold, uint256 tokensBought, uint128 soldId, uint128 boughtId)
func (_HopSwap *HopSwapFilterer) ParseTokenSwap(log types.Log) (*HopSwapTokenSwap, error) {
	event := new(HopSwapTokenSwap)
	if err := _HopSwap.contract.UnpackLog(event, "TokenSwap", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
