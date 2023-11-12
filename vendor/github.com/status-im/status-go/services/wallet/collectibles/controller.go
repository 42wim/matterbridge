package collectibles

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/rpc/network"
	"github.com/status-im/status-go/services/accounts/accountsevent"
	"github.com/status-im/status-go/services/accounts/settingsevent"
	"github.com/status-im/status-go/services/wallet/async"
	walletCommon "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/transfer"
	"github.com/status-im/status-go/services/wallet/walletevent"
)

const (
	activityRefetchMarginSeconds = 30 * 60 // Trigger a fetch if activity is detected this many seconds before the last fetch
)

type commandPerChainID = map[walletCommon.ChainID]*periodicRefreshOwnedCollectiblesCommand
type commandPerAddressAndChainID = map[common.Address]commandPerChainID

type timerPerChainID = map[walletCommon.ChainID]*time.Timer
type timerPerAddressAndChainID = map[common.Address]timerPerChainID

type Controller struct {
	manager      *Manager
	ownershipDB  *OwnershipDB
	walletFeed   *event.Feed
	accountsDB   *accounts.Database
	accountsFeed *event.Feed
	settingsFeed *event.Feed

	networkManager *network.Manager
	cancelFn       context.CancelFunc

	commands            commandPerAddressAndChainID
	timers              timerPerAddressAndChainID
	group               *async.Group
	accountsWatcher     *accountsevent.Watcher
	walletEventsWatcher *walletevent.Watcher
	settingsWatcher     *settingsevent.Watcher

	ownedCollectiblesChangeCb OwnedCollectiblesChangeCb
	collectiblesTransferCb    TransferCb

	commandsLock sync.RWMutex
}

func NewController(
	db *sql.DB,
	walletFeed *event.Feed,
	accountsDB *accounts.Database,
	accountsFeed *event.Feed,
	settingsFeed *event.Feed,
	networkManager *network.Manager,
	manager *Manager) *Controller {
	return &Controller{
		manager:        manager,
		ownershipDB:    NewOwnershipDB(db),
		walletFeed:     walletFeed,
		accountsDB:     accountsDB,
		accountsFeed:   accountsFeed,
		settingsFeed:   settingsFeed,
		networkManager: networkManager,
		commands:       make(commandPerAddressAndChainID),
		timers:         make(timerPerAddressAndChainID),
	}
}

func (c *Controller) SetOwnedCollectiblesChangeCb(cb OwnedCollectiblesChangeCb) {
	c.ownedCollectiblesChangeCb = cb
}

func (c *Controller) SetCollectiblesTransferCb(cb TransferCb) {
	c.collectiblesTransferCb = cb
}

func (c *Controller) Start() {
	// Setup periodical collectibles refresh
	_ = c.startPeriodicalOwnershipFetch()

	// Setup collectibles fetch when a new account gets added
	c.startAccountsWatcher()

	// Setup collectibles fetch when relevant activity is detected
	c.startWalletEventsWatcher()

	// Setup collectibles fetch when chain-related settings change
	c.startSettingsWatcher()
}

func (c *Controller) Stop() {
	c.stopSettingsWatcher()

	c.stopWalletEventsWatcher()

	c.stopAccountsWatcher()

	c.stopPeriodicalOwnershipFetch()
}

func (c *Controller) RefetchOwnedCollectibles() {
	c.stopPeriodicalOwnershipFetch()
	c.manager.ResetConnectionStatus()
	_ = c.startPeriodicalOwnershipFetch()
}

func (c *Controller) GetCommandState(chainID walletCommon.ChainID, address common.Address) OwnershipState {
	c.commandsLock.RLock()
	defer c.commandsLock.RUnlock()

	state := OwnershipStateIdle
	if c.commands[address] != nil && c.commands[address][chainID] != nil {
		state = c.commands[address][chainID].GetState()
	}

	return state
}

func (c *Controller) isPeriodicalOwnershipFetchRunning() bool {
	return c.group != nil
}

