package status

import (
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol"
	"github.com/status-im/status-go/protocol/common/shard"
)

// Make sure that Service implements node.Lifecycle interface.
var _ node.Lifecycle = (*Service)(nil)
var ErrNotInitialized = errors.New("status public api not initialized")

// Service represents out own implementation of personal sign operations.
type Service struct {
	messenger *protocol.Messenger
}

// New returns a new Service.
func New() *Service {
	return &Service{}
}

func (s *Service) Init(messenger *protocol.Messenger) {
	s.messenger = messenger
}

// Protocols returns a new protocols list. In this case, there are none.
func (s *Service) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

// APIs returns a list of new APIs.
func (s *Service) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "status",
			Version:   "1.0",
			Service:   NewPublicAPI(s),
			Public:    true,
		},
	}
}

// NewPublicAPI returns a reference to the PublicAPI object
func NewPublicAPI(s *Service) *PublicAPI {
	api := &PublicAPI{
		service: s,
	}
	return api
}

// Start is run when a service is started.
func (s *Service) Start() error {
	return nil
}

// Stop is run when a service is stopped.
func (s *Service) Stop() error {
	return nil
}

type PublicAPI struct {
	service *Service
}

func (p *PublicAPI) CommunityInfo(communityID types.HexBytes, shard *shard.Shard) (json.RawMessage, error) {
	if p.service.messenger == nil {
		return nil, ErrNotInitialized
	}

	community, err := p.service.messenger.FetchCommunity(&protocol.FetchCommunityRequest{
		CommunityKey:    communityID.String(),
		Shard:           shard,
		TryDatabase:     true,
		WaitForResponse: true,
	})
	if err != nil {
		return nil, err
	}

	return community.MarshalPublicAPIJSON()
}
