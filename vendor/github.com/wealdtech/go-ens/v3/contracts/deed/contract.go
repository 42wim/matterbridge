// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package deed

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
const ContractABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"creationDate\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"destroyDeed\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"setOwner\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"registrar\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"value\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"previousOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newValue\",\"type\":\"uint256\"},{\"name\":\"throwOnFailure\",\"type\":\"bool\"}],\"name\":\"setBalance\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"refundRatio\",\"type\":\"uint256\"}],\"name\":\"closeDeed\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newRegistrar\",\"type\":\"address\"}],\"name\":\"setRegistrar\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"payable\":true,\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnerChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"DeedClosed\",\"type\":\"event\"}]"

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

// CreationDate is a free data retrieval call binding the contract method 0x05b34410.
//
// Solidity: function creationDate() returns(uint256)
func (_Contract *ContractCaller) CreationDate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "creationDate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CreationDate is a free data retrieval call binding the contract method 0x05b34410.
//
// Solidity: function creationDate() returns(uint256)
func (_Contract *ContractSession) CreationDate() (*big.Int, error) {
	return _Contract.Contract.CreationDate(&_Contract.CallOpts)
}

// CreationDate is a free data retrieval call binding the contract method 0x05b34410.
//
// Solidity: function creationDate() returns(uint256)
func (_Contract *ContractCallerSession) CreationDate() (*big.Int, error) {
	return _Contract.Contract.CreationDate(&_Contract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() returns(address)
func (_Contract *ContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() returns(address)
func (_Contract *ContractSession) Owner() (common.Address, error) {
	return _Contract.Contract.Owner(&_Contract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() returns(address)
func (_Contract *ContractCallerSession) Owner() (common.Address, error) {
	return _Contract.Contract.Owner(&_Contract.CallOpts)
}

// PreviousOwner is a free data retrieval call binding the contract method 0x674f220f.
//
// Solidity: function previousOwner() returns(address)
func (_Contract *ContractCaller) PreviousOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "previousOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PreviousOwner is a free data retrieval call binding the contract method 0x674f220f.
//
// Solidity: function previousOwner() returns(address)
func (_Contract *ContractSession) PreviousOwner() (common.Address, error) {
	return _Contract.Contract.PreviousOwner(&_Contract.CallOpts)
}

// PreviousOwner is a free data retrieval call binding the contract method 0x674f220f.
//
// Solidity: function previousOwner() returns(address)
func (_Contract *ContractCallerSession) PreviousOwner() (common.Address, error) {
	return _Contract.Contract.PreviousOwner(&_Contract.CallOpts)
}

// Registrar is a free data retrieval call binding the contract method 0x2b20e397.
//
// Solidity: function registrar() returns(address)
func (_Contract *ContractCaller) Registrar(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "registrar")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Registrar is a free data retrieval call binding the contract method 0x2b20e397.
//
// Solidity: function registrar() returns(address)
func (_Contract *ContractSession) Registrar() (common.Address, error) {
	return _Contract.Contract.Registrar(&_Contract.CallOpts)
}

// Registrar is a free data retrieval call binding the contract method 0x2b20e397.
//
// Solidity: function registrar() returns(address)
func (_Contract *ContractCallerSession) Registrar() (common.Address, error) {
	return _Contract.Contract.Registrar(&_Contract.CallOpts)
}

// Value is a free data retrieval call binding the contract method 0x3fa4f245.
//
// Solidity: function value() returns(uint256)
func (_Contract *ContractCaller) Value(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "value")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Value is a free data retrieval call binding the contract method 0x3fa4f245.
//
// Solidity: function value() returns(uint256)
func (_Contract *ContractSession) Value() (*big.Int, error) {
	return _Contract.Contract.Value(&_Contract.CallOpts)
}

// Value is a free data retrieval call binding the contract method 0x3fa4f245.
//
// Solidity: function value() returns(uint256)
func (_Contract *ContractCallerSession) Value() (*big.Int, error) {
	return _Contract.Contract.Value(&_Contract.CallOpts)
}

// CloseDeed is a paid mutator transaction binding the contract method 0xbbe42771.
//
// Solidity: function closeDeed(uint256 refundRatio) returns()
func (_Contract *ContractTransactor) CloseDeed(opts *bind.TransactOpts, refundRatio *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "closeDeed", refundRatio)
}

// CloseDeed is a paid mutator transaction binding the contract method 0xbbe42771.
//
// Solidity: function closeDeed(uint256 refundRatio) returns()
func (_Contract *ContractSession) CloseDeed(refundRatio *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.CloseDeed(&_Contract.TransactOpts, refundRatio)
}

// CloseDeed is a paid mutator transaction binding the contract method 0xbbe42771.
//
// Solidity: function closeDeed(uint256 refundRatio) returns()
func (_Contract *ContractTransactorSession) CloseDeed(refundRatio *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.CloseDeed(&_Contract.TransactOpts, refundRatio)
}

// DestroyDeed is a paid mutator transaction binding the contract method 0x0b5ab3d5.
//
// Solidity: function destroyDeed() returns()
func (_Contract *ContractTransactor) DestroyDeed(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "destroyDeed")
}

// DestroyDeed is a paid mutator transaction binding the contract method 0x0b5ab3d5.
//
// Solidity: function destroyDeed() returns()
func (_Contract *ContractSession) DestroyDeed() (*types.Transaction, error) {
	return _Contract.Contract.DestroyDeed(&_Contract.TransactOpts)
}

// DestroyDeed is a paid mutator transaction binding the contract method 0x0b5ab3d5.
//
// Solidity: function destroyDeed() returns()
func (_Contract *ContractTransactorSession) DestroyDeed() (*types.Transaction, error) {
	return _Contract.Contract.DestroyDeed(&_Contract.TransactOpts)
}

// SetBalance is a paid mutator transaction binding the contract method 0xb0c80972.
//
// Solidity: function setBalance(uint256 newValue, bool throwOnFailure) returns()
func (_Contract *ContractTransactor) SetBalance(opts *bind.TransactOpts, newValue *big.Int, throwOnFailure bool) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setBalance", newValue, throwOnFailure)
}

// SetBalance is a paid mutator transaction binding the contract method 0xb0c80972.
//
// Solidity: function setBalance(uint256 newValue, bool throwOnFailure) returns()
func (_Contract *ContractSession) SetBalance(newValue *big.Int, throwOnFailure bool) (*types.Transaction, error) {
	return _Contract.Contract.SetBalance(&_Contract.TransactOpts, newValue, throwOnFailure)
}

// SetBalance is a paid mutator transaction binding the contract method 0xb0c80972.
//
// Solidity: function setBalance(uint256 newValue, bool throwOnFailure) returns()
func (_Contract *ContractTransactorSession) SetBalance(newValue *big.Int, throwOnFailure bool) (*types.Transaction, error) {
	return _Contract.Contract.SetBalance(&_Contract.TransactOpts, newValue, throwOnFailure)
}

// SetOwner is a paid mutator transaction binding the contract method 0x13af4035.
//
// Solidity: function setOwner(address newOwner) returns()
func (_Contract *ContractTransactor) SetOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setOwner", newOwner)
}

// SetOwner is a paid mutator transaction binding the contract method 0x13af4035.
//
// Solidity: function setOwner(address newOwner) returns()
func (_Contract *ContractSession) SetOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.SetOwner(&_Contract.TransactOpts, newOwner)
}

