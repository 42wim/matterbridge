// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package hopBridge

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

// BridgeTransferRoot is an auto generated low-level Go binding around an user-defined struct.
type BridgeTransferRoot struct {
	Total           *big.Int
	AmountWithdrawn *big.Int
	CreatedAt       *big.Int
}

// HopBridgeABI is the input ABI used to generate the binding from.
const HopBridgeABI = "[{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"_l1CanonicalToken\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"bonders\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"_governance\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newBonder\",\"type\":\"address\"}],\"name\":\"BonderAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousBonder\",\"type\":\"address\"}],\"name\":\"BonderRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"transferRootId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"originalAmount\",\"type\":\"uint256\"}],\"name\":\"ChallengeResolved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"bonder\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"totalBondsSettled\",\"type\":\"uint256\"}],\"name\":\"MultipleWithdrawalsSettled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Stake\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"transferRootId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"originalAmount\",\"type\":\"uint256\"}],\"name\":\"TransferBondChallenged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"root\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"TransferRootBonded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"originChainId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"destinationChainId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"totalAmount\",\"type\":\"uint256\"}],\"name\":\"TransferRootConfirmed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"totalAmount\",\"type\":\"uint256\"}],\"name\":\"TransferRootSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amountOutMin\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"relayer\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"relayerFee\",\"type\":\"uint256\"}],\"name\":\"TransferSentToL2\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Unstake\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"bonder\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"transferId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"rootHash\",\"type\":\"bytes32\"}],\"name\":\"WithdrawalBondSettled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"transferId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"WithdrawalBonded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"transferId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"transferNonce\",\"type\":\"bytes32\"}],\"name\":\"Withdrew\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"CHALLENGE_AMOUNT_DIVISOR\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"TIME_SLOT_SIZE\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bonder\",\"type\":\"address\"}],\"name\":\"addBonder\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"destinationChainId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalAmount\",\"type\":\"uint256\"}],\"name\":\"bondTransferRoot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"transferNonce\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"bonderFee\",\"type\":\"uint256\"}],\"name\":\"bondWithdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"chainBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"challengePeriod\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"challengeResolutionPeriod\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"originalAmount\",\"type\":\"uint256\"}],\"name\":\"challengeTransferBond\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"originChainId\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"destinationChainId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rootCommittedAt\",\"type\":\"uint256\"}],\"name\":\"confirmTransferRoot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"crossDomainMessengerWrappers\",\"outputs\":[{\"internalType\":\"contractIMessengerWrapper\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"getBondForTransferAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bonder\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"transferId\",\"type\":\"bytes32\"}],\"name\":\"getBondedWithdrawalAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getChainId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"getChallengeAmountForTransferAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bonder\",\"type\":\"address\"}],\"name\":\"getCredit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bonder\",\"type\":\"address\"}],\"name\":\"getDebitAndAdditionalDebit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"maybeBonder\",\"type\":\"address\"}],\"name\":\"getIsBonder\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bonder\",\"type\":\"address\"}],\"name\":\"getRawDebit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"}],\"name\":\"getTimeSlot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"transferNonce\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"bonderFee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amountOutMin\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"name\":\"getTransferId\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"totalAmount\",\"type\":\"uint256\"}],\"name\":\"getTransferRoot\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"total\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amountWithdrawn\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"}],\"internalType\":\"structBridge.TransferRoot\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"totalAmount\",\"type\":\"uint256\"}],\"name\":\"getTransferRootId\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"governance\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"isChainIdPaused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"transferId\",\"type\":\"bytes32\"}],\"name\":\"isTransferIdSpent\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"l1CanonicalToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minTransferRootBondDelay\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bonder\",\"type\":\"address\"}],\"name\":\"removeBonder\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"originalAmount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"}],\"name\":\"rescueTransferRoot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"originalAmount\",\"type\":\"uint256\"}],\"name\":\"resolveChallenge\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amountOutMin\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"relayer\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"relayerFee\",\"type\":\"uint256\"}],\"name\":\"sendToL2\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isPaused\",\"type\":\"bool\"}],\"name\":\"setChainIdDepositsPaused\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_challengePeriod\",\"type\":\"uint256\"}],\"name\":\"setChallengePeriod\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_challengeResolutionPeriod\",\"type\":\"uint256\"}],\"name\":\"setChallengeResolutionPeriod\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"internalType\":\"contractIMessengerWrapper\",\"name\":\"_crossDomainMessengerWrapper\",\"type\":\"address\"}],\"name\":\"setCrossDomainMessengerWrapper\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_newGovernance\",\"type\":\"address\"}],\"name\":\"setGovernance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_minTransferRootBondDelay\",\"type\":\"uint256\"}],\"name\":\"setMinTransferRootBondDelay\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bonder\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"transferId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"transferRootTotalAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"transferIdTreeIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32[]\",\"name\":\"siblings\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"totalLeaves\",\"type\":\"uint256\"}],\"name\":\"settleBondedWithdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bonder\",\"type\":\"address\"},{\"internalType\":\"bytes32[]\",\"name\":\"transferIds\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"totalAmount\",\"type\":\"uint256\"}],\"name\":\"settleBondedWithdrawals\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bonder\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"stake\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"timeSlotToAmountBonded\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"transferBonds\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"bonder\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"challengeStartTime\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"challenger\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"challengeResolved\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"transferRootCommittedAt\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"unstake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"transferNonce\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"bonderFee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amountOutMin\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"rootHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"transferRootTotalAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"transferIdTreeIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32[]\",\"name\":\"siblings\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"totalLeaves\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// HopBridge is an auto generated Go binding around an Ethereum contract.
type HopBridge struct {
	HopBridgeCaller     // Read-only binding to the contract
	HopBridgeTransactor // Write-only binding to the contract
	HopBridgeFilterer   // Log filterer for contract events
}

