package transfer

import (
	"context"
	"errors"
	"math/big"
	"time"

	"golang.org/x/exp/slices" // since 1.21, this is in the standard library

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/status-im/status-go/rpc/chain"
	w_common "github.com/status-im/status-go/services/wallet/common"
)

var (
	zero = big.NewInt(0)
	one  = big.NewInt(1)
	two  = big.NewInt(2)
)

// Partial transaction info obtained by ERC20Downloader.
// A PreloadedTransaction represents a Transaction which contains one
// ERC20/ERC721/ERC1155 transfer event.
// To be converted into one Transfer object post-indexing.
type PreloadedTransaction struct {
	Type    w_common.Type  `json:"type"`
	ID      common.Hash    `json:"-"`
	Address common.Address `json:"address"`
	// Log that was used to generate preloaded transaction.
	Log     *types.Log `json:"log"`
	TokenID *big.Int   `json:"tokenId"`
	Value   *big.Int   `json:"value"`
}

// Transfer stores information about transfer.
// A Transfer represents a plain ETH transfer or some token activity inside a Transaction
// Since ERC1155 transfers can contain multiple tokens, a single Transfer represents a single token transfer,
// that means ERC1155 batch transfers will be represented by multiple Transfer objects.
type Transfer struct {
	Type        w_common.Type      `json:"type"`
	ID          common.Hash        `json:"-"`
	Address     common.Address     `json:"address"`
	BlockNumber *big.Int           `json:"blockNumber"`
	BlockHash   common.Hash        `json:"blockhash"`
	Timestamp   uint64             `json:"timestamp"`
	Transaction *types.Transaction `json:"transaction"`
	Loaded      bool
	NetworkID   uint64
	// From is derived from tx signature in order to offload this computation from UI component.
	From    common.Address `json:"from"`
	Receipt *types.Receipt `json:"receipt"`
	// Log that was used to generate erc20 transfer. Nil for eth transfer.
	Log *types.Log `json:"log"`
	// TokenID is the id of the transferred token. Nil for eth transfer.
	TokenID *big.Int `json:"tokenId"`
	// TokenValue is the value of the token transfer. Nil for eth transfer.
	TokenValue  *big.Int `json:"tokenValue"`
	BaseGasFees string
	// Internal field that is used to track multi-transaction transfers.
	MultiTransactionID MultiTransactionIDType `json:"multi_transaction_id"`
}

// ETHDownloader downloads regular eth transfers and tokens transfers.
type ETHDownloader struct {
	chainClient chain.ClientInterface
	accounts    []common.Address
	signer      types.Signer
	db          *Database
}

var errLogsDownloaderStuck = errors.New("logs downloader stuck")

func (d *ETHDownloader) GetTransfersByNumber(ctx context.Context, number *big.Int) ([]Transfer, error) {
	blk, err := d.chainClient.BlockByNumber(ctx, number)
	if err != nil {
		return nil, err
	}
	rst, err := d.getTransfersInBlock(ctx, blk, d.accounts)
	if err != nil {
		return nil, err
	}
	return rst, err
}

// Only used by status-mobile
func getTransferByHash(ctx context.Context, client chain.ClientInterface, signer types.Signer, address common.Address, hash common.Hash) (*Transfer, error) {
	transaction, _, err := client.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	receipt, err := client.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, err
	}

	eventType, transactionLog := w_common.GetFirstEvent(receipt.Logs)
	transactionType := w_common.EventTypeToSubtransactionType(eventType)

	from, err := types.Sender(signer, transaction)

	if err != nil {
		return nil, err
	}

	baseGasFee, err := client.GetBaseFeeFromBlock(big.NewInt(int64(transactionLog.BlockNumber)))
	if err != nil {
		return nil, err
	}

	transfer := &Transfer{
		Type:        transactionType,
		ID:          hash,
		Address:     address,
		BlockNumber: receipt.BlockNumber,
		BlockHash:   receipt.BlockHash,
		Timestamp:   uint64(time.Now().Unix()),
		Transaction: transaction,
		From:        from,
		Receipt:     receipt,
		Log:         transactionLog,
		BaseGasFees: baseGasFee,
	}

	return transfer, nil
}

