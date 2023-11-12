package dynamic

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// MockBlockChain  is currently a chain of events for different block numbers
// it is used internal by mock client for returning events for a given block number or range in FilterLog rpc call.
type MockBlockChain struct {
	Blocks map[int64]*MockBlock `json:"blocks"`
}

type MockBlock []MockEvent

func containsEntry[T common.Hash | common.Address](topics []T, topicA T) bool {
	for _, topic := range topics {
		if topic == topicA {
			return true
		}
	}
	return false
}

func Topic(topic string) common.Hash {
	return crypto.Keccak256Hash([]byte(topic))
}
func (b MockBlock) getLogs(blockNum uint64, addrs []common.Address, topicA []common.Hash) (txLogs []types.Log) {
	for ind, event := range b {
		txLog := event.GetLog()
		if containsEntry(addrs, txLog.Address) && (len(topicA) == 0 || containsEntry(topicA, txLog.Topics[0])) {
			txLog.BlockNumber = blockNum
			txLog.Index = uint(ind)
			txLogs = append(txLogs, txLog)
		}
	}
	return
}

type MockEvent struct {
	Address common.Address `json:"address"`
	Topics  []string       `json:"topics"`
	Txhash  common.Hash    `json:"txhash"`
	Data    []string       `json:"data"`
}

func (e MockEvent) GetLog() types.Log {
	topics := []common.Hash{Topic(e.Topics[0])}
	for _, topic := range e.Topics[1:] {
		topics = append(topics, parseData(topic))
	}
	//
	var data []byte
	for _, entry := range e.Data {
		data = append(data, parseData(entry).Bytes()...)
	}
	return types.Log{
		Address: e.Address,
		Topics:  topics,
		TxHash:  e.Txhash,
		Data:    data,
	}
}

func parseData(data string) common.Hash {
	splits := strings.Split(data, ":")
	switch splits[0] {
	case "bigint":
		bigInt, ok := new(big.Int).SetString(splits[1], 10)
		if !ok {
			panic("invalid big int")
		}
		return common.BytesToHash(bigInt.Bytes())
	default:
		panic("invalid data type")
	}
}
