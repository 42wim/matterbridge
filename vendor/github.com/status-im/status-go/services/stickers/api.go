package stickers

import (
	"context"
	"math/big"
	"time"

	"github.com/zenthangplus/goccm"
	"olympos.io/encoding/edn"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/contracts"
	"github.com/status-im/status-go/contracts/stickers"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/ipfs"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/server"
	"github.com/status-im/status-go/services/wallet/bigint"
	"github.com/status-im/status-go/transactions"
)

const maxConcurrentRequests = 3
const requestTimeout = time.Duration(5) * time.Second

// ConnectionType constants
type stickerStatus int

const (
	statusAvailable stickerStatus = iota
	statusInstalled
	statusPending
	statusPurchased
)

type API struct {
	contractMaker   *contracts.ContractMaker
	accountsManager *account.GethManager
	accountsDB      *accounts.Database
	pendingTracker  *transactions.PendingTxTracker

	keyStoreDir string
	downloader  *ipfs.Downloader
	httpServer  *server.MediaServer

	ctx context.Context
}

type Sticker struct {
	PackID *bigint.BigInt `json:"packID,omitempty"`
	URL    string         `json:"url,omitempty"`
	Hash   string         `json:"hash,omitempty"`
}

type StickerPack struct {
	ID        *bigint.BigInt `json:"id"`
	Name      string         `json:"name"`
	Author    string         `json:"author"`
	Owner     common.Address `json:"owner,omitempty"`
	Price     *bigint.BigInt `json:"price"`
	Preview   string         `json:"preview"`
	Thumbnail string         `json:"thumbnail"`
	Stickers  []Sticker      `json:"stickers"`

	Status stickerStatus `json:"status"`
}

type StickerPackCollection map[uint]StickerPack

type ednSticker struct {
	Hash string
}

type ednStickerPack struct {
	Name      string
	Author    string
	Thumbnail string
	Preview   string
	Stickers  []ednSticker
}
type ednStickerPackInfo struct {
	Meta ednStickerPack
}

func NewAPI(ctx context.Context, acc *accounts.Database, rpcClient *rpc.Client, accountsManager *account.GethManager, pendingTracker *transactions.PendingTxTracker, keyStoreDir string, downloader *ipfs.Downloader, httpServer *server.MediaServer) *API {
	result := &API{
		contractMaker: &contracts.ContractMaker{
			RPCClient: rpcClient,
		},
		accountsManager: accountsManager,
		accountsDB:      acc,
		pendingTracker:  pendingTracker,
		keyStoreDir:     keyStoreDir,
		downloader:      downloader,
		ctx:             ctx,
		httpServer:      httpServer,
	}

	return result
}

func (api *API) Market(chainID uint64) ([]StickerPack, error) {
	// TODO: eventually this should be changed to include pagination
	accs, err := api.accountsDB.GetActiveAccounts()
	if err != nil {
		return nil, err
	}

	allStickerPacks, err := api.getContractPacks(chainID)
	if err != nil {
		return nil, err
	}

	purchasedPacks := make(map[uint]struct{})

	purchasedPackChan := make(chan *big.Int)
	errChan := make(chan error)
	doneChan := make(chan struct{}, 1)
	go api.getAccountsPurchasedPack(chainID, accs, purchasedPackChan, errChan, doneChan)

	for {
		select {
		case err := <-errChan:
			if err != nil {
				return nil, err
			}
		case packID := <-purchasedPackChan:
			if packID != nil {
				purchasedPacks[uint(packID.Uint64())] = struct{}{}
			}

		case <-doneChan:
			var result []StickerPack
			for _, pack := range allStickerPacks {
				packID := uint(pack.ID.Uint64())
				_, isPurchased := purchasedPacks[packID]
				if isPurchased {
					pack.Status = statusPurchased
				} else {
					pack.Status = statusAvailable
				}
				result = append(result, pack)
			}

			return result, nil
		}
	}
}

func (api *API) execTokenPackID(chainID uint64, tokenIDs []*big.Int, resultChan chan<- *big.Int, errChan chan<- error, doneChan chan<- struct{}) {
	defer close(doneChan)
	defer close(errChan)
	defer close(resultChan)

	stickerPack, err := api.contractMaker.NewStickerPack(chainID)
	if err != nil {
		errChan <- err
		return
	}

	if len(tokenIDs) == 0 {
		return
	}

	callOpts := &bind.CallOpts{Context: api.ctx, Pending: false}

	c := goccm.New(maxConcurrentRequests)
	for _, tokenID := range tokenIDs {
		c.Wait()
		go func(tokenID *big.Int) {
			defer c.Done()
			packID, err := stickerPack.TokenPackId(callOpts, tokenID)
			if err != nil {
				errChan <- err
				return
			}
			resultChan <- packID
		}(tokenID)
	}
	c.WaitAllDone()
}

