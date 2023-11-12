// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package dnssecoracle

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
const ContractABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"dnstype\",\"type\":\"uint16\"},{\"name\":\"name\",\"type\":\"bytes\"}],\"name\":\"rrdata\",\"outputs\":[{\"name\":\"\",\"type\":\"uint32\"},{\"name\":\"\",\"type\":\"uint64\"},{\"name\":\"\",\"type\":\"bytes20\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"input\",\"type\":\"bytes\"},{\"name\":\"sig\",\"type\":\"bytes\"},{\"name\":\"proof\",\"type\":\"bytes\"}],\"name\":\"submitRRSet\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"data\",\"type\":\"bytes\"},{\"name\":\"proof\",\"type\":\"bytes\"}],\"name\":\"submitRRSets\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"deleteType\",\"type\":\"uint16\"},{\"name\":\"deleteName\",\"type\":\"bytes\"},{\"name\":\"nsec\",\"type\":\"bytes\"},{\"name\":\"sig\",\"type\":\"bytes\"},{\"name\":\"proof\",\"type\":\"bytes\"}],\"name\":\"deleteRRSet\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"uint8\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"AlgorithmUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"uint8\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"DigestUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"uint8\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"NSEC3DigestUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"name\",\"type\":\"bytes\"},{\"indexed\":false,\"name\":\"rrset\",\"type\":\"bytes\"}],\"name\":\"RRSetUpdated\",\"type\":\"event\"}]"

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

// Rrdata is a free data retrieval call binding the contract method 0x087991bc.
//
// Solidity: function rrdata(uint16 dnstype, bytes name) view returns(uint32, uint64, bytes20)
func (_Contract *ContractCaller) Rrdata(opts *bind.CallOpts, dnstype uint16, name []byte) (uint32, uint64, [20]byte, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "rrdata", dnstype, name)

	if err != nil {
		return *new(uint32), *new(uint64), *new([20]byte), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)
	out1 := *abi.ConvertType(out[1], new(uint64)).(*uint64)
	out2 := *abi.ConvertType(out[2], new([20]byte)).(*[20]byte)

	return out0, out1, out2, err

}

// Rrdata is a free data retrieval call binding the contract method 0x087991bc.
//
// Solidity: function rrdata(uint16 dnstype, bytes name) view returns(uint32, uint64, bytes20)
func (_Contract *ContractSession) Rrdata(dnstype uint16, name []byte) (uint32, uint64, [20]byte, error) {
	return _Contract.Contract.Rrdata(&_Contract.CallOpts, dnstype, name)
}

// Rrdata is a free data retrieval call binding the contract method 0x087991bc.
//
// Solidity: function rrdata(uint16 dnstype, bytes name) view returns(uint32, uint64, bytes20)
func (_Contract *ContractCallerSession) Rrdata(dnstype uint16, name []byte) (uint32, uint64, [20]byte, error) {
	return _Contract.Contract.Rrdata(&_Contract.CallOpts, dnstype, name)
}

// DeleteRRSet is a paid mutator transaction binding the contract method 0xe60b202f.
//
// Solidity: function deleteRRSet(uint16 deleteType, bytes deleteName, bytes nsec, bytes sig, bytes proof) returns()
func (_Contract *ContractTransactor) DeleteRRSet(opts *bind.TransactOpts, deleteType uint16, deleteName []byte, nsec []byte, sig []byte, proof []byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "deleteRRSet", deleteType, deleteName, nsec, sig, proof)
}

// DeleteRRSet is a paid mutator transaction binding the contract method 0xe60b202f.
//
// Solidity: function deleteRRSet(uint16 deleteType, bytes deleteName, bytes nsec, bytes sig, bytes proof) returns()
func (_Contract *ContractSession) DeleteRRSet(deleteType uint16, deleteName []byte, nsec []byte, sig []byte, proof []byte) (*types.Transaction, error) {
	return _Contract.Contract.DeleteRRSet(&_Contract.TransactOpts, deleteType, deleteName, nsec, sig, proof)
}