// HopBridgeCaller is an auto generated read-only Go binding around an Ethereum contract.
type HopBridgeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HopBridgeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type HopBridgeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HopBridgeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type HopBridgeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HopBridgeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type HopBridgeSession struct {
	Contract     *HopBridge        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HopBridgeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type HopBridgeCallerSession struct {
	Contract *HopBridgeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// HopBridgeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type HopBridgeTransactorSession struct {
	Contract     *HopBridgeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// HopBridgeRaw is an auto generated low-level Go binding around an Ethereum contract.
type HopBridgeRaw struct {
	Contract *HopBridge // Generic contract binding to access the raw methods on
}

// HopBridgeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type HopBridgeCallerRaw struct {
	Contract *HopBridgeCaller // Generic read-only contract binding to access the raw methods on
}

// HopBridgeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type HopBridgeTransactorRaw struct {
	Contract *HopBridgeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewHopBridge creates a new instance of HopBridge, bound to a specific deployed contract.
func NewHopBridge(address common.Address, backend bind.ContractBackend) (*HopBridge, error) {
	contract, err := bindHopBridge(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &HopBridge{HopBridgeCaller: HopBridgeCaller{contract: contract}, HopBridgeTransactor: HopBridgeTransactor{contract: contract}, HopBridgeFilterer: HopBridgeFilterer{contract: contract}}, nil
}

// NewHopBridgeCaller creates a new read-only instance of HopBridge, bound to a specific deployed contract.
func NewHopBridgeCaller(address common.Address, caller bind.ContractCaller) (*HopBridgeCaller, error) {
	contract, err := bindHopBridge(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &HopBridgeCaller{contract: contract}, nil
}

// NewHopBridgeTransactor creates a new write-only instance of HopBridge, bound to a specific deployed contract.
func NewHopBridgeTransactor(address common.Address, transactor bind.ContractTransactor) (*HopBridgeTransactor, error) {
	contract, err := bindHopBridge(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &HopBridgeTransactor{contract: contract}, nil
}

// NewHopBridgeFilterer creates a new log filterer instance of HopBridge, bound to a specific deployed contract.
func NewHopBridgeFilterer(address common.Address, filterer bind.ContractFilterer) (*HopBridgeFilterer, error) {
	contract, err := bindHopBridge(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &HopBridgeFilterer{contract: contract}, nil
}

// bindHopBridge binds a generic wrapper to an already deployed contract.
func bindHopBridge(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(HopBridgeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HopBridge *HopBridgeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _HopBridge.Contract.HopBridgeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HopBridge *HopBridgeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HopBridge.Contract.HopBridgeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HopBridge *HopBridgeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HopBridge.Contract.HopBridgeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HopBridge *HopBridgeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _HopBridge.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HopBridge *HopBridgeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HopBridge.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HopBridge *HopBridgeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HopBridge.Contract.contract.Transact(opts, method, params...)
}

// CHALLENGEAMOUNTDIVISOR is a free data retrieval call binding the contract method 0x98c4f76d.
//
// Solidity: function CHALLENGE_AMOUNT_DIVISOR() view returns(uint256)
func (_HopBridge *HopBridgeCaller) CHALLENGEAMOUNTDIVISOR(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "CHALLENGE_AMOUNT_DIVISOR")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CHALLENGEAMOUNTDIVISOR is a free data retrieval call binding the contract method 0x98c4f76d.
//
// Solidity: function CHALLENGE_AMOUNT_DIVISOR() view returns(uint256)
func (_HopBridge *HopBridgeSession) CHALLENGEAMOUNTDIVISOR() (*big.Int, error) {
	return _HopBridge.Contract.CHALLENGEAMOUNTDIVISOR(&_HopBridge.CallOpts)
}

// CHALLENGEAMOUNTDIVISOR is a free data retrieval call binding the contract method 0x98c4f76d.
//
// Solidity: function CHALLENGE_AMOUNT_DIVISOR() view returns(uint256)
func (_HopBridge *HopBridgeCallerSession) CHALLENGEAMOUNTDIVISOR() (*big.Int, error) {
	return _HopBridge.Contract.CHALLENGEAMOUNTDIVISOR(&_HopBridge.CallOpts)
}

// TIMESLOTSIZE is a free data retrieval call binding the contract method 0x4de8c6e6.
//
// Solidity: function TIME_SLOT_SIZE() view returns(uint256)
func (_HopBridge *HopBridgeCaller) TIMESLOTSIZE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "TIME_SLOT_SIZE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TIMESLOTSIZE is a free data retrieval call binding the contract method 0x4de8c6e6.
//
// Solidity: function TIME_SLOT_SIZE() view returns(uint256)
func (_HopBridge *HopBridgeSession) TIMESLOTSIZE() (*big.Int, error) {
	return _HopBridge.Contract.TIMESLOTSIZE(&_HopBridge.CallOpts)
}

// TIMESLOTSIZE is a free data retrieval call binding the contract method 0x4de8c6e6.
//
// Solidity: function TIME_SLOT_SIZE() view returns(uint256)
func (_HopBridge *HopBridgeCallerSession) TIMESLOTSIZE() (*big.Int, error) {
	return _HopBridge.Contract.TIMESLOTSIZE(&_HopBridge.CallOpts)
}

// ChainBalance is a free data retrieval call binding the contract method 0xfc110b67.
//
// Solidity: function chainBalance(uint256 ) view returns(uint256)
func (_HopBridge *HopBridgeCaller) ChainBalance(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "chainBalance", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ChainBalance is a free data retrieval call binding the contract method 0xfc110b67.
//
// Solidity: function chainBalance(uint256 ) view returns(uint256)
func (_HopBridge *HopBridgeSession) ChainBalance(arg0 *big.Int) (*big.Int, error) {
	return _HopBridge.Contract.ChainBalance(&_HopBridge.CallOpts, arg0)
}

// ChainBalance is a free data retrieval call binding the contract method 0xfc110b67.
//
// Solidity: function chainBalance(uint256 ) view returns(uint256)
func (_HopBridge *HopBridgeCallerSession) ChainBalance(arg0 *big.Int) (*big.Int, error) {
	return _HopBridge.Contract.ChainBalance(&_HopBridge.CallOpts, arg0)
}

// ChallengePeriod is a free data retrieval call binding the contract method 0xf3f480d9.
//
// Solidity: function challengePeriod() view returns(uint256)
func (_HopBridge *HopBridgeCaller) ChallengePeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "challengePeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ChallengePeriod is a free data retrieval call binding the contract method 0xf3f480d9.
//
// Solidity: function challengePeriod() view returns(uint256)
func (_HopBridge *HopBridgeSession) ChallengePeriod() (*big.Int, error) {
	return _HopBridge.Contract.ChallengePeriod(&_HopBridge.CallOpts)
}

// ChallengePeriod is a free data retrieval call binding the contract method 0xf3f480d9.
//
// Solidity: function challengePeriod() view returns(uint256)
func (_HopBridge *HopBridgeCallerSession) ChallengePeriod() (*big.Int, error) {
	return _HopBridge.Contract.ChallengePeriod(&_HopBridge.CallOpts)
}

// ChallengeResolutionPeriod is a free data retrieval call binding the contract method 0x767631d5.
//
// Solidity: function challengeResolutionPeriod() view returns(uint256)
func (_HopBridge *HopBridgeCaller) ChallengeResolutionPeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "challengeResolutionPeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ChallengeResolutionPeriod is a free data retrieval call binding the contract method 0x767631d5.
//
// Solidity: function challengeResolutionPeriod() view returns(uint256)
func (_HopBridge *HopBridgeSession) ChallengeResolutionPeriod() (*big.Int, error) {
	return _HopBridge.Contract.ChallengeResolutionPeriod(&_HopBridge.CallOpts)
}

// ChallengeResolutionPeriod is a free data retrieval call binding the contract method 0x767631d5.
//
// Solidity: function challengeResolutionPeriod() view returns(uint256)
func (_HopBridge *HopBridgeCallerSession) ChallengeResolutionPeriod() (*big.Int, error) {
	return _HopBridge.Contract.ChallengeResolutionPeriod(&_HopBridge.CallOpts)
}

// CrossDomainMessengerWrappers is a free data retrieval call binding the contract method 0xa35962f3.
//
// Solidity: function crossDomainMessengerWrappers(uint256 ) view returns(address)
func (_HopBridge *HopBridgeCaller) CrossDomainMessengerWrappers(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "crossDomainMessengerWrappers", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CrossDomainMessengerWrappers is a free data retrieval call binding the contract method 0xa35962f3.
//
// Solidity: function crossDomainMessengerWrappers(uint256 ) view returns(address)
func (_HopBridge *HopBridgeSession) CrossDomainMessengerWrappers(arg0 *big.Int) (common.Address, error) {
	return _HopBridge.Contract.CrossDomainMessengerWrappers(&_HopBridge.CallOpts, arg0)
}

// CrossDomainMessengerWrappers is a free data retrieval call binding the contract method 0xa35962f3.
//
// Solidity: function crossDomainMessengerWrappers(uint256 ) view returns(address)
func (_HopBridge *HopBridgeCallerSession) CrossDomainMessengerWrappers(arg0 *big.Int) (common.Address, error) {
	return _HopBridge.Contract.CrossDomainMessengerWrappers(&_HopBridge.CallOpts, arg0)
}

// GetBondForTransferAmount is a free data retrieval call binding the contract method 0xe19be150.
//
// Solidity: function getBondForTransferAmount(uint256 amount) pure returns(uint256)
func (_HopBridge *HopBridgeCaller) GetBondForTransferAmount(opts *bind.CallOpts, amount *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "getBondForTransferAmount", amount)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetBondForTransferAmount is a free data retrieval call binding the contract method 0xe19be150.
//
// Solidity: function getBondForTransferAmount(uint256 amount) pure returns(uint256)
func (_HopBridge *HopBridgeSession) GetBondForTransferAmount(amount *big.Int) (*big.Int, error) {
	return _HopBridge.Contract.GetBondForTransferAmount(&_HopBridge.CallOpts, amount)
}

// GetBondForTransferAmount is a free data retrieval call binding the contract method 0xe19be150.
//
// Solidity: function getBondForTransferAmount(uint256 amount) pure returns(uint256)
func (_HopBridge *HopBridgeCallerSession) GetBondForTransferAmount(amount *big.Int) (*big.Int, error) {
	return _HopBridge.Contract.GetBondForTransferAmount(&_HopBridge.CallOpts, amount)
}

// GetBondedWithdrawalAmount is a free data retrieval call binding the contract method 0x302830ab.
//
// Solidity: function getBondedWithdrawalAmount(address bonder, bytes32 transferId) view returns(uint256)
func (_HopBridge *HopBridgeCaller) GetBondedWithdrawalAmount(opts *bind.CallOpts, bonder common.Address, transferId [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "getBondedWithdrawalAmount", bonder, transferId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetBondedWithdrawalAmount is a free data retrieval call binding the contract method 0x302830ab.
//
// Solidity: function getBondedWithdrawalAmount(address bonder, bytes32 transferId) view returns(uint256)
func (_HopBridge *HopBridgeSession) GetBondedWithdrawalAmount(bonder common.Address, transferId [32]byte) (*big.Int, error) {
	return _HopBridge.Contract.GetBondedWithdrawalAmount(&_HopBridge.CallOpts, bonder, transferId)
}

// GetBondedWithdrawalAmount is a free data retrieval call binding the contract method 0x302830ab.
//
// Solidity: function getBondedWithdrawalAmount(address bonder, bytes32 transferId) view returns(uint256)
func (_HopBridge *HopBridgeCallerSession) GetBondedWithdrawalAmount(bonder common.Address, transferId [32]byte) (*big.Int, error) {
	return _HopBridge.Contract.GetBondedWithdrawalAmount(&_HopBridge.CallOpts, bonder, transferId)
}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() view returns(uint256 chainId)
func (_HopBridge *HopBridgeCaller) GetChainId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "getChainId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() view returns(uint256 chainId)
func (_HopBridge *HopBridgeSession) GetChainId() (*big.Int, error) {
	return _HopBridge.Contract.GetChainId(&_HopBridge.CallOpts)
}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() view returns(uint256 chainId)
func (_HopBridge *HopBridgeCallerSession) GetChainId() (*big.Int, error) {
	return _HopBridge.Contract.GetChainId(&_HopBridge.CallOpts)
}

// GetChallengeAmountForTransferAmount is a free data retrieval call binding the contract method 0xa239f5ee.
//
// Solidity: function getChallengeAmountForTransferAmount(uint256 amount) pure returns(uint256)
func (_HopBridge *HopBridgeCaller) GetChallengeAmountForTransferAmount(opts *bind.CallOpts, amount *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "getChallengeAmountForTransferAmount", amount)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetChallengeAmountForTransferAmount is a free data retrieval call binding the contract method 0xa239f5ee.
//
// Solidity: function getChallengeAmountForTransferAmount(uint256 amount) pure returns(uint256)
func (_HopBridge *HopBridgeSession) GetChallengeAmountForTransferAmount(amount *big.Int) (*big.Int, error) {
	return _HopBridge.Contract.GetChallengeAmountForTransferAmount(&_HopBridge.CallOpts, amount)
}

// GetChallengeAmountForTransferAmount is a free data retrieval call binding the contract method 0xa239f5ee.
//
// Solidity: function getChallengeAmountForTransferAmount(uint256 amount) pure returns(uint256)
func (_HopBridge *HopBridgeCallerSession) GetChallengeAmountForTransferAmount(amount *big.Int) (*big.Int, error) {
	return _HopBridge.Contract.GetChallengeAmountForTransferAmount(&_HopBridge.CallOpts, amount)
}

// GetCredit is a free data retrieval call binding the contract method 0x57344e6f.
//
// Solidity: function getCredit(address bonder) view returns(uint256)
func (_HopBridge *HopBridgeCaller) GetCredit(opts *bind.CallOpts, bonder common.Address) (*big.Int, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "getCredit", bonder)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCredit is a free data retrieval call binding the contract method 0x57344e6f.
//
// Solidity: function getCredit(address bonder) view returns(uint256)
func (_HopBridge *HopBridgeSession) GetCredit(bonder common.Address) (*big.Int, error) {
	return _HopBridge.Contract.GetCredit(&_HopBridge.CallOpts, bonder)
}

// GetCredit is a free data retrieval call binding the contract method 0x57344e6f.
//
// Solidity: function getCredit(address bonder) view returns(uint256)
func (_HopBridge *HopBridgeCallerSession) GetCredit(bonder common.Address) (*big.Int, error) {
	return _HopBridge.Contract.GetCredit(&_HopBridge.CallOpts, bonder)
}

// GetDebitAndAdditionalDebit is a free data retrieval call binding the contract method 0xffa9286c.
//
// Solidity: function getDebitAndAdditionalDebit(address bonder) view returns(uint256)
func (_HopBridge *HopBridgeCaller) GetDebitAndAdditionalDebit(opts *bind.CallOpts, bonder common.Address) (*big.Int, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "getDebitAndAdditionalDebit", bonder)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetDebitAndAdditionalDebit is a free data retrieval call binding the contract method 0xffa9286c.
//
// Solidity: function getDebitAndAdditionalDebit(address bonder) view returns(uint256)
func (_HopBridge *HopBridgeSession) GetDebitAndAdditionalDebit(bonder common.Address) (*big.Int, error) {
	return _HopBridge.Contract.GetDebitAndAdditionalDebit(&_HopBridge.CallOpts, bonder)
}

// GetDebitAndAdditionalDebit is a free data retrieval call binding the contract method 0xffa9286c.
//
// Solidity: function getDebitAndAdditionalDebit(address bonder) view returns(uint256)
func (_HopBridge *HopBridgeCallerSession) GetDebitAndAdditionalDebit(bonder common.Address) (*big.Int, error) {
	return _HopBridge.Contract.GetDebitAndAdditionalDebit(&_HopBridge.CallOpts, bonder)
}

// GetIsBonder is a free data retrieval call binding the contract method 0xd5ef7551.
//
// Solidity: function getIsBonder(address maybeBonder) view returns(bool)
func (_HopBridge *HopBridgeCaller) GetIsBonder(opts *bind.CallOpts, maybeBonder common.Address) (bool, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "getIsBonder", maybeBonder)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// GetIsBonder is a free data retrieval call binding the contract method 0xd5ef7551.
//
// Solidity: function getIsBonder(address maybeBonder) view returns(bool)
func (_HopBridge *HopBridgeSession) GetIsBonder(maybeBonder common.Address) (bool, error) {
	return _HopBridge.Contract.GetIsBonder(&_HopBridge.CallOpts, maybeBonder)
}

// GetIsBonder is a free data retrieval call binding the contract method 0xd5ef7551.
//
// Solidity: function getIsBonder(address maybeBonder) view returns(bool)
func (_HopBridge *HopBridgeCallerSession) GetIsBonder(maybeBonder common.Address) (bool, error) {
	return _HopBridge.Contract.GetIsBonder(&_HopBridge.CallOpts, maybeBonder)
}

// GetRawDebit is a free data retrieval call binding the contract method 0x13948c76.
//
// Solidity: function getRawDebit(address bonder) view returns(uint256)
func (_HopBridge *HopBridgeCaller) GetRawDebit(opts *bind.CallOpts, bonder common.Address) (*big.Int, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "getRawDebit", bonder)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRawDebit is a free data retrieval call binding the contract method 0x13948c76.
//
// Solidity: function getRawDebit(address bonder) view returns(uint256)
func (_HopBridge *HopBridgeSession) GetRawDebit(bonder common.Address) (*big.Int, error) {
	return _HopBridge.Contract.GetRawDebit(&_HopBridge.CallOpts, bonder)
}

// GetRawDebit is a free data retrieval call binding the contract method 0x13948c76.
//
// Solidity: function getRawDebit(address bonder) view returns(uint256)
func (_HopBridge *HopBridgeCallerSession) GetRawDebit(bonder common.Address) (*big.Int, error) {
	return _HopBridge.Contract.GetRawDebit(&_HopBridge.CallOpts, bonder)
}

// GetTimeSlot is a free data retrieval call binding the contract method 0x2b85dcc9.
//
// Solidity: function getTimeSlot(uint256 time) pure returns(uint256)
func (_HopBridge *HopBridgeCaller) GetTimeSlot(opts *bind.CallOpts, time *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "getTimeSlot", time)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTimeSlot is a free data retrieval call binding the contract method 0x2b85dcc9.
//
// Solidity: function getTimeSlot(uint256 time) pure returns(uint256)
func (_HopBridge *HopBridgeSession) GetTimeSlot(time *big.Int) (*big.Int, error) {
	return _HopBridge.Contract.GetTimeSlot(&_HopBridge.CallOpts, time)
}

// GetTimeSlot is a free data retrieval call binding the contract method 0x2b85dcc9.
//
// Solidity: function getTimeSlot(uint256 time) pure returns(uint256)
func (_HopBridge *HopBridgeCallerSession) GetTimeSlot(time *big.Int) (*big.Int, error) {
	return _HopBridge.Contract.GetTimeSlot(&_HopBridge.CallOpts, time)
}

// GetTransferId is a free data retrieval call binding the contract method 0xaf215f94.
//
// Solidity: function getTransferId(uint256 chainId, address recipient, uint256 amount, bytes32 transferNonce, uint256 bonderFee, uint256 amountOutMin, uint256 deadline) pure returns(bytes32)
func (_HopBridge *HopBridgeCaller) GetTransferId(opts *bind.CallOpts, chainId *big.Int, recipient common.Address, amount *big.Int, transferNonce [32]byte, bonderFee *big.Int, amountOutMin *big.Int, deadline *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "getTransferId", chainId, recipient, amount, transferNonce, bonderFee, amountOutMin, deadline)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetTransferId is a free data retrieval call binding the contract method 0xaf215f94.
//
// Solidity: function getTransferId(uint256 chainId, address recipient, uint256 amount, bytes32 transferNonce, uint256 bonderFee, uint256 amountOutMin, uint256 deadline) pure returns(bytes32)
func (_HopBridge *HopBridgeSession) GetTransferId(chainId *big.Int, recipient common.Address, amount *big.Int, transferNonce [32]byte, bonderFee *big.Int, amountOutMin *big.Int, deadline *big.Int) ([32]byte, error) {
	return _HopBridge.Contract.GetTransferId(&_HopBridge.CallOpts, chainId, recipient, amount, transferNonce, bonderFee, amountOutMin, deadline)
}

// GetTransferId is a free data retrieval call binding the contract method 0xaf215f94.
//
// Solidity: function getTransferId(uint256 chainId, address recipient, uint256 amount, bytes32 transferNonce, uint256 bonderFee, uint256 amountOutMin, uint256 deadline) pure returns(bytes32)
func (_HopBridge *HopBridgeCallerSession) GetTransferId(chainId *big.Int, recipient common.Address, amount *big.Int, transferNonce [32]byte, bonderFee *big.Int, amountOutMin *big.Int, deadline *big.Int) ([32]byte, error) {
	return _HopBridge.Contract.GetTransferId(&_HopBridge.CallOpts, chainId, recipient, amount, transferNonce, bonderFee, amountOutMin, deadline)
}

// GetTransferRoot is a free data retrieval call binding the contract method 0xce803b4f.
//
// Solidity: function getTransferRoot(bytes32 rootHash, uint256 totalAmount) view returns((uint256,uint256,uint256))
func (_HopBridge *HopBridgeCaller) GetTransferRoot(opts *bind.CallOpts, rootHash [32]byte, totalAmount *big.Int) (BridgeTransferRoot, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "getTransferRoot", rootHash, totalAmount)

	if err != nil {
		return *new(BridgeTransferRoot), err
	}

	out0 := *abi.ConvertType(out[0], new(BridgeTransferRoot)).(*BridgeTransferRoot)

	return out0, err

}

// GetTransferRoot is a free data retrieval call binding the contract method 0xce803b4f.
//
// Solidity: function getTransferRoot(bytes32 rootHash, uint256 totalAmount) view returns((uint256,uint256,uint256))
func (_HopBridge *HopBridgeSession) GetTransferRoot(rootHash [32]byte, totalAmount *big.Int) (BridgeTransferRoot, error) {
	return _HopBridge.Contract.GetTransferRoot(&_HopBridge.CallOpts, rootHash, totalAmount)
}

// GetTransferRoot is a free data retrieval call binding the contract method 0xce803b4f.
//
// Solidity: function getTransferRoot(bytes32 rootHash, uint256 totalAmount) view returns((uint256,uint256,uint256))
func (_HopBridge *HopBridgeCallerSession) GetTransferRoot(rootHash [32]byte, totalAmount *big.Int) (BridgeTransferRoot, error) {
	return _HopBridge.Contract.GetTransferRoot(&_HopBridge.CallOpts, rootHash, totalAmount)
}

// GetTransferRootId is a free data retrieval call binding the contract method 0x960a7afa.
//
// Solidity: function getTransferRootId(bytes32 rootHash, uint256 totalAmount) pure returns(bytes32)
func (_HopBridge *HopBridgeCaller) GetTransferRootId(opts *bind.CallOpts, rootHash [32]byte, totalAmount *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "getTransferRootId", rootHash, totalAmount)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetTransferRootId is a free data retrieval call binding the contract method 0x960a7afa.
//
// Solidity: function getTransferRootId(bytes32 rootHash, uint256 totalAmount) pure returns(bytes32)
func (_HopBridge *HopBridgeSession) GetTransferRootId(rootHash [32]byte, totalAmount *big.Int) ([32]byte, error) {
	return _HopBridge.Contract.GetTransferRootId(&_HopBridge.CallOpts, rootHash, totalAmount)
}

// GetTransferRootId is a free data retrieval call binding the contract method 0x960a7afa.
//
// Solidity: function getTransferRootId(bytes32 rootHash, uint256 totalAmount) pure returns(bytes32)
func (_HopBridge *HopBridgeCallerSession) GetTransferRootId(rootHash [32]byte, totalAmount *big.Int) ([32]byte, error) {
	return _HopBridge.Contract.GetTransferRootId(&_HopBridge.CallOpts, rootHash, totalAmount)
}

// Governance is a free data retrieval call binding the contract method 0x5aa6e675.
//
// Solidity: function governance() view returns(address)
func (_HopBridge *HopBridgeCaller) Governance(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "governance")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Governance is a free data retrieval call binding the contract method 0x5aa6e675.
//
// Solidity: function governance() view returns(address)
func (_HopBridge *HopBridgeSession) Governance() (common.Address, error) {
	return _HopBridge.Contract.Governance(&_HopBridge.CallOpts)
}

// Governance is a free data retrieval call binding the contract method 0x5aa6e675.
//
// Solidity: function governance() view returns(address)
func (_HopBridge *HopBridgeCallerSession) Governance() (common.Address, error) {
	return _HopBridge.Contract.Governance(&_HopBridge.CallOpts)
}

// IsChainIdPaused is a free data retrieval call binding the contract method 0xfa2a69a3.
//
// Solidity: function isChainIdPaused(uint256 ) view returns(bool)
func (_HopBridge *HopBridgeCaller) IsChainIdPaused(opts *bind.CallOpts, arg0 *big.Int) (bool, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "isChainIdPaused", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsChainIdPaused is a free data retrieval call binding the contract method 0xfa2a69a3.
//
// Solidity: function isChainIdPaused(uint256 ) view returns(bool)
func (_HopBridge *HopBridgeSession) IsChainIdPaused(arg0 *big.Int) (bool, error) {
	return _HopBridge.Contract.IsChainIdPaused(&_HopBridge.CallOpts, arg0)
}

// IsChainIdPaused is a free data retrieval call binding the contract method 0xfa2a69a3.
//
// Solidity: function isChainIdPaused(uint256 ) view returns(bool)
func (_HopBridge *HopBridgeCallerSession) IsChainIdPaused(arg0 *big.Int) (bool, error) {
	return _HopBridge.Contract.IsChainIdPaused(&_HopBridge.CallOpts, arg0)
}

// IsTransferIdSpent is a free data retrieval call binding the contract method 0x3a7af631.
//
// Solidity: function isTransferIdSpent(bytes32 transferId) view returns(bool)
func (_HopBridge *HopBridgeCaller) IsTransferIdSpent(opts *bind.CallOpts, transferId [32]byte) (bool, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "isTransferIdSpent", transferId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsTransferIdSpent is a free data retrieval call binding the contract method 0x3a7af631.
//
// Solidity: function isTransferIdSpent(bytes32 transferId) view returns(bool)
func (_HopBridge *HopBridgeSession) IsTransferIdSpent(transferId [32]byte) (bool, error) {
	return _HopBridge.Contract.IsTransferIdSpent(&_HopBridge.CallOpts, transferId)
}

// IsTransferIdSpent is a free data retrieval call binding the contract method 0x3a7af631.
//
// Solidity: function isTransferIdSpent(bytes32 transferId) view returns(bool)
func (_HopBridge *HopBridgeCallerSession) IsTransferIdSpent(transferId [32]byte) (bool, error) {
	return _HopBridge.Contract.IsTransferIdSpent(&_HopBridge.CallOpts, transferId)
}

// L1CanonicalToken is a free data retrieval call binding the contract method 0xb7a0bda6.
//
// Solidity: function l1CanonicalToken() view returns(address)
func (_HopBridge *HopBridgeCaller) L1CanonicalToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "l1CanonicalToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L1CanonicalToken is a free data retrieval call binding the contract method 0xb7a0bda6.
//
// Solidity: function l1CanonicalToken() view returns(address)
func (_HopBridge *HopBridgeSession) L1CanonicalToken() (common.Address, error) {
	return _HopBridge.Contract.L1CanonicalToken(&_HopBridge.CallOpts)
}

// L1CanonicalToken is a free data retrieval call binding the contract method 0xb7a0bda6.
//
// Solidity: function l1CanonicalToken() view returns(address)
func (_HopBridge *HopBridgeCallerSession) L1CanonicalToken() (common.Address, error) {
	return _HopBridge.Contract.L1CanonicalToken(&_HopBridge.CallOpts)
}

// MinTransferRootBondDelay is a free data retrieval call binding the contract method 0x6cff06a7.
//
// Solidity: function minTransferRootBondDelay() view returns(uint256)
func (_HopBridge *HopBridgeCaller) MinTransferRootBondDelay(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "minTransferRootBondDelay")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinTransferRootBondDelay is a free data retrieval call binding the contract method 0x6cff06a7.
//
// Solidity: function minTransferRootBondDelay() view returns(uint256)
func (_HopBridge *HopBridgeSession) MinTransferRootBondDelay() (*big.Int, error) {
	return _HopBridge.Contract.MinTransferRootBondDelay(&_HopBridge.CallOpts)
}

// MinTransferRootBondDelay is a free data retrieval call binding the contract method 0x6cff06a7.
//
// Solidity: function minTransferRootBondDelay() view returns(uint256)
func (_HopBridge *HopBridgeCallerSession) MinTransferRootBondDelay() (*big.Int, error) {
	return _HopBridge.Contract.MinTransferRootBondDelay(&_HopBridge.CallOpts)
}

// TimeSlotToAmountBonded is a free data retrieval call binding the contract method 0x7398d282.
//
// Solidity: function timeSlotToAmountBonded(uint256 , address ) view returns(uint256)
func (_HopBridge *HopBridgeCaller) TimeSlotToAmountBonded(opts *bind.CallOpts, arg0 *big.Int, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "timeSlotToAmountBonded", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TimeSlotToAmountBonded is a free data retrieval call binding the contract method 0x7398d282.
//
// Solidity: function timeSlotToAmountBonded(uint256 , address ) view returns(uint256)
func (_HopBridge *HopBridgeSession) TimeSlotToAmountBonded(arg0 *big.Int, arg1 common.Address) (*big.Int, error) {
	return _HopBridge.Contract.TimeSlotToAmountBonded(&_HopBridge.CallOpts, arg0, arg1)
}

// TimeSlotToAmountBonded is a free data retrieval call binding the contract method 0x7398d282.
//
// Solidity: function timeSlotToAmountBonded(uint256 , address ) view returns(uint256)
func (_HopBridge *HopBridgeCallerSession) TimeSlotToAmountBonded(arg0 *big.Int, arg1 common.Address) (*big.Int, error) {
	return _HopBridge.Contract.TimeSlotToAmountBonded(&_HopBridge.CallOpts, arg0, arg1)
}

// TransferBonds is a free data retrieval call binding the contract method 0x5a7e1083.
//
// Solidity: function transferBonds(bytes32 ) view returns(address bonder, uint256 createdAt, uint256 totalAmount, uint256 challengeStartTime, address challenger, bool challengeResolved)
func (_HopBridge *HopBridgeCaller) TransferBonds(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Bonder             common.Address
	CreatedAt          *big.Int
	TotalAmount        *big.Int
	ChallengeStartTime *big.Int
	Challenger         common.Address
	ChallengeResolved  bool
}, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "transferBonds", arg0)

	outstruct := new(struct {
		Bonder             common.Address
		CreatedAt          *big.Int
		TotalAmount        *big.Int
		ChallengeStartTime *big.Int
		Challenger         common.Address
		ChallengeResolved  bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Bonder = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.CreatedAt = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.TotalAmount = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.ChallengeStartTime = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.Challenger = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.ChallengeResolved = *abi.ConvertType(out[5], new(bool)).(*bool)

	return *outstruct, err

}

// TransferBonds is a free data retrieval call binding the contract method 0x5a7e1083.
//
// Solidity: function transferBonds(bytes32 ) view returns(address bonder, uint256 createdAt, uint256 totalAmount, uint256 challengeStartTime, address challenger, bool challengeResolved)
func (_HopBridge *HopBridgeSession) TransferBonds(arg0 [32]byte) (struct {
	Bonder             common.Address
	CreatedAt          *big.Int
	TotalAmount        *big.Int
	ChallengeStartTime *big.Int
	Challenger         common.Address
	ChallengeResolved  bool
}, error) {
	return _HopBridge.Contract.TransferBonds(&_HopBridge.CallOpts, arg0)
}

// TransferBonds is a free data retrieval call binding the contract method 0x5a7e1083.
//
// Solidity: function transferBonds(bytes32 ) view returns(address bonder, uint256 createdAt, uint256 totalAmount, uint256 challengeStartTime, address challenger, bool challengeResolved)
func (_HopBridge *HopBridgeCallerSession) TransferBonds(arg0 [32]byte) (struct {
	Bonder             common.Address
	CreatedAt          *big.Int
	TotalAmount        *big.Int
	ChallengeStartTime *big.Int
	Challenger         common.Address
	ChallengeResolved  bool
}, error) {
	return _HopBridge.Contract.TransferBonds(&_HopBridge.CallOpts, arg0)
}

// TransferRootCommittedAt is a free data retrieval call binding the contract method 0x4612f40c.
//
// Solidity: function transferRootCommittedAt(bytes32 ) view returns(uint256)
func (_HopBridge *HopBridgeCaller) TransferRootCommittedAt(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _HopBridge.contract.Call(opts, &out, "transferRootCommittedAt", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TransferRootCommittedAt is a free data retrieval call binding the contract method 0x4612f40c.
//
// Solidity: function transferRootCommittedAt(bytes32 ) view returns(uint256)
func (_HopBridge *HopBridgeSession) TransferRootCommittedAt(arg0 [32]byte) (*big.Int, error) {
	return _HopBridge.Contract.TransferRootCommittedAt(&_HopBridge.CallOpts, arg0)
}

// TransferRootCommittedAt is a free data retrieval call binding the contract method 0x4612f40c.
//
// Solidity: function transferRootCommittedAt(bytes32 ) view returns(uint256)
func (_HopBridge *HopBridgeCallerSession) TransferRootCommittedAt(arg0 [32]byte) (*big.Int, error) {
	return _HopBridge.Contract.TransferRootCommittedAt(&_HopBridge.CallOpts, arg0)
}

// AddBonder is a paid mutator transaction binding the contract method 0x5325937f.
//
// Solidity: function addBonder(address bonder) returns()
func (_HopBridge *HopBridgeTransactor) AddBonder(opts *bind.TransactOpts, bonder common.Address) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "addBonder", bonder)
}

// AddBonder is a paid mutator transaction binding the contract method 0x5325937f.
//
// Solidity: function addBonder(address bonder) returns()
func (_HopBridge *HopBridgeSession) AddBonder(bonder common.Address) (*types.Transaction, error) {
	return _HopBridge.Contract.AddBonder(&_HopBridge.TransactOpts, bonder)
}

// AddBonder is a paid mutator transaction binding the contract method 0x5325937f.
//
// Solidity: function addBonder(address bonder) returns()
func (_HopBridge *HopBridgeTransactorSession) AddBonder(bonder common.Address) (*types.Transaction, error) {
	return _HopBridge.Contract.AddBonder(&_HopBridge.TransactOpts, bonder)
}

// BondTransferRoot is a paid mutator transaction binding the contract method 0x8d8798bf.
//
// Solidity: function bondTransferRoot(bytes32 rootHash, uint256 destinationChainId, uint256 totalAmount) returns()
func (_HopBridge *HopBridgeTransactor) BondTransferRoot(opts *bind.TransactOpts, rootHash [32]byte, destinationChainId *big.Int, totalAmount *big.Int) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "bondTransferRoot", rootHash, destinationChainId, totalAmount)
}

// BondTransferRoot is a paid mutator transaction binding the contract method 0x8d8798bf.
//
// Solidity: function bondTransferRoot(bytes32 rootHash, uint256 destinationChainId, uint256 totalAmount) returns()
func (_HopBridge *HopBridgeSession) BondTransferRoot(rootHash [32]byte, destinationChainId *big.Int, totalAmount *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.BondTransferRoot(&_HopBridge.TransactOpts, rootHash, destinationChainId, totalAmount)
}

// BondTransferRoot is a paid mutator transaction binding the contract method 0x8d8798bf.
//
// Solidity: function bondTransferRoot(bytes32 rootHash, uint256 destinationChainId, uint256 totalAmount) returns()
func (_HopBridge *HopBridgeTransactorSession) BondTransferRoot(rootHash [32]byte, destinationChainId *big.Int, totalAmount *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.BondTransferRoot(&_HopBridge.TransactOpts, rootHash, destinationChainId, totalAmount)
}

// BondWithdrawal is a paid mutator transaction binding the contract method 0x23c452cd.
//
// Solidity: function bondWithdrawal(address recipient, uint256 amount, bytes32 transferNonce, uint256 bonderFee) returns()
func (_HopBridge *HopBridgeTransactor) BondWithdrawal(opts *bind.TransactOpts, recipient common.Address, amount *big.Int, transferNonce [32]byte, bonderFee *big.Int) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "bondWithdrawal", recipient, amount, transferNonce, bonderFee)
}

// BondWithdrawal is a paid mutator transaction binding the contract method 0x23c452cd.
//
// Solidity: function bondWithdrawal(address recipient, uint256 amount, bytes32 transferNonce, uint256 bonderFee) returns()
func (_HopBridge *HopBridgeSession) BondWithdrawal(recipient common.Address, amount *big.Int, transferNonce [32]byte, bonderFee *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.BondWithdrawal(&_HopBridge.TransactOpts, recipient, amount, transferNonce, bonderFee)
}

// BondWithdrawal is a paid mutator transaction binding the contract method 0x23c452cd.
//
// Solidity: function bondWithdrawal(address recipient, uint256 amount, bytes32 transferNonce, uint256 bonderFee) returns()
func (_HopBridge *HopBridgeTransactorSession) BondWithdrawal(recipient common.Address, amount *big.Int, transferNonce [32]byte, bonderFee *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.BondWithdrawal(&_HopBridge.TransactOpts, recipient, amount, transferNonce, bonderFee)
}

// ChallengeTransferBond is a paid mutator transaction binding the contract method 0xbacc68af.
//
// Solidity: function challengeTransferBond(bytes32 rootHash, uint256 originalAmount) payable returns()
func (_HopBridge *HopBridgeTransactor) ChallengeTransferBond(opts *bind.TransactOpts, rootHash [32]byte, originalAmount *big.Int) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "challengeTransferBond", rootHash, originalAmount)
}

// ChallengeTransferBond is a paid mutator transaction binding the contract method 0xbacc68af.
//
// Solidity: function challengeTransferBond(bytes32 rootHash, uint256 originalAmount) payable returns()
func (_HopBridge *HopBridgeSession) ChallengeTransferBond(rootHash [32]byte, originalAmount *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.ChallengeTransferBond(&_HopBridge.TransactOpts, rootHash, originalAmount)
}

// ChallengeTransferBond is a paid mutator transaction binding the contract method 0xbacc68af.
//
// Solidity: function challengeTransferBond(bytes32 rootHash, uint256 originalAmount) payable returns()
func (_HopBridge *HopBridgeTransactorSession) ChallengeTransferBond(rootHash [32]byte, originalAmount *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.ChallengeTransferBond(&_HopBridge.TransactOpts, rootHash, originalAmount)
}

// ConfirmTransferRoot is a paid mutator transaction binding the contract method 0xef6ebe5e.
//
// Solidity: function confirmTransferRoot(uint256 originChainId, bytes32 rootHash, uint256 destinationChainId, uint256 totalAmount, uint256 rootCommittedAt) returns()
func (_HopBridge *HopBridgeTransactor) ConfirmTransferRoot(opts *bind.TransactOpts, originChainId *big.Int, rootHash [32]byte, destinationChainId *big.Int, totalAmount *big.Int, rootCommittedAt *big.Int) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "confirmTransferRoot", originChainId, rootHash, destinationChainId, totalAmount, rootCommittedAt)
}

// ConfirmTransferRoot is a paid mutator transaction binding the contract method 0xef6ebe5e.
//
// Solidity: function confirmTransferRoot(uint256 originChainId, bytes32 rootHash, uint256 destinationChainId, uint256 totalAmount, uint256 rootCommittedAt) returns()
func (_HopBridge *HopBridgeSession) ConfirmTransferRoot(originChainId *big.Int, rootHash [32]byte, destinationChainId *big.Int, totalAmount *big.Int, rootCommittedAt *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.ConfirmTransferRoot(&_HopBridge.TransactOpts, originChainId, rootHash, destinationChainId, totalAmount, rootCommittedAt)
}

// ConfirmTransferRoot is a paid mutator transaction binding the contract method 0xef6ebe5e.
//
// Solidity: function confirmTransferRoot(uint256 originChainId, bytes32 rootHash, uint256 destinationChainId, uint256 totalAmount, uint256 rootCommittedAt) returns()
func (_HopBridge *HopBridgeTransactorSession) ConfirmTransferRoot(originChainId *big.Int, rootHash [32]byte, destinationChainId *big.Int, totalAmount *big.Int, rootCommittedAt *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.ConfirmTransferRoot(&_HopBridge.TransactOpts, originChainId, rootHash, destinationChainId, totalAmount, rootCommittedAt)
}

// RemoveBonder is a paid mutator transaction binding the contract method 0x04e6c2c0.
//
// Solidity: function removeBonder(address bonder) returns()
func (_HopBridge *HopBridgeTransactor) RemoveBonder(opts *bind.TransactOpts, bonder common.Address) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "removeBonder", bonder)
}

// RemoveBonder is a paid mutator transaction binding the contract method 0x04e6c2c0.
//
// Solidity: function removeBonder(address bonder) returns()
func (_HopBridge *HopBridgeSession) RemoveBonder(bonder common.Address) (*types.Transaction, error) {
	return _HopBridge.Contract.RemoveBonder(&_HopBridge.TransactOpts, bonder)
}

// RemoveBonder is a paid mutator transaction binding the contract method 0x04e6c2c0.
//
// Solidity: function removeBonder(address bonder) returns()
func (_HopBridge *HopBridgeTransactorSession) RemoveBonder(bonder common.Address) (*types.Transaction, error) {
	return _HopBridge.Contract.RemoveBonder(&_HopBridge.TransactOpts, bonder)
}

// RescueTransferRoot is a paid mutator transaction binding the contract method 0xcbd1642e.
//
// Solidity: function rescueTransferRoot(bytes32 rootHash, uint256 originalAmount, address recipient) returns()
func (_HopBridge *HopBridgeTransactor) RescueTransferRoot(opts *bind.TransactOpts, rootHash [32]byte, originalAmount *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "rescueTransferRoot", rootHash, originalAmount, recipient)
}

// RescueTransferRoot is a paid mutator transaction binding the contract method 0xcbd1642e.
//
// Solidity: function rescueTransferRoot(bytes32 rootHash, uint256 originalAmount, address recipient) returns()
func (_HopBridge *HopBridgeSession) RescueTransferRoot(rootHash [32]byte, originalAmount *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _HopBridge.Contract.RescueTransferRoot(&_HopBridge.TransactOpts, rootHash, originalAmount, recipient)
}

// RescueTransferRoot is a paid mutator transaction binding the contract method 0xcbd1642e.
//
// Solidity: function rescueTransferRoot(bytes32 rootHash, uint256 originalAmount, address recipient) returns()
func (_HopBridge *HopBridgeTransactorSession) RescueTransferRoot(rootHash [32]byte, originalAmount *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _HopBridge.Contract.RescueTransferRoot(&_HopBridge.TransactOpts, rootHash, originalAmount, recipient)
}

// ResolveChallenge is a paid mutator transaction binding the contract method 0x45ca9fc9.
//
// Solidity: function resolveChallenge(bytes32 rootHash, uint256 originalAmount) returns()
func (_HopBridge *HopBridgeTransactor) ResolveChallenge(opts *bind.TransactOpts, rootHash [32]byte, originalAmount *big.Int) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "resolveChallenge", rootHash, originalAmount)
}

