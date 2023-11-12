package activity

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	// used for embedding the sql query in the binary
	_ "embed"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"

	"github.com/status-im/status-go/services/wallet/bigint"
	"github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/services/wallet/transfer"
	"github.com/status-im/status-go/transactions"

	"golang.org/x/exp/constraints"
)

type PayloadType = int

// Beware: please update multiTransactionTypeToActivityType if changing this enum
const (
	MultiTransactionPT PayloadType = iota + 1
	SimpleTransactionPT
	PendingTransactionPT
)

var (
	ZeroAddress = eth.Address{}
)

type TransferType = int

const (
	TransferTypeEth TransferType = iota + 1
	TransferTypeErc20
	TransferTypeErc721
	TransferTypeErc1155
)

type Entry struct {
	payloadType     PayloadType
	transaction     *transfer.TransactionIdentity
	id              transfer.MultiTransactionIDType
	timestamp       int64
	activityType    Type
	activityStatus  Status
	amountOut       *hexutil.Big // Used for activityType SendAT, SwapAT, BridgeAT
	amountIn        *hexutil.Big // Used for activityType ReceiveAT, BuyAT, SwapAT, BridgeAT
	tokenOut        *Token       // Used for activityType SendAT, SwapAT, BridgeAT
	tokenIn         *Token       // Used for activityType ReceiveAT, BuyAT, SwapAT, BridgeAT
	symbolOut       *string
	symbolIn        *string
	sender          *eth.Address
	recipient       *eth.Address
	chainIDOut      *common.ChainID
	chainIDIn       *common.ChainID
	transferType    *TransferType
	contractAddress *eth.Address

	isNew bool // isNew is used to indicate if the entry is newer than session start (changed state also)
}

// Only used for JSON marshalling
type EntryData struct {
	PayloadType     PayloadType                      `json:"payloadType"`
	Transaction     *transfer.TransactionIdentity    `json:"transaction,omitempty"`
	ID              *transfer.MultiTransactionIDType `json:"id,omitempty"`
	Timestamp       *int64                           `json:"timestamp,omitempty"`
	ActivityType    *Type                            `json:"activityType,omitempty"`
	ActivityStatus  *Status                          `json:"activityStatus,omitempty"`
	AmountOut       *hexutil.Big                     `json:"amountOut,omitempty"`
	AmountIn        *hexutil.Big                     `json:"amountIn,omitempty"`
	TokenOut        *Token                           `json:"tokenOut,omitempty"`
	TokenIn         *Token                           `json:"tokenIn,omitempty"`
	SymbolOut       *string                          `json:"symbolOut,omitempty"`
	SymbolIn        *string                          `json:"symbolIn,omitempty"`
	Sender          *eth.Address                     `json:"sender,omitempty"`
	Recipient       *eth.Address                     `json:"recipient,omitempty"`
	ChainIDOut      *common.ChainID                  `json:"chainIdOut,omitempty"`
	ChainIDIn       *common.ChainID                  `json:"chainIdIn,omitempty"`
	TransferType    *TransferType                    `json:"transferType,omitempty"`
	ContractAddress *eth.Address                     `json:"contractAddress,omitempty"`

	IsNew *bool `json:"isNew,omitempty"`

	NftName *string `json:"nftName,omitempty"`
	NftURL  *string `json:"nftUrl,omitempty"`
}

func (e *Entry) MarshalJSON() ([]byte, error) {
	data := EntryData{
		Timestamp:       &e.timestamp,
		ActivityType:    &e.activityType,
		ActivityStatus:  &e.activityStatus,
		AmountOut:       e.amountOut,
		AmountIn:        e.amountIn,
		TokenOut:        e.tokenOut,
		TokenIn:         e.tokenIn,
		SymbolOut:       e.symbolOut,
		SymbolIn:        e.symbolIn,
		Sender:          e.sender,
		Recipient:       e.recipient,
		ChainIDOut:      e.chainIDOut,
		ChainIDIn:       e.chainIDIn,
		TransferType:    e.transferType,
		ContractAddress: e.contractAddress,
	}

	if e.payloadType == MultiTransactionPT {
		data.ID = common.NewAndSet(e.id)
	} else {
		data.Transaction = e.transaction
	}

	data.PayloadType = e.payloadType
	if e.isNew {
		data.IsNew = &e.isNew
	}

	return json.Marshal(data)
}