// DeleteRRSet is a paid mutator transaction binding the contract method 0xe60b202f.
//
// Solidity: function deleteRRSet(uint16 deleteType, bytes deleteName, bytes nsec, bytes sig, bytes proof) returns()
func (_Contract *ContractTransactorSession) DeleteRRSet(deleteType uint16, deleteName []byte, nsec []byte, sig []byte, proof []byte) (*types.Transaction, error) {
	return _Contract.Contract.DeleteRRSet(&_Contract.TransactOpts, deleteType, deleteName, nsec, sig, proof)
}

// SubmitRRSet is a paid mutator transaction binding the contract method 0x4d46d581.
//
// Solidity: function submitRRSet(bytes input, bytes sig, bytes proof) returns(bytes)
func (_Contract *ContractTransactor) SubmitRRSet(opts *bind.TransactOpts, input []byte, sig []byte, proof []byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "submitRRSet", input, sig, proof)
}

// SubmitRRSet is a paid mutator transaction binding the contract method 0x4d46d581.
//
// Solidity: function submitRRSet(bytes input, bytes sig, bytes proof) returns(bytes)
func (_Contract *ContractSession) SubmitRRSet(input []byte, sig []byte, proof []byte) (*types.Transaction, error) {
	return _Contract.Contract.SubmitRRSet(&_Contract.TransactOpts, input, sig, proof)
}

// SubmitRRSet is a paid mutator transaction binding the contract method 0x4d46d581.
//
// Solidity: function submitRRSet(bytes input, bytes sig, bytes proof) returns(bytes)
func (_Contract *ContractTransactorSession) SubmitRRSet(input []byte, sig []byte, proof []byte) (*types.Transaction, error) {
	return _Contract.Contract.SubmitRRSet(&_Contract.TransactOpts, input, sig, proof)
}

// SubmitRRSets is a paid mutator transaction binding the contract method 0x76a14d1d.
//
// Solidity: function submitRRSets(bytes data, bytes proof) returns(bytes)
func (_Contract *ContractTransactor) SubmitRRSets(opts *bind.TransactOpts, data []byte, proof []byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "submitRRSets", data, proof)
}

// SubmitRRSets is a paid mutator transaction binding the contract method 0x76a14d1d.
//
// Solidity: function submitRRSets(bytes data, bytes proof) returns(bytes)
func (_Contract *ContractSession) SubmitRRSets(data []byte, proof []byte) (*types.Transaction, error) {
	return _Contract.Contract.SubmitRRSets(&_Contract.TransactOpts, data, proof)
}

// SubmitRRSets is a paid mutator transaction binding the contract method 0x76a14d1d.
//
// Solidity: function submitRRSets(bytes data, bytes proof) returns(bytes)
func (_Contract *ContractTransactorSession) SubmitRRSets(data []byte, proof []byte) (*types.Transaction, error) {
	return _Contract.Contract.SubmitRRSets(&_Contract.TransactOpts, data, proof)
}