// ResolveChallenge is a paid mutator transaction binding the contract method 0x45ca9fc9.
//
// Solidity: function resolveChallenge(bytes32 rootHash, uint256 originalAmount) returns()
func (_HopBridge *HopBridgeSession) ResolveChallenge(rootHash [32]byte, originalAmount *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.ResolveChallenge(&_HopBridge.TransactOpts, rootHash, originalAmount)
}

// ResolveChallenge is a paid mutator transaction binding the contract method 0x45ca9fc9.
//
// Solidity: function resolveChallenge(bytes32 rootHash, uint256 originalAmount) returns()
func (_HopBridge *HopBridgeTransactorSession) ResolveChallenge(rootHash [32]byte, originalAmount *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.ResolveChallenge(&_HopBridge.TransactOpts, rootHash, originalAmount)
}

// SendToL2 is a paid mutator transaction binding the contract method 0xdeace8f5.
//
// Solidity: function sendToL2(uint256 chainId, address recipient, uint256 amount, uint256 amountOutMin, uint256 deadline, address relayer, uint256 relayerFee) payable returns()
func (_HopBridge *HopBridgeTransactor) SendToL2(opts *bind.TransactOpts, chainId *big.Int, recipient common.Address, amount *big.Int, amountOutMin *big.Int, deadline *big.Int, relayer common.Address, relayerFee *big.Int) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "sendToL2", chainId, recipient, amount, amountOutMin, deadline, relayer, relayerFee)
}

