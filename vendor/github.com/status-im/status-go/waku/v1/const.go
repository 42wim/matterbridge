package v1

// Waku protocol parameters
const (
	Version    = uint64(1) // Peer version number
	VersionStr = "1"       // The same, as a string
	Name       = "waku"    // Nickname of the protocol

	// Waku protocol message codes, according to https://github.com/vacp2p/specs/blob/master/specs/waku/waku-0.md
	statusCode             = 0   // used in the handshake
	messagesCode           = 1   // regular message
	statusUpdateCode       = 22  // update of settings
	batchAcknowledgedCode  = 11  // confirmation that batch of envelopes was received
	messageResponseCode    = 12  // includes confirmation for delivery and information about errors
	p2pRequestCompleteCode = 125 // peer-to-peer message, used by Dapp protocol
	p2pRequestCode         = 126 // peer-to-peer message, used by Dapp protocol
	p2pMessageCode         = 127 // peer-to-peer message (to be consumed by the peer, but not forwarded any further)
	NumberOfMessageCodes   = 128
)
