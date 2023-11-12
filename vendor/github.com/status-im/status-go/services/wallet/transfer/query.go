package transfer

import (
	"bytes"
	"database/sql"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/status-im/status-go/services/wallet/bigint"
	w_common "github.com/status-im/status-go/services/wallet/common"
)

const baseTransfersQuery = "SELECT hash, type, blk_hash, blk_number, timestamp, address, tx, sender, receipt, log, network_id, base_gas_fee, COALESCE(multi_transaction_id, 0) %s FROM transfers"
const preloadedTransfersQuery = "SELECT hash, type, address, log, token_id, amount_padded128hex FROM transfers"

type transfersQuery struct {
	buf        *bytes.Buffer
	args       []interface{}
	whereAdded bool
	subQuery   bool
}

func newTransfersQuery() *transfersQuery {
	newQuery := newEmptyQuery()
	transfersQueryString := fmt.Sprintf(baseTransfersQuery, "")
	newQuery.buf.WriteString(transfersQueryString)
	return newQuery
}

func newTransfersQueryForPreloadedTransactions() *transfersQuery {
	newQuery := newEmptyQuery()
	newQuery.buf.WriteString(preloadedTransfersQuery)
	return newQuery
}

func newSubQuery() *transfersQuery {
	newQuery := newEmptyQuery()
	newQuery.subQuery = true
	return newQuery
}

func newEmptyQuery() *transfersQuery {
	buf := bytes.NewBuffer(nil)
	return &transfersQuery{buf: buf}
}

func (q *transfersQuery) addWhereSeparator(separator SeparatorType) {
	if !q.whereAdded {
		if !q.subQuery {
			q.buf.WriteString(" WHERE")
		}
		q.whereAdded = true
	} else if separator == OrSeparator {
		q.buf.WriteString(" OR")
	} else if separator == AndSeparator {
		q.buf.WriteString(" AND")
	} else if separator != NoSeparator {
		panic("Unknown separator. Need to handle current SeparatorType value")
	}
}

type SeparatorType int

// Beware: please update addWhereSeparator if changing this enum
const (
	NoSeparator SeparatorType = iota + 1
	OrSeparator
	AndSeparator
)

// addSubQuery adds where clause formed as: WHERE/<separator> (<subQuery>)
func (q *transfersQuery) addSubQuery(subQuery *transfersQuery, separator SeparatorType) *transfersQuery {
	q.addWhereSeparator(separator)
	q.buf.WriteString(" (")
	q.buf.Write(subQuery.buf.Bytes())
	q.buf.WriteString(")")
	q.args = append(q.args, subQuery.args...)
	return q
}

func (q *transfersQuery) FilterStart(start *big.Int) *transfersQuery {
	if start != nil {
		q.addWhereSeparator(AndSeparator)
		q.buf.WriteString(" blk_number >= ?")
		q.args = append(q.args, (*bigint.SQLBigInt)(start))
	}
	return q
}

func (q *transfersQuery) FilterEnd(end *big.Int) *transfersQuery {
	if end != nil {
		q.addWhereSeparator(AndSeparator)
		q.buf.WriteString(" blk_number <= ?")
		q.args = append(q.args, (*bigint.SQLBigInt)(end))
	}
	return q
}

func (q *transfersQuery) FilterLoaded(loaded int) *transfersQuery {
	q.addWhereSeparator(AndSeparator)
	q.buf.WriteString(" loaded = ? ")
	q.args = append(q.args, loaded)

	return q
}

func (q *transfersQuery) FilterNetwork(network uint64) *transfersQuery {
	q.addWhereSeparator(AndSeparator)
	q.buf.WriteString(" network_id = ?")
	q.args = append(q.args, network)
	return q
}

func (q *transfersQuery) FilterAddress(address common.Address) *transfersQuery {
	q.addWhereSeparator(AndSeparator)
	q.buf.WriteString(" address = ?")
	q.args = append(q.args, address)
	return q
}

func (q *transfersQuery) FilterTransactionID(hash common.Hash) *transfersQuery {
	q.addWhereSeparator(AndSeparator)
	q.buf.WriteString(" hash = ?")
	q.args = append(q.args, hash)
	return q
}

func (q *transfersQuery) FilterTransactionHash(hash common.Hash) *transfersQuery {
	q.addWhereSeparator(AndSeparator)
	q.buf.WriteString(" tx_hash = ?")
	q.args = append(q.args, hash)
	return q
}

