// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

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

// RLNMetaData contains all meta data concerning the RLN contract.
var RLNMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_poseidonHasher\",\"type\":\"address\"},{\"internalType\":\"uint16\",\"name\":\"_contractIndex\",\"type\":\"uint16\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"DuplicateIdCommitment\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FullTree\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"idCommitment\",\"type\":\"uint256\"}],\"name\":\"InvalidIdCommitment\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotImplemented\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"idCommitment\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"MemberRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"idCommitment\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"MemberWithdrawn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"DEPTH\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MEMBERSHIP_DEPOSIT\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SET_SIZE\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"contractIndex\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"deployedBlockNumber\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"idCommitmentIndex\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"idCommitment\",\"type\":\"uint256\"}],\"name\":\"isValidCommitment\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"memberExists\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"members\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"poseidonHasher\",\"outputs\":[{\"internalType\":\"contractPoseidonHasher\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"idCommitments\",\"type\":\"uint256[]\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"idCommitment\",\"type\":\"uint256\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"idCommitment\",\"type\":\"uint256\"},{\"internalType\":\"addresspayable\",\"name\":\"receiver\",\"type\":\"address\"},{\"internalType\":\"uint256[8]\",\"name\":\"proof\",\"type\":\"uint256[8]\"}],\"name\":\"slash\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"stakedAmounts\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"verifier\",\"outputs\":[{\"internalType\":\"contractIVerifier\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"withdrawalBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x61016060405260006001553480156200001757600080fd5b50604051620014fa380380620014fa83398181016040528101906200003d919062000290565b6000601483600062000064620000586200011b60201b60201c565b6200012360201b60201c565b83608081815250508260a08181525050826001901b60c081815250508173ffffffffffffffffffffffffffffffffffffffff1660e08173ffffffffffffffffffffffffffffffffffffffff16815250508073ffffffffffffffffffffffffffffffffffffffff166101008173ffffffffffffffffffffffffffffffffffffffff16815250504363ffffffff166101208163ffffffff1681525050505050508061ffff166101408161ffff16815250505050620002d7565b600033905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050816000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a35050565b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006200021982620001ec565b9050919050565b6200022b816200020c565b81146200023757600080fd5b50565b6000815190506200024b8162000220565b92915050565b600061ffff82169050919050565b6200026a8162000251565b81146200027657600080fd5b50565b6000815190506200028a816200025f565b92915050565b60008060408385031215620002aa57620002a9620001e7565b5b6000620002ba858286016200023a565b9250506020620002cd8582860162000279565b9150509250929050565b60805160a05160c05160e0516101005161012051610140516111ba620003406000396000610545015260006105e3015260006105690152600081816104ac015261058d0152600081816107560152610aea015260006106fc015260006107ac01526111ba6000f3fe60806040526004361061011f5760003560e01c80638be9b119116100a0578063c5b208ff11610064578063c5b208ff146103c5578063d0383d6814610402578063f207564e1461042d578063f220b9ec14610449578063f2fde38b146104745761011f565b80638be9b119146102de5780638da5cb5b1461030757806398366e3514610332578063ae74552a1461035d578063bc499128146103885761011f565b80634add651e116100e75780634add651e146101f95780635daf08ca146102245780636bdcc8ab14610261578063715018a61461029e5780637a34289d146102b55761011f565b806322d9730c1461012457806328b070e0146101615780632b7ac3f31461018c578063331b6ab3146101b75780633ccfd60b146101e2575b600080fd5b34801561013057600080fd5b5061014b60048036038101906101469190610b86565b61049d565b6040516101589190610bce565b60405180910390f35b34801561016d57600080fd5b50610176610543565b6040516101839190610c06565b60405180910390f35b34801561019857600080fd5b506101a1610567565b6040516101ae9190610ca0565b60405180910390f35b3480156101c357600080fd5b506101cc61058b565b6040516101d99190610cdc565b60405180910390f35b3480156101ee57600080fd5b506101f76105af565b005b34801561020557600080fd5b5061020e6105e1565b60405161021b9190610d16565b60405180910390f35b34801561023057600080fd5b5061024b60048036038101906102469190610b86565b610605565b6040516102589190610d40565b60405180910390f35b34801561026d57600080fd5b5061028860048036038101906102839190610b86565b61061d565b6040516102959190610bce565b60405180910390f35b3480156102aa57600080fd5b506102b361063d565b005b3480156102c157600080fd5b506102dc60048036038101906102d79190610dc0565b610651565b005b3480156102ea57600080fd5b5061030560048036038101906103009190610e6d565b61069f565b005b34801561031357600080fd5b5061031c6106d1565b6040516103299190610ee2565b60405180910390f35b34801561033e57600080fd5b506103476106fa565b6040516103549190610d40565b60405180910390f35b34801561036957600080fd5b5061037261071e565b60405161037f9190610d40565b60405180910390f35b34801561039457600080fd5b506103af60048036038101906103aa9190610b86565b610724565b6040516103bc9190610d40565b60405180910390f35b3480156103d157600080fd5b506103ec60048036038101906103e79190610f29565b61073c565b6040516103f99190610d40565b60405180910390f35b34801561040e57600080fd5b50610417610754565b6040516104249190610d40565b60405180910390f35b61044760048036038101906104429190610b86565b610778565b005b34801561045557600080fd5b5061045e6107aa565b60405161046b9190610d40565b60405180910390f35b34801561048057600080fd5b5061049b60048036038101906104969190610f29565b6107ce565b005b600080821415801561053c57507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663e493ef8c6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610515573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906105399190610f6b565b82105b9050919050565b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000081565b6040517fd623472500000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b7f000000000000000000000000000000000000000000000000000000000000000081565b60036020528060005260406000206000915090505481565b60046020528060005260406000206000915054906101000a900460ff1681565b610645610851565b61064f60006108cf565b565b610659610851565b600082829050905060005b818110156106995761068e84848381811061068257610681610f98565b5b90506020020135610993565b806001019050610664565b50505050565b6040517fd623472500000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b7f000000000000000000000000000000000000000000000000000000000000000081565b60015481565b60026020528060005260406000206000915090505481565b60056020528060005260406000206000915090505481565b7f000000000000000000000000000000000000000000000000000000000000000081565b6040517fd623472500000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b7f000000000000000000000000000000000000000000000000000000000000000081565b6107d6610851565b600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1603610845576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161083c9061104a565b60405180910390fd5b61084e816108cf565b50565b610859610a39565b73ffffffffffffffffffffffffffffffffffffffff166108776106d1565b73ffffffffffffffffffffffffffffffffffffffff16146108cd576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016108c4906110b6565b60405180910390fd5b565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050816000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a35050565b61099c81610a41565b600154600360008381526020019081526020016000208190555060016004600083815260200190815260200160002060006101000a81548160ff0219169083151502179055507f5a92c2530f207992057b9c3e544108ffce3beda4a63719f316967c49bf6159d281600154604051610a159291906110d6565b60405180910390a16001806000828254610a2f919061112e565b9250508190555050565b600033905090565b610a4a8161049d565b610a8b57806040517f7f3e75af000000000000000000000000000000000000000000000000000000008152600401610a829190610d40565b60405180910390fd5b600115156004600083815260200190815260200160002060009054906101000a900460ff16151503610ae8576040517e0a60f700000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b7f000000000000000000000000000000000000000000000000000000000000000060015410610b43576040517f57f6953100000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b50565b600080fd5b600080fd5b6000819050919050565b610b6381610b50565b8114610b6e57600080fd5b50565b600081359050610b8081610b5a565b92915050565b600060208284031215610b9c57610b9b610b46565b5b6000610baa84828501610b71565b91505092915050565b60008115159050919050565b610bc881610bb3565b82525050565b6000602082019050610be36000830184610bbf565b92915050565b600061ffff82169050919050565b610c0081610be9565b82525050565b6000602082019050610c1b6000830184610bf7565b92915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b6000610c66610c61610c5c84610c21565b610c41565b610c21565b9050919050565b6000610c7882610c4b565b9050919050565b6000610c8a82610c6d565b9050919050565b610c9a81610c7f565b82525050565b6000602082019050610cb56000830184610c91565b92915050565b6000610cc682610c6d565b9050919050565b610cd681610cbb565b82525050565b6000602082019050610cf16000830184610ccd565b92915050565b600063ffffffff82169050919050565b610d1081610cf7565b82525050565b6000602082019050610d2b6000830184610d07565b92915050565b610d3a81610b50565b82525050565b6000602082019050610d556000830184610d31565b92915050565b600080fd5b600080fd5b600080fd5b60008083601f840112610d8057610d7f610d5b565b5b8235905067ffffffffffffffff811115610d9d57610d9c610d60565b5b602083019150836020820283011115610db957610db8610d65565b5b9250929050565b60008060208385031215610dd757610dd6610b46565b5b600083013567ffffffffffffffff811115610df557610df4610b4b565b5b610e0185828601610d6a565b92509250509250929050565b6000610e1882610c21565b9050919050565b610e2881610e0d565b8114610e3357600080fd5b50565b600081359050610e4581610e1f565b92915050565b600081905082602060080282011115610e6757610e66610d65565b5b92915050565b60008060006101408486031215610e8757610e86610b46565b5b6000610e9586828701610b71565b9350506020610ea686828701610e36565b9250506040610eb786828701610e4b565b9150509250925092565b6000610ecc82610c21565b9050919050565b610edc81610ec1565b82525050565b6000602082019050610ef76000830184610ed3565b92915050565b610f0681610ec1565b8114610f1157600080fd5b50565b600081359050610f2381610efd565b92915050565b600060208284031215610f3f57610f3e610b46565b5b6000610f4d84828501610f14565b91505092915050565b600081519050610f6581610b5a565b92915050565b600060208284031215610f8157610f80610b46565b5b6000610f8f84828501610f56565b91505092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b600082825260208201905092915050565b7f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160008201527f6464726573730000000000000000000000000000000000000000000000000000602082015250565b6000611034602683610fc7565b915061103f82610fd8565b604082019050919050565b6000602082019050818103600083015261106381611027565b9050919050565b7f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572600082015250565b60006110a0602083610fc7565b91506110ab8261106a565b602082019050919050565b600060208201905081810360008301526110cf81611093565b9050919050565b60006040820190506110eb6000830185610d31565b6110f86020830184610d31565b9392505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061113982610b50565b915061114483610b50565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff03821115611179576111786110ff565b5b82820190509291505056fea26469706673582212205b9493b64379a82ad7d7b62004d979695bc51f0217dab7805001188ca178c99964736f6c634300080f0033",
}