// SetOwner is a paid mutator transaction binding the contract method 0x13af4035.
//
// Solidity: function setOwner(address newOwner) returns()
func (_Contract *ContractTransactorSession) SetOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.SetOwner(&_Contract.TransactOpts, newOwner)
}

// SetRegistrar is a paid mutator transaction binding the contract method 0xfaab9d39.
//
// Solidity: function setRegistrar(address newRegistrar) returns()
func (_Contract *ContractTransactor) SetRegistrar(opts *bind.TransactOpts, newRegistrar common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setRegistrar", newRegistrar)
}

// SetRegistrar is a paid mutator transaction binding the contract method 0xfaab9d39.
//
// Solidity: function setRegistrar(address newRegistrar) returns()
func (_Contract *ContractSession) SetRegistrar(newRegistrar common.Address) (*types.Transaction, error) {
	return _Contract.Contract.SetRegistrar(&_Contract.TransactOpts, newRegistrar)
}

// SetRegistrar is a paid mutator transaction binding the contract method 0xfaab9d39.
//
// Solidity: function setRegistrar(address newRegistrar) returns()
func (_Contract *ContractTransactorSession) SetRegistrar(newRegistrar common.Address) (*types.Transaction, error) {
	return _Contract.Contract.SetRegistrar(&_Contract.TransactOpts, newRegistrar)
}

