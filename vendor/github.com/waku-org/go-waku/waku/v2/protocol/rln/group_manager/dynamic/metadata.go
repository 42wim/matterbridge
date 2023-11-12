package dynamic

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/waku-org/go-waku/waku/v2/protocol/rln/group_manager"
)

// RLNMetadata persists attributes in the RLN database
type RLNMetadata struct {
	LastProcessedBlock uint64
	ChainID            *big.Int
	ContractAddress    common.Address
	ValidRootsPerBlock []group_manager.RootsPerBlock
}

// Serialize converts a RLNMetadata into a binary format expected by zerokit's RLN
func (r RLNMetadata) Serialize() ([]byte, error) {
	chainID := r.ChainID
	if chainID == nil {
		return nil, errors.New("chain-id not specified and cannot be 0")
	}

	var result []byte
	result = binary.LittleEndian.AppendUint64(result, r.LastProcessedBlock)
	result = binary.LittleEndian.AppendUint64(result, chainID.Uint64())
	result = append(result, r.ContractAddress.Bytes()...)
	result = binary.LittleEndian.AppendUint64(result, uint64(len(r.ValidRootsPerBlock)))
	for _, v := range r.ValidRootsPerBlock {
		result = append(result, v.Root[:]...)
		result = binary.LittleEndian.AppendUint64(result, v.BlockNumber)
	}

	return result, nil
}

const lastProcessedBlockOffset = 0
const chainIDOffset = lastProcessedBlockOffset + 8
const contractAddressOffset = chainIDOffset + 8
const validRootsLenOffset = contractAddressOffset + 20
const validRootsValOffset = validRootsLenOffset + 8
const metadataByteLen = 8 + 8 + 20 + 8 // uint64 + uint64 + ethAddress + uint64

// DeserializeMetadata converts a byte slice into a RLNMetadata instance
func DeserializeMetadata(b []byte) (RLNMetadata, error) {
	if len(b) < metadataByteLen {
		return RLNMetadata{}, errors.New("wrong size")
	}

	validRootLen := binary.LittleEndian.Uint64(b[validRootsLenOffset:validRootsValOffset])
	if len(b) < int(metadataByteLen+validRootLen*(32+8)) { // len of a merkle node and len of a uint64 for the block number
		return RLNMetadata{}, errors.New("wrong size")
	}

	validRoots := make([]group_manager.RootsPerBlock, 0, validRootLen)
	for i := 0; i < int(validRootLen); i++ {
		rootOffset := validRootsValOffset + (i * (32 + 8))
		blockOffset := rootOffset + 32
		root := group_manager.RootsPerBlock{
			BlockNumber: binary.LittleEndian.Uint64(b[blockOffset : blockOffset+8]),
		}
		copy(root.Root[:], b[rootOffset:blockOffset])
		validRoots = append(validRoots, root)
	}

	return RLNMetadata{
		LastProcessedBlock: binary.LittleEndian.Uint64(b[lastProcessedBlockOffset:chainIDOffset]),
		ChainID:            new(big.Int).SetUint64(binary.LittleEndian.Uint64(b[chainIDOffset:contractAddressOffset])),
		ContractAddress:    common.BytesToAddress(b[contractAddressOffset:validRootsLenOffset]),
		ValidRootsPerBlock: validRoots,
	}, nil
}

// SetMetadata stores some metadata into the zerokit's RLN database
func (gm *DynamicGroupManager) SetMetadata(meta RLNMetadata) error {
	b, err := meta.Serialize()
	if err != nil {
		return err
	}
	return gm.rln.SetMetadata(b)
}