func (api *API) getTokenPackIDs(chainID uint64, tokenIDs []*big.Int) ([]*big.Int, error) {
	tokenPackIDChan := make(chan *big.Int)
	errChan := make(chan error)
	doneChan := make(chan struct{}, 1)

	go api.execTokenPackID(chainID, tokenIDs, tokenPackIDChan, errChan, doneChan)

	var tokenPackIDs []*big.Int
	for {
		select {
		case <-doneChan:
			return tokenPackIDs, nil
		case err := <-errChan:
			if err != nil {
				return nil, err
			}
		case t := <-tokenPackIDChan:
			if t != nil {
				tokenPackIDs = append(tokenPackIDs, t)
			}
		}
	}
}

func (api *API) getPurchasedPackIDs(chainID uint64, account types.Address) ([]*big.Int, error) {
	// TODO: this should be replaced in the future by something like TheGraph to reduce the number of requests to infura

	stickerPack, err := api.contractMaker.NewStickerPack(chainID)
	if err != nil {
		return nil, err
	}

	callOpts := &bind.CallOpts{Context: api.ctx, Pending: false}

	balance, err := stickerPack.BalanceOf(callOpts, common.Address(account))
	if err != nil {
		return nil, err
	}

	tokenIDs, err := api.getTokenOwnerOfIndex(chainID, account, balance)
	if err != nil {
		return nil, err
	}

	return api.getTokenPackIDs(chainID, tokenIDs)
}

func (api *API) fetchStickerPacks(chainID uint64, resultChan chan<- *StickerPack, errChan chan<- error, doneChan chan<- struct{}) {
	defer close(doneChan)
	defer close(errChan)
	defer close(resultChan)

	installedPacks, err := api.Installed()
	if err != nil {
		errChan <- err
		return
	}

	pendingPacks, err := api.pendingStickerPacks()
	if err != nil {
		errChan <- err
		return
	}

	stickerType, err := api.contractMaker.NewStickerType(chainID)
	if err != nil {
		errChan <- err
		return
	}

	callOpts := &bind.CallOpts{Context: api.ctx, Pending: false}

	numPacks, err := stickerType.PackCount(callOpts)
	if err != nil {
		errChan <- err
		return
	}

	if numPacks.Uint64() == 0 {
		return
	}

	c := goccm.New(maxConcurrentRequests)
	for i := uint64(0); i < numPacks.Uint64(); i++ {
		c.Wait()
		go func(i uint64) {
			defer c.Done()

			packID := new(big.Int).SetUint64(i)

			_, exists := installedPacks[uint(i)]
			if exists {
				return // We already have the sticker pack data, no need to query it
			}

			_, exists = pendingPacks[uint(i)]
			if exists {
				return // We already have the sticker pack data, no need to query it
			}

			stickerPack, err := api.fetchPackData(stickerType, packID, true)
			if err != nil {
				log.Warn("Could not retrieve stickerpack data", "packID", packID, "error", err)
				errChan <- err
				return
			}

			resultChan <- stickerPack
		}(i)
	}

	c.WaitAllDone()
}

func (api *API) fetchPackData(stickerType *stickers.StickerType, packID *big.Int, translateHashes bool) (*StickerPack, error) {

	timeoutContext, timeoutCancel := context.WithTimeout(api.ctx, requestTimeout)
	defer timeoutCancel()

	callOpts := &bind.CallOpts{Context: timeoutContext, Pending: false}
	packData, err := stickerType.GetPackData(callOpts, packID)
	if err != nil {
		return nil, err
	}

	stickerPack := &StickerPack{
		ID:    &bigint.BigInt{Int: packID},
		Owner: packData.Owner,
		Price: &bigint.BigInt{Int: packData.Price},
	}

	err = api.downloadPackData(stickerPack, packData.Contenthash, translateHashes)
	if err != nil {
		return nil, err
	}

	return stickerPack, nil
}

func (api *API) downloadPackData(stickerPack *StickerPack, contentHash []byte, translateHashes bool) error {
	fileContent, err := api.downloader.Get(hexutil.Encode(contentHash)[2:], true)
	if err != nil {
		return err
	}
	return api.populateStickerPackAttributes(stickerPack, fileContent, translateHashes)
}

