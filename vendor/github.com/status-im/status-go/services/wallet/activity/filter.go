package activity

import (
	"context"
	"database/sql"
	"fmt"

	// used for embedding the sql query in the binary
	_ "embed"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/transactions"
)

const NoLimitTimestampForPeriod = 0

//go:embed get_collectibles.sql
var getCollectiblesQueryFormatString string

//go:embed oldest_timestamp.sql
var oldestTimestampQueryFormatString string

//go:embed recipients.sql
var recipientsQueryFormatString string

type Period struct {
	StartTimestamp int64 `json:"startTimestamp"`
	EndTimestamp   int64 `json:"endTimestamp"`
}

type Type int

const (
	SendAT Type = iota
	ReceiveAT
	BuyAT
	SwapAT
	BridgeAT
	ContractDeploymentAT
	MintAT
)

func allActivityTypesFilter() []Type {
	return []Type{}
}

type Status int

const (
	FailedAS    Status = iota // failed status or at least one failed transaction for multi-transactions
	PendingAS                 // in pending DB or at least one transaction in pending for multi-transactions
	CompleteAS                // success status
	FinalizedAS               // all multi-transactions have success status
)

func allActivityStatusesFilter() []Status {
	return []Status{}
}

type TokenType int

const (
	Native TokenType = iota
	Erc20
	Erc721
	Erc1155
)

// Token supports all tokens. Some fields might be optional, depending on the TokenType
type Token struct {
	TokenType TokenType `json:"tokenType"`
	// ChainID is used for TokenType.Native only to lookup the symbol, all chains will be included in the token filter
	ChainID common.ChainID `json:"chainId"`
	Address eth.Address    `json:"address,omitempty"`
	TokenID *hexutil.Big   `json:"tokenId,omitempty"`
}

func allTokensFilter() []Token {
	return []Token{}
}

func allNetworksFilter() []common.ChainID {
	return []common.ChainID{}
}

type Filter struct {
	Period                Period        `json:"period"`
	Types                 []Type        `json:"types"`
	Statuses              []Status      `json:"statuses"`
	CounterpartyAddresses []eth.Address `json:"counterpartyAddresses"`

	// Tokens
	Assets                []Token `json:"assets"`
	Collectibles          []Token `json:"collectibles"`
	FilterOutAssets       bool    `json:"filterOutAssets"`
	FilterOutCollectibles bool    `json:"filterOutCollectibles"`
}

func GetRecipients(ctx context.Context, db *sql.DB, chainIDs []common.ChainID, addresses []eth.Address, offset int, limit int) (recipients []eth.Address, hasMore bool, err error) {
	filterAllAddresses := len(addresses) == 0
	involvedAddresses := noEntriesInTmpTableSQLValues
	if !filterAllAddresses {
		involvedAddresses = joinAddresses(addresses)
	}

	includeAllNetworks := len(chainIDs) == 0
	networks := noEntriesInTmpTableSQLValues
	if !includeAllNetworks {
		networks = joinItems(chainIDs, nil)
	}

	queryString := fmt.Sprintf(recipientsQueryFormatString, involvedAddresses, networks)

	rows, err := db.QueryContext(ctx, queryString, filterAllAddresses, includeAllNetworks, transactions.Pending, limit, offset)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var entries []eth.Address
	for rows.Next() {
		var toAddress eth.Address
		var timestamp int64
		err := rows.Scan(&toAddress, &timestamp)
		if err != nil {
			return nil, false, err
		}
		entries = append(entries, toAddress)
	}

	if err = rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore = len(entries) == limit

	return entries, hasMore, nil
}

func GetOldestTimestamp(ctx context.Context, db *sql.DB, addresses []eth.Address) (timestamp int64, err error) {
	filterAllAddresses := len(addresses) == 0
	involvedAddresses := noEntriesInTmpTableSQLValues
	if !filterAllAddresses {
		involvedAddresses = joinAddresses(addresses)
	}

	queryString := fmt.Sprintf(oldestTimestampQueryFormatString, involvedAddresses)

	row := db.QueryRowContext(ctx, queryString, filterAllAddresses)
	var fromAddress, toAddress sql.NullString
	err = row.Scan(&fromAddress, &toAddress, &timestamp)
	if err == sql.ErrNoRows {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return timestamp, nil
}

func GetActivityCollectibles(ctx context.Context, db *sql.DB, chainIDs []common.ChainID, owners []eth.Address, offset int, limit int) ([]thirdparty.CollectibleUniqueID, error) {
	filterAllAddresses := len(owners) == 0
	involvedAddresses := noEntriesInTmpTableSQLValues
	if !filterAllAddresses {
		involvedAddresses = joinAddresses(owners)
	}

	includeAllNetworks := len(chainIDs) == 0
	networks := noEntriesInTmpTableSQLValues
	if !includeAllNetworks {
		networks = joinItems(chainIDs, nil)
	}

	queryString := fmt.Sprintf(getCollectiblesQueryFormatString, involvedAddresses, networks)

	rows, err := db.QueryContext(ctx, queryString, filterAllAddresses, includeAllNetworks, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return thirdparty.RowsToCollectibles(rows)
}
