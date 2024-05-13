package transactions

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/status-go/rpc/chain"
	"github.com/status-im/status-go/services/wallet/bigint"
	"github.com/status-im/status-go/services/wallet/common"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockETHClient struct {
	mock.Mock
}

func (m *MockETHClient) BatchCallContext(ctx context.Context, b []rpc.BatchElem) error {
	args := m.Called(ctx, b)
	return args.Error(0)
}

type MockChainClient struct {
	mock.Mock

	Clients map[common.ChainID]*MockETHClient
}

func NewMockChainClient() *MockChainClient {
	return &MockChainClient{
		Clients: make(map[common.ChainID]*MockETHClient),
	}
}

func (m *MockChainClient) SetAvailableClients(chainIDs []common.ChainID) *MockChainClient {
	for _, chainID := range chainIDs {
		if _, ok := m.Clients[chainID]; !ok {
			m.Clients[chainID] = new(MockETHClient)
		}
	}
	return m
}

func (m *MockChainClient) AbstractEthClient(chainID common.ChainID) (chain.BatchCallClient, error) {
	if _, ok := m.Clients[chainID]; !ok {
		panic(fmt.Sprintf("no mock client for chainID %d", chainID))
	}
	return m.Clients[chainID], nil
}

func GenerateTestPendingTransactions(start int, count int) []PendingTransaction {
	if count > 127 {
		panic("can't generate more than 127 distinct transactions")
	}

	txs := make([]PendingTransaction, count)
	for i := start; i < count; i++ {
		txs[i] = PendingTransaction{
			Hash:           eth.HexToHash(fmt.Sprintf("0x1%d", i)),
			From:           eth.HexToAddress(fmt.Sprintf("0x2%d", i)),
			To:             eth.HexToAddress(fmt.Sprintf("0x3%d", i)),
			Type:           RegisterENS,
			AdditionalData: "someuser.stateofus.eth",
			Value:          bigint.BigInt{Int: big.NewInt(int64(i))},
			GasLimit:       bigint.BigInt{Int: big.NewInt(21000)},
			GasPrice:       bigint.BigInt{Int: big.NewInt(int64(i))},
			ChainID:        777,
			Status:         new(TxStatus),
			AutoDelete:     new(bool),
			Symbol:         "ETH",
			Timestamp:      uint64(i),
		}
		*txs[i].Status = Pending  // set to pending by default
		*txs[i].AutoDelete = true // set to true by default
	}
	return txs
}

// groupSliceInMap groups a slice of S into a map[K][]N using the getKeyValue function to extract the key and new value for each entry
func groupSliceInMap[S any, K comparable, N any](s []S, getKeyValue func(entry S, i int) (K, N)) map[K][]N {
	m := make(map[K][]N)
	for i, x := range s {
		k, v := getKeyValue(x, i)
		m[k] = append(m[k], v)
	}
	return m
}

func keysInMap[K comparable, V any](m map[K]V) (res []K) {
	if len(m) > 0 {
		res = make([]K, 0, len(m))
	}

	for k := range m {
		res = append(res, k)
	}
	return
}

type TestTxSummary struct {
	failStatus  bool
	DontConfirm bool
	// Timestamp will be used to mock the Timestamp if greater than 0
	Timestamp int
}

type summaryTxPair struct {
	summary  TestTxSummary
	tx       PendingTransaction
	answered bool
}

func MockTestTransactions(t *testing.T, chainClient *MockChainClient, testTxs []TestTxSummary) []PendingTransaction {
	genTxs := GenerateTestPendingTransactions(0, len(testTxs))
	for i, tx := range testTxs {
		if tx.Timestamp > 0 {
			genTxs[i].Timestamp = uint64(tx.Timestamp)
		}
	}

	grouped := groupSliceInMap(genTxs, func(tx PendingTransaction, i int) (common.ChainID, summaryTxPair) {
		return tx.ChainID, summaryTxPair{
			summary: testTxs[i],
			tx:      tx,
		}
	})

	chains := keysInMap(grouped)
	chainClient.SetAvailableClients(chains)

	for chainID, chainSummaries := range grouped {
		// Mock the one call to getTransactionReceipt
		// It is expected that pending transactions manager will call out of order, therefore match based on hash
		cl := chainClient.Clients[chainID]
		call := cl.On("BatchCallContext", mock.Anything, mock.MatchedBy(func(b []rpc.BatchElem) bool {
			if len(b) > len(chainSummaries) {
				return false
			}
			for i := range b {
				for _, sum := range chainSummaries {
					tx := &sum.tx
					if sum.answered {
						continue
					}
					require.Equal(t, GetTransactionReceiptRPCName, b[i].Method)
					if tx.Hash == b[i].Args[0].(eth.Hash) {
						sum.answered = true
						return true
					}
				}
			}
			return false
		})).Return(nil)

		call.Run(func(args mock.Arguments) {
			elems := args.Get(1).([]rpc.BatchElem)
			for i := range elems {
				receiptWrapper, ok := elems[i].Result.(*nullableReceipt)
				require.True(t, ok)
				require.NotNil(t, receiptWrapper)
				// Simulate parsing of eth_getTransactionReceipt response
				for _, sum := range chainSummaries {
					tx := &sum.tx
					if tx.Hash == elems[i].Args[0].(eth.Hash) {
						if !sum.summary.DontConfirm {
							status := types.ReceiptStatusSuccessful
							if sum.summary.failStatus {
								status = types.ReceiptStatusFailed
							}

							receiptWrapper.Receipt = &types.Receipt{
								BlockNumber: new(big.Int).SetUint64(1),
								Status:      status,
							}
						}
					}
				}
			}
		})
	}
	return genTxs
}
