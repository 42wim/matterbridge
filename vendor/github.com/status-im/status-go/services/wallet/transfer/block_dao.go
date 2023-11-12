package transfer

import (
	"database/sql"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/services/wallet/bigint"
)

type BlocksRange struct {
	from *big.Int
	to   *big.Int
}

type Block struct {
	Number  *big.Int
	Balance *big.Int
	Nonce   *int64
}

type BlockView struct {
	Address common.Address `json:"address"`
	Number  *big.Int       `json:"blockNumber"`
	Balance bigint.BigInt  `json:"balance"`
	Nonce   *int64         `json:"nonce"`
}

func blocksToViews(blocks map[common.Address]*Block) []BlockView {
	blocksViews := []BlockView{}
	for address, block := range blocks {
		view := BlockView{
			Address: address,
			Number:  block.Number,
			Balance: bigint.BigInt{Int: block.Balance},
			Nonce:   block.Nonce,
		}
		blocksViews = append(blocksViews, view)
	}

	return blocksViews
}

type BlockDAO struct {
	db *sql.DB
}

// MergeBlocksRanges merge old blocks ranges if possible
func (b *BlockDAO) mergeBlocksRanges(chainIDs []uint64, accounts []common.Address) error {
	for _, chainID := range chainIDs {
		for _, account := range accounts {
			err := b.mergeRanges(chainID, account)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *BlockDAO) mergeRanges(chainID uint64, account common.Address) (err error) {
	var (
		tx *sql.Tx
	)

	ranges, err := b.getOldRanges(chainID, account)
	if err != nil {
		return err
	}

	log.Info("merge old ranges", "account", account, "network", chainID, "ranges", len(ranges))

	if len(ranges) <= 1 {
		return nil
	}

	tx, err = b.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	newRanges, deletedRanges := getNewRanges(ranges)

	for _, rangeToDelete := range deletedRanges {
		err = deleteRange(chainID, tx, account, rangeToDelete.from, rangeToDelete.to)
		if err != nil {
			return err
		}
	}

	for _, newRange := range newRanges {
		err = insertRange(chainID, tx, account, newRange.from, newRange.to)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *BlockDAO) insertRange(chainID uint64, account common.Address, from, to, balance *big.Int, nonce uint64) error {
	log.Debug("insert blocks range", "account", account, "network id", chainID, "from", from, "to", to, "balance", balance, "nonce", nonce)
	insert, err := b.db.Prepare("INSERT INTO blocks_ranges (network_id, address, blk_from, blk_to, balance, nonce) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = insert.Exec(chainID, account, (*bigint.SQLBigInt)(from), (*bigint.SQLBigInt)(to), (*bigint.SQLBigIntBytes)(balance), &nonce)
	return err
}

func (b *BlockDAO) getOldRanges(chainID uint64, account common.Address) ([]*BlocksRange, error) {
	query := `select blk_from, blk_to from blocks_ranges
	          where address = ?
	          and network_id = ?
	          order by blk_from`

	rows, err := b.db.Query(query, account, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ranges := []*BlocksRange{}
	for rows.Next() {
		from := &big.Int{}
		to := &big.Int{}
		err = rows.Scan((*bigint.SQLBigInt)(from), (*bigint.SQLBigInt)(to))
		if err != nil {
			return nil, err
		}

		ranges = append(ranges, &BlocksRange{
			from: from,
			to:   to,
		})
	}

	return ranges, nil
}

// GetBlocksToLoadByAddress gets unloaded blocks for a given address.
func (b *BlockDAO) GetBlocksToLoadByAddress(chainID uint64, address common.Address, limit int) (rst []*big.Int, err error) {
	query := `SELECT blk_number FROM blocks
	WHERE address = ? AND network_id = ? AND loaded = 0
	ORDER BY blk_number DESC
	LIMIT ?`
	rows, err := b.db.Query(query, address, chainID, limit)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		block := &big.Int{}
		err = rows.Scan((*bigint.SQLBigInt)(block))
		if err != nil {
			return nil, err
		}
		rst = append(rst, block)
	}
	return rst, nil
}

func (b *BlockDAO) GetLastBlockByAddress(chainID uint64, address common.Address, limit int) (rst *big.Int, err error) {
	query := `SELECT * FROM
	(SELECT blk_number FROM blocks WHERE address = ? AND network_id = ? ORDER BY blk_number DESC LIMIT ?)
	ORDER BY blk_number LIMIT 1`
	rows, err := b.db.Query(query, address, chainID, limit)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		block := &big.Int{}
		err = rows.Scan((*bigint.SQLBigInt)(block))
		if err != nil {
			return nil, err
		}

		return block, nil
	}

	return nil, nil
}

func (b *BlockDAO) GetFirstSavedBlock(chainID uint64, address common.Address) (rst *DBHeader, err error) {
	query := `SELECT blk_number, blk_hash, loaded
	FROM blocks
	WHERE network_id = ? AND address = ?
	ORDER BY blk_number LIMIT 1`
	rows, err := b.db.Query(query, chainID, address)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		header := &DBHeader{Hash: common.Hash{}, Number: new(big.Int)}
		err = rows.Scan((*bigint.SQLBigInt)(header.Number), &header.Hash, &header.Loaded)
		if err != nil {
			return nil, err
		}

		return header, nil
	}

	return nil, nil
}

func (b *BlockDAO) GetFirstKnownBlock(chainID uint64, address common.Address) (rst *big.Int, err error) {
	query := `SELECT blk_from FROM blocks_ranges
	WHERE address = ?
	AND network_id = ?
	ORDER BY blk_from
	LIMIT 1`

	rows, err := b.db.Query(query, address, chainID)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		block := &big.Int{}
		err = rows.Scan((*bigint.SQLBigInt)(block))
		if err != nil {
			return nil, err
		}

		return block, nil
	}

	return nil, nil
}

func (b *BlockDAO) GetLastKnownBlockByAddress(chainID uint64, address common.Address) (block *Block, err error) {
	query := `SELECT blk_to, balance, nonce FROM blocks_ranges
	WHERE address = ?
	AND network_id = ?
	ORDER BY blk_to DESC
	LIMIT 1`

	rows, err := b.db.Query(query, address, chainID)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		var nonce sql.NullInt64
		block = &Block{Number: &big.Int{}, Balance: &big.Int{}}
		err = rows.Scan((*bigint.SQLBigInt)(block.Number), (*bigint.SQLBigIntBytes)(block.Balance), &nonce)
		if err != nil {
			return nil, err
		}

		if nonce.Valid {
			block.Nonce = &nonce.Int64
		}
		return block, nil
	}

	return nil, nil
}

