package types

import (
	"math/big"

	"github.com/status-im/status-go/eth-node/types"
)

type TransactionStatus uint64

const (
	TransactionStatusFailed  = 0
	TransactionStatusSuccess = 1
	TransactionStatusPending = 2
)

type Message struct {
	to         *types.Address
	from       types.Address
	nonce      uint64
	amount     *big.Int
	gasLimit   uint64
	gasPrice   *big.Int
	data       []byte
	checkNonce bool
}

func NewMessage(from types.Address, to *types.Address, nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, checkNonce bool) Message {
	return Message{
		from:       from,
		to:         to,
		nonce:      nonce,
		amount:     amount,
		gasLimit:   gasLimit,
		gasPrice:   gasPrice,
		data:       data,
		checkNonce: checkNonce,
	}
}

func (m Message) From() types.Address { return m.from }
func (m Message) To() *types.Address  { return m.to }
func (m Message) GasPrice() *big.Int  { return m.gasPrice }
func (m Message) Value() *big.Int     { return m.amount }
func (m Message) Gas() uint64         { return m.gasLimit }
func (m Message) Nonce() uint64       { return m.nonce }
func (m Message) Data() []byte        { return m.data }
func (m Message) CheckNonce() bool    { return m.checkNonce }