// SendToL2 is a paid mutator transaction binding the contract method 0xdeace8f5.
//
// Solidity: function sendToL2(uint256 chainId, address recipient, uint256 amount, uint256 amountOutMin, uint256 deadline, address relayer, uint256 relayerFee) payable returns()
func (_HopBridge *HopBridgeSession) SendToL2(chainId *big.Int, recipient common.Address, amount *big.Int, amountOutMin *big.Int, deadline *big.Int, relayer common.Address, relayerFee *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.SendToL2(&_HopBridge.TransactOpts, chainId, recipient, amount, amountOutMin, deadline, relayer, relayerFee)
}

// SendToL2 is a paid mutator transaction binding the contract method 0xdeace8f5.
//
// Solidity: function sendToL2(uint256 chainId, address recipient, uint256 amount, uint256 amountOutMin, uint256 deadline, address relayer, uint256 relayerFee) payable returns()
func (_HopBridge *HopBridgeTransactorSession) SendToL2(chainId *big.Int, recipient common.Address, amount *big.Int, amountOutMin *big.Int, deadline *big.Int, relayer common.Address, relayerFee *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.SendToL2(&_HopBridge.TransactOpts, chainId, recipient, amount, amountOutMin, deadline, relayer, relayerFee)
}

// SetChainIdDepositsPaused is a paid mutator transaction binding the contract method 0x14942024.
//
// Solidity: function setChainIdDepositsPaused(uint256 chainId, bool isPaused) returns()
func (_HopBridge *HopBridgeTransactor) SetChainIdDepositsPaused(opts *bind.TransactOpts, chainId *big.Int, isPaused bool) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "setChainIdDepositsPaused", chainId, isPaused)
}

// SetChainIdDepositsPaused is a paid mutator transaction binding the contract method 0x14942024.
//
// Solidity: function setChainIdDepositsPaused(uint256 chainId, bool isPaused) returns()
func (_HopBridge *HopBridgeSession) SetChainIdDepositsPaused(chainId *big.Int, isPaused bool) (*types.Transaction, error) {
	return _HopBridge.Contract.SetChainIdDepositsPaused(&_HopBridge.TransactOpts, chainId, isPaused)
}

// SetChainIdDepositsPaused is a paid mutator transaction binding the contract method 0x14942024.
//
// Solidity: function setChainIdDepositsPaused(uint256 chainId, bool isPaused) returns()
func (_HopBridge *HopBridgeTransactorSession) SetChainIdDepositsPaused(chainId *big.Int, isPaused bool) (*types.Transaction, error) {
	return _HopBridge.Contract.SetChainIdDepositsPaused(&_HopBridge.TransactOpts, chainId, isPaused)
}

// SetChallengePeriod is a paid mutator transaction binding the contract method 0x5d475fdd.
//
// Solidity: function setChallengePeriod(uint256 _challengePeriod) returns()
func (_HopBridge *HopBridgeTransactor) SetChallengePeriod(opts *bind.TransactOpts, _challengePeriod *big.Int) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "setChallengePeriod", _challengePeriod)
}

// SetChallengePeriod is a paid mutator transaction binding the contract method 0x5d475fdd.
//
// Solidity: function setChallengePeriod(uint256 _challengePeriod) returns()
func (_HopBridge *HopBridgeSession) SetChallengePeriod(_challengePeriod *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.SetChallengePeriod(&_HopBridge.TransactOpts, _challengePeriod)
}

// SetChallengePeriod is a paid mutator transaction binding the contract method 0x5d475fdd.
//
// Solidity: function setChallengePeriod(uint256 _challengePeriod) returns()
func (_HopBridge *HopBridgeTransactorSession) SetChallengePeriod(_challengePeriod *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.SetChallengePeriod(&_HopBridge.TransactOpts, _challengePeriod)
}

// SetChallengeResolutionPeriod is a paid mutator transaction binding the contract method 0xeecd57e6.
//
// Solidity: function setChallengeResolutionPeriod(uint256 _challengeResolutionPeriod) returns()
func (_HopBridge *HopBridgeTransactor) SetChallengeResolutionPeriod(opts *bind.TransactOpts, _challengeResolutionPeriod *big.Int) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "setChallengeResolutionPeriod", _challengeResolutionPeriod)
}

// SetChallengeResolutionPeriod is a paid mutator transaction binding the contract method 0xeecd57e6.
//
// Solidity: function setChallengeResolutionPeriod(uint256 _challengeResolutionPeriod) returns()
func (_HopBridge *HopBridgeSession) SetChallengeResolutionPeriod(_challengeResolutionPeriod *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.SetChallengeResolutionPeriod(&_HopBridge.TransactOpts, _challengeResolutionPeriod)
}

// SetChallengeResolutionPeriod is a paid mutator transaction binding the contract method 0xeecd57e6.
//
// Solidity: function setChallengeResolutionPeriod(uint256 _challengeResolutionPeriod) returns()
func (_HopBridge *HopBridgeTransactorSession) SetChallengeResolutionPeriod(_challengeResolutionPeriod *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.SetChallengeResolutionPeriod(&_HopBridge.TransactOpts, _challengeResolutionPeriod)
}

// SetCrossDomainMessengerWrapper is a paid mutator transaction binding the contract method 0xd4448163.
//
// Solidity: function setCrossDomainMessengerWrapper(uint256 chainId, address _crossDomainMessengerWrapper) returns()
func (_HopBridge *HopBridgeTransactor) SetCrossDomainMessengerWrapper(opts *bind.TransactOpts, chainId *big.Int, _crossDomainMessengerWrapper common.Address) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "setCrossDomainMessengerWrapper", chainId, _crossDomainMessengerWrapper)
}

// SetCrossDomainMessengerWrapper is a paid mutator transaction binding the contract method 0xd4448163.
//
// Solidity: function setCrossDomainMessengerWrapper(uint256 chainId, address _crossDomainMessengerWrapper) returns()
func (_HopBridge *HopBridgeSession) SetCrossDomainMessengerWrapper(chainId *big.Int, _crossDomainMessengerWrapper common.Address) (*types.Transaction, error) {
	return _HopBridge.Contract.SetCrossDomainMessengerWrapper(&_HopBridge.TransactOpts, chainId, _crossDomainMessengerWrapper)
}

// SetCrossDomainMessengerWrapper is a paid mutator transaction binding the contract method 0xd4448163.
//
// Solidity: function setCrossDomainMessengerWrapper(uint256 chainId, address _crossDomainMessengerWrapper) returns()
func (_HopBridge *HopBridgeTransactorSession) SetCrossDomainMessengerWrapper(chainId *big.Int, _crossDomainMessengerWrapper common.Address) (*types.Transaction, error) {
	return _HopBridge.Contract.SetCrossDomainMessengerWrapper(&_HopBridge.TransactOpts, chainId, _crossDomainMessengerWrapper)
}

