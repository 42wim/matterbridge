package ens

import (
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"github.com/wealdtech/go-ens/v3"
	"github.com/wealdtech/go-multicodec"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/contracts"
	"github.com/status-im/status-go/contracts/registrar"
	"github.com/status-im/status-go/contracts/resolver"
	"github.com/status-im/status-go/contracts/snt"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/services/utils"
	wcommon "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/transactions"
)

const StatusDomain = "stateofus.eth"

func NewAPI(rpcClient *rpc.Client, accountsManager *account.GethManager, pendingTracker *transactions.PendingTxTracker, config *params.NodeConfig, appDb *sql.DB, timeSource func() time.Time, syncUserDetailFunc *syncUsernameDetail) *API {
	return &API{
		contractMaker: &contracts.ContractMaker{
			RPCClient: rpcClient,
		},
		accountsManager: accountsManager,
		pendingTracker:  pendingTracker,
		config:          config,
		addrPerChain:    make(map[uint64]common.Address),
		db:              NewEnsDatabase(appDb),

		quit:               make(chan struct{}),
		timeSource:         timeSource,
		syncUserDetailFunc: syncUserDetailFunc,
	}
}

type URI struct {
	Scheme string
	Host   string
	Path   string
}

// use this to avoid using messenger directly to avoid circular dependency (protocol->ens->protocol)
type syncUsernameDetail func(context.Context, *UsernameDetail) error

type API struct {
	contractMaker   *contracts.ContractMaker
	accountsManager *account.GethManager
	pendingTracker  *transactions.PendingTxTracker
	config          *params.NodeConfig

	addrPerChain      map[uint64]common.Address
	addrPerChainMutex sync.Mutex

	quitOnce sync.Once
	quit     chan struct{}

	db                 *Database
	syncUserDetailFunc *syncUsernameDetail

	timeSource func() time.Time
}

func (api *API) Stop() {
	api.quitOnce.Do(func() {
		close(api.quit)
	})
}

func (api *API) unixTime() uint64 {
	return uint64(api.timeSource().Unix())
}

func (api *API) GetEnsUsernames(ctx context.Context) ([]*UsernameDetail, error) {
	removed := false
	return api.db.GetEnsUsernames(&removed)
}

func (api *API) Add(ctx context.Context, chainID uint64, username string) error {
	ud := &UsernameDetail{Username: username, ChainID: chainID, Clock: api.unixTime()}
	err := api.db.AddEnsUsername(ud)
	if err != nil {
		return err
	}
	return (*api.syncUserDetailFunc)(ctx, ud)
}

func (api *API) Remove(ctx context.Context, chainID uint64, username string) error {
	ud := &UsernameDetail{Username: username, ChainID: chainID, Clock: api.unixTime()}
	affected, err := api.db.RemoveEnsUsername(ud)
	if err != nil {
		return err
	}
	if affected {
		return (*api.syncUserDetailFunc)(ctx, ud)
	}
	return nil
}

func (api *API) GetRegistrarAddress(ctx context.Context, chainID uint64) (common.Address, error) {
	return api.usernameRegistrarAddr(ctx, chainID)
}

func (api *API) Resolver(ctx context.Context, chainID uint64, username string) (*common.Address, error) {
	err := validateENSUsername(username)
	if err != nil {
		return nil, err
	}

	registry, err := api.contractMaker.NewRegistry(chainID)
	if err != nil {
		return nil, err
	}

	callOpts := &bind.CallOpts{Context: ctx, Pending: false}
	resolver, err := registry.Resolver(callOpts, nameHash(username))
	if err != nil {
		return nil, err
	}

	return &resolver, nil
}

func (api *API) GetName(ctx context.Context, chainID uint64, address common.Address) (string, error) {
	backend, err := api.contractMaker.RPCClient.EthClient(chainID)
	if err != nil {
		return "", err
	}
	return ens.ReverseResolve(backend, address)
}

func (api *API) OwnerOf(ctx context.Context, chainID uint64, username string) (*common.Address, error) {
	err := validateENSUsername(username)
	if err != nil {
		return nil, err
	}

	registry, err := api.contractMaker.NewRegistry(chainID)
	if err != nil {
		return nil, err
	}

	callOpts := &bind.CallOpts{Context: ctx, Pending: false}
	owner, err := registry.Owner(callOpts, nameHash(username))
	if err != nil {
		return nil, err
	}

	return &owner, nil
}

