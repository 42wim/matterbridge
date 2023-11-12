package static

import (
	"context"
	"errors"

	"github.com/waku-org/go-waku/waku/v2/protocol/rln/group_manager"
	"github.com/waku-org/go-zerokit-rln/rln"
	"go.uber.org/zap"
)

type StaticGroupManager struct {
	rln *rln.RLN
	log *zap.Logger

	identityCredential *rln.IdentityCredential
	membershipIndex    rln.MembershipIndex

	group       []rln.IDCommitment
	rootTracker *group_manager.MerkleRootTracker
	nextIndex   uint64
}

func NewStaticGroupManager(
	group []rln.IDCommitment,
	identityCredential rln.IdentityCredential,
	index rln.MembershipIndex,
	rlnInstance *rln.RLN,
	rootTracker *group_manager.MerkleRootTracker,
	log *zap.Logger,
) (*StaticGroupManager, error) {
	// check the peer's index and the inclusion of user's identity commitment in the group
	if identityCredential.IDCommitment != group[int(index)] {
		return nil, errors.New("peer's IDCommitment does not match commitment in group")
	}

	return &StaticGroupManager{
		log:                log.Named("rln-static"),
		group:              group,
		identityCredential: &identityCredential,
		membershipIndex:    index,
		rln:                rlnInstance,
		rootTracker:        rootTracker,
	}, nil
}

func (gm *StaticGroupManager) Start(ctx context.Context) error {
	gm.log.Info("mounting rln-relay in off-chain/static mode")

	// add members to the Merkle tree

	err := gm.insertMembers(gm.group)
	if err != nil {
		return err
	}

	gm.group = nil // Deleting group to release memory

	return nil
}

func (gm *StaticGroupManager) insertMembers(idCommitments []rln.IDCommitment) error {
	err := gm.rln.InsertMembers(rln.MembershipIndex(gm.nextIndex), idCommitments)
	if err != nil {
		gm.log.Error("inserting members into merkletree", zap.Error(err))
		return err
	}

	latestIndex := gm.nextIndex + uint64(len(idCommitments))

	gm.rootTracker.UpdateLatestRoot(latestIndex)

	gm.nextIndex = latestIndex + 1

	return nil
}

func (gm *StaticGroupManager) IdentityCredentials() (rln.IdentityCredential, error) {
	if gm.identityCredential == nil {
		return rln.IdentityCredential{}, errors.New("identity credential has not been setup")
	}

	return *gm.identityCredential, nil
}

func (gm *StaticGroupManager) MembershipIndex() rln.MembershipIndex {
	return gm.membershipIndex
}

// Stop is a function created just to comply with the GroupManager interface (it does nothing)
func (gm *StaticGroupManager) Stop() error {
	// Do nothing
	return nil
}

func (gm *StaticGroupManager) IsReady(ctx context.Context) (bool, error) {
	return true, nil
}
