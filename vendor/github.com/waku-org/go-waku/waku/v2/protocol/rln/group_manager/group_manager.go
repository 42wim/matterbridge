package group_manager

import (
	"context"

	"github.com/waku-org/go-zerokit-rln/rln"
)

type GroupManager interface {
	Start(ctx context.Context) error
	IdentityCredentials() (rln.IdentityCredential, error)
	MembershipIndex() rln.MembershipIndex
	Stop() error
	IsReady(ctx context.Context) (bool, error)
}

type Details struct {
	GroupManager GroupManager
	RootTracker  *MerkleRootTracker

	RLN *rln.RLN
}
