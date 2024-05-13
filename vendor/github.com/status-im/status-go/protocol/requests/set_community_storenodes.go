package requests

import (
	"bytes"
	"errors"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/storenodes"
)

var (
	ErrSetCommunityStorenodesEmpty            = errors.New("set-community-storenodes: empty payload")
	ErrSetCommunityStorenodesTooMany          = errors.New("set-community-storenodes: too many")
	ErrSetCommunityStorenodesMismatch         = errors.New("set-community-storenodes: communityId mismatch")
	ErrSetCommunityStorenodesMissingCommunity = errors.New("set-community-storenodes: missing community")
	ErrSetCommunityStorenodesBadVersion       = errors.New("set-community-storenodes: bad version")
)

type SetCommunityStorenodes struct {
	CommunityID types.HexBytes         `json:"communityId"`
	Storenodes  []storenodes.Storenode `json:"storenodes"`
}

func (s *SetCommunityStorenodes) Validate() error {
	if s == nil || len(s.Storenodes) == 0 {
		return ErrSetCommunityStorenodesEmpty
	}
	if len(s.Storenodes) > 1 {
		// TODO for now only allow one
		return ErrSetCommunityStorenodesTooMany
	}
	if len(s.CommunityID) == 0 {
		return ErrSetCommunityStorenodesMissingCommunity
	}
	for _, sn := range s.Storenodes {
		if !bytes.Equal(sn.CommunityID, s.CommunityID) {
			return ErrSetCommunityStorenodesMismatch
		}
		if sn.Version == 0 {
			return ErrSetCommunityStorenodesBadVersion
		}
		// TODO validate address and other fields
	}
	return nil
}