func (d *ETHDownloader) getTransfersInBlock(ctx context.Context, blk *types.Block, accounts []common.Address) ([]Transfer, error) {
	startTs := time.Now()

	rst := make([]Transfer, 0, len(blk.Transactions()))

	receiptsByAddressAndTxHash := make(map[common.Address]map[common.Hash]*types.Receipt)
	txsByAddressAndTxHash := make(map[common.Address]map[common.Hash]*types.Transaction)

	addReceiptToCache := func(address common.Address, txHash common.Hash, receipt *types.Receipt) {
		if receiptsByAddressAndTxHash[address] == nil {
			receiptsByAddressAndTxHash[address] = make(map[common.Hash]*types.Receipt)
		}
		receiptsByAddressAndTxHash[address][txHash] = receipt
	}

	addTxToCache := func(address common.Address, txHash common.Hash, tx *types.Transaction) {
		if txsByAddressAndTxHash[address] == nil {
			txsByAddressAndTxHash[address] = make(map[common.Hash]*types.Transaction)
		}
		txsByAddressAndTxHash[address][txHash] = tx
	}

	getReceiptFromCache := func(address common.Address, txHash common.Hash) *types.Receipt {
		if receiptsByAddressAndTxHash[address] == nil {
			return nil
		}
		return receiptsByAddressAndTxHash[address][txHash]
	}

	getTxFromCache := func(address common.Address, txHash common.Hash) *types.Transaction {
		if txsByAddressAndTxHash[address] == nil {
			return nil
		}
		return txsByAddressAndTxHash[address][txHash]
	}

	getReceipt := func(address common.Address, txHash common.Hash) (receipt *types.Receipt, err error) {
		receipt = getReceiptFromCache(address, txHash)
		if receipt == nil {
			receipt, err = d.fetchTransactionReceipt(ctx, txHash)
			if err != nil {
				return nil, err
			}
			addReceiptToCache(address, txHash, receipt)
		}
		return receipt, nil
	}

	getTx := func(address common.Address, txHash common.Hash) (tx *types.Transaction, err error) {
		tx = getTxFromCache(address, txHash)
		if tx == nil {
			tx, err = d.fetchTransaction(ctx, txHash)
			if err != nil {
				return nil, err
			}
			addTxToCache(address, txHash, tx)
		}
		return tx, nil
	}

	for _, address := range accounts {
		// During block discovery, we should have populated the DB with 1 item per transfer log containing
		// erc20/erc721/erc1155 transfers.
		// ID is a hash of the tx hash and the log index. log_index is unique per ERC20/721 tx, but not per ERC1155 tx.
		transactionsToLoad, err := d.db.GetTransactionsToLoad(d.chainClient.NetworkID(), address, blk.Number())
		if err != nil {
			return nil, err
		}

		areSubTxsCheckedForTxHash := make(map[common.Hash]bool)

		log.Debug("getTransfersInBlock", "block", blk.Number(), "transactionsToLoad", len(transactionsToLoad))

		for _, t := range transactionsToLoad {
			receipt, err := getReceipt(address, t.Log.TxHash)
			if err != nil {
				return nil, err
			}

			tx, err := getTx(address, t.Log.TxHash)
			if err != nil {
				return nil, err
			}

			subtransactions, err := d.subTransactionsFromPreloaded(t, tx, receipt, blk)
			if err != nil {
				log.Error("can't fetch subTxs for erc20/erc721/erc1155 transfer", "error", err)
				return nil, err
			}
			rst = append(rst, subtransactions...)
			areSubTxsCheckedForTxHash[t.Log.TxHash] = true
		}

		for _, tx := range blk.Transactions() {
			// Skip dummy blob transactions, as they are not supported by us
			if tx.Type() == types.BlobTxType {
				continue
			}
			if tx.ChainId().Cmp(big.NewInt(0)) != 0 && tx.ChainId().Cmp(d.chainClient.ToBigInt()) != 0 {
				log.Info("chain id mismatch", "tx hash", tx.Hash(), "tx chain id", tx.ChainId(), "expected chain id", d.chainClient.NetworkID())
				continue
			}
			from, err := types.Sender(d.signer, tx)

			if err != nil {
				if err == core.ErrTxTypeNotSupported {
					log.Error("Tx Type not supported", "tx chain id", tx.ChainId(), "type", tx.Type(), "error", err)
					continue
				}
				return nil, err
			}

			isPlainTransfer := from == address || (tx.To() != nil && *tx.To() == address)
			mustCheckSubTxs := false

			if !isPlainTransfer {
				// We might miss some subTransactions of interest for some transaction types. We need to check if we
				// find the address in the transaction data.
				switch tx.Type() {
				case types.DynamicFeeTxType, types.OptimismDepositTxType, types.ArbitrumDepositTxType, types.ArbitrumRetryTxType:
					mustCheckSubTxs = !areSubTxsCheckedForTxHash[tx.Hash()] && w_common.TxDataContainsAddress(tx.Type(), tx.Data(), address)
				}
			}

			if isPlainTransfer || mustCheckSubTxs {
				receipt, err := getReceipt(address, tx.Hash())
				if err != nil {
					return nil, err
				}

				// Since we've already got the receipt, check for subTxs of
				// interest in case we haven't already.
				if !areSubTxsCheckedForTxHash[tx.Hash()] {
					subtransactions, err := d.subTransactionsFromTransactionData(address, from, tx, receipt, blk)
					if err != nil {
						log.Error("can't fetch subTxs for eth transfer", "error", err)
						return nil, err
					}
					rst = append(rst, subtransactions...)
					areSubTxsCheckedForTxHash[tx.Hash()] = true
				}

				// If it's a plain ETH transfer, add it to the list
				if isPlainTransfer {
					rst = append(rst, Transfer{
						Type:               w_common.EthTransfer,
						NetworkID:          tx.ChainId().Uint64(),
						ID:                 tx.Hash(),
						Address:            address,
						BlockNumber:        blk.Number(),
						BlockHash:          receipt.BlockHash,
						Timestamp:          blk.Time(),
						Transaction:        tx,
						From:               from,
						Receipt:            receipt,
						Log:                nil,
						BaseGasFees:        blk.BaseFee().String(),
						MultiTransactionID: NoMultiTransactionID})
				}
			}
		}
	}
	log.Debug("getTransfersInBlock found", "block", blk.Number(), "len", len(rst), "time", time.Since(startTs))
	// TODO(dshulyak) test that balance difference was covered by transactions
	return rst, nil
}