func (api *API) ContentHash(ctx context.Context, chainID uint64, username string) ([]byte, error) {
	err := validateENSUsername(username)
	if err != nil {
		return nil, err
	}

	resolverAddress, err := api.Resolver(ctx, chainID, username)
	if err != nil {
		return nil, err
	}

	resolver, err := api.contractMaker.NewPublicResolver(chainID, resolverAddress)
	if err != nil {
		return nil, err
	}

	callOpts := &bind.CallOpts{Context: ctx, Pending: false}
	contentHash, err := resolver.Contenthash(callOpts, nameHash(username))
	if err != nil {
		return nil, nil
	}

	return contentHash, nil
}

func (api *API) PublicKeyOf(ctx context.Context, chainID uint64, username string) (string, error) {
	err := validateENSUsername(username)
	if err != nil {
		return "", err
	}

	resolverAddress, err := api.Resolver(ctx, chainID, username)
	if err != nil {
		return "", err
	}

	resolver, err := api.contractMaker.NewPublicResolver(chainID, resolverAddress)
	if err != nil {
		return "", err
	}

	callOpts := &bind.CallOpts{Context: ctx, Pending: false}
	pubKey, err := resolver.Pubkey(callOpts, nameHash(username))
	if err != nil {
		return "", err
	}
	return "0x04" + hex.EncodeToString(pubKey.X[:]) + hex.EncodeToString(pubKey.Y[:]), nil
}

func (api *API) AddressOf(ctx context.Context, chainID uint64, username string) (*common.Address, error) {
	err := validateENSUsername(username)
	if err != nil {
		return nil, err
	}

	resolverAddress, err := api.Resolver(ctx, chainID, username)
	if err != nil {
		return nil, err
	}

	resolver, err := api.contractMaker.NewPublicResolver(chainID, resolverAddress)
	if err != nil {
		return nil, err
	}

	callOpts := &bind.CallOpts{Context: ctx, Pending: false}
	addr, err := resolver.Addr(callOpts, nameHash(username))
	if err != nil {
		return nil, err
	}

	return &addr, nil
}

func (api *API) usernameRegistrarAddr(ctx context.Context, chainID uint64) (common.Address, error) {
	log.Info("obtaining username registrar address")
	api.addrPerChainMutex.Lock()
	defer api.addrPerChainMutex.Unlock()
	addr, ok := api.addrPerChain[chainID]
	if ok {
		return addr, nil
	}

	registryAddr, err := api.OwnerOf(ctx, chainID, StatusDomain)
	if err != nil {
		return common.Address{}, err
	}

	api.addrPerChain[chainID] = *registryAddr

	go func() {
		registry, err := api.contractMaker.NewRegistry(chainID)
		if err != nil {
			return
		}

		logs := make(chan *resolver.ENSRegistryWithFallbackNewOwner)

		sub, err := registry.WatchNewOwner(&bind.WatchOpts{}, logs, nil, nil)
		if err != nil {
			return
		}

		for {
			select {
			case <-api.quit:
				log.Info("quitting ens contract subscription")
				sub.Unsubscribe()
				return
			case err := <-sub.Err():
				if err != nil {
					log.Error("ens contract subscription error: " + err.Error())
				}
				return
			case vLog := <-logs:
				api.addrPerChainMutex.Lock()
				api.addrPerChain[chainID] = vLog.Owner
				api.addrPerChainMutex.Unlock()
			}
		}
	}()

	return *registryAddr, nil
}

func (api *API) ExpireAt(ctx context.Context, chainID uint64, username string) (string, error) {
	registryAddr, err := api.usernameRegistrarAddr(ctx, chainID)
	if err != nil {
		return "", err
	}

	registrar, err := api.contractMaker.NewUsernameRegistrar(chainID, registryAddr)
	if err != nil {
		return "", err
	}

	callOpts := &bind.CallOpts{Context: ctx, Pending: false}
	expTime, err := registrar.GetExpirationTime(callOpts, usernameToLabel(username))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", expTime), nil
}

