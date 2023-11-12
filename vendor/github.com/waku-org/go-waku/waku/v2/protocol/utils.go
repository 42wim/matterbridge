package protocol

import (
	"strings"

	"github.com/libp2p/go-libp2p/core/protocol"
)

const GossipSubOptimalFullMeshSize = 6

// FulltextMatch is the default matching function used for checking if a peer
// supports a protocol or not
func FulltextMatch(expectedProtocol string) func(string) bool {
	return func(receivedProtocol string) bool {
		return receivedProtocol == expectedProtocol
	}
}

// PrefixTextMatch is a matching function used for checking if a peer's
// supported protocols begin with a particular prefix
func PrefixTextMatch(prefix string) func(protocol.ID) bool {
	return func(receivedProtocol protocol.ID) bool {
		return strings.HasPrefix(string(receivedProtocol), prefix)
	}
}
