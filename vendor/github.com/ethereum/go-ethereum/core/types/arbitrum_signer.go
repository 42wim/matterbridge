package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

var ArbosAddress = common.HexToAddress("0xa4b05")
var ArbSysAddress = common.HexToAddress("0x64")
var ArbGasInfoAddress = common.HexToAddress("0x6c")
var ArbRetryableTxAddress = common.HexToAddress("0x6e")
var NodeInterfaceAddress = common.HexToAddress("0xc8")
var NodeInterfaceDebugAddress = common.HexToAddress("0xc9")

type arbitrumSigner struct{ Signer }

func NewArbitrumSigner(signer Signer) Signer {
	return arbitrumSigner{Signer: signer}
}

func (s arbitrumSigner) Sender(tx *Transaction) (common.Address, error) {
	switch inner := tx.inner.(type) {
	case *ArbitrumUnsignedTx:
		return inner.From, nil
	case *ArbitrumContractTx:
		return inner.From, nil
	case *ArbitrumDepositTx:
		return inner.From, nil
	case *ArbitrumInternalTx:
		return ArbosAddress, nil
	case *ArbitrumRetryTx:
		return inner.From, nil
	case *ArbitrumSubmitRetryableTx:
		return inner.From, nil
	case *ArbitrumLegacyTxData:
		legacyData := tx.inner.(*ArbitrumLegacyTxData)
		if legacyData.Sender != nil {
			return *legacyData.Sender, nil
		}
		fakeTx := NewTx(&legacyData.LegacyTx)
		return s.Signer.Sender(fakeTx)
	default:
		return s.Signer.Sender(tx)
	}
}

func (s arbitrumSigner) Equal(s2 Signer) bool {
	x, ok := s2.(arbitrumSigner)
	return ok && x.Signer.Equal(s.Signer)
}

func (s arbitrumSigner) SignatureValues(tx *Transaction, sig []byte) (R, S, V *big.Int, err error) {
	switch tx.inner.(type) {
	case *ArbitrumUnsignedTx:
		return bigZero, bigZero, bigZero, nil
	case *ArbitrumContractTx:
		return bigZero, bigZero, bigZero, nil
	case *ArbitrumDepositTx:
		return bigZero, bigZero, bigZero, nil
	case *ArbitrumInternalTx:
		return bigZero, bigZero, bigZero, nil
	case *ArbitrumRetryTx:
		return bigZero, bigZero, bigZero, nil
	case *ArbitrumSubmitRetryableTx:
		return bigZero, bigZero, bigZero, nil
	case *ArbitrumLegacyTxData:
		legacyData := tx.inner.(*ArbitrumLegacyTxData)
		fakeTx := NewTx(&legacyData.LegacyTx)
		return s.Signer.SignatureValues(fakeTx, sig)
	default:
		return s.Signer.SignatureValues(tx, sig)
	}
}

// Hash returns the hash to be signed by the sender.
// It does not uniquely identify the transaction.
func (s arbitrumSigner) Hash(tx *Transaction) common.Hash {
	if legacyData, isArbLegacy := tx.inner.(*ArbitrumLegacyTxData); isArbLegacy {
		fakeTx := NewTx(&legacyData.LegacyTx)
		return s.Signer.Hash(fakeTx)
	}
	return s.Signer.Hash(tx)
}