func (api *API) Price(ctx context.Context, chainID uint64) (string, error) {
	registryAddr, err := api.usernameRegistrarAddr(ctx, chainID)
	if err != nil {
		return "", err
	}

	registrar, err := api.contractMaker.NewUsernameRegistrar(chainID, registryAddr)
	if err != nil {
		return "", err
	}

	callOpts := &bind.CallOpts{Context: ctx, Pending: false}
	price, err := registrar.GetPrice(callOpts)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", price), nil
}

func (api *API) Release(ctx context.Context, chainID uint64, txArgs transactions.SendTxArgs, password string, username string) (string, error) {
	registryAddr, err := api.usernameRegistrarAddr(ctx, chainID)
	if err != nil {
		return "", err
	}

	registrar, err := api.contractMaker.NewUsernameRegistrar(chainID, registryAddr)
	if err != nil {
		return "", err
	}

	txOpts := txArgs.ToTransactOpts(utils.GetSigner(chainID, api.accountsManager, api.config.KeyStoreDir, txArgs.From, password))
	tx, err := registrar.Release(txOpts, usernameToLabel(username))
	if err != nil {
		return "", err
	}

	err = api.pendingTracker.TrackPendingTransaction(
		wcommon.ChainID(chainID),
		tx.Hash(),
		common.Address(txArgs.From),
		transactions.ReleaseENS,
		transactions.AutoDelete,
	)
	if err != nil {
		log.Error("TrackPendingTransaction error", "error", err)
		return "", err
	}

	err = api.Remove(ctx, chainID, fullDomainName(username))

	if err != nil {
		log.Warn("Releasing ENS username: transaction successful, but removing failed")
	}

	return tx.Hash().String(), nil
}

func (api *API) ReleasePrepareTxCallMsg(ctx context.Context, chainID uint64, txArgs transactions.SendTxArgs, username string) (ethereum.CallMsg, error) {
	registrarABI, err := abi.JSON(strings.NewReader(registrar.UsernameRegistrarABI))
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	data, err := registrarABI.Pack("release", usernameToLabel(username))
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	sntAddress, err := snt.ContractAddress(chainID)
	if err != nil {
		return ethereum.CallMsg{}, err
	}
	return ethereum.CallMsg{
		From:  common.Address(txArgs.From),
		To:    &sntAddress,
		Value: big.NewInt(0),
		Data:  data,
	}, nil
}

func (api *API) ReleasePrepareTx(ctx context.Context, chainID uint64, txArgs transactions.SendTxArgs, username string) (interface{}, error) {
	callMsg, err := api.ReleasePrepareTxCallMsg(ctx, chainID, txArgs, username)
	if err != nil {
		return nil, err
	}

	return toCallArg(callMsg), nil
}

func (api *API) ReleaseEstimate(ctx context.Context, chainID uint64, txArgs transactions.SendTxArgs, username string) (uint64, error) {
	registrarABI, err := abi.JSON(strings.NewReader(registrar.UsernameRegistrarABI))
	if err != nil {
		return 0, err
	}

	data, err := registrarABI.Pack("release", usernameToLabel(username))
	if err != nil {
		return 0, err
	}

	ethClient, err := api.contractMaker.RPCClient.EthClient(chainID)
	if err != nil {
		return 0, err
	}

	registryAddr, err := api.usernameRegistrarAddr(ctx, chainID)
	if err != nil {
		return 0, err
	}

	estimate, err := ethClient.EstimateGas(ctx, ethereum.CallMsg{
		From:  common.Address(txArgs.From),
		To:    &registryAddr,
		Value: big.NewInt(0),
		Data:  data,
	})
	if err != nil {
		return 0, err
	}
	return estimate + 1000, nil
}

