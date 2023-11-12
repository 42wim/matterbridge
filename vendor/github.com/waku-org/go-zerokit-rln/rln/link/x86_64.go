//go:build (linux || windows) && amd64 && !android
// +build linux windows
// +build amd64
// +build !android

package link

import r "github.com/waku-org/go-zerokit-rln-x86_64/rln"

type RLNWrapper struct {
	ffi *r.RLN
}

func NewWithParams(depth int, wasm []byte, zkey []byte, verifKey []byte, treeConfig []byte) (*RLNWrapper, error) {
	rln, err := r.NewWithParams(depth, wasm, zkey, verifKey, treeConfig)
	if err != nil {
		return nil, err
	}
	return &RLNWrapper{ffi: rln}, nil
}

func New(depth int, config []byte) (*RLNWrapper, error) {
	rln, err := r.New(uint(depth), config)
	if err != nil {
		return nil, err
	}
	return &RLNWrapper{ffi: rln}, nil
}

func (i RLNWrapper) SetTree(treeHeight uint) bool {
	return i.ffi.SetTree(treeHeight)
}

func (i RLNWrapper) InitTreeWithLeaves(idcommitments []byte) bool {
	return i.ffi.InitTreeWithLeaves(idcommitments)
}

func (i RLNWrapper) KeyGen() []byte {
	return i.ffi.KeyGen()
}

func (i RLNWrapper) SeededKeyGen(seed []byte) []byte {
	return i.ffi.SeededKeyGen(seed)
}

func (i RLNWrapper) ExtendedKeyGen() []byte {
	return i.ffi.ExtendedKeyGen()
}

func (i RLNWrapper) ExtendedSeededKeyGen(seed []byte) []byte {
	return i.ffi.ExtendedSeededKeyGen(seed)
}

func (i RLNWrapper) Hash(input []byte) ([]byte, error) {
	return i.ffi.Hash(input)
}

func (i RLNWrapper) PoseidonHash(input []byte) ([]byte, error) {
	return i.ffi.PoseidonHash(input)
}

func (i RLNWrapper) SetLeaf(index uint, idcommitment []byte) bool {
	return i.ffi.SetLeaf(index, idcommitment)
}

func (i RLNWrapper) SetNextLeaf(idcommitment []byte) bool {
	return i.ffi.SetNextLeaf(idcommitment)
}

func (i RLNWrapper) SetLeavesFrom(index uint, idcommitments []byte) bool {
	return i.ffi.SetLeavesFrom(index, idcommitments)
}

func (i RLNWrapper) DeleteLeaf(index uint) bool {
	return i.ffi.DeleteLeaf(index)
}

func (i RLNWrapper) GetRoot() ([]byte, error) {
	return i.ffi.GetRoot()
}

func (i RLNWrapper) GetLeaf(index uint) ([]byte, error) {
	return i.ffi.GetLeaf(index)
}

func (i RLNWrapper) GenerateRLNProof(input []byte) ([]byte, error) {
	return i.ffi.GenerateRLNProof(input)
}

func (i RLNWrapper) VerifyWithRoots(input []byte, roots []byte) (bool, error) {
	return i.ffi.VerifyWithRoots(input, roots)
}

func (i RLNWrapper) AtomicOperation(index uint, leaves []byte, indices []byte) bool {
	return i.ffi.AtomicOperation(index, leaves, indices)
}

func (i RLNWrapper) SeqAtomicOperation(leaves []byte, indices []byte) bool {
	return i.ffi.SeqAtomicOperation(leaves, indices)
}

func (i RLNWrapper) RecoverIDSecret(proof1 []byte, proof2 []byte) ([]byte, error) {
	return i.ffi.RecoverIDSecret(proof1, proof2)
}

func (i RLNWrapper) SetMetadata(metadata []byte) bool {
	return i.ffi.SetMetadata(metadata)
}

func (i RLNWrapper) GetMetadata() ([]byte, error) {
	return i.ffi.GetMetadata()
}

func (i RLNWrapper) Flush() bool {
	return i.ffi.Flush()
}

func (i RLNWrapper) LeavesSet() uint {
	return i.ffi.LeavesSet()
}
