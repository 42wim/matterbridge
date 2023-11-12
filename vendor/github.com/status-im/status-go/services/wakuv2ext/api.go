package wakuv2ext

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/services/ext"
)

// PublicAPI extends waku public API.
type PublicAPI struct {
	*ext.PublicAPI
	service   *Service
	publicAPI types.PublicWakuAPI
	log       log.Logger
}

// NewPublicAPI returns instance of the public API.
func NewPublicAPI(s *Service) *PublicAPI {
	return &PublicAPI{
		PublicAPI: ext.NewPublicAPI(s.Service, s.w),
		service:   s,
		publicAPI: s.w.PublicWakuAPI(),
		log:       log.New("package", "status-go/services/wakuext.PublicAPI"),
	}
}
