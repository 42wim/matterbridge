package types

import (
	"math/big"
	"github.com/ethereum/go-ethereum/common"
)

type DummyBlobTx struct {
}

// accessors for innerTx.
func (tx *DummyBlobTx) txType() byte           { return BlobTxType }
func (tx *DummyBlobTx) chainID() *big.Int      { return nil }
func (tx *DummyBlobTx) accessList() AccessList { return AccessList{} }
func (tx *DummyBlobTx) data() []byte           { return nil }
func (tx *DummyBlobTx) gas() uint64            { return 0 }
func (tx *DummyBlobTx) gasFeeCap() *big.Int    { return nil }
func (tx *DummyBlobTx) gasTipCap() *big.Int    { return nil }
func (tx *DummyBlobTx) gasPrice() *big.Int     { return nil }
func (tx *DummyBlobTx) value() *big.Int        { return nil }
func (tx *DummyBlobTx) nonce() uint64          { return 0 }
func (tx *DummyBlobTx) to() *common.Address    { return nil }
func (tx *DummyBlobTx) isSystemTx() bool       { return false }
func (tx *DummyBlobTx) copy() TxData 		   { return nil }
func (tx *DummyBlobTx) effectiveGasPrice(dst *big.Int, baseFee *big.Int) *big.Int { return nil }
func (tx *DummyBlobTx) isFake() bool           { return true }
func (tx *DummyBlobTx) rawSignatureValues() (v, r, s *big.Int) { return nil, nil, nil }
func (tx *DummyBlobTx) setSignatureValues(chainID, v, r, s *big.Int) {}