func (api *API) Register(ctx context.Context, chainID uint64, txArgs transactions.SendTxArgs, password string, username string, pubkey string) (string, error) {
	snt, err := api.contractMaker.NewSNT(chainID)
	if err != nil {
		return "", err
	}

	priceHex, err := api.Price(ctx, chainID)
	if err != nil {
		return "", err
	}
	price := new(big.Int)
	price.SetString(priceHex, 16)

	registrarABI, err := abi.JSON(strings.NewReader(registrar.UsernameRegistrarABI))
	if err != nil {
		return "", err
	}

	x, y := extractCoordinates(pubkey)
	extraData, err := registrarABI.Pack("register", usernameToLabel(username), common.Address(txArgs.From), x, y)
	if err != nil {
		return "", err
	}

	registryAddr, err := api.usernameRegistrarAddr(ctx, chainID)
	if err != nil {
		return "", err
	}

	txOpts := txArgs.ToTransactOpts(utils.GetSigner(chainID, api.accountsManager, api.config.KeyStoreDir, txArgs.From, password))
	tx, err := snt.ApproveAndCall(
		txOpts,
		registryAddr,
		price,
		extraData,
	)

	if err != nil {
		return "", err
	}

	err = api.pendingTracker.TrackPendingTransaction(
		wcommon.ChainID(chainID),
		tx.Hash(),
		common.Address(txArgs.From),
		transactions.RegisterENS,
		transactions.AutoDelete,
	)
	if err != nil {
		log.Error("TrackPendingTransaction error", "error", err)
		return "", err
	}

	err = api.Add(ctx, chainID, fullDomainName(username))
	if err != nil {
		log.Warn("Registering ENS username: transaction successful, but adding failed")
	}

	return tx.Hash().String(), nil
}

func (api *API) RegisterPrepareTxCallMsg(ctx context.Context, chainID uint64, txArgs transactions.SendTxArgs, username string, pubkey string) (ethereum.CallMsg, error) {
	priceHex, err := api.Price(ctx, chainID)
	if err != nil {
		return ethereum.CallMsg{}, err
	}
	price := new(big.Int)
	price.SetString(priceHex, 16)

	registrarABI, err := abi.JSON(strings.NewReader(registrar.UsernameRegistrarABI))
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	x, y := extractCoordinates(pubkey)
	extraData, err := registrarABI.Pack("register", usernameToLabel(username), common.Address(txArgs.From), x, y)
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	sntABI, err := abi.JSON(strings.NewReader(snt.SNTABI))
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	registryAddr, err := api.usernameRegistrarAddr(ctx, chainID)
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	data, err := sntABI.Pack("approveAndCall", registryAddr, price, extraData)
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	sntAddress, err := snt.ContractAddress(chainID)
	if err != nil {
		return ethereum.CallMsg{}, err
	}
	return ethereum.CallMsg{
		From:  common.Address(txArgs.From),
		To:    &sntAddress,
		Value: big.NewInt(0),
		Data:  data,
	}, nil
}

func (api *API) RegisterPrepareTx(ctx context.Context, chainID uint64, txArgs transactions.SendTxArgs, username string, pubkey string) (interface{}, error) {
	callMsg, err := api.RegisterPrepareTxCallMsg(ctx, chainID, txArgs, username, pubkey)
	if err != nil {
		return nil, err
	}

	return toCallArg(callMsg), nil
}

func (api *API) RegisterEstimate(ctx context.Context, chainID uint64, txArgs transactions.SendTxArgs, username string, pubkey string) (uint64, error) {
	ethClient, err := api.contractMaker.RPCClient.EthClient(chainID)
	if err != nil {
		return 0, err
	}

	callMsg, err := api.RegisterPrepareTxCallMsg(ctx, chainID, txArgs, username, pubkey)
	if err != nil {
		return 0, err
	}

	estimate, err := ethClient.EstimateGas(ctx, callMsg)
	if err != nil {
		return 0, err
	}
	return estimate + 1000, nil
}

func (api *API) SetPubKey(ctx context.Context, chainID uint64, txArgs transactions.SendTxArgs, password string, username string, pubkey string) (string, error) {
	err := validateENSUsername(username)
	if err != nil {
		return "", err
	}

	resolverAddress, err := api.Resolver(ctx, chainID, username)
	if err != nil {
		return "", err
	}

	resolver, err := api.contractMaker.NewPublicResolver(chainID, resolverAddress)
	if err != nil {
		return "", err
	}

	x, y := extractCoordinates(pubkey)
	txOpts := txArgs.ToTransactOpts(utils.GetSigner(chainID, api.accountsManager, api.config.KeyStoreDir, txArgs.From, password))
	tx, err := resolver.SetPubkey(txOpts, nameHash(username), x, y)
	if err != nil {
		return "", err
	}

	err = api.pendingTracker.TrackPendingTransaction(
		wcommon.ChainID(chainID),
		tx.Hash(),
		common.Address(txArgs.From),
		transactions.SetPubKey,
		transactions.AutoDelete,
	)
	if err != nil {
		log.Error("TrackPendingTransaction error", "error", err)
		return "", err
	}

	err = api.Add(ctx, chainID, fullDomainName(username))

	if err != nil {
		log.Warn("Registering ENS username: transaction successful, but adding failed")
	}

	return tx.Hash().String(), nil
}

