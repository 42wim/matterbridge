package transfer

import (
	"database/sql"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/services/wallet/bigint"
)

type BlockRangeDAOer interface {
	getBlockRange(chainID uint64, address common.Address) (blockRange *ethTokensBlockRanges, err error)
	upsertRange(chainID uint64, account common.Address, newBlockRange *ethTokensBlockRanges) (err error)
	updateTokenRange(chainID uint64, account common.Address, newBlockRange *BlockRange) (err error)
	upsertEthRange(chainID uint64, account common.Address, newBlockRange *BlockRange) (err error)
}

type BlockRangeSequentialDAO struct {
	db *sql.DB
}

type BlockRange struct {
	Start      *big.Int // Block of first transfer
	FirstKnown *big.Int // Oldest scanned block
	LastKnown  *big.Int // Last scanned block
}

func NewBlockRange() *BlockRange {
	return &BlockRange{Start: &big.Int{}, FirstKnown: &big.Int{}, LastKnown: &big.Int{}}
}

type ethTokensBlockRanges struct {
	eth              *BlockRange
	tokens           *BlockRange
	balanceCheckHash string
}

func newEthTokensBlockRanges() *ethTokensBlockRanges {
	return &ethTokensBlockRanges{eth: NewBlockRange(), tokens: NewBlockRange()}
}

func (b *BlockRangeSequentialDAO) getBlockRange(chainID uint64, address common.Address) (blockRange *ethTokensBlockRanges, err error) {
	query := `SELECT blk_start, blk_first, blk_last, token_blk_start, token_blk_first, token_blk_last, balance_check_hash FROM blocks_ranges_sequential
	WHERE address = ?
	AND network_id = ?`

	rows, err := b.db.Query(query, address, chainID)
	if err != nil {
		return
	}
	defer rows.Close()

	blockRange = &ethTokensBlockRanges{}
	if rows.Next() {
		blockRange = newEthTokensBlockRanges()
		err = rows.Scan((*bigint.SQLBigInt)(blockRange.eth.Start),
			(*bigint.SQLBigInt)(blockRange.eth.FirstKnown),
			(*bigint.SQLBigInt)(blockRange.eth.LastKnown),
			(*bigint.SQLBigInt)(blockRange.tokens.Start),
			(*bigint.SQLBigInt)(blockRange.tokens.FirstKnown),
			(*bigint.SQLBigInt)(blockRange.tokens.LastKnown),
			&blockRange.balanceCheckHash,
		)
		if err != nil {
			return nil, err
		}

		return blockRange, nil
	}

	return blockRange, nil
}

func (b *BlockRangeSequentialDAO) deleteRange(account common.Address) error {
	log.Debug("delete blocks range", "account", account)
	delete, err := b.db.Prepare(`DELETE FROM blocks_ranges_sequential WHERE address = ?`)
	if err != nil {
		log.Error("Failed to prepare deletion of sequential block range", "error", err)
		return err
	}

	_, err = delete.Exec(account)
	return err
}