func (api *API) hashToURL(hash string) string {
	return api.httpServer.MakeStickerURL(hash)
}

func (api *API) populateStickerPackAttributes(stickerPack *StickerPack, ednSource []byte, translateHashes bool) error {
	var stickerpackIPFSInfo ednStickerPackInfo
	err := edn.Unmarshal(ednSource, &stickerpackIPFSInfo)
	if err != nil {
		return err
	}

	stickerPack.Author = stickerpackIPFSInfo.Meta.Author
	stickerPack.Name = stickerpackIPFSInfo.Meta.Name

	if translateHashes {
		stickerPack.Preview = api.hashToURL(stickerpackIPFSInfo.Meta.Preview)
		stickerPack.Thumbnail = api.hashToURL(stickerpackIPFSInfo.Meta.Thumbnail)
	} else {
		stickerPack.Preview = stickerpackIPFSInfo.Meta.Preview
		stickerPack.Thumbnail = stickerpackIPFSInfo.Meta.Thumbnail
	}

	for _, s := range stickerpackIPFSInfo.Meta.Stickers {
		url := ""
		if translateHashes {
			url = api.hashToURL(s.Hash)
		}

		stickerPack.Stickers = append(stickerPack.Stickers, Sticker{
			PackID: stickerPack.ID,
			URL:    url,
			Hash:   s.Hash,
		})
	}

	return nil
}

func (api *API) getContractPacks(chainID uint64) ([]StickerPack, error) {
	stickerPackChan := make(chan *StickerPack)
	errChan := make(chan error)
	doneChan := make(chan struct{}, 1)

	go api.fetchStickerPacks(chainID, stickerPackChan, errChan, doneChan)

	var packs []StickerPack

	for {
		select {
		case <-doneChan:
			return packs, nil
		case err := <-errChan:
			if err != nil {
				return nil, err
			}
		case pack := <-stickerPackChan:
			if pack != nil {
				packs = append(packs, *pack)
			}
		}
	}
}

func (api *API) getAccountsPurchasedPack(chainID uint64, accs []*accounts.Account, resultChan chan<- *big.Int, errChan chan<- error, doneChan chan<- struct{}) {
	defer close(doneChan)
	defer close(errChan)
	defer close(resultChan)

	if len(accs) == 0 {
		return
	}

	c := goccm.New(maxConcurrentRequests)
	for _, account := range accs {
		c.Wait()
		go func(acc *accounts.Account) {
			defer c.Done()
			packs, err := api.getPurchasedPackIDs(chainID, acc.Address)
			if err != nil {
				errChan <- err
				return
			}

			for _, p := range packs {
				resultChan <- p
			}
		}(account)
	}
	c.WaitAllDone()
}

func (api *API) execTokenOwnerOfIndex(chainID uint64, account types.Address, balance *big.Int, resultChan chan<- *big.Int, errChan chan<- error, doneChan chan<- struct{}) {
	defer close(doneChan)
	defer close(errChan)
	defer close(resultChan)

	stickerPack, err := api.contractMaker.NewStickerPack(chainID)
	if err != nil {
		errChan <- err
		return
	}

	if balance.Int64() == 0 {
		return
	}

	callOpts := &bind.CallOpts{Context: api.ctx, Pending: false}

	c := goccm.New(maxConcurrentRequests)
	for i := uint64(0); i < balance.Uint64(); i++ {
		c.Wait()
		go func(i uint64) {
			defer c.Done()
			tokenID, err := stickerPack.TokenOfOwnerByIndex(callOpts, common.Address(account), new(big.Int).SetUint64(i))
			if err != nil {
				errChan <- err
				return
			}

			resultChan <- tokenID
		}(i)
	}
	c.WaitAllDone()
}

func (api *API) getTokenOwnerOfIndex(chainID uint64, account types.Address, balance *big.Int) ([]*big.Int, error) {
	tokenIDChan := make(chan *big.Int)
	errChan := make(chan error)
	doneChan := make(chan struct{}, 1)

	go api.execTokenOwnerOfIndex(chainID, account, balance, tokenIDChan, errChan, doneChan)

	var tokenIDs []*big.Int
	for {
		select {
		case <-doneChan:
			return tokenIDs, nil
		case err := <-errChan:
			if err != nil {
				return nil, err
			}
		case tokenID := <-tokenIDChan:
			if tokenID != nil {
				tokenIDs = append(tokenIDs, tokenID)
			}
		}
	}
}
