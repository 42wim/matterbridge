package dynamic

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"sort"
	"sync/atomic"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ErrCount struct {
	err   error
	count int
}

type MockClient struct {
	ethclient.Client
	blockChain     MockBlockChain
	latestBlockNum atomic.Int64
	errOnBlock     map[int64]*ErrCount
}

func (c *MockClient) SetLatestBlockNumber(num int64) {
	c.latestBlockNum.Store(num)
}

func (c *MockClient) Close() {

}
func (c *MockClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return types.NewBlock(&types.Header{Number: big.NewInt(c.latestBlockNum.Load())}, nil, nil, nil, nil), nil
}
func NewMockClient(t *testing.T, blockFile string) *MockClient {
	blockChain := MockBlockChain{}
	data, err := os.ReadFile(blockFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &blockChain); err != nil {
		t.Fatal(err)
	}
	return &MockClient{blockChain: blockChain, errOnBlock: map[int64]*ErrCount{}}
}

func (c *MockClient) SetErrorOnBlock(blockNum int64, err error, count int) {
	c.errOnBlock[blockNum] = &ErrCount{err: err, count: count}
}

func (c *MockClient) getFromAndToRange(query ethereum.FilterQuery) (int64, int64) {
	var fromBlock int64
	if query.FromBlock == nil {
		fromBlock = 0
	} else {
		fromBlock = query.FromBlock.Int64()
	}

	var toBlock int64
	if query.ToBlock == nil {
		toBlock = 0
	} else {
		toBlock = query.ToBlock.Int64()
	}
	return fromBlock, toBlock
}
func (c *MockClient) FilterLogs(ctx context.Context, query ethereum.FilterQuery) (allTxLogs []types.Log, err error) {
	fromBlock, toBlock := c.getFromAndToRange(query)
	for block, details := range c.blockChain.Blocks {
		if block >= fromBlock && block <= toBlock {
			if txLogs := details.getLogs(uint64(block), query.Addresses, query.Topics[0]); len(txLogs) != 0 {
				allTxLogs = append(allTxLogs, txLogs...)
			}
			if errCount, ok := c.errOnBlock[block]; ok && errCount.count != 0 {
				errCount.count--
				return nil, errCount.err
			}
		}
	}
	sort.Slice(allTxLogs, func(i, j int) bool {
		return allTxLogs[i].BlockNumber < allTxLogs[j].BlockNumber ||
			(allTxLogs[i].BlockNumber == allTxLogs[j].BlockNumber && allTxLogs[i].Index < allTxLogs[j].Index)
	})
	return allTxLogs, nil
}

func (c *MockClient) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	for {
		next := c.latestBlockNum.Load() + 1
		if c.blockChain.Blocks[next] != nil {
			ch <- &types.Header{Number: big.NewInt(next)}
			c.latestBlockNum.Store(next)
		} else {
			break
		}
	}
	return testNoopSub{}, nil
}

type testNoopSub struct {
}

func (testNoopSub) Unsubscribe() {

}

// Err returns the subscription error channel. The error channel receives
// a value if there is an issue with the subscription (e.g. the network connection
// delivering the events has been closed). Only one value will ever be sent.
// The error channel is closed by Unsubscribe.
func (testNoopSub) Err() <-chan error {
	ch := make(chan error)
	return ch
}