// NewERC20TransfersDownloader returns new instance.
func NewERC20TransfersDownloader(client chain.ClientInterface, accounts []common.Address, signer types.Signer, incomingOnly bool) *ERC20TransfersDownloader {
	signature := w_common.GetEventSignatureHash(w_common.Erc20_721TransferEventSignature)

	return &ERC20TransfersDownloader{
		client:                 client,
		accounts:               accounts,
		signature:              signature,
		incomingOnly:           incomingOnly,
		signatureErc1155Single: w_common.GetEventSignatureHash(w_common.Erc1155TransferSingleEventSignature),
		signatureErc1155Batch:  w_common.GetEventSignatureHash(w_common.Erc1155TransferBatchEventSignature),
		signer:                 signer,
	}
}

// ERC20TransfersDownloader is a downloader for erc20 and erc721 tokens transfers.
// Since both transaction types share the same signature, both will be assigned
// type Erc20Transfer. Until the downloader gets refactored and a migration of the
// database gets implemented, differentiation between erc20 and erc721 will handled
// in the controller.
type ERC20TransfersDownloader struct {
	client       chain.ClientInterface
	accounts     []common.Address
	incomingOnly bool

	// hash of the Transfer event signature
	signature              common.Hash
	signatureErc1155Single common.Hash
	signatureErc1155Batch  common.Hash

	// signer is used to derive tx sender from tx signature
	signer types.Signer
}

func topicFromAddressSlice(addresses []common.Address) []common.Hash {
	rst := make([]common.Hash, len(addresses))
	for i, address := range addresses {
		rst[i] = common.BytesToHash(address.Bytes())
	}
	return rst
}

func (d *ERC20TransfersDownloader) inboundTopics(addresses []common.Address) [][]common.Hash {
	return [][]common.Hash{{d.signature}, {}, topicFromAddressSlice(addresses)}
}

func (d *ERC20TransfersDownloader) outboundTopics(addresses []common.Address) [][]common.Hash {
	return [][]common.Hash{{d.signature}, topicFromAddressSlice(addresses), {}}
}

func (d *ERC20TransfersDownloader) inboundERC20OutboundERC1155Topics(addresses []common.Address) [][]common.Hash {
	return [][]common.Hash{{d.signature, d.signatureErc1155Single, d.signatureErc1155Batch}, {}, topicFromAddressSlice(addresses)}
}

func (d *ERC20TransfersDownloader) inboundTopicsERC1155(addresses []common.Address) [][]common.Hash {
	return [][]common.Hash{{d.signatureErc1155Single, d.signatureErc1155Batch}, {}, {}, topicFromAddressSlice(addresses)}
}

func (d *ETHDownloader) fetchTransactionReceipt(parent context.Context, txHash common.Hash) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
	receipt, err := d.chainClient.TransactionReceipt(ctx, txHash)
	cancel()
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

