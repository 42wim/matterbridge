package rpc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	gethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/rpc/chain"
	"github.com/status-im/status-go/rpc/network"
	"github.com/status-im/status-go/services/rpcstats"
	"github.com/status-im/status-go/services/wallet/common"
)

const (
	// DefaultCallTimeout is a default timeout for an RPC call
	DefaultCallTimeout = time.Minute
)

// List of RPC client errors.
var (
	ErrMethodNotFound = fmt.Errorf("The method does not exist/is not available")
)

// Handler defines handler for RPC methods.
type Handler func(context.Context, uint64, ...interface{}) (interface{}, error)

type ClientInterface interface {
	AbstractEthClient(chainID common.ChainID) (chain.BatchCallClient, error)
}

// Client represents RPC client with custom routing
// scheme. It automatically decides where RPC call
// goes - Upstream or Local node.
type Client struct {
	sync.RWMutex

	upstreamEnabled bool
	upstreamURL     string
	UpstreamChainID uint64

	local              *gethrpc.Client
	upstream           chain.ClientInterface
	rpcClientsMutex    sync.RWMutex
	rpcClients         map[uint64]chain.ClientInterface
	rpcLimiterMutex    sync.RWMutex
	limiterPerProvider map[string]*chain.RPCLimiter

	router         *router
	NetworkManager *network.Manager

	handlersMx sync.RWMutex       // mx guards handlers
	handlers   map[string]Handler // locally registered handlers
	log        log.Logger

	walletNotifier func(chainID uint64, message string)
}

// Is initialized in a build-tag-dependent module
var verifProxyInitFn func(c *Client)

// NewClient initializes Client and tries to connect to both,
// upstream and local node.
//
// Client is safe for concurrent use and will automatically
// reconnect to the server if connection is lost.
func NewClient(client *gethrpc.Client, upstreamChainID uint64, upstream params.UpstreamRPCConfig, networks []params.Network, db *sql.DB) (*Client, error) {
	var err error

	log := log.New("package", "status-go/rpc.Client")
	networkManager := network.NewManager(db)
	err = networkManager.Init(networks)
	if err != nil {
		log.Error("Network manager failed to initialize", "error", err)
	}

	c := Client{
		local:              client,
		NetworkManager:     networkManager,
		handlers:           make(map[string]Handler),
		rpcClients:         make(map[uint64]chain.ClientInterface),
		limiterPerProvider: make(map[string]*chain.RPCLimiter),
		log:                log,
	}

	if upstream.Enabled {
		c.UpstreamChainID = upstreamChainID
		c.upstreamEnabled = upstream.Enabled
		c.upstreamURL = upstream.URL
		upstreamClient, err := gethrpc.Dial(c.upstreamURL)
		if err != nil {
			return nil, fmt.Errorf("dial upstream server: %s", err)
		}
		limiter, err := c.getRPCLimiter(c.upstreamURL)
		if err != nil {
			return nil, fmt.Errorf("get RPC limiter: %s", err)
		}
		c.upstream = chain.NewSimpleClient(limiter, upstreamClient, upstreamChainID)
	}

	c.router = newRouter(c.upstreamEnabled)

	if verifProxyInitFn != nil {
		verifProxyInitFn(&c)
	}

	return &c, nil
}

func (c *Client) SetWalletNotifier(notifier func(chainID uint64, message string)) {
	c.walletNotifier = notifier
}

func extractLastParamFromURL(inputURL string) (string, error) {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", err
	}

	pathSegments := strings.Split(parsedURL.Path, "/")
	lastSegment := pathSegments[len(pathSegments)-1]

	return lastSegment, nil
}

func (c *Client) getRPCLimiter(URL string) (*chain.RPCLimiter, error) {
	apiKey, err := extractLastParamFromURL(URL)
	if err != nil {
		return nil, err
	}
	c.rpcLimiterMutex.Lock()
	defer c.rpcLimiterMutex.Unlock()
	if limiter, ok := c.limiterPerProvider[apiKey]; ok {
		return limiter, nil
	}
	limiter := chain.NewRPCLimiter()
	c.limiterPerProvider[apiKey] = limiter
	return limiter, nil
}