// SetGovernance is a paid mutator transaction binding the contract method 0xab033ea9.
//
// Solidity: function setGovernance(address _newGovernance) returns()
func (_HopBridge *HopBridgeTransactor) SetGovernance(opts *bind.TransactOpts, _newGovernance common.Address) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "setGovernance", _newGovernance)
}

// SetGovernance is a paid mutator transaction binding the contract method 0xab033ea9.
//
// Solidity: function setGovernance(address _newGovernance) returns()
func (_HopBridge *HopBridgeSession) SetGovernance(_newGovernance common.Address) (*types.Transaction, error) {
	return _HopBridge.Contract.SetGovernance(&_HopBridge.TransactOpts, _newGovernance)
}

// SetGovernance is a paid mutator transaction binding the contract method 0xab033ea9.
//
// Solidity: function setGovernance(address _newGovernance) returns()
func (_HopBridge *HopBridgeTransactorSession) SetGovernance(_newGovernance common.Address) (*types.Transaction, error) {
	return _HopBridge.Contract.SetGovernance(&_HopBridge.TransactOpts, _newGovernance)
}

// SetMinTransferRootBondDelay is a paid mutator transaction binding the contract method 0x39ada669.
//
// Solidity: function setMinTransferRootBondDelay(uint256 _minTransferRootBondDelay) returns()
func (_HopBridge *HopBridgeTransactor) SetMinTransferRootBondDelay(opts *bind.TransactOpts, _minTransferRootBondDelay *big.Int) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "setMinTransferRootBondDelay", _minTransferRootBondDelay)
}

// SetMinTransferRootBondDelay is a paid mutator transaction binding the contract method 0x39ada669.
//
// Solidity: function setMinTransferRootBondDelay(uint256 _minTransferRootBondDelay) returns()
func (_HopBridge *HopBridgeSession) SetMinTransferRootBondDelay(_minTransferRootBondDelay *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.SetMinTransferRootBondDelay(&_HopBridge.TransactOpts, _minTransferRootBondDelay)
}

// SetMinTransferRootBondDelay is a paid mutator transaction binding the contract method 0x39ada669.
//
// Solidity: function setMinTransferRootBondDelay(uint256 _minTransferRootBondDelay) returns()
func (_HopBridge *HopBridgeTransactorSession) SetMinTransferRootBondDelay(_minTransferRootBondDelay *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.SetMinTransferRootBondDelay(&_HopBridge.TransactOpts, _minTransferRootBondDelay)
}

// SettleBondedWithdrawal is a paid mutator transaction binding the contract method 0xc7525dd3.
//
// Solidity: function settleBondedWithdrawal(address bonder, bytes32 transferId, bytes32 rootHash, uint256 transferRootTotalAmount, uint256 transferIdTreeIndex, bytes32[] siblings, uint256 totalLeaves) returns()
func (_HopBridge *HopBridgeTransactor) SettleBondedWithdrawal(opts *bind.TransactOpts, bonder common.Address, transferId [32]byte, rootHash [32]byte, transferRootTotalAmount *big.Int, transferIdTreeIndex *big.Int, siblings [][32]byte, totalLeaves *big.Int) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "settleBondedWithdrawal", bonder, transferId, rootHash, transferRootTotalAmount, transferIdTreeIndex, siblings, totalLeaves)
}

// SettleBondedWithdrawal is a paid mutator transaction binding the contract method 0xc7525dd3.
//
// Solidity: function settleBondedWithdrawal(address bonder, bytes32 transferId, bytes32 rootHash, uint256 transferRootTotalAmount, uint256 transferIdTreeIndex, bytes32[] siblings, uint256 totalLeaves) returns()
func (_HopBridge *HopBridgeSession) SettleBondedWithdrawal(bonder common.Address, transferId [32]byte, rootHash [32]byte, transferRootTotalAmount *big.Int, transferIdTreeIndex *big.Int, siblings [][32]byte, totalLeaves *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.SettleBondedWithdrawal(&_HopBridge.TransactOpts, bonder, transferId, rootHash, transferRootTotalAmount, transferIdTreeIndex, siblings, totalLeaves)
}

// SettleBondedWithdrawal is a paid mutator transaction binding the contract method 0xc7525dd3.
//
// Solidity: function settleBondedWithdrawal(address bonder, bytes32 transferId, bytes32 rootHash, uint256 transferRootTotalAmount, uint256 transferIdTreeIndex, bytes32[] siblings, uint256 totalLeaves) returns()
func (_HopBridge *HopBridgeTransactorSession) SettleBondedWithdrawal(bonder common.Address, transferId [32]byte, rootHash [32]byte, transferRootTotalAmount *big.Int, transferIdTreeIndex *big.Int, siblings [][32]byte, totalLeaves *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.SettleBondedWithdrawal(&_HopBridge.TransactOpts, bonder, transferId, rootHash, transferRootTotalAmount, transferIdTreeIndex, siblings, totalLeaves)
}

// SettleBondedWithdrawals is a paid mutator transaction binding the contract method 0xb162717e.
//
// Solidity: function settleBondedWithdrawals(address bonder, bytes32[] transferIds, uint256 totalAmount) returns()
func (_HopBridge *HopBridgeTransactor) SettleBondedWithdrawals(opts *bind.TransactOpts, bonder common.Address, transferIds [][32]byte, totalAmount *big.Int) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "settleBondedWithdrawals", bonder, transferIds, totalAmount)
}

// SettleBondedWithdrawals is a paid mutator transaction binding the contract method 0xb162717e.
//
// Solidity: function settleBondedWithdrawals(address bonder, bytes32[] transferIds, uint256 totalAmount) returns()
func (_HopBridge *HopBridgeSession) SettleBondedWithdrawals(bonder common.Address, transferIds [][32]byte, totalAmount *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.SettleBondedWithdrawals(&_HopBridge.TransactOpts, bonder, transferIds, totalAmount)
}

// SettleBondedWithdrawals is a paid mutator transaction binding the contract method 0xb162717e.
//
// Solidity: function settleBondedWithdrawals(address bonder, bytes32[] transferIds, uint256 totalAmount) returns()
func (_HopBridge *HopBridgeTransactorSession) SettleBondedWithdrawals(bonder common.Address, transferIds [][32]byte, totalAmount *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.SettleBondedWithdrawals(&_HopBridge.TransactOpts, bonder, transferIds, totalAmount)
}

// Stake is a paid mutator transaction binding the contract method 0xadc9772e.
//
// Solidity: function stake(address bonder, uint256 amount) payable returns()
func (_HopBridge *HopBridgeTransactor) Stake(opts *bind.TransactOpts, bonder common.Address, amount *big.Int) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "stake", bonder, amount)
}

// Stake is a paid mutator transaction binding the contract method 0xadc9772e.
//
// Solidity: function stake(address bonder, uint256 amount) payable returns()
func (_HopBridge *HopBridgeSession) Stake(bonder common.Address, amount *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.Stake(&_HopBridge.TransactOpts, bonder, amount)
}

// Stake is a paid mutator transaction binding the contract method 0xadc9772e.
//
// Solidity: function stake(address bonder, uint256 amount) payable returns()
func (_HopBridge *HopBridgeTransactorSession) Stake(bonder common.Address, amount *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.Stake(&_HopBridge.TransactOpts, bonder, amount)
}

// Unstake is a paid mutator transaction binding the contract method 0x2e17de78.
//
// Solidity: function unstake(uint256 amount) returns()
func (_HopBridge *HopBridgeTransactor) Unstake(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "unstake", amount)
}

// Unstake is a paid mutator transaction binding the contract method 0x2e17de78.
//
// Solidity: function unstake(uint256 amount) returns()
func (_HopBridge *HopBridgeSession) Unstake(amount *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.Unstake(&_HopBridge.TransactOpts, amount)
}