func (d *ETHDownloader) fetchTransaction(parent context.Context, txHash common.Hash) (*types.Transaction, error) {
	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
	tx, _, err := d.chainClient.TransactionByHash(ctx, txHash) // TODO Save on requests by checking in the DB first
	cancel()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (d *ETHDownloader) subTransactionsFromPreloaded(preloadedTx *PreloadedTransaction, tx *types.Transaction, receipt *types.Receipt, blk *types.Block) ([]Transfer, error) {
	log.Debug("subTransactionsFromPreloaded start", "txHash", tx.Hash().Hex(), "address", preloadedTx.Address, "tokenID", preloadedTx.TokenID, "value", preloadedTx.Value)
	address := preloadedTx.Address
	txLog := preloadedTx.Log

	rst := make([]Transfer, 0, 1)

	from, err := types.Sender(d.signer, tx)
	if err != nil {
		if err == core.ErrTxTypeNotSupported {
			return nil, nil
		}
		return nil, err
	}

	eventType := w_common.GetEventType(preloadedTx.Log)
	// Only add ERC20/ERC721/ERC1155 transfers from/to the given account
	// from/to matching is already handled by getLogs filter
	switch eventType {
	case w_common.Erc20TransferEventType,
		w_common.Erc721TransferEventType,
		w_common.Erc1155TransferSingleEventType, w_common.Erc1155TransferBatchEventType:
		log.Debug("subTransactionsFromPreloaded transfer", "eventType", eventType, "logIdx", txLog.Index, "txHash", tx.Hash().Hex(), "address", address.Hex(), "tokenID", preloadedTx.TokenID, "value", preloadedTx.Value, "baseFee", blk.BaseFee().String())

		transfer := Transfer{
			Type:               w_common.EventTypeToSubtransactionType(eventType),
			ID:                 preloadedTx.ID,
			Address:            address,
			BlockNumber:        new(big.Int).SetUint64(txLog.BlockNumber),
			BlockHash:          txLog.BlockHash,
			Loaded:             true,
			NetworkID:          d.signer.ChainID().Uint64(),
			From:               from,
			Log:                txLog,
			TokenID:            preloadedTx.TokenID,
			TokenValue:         preloadedTx.Value,
			BaseGasFees:        blk.BaseFee().String(),
			Transaction:        tx,
			Receipt:            receipt,
			Timestamp:          blk.Time(),
			MultiTransactionID: NoMultiTransactionID,
		}

		rst = append(rst, transfer)
	}

	log.Debug("subTransactionsFromPreloaded end", "txHash", tx.Hash().Hex(), "address", address.Hex(), "tokenID", preloadedTx.TokenID, "value", preloadedTx.Value)
	return rst, nil
}

func (d *ETHDownloader) subTransactionsFromTransactionData(address, from common.Address, tx *types.Transaction, receipt *types.Receipt, blk *types.Block) ([]Transfer, error) {
	log.Debug("subTransactionsFromTransactionData start", "txHash", tx.Hash().Hex(), "address", address)

	rst := make([]Transfer, 0, 1)

	for _, txLog := range receipt.Logs {
		eventType := w_common.GetEventType(txLog)
		switch eventType {
		case w_common.UniswapV2SwapEventType, w_common.UniswapV3SwapEventType,
			w_common.HopBridgeTransferSentToL2EventType, w_common.HopBridgeTransferFromL1CompletedEventType,
			w_common.HopBridgeWithdrawalBondedEventType, w_common.HopBridgeTransferSentEventType:
			transfer := Transfer{
				Type:               w_common.EventTypeToSubtransactionType(eventType),
				ID:                 w_common.GetLogSubTxID(*txLog),
				Address:            address,
				BlockNumber:        new(big.Int).SetUint64(txLog.BlockNumber),
				BlockHash:          txLog.BlockHash,
				Loaded:             true,
				NetworkID:          d.signer.ChainID().Uint64(),
				From:               from,
				Log:                txLog,
				BaseGasFees:        blk.BaseFee().String(),
				Transaction:        tx,
				Receipt:            receipt,
				Timestamp:          blk.Time(),
				MultiTransactionID: NoMultiTransactionID,
			}

			rst = append(rst, transfer)
		}
	}

	log.Debug("subTransactionsFromTransactionData end", "txHash", tx.Hash().Hex(), "address", address.Hex())
	return rst, nil
}

func (d *ERC20TransfersDownloader) blocksFromLogs(parent context.Context, logs []types.Log) ([]*DBHeader, error) {
	concurrent := NewConcurrentDownloader(parent, NoThreadLimit)

	for i := range logs {
		l := logs[i]

		if l.Removed {
			continue
		}

		var address common.Address
		from, to, txIDs, tokenIDs, values, err := w_common.ParseTransferLog(l)
		if err != nil {
			log.Error("failed to parse transfer log", "log", l, "address", d.accounts, "error", err)
			continue
		}

		// Double check provider returned the correct log
		if slices.Contains(d.accounts, from) {
			address = from
		} else if slices.Contains(d.accounts, to) {
			address = to
		} else {
			log.Error("from/to address mismatch", "log", l, "addresses", d.accounts)
			continue
		}

		eventType := w_common.GetEventType(&l)
		logType := w_common.EventTypeToSubtransactionType(eventType)

		for i, txID := range txIDs {
			log.Debug("block from logs", "block", l.BlockNumber, "log", l, "logType", logType, "txID", txID)

			// For ERC20 there is no tokenID, so we use nil
			var tokenID *big.Int
			if len(tokenIDs) > i {
				tokenID = tokenIDs[i]
			}

			header := &DBHeader{
				Number:  big.NewInt(int64(l.BlockNumber)),
				Hash:    l.BlockHash,
				Address: address,
				PreloadedTransactions: []*PreloadedTransaction{{
					ID:      txID,
					Type:    logType,
					Log:     &l,
					TokenID: tokenID,
					Value:   values[i],
				}},
				Loaded: false,
			}

			concurrent.Add(func(ctx context.Context) error {
				concurrent.PushHeader(header)
				return nil
			})
		}
	}
	select {
	case <-concurrent.WaitAsync():
	case <-parent.Done():
		return nil, errLogsDownloaderStuck
	}
	return concurrent.GetHeaders(), concurrent.Error()
}

// GetHeadersInRange returns transfers between two blocks.
// time to get logs for 100000 blocks = 1.144686979s. with 249 events in the result set.
func (d *ERC20TransfersDownloader) GetHeadersInRange(parent context.Context, from, to *big.Int) ([]*DBHeader, error) {
	start := time.Now()
	log.Debug("get erc20 transfers in range start", "chainID", d.client.NetworkID(), "from", from, "to", to)
	headers := []*DBHeader{}
	ctx := context.Background()
	var err error
	outbound := []types.Log{}
	var inboundOrMixed []types.Log // inbound ERC20 or outbound ERC1155 share the same signature for our purposes
	if !d.incomingOnly {
		outbound, err = d.client.FilterLogs(ctx, ethereum.FilterQuery{
			FromBlock: from,
			ToBlock:   to,
			Topics:    d.outboundTopics(d.accounts),
		})
		if err != nil {
			return nil, err
		}
		inboundOrMixed, err = d.client.FilterLogs(ctx, ethereum.FilterQuery{
			FromBlock: from,
			ToBlock:   to,
			Topics:    d.inboundERC20OutboundERC1155Topics(d.accounts),
		})
		if err != nil {
			return nil, err
		}
	} else {
		inboundOrMixed, err = d.client.FilterLogs(ctx, ethereum.FilterQuery{
			FromBlock: from,
			ToBlock:   to,
			Topics:    d.inboundTopics(d.accounts),
		})
		if err != nil {
			return nil, err
		}
	}

	inbound1155, err := d.client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: from,
		ToBlock:   to,
		Topics:    d.inboundTopicsERC1155(d.accounts),
	})
	if err != nil {
		return nil, err
	}

	logs := concatLogs(outbound, inboundOrMixed, inbound1155)

	if len(logs) == 0 {
		log.Debug("no logs found for account")
		return nil, nil
	}

	rst, err := d.blocksFromLogs(parent, logs)
	if err != nil {
		return nil, err
	}
	if len(rst) == 0 {
		log.Warn("no headers found in logs for account", "chainID", d.client.NetworkID(), "addresses", d.accounts, "from", from, "to", to)
	} else {
		headers = append(headers, rst...)
		log.Debug("found erc20 transfers for account", "chainID", d.client.NetworkID(), "addresses", d.accounts,
			"from", from, "to", to, "headers", len(headers))
	}

	log.Debug("get erc20 transfers in range end", "chainID", d.client.NetworkID(),
		"from", from, "to", to, "headers", len(headers), "took", time.Since(start))
	return headers, nil
}

func concatLogs(slices ...[]types.Log) []types.Log {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}
	tmp := make([]types.Log, totalLen)
	var i int
	for _, s := range slices {
		i += copy(tmp[i:], s)
	}

	return tmp
}
