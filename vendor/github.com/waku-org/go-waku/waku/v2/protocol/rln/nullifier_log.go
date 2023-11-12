package rln

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/waku-org/go-zerokit-rln/rln"
	"go.uber.org/zap"
)

// NullifierLog is the log of nullifiers and Shamir shares of the past messages grouped per epoch
type NullifierLog struct {
	sync.RWMutex

	log            *zap.Logger
	nullifierLog   map[rln.Nullifier][]rln.ProofMetadata // Might make sense to replace this map by a shrinkable map due to https://github.com/golang/go/issues/20135.
	nullifierQueue []rln.Nullifier
}

// NewNullifierLog creates an instance of NullifierLog
func NewNullifierLog(ctx context.Context, log *zap.Logger) *NullifierLog {
	result := &NullifierLog{
		nullifierLog: make(map[rln.Nullifier][]rln.ProofMetadata),
		log:          log,
	}

	go result.cleanup(ctx)

	return result
}

var errAlreadyExists = errors.New("proof already exists")

// Insert stores a proof in the nullifier log only if it doesnt exist already
func (n *NullifierLog) Insert(proofMD rln.ProofMetadata) error {
	n.Lock()
	defer n.Unlock()

	proofs, ok := n.nullifierLog[proofMD.ExternalNullifier]
	if ok {
		// check if an identical record exists
		for _, p := range proofs {
			if p.Equals(proofMD) {
				// TODO: slashing logic
				return errAlreadyExists
			}
		}
	}

	n.nullifierLog[proofMD.ExternalNullifier] = append(proofs, proofMD)
	n.nullifierQueue = append(n.nullifierQueue, proofMD.ExternalNullifier)
	return nil
}

// HasDuplicate returns true if there is another message in the  `nullifierLog` with the same
// epoch and nullifier as `msg`'s epoch and nullifier but different Shamir secret shares
// otherwise, returns false
func (n *NullifierLog) HasDuplicate(proofMD rln.ProofMetadata) (bool, error) {
	n.RLock()
	defer n.RUnlock()

	proofs, ok := n.nullifierLog[proofMD.ExternalNullifier]
	if !ok {
		// epoch does not exist
		return false, nil
	}

	for _, p := range proofs {
		if p.Equals(proofMD) {
			// there is an identical record, ignore the msg
			return true, nil
		}
	}

	// check for a message with the same nullifier but different secret shares
	matched := false
	for _, it := range proofs {
		if bytes.Equal(it.Nullifier[:], proofMD.Nullifier[:]) && (!bytes.Equal(it.ShareX[:], proofMD.ShareX[:]) || !bytes.Equal(it.ShareY[:], proofMD.ShareY[:])) {
			matched = true
			break
		}
	}

	return matched, nil
}

// cleanup cleans up the log every time there are more than MaxEpochGap epochs stored in it
func (n *NullifierLog) cleanup(ctx context.Context) {
	t := time.NewTicker(1 * time.Minute) // TODO: tune this
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-t.C:
			func() {
				n.Lock()
				defer n.Unlock()

				if int64(len(n.nullifierQueue)) < maxEpochGap {
					return
				}

				n.log.Debug("clearing epochs from the nullifier log", zap.Int64("count", maxEpochGap))

				toDelete := n.nullifierQueue[0:maxEpochGap]
				for _, l := range toDelete {
					delete(n.nullifierLog, l)
				}
				n.nullifierQueue = n.nullifierQueue[maxEpochGap:]
			}()
		}
	}

}
