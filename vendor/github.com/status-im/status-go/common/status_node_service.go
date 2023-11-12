package common

import (
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
)

type StatusService interface {
	Start() error
	Stop() error
	Protocols() []p2p.Protocol
	APIs() []rpc.API
}