func (c *Client) getClientUsingCache(chainID uint64) (chain.ClientInterface, error) {
	c.rpcClientsMutex.Lock()
	defer c.rpcClientsMutex.Unlock()
	if rpcClient, ok := c.rpcClients[chainID]; ok {
		if rpcClient.GetWalletNotifier() == nil {
			rpcClient.SetWalletNotifier(c.walletNotifier)
		}
		return rpcClient, nil
	}

	network := c.NetworkManager.Find(chainID)
	if network == nil {
		if c.UpstreamChainID == chainID {
			return c.upstream, nil
		}
		return nil, fmt.Errorf("could not find network: %d", chainID)
	}

	rpcClient, err := gethrpc.Dial(network.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("dial upstream server: %s", err)
	}

	rpcLimiter, err := c.getRPCLimiter(network.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("get RPC limiter: %s", err)
	}

	var (
		rpcFallbackClient  *gethrpc.Client
		rpcFallbackLimiter *chain.RPCLimiter
	)
	if len(network.FallbackURL) > 0 {
		rpcFallbackClient, err = gethrpc.Dial(network.FallbackURL)
		if err != nil {
			return nil, fmt.Errorf("dial upstream server: %s", err)
		}

		rpcFallbackLimiter, err = c.getRPCLimiter(network.FallbackURL)
		if err != nil {
			return nil, fmt.Errorf("get RPC fallback limiter: %s", err)
		}
	}

	client := chain.NewClient(rpcLimiter, rpcClient, rpcFallbackLimiter, rpcFallbackClient, chainID)
	client.WalletNotifier = c.walletNotifier
	c.rpcClients[chainID] = client
	return client, nil
}

