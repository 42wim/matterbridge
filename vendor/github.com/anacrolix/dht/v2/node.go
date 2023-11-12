package dht

import (
	"time"

	"github.com/anacrolix/dht/v2/int160"
	"github.com/anacrolix/dht/v2/krpc"
)

type nodeKey struct {
	Addr Addr
	Id   int160.T
}

type node struct {
	nodeKey

	lastGotQuery    time.Time // From the remote node
	lastGotResponse time.Time // From the remote node

	numReceivesFrom            int
	failedLastQuestionablePing bool
}

func (s *Server) IsQuestionable(n *node) bool {
	return !s.IsGood(n) && !s.nodeIsBad(n)
}

func (n *node) hasAddrAndID(addr Addr, id int160.T) bool {
	return id == n.Id && n.Addr.String() == addr.String()
}

func (n *node) IsSecure() bool {
	return NodeIdSecure(n.Id.AsByteArray(), n.Addr.IP())
}

func (n *node) idString() string {
	return n.Id.ByteString()
}

func (n *node) NodeInfo() (ret krpc.NodeInfo) {
	ret.Addr = n.Addr.KRPC()
	if n := copy(ret.ID[:], n.idString()); n != 20 {
		panic(n)
	}
	return
}

// Per the spec in BEP 5.
func (s *Server) IsGood(n *node) bool {
	if s.nodeIsBad(n) {
		return false
	}
	return time.Since(n.lastGotResponse) < 15*time.Minute ||
		!n.lastGotResponse.IsZero() && time.Since(n.lastGotQuery) < 15*time.Minute
}
