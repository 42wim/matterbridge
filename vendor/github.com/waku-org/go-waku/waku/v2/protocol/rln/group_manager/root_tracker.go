package group_manager

import (
	"bytes"
	"sync"

	"github.com/waku-org/go-waku/waku/v2/utils"
	"github.com/waku-org/go-zerokit-rln/rln"
	"go.uber.org/zap"
)

// RootsPerBlock stores the merkle root generated at N block number
type RootsPerBlock struct {
	Root        rln.MerkleNode
	BlockNumber uint64
}

// MerkleRootTracker keeps track of the latest N merkle roots considered
// valid for RLN proofs.
type MerkleRootTracker struct {
	sync.RWMutex

	rln                      *rln.RLN
	acceptableRootWindowSize int
	validMerkleRoots         []RootsPerBlock
	merkleRootBuffer         []RootsPerBlock
}

const maxBufferSize = 20

// NewMerkleRootTracker creates an instance of MerkleRootTracker
func NewMerkleRootTracker(acceptableRootWindowSize int, rlnInstance *rln.RLN) *MerkleRootTracker {
	result := &MerkleRootTracker{
		acceptableRootWindowSize: acceptableRootWindowSize,
		rln:                      rlnInstance,
	}

	result.UpdateLatestRoot(0)

	return result
}

// Backfill is used to pop merkle roots when there is a chain fork
func (m *MerkleRootTracker) Backfill(fromBlockNumber uint64) {
	m.Lock()
	defer m.Unlock()

	numBlocks := 0
	for i := len(m.validMerkleRoots) - 1; i >= 0; i-- {
		if m.validMerkleRoots[i].BlockNumber >= fromBlockNumber {
			numBlocks++
		}
	}

	if numBlocks == 0 {
		return
	}

	// Remove last roots
	rootsToPop := numBlocks
	if len(m.validMerkleRoots) < rootsToPop {
		rootsToPop = len(m.validMerkleRoots)
	}
	m.validMerkleRoots = m.validMerkleRoots[0 : len(m.validMerkleRoots)-rootsToPop]

	if len(m.merkleRootBuffer) == 0 {
		return
	}

	// Backfill the tree's acceptable roots
	rootsToRestore := numBlocks
	bufferLen := len(m.merkleRootBuffer)
	if bufferLen < rootsToRestore {
		rootsToRestore = bufferLen
	}
	for i := 0; i < rootsToRestore; i++ {
		x, newRootBuffer := m.merkleRootBuffer[len(m.merkleRootBuffer)-1], m.merkleRootBuffer[:len(m.merkleRootBuffer)-1] // Pop
		m.validMerkleRoots = append([]RootsPerBlock{x}, m.validMerkleRoots...)
		m.merkleRootBuffer = newRootBuffer
	}
}

// ContainsRoot is used to check whether a merkle tree root is contained in the list of valid merkle roots or not
func (m *MerkleRootTracker) ContainsRoot(root [32]byte) bool {
	return m.IndexOf(root) > -1
}

// IndexOf returns the index of a root if present in the list of valid merkle roots
func (m *MerkleRootTracker) IndexOf(root [32]byte) int {
	m.RLock()
	defer m.RUnlock()

	for i := range m.validMerkleRoots {
		if bytes.Equal(m.validMerkleRoots[i].Root[:], root[:]) {
			return i
		}
	}

	return -1
}

// UpdateLatestRoot should be called when a block containing a new
// IDCommitment is received so we can keep track of the merkle root change
func (m *MerkleRootTracker) UpdateLatestRoot(blockNumber uint64) rln.MerkleNode {
	m.Lock()
	defer m.Unlock()

	root, err := m.rln.GetMerkleRoot()
	if err != nil {
		utils.Logger().Named("root-tracker").Panic("could not retrieve merkle root", zap.Error(err))
	}

	m.pushRoot(blockNumber, root)

	return root
}

func (m *MerkleRootTracker) pushRoot(blockNumber uint64, root [32]byte) {
	m.validMerkleRoots = append(m.validMerkleRoots, RootsPerBlock{
		Root:        root,
		BlockNumber: blockNumber,
	})

	// Maintain valid merkle root window
	if len(m.validMerkleRoots) > m.acceptableRootWindowSize {
		m.merkleRootBuffer = append(m.merkleRootBuffer, m.validMerkleRoots[0])
		m.validMerkleRoots = m.validMerkleRoots[1:]
	}

	// Maintain merkle root buffer
	if len(m.merkleRootBuffer) > maxBufferSize {
		m.merkleRootBuffer = m.merkleRootBuffer[1:]
	}
}

// Roots return the list of valid merkle roots
func (m *MerkleRootTracker) Roots() []rln.MerkleNode {
	m.RLock()
	defer m.RUnlock()

	result := make([]rln.MerkleNode, len(m.validMerkleRoots))
	for i := range m.validMerkleRoots {
		result[i] = m.validMerkleRoots[i].Root
	}

	return result
}

// Buffer is used as a repository of older merkle roots that although
// they were valid once, they have left the acceptable window of
// merkle roots. We keep track of them in case a chain fork occurs
// and we need to restore the valid merkle roots to a previous point
// of time
func (m *MerkleRootTracker) Buffer() []rln.MerkleNode {
	m.RLock()
	defer m.RUnlock()

	result := make([]rln.MerkleNode, len(m.merkleRootBuffer))
	for i := range m.merkleRootBuffer {
		result[i] = m.merkleRootBuffer[i].Root
	}

	return result
}

// ValidRootsPerBlock returns the current valid merkle roots and block numbers
func (m *MerkleRootTracker) ValidRootsPerBlock() []RootsPerBlock {
	m.RLock()
	defer m.RUnlock()

	return m.validMerkleRoots
}

// SetValidRootsPerBlock is used to overwrite the valid merkle roots
func (m *MerkleRootTracker) SetValidRootsPerBlock(roots []RootsPerBlock) {
	m.Lock()
	defer m.Unlock()

	m.validMerkleRoots = roots
}