func (q *transfersQuery) FilterBlockHash(blockHash common.Hash) *transfersQuery {
	q.addWhereSeparator(AndSeparator)
	q.buf.WriteString(" blk_hash = ?")
	q.args = append(q.args, blockHash)
	return q
}

func (q *transfersQuery) FilterBlockNumber(blockNumber *big.Int) *transfersQuery {
	q.addWhereSeparator(AndSeparator)
	q.buf.WriteString(" blk_number = ?")
	q.args = append(q.args, (*bigint.SQLBigInt)(blockNumber))
	return q
}

func ascendingString(ascending bool) string {
	if ascending {
		return "ASC"
	}
	return "DESC"
}

func (q *transfersQuery) SortByBlockNumberAndHash() *transfersQuery {
	q.buf.WriteString(" ORDER BY blk_number DESC, hash ASC ")
	return q
}

func (q *transfersQuery) SortByTimestamp(ascending bool) *transfersQuery {
	q.buf.WriteString(fmt.Sprintf(" ORDER BY timestamp %s ", ascendingString(ascending)))
	return q
}

func (q *transfersQuery) Limit(pageSize int64) *transfersQuery {
	q.buf.WriteString(" LIMIT ?")
	q.args = append(q.args, pageSize)
	return q
}

func (q *transfersQuery) FilterType(txType w_common.Type) *transfersQuery {
	q.addWhereSeparator(AndSeparator)
	q.buf.WriteString(" type = ?")
	q.args = append(q.args, txType)
	return q
}

func (q *transfersQuery) FilterTokenAddress(address common.Address) *transfersQuery {
	q.addWhereSeparator(AndSeparator)
	q.buf.WriteString(" token_address = ?")
	q.args = append(q.args, address)
	return q
}

func (q *transfersQuery) FilterTokenID(tokenID *big.Int) *transfersQuery {
	q.addWhereSeparator(AndSeparator)
	q.buf.WriteString(" token_id = ?")
	q.args = append(q.args, (*bigint.SQLBigIntBytes)(tokenID))
	return q
}

func (q *transfersQuery) String() string {
	return q.buf.String()
}

func (q *transfersQuery) Args() []interface{} {
	return q.args
}

func (q *transfersQuery) TransferScan(rows *sql.Rows) (rst []Transfer, err error) {
	for rows.Next() {
		transfer := Transfer{
			BlockNumber: &big.Int{},
			Transaction: &types.Transaction{},
			Receipt:     &types.Receipt{},
			Log:         &types.Log{},
		}
		err = rows.Scan(
			&transfer.ID, &transfer.Type, &transfer.BlockHash,
			(*bigint.SQLBigInt)(transfer.BlockNumber), &transfer.Timestamp, &transfer.Address,
			&JSONBlob{transfer.Transaction}, &transfer.From, &JSONBlob{transfer.Receipt}, &JSONBlob{transfer.Log}, &transfer.NetworkID, &transfer.BaseGasFees, &transfer.MultiTransactionID)
		if err != nil {
			return nil, err
		}
		rst = append(rst, transfer)
	}

	return rst, nil
}

func (q *transfersQuery) PreloadedTransactionScan(rows *sql.Rows) (rst []*PreloadedTransaction, err error) {
	transfers := make([]Transfer, 0)
	for rows.Next() {
		transfer := Transfer{
			Log: &types.Log{},
		}
		tokenValue := sql.NullString{}
		tokenID := sql.RawBytes{}
		err = rows.Scan(
			&transfer.ID, &transfer.Type,
			&transfer.Address,
			&JSONBlob{transfer.Log},
			&tokenID, &tokenValue)

		if len(tokenID) > 0 {
			transfer.TokenID = new(big.Int).SetBytes(tokenID)
		}

		if tokenValue.Valid {
			var ok bool
			transfer.TokenValue, ok = new(big.Int).SetString(tokenValue.String, 16)
			if !ok {
				panic("failed to parse token value")
			}
		}

		if err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer)
	}

	rst = make([]*PreloadedTransaction, 0, len(transfers))

	for _, transfer := range transfers {
		preloadedTransaction := &PreloadedTransaction{
			ID:      transfer.ID,
			Type:    transfer.Type,
			Address: transfer.Address,
			Log:     transfer.Log,
			TokenID: transfer.TokenID,
			Value:   transfer.TokenValue,
		}

		rst = append(rst, preloadedTransaction)
	}

	return rst, nil
}
