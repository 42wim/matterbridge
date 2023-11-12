package rpc

// router implements logic for routing
// JSON-RPC requests either to Upstream or
// Local node.
type router struct {
	methods         map[string]bool
	blockedMethods  map[string]struct{}
	upstreamEnabled bool
}

// newRouter inits new router.
func newRouter(upstreamEnabled bool) *router {
	r := &router{
		methods:         make(map[string]bool),
		blockedMethods:  make(map[string]struct{}),
		upstreamEnabled: upstreamEnabled,
	}

	for _, m := range remoteMethods {
		r.methods[m] = true
	}

	for _, m := range blockedMethods {
		r.blockedMethods[m] = struct{}{}
	}

	return r
}

// routeRemote returns true if given method should be routed to the remote node
func (r *router) routeRemote(method string) bool {
	if !r.upstreamEnabled {
		return false
	}

	// else check route using the methods list
	return r.methods[method]
}

func (r *router) routeBlocked(method string) bool {
	_, ok := r.blockedMethods[method]
	return ok
}

// blockedMethods is a list of dangerous or having security implications JSON-RPC methods
// that are not allowed to be called.
var blockedMethods = [...]string{
	"shh_getPrivateKey",
}

// BlockedMethods returns a list of methods that are not allowed to be called.
// A copy of a slice is returned in order to prevent from changing it from outside.
func BlockedMethods() []string {
	return append([]string(nil), blockedMethods[:]...)
}

// remoteMethods contains methods that should be routed to
// the upstream node; the rest is considered to be routed to
// the local node.
// A list of supported methods:
//
//	curl --include \
//	   --header "Content-Type: application/json" \
//	   --header "Accept: application/json" 'https://api.infura.io/v1/jsonrpc/ropsten/methods'
//
// Although it's tempting to only list methods coming to the local node as there're fewer of them
// but it's deceptive: we want to ensure that only known requests leave our zone of responsibility.
// Also, we want new requests in newer Geth versions not to be accidentally routed to the upstream.
// The list of methods: https://github.com/ethereum/wiki/wiki/JSON-RPC
var remoteMethods = [...]string{
	"eth_protocolVersion",
	"eth_syncing",
	"eth_coinbase",
	"eth_mining",
	"eth_hashrate",
	"eth_gasPrice",
	"eth_maxPriorityFeePerGas",
	"eth_feeHistory",
	//"eth_accounts", // due to sub-accounts handling
	"eth_blockNumber",
	"eth_getBalance",
	"eth_getStorageAt",
	"eth_getTransactionCount",
	"eth_getBlockTransactionCountByHash",
	"eth_getBlockTransactionCountByNumber",
	"eth_getUncleCountByBlockHash",
	"eth_getUncleCountByBlockNumber",
	"eth_getCode",
	//"eth_sign", // only the local node has an injected account to sign the payload with
	//"eth_sendTransaction", // we handle this specially calling eth_estimateGas, signing it locally and sending eth_sendRawTransaction afterwards
	"eth_sendRawTransaction",
	"eth_call",
	"eth_estimateGas",
	"eth_getBlockByHash",
	"eth_getBlockByNumber",
	"eth_getTransactionByHash",
	"eth_getTransactionByBlockHashAndIndex",
	"eth_getTransactionByBlockNumberAndIndex",
	"eth_getTransactionReceipt",
	"eth_getUncleByBlockHashAndIndex",
	"eth_getUncleByBlockNumberAndIndex",
	//"eth_getCompilers",    // goes to the local because there's no need to send it anywhere
	//"eth_compileLLL",      // goes to the local because there's no need to send it anywhere
	//"eth_compileSolidity", // goes to the local because there's no need to send it anywhere
	//"eth_compileSerpent",  // goes to the local because there's no need to send it anywhere

	"eth_getLogs",
	"eth_getWork",
	"eth_submitWork",
	"eth_submitHashrate",
	"net_version",
	"net_peerCount",
	"net_listening",
}
