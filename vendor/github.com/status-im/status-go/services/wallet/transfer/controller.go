package transfer

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"

	"golang.org/x/exp/slices" // since 1.21, this is in the standard library

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	statusaccounts "github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/services/accounts/accountsevent"
	"github.com/status-im/status-go/services/wallet/balance"
	"github.com/status-im/status-go/services/wallet/blockchainstate"
	"github.com/status-im/status-go/services/wallet/token"
	"github.com/status-im/status-go/transactions"
)

type Controller struct {
	db                 *Database
	accountsDB         *statusaccounts.Database
	rpcClient          *rpc.Client
	blockDAO           *BlockDAO
	blockRangesSeqDAO  *BlockRangeSequentialDAO
	reactor            *Reactor
	accountFeed        *event.Feed
	TransferFeed       *event.Feed
	accWatcher         *accountsevent.Watcher
	transactionManager *TransactionManager
	pendingTxManager   *transactions.PendingTxTracker
	tokenManager       *token.Manager
	balanceCacher      balance.Cacher
	blockChainState    *blockchainstate.BlockChainState
}

func NewTransferController(db *sql.DB, accountsDB *statusaccounts.Database, rpcClient *rpc.Client, accountFeed *event.Feed, transferFeed *event.Feed,
	transactionManager *TransactionManager, pendingTxManager *transactions.PendingTxTracker, tokenManager *token.Manager,
	balanceCacher balance.Cacher, blockChainState *blockchainstate.BlockChainState) *Controller {

	blockDAO := &BlockDAO{db}
	return &Controller{
		db:                 NewDB(db),
		accountsDB:         accountsDB,
		blockDAO:           blockDAO,
		blockRangesSeqDAO:  &BlockRangeSequentialDAO{db},
		rpcClient:          rpcClient,
		accountFeed:        accountFeed,
		TransferFeed:       transferFeed,
		transactionManager: transactionManager,
		pendingTxManager:   pendingTxManager,
		tokenManager:       tokenManager,
		balanceCacher:      balanceCacher,
		blockChainState:    blockChainState,
	}
}

func (c *Controller) Start() {
	go func() { _ = c.cleanupAccountsLeftovers() }()
}

func (c *Controller) Stop() {
	if c.reactor != nil {
		c.reactor.stop()
	}

	if c.accWatcher != nil {
		c.accWatcher.Stop()
		c.accWatcher = nil
	}
}

func sameChains(chainIDs1 []uint64, chainIDs2 []uint64) bool {
	if len(chainIDs1) != len(chainIDs2) {
		return false
	}

	for _, chainID := range chainIDs1 {
		if !slices.Contains(chainIDs2, chainID) {
			return false
		}
	}

	return true
}

func (c *Controller) CheckRecentHistory(chainIDs []uint64, accounts []common.Address) error {
	if len(accounts) == 0 {
		return nil
	}

	if len(chainIDs) == 0 {
		return nil
	}

	err := c.blockDAO.mergeBlocksRanges(chainIDs, accounts)
	if err != nil {
		return err
	}

	chainClients, err := c.rpcClient.EthClients(chainIDs)
	if err != nil {
		return err
	}

	if c.reactor != nil {
		if !sameChains(chainIDs, c.reactor.chainIDs) {
			err := c.reactor.restart(chainClients, accounts)
			if err != nil {
				return err
			}
		}

		return nil
	}

	multiaccSettings, err := c.accountsDB.GetSettings()
	if err != nil {
		return err
	}

	omitHistory := multiaccSettings.OmitTransfersHistoryScan
	if omitHistory {
		err := c.accountsDB.SaveSettingField(settings.OmitTransfersHistoryScan, false)
		if err != nil {
			return err
		}
	}

	c.reactor = NewReactor(c.db, c.blockDAO, c.blockRangesSeqDAO, c.accountsDB, c.TransferFeed, c.transactionManager,
		c.pendingTxManager, c.tokenManager, c.balanceCacher, omitHistory, c.blockChainState)

	err = c.reactor.start(chainClients, accounts)
	if err != nil {
		return err
	}

	c.startAccountWatcher(chainIDs)

	return nil
}

func (c *Controller) startAccountWatcher(chainIDs []uint64) {
	if c.accWatcher == nil {
		c.accWatcher = accountsevent.NewWatcher(c.accountsDB, c.accountFeed, func(changedAddresses []common.Address, eventType accountsevent.EventType, currentAddresses []common.Address) {
			c.onAccountsChanged(changedAddresses, eventType, currentAddresses, chainIDs)
		})
	}
	c.accWatcher.Start()
}