// Ethclient returns ethclient.Client per chain
func (c *Client) EthClient(chainID uint64) (chain.ClientInterface, error) {
	client, err := c.getClientUsingCache(chainID)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// AbstractEthClient returns a partial abstraction used by new components for testing purposes
func (c *Client) AbstractEthClient(chainID common.ChainID) (chain.BatchCallClient, error) {
	client, err := c.getClientUsingCache(uint64(chainID))
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) EthClients(chainIDs []uint64) (map[uint64]chain.ClientInterface, error) {
	clients := make(map[uint64]chain.ClientInterface, 0)
	for _, chainID := range chainIDs {
		client, err := c.getClientUsingCache(chainID)
		if err != nil {
			return nil, err
		}
		clients[chainID] = client
	}

	return clients, nil
}

// SetClient strictly for testing purposes
func (c *Client) SetClient(chainID uint64, client chain.ClientInterface) {
	c.rpcClientsMutex.Lock()
	defer c.rpcClientsMutex.Unlock()
	c.rpcClients[chainID] = client
}

// UpdateUpstreamURL changes the upstream RPC client URL, if the upstream is enabled.
func (c *Client) UpdateUpstreamURL(url string) error {
	if c.upstream == nil {
		return nil
	}

	rpcClient, err := gethrpc.Dial(url)
	if err != nil {
		return err
	}
	rpcLimiter, err := c.getRPCLimiter(url)
	if err != nil {
		return err
	}
	c.Lock()
	c.upstream = chain.NewSimpleClient(rpcLimiter, rpcClient, c.UpstreamChainID)
	c.upstreamURL = url
	c.Unlock()

	return nil
}

// Call performs a JSON-RPC call with the given arguments and unmarshals into
// result if no error occurred.
//
// The result must be a pointer so that package json can unmarshal into it. You
// can also pass nil, in which case the result is ignored.
//
// It uses custom routing scheme for calls.
func (c *Client) Call(result interface{}, chainID uint64, method string, args ...interface{}) error {
	ctx := context.Background()
	return c.CallContext(ctx, result, chainID, method, args...)
}

// CallContext performs a JSON-RPC call with the given arguments. If the context is
// canceled before the call has successfully returned, CallContext returns immediately.
//
// The result must be a pointer so that package json can unmarshal into it. You
// can also pass nil, in which case the result is ignored.
//
// It uses custom routing scheme for calls.
// If there are any local handlers registered for this call, they will handle it.
func (c *Client) CallContext(ctx context.Context, result interface{}, chainID uint64, method string, args ...interface{}) error {
	rpcstats.CountCall(method)
	if c.router.routeBlocked(method) {
		return ErrMethodNotFound
	}

	// check locally registered handlers first
	if handler, ok := c.handler(method); ok {
		return c.callMethod(ctx, result, chainID, handler, args...)
	}

	return c.CallContextIgnoringLocalHandlers(ctx, result, chainID, method, args...)
}

// CallContextIgnoringLocalHandlers performs a JSON-RPC call with the given
// arguments.
//
// If there are local handlers registered for this call, they would
// be ignored. It is useful if the call is happening from within a local
// handler itself.
// Upstream calls routing will be used anyway.
func (c *Client) CallContextIgnoringLocalHandlers(ctx context.Context, result interface{}, chainID uint64, method string, args ...interface{}) error {
	if c.router.routeBlocked(method) {
		return ErrMethodNotFound
	}

	if c.router.routeRemote(method) {
		client, err := c.getClientUsingCache(chainID)
		if err != nil {
			return err
		}
		return client.CallContext(ctx, result, method, args...)
	}

	if c.local == nil {
		c.log.Warn("Local JSON-RPC endpoint missing", "method", method)
		return errors.New("missing local JSON-RPC endpoint")
	}
	return c.local.CallContext(ctx, result, method, args...)
}

// RegisterHandler registers local handler for specific RPC method.
//
// If method is registered, it will be executed with given handler and
// never routed to the upstream or local servers.
func (c *Client) RegisterHandler(method string, handler Handler) {
	c.handlersMx.Lock()
	defer c.handlersMx.Unlock()

	c.handlers[method] = handler
}

// UnregisterHandler removes a previously registered handler.
func (c *Client) UnregisterHandler(method string) {
	c.handlersMx.Lock()
	defer c.handlersMx.Unlock()

	delete(c.handlers, method)
}

// callMethod calls registered RPC handler with given args and pointer to result.
// It handles proper params and result converting
//
// TODO(divan): use cancellation via context here?
func (c *Client) callMethod(ctx context.Context, result interface{}, chainID uint64, handler Handler, args ...interface{}) error {
	response, err := handler(ctx, chainID, args...)
	if err != nil {
		return err
	}

	// if result is nil, just ignore result -
	// the same way as gethrpc.CallContext() caller would expect
	if result == nil {
		return nil
	}

	return setResultFromRPCResponse(result, response)
}

// handler is a concurrently safe method to get registered handler by name.
func (c *Client) handler(method string) (Handler, bool) {
	c.handlersMx.RLock()
	defer c.handlersMx.RUnlock()
	handler, ok := c.handlers[method]
	return handler, ok
}

// setResultFromRPCResponse tries to set result value from response using reflection
// as concrete types are unknown.
func setResultFromRPCResponse(result, response interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("invalid result type: %s", r)
		}
	}()

	responseValue := reflect.ValueOf(response)

	// If it is called via CallRaw, result has type json.RawMessage and
	// we should marshal the response before setting it.
	// Otherwise, it is called with CallContext and result is of concrete type,
	// thus we should try to set it as it is.
	// If response type and result type are incorrect, an error should be returned.
	// TODO(divan): add additional checks for result underlying value, if needed:
	// some example: https://golang.org/src/encoding/json/decode.go#L596
	switch reflect.ValueOf(result).Elem().Type() {
	case reflect.TypeOf(json.RawMessage{}), reflect.TypeOf([]byte{}):
		data, err := json.Marshal(response)
		if err != nil {
			return err
		}

		responseValue = reflect.ValueOf(data)
	}

	value := reflect.ValueOf(result).Elem()
	if !value.CanSet() {
		return errors.New("can't assign value to result")
	}
	value.Set(responseValue)

	return nil
}
