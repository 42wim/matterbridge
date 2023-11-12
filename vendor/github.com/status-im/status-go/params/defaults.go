package params

import "github.com/ethereum/go-ethereum/p2p/discv5"

const (
	// StatusDatabase path relative to DataDir.
	StatusDatabase = "status-db"

	// SendTransactionMethodName https://docs.walletconnect.com/advanced/rpc-reference/ethereum-rpc#eth_sendtransaction
	SendTransactionMethodName = "eth_sendTransaction"

	// SendTransactionMethodName https://docs.walletconnect.com/advanced/rpc-reference/ethereum-rpc#eth_sendrawtransaction
	SendRawTransactionMethodName = "eth_sendRawTransaction"

	BalanceMethodName = "eth_getBalance"

	// AccountsMethodName defines the name for listing the currently signed accounts.
	AccountsMethodName = "eth_accounts"

	// PersonalSignMethodName https://docs.walletconnect.com/advanced/rpc-reference/ethereum-rpc#personal_sign
	PersonalSignMethodName = "personal_sign"

	// SignMethodName https://docs.walletconnect.com/advanced/rpc-reference/ethereum-rpc#eth_sign
	SignMethodName = "eth_sign"

	// SignTransactionMethodName https://docs.walletconnect.com/advanced/rpc-reference/ethereum-rpc#eth_signtransaction
	SignTransactionMethodName = "eth_signTransaction"

	// SignTypedDataMethodName https://docs.walletconnect.com/advanced/rpc-reference/ethereum-rpc#eth_signtypeddata
	SignTypedDataMethodName   = "eth_signTypedData"
	SignTypedDataV3MethodName = "eth_signTypedData_v3"
	SignTypedDataV4MethodName = "eth_signTypedData_v4"

	WalletSwitchEthereumChainMethodName = "wallet_switchEthereumChain"

	// PersonalRecoverMethodName defines the name for `personal.recover` API.
	PersonalRecoverMethodName = "personal_ecRecover"

	// DefaultGas default amount of gas used for transactions
	DefaultGas = 180000

	// WhisperMinimumPoW amount of work for Whisper message to be added to sending queue
	// We enforce a minimum as a bland spam prevention mechanism.
	WhisperMinimumPoW = 0.000002

	// WhisperTTL is time to live for messages, in seconds
	WhisperTTL = 120

	// WakuMinimumPoW amount of work for Whisper message to be added to sending queue
	// We enforce a minimum as a bland spam prevention mechanism.
	WakuMinimumPoW = 0.000002

	// WakuTTL is time to live for messages, in seconds
	WakuTTL = 120

	// MainnetEthereumNetworkURL is URL where the upstream ethereum network is loaded to
	// allow us avoid syncing node.
	MainnetEthereumNetworkURL = "https://mainnet.infura.io/nKmXgiFgc2KqtoQ8BCGJ"

	// GoerliEthereumNetworkURL is an open RPC endpoint to Goerli network
	// Other RPC endpoints are available here: http://goerli.blockscout.com/
	GoerliEthereumNetworkURL = "http://goerli.blockscout.com/"

	// MainNetworkID is id of the main network
	MainNetworkID = 1

	// GoerliNetworkID is id of goerli test network (PoA)
	GoerliNetworkID = 5

	// StatusChainNetworkID is id of a test network (private chain)
	StatusChainNetworkID = 777

	// WhisperDiscv5Topic used to register and search for whisper peers using discovery v5.
	WhisperDiscv5Topic = discv5.Topic("whisper")

	// MailServerDiscv5Topic used to register and search for mail server peers using discovery v5.
	MailServerDiscv5Topic = discv5.Topic("whispermail")

	// LESDiscoveryIdentifier is a prefix for topic used for LES peers discovery.
	LESDiscoveryIdentifier = "LES2@"
)

var (
	// WhisperDiscv5Limits declares min and max limits for peers with whisper topic.
	WhisperDiscv5Limits = Limits{2, 2}
	// LesDiscoveryLimits default limits used if LES and discovery are enabled.
	LesDiscoveryLimits = Limits{2, 2}
)