// Unstake is a paid mutator transaction binding the contract method 0x2e17de78.
//
// Solidity: function unstake(uint256 amount) returns()
func (_HopBridge *HopBridgeTransactorSession) Unstake(amount *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.Unstake(&_HopBridge.TransactOpts, amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x0f7aadb7.
//
// Solidity: function withdraw(address recipient, uint256 amount, bytes32 transferNonce, uint256 bonderFee, uint256 amountOutMin, uint256 deadline, bytes32 rootHash, uint256 transferRootTotalAmount, uint256 transferIdTreeIndex, bytes32[] siblings, uint256 totalLeaves) returns()
func (_HopBridge *HopBridgeTransactor) Withdraw(opts *bind.TransactOpts, recipient common.Address, amount *big.Int, transferNonce [32]byte, bonderFee *big.Int, amountOutMin *big.Int, deadline *big.Int, rootHash [32]byte, transferRootTotalAmount *big.Int, transferIdTreeIndex *big.Int, siblings [][32]byte, totalLeaves *big.Int) (*types.Transaction, error) {
	return _HopBridge.contract.Transact(opts, "withdraw", recipient, amount, transferNonce, bonderFee, amountOutMin, deadline, rootHash, transferRootTotalAmount, transferIdTreeIndex, siblings, totalLeaves)
}

// Withdraw is a paid mutator transaction binding the contract method 0x0f7aadb7.
//
// Solidity: function withdraw(address recipient, uint256 amount, bytes32 transferNonce, uint256 bonderFee, uint256 amountOutMin, uint256 deadline, bytes32 rootHash, uint256 transferRootTotalAmount, uint256 transferIdTreeIndex, bytes32[] siblings, uint256 totalLeaves) returns()
func (_HopBridge *HopBridgeSession) Withdraw(recipient common.Address, amount *big.Int, transferNonce [32]byte, bonderFee *big.Int, amountOutMin *big.Int, deadline *big.Int, rootHash [32]byte, transferRootTotalAmount *big.Int, transferIdTreeIndex *big.Int, siblings [][32]byte, totalLeaves *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.Withdraw(&_HopBridge.TransactOpts, recipient, amount, transferNonce, bonderFee, amountOutMin, deadline, rootHash, transferRootTotalAmount, transferIdTreeIndex, siblings, totalLeaves)
}

// Withdraw is a paid mutator transaction binding the contract method 0x0f7aadb7.
//
// Solidity: function withdraw(address recipient, uint256 amount, bytes32 transferNonce, uint256 bonderFee, uint256 amountOutMin, uint256 deadline, bytes32 rootHash, uint256 transferRootTotalAmount, uint256 transferIdTreeIndex, bytes32[] siblings, uint256 totalLeaves) returns()
func (_HopBridge *HopBridgeTransactorSession) Withdraw(recipient common.Address, amount *big.Int, transferNonce [32]byte, bonderFee *big.Int, amountOutMin *big.Int, deadline *big.Int, rootHash [32]byte, transferRootTotalAmount *big.Int, transferIdTreeIndex *big.Int, siblings [][32]byte, totalLeaves *big.Int) (*types.Transaction, error) {
	return _HopBridge.Contract.Withdraw(&_HopBridge.TransactOpts, recipient, amount, transferNonce, bonderFee, amountOutMin, deadline, rootHash, transferRootTotalAmount, transferIdTreeIndex, siblings, totalLeaves)
}

// HopBridgeBonderAddedIterator is returned from FilterBonderAdded and is used to iterate over the raw logs and unpacked data for BonderAdded events raised by the HopBridge contract.
type HopBridgeBonderAddedIterator struct {
	Event *HopBridgeBonderAdded // Event containing the contract specifics and raw log

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
func (it *HopBridgeBonderAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopBridgeBonderAdded)
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
		it.Event = new(HopBridgeBonderAdded)
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
func (it *HopBridgeBonderAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopBridgeBonderAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopBridgeBonderAdded represents a BonderAdded event raised by the HopBridge contract.
type HopBridgeBonderAdded struct {
	NewBonder common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterBonderAdded is a free log retrieval operation binding the contract event 0x2cec73b7434d3b91198ad1a618f63e6a0761ce281af5ec9ec76606d948d03e23.
//
// Solidity: event BonderAdded(address indexed newBonder)
func (_HopBridge *HopBridgeFilterer) FilterBonderAdded(opts *bind.FilterOpts, newBonder []common.Address) (*HopBridgeBonderAddedIterator, error) {

	var newBonderRule []interface{}
	for _, newBonderItem := range newBonder {
		newBonderRule = append(newBonderRule, newBonderItem)
	}

	logs, sub, err := _HopBridge.contract.FilterLogs(opts, "BonderAdded", newBonderRule)
	if err != nil {
		return nil, err
	}
	return &HopBridgeBonderAddedIterator{contract: _HopBridge.contract, event: "BonderAdded", logs: logs, sub: sub}, nil
}

// WatchBonderAdded is a free log subscription operation binding the contract event 0x2cec73b7434d3b91198ad1a618f63e6a0761ce281af5ec9ec76606d948d03e23.
//
// Solidity: event BonderAdded(address indexed newBonder)
func (_HopBridge *HopBridgeFilterer) WatchBonderAdded(opts *bind.WatchOpts, sink chan<- *HopBridgeBonderAdded, newBonder []common.Address) (event.Subscription, error) {

	var newBonderRule []interface{}
	for _, newBonderItem := range newBonder {
		newBonderRule = append(newBonderRule, newBonderItem)
	}

	logs, sub, err := _HopBridge.contract.WatchLogs(opts, "BonderAdded", newBonderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopBridgeBonderAdded)
				if err := _HopBridge.contract.UnpackLog(event, "BonderAdded", log); err != nil {
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

// ParseBonderAdded is a log parse operation binding the contract event 0x2cec73b7434d3b91198ad1a618f63e6a0761ce281af5ec9ec76606d948d03e23.
//
// Solidity: event BonderAdded(address indexed newBonder)
func (_HopBridge *HopBridgeFilterer) ParseBonderAdded(log types.Log) (*HopBridgeBonderAdded, error) {
	event := new(HopBridgeBonderAdded)
	if err := _HopBridge.contract.UnpackLog(event, "BonderAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopBridgeBonderRemovedIterator is returned from FilterBonderRemoved and is used to iterate over the raw logs and unpacked data for BonderRemoved events raised by the HopBridge contract.
type HopBridgeBonderRemovedIterator struct {
	Event *HopBridgeBonderRemoved // Event containing the contract specifics and raw log

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
func (it *HopBridgeBonderRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopBridgeBonderRemoved)
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
		it.Event = new(HopBridgeBonderRemoved)
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
func (it *HopBridgeBonderRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopBridgeBonderRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopBridgeBonderRemoved represents a BonderRemoved event raised by the HopBridge contract.
type HopBridgeBonderRemoved struct {
	PreviousBonder common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterBonderRemoved is a free log retrieval operation binding the contract event 0x4234ba611d325b3ba434c4e1b037967b955b1274d4185ee9847b7491111a48ff.
//
// Solidity: event BonderRemoved(address indexed previousBonder)
func (_HopBridge *HopBridgeFilterer) FilterBonderRemoved(opts *bind.FilterOpts, previousBonder []common.Address) (*HopBridgeBonderRemovedIterator, error) {

	var previousBonderRule []interface{}
	for _, previousBonderItem := range previousBonder {
		previousBonderRule = append(previousBonderRule, previousBonderItem)
	}

	logs, sub, err := _HopBridge.contract.FilterLogs(opts, "BonderRemoved", previousBonderRule)
	if err != nil {
		return nil, err
	}
	return &HopBridgeBonderRemovedIterator{contract: _HopBridge.contract, event: "BonderRemoved", logs: logs, sub: sub}, nil
}

// WatchBonderRemoved is a free log subscription operation binding the contract event 0x4234ba611d325b3ba434c4e1b037967b955b1274d4185ee9847b7491111a48ff.
//
// Solidity: event BonderRemoved(address indexed previousBonder)
func (_HopBridge *HopBridgeFilterer) WatchBonderRemoved(opts *bind.WatchOpts, sink chan<- *HopBridgeBonderRemoved, previousBonder []common.Address) (event.Subscription, error) {

	var previousBonderRule []interface{}
	for _, previousBonderItem := range previousBonder {
		previousBonderRule = append(previousBonderRule, previousBonderItem)
	}

	logs, sub, err := _HopBridge.contract.WatchLogs(opts, "BonderRemoved", previousBonderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopBridgeBonderRemoved)
				if err := _HopBridge.contract.UnpackLog(event, "BonderRemoved", log); err != nil {
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

// ParseBonderRemoved is a log parse operation binding the contract event 0x4234ba611d325b3ba434c4e1b037967b955b1274d4185ee9847b7491111a48ff.
//
// Solidity: event BonderRemoved(address indexed previousBonder)
func (_HopBridge *HopBridgeFilterer) ParseBonderRemoved(log types.Log) (*HopBridgeBonderRemoved, error) {
	event := new(HopBridgeBonderRemoved)
	if err := _HopBridge.contract.UnpackLog(event, "BonderRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopBridgeChallengeResolvedIterator is returned from FilterChallengeResolved and is used to iterate over the raw logs and unpacked data for ChallengeResolved events raised by the HopBridge contract.
type HopBridgeChallengeResolvedIterator struct {
	Event *HopBridgeChallengeResolved // Event containing the contract specifics and raw log

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
func (it *HopBridgeChallengeResolvedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopBridgeChallengeResolved)
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
		it.Event = new(HopBridgeChallengeResolved)
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
func (it *HopBridgeChallengeResolvedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopBridgeChallengeResolvedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopBridgeChallengeResolved represents a ChallengeResolved event raised by the HopBridge contract.
type HopBridgeChallengeResolved struct {
	TransferRootId [32]byte
	RootHash       [32]byte
	OriginalAmount *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterChallengeResolved is a free log retrieval operation binding the contract event 0x4a99228a8a6d774d261be57ab0ed833bb1bae1f22bbbd3d4767b75ad03fdddf7.
//
// Solidity: event ChallengeResolved(bytes32 indexed transferRootId, bytes32 indexed rootHash, uint256 originalAmount)
func (_HopBridge *HopBridgeFilterer) FilterChallengeResolved(opts *bind.FilterOpts, transferRootId [][32]byte, rootHash [][32]byte) (*HopBridgeChallengeResolvedIterator, error) {

	var transferRootIdRule []interface{}
	for _, transferRootIdItem := range transferRootId {
		transferRootIdRule = append(transferRootIdRule, transferRootIdItem)
	}
	var rootHashRule []interface{}
	for _, rootHashItem := range rootHash {
		rootHashRule = append(rootHashRule, rootHashItem)
	}

	logs, sub, err := _HopBridge.contract.FilterLogs(opts, "ChallengeResolved", transferRootIdRule, rootHashRule)
	if err != nil {
		return nil, err
	}
	return &HopBridgeChallengeResolvedIterator{contract: _HopBridge.contract, event: "ChallengeResolved", logs: logs, sub: sub}, nil
}

// WatchChallengeResolved is a free log subscription operation binding the contract event 0x4a99228a8a6d774d261be57ab0ed833bb1bae1f22bbbd3d4767b75ad03fdddf7.
//
// Solidity: event ChallengeResolved(bytes32 indexed transferRootId, bytes32 indexed rootHash, uint256 originalAmount)
func (_HopBridge *HopBridgeFilterer) WatchChallengeResolved(opts *bind.WatchOpts, sink chan<- *HopBridgeChallengeResolved, transferRootId [][32]byte, rootHash [][32]byte) (event.Subscription, error) {

	var transferRootIdRule []interface{}
	for _, transferRootIdItem := range transferRootId {
		transferRootIdRule = append(transferRootIdRule, transferRootIdItem)
	}
	var rootHashRule []interface{}
	for _, rootHashItem := range rootHash {
		rootHashRule = append(rootHashRule, rootHashItem)
	}

	logs, sub, err := _HopBridge.contract.WatchLogs(opts, "ChallengeResolved", transferRootIdRule, rootHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopBridgeChallengeResolved)
				if err := _HopBridge.contract.UnpackLog(event, "ChallengeResolved", log); err != nil {
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

// ParseChallengeResolved is a log parse operation binding the contract event 0x4a99228a8a6d774d261be57ab0ed833bb1bae1f22bbbd3d4767b75ad03fdddf7.
//
// Solidity: event ChallengeResolved(bytes32 indexed transferRootId, bytes32 indexed rootHash, uint256 originalAmount)
func (_HopBridge *HopBridgeFilterer) ParseChallengeResolved(log types.Log) (*HopBridgeChallengeResolved, error) {
	event := new(HopBridgeChallengeResolved)
	if err := _HopBridge.contract.UnpackLog(event, "ChallengeResolved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopBridgeMultipleWithdrawalsSettledIterator is returned from FilterMultipleWithdrawalsSettled and is used to iterate over the raw logs and unpacked data for MultipleWithdrawalsSettled events raised by the HopBridge contract.
type HopBridgeMultipleWithdrawalsSettledIterator struct {
	Event *HopBridgeMultipleWithdrawalsSettled // Event containing the contract specifics and raw log

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
func (it *HopBridgeMultipleWithdrawalsSettledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopBridgeMultipleWithdrawalsSettled)
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
		it.Event = new(HopBridgeMultipleWithdrawalsSettled)
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
func (it *HopBridgeMultipleWithdrawalsSettledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopBridgeMultipleWithdrawalsSettledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopBridgeMultipleWithdrawalsSettled represents a MultipleWithdrawalsSettled event raised by the HopBridge contract.
type HopBridgeMultipleWithdrawalsSettled struct {
	Bonder            common.Address
	RootHash          [32]byte
	TotalBondsSettled *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterMultipleWithdrawalsSettled is a free log retrieval operation binding the contract event 0x78e830d08be9d5f957414c84d685c061ecbd8467be98b42ebb64f0118b57d2ff.
//
// Solidity: event MultipleWithdrawalsSettled(address indexed bonder, bytes32 indexed rootHash, uint256 totalBondsSettled)
func (_HopBridge *HopBridgeFilterer) FilterMultipleWithdrawalsSettled(opts *bind.FilterOpts, bonder []common.Address, rootHash [][32]byte) (*HopBridgeMultipleWithdrawalsSettledIterator, error) {

	var bonderRule []interface{}
	for _, bonderItem := range bonder {
		bonderRule = append(bonderRule, bonderItem)
	}
	var rootHashRule []interface{}
	for _, rootHashItem := range rootHash {
		rootHashRule = append(rootHashRule, rootHashItem)
	}

	logs, sub, err := _HopBridge.contract.FilterLogs(opts, "MultipleWithdrawalsSettled", bonderRule, rootHashRule)
	if err != nil {
		return nil, err
	}
	return &HopBridgeMultipleWithdrawalsSettledIterator{contract: _HopBridge.contract, event: "MultipleWithdrawalsSettled", logs: logs, sub: sub}, nil
}

// WatchMultipleWithdrawalsSettled is a free log subscription operation binding the contract event 0x78e830d08be9d5f957414c84d685c061ecbd8467be98b42ebb64f0118b57d2ff.
//
// Solidity: event MultipleWithdrawalsSettled(address indexed bonder, bytes32 indexed rootHash, uint256 totalBondsSettled)
func (_HopBridge *HopBridgeFilterer) WatchMultipleWithdrawalsSettled(opts *bind.WatchOpts, sink chan<- *HopBridgeMultipleWithdrawalsSettled, bonder []common.Address, rootHash [][32]byte) (event.Subscription, error) {

	var bonderRule []interface{}
	for _, bonderItem := range bonder {
		bonderRule = append(bonderRule, bonderItem)
	}
	var rootHashRule []interface{}
	for _, rootHashItem := range rootHash {
		rootHashRule = append(rootHashRule, rootHashItem)
	}

	logs, sub, err := _HopBridge.contract.WatchLogs(opts, "MultipleWithdrawalsSettled", bonderRule, rootHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopBridgeMultipleWithdrawalsSettled)
				if err := _HopBridge.contract.UnpackLog(event, "MultipleWithdrawalsSettled", log); err != nil {
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

// ParseMultipleWithdrawalsSettled is a log parse operation binding the contract event 0x78e830d08be9d5f957414c84d685c061ecbd8467be98b42ebb64f0118b57d2ff.
//
// Solidity: event MultipleWithdrawalsSettled(address indexed bonder, bytes32 indexed rootHash, uint256 totalBondsSettled)
func (_HopBridge *HopBridgeFilterer) ParseMultipleWithdrawalsSettled(log types.Log) (*HopBridgeMultipleWithdrawalsSettled, error) {
	event := new(HopBridgeMultipleWithdrawalsSettled)
	if err := _HopBridge.contract.UnpackLog(event, "MultipleWithdrawalsSettled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopBridgeStakeIterator is returned from FilterStake and is used to iterate over the raw logs and unpacked data for Stake events raised by the HopBridge contract.
type HopBridgeStakeIterator struct {
	Event *HopBridgeStake // Event containing the contract specifics and raw log

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
func (it *HopBridgeStakeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopBridgeStake)
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
		it.Event = new(HopBridgeStake)
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
func (it *HopBridgeStakeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopBridgeStakeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopBridgeStake represents a Stake event raised by the HopBridge contract.
type HopBridgeStake struct {
	Account common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterStake is a free log retrieval operation binding the contract event 0xebedb8b3c678666e7f36970bc8f57abf6d8fa2e828c0da91ea5b75bf68ed101a.
//
// Solidity: event Stake(address indexed account, uint256 amount)
func (_HopBridge *HopBridgeFilterer) FilterStake(opts *bind.FilterOpts, account []common.Address) (*HopBridgeStakeIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _HopBridge.contract.FilterLogs(opts, "Stake", accountRule)
	if err != nil {
		return nil, err
	}
	return &HopBridgeStakeIterator{contract: _HopBridge.contract, event: "Stake", logs: logs, sub: sub}, nil
}

// WatchStake is a free log subscription operation binding the contract event 0xebedb8b3c678666e7f36970bc8f57abf6d8fa2e828c0da91ea5b75bf68ed101a.
//
// Solidity: event Stake(address indexed account, uint256 amount)
func (_HopBridge *HopBridgeFilterer) WatchStake(opts *bind.WatchOpts, sink chan<- *HopBridgeStake, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _HopBridge.contract.WatchLogs(opts, "Stake", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopBridgeStake)
				if err := _HopBridge.contract.UnpackLog(event, "Stake", log); err != nil {
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

// ParseStake is a log parse operation binding the contract event 0xebedb8b3c678666e7f36970bc8f57abf6d8fa2e828c0da91ea5b75bf68ed101a.
//
// Solidity: event Stake(address indexed account, uint256 amount)
func (_HopBridge *HopBridgeFilterer) ParseStake(log types.Log) (*HopBridgeStake, error) {
	event := new(HopBridgeStake)
	if err := _HopBridge.contract.UnpackLog(event, "Stake", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopBridgeTransferBondChallengedIterator is returned from FilterTransferBondChallenged and is used to iterate over the raw logs and unpacked data for TransferBondChallenged events raised by the HopBridge contract.
type HopBridgeTransferBondChallengedIterator struct {
	Event *HopBridgeTransferBondChallenged // Event containing the contract specifics and raw log

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
func (it *HopBridgeTransferBondChallengedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopBridgeTransferBondChallenged)
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
		it.Event = new(HopBridgeTransferBondChallenged)
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
func (it *HopBridgeTransferBondChallengedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopBridgeTransferBondChallengedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopBridgeTransferBondChallenged represents a TransferBondChallenged event raised by the HopBridge contract.
type HopBridgeTransferBondChallenged struct {
	TransferRootId [32]byte
	RootHash       [32]byte
	OriginalAmount *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterTransferBondChallenged is a free log retrieval operation binding the contract event 0xec2697dcba539a0ac947cdf1f6d0b6314c065429eca8be2435859b10209d4c27.
//
// Solidity: event TransferBondChallenged(bytes32 indexed transferRootId, bytes32 indexed rootHash, uint256 originalAmount)
func (_HopBridge *HopBridgeFilterer) FilterTransferBondChallenged(opts *bind.FilterOpts, transferRootId [][32]byte, rootHash [][32]byte) (*HopBridgeTransferBondChallengedIterator, error) {

	var transferRootIdRule []interface{}
	for _, transferRootIdItem := range transferRootId {
		transferRootIdRule = append(transferRootIdRule, transferRootIdItem)
	}
	var rootHashRule []interface{}
	for _, rootHashItem := range rootHash {
		rootHashRule = append(rootHashRule, rootHashItem)
	}

	logs, sub, err := _HopBridge.contract.FilterLogs(opts, "TransferBondChallenged", transferRootIdRule, rootHashRule)
	if err != nil {
		return nil, err
	}
	return &HopBridgeTransferBondChallengedIterator{contract: _HopBridge.contract, event: "TransferBondChallenged", logs: logs, sub: sub}, nil
}

// WatchTransferBondChallenged is a free log subscription operation binding the contract event 0xec2697dcba539a0ac947cdf1f6d0b6314c065429eca8be2435859b10209d4c27.
//
// Solidity: event TransferBondChallenged(bytes32 indexed transferRootId, bytes32 indexed rootHash, uint256 originalAmount)
func (_HopBridge *HopBridgeFilterer) WatchTransferBondChallenged(opts *bind.WatchOpts, sink chan<- *HopBridgeTransferBondChallenged, transferRootId [][32]byte, rootHash [][32]byte) (event.Subscription, error) {

	var transferRootIdRule []interface{}
	for _, transferRootIdItem := range transferRootId {
		transferRootIdRule = append(transferRootIdRule, transferRootIdItem)
	}
	var rootHashRule []interface{}
	for _, rootHashItem := range rootHash {
		rootHashRule = append(rootHashRule, rootHashItem)
	}

	logs, sub, err := _HopBridge.contract.WatchLogs(opts, "TransferBondChallenged", transferRootIdRule, rootHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopBridgeTransferBondChallenged)
				if err := _HopBridge.contract.UnpackLog(event, "TransferBondChallenged", log); err != nil {
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

// ParseTransferBondChallenged is a log parse operation binding the contract event 0xec2697dcba539a0ac947cdf1f6d0b6314c065429eca8be2435859b10209d4c27.
//
// Solidity: event TransferBondChallenged(bytes32 indexed transferRootId, bytes32 indexed rootHash, uint256 originalAmount)
func (_HopBridge *HopBridgeFilterer) ParseTransferBondChallenged(log types.Log) (*HopBridgeTransferBondChallenged, error) {
	event := new(HopBridgeTransferBondChallenged)
	if err := _HopBridge.contract.UnpackLog(event, "TransferBondChallenged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopBridgeTransferRootBondedIterator is returned from FilterTransferRootBonded and is used to iterate over the raw logs and unpacked data for TransferRootBonded events raised by the HopBridge contract.
type HopBridgeTransferRootBondedIterator struct {
	Event *HopBridgeTransferRootBonded // Event containing the contract specifics and raw log

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
func (it *HopBridgeTransferRootBondedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopBridgeTransferRootBonded)
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
		it.Event = new(HopBridgeTransferRootBonded)
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
func (it *HopBridgeTransferRootBondedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopBridgeTransferRootBondedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopBridgeTransferRootBonded represents a TransferRootBonded event raised by the HopBridge contract.
type HopBridgeTransferRootBonded struct {
	Root   [32]byte
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterTransferRootBonded is a free log retrieval operation binding the contract event 0xa57b3e1f3af9eca02201028629700658608222c365064584cfe65d9630ef4f7b.
//
// Solidity: event TransferRootBonded(bytes32 indexed root, uint256 amount)
func (_HopBridge *HopBridgeFilterer) FilterTransferRootBonded(opts *bind.FilterOpts, root [][32]byte) (*HopBridgeTransferRootBondedIterator, error) {

	var rootRule []interface{}
	for _, rootItem := range root {
		rootRule = append(rootRule, rootItem)
	}

	logs, sub, err := _HopBridge.contract.FilterLogs(opts, "TransferRootBonded", rootRule)
	if err != nil {
		return nil, err
	}
	return &HopBridgeTransferRootBondedIterator{contract: _HopBridge.contract, event: "TransferRootBonded", logs: logs, sub: sub}, nil
}

// WatchTransferRootBonded is a free log subscription operation binding the contract event 0xa57b3e1f3af9eca02201028629700658608222c365064584cfe65d9630ef4f7b.
//
// Solidity: event TransferRootBonded(bytes32 indexed root, uint256 amount)
func (_HopBridge *HopBridgeFilterer) WatchTransferRootBonded(opts *bind.WatchOpts, sink chan<- *HopBridgeTransferRootBonded, root [][32]byte) (event.Subscription, error) {

	var rootRule []interface{}
	for _, rootItem := range root {
		rootRule = append(rootRule, rootItem)
	}

	logs, sub, err := _HopBridge.contract.WatchLogs(opts, "TransferRootBonded", rootRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopBridgeTransferRootBonded)
				if err := _HopBridge.contract.UnpackLog(event, "TransferRootBonded", log); err != nil {
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

// ParseTransferRootBonded is a log parse operation binding the contract event 0xa57b3e1f3af9eca02201028629700658608222c365064584cfe65d9630ef4f7b.
//
// Solidity: event TransferRootBonded(bytes32 indexed root, uint256 amount)
func (_HopBridge *HopBridgeFilterer) ParseTransferRootBonded(log types.Log) (*HopBridgeTransferRootBonded, error) {
	event := new(HopBridgeTransferRootBonded)
	if err := _HopBridge.contract.UnpackLog(event, "TransferRootBonded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopBridgeTransferRootConfirmedIterator is returned from FilterTransferRootConfirmed and is used to iterate over the raw logs and unpacked data for TransferRootConfirmed events raised by the HopBridge contract.
type HopBridgeTransferRootConfirmedIterator struct {
	Event *HopBridgeTransferRootConfirmed // Event containing the contract specifics and raw log

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
func (it *HopBridgeTransferRootConfirmedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopBridgeTransferRootConfirmed)
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
		it.Event = new(HopBridgeTransferRootConfirmed)
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
func (it *HopBridgeTransferRootConfirmedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopBridgeTransferRootConfirmedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopBridgeTransferRootConfirmed represents a TransferRootConfirmed event raised by the HopBridge contract.
type HopBridgeTransferRootConfirmed struct {
	OriginChainId      *big.Int
	DestinationChainId *big.Int
	RootHash           [32]byte
	TotalAmount        *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterTransferRootConfirmed is a free log retrieval operation binding the contract event 0xfdfb0eefa96935b8a8c0edf528e125dc6f3934fdbbfce31b38967e8ff413dccd.
//
// Solidity: event TransferRootConfirmed(uint256 indexed originChainId, uint256 indexed destinationChainId, bytes32 indexed rootHash, uint256 totalAmount)
func (_HopBridge *HopBridgeFilterer) FilterTransferRootConfirmed(opts *bind.FilterOpts, originChainId []*big.Int, destinationChainId []*big.Int, rootHash [][32]byte) (*HopBridgeTransferRootConfirmedIterator, error) {

	var originChainIdRule []interface{}
	for _, originChainIdItem := range originChainId {
		originChainIdRule = append(originChainIdRule, originChainIdItem)
	}
	var destinationChainIdRule []interface{}
	for _, destinationChainIdItem := range destinationChainId {
		destinationChainIdRule = append(destinationChainIdRule, destinationChainIdItem)
	}
	var rootHashRule []interface{}
	for _, rootHashItem := range rootHash {
		rootHashRule = append(rootHashRule, rootHashItem)
	}

	logs, sub, err := _HopBridge.contract.FilterLogs(opts, "TransferRootConfirmed", originChainIdRule, destinationChainIdRule, rootHashRule)
	if err != nil {
		return nil, err
	}
	return &HopBridgeTransferRootConfirmedIterator{contract: _HopBridge.contract, event: "TransferRootConfirmed", logs: logs, sub: sub}, nil
}

// WatchTransferRootConfirmed is a free log subscription operation binding the contract event 0xfdfb0eefa96935b8a8c0edf528e125dc6f3934fdbbfce31b38967e8ff413dccd.
//
// Solidity: event TransferRootConfirmed(uint256 indexed originChainId, uint256 indexed destinationChainId, bytes32 indexed rootHash, uint256 totalAmount)
func (_HopBridge *HopBridgeFilterer) WatchTransferRootConfirmed(opts *bind.WatchOpts, sink chan<- *HopBridgeTransferRootConfirmed, originChainId []*big.Int, destinationChainId []*big.Int, rootHash [][32]byte) (event.Subscription, error) {

	var originChainIdRule []interface{}
	for _, originChainIdItem := range originChainId {
		originChainIdRule = append(originChainIdRule, originChainIdItem)
	}
	var destinationChainIdRule []interface{}
	for _, destinationChainIdItem := range destinationChainId {
		destinationChainIdRule = append(destinationChainIdRule, destinationChainIdItem)
	}
	var rootHashRule []interface{}
	for _, rootHashItem := range rootHash {
		rootHashRule = append(rootHashRule, rootHashItem)
	}

	logs, sub, err := _HopBridge.contract.WatchLogs(opts, "TransferRootConfirmed", originChainIdRule, destinationChainIdRule, rootHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopBridgeTransferRootConfirmed)
				if err := _HopBridge.contract.UnpackLog(event, "TransferRootConfirmed", log); err != nil {
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

// ParseTransferRootConfirmed is a log parse operation binding the contract event 0xfdfb0eefa96935b8a8c0edf528e125dc6f3934fdbbfce31b38967e8ff413dccd.
//
// Solidity: event TransferRootConfirmed(uint256 indexed originChainId, uint256 indexed destinationChainId, bytes32 indexed rootHash, uint256 totalAmount)
func (_HopBridge *HopBridgeFilterer) ParseTransferRootConfirmed(log types.Log) (*HopBridgeTransferRootConfirmed, error) {
	event := new(HopBridgeTransferRootConfirmed)
	if err := _HopBridge.contract.UnpackLog(event, "TransferRootConfirmed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopBridgeTransferRootSetIterator is returned from FilterTransferRootSet and is used to iterate over the raw logs and unpacked data for TransferRootSet events raised by the HopBridge contract.
type HopBridgeTransferRootSetIterator struct {
	Event *HopBridgeTransferRootSet // Event containing the contract specifics and raw log

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
func (it *HopBridgeTransferRootSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopBridgeTransferRootSet)
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
		it.Event = new(HopBridgeTransferRootSet)
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
func (it *HopBridgeTransferRootSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopBridgeTransferRootSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopBridgeTransferRootSet represents a TransferRootSet event raised by the HopBridge contract.
type HopBridgeTransferRootSet struct {
	RootHash    [32]byte
	TotalAmount *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterTransferRootSet is a free log retrieval operation binding the contract event 0xb33d2162aead99dab59e77a7a67ea025b776bf8ca8079e132afdf9b23e03bd42.
//
// Solidity: event TransferRootSet(bytes32 indexed rootHash, uint256 totalAmount)
func (_HopBridge *HopBridgeFilterer) FilterTransferRootSet(opts *bind.FilterOpts, rootHash [][32]byte) (*HopBridgeTransferRootSetIterator, error) {

	var rootHashRule []interface{}
	for _, rootHashItem := range rootHash {
		rootHashRule = append(rootHashRule, rootHashItem)
	}

	logs, sub, err := _HopBridge.contract.FilterLogs(opts, "TransferRootSet", rootHashRule)
	if err != nil {
		return nil, err
	}
	return &HopBridgeTransferRootSetIterator{contract: _HopBridge.contract, event: "TransferRootSet", logs: logs, sub: sub}, nil
}

// WatchTransferRootSet is a free log subscription operation binding the contract event 0xb33d2162aead99dab59e77a7a67ea025b776bf8ca8079e132afdf9b23e03bd42.
//
// Solidity: event TransferRootSet(bytes32 indexed rootHash, uint256 totalAmount)
func (_HopBridge *HopBridgeFilterer) WatchTransferRootSet(opts *bind.WatchOpts, sink chan<- *HopBridgeTransferRootSet, rootHash [][32]byte) (event.Subscription, error) {

	var rootHashRule []interface{}
	for _, rootHashItem := range rootHash {
		rootHashRule = append(rootHashRule, rootHashItem)
	}

	logs, sub, err := _HopBridge.contract.WatchLogs(opts, "TransferRootSet", rootHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopBridgeTransferRootSet)
				if err := _HopBridge.contract.UnpackLog(event, "TransferRootSet", log); err != nil {
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

// ParseTransferRootSet is a log parse operation binding the contract event 0xb33d2162aead99dab59e77a7a67ea025b776bf8ca8079e132afdf9b23e03bd42.
//
// Solidity: event TransferRootSet(bytes32 indexed rootHash, uint256 totalAmount)
func (_HopBridge *HopBridgeFilterer) ParseTransferRootSet(log types.Log) (*HopBridgeTransferRootSet, error) {
	event := new(HopBridgeTransferRootSet)
	if err := _HopBridge.contract.UnpackLog(event, "TransferRootSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopBridgeTransferSentToL2Iterator is returned from FilterTransferSentToL2 and is used to iterate over the raw logs and unpacked data for TransferSentToL2 events raised by the HopBridge contract.
type HopBridgeTransferSentToL2Iterator struct {
	Event *HopBridgeTransferSentToL2 // Event containing the contract specifics and raw log

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
func (it *HopBridgeTransferSentToL2Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopBridgeTransferSentToL2)
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
		it.Event = new(HopBridgeTransferSentToL2)
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
func (it *HopBridgeTransferSentToL2Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopBridgeTransferSentToL2Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopBridgeTransferSentToL2 represents a TransferSentToL2 event raised by the HopBridge contract.
type HopBridgeTransferSentToL2 struct {
	ChainId      *big.Int
	Recipient    common.Address
	Amount       *big.Int
	AmountOutMin *big.Int
	Deadline     *big.Int
	Relayer      common.Address
	RelayerFee   *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterTransferSentToL2 is a free log retrieval operation binding the contract event 0x0a0607688c86ec1775abcdbab7b33a3a35a6c9cde677c9be880150c231cc6b0b.
//
// Solidity: event TransferSentToL2(uint256 indexed chainId, address indexed recipient, uint256 amount, uint256 amountOutMin, uint256 deadline, address indexed relayer, uint256 relayerFee)
func (_HopBridge *HopBridgeFilterer) FilterTransferSentToL2(opts *bind.FilterOpts, chainId []*big.Int, recipient []common.Address, relayer []common.Address) (*HopBridgeTransferSentToL2Iterator, error) {

	var chainIdRule []interface{}
	for _, chainIdItem := range chainId {
		chainIdRule = append(chainIdRule, chainIdItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	var relayerRule []interface{}
	for _, relayerItem := range relayer {
		relayerRule = append(relayerRule, relayerItem)
	}

	logs, sub, err := _HopBridge.contract.FilterLogs(opts, "TransferSentToL2", chainIdRule, recipientRule, relayerRule)
	if err != nil {
		return nil, err
	}
	return &HopBridgeTransferSentToL2Iterator{contract: _HopBridge.contract, event: "TransferSentToL2", logs: logs, sub: sub}, nil
}

// WatchTransferSentToL2 is a free log subscription operation binding the contract event 0x0a0607688c86ec1775abcdbab7b33a3a35a6c9cde677c9be880150c231cc6b0b.
//
// Solidity: event TransferSentToL2(uint256 indexed chainId, address indexed recipient, uint256 amount, uint256 amountOutMin, uint256 deadline, address indexed relayer, uint256 relayerFee)
func (_HopBridge *HopBridgeFilterer) WatchTransferSentToL2(opts *bind.WatchOpts, sink chan<- *HopBridgeTransferSentToL2, chainId []*big.Int, recipient []common.Address, relayer []common.Address) (event.Subscription, error) {

	var chainIdRule []interface{}
	for _, chainIdItem := range chainId {
		chainIdRule = append(chainIdRule, chainIdItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	var relayerRule []interface{}
	for _, relayerItem := range relayer {
		relayerRule = append(relayerRule, relayerItem)
	}

	logs, sub, err := _HopBridge.contract.WatchLogs(opts, "TransferSentToL2", chainIdRule, recipientRule, relayerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopBridgeTransferSentToL2)
				if err := _HopBridge.contract.UnpackLog(event, "TransferSentToL2", log); err != nil {
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

// ParseTransferSentToL2 is a log parse operation binding the contract event 0x0a0607688c86ec1775abcdbab7b33a3a35a6c9cde677c9be880150c231cc6b0b.
//
// Solidity: event TransferSentToL2(uint256 indexed chainId, address indexed recipient, uint256 amount, uint256 amountOutMin, uint256 deadline, address indexed relayer, uint256 relayerFee)
func (_HopBridge *HopBridgeFilterer) ParseTransferSentToL2(log types.Log) (*HopBridgeTransferSentToL2, error) {
	event := new(HopBridgeTransferSentToL2)
	if err := _HopBridge.contract.UnpackLog(event, "TransferSentToL2", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopBridgeUnstakeIterator is returned from FilterUnstake and is used to iterate over the raw logs and unpacked data for Unstake events raised by the HopBridge contract.
type HopBridgeUnstakeIterator struct {
	Event *HopBridgeUnstake // Event containing the contract specifics and raw log

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
func (it *HopBridgeUnstakeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopBridgeUnstake)
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
		it.Event = new(HopBridgeUnstake)
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
func (it *HopBridgeUnstakeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopBridgeUnstakeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopBridgeUnstake represents a Unstake event raised by the HopBridge contract.
type HopBridgeUnstake struct {
	Account common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUnstake is a free log retrieval operation binding the contract event 0x85082129d87b2fe11527cb1b3b7a520aeb5aa6913f88a3d8757fe40d1db02fdd.
//
// Solidity: event Unstake(address indexed account, uint256 amount)
func (_HopBridge *HopBridgeFilterer) FilterUnstake(opts *bind.FilterOpts, account []common.Address) (*HopBridgeUnstakeIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _HopBridge.contract.FilterLogs(opts, "Unstake", accountRule)
	if err != nil {
		return nil, err
	}
	return &HopBridgeUnstakeIterator{contract: _HopBridge.contract, event: "Unstake", logs: logs, sub: sub}, nil
}

// WatchUnstake is a free log subscription operation binding the contract event 0x85082129d87b2fe11527cb1b3b7a520aeb5aa6913f88a3d8757fe40d1db02fdd.
//
// Solidity: event Unstake(address indexed account, uint256 amount)
func (_HopBridge *HopBridgeFilterer) WatchUnstake(opts *bind.WatchOpts, sink chan<- *HopBridgeUnstake, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _HopBridge.contract.WatchLogs(opts, "Unstake", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopBridgeUnstake)
				if err := _HopBridge.contract.UnpackLog(event, "Unstake", log); err != nil {
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

// ParseUnstake is a log parse operation binding the contract event 0x85082129d87b2fe11527cb1b3b7a520aeb5aa6913f88a3d8757fe40d1db02fdd.
//
// Solidity: event Unstake(address indexed account, uint256 amount)
func (_HopBridge *HopBridgeFilterer) ParseUnstake(log types.Log) (*HopBridgeUnstake, error) {
	event := new(HopBridgeUnstake)
	if err := _HopBridge.contract.UnpackLog(event, "Unstake", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopBridgeWithdrawalBondSettledIterator is returned from FilterWithdrawalBondSettled and is used to iterate over the raw logs and unpacked data for WithdrawalBondSettled events raised by the HopBridge contract.
type HopBridgeWithdrawalBondSettledIterator struct {
	Event *HopBridgeWithdrawalBondSettled // Event containing the contract specifics and raw log

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
func (it *HopBridgeWithdrawalBondSettledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopBridgeWithdrawalBondSettled)
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
		it.Event = new(HopBridgeWithdrawalBondSettled)
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
func (it *HopBridgeWithdrawalBondSettledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopBridgeWithdrawalBondSettledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopBridgeWithdrawalBondSettled represents a WithdrawalBondSettled event raised by the HopBridge contract.
type HopBridgeWithdrawalBondSettled struct {
	Bonder     common.Address
	TransferId [32]byte
	RootHash   [32]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterWithdrawalBondSettled is a free log retrieval operation binding the contract event 0x84eb21b24c31b27a3bc67dde4a598aad06db6e9415cd66544492b9616996143c.
//
// Solidity: event WithdrawalBondSettled(address indexed bonder, bytes32 indexed transferId, bytes32 indexed rootHash)
func (_HopBridge *HopBridgeFilterer) FilterWithdrawalBondSettled(opts *bind.FilterOpts, bonder []common.Address, transferId [][32]byte, rootHash [][32]byte) (*HopBridgeWithdrawalBondSettledIterator, error) {

	var bonderRule []interface{}
	for _, bonderItem := range bonder {
		bonderRule = append(bonderRule, bonderItem)
	}
	var transferIdRule []interface{}
	for _, transferIdItem := range transferId {
		transferIdRule = append(transferIdRule, transferIdItem)
	}
	var rootHashRule []interface{}
	for _, rootHashItem := range rootHash {
		rootHashRule = append(rootHashRule, rootHashItem)
	}

	logs, sub, err := _HopBridge.contract.FilterLogs(opts, "WithdrawalBondSettled", bonderRule, transferIdRule, rootHashRule)
	if err != nil {
		return nil, err
	}
	return &HopBridgeWithdrawalBondSettledIterator{contract: _HopBridge.contract, event: "WithdrawalBondSettled", logs: logs, sub: sub}, nil
}

// WatchWithdrawalBondSettled is a free log subscription operation binding the contract event 0x84eb21b24c31b27a3bc67dde4a598aad06db6e9415cd66544492b9616996143c.
//
// Solidity: event WithdrawalBondSettled(address indexed bonder, bytes32 indexed transferId, bytes32 indexed rootHash)
func (_HopBridge *HopBridgeFilterer) WatchWithdrawalBondSettled(opts *bind.WatchOpts, sink chan<- *HopBridgeWithdrawalBondSettled, bonder []common.Address, transferId [][32]byte, rootHash [][32]byte) (event.Subscription, error) {

	var bonderRule []interface{}
	for _, bonderItem := range bonder {
		bonderRule = append(bonderRule, bonderItem)
	}
	var transferIdRule []interface{}
	for _, transferIdItem := range transferId {
		transferIdRule = append(transferIdRule, transferIdItem)
	}
	var rootHashRule []interface{}
	for _, rootHashItem := range rootHash {
		rootHashRule = append(rootHashRule, rootHashItem)
	}

	logs, sub, err := _HopBridge.contract.WatchLogs(opts, "WithdrawalBondSettled", bonderRule, transferIdRule, rootHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopBridgeWithdrawalBondSettled)
				if err := _HopBridge.contract.UnpackLog(event, "WithdrawalBondSettled", log); err != nil {
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

// ParseWithdrawalBondSettled is a log parse operation binding the contract event 0x84eb21b24c31b27a3bc67dde4a598aad06db6e9415cd66544492b9616996143c.
//
// Solidity: event WithdrawalBondSettled(address indexed bonder, bytes32 indexed transferId, bytes32 indexed rootHash)
func (_HopBridge *HopBridgeFilterer) ParseWithdrawalBondSettled(log types.Log) (*HopBridgeWithdrawalBondSettled, error) {
	event := new(HopBridgeWithdrawalBondSettled)
	if err := _HopBridge.contract.UnpackLog(event, "WithdrawalBondSettled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopBridgeWithdrawalBondedIterator is returned from FilterWithdrawalBonded and is used to iterate over the raw logs and unpacked data for WithdrawalBonded events raised by the HopBridge contract.
type HopBridgeWithdrawalBondedIterator struct {
	Event *HopBridgeWithdrawalBonded // Event containing the contract specifics and raw log

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
func (it *HopBridgeWithdrawalBondedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopBridgeWithdrawalBonded)
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
		it.Event = new(HopBridgeWithdrawalBonded)
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
func (it *HopBridgeWithdrawalBondedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopBridgeWithdrawalBondedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopBridgeWithdrawalBonded represents a WithdrawalBonded event raised by the HopBridge contract.
type HopBridgeWithdrawalBonded struct {
	TransferId [32]byte
	Amount     *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterWithdrawalBonded is a free log retrieval operation binding the contract event 0x0c3d250c7831051e78aa6a56679e590374c7c424415ffe4aa474491def2fe705.
//
// Solidity: event WithdrawalBonded(bytes32 indexed transferId, uint256 amount)
func (_HopBridge *HopBridgeFilterer) FilterWithdrawalBonded(opts *bind.FilterOpts, transferId [][32]byte) (*HopBridgeWithdrawalBondedIterator, error) {

	var transferIdRule []interface{}
	for _, transferIdItem := range transferId {
		transferIdRule = append(transferIdRule, transferIdItem)
	}

	logs, sub, err := _HopBridge.contract.FilterLogs(opts, "WithdrawalBonded", transferIdRule)
	if err != nil {
		return nil, err
	}
	return &HopBridgeWithdrawalBondedIterator{contract: _HopBridge.contract, event: "WithdrawalBonded", logs: logs, sub: sub}, nil
}

// WatchWithdrawalBonded is a free log subscription operation binding the contract event 0x0c3d250c7831051e78aa6a56679e590374c7c424415ffe4aa474491def2fe705.
//
// Solidity: event WithdrawalBonded(bytes32 indexed transferId, uint256 amount)
func (_HopBridge *HopBridgeFilterer) WatchWithdrawalBonded(opts *bind.WatchOpts, sink chan<- *HopBridgeWithdrawalBonded, transferId [][32]byte) (event.Subscription, error) {

	var transferIdRule []interface{}
	for _, transferIdItem := range transferId {
		transferIdRule = append(transferIdRule, transferIdItem)
	}

	logs, sub, err := _HopBridge.contract.WatchLogs(opts, "WithdrawalBonded", transferIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopBridgeWithdrawalBonded)
				if err := _HopBridge.contract.UnpackLog(event, "WithdrawalBonded", log); err != nil {
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

// ParseWithdrawalBonded is a log parse operation binding the contract event 0x0c3d250c7831051e78aa6a56679e590374c7c424415ffe4aa474491def2fe705.
//
// Solidity: event WithdrawalBonded(bytes32 indexed transferId, uint256 amount)
func (_HopBridge *HopBridgeFilterer) ParseWithdrawalBonded(log types.Log) (*HopBridgeWithdrawalBonded, error) {
	event := new(HopBridgeWithdrawalBonded)
	if err := _HopBridge.contract.UnpackLog(event, "WithdrawalBonded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HopBridgeWithdrewIterator is returned from FilterWithdrew and is used to iterate over the raw logs and unpacked data for Withdrew events raised by the HopBridge contract.
type HopBridgeWithdrewIterator struct {
	Event *HopBridgeWithdrew // Event containing the contract specifics and raw log

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
func (it *HopBridgeWithdrewIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HopBridgeWithdrew)
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
		it.Event = new(HopBridgeWithdrew)
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
func (it *HopBridgeWithdrewIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HopBridgeWithdrewIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HopBridgeWithdrew represents a Withdrew event raised by the HopBridge contract.
type HopBridgeWithdrew struct {
	TransferId    [32]byte
	Recipient     common.Address
	Amount        *big.Int
	TransferNonce [32]byte
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterWithdrew is a free log retrieval operation binding the contract event 0x9475cdbde5fc71fe2ccd413c82878ee54d061b9f74f9e2e1a03ff1178821502c.
//
// Solidity: event Withdrew(bytes32 indexed transferId, address indexed recipient, uint256 amount, bytes32 transferNonce)
func (_HopBridge *HopBridgeFilterer) FilterWithdrew(opts *bind.FilterOpts, transferId [][32]byte, recipient []common.Address) (*HopBridgeWithdrewIterator, error) {

	var transferIdRule []interface{}
	for _, transferIdItem := range transferId {
		transferIdRule = append(transferIdRule, transferIdItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _HopBridge.contract.FilterLogs(opts, "Withdrew", transferIdRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &HopBridgeWithdrewIterator{contract: _HopBridge.contract, event: "Withdrew", logs: logs, sub: sub}, nil
}

// WatchWithdrew is a free log subscription operation binding the contract event 0x9475cdbde5fc71fe2ccd413c82878ee54d061b9f74f9e2e1a03ff1178821502c.
//
// Solidity: event Withdrew(bytes32 indexed transferId, address indexed recipient, uint256 amount, bytes32 transferNonce)
func (_HopBridge *HopBridgeFilterer) WatchWithdrew(opts *bind.WatchOpts, sink chan<- *HopBridgeWithdrew, transferId [][32]byte, recipient []common.Address) (event.Subscription, error) {

	var transferIdRule []interface{}
	for _, transferIdItem := range transferId {
		transferIdRule = append(transferIdRule, transferIdItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _HopBridge.contract.WatchLogs(opts, "Withdrew", transferIdRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HopBridgeWithdrew)
				if err := _HopBridge.contract.UnpackLog(event, "Withdrew", log); err != nil {
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

// ParseWithdrew is a log parse operation binding the contract event 0x9475cdbde5fc71fe2ccd413c82878ee54d061b9f74f9e2e1a03ff1178821502c.
//
// Solidity: event Withdrew(bytes32 indexed transferId, address indexed recipient, uint256 amount, bytes32 transferNonce)
func (_HopBridge *HopBridgeFilterer) ParseWithdrew(log types.Log) (*HopBridgeWithdrew, error) {
	event := new(HopBridgeWithdrew)
	if err := _HopBridge.contract.UnpackLog(event, "Withdrew", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