// Starts periodical fetching for the all wallet addresses and all chains
func (c *Controller) startPeriodicalOwnershipFetch() error {
	c.commandsLock.Lock()
	defer c.commandsLock.Unlock()

	if c.isPeriodicalOwnershipFetchRunning() {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFn = cancel

	c.group = async.NewGroup(ctx)

	addresses, err := c.accountsDB.GetWalletAddresses()
	if err != nil {
		return err
	}

	for _, addr := range addresses {
		err := c.startPeriodicalOwnershipFetchForAccount(common.Address(addr))
		if err != nil {
			log.Error("Error starting periodical collectibles fetch for accpunt", "address", addr, "error", err)
			return err
		}
	}

	return nil
}

func (c *Controller) stopPeriodicalOwnershipFetch() {
	c.commandsLock.Lock()
	defer c.commandsLock.Unlock()

	if !c.isPeriodicalOwnershipFetchRunning() {
		return
	}

	if c.cancelFn != nil {
		c.cancelFn()
		c.cancelFn = nil
	}
	if c.group != nil {
		c.group.Stop()
		c.group.Wait()
		c.group = nil
		c.commands = make(commandPerAddressAndChainID)
	}
}

// Starts (or restarts) periodical fetching for the given account address for all chains
func (c *Controller) startPeriodicalOwnershipFetchForAccount(address common.Address) error {
	log.Debug("wallet.api.collectibles.Controller Start periodical fetching", "address", address)

	networks, err := c.networkManager.Get(false)
	if err != nil {
		return err
	}

	areTestNetworksEnabled, err := c.accountsDB.GetTestNetworksEnabled()
	if err != nil {
		return err
	}

	for _, network := range networks {
		if network.IsTest != areTestNetworksEnabled {
			continue
		}
		chainID := walletCommon.ChainID(network.ChainID)

		err := c.startPeriodicalOwnershipFetchForAccountAndChainID(address, chainID, false)
		if err != nil {
			return err
		}
	}

	return nil
}

// Starts (or restarts) periodical fetching for the given account address for all chains
func (c *Controller) startPeriodicalOwnershipFetchForAccountAndChainID(address common.Address, chainID walletCommon.ChainID, delayed bool) error {
	log.Debug("wallet.api.collectibles.Controller Start periodical fetching", "address", address, "chainID", chainID, "delayed", delayed)

	if !c.isPeriodicalOwnershipFetchRunning() {
		return errors.New("periodical fetch not initialized")
	}

	err := c.stopPeriodicalOwnershipFetchForAccountAndChainID(address, chainID)
	if err != nil {
		return err
	}

	if _, ok := c.commands[address]; !ok {
		c.commands[address] = make(commandPerChainID)
	}

	command := newPeriodicRefreshOwnedCollectiblesCommand(
		c.manager,
		c.ownershipDB,
		c.walletFeed,
		chainID,
		address,
		c.ownedCollectiblesChangeCb,
	)

	c.commands[address][chainID] = command
	if delayed {
		c.group.Add(command.DelayedCommand())
	} else {
		c.group.Add(command.Command())
	}

	return nil
}

// Stop periodical fetching for the given account address for all chains
func (c *Controller) stopPeriodicalOwnershipFetchForAccount(address common.Address) error {
	log.Debug("wallet.api.collectibles.Controller Stop periodical fetching", "address", address)

	if !c.isPeriodicalOwnershipFetchRunning() {
		return errors.New("periodical fetch not initialized")
	}

	if _, ok := c.commands[address]; ok {
		for chainID := range c.commands[address] {
			err := c.stopPeriodicalOwnershipFetchForAccountAndChainID(address, chainID)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (c *Controller) stopPeriodicalOwnershipFetchForAccountAndChainID(address common.Address, chainID walletCommon.ChainID) error {
	log.Debug("wallet.api.collectibles.Controller Stop periodical fetching", "address", address, "chainID", chainID)

	if !c.isPeriodicalOwnershipFetchRunning() {
		return errors.New("periodical fetch not initialized")
	}

	if _, ok := c.commands[address]; ok {
		if _, ok := c.commands[address][chainID]; ok {
			c.commands[address][chainID].Stop()
			delete(c.commands[address], chainID)
		}
		// If it was the last chain, delete the address as well
		if len(c.commands[address]) == 0 {
			delete(c.commands, address)
		}
	}

	return nil
}

func (c *Controller) startAccountsWatcher() {
	if c.accountsWatcher != nil {
		return
	}

	accountChangeCb := func(changedAddresses []common.Address, eventType accountsevent.EventType, currentAddresses []common.Address) {
		c.commandsLock.Lock()
		defer c.commandsLock.Unlock()
		// Whenever an account gets added, start fetching
		if eventType == accountsevent.EventTypeAdded {
			for _, address := range changedAddresses {
				err := c.startPeriodicalOwnershipFetchForAccount(address)
				if err != nil {
					log.Error("Error starting periodical collectibles fetch", "address", address, "error", err)
				}
			}
		} else if eventType == accountsevent.EventTypeRemoved {
			for _, address := range changedAddresses {
				err := c.stopPeriodicalOwnershipFetchForAccount(address)
				if err != nil {
					log.Error("Error starting periodical collectibles fetch", "address", address, "error", err)
				}
			}
		}
	}

	c.accountsWatcher = accountsevent.NewWatcher(c.accountsDB, c.accountsFeed, accountChangeCb)

	c.accountsWatcher.Start()
}

func (c *Controller) stopAccountsWatcher() {
	if c.accountsWatcher != nil {
		c.accountsWatcher.Stop()
		c.accountsWatcher = nil
	}
}

func (c *Controller) startWalletEventsWatcher() {
	if c.walletEventsWatcher != nil {
		return
	}

	walletEventCb := func(event walletevent.Event) {
		// EventRecentHistoryReady ?
		if event.Type != transfer.EventInternalERC721TransferDetected &&
			event.Type != transfer.EventInternalERC1155TransferDetected {
			return
		}

		chainID := walletCommon.ChainID(event.ChainID)
		for _, account := range event.Accounts {
			// Call external callback
			if c.collectiblesTransferCb != nil {
				c.collectiblesTransferCb(account, chainID, event.EventParams.([]transfer.Transfer))
			}

			c.refetchOwnershipIfRecentTransfer(account, chainID, event.At)
		}
	}

	c.walletEventsWatcher = walletevent.NewWatcher(c.walletFeed, walletEventCb)

	c.walletEventsWatcher.Start()
}

func (c *Controller) stopWalletEventsWatcher() {
	if c.walletEventsWatcher != nil {
		c.walletEventsWatcher.Stop()
		c.walletEventsWatcher = nil
	}
}

func (c *Controller) startSettingsWatcher() {
	if c.settingsWatcher != nil {
		return
	}

	settingChangeCb := func(setting settings.SettingField, value interface{}) {
		if setting.Equals(settings.TestNetworksEnabled) || setting.Equals(settings.IsSepoliaEnabled) {
			c.stopPeriodicalOwnershipFetch()
			err := c.startPeriodicalOwnershipFetch()
			if err != nil {
				log.Error("Error starting periodical collectibles fetch", "error", err)
			}
		}
	}

	c.settingsWatcher = settingsevent.NewWatcher(c.settingsFeed, settingChangeCb)

	c.settingsWatcher.Start()
}

func (c *Controller) stopSettingsWatcher() {
	if c.settingsWatcher != nil {
		c.settingsWatcher.Stop()
		c.settingsWatcher = nil
	}
}

func (c *Controller) refetchOwnershipIfRecentTransfer(account common.Address, chainID walletCommon.ChainID, latestTxTimestamp int64) {

	// Check last ownership update timestamp
	timestamp, err := c.ownershipDB.GetOwnershipUpdateTimestamp(account, chainID)

	if err != nil {
		log.Error("Error getting ownership update timestamp", "error", err)
		return
	}
	if timestamp == InvalidTimestamp {
		// Ownership was never fetched for this account
		return
	}

	timeCheck := timestamp - activityRefetchMarginSeconds
	if timeCheck < 0 {
		timeCheck = 0
	}

	if latestTxTimestamp > timeCheck {
		// Restart fetching for account + chainID
		c.commandsLock.Lock()
		err := c.startPeriodicalOwnershipFetchForAccountAndChainID(account, chainID, true)
		c.commandsLock.Unlock()
		if err != nil {
			log.Error("Error starting periodical collectibles fetch", "address", account, "error", err)
		}
	}
}