// ContractDeedClosedIterator is returned from FilterDeedClosed and is used to iterate over the raw logs and unpacked data for DeedClosed events raised by the Contract contract.
type ContractDeedClosedIterator struct {
	Event *ContractDeedClosed // Event containing the contract specifics and raw log

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
func (it *ContractDeedClosedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDeedClosed)
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
		it.Event = new(ContractDeedClosed)
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
func (it *ContractDeedClosedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDeedClosedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDeedClosed represents a DeedClosed event raised by the Contract contract.
type ContractDeedClosed struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterDeedClosed is a free log retrieval operation binding the contract event 0xbb2ce2f51803bba16bc85282b47deeea9a5c6223eabea1077be696b3f265cf13.
//
// Solidity: event DeedClosed()
func (_Contract *ContractFilterer) FilterDeedClosed(opts *bind.FilterOpts) (*ContractDeedClosedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "DeedClosed")
	if err != nil {
		return nil, err
	}
	return &ContractDeedClosedIterator{contract: _Contract.contract, event: "DeedClosed", logs: logs, sub: sub}, nil
}

// WatchDeedClosed is a free log subscription operation binding the contract event 0xbb2ce2f51803bba16bc85282b47deeea9a5c6223eabea1077be696b3f265cf13.
//
// Solidity: event DeedClosed()
func (_Contract *ContractFilterer) WatchDeedClosed(opts *bind.WatchOpts, sink chan<- *ContractDeedClosed) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "DeedClosed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDeedClosed)
				if err := _Contract.contract.UnpackLog(event, "DeedClosed", log); err != nil {
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

// ParseDeedClosed is a log parse operation binding the contract event 0xbb2ce2f51803bba16bc85282b47deeea9a5c6223eabea1077be696b3f265cf13.
//
// Solidity: event DeedClosed()
func (_Contract *ContractFilterer) ParseDeedClosed(log types.Log) (*ContractDeedClosed, error) {
	event := new(ContractDeedClosed)
	if err := _Contract.contract.UnpackLog(event, "DeedClosed", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractOwnerChangedIterator is returned from FilterOwnerChanged and is used to iterate over the raw logs and unpacked data for OwnerChanged events raised by the Contract contract.
type ContractOwnerChangedIterator struct {
	Event *ContractOwnerChanged // Event containing the contract specifics and raw log

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
func (it *ContractOwnerChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractOwnerChanged)
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
		it.Event = new(ContractOwnerChanged)
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
func (it *ContractOwnerChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractOwnerChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractOwnerChanged represents a OwnerChanged event raised by the Contract contract.
type ContractOwnerChanged struct {
	NewOwner common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterOwnerChanged is a free log retrieval operation binding the contract event 0xa2ea9883a321a3e97b8266c2b078bfeec6d50c711ed71f874a90d500ae2eaf36.
//
// Solidity: event OwnerChanged(address newOwner)
func (_Contract *ContractFilterer) FilterOwnerChanged(opts *bind.FilterOpts) (*ContractOwnerChangedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "OwnerChanged")
	if err != nil {
		return nil, err
	}
	return &ContractOwnerChangedIterator{contract: _Contract.contract, event: "OwnerChanged", logs: logs, sub: sub}, nil
}

// WatchOwnerChanged is a free log subscription operation binding the contract event 0xa2ea9883a321a3e97b8266c2b078bfeec6d50c711ed71f874a90d500ae2eaf36.
//
// Solidity: event OwnerChanged(address newOwner)
func (_Contract *ContractFilterer) WatchOwnerChanged(opts *bind.WatchOpts, sink chan<- *ContractOwnerChanged) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "OwnerChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractOwnerChanged)
				if err := _Contract.contract.UnpackLog(event, "OwnerChanged", log); err != nil {
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

// ParseOwnerChanged is a log parse operation binding the contract event 0xa2ea9883a321a3e97b8266c2b078bfeec6d50c711ed71f874a90d500ae2eaf36.
//
// Solidity: event OwnerChanged(address newOwner)
func (_Contract *ContractFilterer) ParseOwnerChanged(log types.Log) (*ContractOwnerChanged, error) {
	event := new(ContractOwnerChanged)
	if err := _Contract.contract.UnpackLog(event, "OwnerChanged", log); err != nil {
		return nil, err
	}
	return event, nil
}