func (b *BlockRangeSequentialDAO) upsertRange(chainID uint64, account common.Address, newBlockRange *ethTokensBlockRanges) (err error) {
	ethTokensBlockRange, err := b.getBlockRange(chainID, account)
	if err != nil {
		return err
	}

	ethBlockRange := prepareUpdatedBlockRange(ethTokensBlockRange.eth, newBlockRange.eth)
	tokensBlockRange := prepareUpdatedBlockRange(ethTokensBlockRange.tokens, newBlockRange.tokens)

	log.Debug("update eth and tokens blocks range", "account", account, "chainID", chainID,
		"eth.start", ethBlockRange.Start, "eth.first", ethBlockRange.FirstKnown, "eth.last", ethBlockRange.LastKnown,
		"tokens.start", tokensBlockRange.Start, "tokens.first", ethBlockRange.FirstKnown, "eth.last", ethBlockRange.LastKnown, "hash", newBlockRange.balanceCheckHash)

	upsert, err := b.db.Prepare(`REPLACE INTO blocks_ranges_sequential
					(network_id, address, blk_start, blk_first, blk_last, token_blk_start, token_blk_first, token_blk_last, balance_check_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = upsert.Exec(chainID, account, (*bigint.SQLBigInt)(ethBlockRange.Start), (*bigint.SQLBigInt)(ethBlockRange.FirstKnown), (*bigint.SQLBigInt)(ethBlockRange.LastKnown),
		(*bigint.SQLBigInt)(tokensBlockRange.Start), (*bigint.SQLBigInt)(tokensBlockRange.FirstKnown), (*bigint.SQLBigInt)(tokensBlockRange.LastKnown), newBlockRange.balanceCheckHash)

	return err
}

func (b *BlockRangeSequentialDAO) upsertEthRange(chainID uint64, account common.Address,
	newBlockRange *BlockRange) (err error) {

	ethTokensBlockRange, err := b.getBlockRange(chainID, account)
	if err != nil {
		return err
	}

	blockRange := prepareUpdatedBlockRange(ethTokensBlockRange.eth, newBlockRange)

	log.Debug("update eth blocks range", "account", account, "chainID", chainID,
		"start", blockRange.Start, "first", blockRange.FirstKnown, "last", blockRange.LastKnown, "old hash", ethTokensBlockRange.balanceCheckHash)

	upsert, err := b.db.Prepare(`REPLACE INTO blocks_ranges_sequential
					(network_id, address, blk_start, blk_first, blk_last, token_blk_start, token_blk_first, token_blk_last, balance_check_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	if ethTokensBlockRange.tokens == nil {
		ethTokensBlockRange.tokens = NewBlockRange()
	}

	_, err = upsert.Exec(chainID, account, (*bigint.SQLBigInt)(blockRange.Start), (*bigint.SQLBigInt)(blockRange.FirstKnown), (*bigint.SQLBigInt)(blockRange.LastKnown),
		(*bigint.SQLBigInt)(ethTokensBlockRange.tokens.Start), (*bigint.SQLBigInt)(ethTokensBlockRange.tokens.FirstKnown), (*bigint.SQLBigInt)(ethTokensBlockRange.tokens.LastKnown), ethTokensBlockRange.balanceCheckHash)

	return err
}

func (b *BlockRangeSequentialDAO) updateTokenRange(chainID uint64, account common.Address,
	newBlockRange *BlockRange) (err error) {

	ethTokensBlockRange, err := b.getBlockRange(chainID, account)
	if err != nil {
		return err
	}

	blockRange := prepareUpdatedBlockRange(ethTokensBlockRange.tokens, newBlockRange)

	log.Debug("update tokens blocks range", "account", account, "chainID", chainID,
		"start", blockRange.Start, "first", blockRange.FirstKnown, "last", blockRange.LastKnown, "old hash", ethTokensBlockRange.balanceCheckHash)

	update, err := b.db.Prepare(`UPDATE blocks_ranges_sequential SET token_blk_start = ?, token_blk_first = ?, token_blk_last = ? WHERE network_id = ? AND address = ?`)
	if err != nil {
		return err
	}

	_, err = update.Exec((*bigint.SQLBigInt)(blockRange.Start), (*bigint.SQLBigInt)(blockRange.FirstKnown),
		(*bigint.SQLBigInt)(blockRange.LastKnown), chainID, account)

	return err
}

func prepareUpdatedBlockRange(blockRange, newBlockRange *BlockRange) *BlockRange {
	// Update existing range
	if blockRange != nil {
		if newBlockRange != nil {
			// Ovewrite start block if there was not any or if new one is older, because it can be precised only
			// to a greater value, because no history can be before some block that is considered
			// as a start of history, but due to concurrent block range checks, a newer greater block
			// can be found that matches criteria of a start block (nonce is zero, balances are equal)
			if newBlockRange.Start != nil && (blockRange.Start == nil || blockRange.Start.Cmp(newBlockRange.Start) < 0) {
				blockRange.Start = newBlockRange.Start
			}

			// Overwrite first known block if there was not any or if new one is older
			if (blockRange.FirstKnown == nil && newBlockRange.FirstKnown != nil) ||
				(blockRange.FirstKnown != nil && newBlockRange.FirstKnown != nil && blockRange.FirstKnown.Cmp(newBlockRange.FirstKnown) > 0) {
				blockRange.FirstKnown = newBlockRange.FirstKnown
			}

			// Overwrite last known block if there was not any or if new one is newer
			if (blockRange.LastKnown == nil && newBlockRange.LastKnown != nil) ||
				(blockRange.LastKnown != nil && newBlockRange.LastKnown != nil && blockRange.LastKnown.Cmp(newBlockRange.LastKnown) < 0) {
				blockRange.LastKnown = newBlockRange.LastKnown
			}
		}
	} else {
		blockRange = newBlockRange
	}

	return blockRange
}