func (api *API) SetPubKeyPrepareTxCallMsg(ctx context.Context, chainID uint64, txArgs transactions.SendTxArgs, username string, pubkey string) (ethereum.CallMsg, error) {
	err := validateENSUsername(username)
	if err != nil {
		return ethereum.CallMsg{}, err
	}
	x, y := extractCoordinates(pubkey)

	resolverABI, err := abi.JSON(strings.NewReader(resolver.PublicResolverABI))
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	data, err := resolverABI.Pack("setPubkey", nameHash(username), x, y)
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	resolverAddress, err := api.Resolver(ctx, chainID, username)
	if err != nil {
		return ethereum.CallMsg{}, err
	}

	return ethereum.CallMsg{
		From:  common.Address(txArgs.From),
		To:    resolverAddress,
		Value: big.NewInt(0),
		Data:  data,
	}, nil
}

func (api *API) SetPubKeyPrepareTx(ctx context.Context, chainID uint64, txArgs transactions.SendTxArgs, username string, pubkey string) (interface{}, error) {
	callMsg, err := api.SetPubKeyPrepareTxCallMsg(ctx, chainID, txArgs, username, pubkey)
	if err != nil {
		return nil, err
	}

	return toCallArg(callMsg), nil
}

func (api *API) SetPubKeyEstimate(ctx context.Context, chainID uint64, txArgs transactions.SendTxArgs, username string, pubkey string) (uint64, error) {
	ethClient, err := api.contractMaker.RPCClient.EthClient(chainID)
	if err != nil {
		return 0, err
	}

	callMsg, err := api.SetPubKeyPrepareTxCallMsg(ctx, chainID, txArgs, username, pubkey)
	if err != nil {
		return 0, err
	}

	estimate, err := ethClient.EstimateGas(ctx, callMsg)
	if err != nil {
		return 0, err
	}
	return estimate + 1000, nil
}

func (api *API) ResourceURL(ctx context.Context, chainID uint64, username string) (*URI, error) {
	scheme := "https"
	contentHash, err := api.ContentHash(ctx, chainID, username)
	if err != nil {
		return nil, err
	}

	if len(contentHash) == 0 {
		return &URI{}, nil
	}

	data, codec, err := multicodec.RemoveCodec(contentHash)
	if err != nil {
		return nil, err
	}
	codecName, err := multicodec.Name(codec)
	if err != nil {
		return nil, err
	}

	switch codecName {
	case "ipfs-ns":
		thisCID, err := cid.Parse(data)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse CID")
		}
		str, err := thisCID.StringOfBase(multibase.Base32)
		if err != nil {
			return nil, errors.Wrap(err, "failed to obtain base36 representation")
		}

		parsedURL, _ := url.Parse(params.IpfsGatewayURL)
		// Remove scheme from the url
		host := parsedURL.Hostname() + parsedURL.Path + str
		return &URI{scheme, host, ""}, nil
	case "ipns-ns":
		id, offset := binary.Uvarint(data)
		if id == 0 {
			return nil, fmt.Errorf("unknown CID")
		}

		data, _, err := multicodec.RemoveCodec(data[offset:])
		if err != nil {
			return nil, err
		}
		decodedMHash, err := multihash.Decode(data)
		if err != nil {
			return nil, err
		}

		return &URI{scheme, string(decodedMHash.Digest), ""}, nil
	case "swarm-ns":
		id, offset := binary.Uvarint(data)
		if id == 0 {
			return nil, fmt.Errorf("unknown CID")
		}
		data, _, err := multicodec.RemoveCodec(data[offset:])
		if err != nil {
			return nil, err
		}
		decodedMHash, err := multihash.Decode(data)
		if err != nil {
			return nil, err
		}
		path := "/bzz:/" + hex.EncodeToString(decodedMHash.Digest) + "/"
		return &URI{scheme, "swarm-gateways.net", path}, nil
	default:
		return nil, fmt.Errorf("unknown codec name %s", codecName)
	}
}

func toCallArg(msg ethereum.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	return arg
}

func fullDomainName(username string) string {
	return username + "." + StatusDomain
}