func (c *Controller) onAccountsChanged(changedAddresses []common.Address, eventType accountsevent.EventType, currentAddresses []common.Address, chainIDs []uint64) {
	if eventType == accountsevent.EventTypeRemoved {
		for _, address := range changedAddresses {
			c.cleanUpRemovedAccount(address)
		}
	}

	if c.reactor == nil {
		log.Warn("reactor is not initialized")
		return
	}

	if eventType == accountsevent.EventTypeAdded || eventType == accountsevent.EventTypeRemoved {
		log.Debug("list of accounts was changed from a previous version. reactor will be restarted", "new", currentAddresses)

		chainClients, err := c.rpcClient.EthClients(chainIDs)
		if err != nil {
			return
		}

		err = c.reactor.restart(chainClients, currentAddresses)
		if err != nil {
			log.Error("failed to restart reactor with new accounts", "error", err)
		}
	}
}

// Only used by status-mobile
func (c *Controller) LoadTransferByHash(ctx context.Context, rpcClient *rpc.Client, address common.Address, hash common.Hash) error {
	chainClient, err := rpcClient.EthClient(rpcClient.UpstreamChainID)
	if err != nil {
		return err
	}

	signer := types.LatestSignerForChainID(chainClient.ToBigInt())

	transfer, err := getTransferByHash(ctx, chainClient, signer, address, hash)
	if err != nil {
		return err
	}

	transfers := []Transfer{*transfer}

	err = c.db.InsertBlock(rpcClient.UpstreamChainID, address, transfer.BlockNumber, transfer.BlockHash)
	if err != nil {
		return err
	}

	tx, err := c.db.client.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	blocks := []*big.Int{transfer.BlockNumber}
	err = saveTransfersMarkBlocksLoaded(tx, rpcClient.UpstreamChainID, address, transfers, blocks)
	if err != nil {
		rollErr := tx.Rollback()
		if rollErr != nil {
			return fmt.Errorf("failed to rollback transaction due to error: %v", err)
		}
		return err
	}

	return nil
}

func (c *Controller) GetTransfersByAddress(ctx context.Context, chainID uint64, address common.Address, toBlock *big.Int,
	limit int64, fetchMore bool) ([]View, error) {

	rst, err := c.reactor.getTransfersByAddress(ctx, chainID, address, toBlock, limit)
	if err != nil {
		log.Error("[WalletAPI:: GetTransfersByAddress] can't fetch transfers", "err", err)
		return nil, err
	}

	return castToTransferViews(rst), nil
}

func (c *Controller) GetTransfersForIdentities(ctx context.Context, identities []TransactionIdentity) ([]View, error) {
	rst, err := c.db.GetTransfersForIdentities(ctx, identities)
	if err != nil {
		log.Error("[transfer.Controller.GetTransfersForIdentities] DB err", err)
		return nil, err
	}

	return castToTransferViews(rst), nil
}

func (c *Controller) GetCachedBalances(ctx context.Context, chainID uint64, addresses []common.Address) ([]BlockView, error) {
	result, error := c.blockDAO.getLastKnownBlocks(chainID, addresses)
	if error != nil {
		return nil, error
	}

	return blocksToViews(result), nil
}

func (c *Controller) cleanUpRemovedAccount(address common.Address) {
	// Transfers will be deleted by foreign key constraint by cascade
	err := deleteBlocks(c.db.client, address)
	if err != nil {
		log.Error("Failed to delete blocks", "error", err)
	}
	err = deleteAllRanges(c.db.client, address)
	if err != nil {
		log.Error("Failed to delete old blocks ranges", "error", err)
	}

	err = c.blockRangesSeqDAO.deleteRange(address)
	if err != nil {
		log.Error("Failed to delete blocks ranges sequential", "error", err)
	}
}

func (c *Controller) cleanupAccountsLeftovers() error {
	// We clean up accounts that were deleted and soft removed
	accounts, err := c.accountsDB.GetWalletAddresses()
	if err != nil {
		log.Error("Failed to get accounts", "error", err)
		return err
	}

	existingAddresses := make([]common.Address, len(accounts))
	for i, account := range accounts {
		existingAddresses[i] = (common.Address)(account)
	}

	addressesInWalletDB, err := getAddresses(c.db.client)
	if err != nil {
		log.Error("Failed to get addresses from wallet db", "error", err)
		return err
	}

	missing := findMissingItems(addressesInWalletDB, existingAddresses)
	for _, address := range missing {
		c.cleanUpRemovedAccount(address)
	}

	return nil
}

// find items from one slice that are not in another
func findMissingItems(slice1 []common.Address, slice2 []common.Address) []common.Address {
	var missing []common.Address
	for _, item := range slice1 {
		if !slices.Contains(slice2, item) {
			missing = append(missing, item)
		}
	}
	return missing
}