// ContractAlgorithmUpdatedIterator is returned from FilterAlgorithmUpdated and is used to iterate over the raw logs and unpacked data for AlgorithmUpdated events raised by the Contract contract.
type ContractAlgorithmUpdatedIterator struct {
	Event *ContractAlgorithmUpdated // Event containing the contract specifics and raw log

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
func (it *ContractAlgorithmUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractAlgorithmUpdated)
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
		it.Event = new(ContractAlgorithmUpdated)
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
func (it *ContractAlgorithmUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractAlgorithmUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractAlgorithmUpdated represents a AlgorithmUpdated event raised by the Contract contract.
type ContractAlgorithmUpdated struct {
	Id   uint8
	Addr common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlgorithmUpdated is a free log retrieval operation binding the contract event 0xf73c3c226af96b7f1ba666a21b3ceaf2be3ee6a365e3178fd9cd1eaae0075aa8.
//
// Solidity: event AlgorithmUpdated(uint8 id, address addr)
func (_Contract *ContractFilterer) FilterAlgorithmUpdated(opts *bind.FilterOpts) (*ContractAlgorithmUpdatedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "AlgorithmUpdated")
	if err != nil {
		return nil, err
	}
	return &ContractAlgorithmUpdatedIterator{contract: _Contract.contract, event: "AlgorithmUpdated", logs: logs, sub: sub}, nil
}

// WatchAlgorithmUpdated is a free log subscription operation binding the contract event 0xf73c3c226af96b7f1ba666a21b3ceaf2be3ee6a365e3178fd9cd1eaae0075aa8.
//
// Solidity: event AlgorithmUpdated(uint8 id, address addr)
func (_Contract *ContractFilterer) WatchAlgorithmUpdated(opts *bind.WatchOpts, sink chan<- *ContractAlgorithmUpdated) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "AlgorithmUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractAlgorithmUpdated)
				if err := _Contract.contract.UnpackLog(event, "AlgorithmUpdated", log); err != nil {
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

// ParseAlgorithmUpdated is a log parse operation binding the contract event 0xf73c3c226af96b7f1ba666a21b3ceaf2be3ee6a365e3178fd9cd1eaae0075aa8.
//
// Solidity: event AlgorithmUpdated(uint8 id, address addr)
func (_Contract *ContractFilterer) ParseAlgorithmUpdated(log types.Log) (*ContractAlgorithmUpdated, error) {
	event := new(ContractAlgorithmUpdated)
	if err := _Contract.contract.UnpackLog(event, "AlgorithmUpdated", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractDigestUpdatedIterator is returned from FilterDigestUpdated and is used to iterate over the raw logs and unpacked data for DigestUpdated events raised by the Contract contract.
type ContractDigestUpdatedIterator struct {
	Event *ContractDigestUpdated // Event containing the contract specifics and raw log

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
func (it *ContractDigestUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDigestUpdated)
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
		it.Event = new(ContractDigestUpdated)
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
func (it *ContractDigestUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDigestUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDigestUpdated represents a DigestUpdated event raised by the Contract contract.
type ContractDigestUpdated struct {
	Id   uint8
	Addr common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterDigestUpdated is a free log retrieval operation binding the contract event 0x2fcc274c3b72dd483ab201bfa87295e3817e8b9b10693219873b722ca1af00c7.
//
// Solidity: event DigestUpdated(uint8 id, address addr)
func (_Contract *ContractFilterer) FilterDigestUpdated(opts *bind.FilterOpts) (*ContractDigestUpdatedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "DigestUpdated")
	if err != nil {
		return nil, err
	}
	return &ContractDigestUpdatedIterator{contract: _Contract.contract, event: "DigestUpdated", logs: logs, sub: sub}, nil
}

// WatchDigestUpdated is a free log subscription operation binding the contract event 0x2fcc274c3b72dd483ab201bfa87295e3817e8b9b10693219873b722ca1af00c7.
//
// Solidity: event DigestUpdated(uint8 id, address addr)
func (_Contract *ContractFilterer) WatchDigestUpdated(opts *bind.WatchOpts, sink chan<- *ContractDigestUpdated) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "DigestUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDigestUpdated)
				if err := _Contract.contract.UnpackLog(event, "DigestUpdated", log); err != nil {
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

// ParseDigestUpdated is a log parse operation binding the contract event 0x2fcc274c3b72dd483ab201bfa87295e3817e8b9b10693219873b722ca1af00c7.
//
// Solidity: event DigestUpdated(uint8 id, address addr)
func (_Contract *ContractFilterer) ParseDigestUpdated(log types.Log) (*ContractDigestUpdated, error) {
	event := new(ContractDigestUpdated)
	if err := _Contract.contract.UnpackLog(event, "DigestUpdated", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractNSEC3DigestUpdatedIterator is returned from FilterNSEC3DigestUpdated and is used to iterate over the raw logs and unpacked data for NSEC3DigestUpdated events raised by the Contract contract.
type ContractNSEC3DigestUpdatedIterator struct {
	Event *ContractNSEC3DigestUpdated // Event containing the contract specifics and raw log

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
func (it *ContractNSEC3DigestUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractNSEC3DigestUpdated)
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
		it.Event = new(ContractNSEC3DigestUpdated)
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
func (it *ContractNSEC3DigestUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractNSEC3DigestUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractNSEC3DigestUpdated represents a NSEC3DigestUpdated event raised by the Contract contract.
type ContractNSEC3DigestUpdated struct {
	Id   uint8
	Addr common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterNSEC3DigestUpdated is a free log retrieval operation binding the contract event 0xc7eec866a7a1386188cc3ca20ffea75b71bd3e90a60b6791b1d3f0971145118d.
//
// Solidity: event NSEC3DigestUpdated(uint8 id, address addr)
func (_Contract *ContractFilterer) FilterNSEC3DigestUpdated(opts *bind.FilterOpts) (*ContractNSEC3DigestUpdatedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "NSEC3DigestUpdated")
	if err != nil {
		return nil, err
	}
	return &ContractNSEC3DigestUpdatedIterator{contract: _Contract.contract, event: "NSEC3DigestUpdated", logs: logs, sub: sub}, nil
}

// WatchNSEC3DigestUpdated is a free log subscription operation binding the contract event 0xc7eec866a7a1386188cc3ca20ffea75b71bd3e90a60b6791b1d3f0971145118d.
//
// Solidity: event NSEC3DigestUpdated(uint8 id, address addr)
func (_Contract *ContractFilterer) WatchNSEC3DigestUpdated(opts *bind.WatchOpts, sink chan<- *ContractNSEC3DigestUpdated) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "NSEC3DigestUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractNSEC3DigestUpdated)
				if err := _Contract.contract.UnpackLog(event, "NSEC3DigestUpdated", log); err != nil {
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

// ParseNSEC3DigestUpdated is a log parse operation binding the contract event 0xc7eec866a7a1386188cc3ca20ffea75b71bd3e90a60b6791b1d3f0971145118d.
//
// Solidity: event NSEC3DigestUpdated(uint8 id, address addr)
func (_Contract *ContractFilterer) ParseNSEC3DigestUpdated(log types.Log) (*ContractNSEC3DigestUpdated, error) {
	event := new(ContractNSEC3DigestUpdated)
	if err := _Contract.contract.UnpackLog(event, "NSEC3DigestUpdated", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractRRSetUpdatedIterator is returned from FilterRRSetUpdated and is used to iterate over the raw logs and unpacked data for RRSetUpdated events raised by the Contract contract.
type ContractRRSetUpdatedIterator struct {
	Event *ContractRRSetUpdated // Event containing the contract specifics and raw log

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
func (it *ContractRRSetUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractRRSetUpdated)
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
		it.Event = new(ContractRRSetUpdated)
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
func (it *ContractRRSetUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractRRSetUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractRRSetUpdated represents a RRSetUpdated event raised by the Contract contract.
type ContractRRSetUpdated struct {
	Name  []byte
	Rrset []byte
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterRRSetUpdated is a free log retrieval operation binding the contract event 0x55ced933cdd5a34dd03eb5d4bef19ec6ebb251dcd7a988eee0c1b9a13baaa88b.
//
// Solidity: event RRSetUpdated(bytes name, bytes rrset)
func (_Contract *ContractFilterer) FilterRRSetUpdated(opts *bind.FilterOpts) (*ContractRRSetUpdatedIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "RRSetUpdated")
	if err != nil {
		return nil, err
	}
	return &ContractRRSetUpdatedIterator{contract: _Contract.contract, event: "RRSetUpdated", logs: logs, sub: sub}, nil
}

// WatchRRSetUpdated is a free log subscription operation binding the contract event 0x55ced933cdd5a34dd03eb5d4bef19ec6ebb251dcd7a988eee0c1b9a13baaa88b.
//
// Solidity: event RRSetUpdated(bytes name, bytes rrset)
func (_Contract *ContractFilterer) WatchRRSetUpdated(opts *bind.WatchOpts, sink chan<- *ContractRRSetUpdated) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "RRSetUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractRRSetUpdated)
				if err := _Contract.contract.UnpackLog(event, "RRSetUpdated", log); err != nil {
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

// ParseRRSetUpdated is a log parse operation binding the contract event 0x55ced933cdd5a34dd03eb5d4bef19ec6ebb251dcd7a988eee0c1b9a13baaa88b.
//
// Solidity: event RRSetUpdated(bytes name, bytes rrset)
func (_Contract *ContractFilterer) ParseRRSetUpdated(log types.Log) (*ContractRRSetUpdated, error) {
	event := new(ContractRRSetUpdated)
	if err := _Contract.contract.UnpackLog(event, "RRSetUpdated", log); err != nil {
		return nil, err
	}
	return event, nil
}