func (e *Entry) UnmarshalJSON(data []byte) error {
	aux := EntryData{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	e.payloadType = aux.PayloadType
	e.transaction = aux.Transaction
	if aux.ID != nil {
		e.id = *aux.ID
	}
	if aux.Timestamp != nil {
		e.timestamp = *aux.Timestamp
	}
	if aux.ActivityType != nil {
		e.activityType = *aux.ActivityType
	}
	if aux.ActivityStatus != nil {
		e.activityStatus = *aux.ActivityStatus
	}
	e.amountOut = aux.AmountOut
	e.amountIn = aux.AmountIn
	e.tokenOut = aux.TokenOut
	e.tokenIn = aux.TokenIn
	e.symbolOut = aux.SymbolOut
	e.symbolIn = aux.SymbolIn
	e.sender = aux.Sender
	e.recipient = aux.Recipient
	e.chainIDOut = aux.ChainIDOut
	e.chainIDIn = aux.ChainIDIn
	e.transferType = aux.TransferType

	e.isNew = aux.IsNew != nil && *aux.IsNew

	return nil
}

func newActivityEntryWithPendingTransaction(transaction *transfer.TransactionIdentity, timestamp int64, activityType Type, activityStatus Status) Entry {
	return newActivityEntryWithTransaction(true, transaction, timestamp, activityType, activityStatus)
}

func newActivityEntryWithSimpleTransaction(transaction *transfer.TransactionIdentity, timestamp int64, activityType Type, activityStatus Status) Entry {
	return newActivityEntryWithTransaction(false, transaction, timestamp, activityType, activityStatus)
}

func newActivityEntryWithTransaction(pending bool, transaction *transfer.TransactionIdentity, timestamp int64, activityType Type, activityStatus Status) Entry {
	payloadType := SimpleTransactionPT
	if pending {
		payloadType = PendingTransactionPT
	}

	return Entry{
		payloadType:    payloadType,
		transaction:    transaction,
		id:             0,
		timestamp:      timestamp,
		activityType:   activityType,
		activityStatus: activityStatus,
	}
}

func NewActivityEntryWithMultiTransaction(id transfer.MultiTransactionIDType, timestamp int64, activityType Type, activityStatus Status) Entry {
	return Entry{
		payloadType:    MultiTransactionPT,
		id:             id,
		timestamp:      timestamp,
		activityType:   activityType,
		activityStatus: activityStatus,
	}
}

func (e *Entry) PayloadType() PayloadType {
	return e.payloadType
}

func (e *Entry) isNFT() bool {
	tt := e.transferType
	return tt != nil && (*tt == TransferTypeErc721 || *tt == TransferTypeErc1155) && ((e.tokenIn != nil && e.tokenIn.TokenID != nil) || (e.tokenOut != nil && e.tokenOut.TokenID != nil))
}

func tokenIDToWalletBigInt(tokenID *hexutil.Big) *bigint.BigInt {
	if tokenID == nil {
		return nil
	}

	bi := new(big.Int).Set((*big.Int)(tokenID))
	return &bigint.BigInt{Int: bi}
}

func (e *Entry) anyIdentity() *thirdparty.CollectibleUniqueID {
	if e.tokenIn != nil {
		return &thirdparty.CollectibleUniqueID{
			ContractID: thirdparty.ContractID{
				ChainID: e.tokenIn.ChainID,
				Address: e.tokenIn.Address,
			},
			TokenID: tokenIDToWalletBigInt(e.tokenIn.TokenID),
		}
	} else if e.tokenOut != nil {
		return &thirdparty.CollectibleUniqueID{
			ContractID: thirdparty.ContractID{
				ChainID: e.tokenOut.ChainID,
				Address: e.tokenOut.Address,
			},
			TokenID: tokenIDToWalletBigInt(e.tokenOut.TokenID),
		}
	}
	return nil
}

func (e *Entry) getIdentity() EntryIdentity {
	return EntryIdentity{
		payloadType: e.payloadType,
		id:          e.id,
		transaction: e.transaction,
	}
}

func multiTransactionTypeToActivityType(mtType transfer.MultiTransactionType) Type {
	if mtType == transfer.MultiTransactionSend {
		return SendAT
	} else if mtType == transfer.MultiTransactionSwap {
		return SwapAT
	} else if mtType == transfer.MultiTransactionBridge {
		return BridgeAT
	}
	panic("unknown multi transaction type")
}

func sliceContains[T constraints.Ordered](slice []T, item T) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func sliceChecksCondition[T any](slice []T, condition func(*T) bool) bool {
	for i := range slice {
		if condition(&slice[i]) {
			return true
		}
	}
	return false
}

func joinItems[T interface{}](items []T, itemConversion func(T) string) string {
	if len(items) == 0 {
		return ""
	}
	var sb strings.Builder
	if itemConversion == nil {
		itemConversion = func(item T) string {
			return fmt.Sprintf("%v", item)
		}
	}
	for i, item := range items {
		if i == 0 {
			sb.WriteString("(")
		} else {
			sb.WriteString("),(")
		}
		sb.WriteString(itemConversion(item))
	}
	sb.WriteString(")")

	return sb.String()
}

func joinAddresses(addresses []eth.Address) string {
	return joinItems(addresses, func(a eth.Address) string {
		return fmt.Sprintf("X'%s'", hex.EncodeToString(a[:]))
	})
}

func activityTypesToMultiTransactionTypes(trTypes []Type) []transfer.MultiTransactionType {
	mtTypes := make([]transfer.MultiTransactionType, 0, len(trTypes))
	for _, t := range trTypes {
		var mtType transfer.MultiTransactionType
		if t == SendAT {
			mtType = transfer.MultiTransactionSend
		} else if t == SwapAT {
			mtType = transfer.MultiTransactionSwap
		} else if t == BridgeAT {
			mtType = transfer.MultiTransactionBridge
		} else {
			continue
		}
		mtTypes = append(mtTypes, mtType)
	}
	return mtTypes
}

const (
	fromTrType = byte(1)
	toTrType   = byte(2)

	noEntriesInTmpTableSQLValues             = "(NULL)"
	noEntriesInTwoColumnsTmpTableSQLValues   = "(NULL, NULL)"
	noEntriesInThreeColumnsTmpTableSQLValues = "(NULL, NULL, NULL)"
)

//go:embed filter.sql
var queryFormatString string
var mintATQuery = "SELECT hash FROM input_data WHERE method IN ('mint', 'mintToken')"

type FilterDependencies struct {
	db *sql.DB
	// use token.TokenType, token.ChainID and token.Address to find the available symbol
	tokenSymbol func(token Token) string
	// use the chainID and symbol to look up token.TokenType and token.Address. Return nil if not found
	tokenFromSymbol func(chainID *common.ChainID, symbol string) *Token
	// use to get current timestamp
	currentTimestamp func() int64
}

// getActivityEntries queries the transfers, pending_transactions, and multi_transactions tables based on filter parameters and arguments
// it returns metadata for all entries ordered by timestamp column
//
// addresses are mandatory and used to detect activity types SendAT and ReceiveAT for transfers entries
//
// allAddresses optimization indicates if the passed addresses include all the owners in the wallet DB
//
// Adding a no-limit option was never considered or required.
func getActivityEntries(ctx context.Context, deps FilterDependencies, addresses []eth.Address, allAddresses bool, chainIDs []common.ChainID, filter Filter, offset int, limit int) ([]Entry, error) {
	if len(addresses) == 0 {
		return nil, errors.New("no addresses provided")
	}

	includeAllTokenTypeAssets := len(filter.Assets) == 0 && !filter.FilterOutAssets

	// Used for symbol bearing tables multi_transactions and pending_transactions
	assetsTokenCodes := noEntriesInTmpTableSQLValues
	// Used for identity bearing tables transfers
	assetsERC20 := noEntriesInTwoColumnsTmpTableSQLValues
	if !includeAllTokenTypeAssets && !filter.FilterOutAssets {
		symbolsSet := make(map[string]struct{})
		var symbols []string
		for _, item := range filter.Assets {
			symbol := deps.tokenSymbol(item)
			if _, ok := symbolsSet[symbol]; !ok {
				symbols = append(symbols, symbol)
				symbolsSet[symbol] = struct{}{}
			}
		}
		assetsTokenCodes = joinItems(symbols, func(s string) string {
			return fmt.Sprintf("'%s'", s)
		})

		if sliceChecksCondition(filter.Assets, func(item *Token) bool { return item.TokenType == Erc20 }) {
			assetsERC20 = joinItems(filter.Assets, func(item Token) string {
				if item.TokenType == Erc20 {
					return fmt.Sprintf("%d, X'%s'", item.ChainID, item.Address.Hex()[2:])
				}
				return ""
			})
		}
	}

	includeAllCollectibles := len(filter.Collectibles) == 0 && !filter.FilterOutCollectibles
	assetsERC721 := noEntriesInThreeColumnsTmpTableSQLValues
	if !includeAllCollectibles && !filter.FilterOutCollectibles {
		assetsERC721 = joinItems(filter.Collectibles, func(item Token) string {
			tokenID := item.TokenID.String()[2:]
			address := item.Address.Hex()[2:]
			// SQLite mandates that byte length is an even number which hexutil.EncodeBig doesn't guarantee
			if len(tokenID)%2 == 1 {
				tokenID = "0" + tokenID
			}
			return fmt.Sprintf("%d, X'%s', X'%s'", item.ChainID, tokenID, address)
		})
	}

	// construct chain IDs
	includeAllNetworks := len(chainIDs) == 0
	networks := noEntriesInTmpTableSQLValues
	if !includeAllNetworks {
		networks = joinItems(chainIDs, nil)
	}

	layer2Chains := []uint64{common.OptimismMainnet, common.OptimismGoerli, common.ArbitrumMainnet, common.ArbitrumGoerli}
	layer2Networks := joinItems(layer2Chains, func(chainID uint64) string {
		return fmt.Sprintf("%d", chainID)
	})

	startFilterDisabled := !(filter.Period.StartTimestamp > 0)
	endFilterDisabled := !(filter.Period.EndTimestamp > 0)
	filterActivityTypeAll := len(filter.Types) == 0
	filterAllToAddresses := len(filter.CounterpartyAddresses) == 0
	includeAllStatuses := len(filter.Statuses) == 0

	filterStatusPending := false
	filterStatusCompleted := false
	filterStatusFailed := false
	filterStatusFinalized := false
	if !includeAllStatuses {
		filterStatusPending = sliceContains(filter.Statuses, PendingAS)
		filterStatusCompleted = sliceContains(filter.Statuses, CompleteAS)
		filterStatusFailed = sliceContains(filter.Statuses, FailedAS)
		filterStatusFinalized = sliceContains(filter.Statuses, FinalizedAS)
	}

	involvedAddresses := joinAddresses(addresses)
	toAddresses := noEntriesInTmpTableSQLValues
	if !filterAllToAddresses {
		toAddresses = joinAddresses(filter.CounterpartyAddresses)
	}

	mtTypes := activityTypesToMultiTransactionTypes(filter.Types)
	joinedMTTypes := joinItems(mtTypes, func(t transfer.MultiTransactionType) string {
		return strconv.Itoa(int(t))
	})

	inputDataMethods := make([]string, 0)

	if includeAllStatuses || sliceContains(filter.Types, MintAT) || sliceContains(filter.Types, ReceiveAT) {
		inputDataRows, err := deps.db.QueryContext(ctx, mintATQuery)

		if err != nil {
			return nil, err
		}

		for inputDataRows.Next() {
			var inputData sql.NullString
			err := inputDataRows.Scan(&inputData)
			if err == nil && inputData.Valid {
				inputDataMethods = append(inputDataMethods, inputData.String)
			}
		}
	}

	queryString := fmt.Sprintf(queryFormatString, involvedAddresses, toAddresses, assetsTokenCodes, assetsERC20, assetsERC721, networks,
		layer2Networks, mintATQuery, joinedMTTypes)

	// The duplicated temporary table UNION with CTE acts as an optimization
	// As soon as we use filter_addresses CTE or filter_addresses_table temp table
	// or switch them alternatively for JOIN or IN clauses the performance drops significantly
	_, err := deps.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS filter_addresses_table; CREATE TEMP TABLE filter_addresses_table (address VARCHAR PRIMARY KEY); INSERT INTO filter_addresses_table (address) VALUES %s;\n", involvedAddresses))
	if err != nil {
		return nil, err
	}

	rows, err := deps.db.QueryContext(ctx, queryString,
		startFilterDisabled, filter.Period.StartTimestamp, endFilterDisabled, filter.Period.EndTimestamp,
		filterActivityTypeAll, sliceContains(filter.Types, SendAT), sliceContains(filter.Types, ReceiveAT),
		sliceContains(filter.Types, ContractDeploymentAT), sliceContains(filter.Types, MintAT),
		transfer.MultiTransactionSend,
		fromTrType, toTrType,
		allAddresses, filterAllToAddresses,
		includeAllStatuses, filterStatusCompleted, filterStatusFailed, filterStatusFinalized, filterStatusPending,
		FailedAS, CompleteAS, FinalizedAS, PendingAS,
		includeAllTokenTypeAssets,
		includeAllCollectibles,
		includeAllNetworks,
		transactions.Pending,
		deps.currentTimestamp(),
		648000, // 7.5 days in seconds for layer 2 finalization. 0.5 day is buffer to not create false positive.
		960,    // A block on layer 1 is every 12s, finalization require 64 blocks. A buffer of 16 blocks is added to not create false positives.
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var transferHash, pendingHash []byte
		var chainID, outChainIDDB, inChainIDDB, multiTxID, aggregatedCount sql.NullInt64
		var timestamp int64
		var dbMtType, dbTrType sql.NullByte
		var toAddress, fromAddress eth.Address
		var toAddressDB, ownerAddressDB, contractAddressDB, dbTokenID sql.RawBytes
		var tokenAddress, contractAddress *eth.Address
		var aggregatedStatus int
		var dbTrAmount sql.NullString
		dbPTrAmount := new(big.Int)
		var dbMtFromAmount, dbMtToAmount, contractType sql.NullString
		var tokenCode, fromTokenCode, toTokenCode sql.NullString
		var methodHash sql.NullString
		var transferType *TransferType
		var communityMintEventDB sql.NullBool
		var communityMintEvent bool
		err := rows.Scan(&transferHash, &pendingHash, &chainID, &multiTxID, &timestamp, &dbMtType, &dbTrType, &fromAddress,
			&toAddressDB, &ownerAddressDB, &dbTrAmount, (*bigint.SQLBigIntBytes)(dbPTrAmount), &dbMtFromAmount, &dbMtToAmount, &aggregatedStatus, &aggregatedCount,
			&tokenAddress, &dbTokenID, &tokenCode, &fromTokenCode, &toTokenCode, &outChainIDDB, &inChainIDDB, &contractType,
			&contractAddressDB, &methodHash, &communityMintEventDB)
		if err != nil {
			return nil, err
		}

		if len(toAddressDB) > 0 {
			toAddress = eth.BytesToAddress(toAddressDB)
		}

		if contractType.Valid {
			transferType = contractTypeFromDBType(contractType.String)
		}

		if communityMintEventDB.Valid {
			communityMintEvent = communityMintEventDB.Bool
		}

		if len(contractAddressDB) > 0 {
			contractAddress = new(eth.Address)
			*contractAddress = eth.BytesToAddress(contractAddressDB)
		}

		getActivityType := func(trType sql.NullByte) (activityType Type, filteredAddress eth.Address) {
			if trType.Valid {
				if trType.Byte == fromTrType {
					if toAddress == ZeroAddress && transferType != nil && *transferType == TransferTypeEth && contractAddress != nil && *contractAddress != ZeroAddress {
						return ContractDeploymentAT, fromAddress
					}
					return SendAT, fromAddress
				} else if trType.Byte == toTrType {
					at := ReceiveAT
					if fromAddress == ZeroAddress && transferType != nil {
						if *transferType == TransferTypeErc721 || (*transferType == TransferTypeErc20 && methodHash.Valid && (communityMintEvent || sliceContains(inputDataMethods, methodHash.String))) {
							at = MintAT
						}
					}
					return at, toAddress
				}
			}
			log.Warn(fmt.Sprintf("unexpected activity type. Missing from [%s] or to [%s] in addresses?", fromAddress, toAddress))
			return ReceiveAT, toAddress
		}

		// Can be mapped directly because the values are injected into the query
		activityStatus := Status(aggregatedStatus)
		var outChainID, inChainID *common.ChainID
		var entry Entry
		var tokenID *hexutil.Big
		if len(dbTokenID) > 0 {
			tokenID = (*hexutil.Big)(new(big.Int).SetBytes(dbTokenID))
		}

		if transferHash != nil && chainID.Valid {
			// Process `transfers` row

			// Extract activity type: SendAT/ReceiveAT
			activityType, _ := getActivityType(dbTrType)

			ownerAddress := eth.BytesToAddress(ownerAddressDB)
			inAmount, outAmount := getTrInAndOutAmounts(activityType, dbTrAmount, dbPTrAmount)

			// Extract tokens and chains
			var involvedToken *Token
			if tokenAddress != nil && *tokenAddress != ZeroAddress {
				involvedToken = &Token{TokenType: Erc20, ChainID: common.ChainID(chainID.Int64), TokenID: tokenID, Address: *tokenAddress}
			} else {
				involvedToken = &Token{TokenType: Native, ChainID: common.ChainID(chainID.Int64), TokenID: tokenID}
			}

			entry = newActivityEntryWithSimpleTransaction(
				&transfer.TransactionIdentity{ChainID: common.ChainID(chainID.Int64),
					Hash:    eth.BytesToHash(transferHash),
					Address: ownerAddress,
				},
				timestamp, activityType, activityStatus,
			)

			// Extract tokens
			if activityType == SendAT || activityType == ContractDeploymentAT {
				entry.tokenOut = involvedToken
				outChainID = new(common.ChainID)
				*outChainID = common.ChainID(chainID.Int64)
			} else {
				entry.tokenIn = involvedToken
				inChainID = new(common.ChainID)
				*inChainID = common.ChainID(chainID.Int64)
			}

			entry.symbolOut, entry.symbolIn = lookupAndFillInTokens(deps, entry.tokenOut, entry.tokenIn)

			// Complete the data
			entry.amountOut = outAmount
			entry.amountIn = inAmount
		} else if pendingHash != nil && chainID.Valid {
			// Process `pending_transactions` row

			// Extract activity type: SendAT/ReceiveAT
			activityType, _ := getActivityType(dbTrType)

			inAmount, outAmount := getTrInAndOutAmounts(activityType, dbTrAmount, dbPTrAmount)

			outChainID = new(common.ChainID)
			*outChainID = common.ChainID(chainID.Int64)

			entry = newActivityEntryWithPendingTransaction(
				&transfer.TransactionIdentity{ChainID: common.ChainID(chainID.Int64),
					Hash: eth.BytesToHash(pendingHash),
				},
				timestamp, activityType, activityStatus,
			)

			// Extract tokens
			if tokenCode.Valid {
				cID := common.ChainID(chainID.Int64)
				entry.tokenOut = deps.tokenFromSymbol(&cID, tokenCode.String)
			}
			entry.symbolOut, entry.symbolIn = lookupAndFillInTokens(deps, entry.tokenOut, nil)

			// Complete the data
			entry.amountOut = outAmount
			entry.amountIn = inAmount

		} else if multiTxID.Valid {
			// Process `multi_transactions` row

			mtInAmount, mtOutAmount := getMtInAndOutAmounts(dbMtFromAmount, dbMtToAmount)

			// Extract activity type: SendAT/SwapAT/BridgeAT
			activityType := multiTransactionTypeToActivityType(transfer.MultiTransactionType(dbMtType.Byte))

			if outChainIDDB.Valid && outChainIDDB.Int64 != 0 {
				outChainID = new(common.ChainID)
				*outChainID = common.ChainID(outChainIDDB.Int64)
			}
			if inChainIDDB.Valid && inChainIDDB.Int64 != 0 {
				inChainID = new(common.ChainID)
				*inChainID = common.ChainID(inChainIDDB.Int64)
			}

			entry = NewActivityEntryWithMultiTransaction(transfer.MultiTransactionIDType(multiTxID.Int64),
				timestamp, activityType, activityStatus)

			// Extract tokens
			if fromTokenCode.Valid {
				entry.tokenOut = deps.tokenFromSymbol(outChainID, fromTokenCode.String)
				entry.symbolOut = common.NewAndSet(fromTokenCode.String)
			}
			if toTokenCode.Valid {
				entry.tokenIn = deps.tokenFromSymbol(inChainID, toTokenCode.String)
				entry.symbolIn = common.NewAndSet(toTokenCode.String)
			}

			// Complete the data
			entry.amountOut = mtOutAmount
			entry.amountIn = mtInAmount
		} else {
			return nil, errors.New("invalid row data")
		}

		// Complete common data
		entry.recipient = &toAddress
		entry.sender = &fromAddress
		entry.recipient = &toAddress
		entry.chainIDOut = outChainID
		entry.chainIDIn = inChainID
		entry.transferType = transferType

		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func getTrInAndOutAmounts(activityType Type, trAmount sql.NullString, pTrAmount *big.Int) (inAmount *hexutil.Big, outAmount *hexutil.Big) {
	var amount *big.Int
	ok := false
	if trAmount.Valid {
		amount, ok = new(big.Int).SetString(trAmount.String, 16)
	} else if pTrAmount != nil {
		// Process pending transaction value
		amount = pTrAmount
		ok = true
	} else {
		log.Warn(fmt.Sprintf("invalid transaction amount for type %d", activityType))
	}

	if ok {
		switch activityType {
		case ContractDeploymentAT:
			fallthrough
		case SendAT:
			inAmount = (*hexutil.Big)(big.NewInt(0))
			outAmount = (*hexutil.Big)(amount)
			return
		case MintAT:
			fallthrough
		case ReceiveAT:
			inAmount = (*hexutil.Big)(amount)
			outAmount = (*hexutil.Big)(big.NewInt(0))
			return
		default:
			log.Warn(fmt.Sprintf("unexpected activity type %d", activityType))
		}
	} else {
		log.Warn(fmt.Sprintf("could not parse amount %s", trAmount.String))
	}

	inAmount = (*hexutil.Big)(big.NewInt(0))
	outAmount = (*hexutil.Big)(big.NewInt(0))
	return
}

func getMtInAndOutAmounts(dbFromAmount sql.NullString, dbToAmount sql.NullString) (inAmount *hexutil.Big, outAmount *hexutil.Big) {
	if dbFromAmount.Valid && dbToAmount.Valid {
		fromHexStr := dbFromAmount.String
		toHexStr := dbToAmount.String
		if len(fromHexStr) > 2 && len(toHexStr) > 2 {
			fromAmount, frOk := new(big.Int).SetString(dbFromAmount.String[2:], 16)
			toAmount, toOk := new(big.Int).SetString(dbToAmount.String[2:], 16)
			if frOk && toOk {
				inAmount = (*hexutil.Big)(toAmount)
				outAmount = (*hexutil.Big)(fromAmount)
				return
			}
		}
		log.Warn(fmt.Sprintf("could not parse amounts %s %s", fromHexStr, toHexStr))
	} else {
		log.Warn("invalid transaction amounts")
	}
	inAmount = (*hexutil.Big)(big.NewInt(0))
	outAmount = (*hexutil.Big)(big.NewInt(0))
	return
}

func contractTypeFromDBType(dbType string) (transferType *TransferType) {
	transferType = new(TransferType)
	switch common.Type(dbType) {
	case common.EthTransfer:
		*transferType = TransferTypeEth
	case common.Erc20Transfer:
		*transferType = TransferTypeErc20
	case common.Erc721Transfer:
		*transferType = TransferTypeErc721
	default:
		return nil
	}
	return transferType
}

// lookupAndFillInTokens ignores NFTs
func lookupAndFillInTokens(deps FilterDependencies, tokenOut *Token, tokenIn *Token) (symbolOut *string, symbolIn *string) {
	if tokenOut != nil && tokenOut.TokenID == nil {
		symbol := deps.tokenSymbol(*tokenOut)
		if len(symbol) > 0 {
			symbolOut = common.NewAndSet(symbol)
		}
	}
	if tokenIn != nil && tokenIn.TokenID == nil {
		symbol := deps.tokenSymbol(*tokenIn)
		if len(symbol) > 0 {
			symbolIn = common.NewAndSet(symbol)
		}
	}
	return symbolOut, symbolIn
}