// RLNABI is the input ABI used to generate the binding from.
// Deprecated: Use RLNMetaData.ABI instead.
var RLNABI = RLNMetaData.ABI

// RLNBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use RLNMetaData.Bin instead.
var RLNBin = RLNMetaData.Bin

// DeployRLN deploys a new Ethereum contract, binding an instance of RLN to it.
func DeployRLN(auth *bind.TransactOpts, backend bind.ContractBackend, _poseidonHasher common.Address, _contractIndex uint16) (common.Address, *types.Transaction, *RLN, error) {
	parsed, err := RLNMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(RLNBin), backend, _poseidonHasher, _contractIndex)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &RLN{RLNCaller: RLNCaller{contract: contract}, RLNTransactor: RLNTransactor{contract: contract}, RLNFilterer: RLNFilterer{contract: contract}}, nil
}

// RLN is an auto generated Go binding around an Ethereum contract.
type RLN struct {
	RLNCaller     // Read-only binding to the contract
	RLNTransactor // Write-only binding to the contract
	RLNFilterer   // Log filterer for contract events
}

// RLNCaller is an auto generated read-only Go binding around an Ethereum contract.
type RLNCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RLNTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RLNTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RLNFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RLNFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RLNSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RLNSession struct {
	Contract     *RLN              // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RLNCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RLNCallerSession struct {
	Contract *RLNCaller    // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// RLNTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RLNTransactorSession struct {
	Contract     *RLNTransactor    // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RLNRaw is an auto generated low-level Go binding around an Ethereum contract.
type RLNRaw struct {
	Contract *RLN // Generic contract binding to access the raw methods on
}

// RLNCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RLNCallerRaw struct {
	Contract *RLNCaller // Generic read-only contract binding to access the raw methods on
}

// RLNTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RLNTransactorRaw struct {
	Contract *RLNTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRLN creates a new instance of RLN, bound to a specific deployed contract.
func NewRLN(address common.Address, backend bind.ContractBackend) (*RLN, error) {
	contract, err := bindRLN(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &RLN{RLNCaller: RLNCaller{contract: contract}, RLNTransactor: RLNTransactor{contract: contract}, RLNFilterer: RLNFilterer{contract: contract}}, nil
}

// NewRLNCaller creates a new read-only instance of RLN, bound to a specific deployed contract.
func NewRLNCaller(address common.Address, caller bind.ContractCaller) (*RLNCaller, error) {
	contract, err := bindRLN(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RLNCaller{contract: contract}, nil
}

// NewRLNTransactor creates a new write-only instance of RLN, bound to a specific deployed contract.
func NewRLNTransactor(address common.Address, transactor bind.ContractTransactor) (*RLNTransactor, error) {
	contract, err := bindRLN(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RLNTransactor{contract: contract}, nil
}

// NewRLNFilterer creates a new log filterer instance of RLN, bound to a specific deployed contract.
func NewRLNFilterer(address common.Address, filterer bind.ContractFilterer) (*RLNFilterer, error) {
	contract, err := bindRLN(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RLNFilterer{contract: contract}, nil
}

// bindRLN binds a generic wrapper to an already deployed contract.
func bindRLN(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := RLNMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RLN *RLNRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _RLN.Contract.RLNCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RLN *RLNRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RLN.Contract.RLNTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RLN *RLNRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RLN.Contract.RLNTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RLN *RLNCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _RLN.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RLN *RLNTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RLN.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RLN *RLNTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RLN.Contract.contract.Transact(opts, method, params...)
}

// DEPTH is a free data retrieval call binding the contract method 0x98366e35.
//
// Solidity: function DEPTH() view returns(uint256)
func (_RLN *RLNCaller) DEPTH(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _RLN.contract.Call(opts, &out, "DEPTH")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DEPTH is a free data retrieval call binding the contract method 0x98366e35.
//
// Solidity: function DEPTH() view returns(uint256)
func (_RLN *RLNSession) DEPTH() (*big.Int, error) {
	return _RLN.Contract.DEPTH(&_RLN.CallOpts)
}

// DEPTH is a free data retrieval call binding the contract method 0x98366e35.
//
// Solidity: function DEPTH() view returns(uint256)
func (_RLN *RLNCallerSession) DEPTH() (*big.Int, error) {
	return _RLN.Contract.DEPTH(&_RLN.CallOpts)
}

// MEMBERSHIPDEPOSIT is a free data retrieval call binding the contract method 0xf220b9ec.
//
// Solidity: function MEMBERSHIP_DEPOSIT() view returns(uint256)
func (_RLN *RLNCaller) MEMBERSHIPDEPOSIT(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _RLN.contract.Call(opts, &out, "MEMBERSHIP_DEPOSIT")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MEMBERSHIPDEPOSIT is a free data retrieval call binding the contract method 0xf220b9ec.
//
// Solidity: function MEMBERSHIP_DEPOSIT() view returns(uint256)
func (_RLN *RLNSession) MEMBERSHIPDEPOSIT() (*big.Int, error) {
	return _RLN.Contract.MEMBERSHIPDEPOSIT(&_RLN.CallOpts)
}

// MEMBERSHIPDEPOSIT is a free data retrieval call binding the contract method 0xf220b9ec.
//
// Solidity: function MEMBERSHIP_DEPOSIT() view returns(uint256)
func (_RLN *RLNCallerSession) MEMBERSHIPDEPOSIT() (*big.Int, error) {
	return _RLN.Contract.MEMBERSHIPDEPOSIT(&_RLN.CallOpts)
}

// SETSIZE is a free data retrieval call binding the contract method 0xd0383d68.
//
// Solidity: function SET_SIZE() view returns(uint256)
func (_RLN *RLNCaller) SETSIZE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _RLN.contract.Call(opts, &out, "SET_SIZE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SETSIZE is a free data retrieval call binding the contract method 0xd0383d68.
//
// Solidity: function SET_SIZE() view returns(uint256)
func (_RLN *RLNSession) SETSIZE() (*big.Int, error) {
	return _RLN.Contract.SETSIZE(&_RLN.CallOpts)
}

// SETSIZE is a free data retrieval call binding the contract method 0xd0383d68.
//
// Solidity: function SET_SIZE() view returns(uint256)
func (_RLN *RLNCallerSession) SETSIZE() (*big.Int, error) {
	return _RLN.Contract.SETSIZE(&_RLN.CallOpts)
}

// ContractIndex is a free data retrieval call binding the contract method 0x28b070e0.
//
// Solidity: function contractIndex() view returns(uint16)
func (_RLN *RLNCaller) ContractIndex(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _RLN.contract.Call(opts, &out, "contractIndex")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// ContractIndex is a free data retrieval call binding the contract method 0x28b070e0.
//
// Solidity: function contractIndex() view returns(uint16)
func (_RLN *RLNSession) ContractIndex() (uint16, error) {
	return _RLN.Contract.ContractIndex(&_RLN.CallOpts)
}

// ContractIndex is a free data retrieval call binding the contract method 0x28b070e0.
//
// Solidity: function contractIndex() view returns(uint16)
func (_RLN *RLNCallerSession) ContractIndex() (uint16, error) {
	return _RLN.Contract.ContractIndex(&_RLN.CallOpts)
}

// DeployedBlockNumber is a free data retrieval call binding the contract method 0x4add651e.
//
// Solidity: function deployedBlockNumber() view returns(uint32)
func (_RLN *RLNCaller) DeployedBlockNumber(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _RLN.contract.Call(opts, &out, "deployedBlockNumber")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// DeployedBlockNumber is a free data retrieval call binding the contract method 0x4add651e.
//
// Solidity: function deployedBlockNumber() view returns(uint32)
func (_RLN *RLNSession) DeployedBlockNumber() (uint32, error) {
	return _RLN.Contract.DeployedBlockNumber(&_RLN.CallOpts)
}

// DeployedBlockNumber is a free data retrieval call binding the contract method 0x4add651e.
//
// Solidity: function deployedBlockNumber() view returns(uint32)
func (_RLN *RLNCallerSession) DeployedBlockNumber() (uint32, error) {
	return _RLN.Contract.DeployedBlockNumber(&_RLN.CallOpts)
}

// IdCommitmentIndex is a free data retrieval call binding the contract method 0xae74552a.
//
// Solidity: function idCommitmentIndex() view returns(uint256)
func (_RLN *RLNCaller) IdCommitmentIndex(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _RLN.contract.Call(opts, &out, "idCommitmentIndex")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// IdCommitmentIndex is a free data retrieval call binding the contract method 0xae74552a.
//
// Solidity: function idCommitmentIndex() view returns(uint256)
func (_RLN *RLNSession) IdCommitmentIndex() (*big.Int, error) {
	return _RLN.Contract.IdCommitmentIndex(&_RLN.CallOpts)
}

// IdCommitmentIndex is a free data retrieval call binding the contract method 0xae74552a.
//
// Solidity: function idCommitmentIndex() view returns(uint256)
func (_RLN *RLNCallerSession) IdCommitmentIndex() (*big.Int, error) {
	return _RLN.Contract.IdCommitmentIndex(&_RLN.CallOpts)
}

// IsValidCommitment is a free data retrieval call binding the contract method 0x22d9730c.
//
// Solidity: function isValidCommitment(uint256 idCommitment) view returns(bool)
func (_RLN *RLNCaller) IsValidCommitment(opts *bind.CallOpts, idCommitment *big.Int) (bool, error) {
	var out []interface{}
	err := _RLN.contract.Call(opts, &out, "isValidCommitment", idCommitment)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValidCommitment is a free data retrieval call binding the contract method 0x22d9730c.
//
// Solidity: function isValidCommitment(uint256 idCommitment) view returns(bool)
func (_RLN *RLNSession) IsValidCommitment(idCommitment *big.Int) (bool, error) {
	return _RLN.Contract.IsValidCommitment(&_RLN.CallOpts, idCommitment)
}

// IsValidCommitment is a free data retrieval call binding the contract method 0x22d9730c.
//
// Solidity: function isValidCommitment(uint256 idCommitment) view returns(bool)
func (_RLN *RLNCallerSession) IsValidCommitment(idCommitment *big.Int) (bool, error) {
	return _RLN.Contract.IsValidCommitment(&_RLN.CallOpts, idCommitment)
}

// MemberExists is a free data retrieval call binding the contract method 0x6bdcc8ab.
//
// Solidity: function memberExists(uint256 ) view returns(bool)
func (_RLN *RLNCaller) MemberExists(opts *bind.CallOpts, arg0 *big.Int) (bool, error) {
	var out []interface{}
	err := _RLN.contract.Call(opts, &out, "memberExists", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// MemberExists is a free data retrieval call binding the contract method 0x6bdcc8ab.
//
// Solidity: function memberExists(uint256 ) view returns(bool)
func (_RLN *RLNSession) MemberExists(arg0 *big.Int) (bool, error) {
	return _RLN.Contract.MemberExists(&_RLN.CallOpts, arg0)
}

// MemberExists is a free data retrieval call binding the contract method 0x6bdcc8ab.
//
// Solidity: function memberExists(uint256 ) view returns(bool)
func (_RLN *RLNCallerSession) MemberExists(arg0 *big.Int) (bool, error) {
	return _RLN.Contract.MemberExists(&_RLN.CallOpts, arg0)
}

// Members is a free data retrieval call binding the contract method 0x5daf08ca.
//
// Solidity: function members(uint256 ) view returns(uint256)
func (_RLN *RLNCaller) Members(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _RLN.contract.Call(opts, &out, "members", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Members is a free data retrieval call binding the contract method 0x5daf08ca.
//
// Solidity: function members(uint256 ) view returns(uint256)
func (_RLN *RLNSession) Members(arg0 *big.Int) (*big.Int, error) {
	return _RLN.Contract.Members(&_RLN.CallOpts, arg0)
}

// Members is a free data retrieval call binding the contract method 0x5daf08ca.
//
// Solidity: function members(uint256 ) view returns(uint256)
func (_RLN *RLNCallerSession) Members(arg0 *big.Int) (*big.Int, error) {
	return _RLN.Contract.Members(&_RLN.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_RLN *RLNCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _RLN.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_RLN *RLNSession) Owner() (common.Address, error) {
	return _RLN.Contract.Owner(&_RLN.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_RLN *RLNCallerSession) Owner() (common.Address, error) {
	return _RLN.Contract.Owner(&_RLN.CallOpts)
}

// PoseidonHasher is a free data retrieval call binding the contract method 0x331b6ab3.
//
// Solidity: function poseidonHasher() view returns(address)
func (_RLN *RLNCaller) PoseidonHasher(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _RLN.contract.Call(opts, &out, "poseidonHasher")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PoseidonHasher is a free data retrieval call binding the contract method 0x331b6ab3.
//
// Solidity: function poseidonHasher() view returns(address)
func (_RLN *RLNSession) PoseidonHasher() (common.Address, error) {
	return _RLN.Contract.PoseidonHasher(&_RLN.CallOpts)
}

// PoseidonHasher is a free data retrieval call binding the contract method 0x331b6ab3.
//
// Solidity: function poseidonHasher() view returns(address)
func (_RLN *RLNCallerSession) PoseidonHasher() (common.Address, error) {
	return _RLN.Contract.PoseidonHasher(&_RLN.CallOpts)
}

// Slash is a free data retrieval call binding the contract method 0x8be9b119.
//
// Solidity: function slash(uint256 idCommitment, address receiver, uint256[8] proof) pure returns()
func (_RLN *RLNCaller) Slash(opts *bind.CallOpts, idCommitment *big.Int, receiver common.Address, proof [8]*big.Int) error {
	var out []interface{}
	err := _RLN.contract.Call(opts, &out, "slash", idCommitment, receiver, proof)

	if err != nil {
		return err
	}

	return err

}

// Slash is a free data retrieval call binding the contract method 0x8be9b119.
//
// Solidity: function slash(uint256 idCommitment, address receiver, uint256[8] proof) pure returns()
func (_RLN *RLNSession) Slash(idCommitment *big.Int, receiver common.Address, proof [8]*big.Int) error {
	return _RLN.Contract.Slash(&_RLN.CallOpts, idCommitment, receiver, proof)
}

// Slash is a free data retrieval call binding the contract method 0x8be9b119.
//
// Solidity: function slash(uint256 idCommitment, address receiver, uint256[8] proof) pure returns()
func (_RLN *RLNCallerSession) Slash(idCommitment *big.Int, receiver common.Address, proof [8]*big.Int) error {
	return _RLN.Contract.Slash(&_RLN.CallOpts, idCommitment, receiver, proof)
}

// StakedAmounts is a free data retrieval call binding the contract method 0xbc499128.
//
// Solidity: function stakedAmounts(uint256 ) view returns(uint256)
func (_RLN *RLNCaller) StakedAmounts(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _RLN.contract.Call(opts, &out, "stakedAmounts", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StakedAmounts is a free data retrieval call binding the contract method 0xbc499128.
//
// Solidity: function stakedAmounts(uint256 ) view returns(uint256)
func (_RLN *RLNSession) StakedAmounts(arg0 *big.Int) (*big.Int, error) {
	return _RLN.Contract.StakedAmounts(&_RLN.CallOpts, arg0)
}

// StakedAmounts is a free data retrieval call binding the contract method 0xbc499128.
//
// Solidity: function stakedAmounts(uint256 ) view returns(uint256)
func (_RLN *RLNCallerSession) StakedAmounts(arg0 *big.Int) (*big.Int, error) {
	return _RLN.Contract.StakedAmounts(&_RLN.CallOpts, arg0)
}

// Verifier is a free data retrieval call binding the contract method 0x2b7ac3f3.
//
// Solidity: function verifier() view returns(address)
func (_RLN *RLNCaller) Verifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _RLN.contract.Call(opts, &out, "verifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Verifier is a free data retrieval call binding the contract method 0x2b7ac3f3.
//
// Solidity: function verifier() view returns(address)
func (_RLN *RLNSession) Verifier() (common.Address, error) {
	return _RLN.Contract.Verifier(&_RLN.CallOpts)
}

// Verifier is a free data retrieval call binding the contract method 0x2b7ac3f3.
//
// Solidity: function verifier() view returns(address)
func (_RLN *RLNCallerSession) Verifier() (common.Address, error) {
	return _RLN.Contract.Verifier(&_RLN.CallOpts)
}

// Withdraw is a free data retrieval call binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() pure returns()
func (_RLN *RLNCaller) Withdraw(opts *bind.CallOpts) error {
	var out []interface{}
	err := _RLN.contract.Call(opts, &out, "withdraw")

	if err != nil {
		return err
	}

	return err

}

// Withdraw is a free data retrieval call binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() pure returns()
func (_RLN *RLNSession) Withdraw() error {
	return _RLN.Contract.Withdraw(&_RLN.CallOpts)
}

// Withdraw is a free data retrieval call binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() pure returns()
func (_RLN *RLNCallerSession) Withdraw() error {
	return _RLN.Contract.Withdraw(&_RLN.CallOpts)
}

// WithdrawalBalance is a free data retrieval call binding the contract method 0xc5b208ff.
//
// Solidity: function withdrawalBalance(address ) view returns(uint256)
func (_RLN *RLNCaller) WithdrawalBalance(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _RLN.contract.Call(opts, &out, "withdrawalBalance", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// WithdrawalBalance is a free data retrieval call binding the contract method 0xc5b208ff.
//
// Solidity: function withdrawalBalance(address ) view returns(uint256)
func (_RLN *RLNSession) WithdrawalBalance(arg0 common.Address) (*big.Int, error) {
	return _RLN.Contract.WithdrawalBalance(&_RLN.CallOpts, arg0)
}

// WithdrawalBalance is a free data retrieval call binding the contract method 0xc5b208ff.
//
// Solidity: function withdrawalBalance(address ) view returns(uint256)
func (_RLN *RLNCallerSession) WithdrawalBalance(arg0 common.Address) (*big.Int, error) {
	return _RLN.Contract.WithdrawalBalance(&_RLN.CallOpts, arg0)
}

// Register is a paid mutator transaction binding the contract method 0x7a34289d.
//
// Solidity: function register(uint256[] idCommitments) returns()
func (_RLN *RLNTransactor) Register(opts *bind.TransactOpts, idCommitments []*big.Int) (*types.Transaction, error) {
	return _RLN.contract.Transact(opts, "register", idCommitments)
}

// Register is a paid mutator transaction binding the contract method 0x7a34289d.
//
// Solidity: function register(uint256[] idCommitments) returns()
func (_RLN *RLNSession) Register(idCommitments []*big.Int) (*types.Transaction, error) {
	return _RLN.Contract.Register(&_RLN.TransactOpts, idCommitments)
}

// Register is a paid mutator transaction binding the contract method 0x7a34289d.
//
// Solidity: function register(uint256[] idCommitments) returns()
func (_RLN *RLNTransactorSession) Register(idCommitments []*big.Int) (*types.Transaction, error) {
	return _RLN.Contract.Register(&_RLN.TransactOpts, idCommitments)
}

// Register0 is a paid mutator transaction binding the contract method 0xf207564e.
//
// Solidity: function register(uint256 idCommitment) payable returns()
func (_RLN *RLNTransactor) Register0(opts *bind.TransactOpts, idCommitment *big.Int) (*types.Transaction, error) {
	return _RLN.contract.Transact(opts, "register0", idCommitment)
}

// Register0 is a paid mutator transaction binding the contract method 0xf207564e.
//
// Solidity: function register(uint256 idCommitment) payable returns()
func (_RLN *RLNSession) Register0(idCommitment *big.Int) (*types.Transaction, error) {
	return _RLN.Contract.Register0(&_RLN.TransactOpts, idCommitment)
}

// Register0 is a paid mutator transaction binding the contract method 0xf207564e.
//
// Solidity: function register(uint256 idCommitment) payable returns()
func (_RLN *RLNTransactorSession) Register0(idCommitment *big.Int) (*types.Transaction, error) {
	return _RLN.Contract.Register0(&_RLN.TransactOpts, idCommitment)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_RLN *RLNTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RLN.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_RLN *RLNSession) RenounceOwnership() (*types.Transaction, error) {
	return _RLN.Contract.RenounceOwnership(&_RLN.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_RLN *RLNTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _RLN.Contract.RenounceOwnership(&_RLN.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_RLN *RLNTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _RLN.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_RLN *RLNSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _RLN.Contract.TransferOwnership(&_RLN.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_RLN *RLNTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _RLN.Contract.TransferOwnership(&_RLN.TransactOpts, newOwner)
}

// RLNMemberRegisteredIterator is returned from FilterMemberRegistered and is used to iterate over the raw logs and unpacked data for MemberRegistered events raised by the RLN contract.
type RLNMemberRegisteredIterator struct {
	Event *RLNMemberRegistered // Event containing the contract specifics and raw log

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
func (it *RLNMemberRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RLNMemberRegistered)
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
		it.Event = new(RLNMemberRegistered)
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
func (it *RLNMemberRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RLNMemberRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RLNMemberRegistered represents a MemberRegistered event raised by the RLN contract.
type RLNMemberRegistered struct {
	IdCommitment *big.Int
	Index        *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterMemberRegistered is a free log retrieval operation binding the contract event 0x5a92c2530f207992057b9c3e544108ffce3beda4a63719f316967c49bf6159d2.
//
// Solidity: event MemberRegistered(uint256 idCommitment, uint256 index)
func (_RLN *RLNFilterer) FilterMemberRegistered(opts *bind.FilterOpts) (*RLNMemberRegisteredIterator, error) {

	logs, sub, err := _RLN.contract.FilterLogs(opts, "MemberRegistered")
	if err != nil {
		return nil, err
	}
	return &RLNMemberRegisteredIterator{contract: _RLN.contract, event: "MemberRegistered", logs: logs, sub: sub}, nil
}

// WatchMemberRegistered is a free log subscription operation binding the contract event 0x5a92c2530f207992057b9c3e544108ffce3beda4a63719f316967c49bf6159d2.
//
// Solidity: event MemberRegistered(uint256 idCommitment, uint256 index)
func (_RLN *RLNFilterer) WatchMemberRegistered(opts *bind.WatchOpts, sink chan<- *RLNMemberRegistered) (event.Subscription, error) {

	logs, sub, err := _RLN.contract.WatchLogs(opts, "MemberRegistered")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RLNMemberRegistered)
				if err := _RLN.contract.UnpackLog(event, "MemberRegistered", log); err != nil {
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

// ParseMemberRegistered is a log parse operation binding the contract event 0x5a92c2530f207992057b9c3e544108ffce3beda4a63719f316967c49bf6159d2.
//
// Solidity: event MemberRegistered(uint256 idCommitment, uint256 index)
func (_RLN *RLNFilterer) ParseMemberRegistered(log types.Log) (*RLNMemberRegistered, error) {
	event := new(RLNMemberRegistered)
	if err := _RLN.contract.UnpackLog(event, "MemberRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RLNMemberWithdrawnIterator is returned from FilterMemberWithdrawn and is used to iterate over the raw logs and unpacked data for MemberWithdrawn events raised by the RLN contract.
type RLNMemberWithdrawnIterator struct {
	Event *RLNMemberWithdrawn // Event containing the contract specifics and raw log

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
func (it *RLNMemberWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RLNMemberWithdrawn)
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
		it.Event = new(RLNMemberWithdrawn)
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
func (it *RLNMemberWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RLNMemberWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RLNMemberWithdrawn represents a MemberWithdrawn event raised by the RLN contract.
type RLNMemberWithdrawn struct {
	IdCommitment *big.Int
	Index        *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterMemberWithdrawn is a free log retrieval operation binding the contract event 0x62ec3a516d22a993ce5cb4e7593e878c74f4d799dde522a88dc27a994fd5a943.
//
// Solidity: event MemberWithdrawn(uint256 idCommitment, uint256 index)
func (_RLN *RLNFilterer) FilterMemberWithdrawn(opts *bind.FilterOpts) (*RLNMemberWithdrawnIterator, error) {

	logs, sub, err := _RLN.contract.FilterLogs(opts, "MemberWithdrawn")
	if err != nil {
		return nil, err
	}
	return &RLNMemberWithdrawnIterator{contract: _RLN.contract, event: "MemberWithdrawn", logs: logs, sub: sub}, nil
}

// WatchMemberWithdrawn is a free log subscription operation binding the contract event 0x62ec3a516d22a993ce5cb4e7593e878c74f4d799dde522a88dc27a994fd5a943.
//
// Solidity: event MemberWithdrawn(uint256 idCommitment, uint256 index)
func (_RLN *RLNFilterer) WatchMemberWithdrawn(opts *bind.WatchOpts, sink chan<- *RLNMemberWithdrawn) (event.Subscription, error) {

	logs, sub, err := _RLN.contract.WatchLogs(opts, "MemberWithdrawn")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RLNMemberWithdrawn)
				if err := _RLN.contract.UnpackLog(event, "MemberWithdrawn", log); err != nil {
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

// ParseMemberWithdrawn is a log parse operation binding the contract event 0x62ec3a516d22a993ce5cb4e7593e878c74f4d799dde522a88dc27a994fd5a943.
//
// Solidity: event MemberWithdrawn(uint256 idCommitment, uint256 index)
func (_RLN *RLNFilterer) ParseMemberWithdrawn(log types.Log) (*RLNMemberWithdrawn, error) {
	event := new(RLNMemberWithdrawn)
	if err := _RLN.contract.UnpackLog(event, "MemberWithdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RLNOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the RLN contract.
type RLNOwnershipTransferredIterator struct {
	Event *RLNOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *RLNOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RLNOwnershipTransferred)
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
		it.Event = new(RLNOwnershipTransferred)
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
func (it *RLNOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RLNOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RLNOwnershipTransferred represents a OwnershipTransferred event raised by the RLN contract.
type RLNOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_RLN *RLNFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*RLNOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _RLN.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &RLNOwnershipTransferredIterator{contract: _RLN.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_RLN *RLNFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *RLNOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _RLN.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RLNOwnershipTransferred)
				if err := _RLN.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_RLN *RLNFilterer) ParseOwnershipTransferred(log types.Log) (*RLNOwnershipTransferred, error) {
	event := new(RLNOwnershipTransferred)
	if err := _RLN.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