func (b *BlockDAO) getLastKnownBlocks(chainID uint64, addresses []common.Address) (map[common.Address]*Block, error) {
	result := map[common.Address]*Block{}
	for _, address := range addresses {
		block, error := b.GetLastKnownBlockByAddress(chainID, address)
		if error != nil {
			return nil, error
		}

		if block != nil {
			result[address] = block
		}
	}

	return result, nil
}

// TODO Remove the method below, it is used in one place and duplicates getLastKnownBlocks method with slight unneeded change
func (b *BlockDAO) GetLastKnownBlockByAddresses(chainID uint64, addresses []common.Address) (map[common.Address]*Block, []common.Address, error) {
	res := map[common.Address]*Block{}
	accountsWithoutHistory := []common.Address{}
	for _, address := range addresses {
		block, err := b.GetLastKnownBlockByAddress(chainID, address)
		if err != nil {
			log.Info("Can't get last block", "error", err)
			return nil, nil, err
		}

		if block != nil {
			res[address] = block
		} else {
			accountsWithoutHistory = append(accountsWithoutHistory, address)
		}
	}

	return res, accountsWithoutHistory, nil
}

func getNewRanges(ranges []*BlocksRange) ([]*BlocksRange, []*BlocksRange) {
	initValue := big.NewInt(-1)
	prevFrom := big.NewInt(-1)
	prevTo := big.NewInt(-1)
	hasMergedRanges := false
	var newRanges []*BlocksRange
	var deletedRanges []*BlocksRange
	for idx, blocksRange := range ranges {
		if prevTo.Cmp(initValue) == 0 {
			prevTo = blocksRange.to
			prevFrom = blocksRange.from
		} else if prevTo.Cmp(blocksRange.from) >= 0 {
			hasMergedRanges = true
			deletedRanges = append(deletedRanges, ranges[idx-1])
			if prevTo.Cmp(blocksRange.to) <= 0 {
				prevTo = blocksRange.to
			}
		} else {
			if hasMergedRanges {
				deletedRanges = append(deletedRanges, ranges[idx-1])
				newRanges = append(newRanges, &BlocksRange{
					from: prevFrom,
					to:   prevTo,
				})
			}
			log.Info("blocks ranges gap detected", "from", prevTo, "to", blocksRange.from)
			hasMergedRanges = false

			prevFrom = blocksRange.from
			prevTo = blocksRange.to
		}
	}

	if hasMergedRanges {
		deletedRanges = append(deletedRanges, ranges[len(ranges)-1])
		newRanges = append(newRanges, &BlocksRange{
			from: prevFrom,
			to:   prevTo,
		})
	}

	return newRanges, deletedRanges
}

func deleteRange(chainID uint64, creator statementCreator, account common.Address, from *big.Int, to *big.Int) error {
	log.Info("delete blocks range", "account", account, "network", chainID, "from", from, "to", to)
	delete, err := creator.Prepare(`DELETE FROM blocks_ranges
                                        WHERE address = ?
                                        AND network_id = ?
                                        AND blk_from = ?
                                        AND blk_to = ?`)
	if err != nil {
		log.Info("some error", "error", err)
		return err
	}

	_, err = delete.Exec(account, chainID, (*bigint.SQLBigInt)(from), (*bigint.SQLBigInt)(to))
	return err
}

func deleteAllRanges(creator statementCreator, account common.Address) error {
	delete, err := creator.Prepare(`DELETE FROM blocks_ranges WHERE address = ?`)
	if err != nil {
		return err
	}

	_, err = delete.Exec(account)
	return err
}

func insertRange(chainID uint64, creator statementCreator, account common.Address, from *big.Int, to *big.Int) error {
	log.Info("insert blocks range", "account", account, "network", chainID, "from", from, "to", to)
	insert, err := creator.Prepare("INSERT INTO blocks_ranges (network_id, address, blk_from, blk_to) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}

	_, err = insert.Exec(chainID, account, (*bigint.SQLBigInt)(from), (*bigint.SQLBigInt)(to))
	return err
}
