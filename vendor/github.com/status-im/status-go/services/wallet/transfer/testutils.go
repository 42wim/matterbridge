package transfer

import (
	"database/sql"
	"fmt"
	"math/big"
	"testing"

	eth_common "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/status-im/status-go/services/wallet/bigint"
	"github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/testutils"
	"github.com/status-im/status-go/services/wallet/token"

	"github.com/stretchr/testify/require"
)

type TestTransaction struct {
	Hash               eth_common.Hash
	ChainID            common.ChainID
	From               eth_common.Address // [sender]
	Timestamp          int64
	BlkNumber          int64
	Success            bool
	Nonce              uint64
	Contract           eth_common.Address
	MultiTransactionID common.MultiTransactionIDType
}

type TestTransfer struct {
	TestTransaction
	To    eth_common.Address // [address]
	Value int64
	Token *token.Token
}

type TestMultiTransaction struct {
	MultiTransactionID   common.MultiTransactionIDType
	MultiTransactionType MultiTransactionType
	FromAddress          eth_common.Address
	ToAddress            eth_common.Address
	FromToken            string
	ToToken              string
	FromAmount           int64
	ToAmount             int64
	Timestamp            int64
	FromNetworkID        *uint64
	ToNetworkID          *uint64
}

func SeedToToken(seed int) *token.Token {
	tokenIndex := seed % len(TestTokens)
	return TestTokens[tokenIndex]
}

func TestTrToToken(t *testing.T, tt *TestTransaction) (token *token.Token, isNative bool) {
	// Sanity check that none of the markers changed and they should be equal to seed
	require.Equal(t, tt.Timestamp, tt.BlkNumber)

	tokenIndex := int(tt.Timestamp) % len(TestTokens)
	isNative = testutils.SliceContains(NativeTokenIndices, tokenIndex)

	return TestTokens[tokenIndex], isNative
}

func generateTestTransaction(seed int) TestTransaction {
	token := SeedToToken(seed)
	return TestTransaction{
		Hash:      eth_common.HexToHash(fmt.Sprintf("0x1%d", seed)),
		ChainID:   common.ChainID(token.ChainID),
		From:      eth_common.HexToAddress(fmt.Sprintf("0x2%d", seed)),
		Timestamp: int64(seed),
		BlkNumber: int64(seed),
		Success:   true,
		Nonce:     uint64(seed),
		// In practice this is last20Bytes(Keccak256(RLP(From, nonce)))
		Contract:           eth_common.HexToAddress(fmt.Sprintf("0x4%d", seed)),
		MultiTransactionID: common.NoMultiTransactionID,
	}
}

func generateTestTransfer(seed int) TestTransfer {
	tokenIndex := seed % len(TestTokens)
	token := TestTokens[tokenIndex]
	return TestTransfer{
		TestTransaction: generateTestTransaction(seed),
		To:              eth_common.HexToAddress(fmt.Sprintf("0x3%d", seed)),
		Value:           int64(seed),
		Token:           token,
	}
}

func GenerateTestSendMultiTransaction(tr TestTransfer) TestMultiTransaction {
	return TestMultiTransaction{
		MultiTransactionType: MultiTransactionSend,
		FromAddress:          tr.From,
		ToAddress:            tr.To,
		FromToken:            tr.Token.Symbol,
		ToToken:              tr.Token.Symbol,
		FromAmount:           tr.Value,
		ToAmount:             0,
		Timestamp:            tr.Timestamp,
	}
}

func GenerateTestSwapMultiTransaction(tr TestTransfer, toToken string, toAmount int64) TestMultiTransaction {
	return TestMultiTransaction{
		MultiTransactionType: MultiTransactionSwap,
		FromAddress:          tr.From,
		ToAddress:            tr.To,
		FromToken:            tr.Token.Symbol,
		ToToken:              toToken,
		FromAmount:           tr.Value,
		ToAmount:             toAmount,
		Timestamp:            tr.Timestamp,
	}
}

func GenerateTestBridgeMultiTransaction(fromTr, toTr TestTransfer) TestMultiTransaction {
	return TestMultiTransaction{
		MultiTransactionType: MultiTransactionBridge,
		FromAddress:          fromTr.From,
		ToAddress:            toTr.To,
		FromToken:            fromTr.Token.Symbol,
		ToToken:              toTr.Token.Symbol,
		FromAmount:           fromTr.Value,
		ToAmount:             toTr.Value,
		Timestamp:            fromTr.Timestamp,
	}
}

// GenerateTestTransfers will generate transaction based on the TestTokens index and roll over if there are more than
// len(TestTokens) transactions
func GenerateTestTransfers(tb testing.TB, db *sql.DB, firstStartIndex int, count int) (result []TestTransfer, fromAddresses, toAddresses []eth_common.Address) {
	for i := firstStartIndex; i < (firstStartIndex + count); i++ {
		tr := generateTestTransfer(i)
		fromAddresses = append(fromAddresses, tr.From)
		toAddresses = append(toAddresses, tr.To)
		result = append(result, tr)
	}
	return
}

type TestCollectible struct {
	TokenAddress eth_common.Address
	TokenID      *big.Int
	ChainID      common.ChainID
}

var TestCollectibles = []TestCollectible{
	TestCollectible{
		TokenAddress: eth_common.HexToAddress("0x97a04fda4d97c6e3547d66b572e29f4a4ff40392"),
		TokenID:      big.NewInt(1),
		ChainID:      1,
	},
	TestCollectible{ // Same token ID as above but different address
		TokenAddress: eth_common.HexToAddress("0x2cec8879915cdbd80c88d8b1416aa9413a24ddfa"),
		TokenID:      big.NewInt(1),
		ChainID:      1,
	},
	TestCollectible{ // TokenID (big.Int) value 0 might be problematic if not handled properly
		TokenAddress: eth_common.HexToAddress("0x97a04fda4d97c6e3547d66b572e29f4a4ff4ABCD"),
		TokenID:      big.NewInt(0),
		ChainID:      420,
	},
	TestCollectible{
		TokenAddress: eth_common.HexToAddress("0x1dea7a3e04849840c0eb15fd26a55f6c40c4a69b"),
		TokenID:      big.NewInt(11),
		ChainID:      5,
	},
	TestCollectible{ // Same address as above but different token ID
		TokenAddress: eth_common.HexToAddress("0x1dea7a3e04849840c0eb15fd26a55f6c40c4a69b"),
		TokenID:      big.NewInt(12),
		ChainID:      5,
	},
}

var EthMainnet = token.Token{
	Address: eth_common.HexToAddress("0x"),
	Name:    "Ether",
	Symbol:  "ETH",
	ChainID: 1,
}

var EthGoerli = token.Token{
	Address: eth_common.HexToAddress("0x"),
	Name:    "Ether",
	Symbol:  "ETH",
	ChainID: 5,
}

var EthOptimism = token.Token{
	Address: eth_common.HexToAddress("0x"),
	Name:    "Ether",
	Symbol:  "ETH",
	ChainID: 10,
}

var UsdcMainnet = token.Token{
	Address: eth_common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"),
	Name:    "USD Coin",
	Symbol:  "USDC",
	ChainID: 1,
}

var UsdcGoerli = token.Token{
	Address: eth_common.HexToAddress("0x98339d8c260052b7ad81c28c16c0b98420f2b46a"),
	Name:    "USD Coin",
	Symbol:  "USDC",
	ChainID: 5,
}

var UsdcOptimism = token.Token{
	Address: eth_common.HexToAddress("0x7f5c764cbc14f9669b88837ca1490cca17c31607"),
	Name:    "USD Coin",
	Symbol:  "USDC",
	ChainID: 10,
}

var SntMainnet = token.Token{
	Address: eth_common.HexToAddress("0x744d70fdbe2ba4cf95131626614a1763df805b9e"),
	Name:    "Status Network Token",
	Symbol:  "SNT",
	ChainID: 1,
}

var DaiMainnet = token.Token{
	Address: eth_common.HexToAddress("0xf2edF1c091f683E3fb452497d9a98A49cBA84666"),
	Name:    "DAI Stablecoin",
	Symbol:  "DAI",
	ChainID: 5,
}

var DaiGoerli = token.Token{
	Address: eth_common.HexToAddress("0xf2edF1c091f683E3fb452497d9a98A49cBA84666"),
	Name:    "DAI Stablecoin",
	Symbol:  "DAI",
	ChainID: 5,
}

// TestTokens contains ETH/Mainnet, ETH/Goerli, ETH/Optimism, USDC/Mainnet, USDC/Goerli, USDC/Optimism, SNT/Mainnet, DAI/Mainnet, DAI/Goerli
var TestTokens = []*token.Token{
	&EthMainnet, &EthGoerli, &EthOptimism, &UsdcMainnet, &UsdcGoerli, &UsdcOptimism, &SntMainnet, &DaiMainnet, &DaiGoerli,
}

func LookupTokenIdentity(chainID uint64, address eth_common.Address, native bool) *token.Token {
	for _, token := range TestTokens {
		if token.ChainID == chainID && token.Address == address && token.IsNative() == native {
			return token
		}
	}
	return nil
}

var NativeTokenIndices = []int{0, 1, 2}

func InsertTestTransfer(tb testing.TB, db *sql.DB, address eth_common.Address, tr *TestTransfer) {
	token := TestTokens[int(tr.Timestamp)%len(TestTokens)]
	InsertTestTransferWithOptions(tb, db, address, tr, &TestTransferOptions{
		TokenAddress: token.Address,
	})
}

type TestTransferOptions struct {
	TokenAddress     eth_common.Address
	TokenID          *big.Int
	NullifyAddresses []eth_common.Address
	Tx               *types.Transaction
	Receipt          *types.Receipt
}

func GenerateTxField(data []byte) *types.Transaction {
	return types.NewTx(&types.DynamicFeeTx{
		Data: data,
	})
}

func InsertTestTransferWithOptions(tb testing.TB, db *sql.DB, address eth_common.Address, tr *TestTransfer, opt *TestTransferOptions) {
	var (
		tx *sql.Tx
	)
	tx, err := db.Begin()
	require.NoError(tb, err)
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	blkHash := eth_common.HexToHash("4")

	block := blockDBFields{
		chainID:     uint64(tr.ChainID),
		account:     address,
		blockNumber: big.NewInt(tr.BlkNumber),
		blockHash:   blkHash,
	}

	// Respect `FOREIGN KEY(network_id,address,blk_hash)` of `transfers` table
	err = insertBlockDBFields(tx, block)
	require.NoError(tb, err)

	receiptStatus := uint64(0)
	if tr.Success {
		receiptStatus = 1
	}

	tokenType := "eth"
	if (opt.TokenAddress != eth_common.Address{}) {
		if opt.TokenID == nil {
			tokenType = "erc20"
		} else {
			tokenType = "erc721"
		}
	}

	// Workaround to simulate writing of NULL values for addresses
	txTo := &tr.To
	txFrom := &tr.From
	for i := 0; i < len(opt.NullifyAddresses); i++ {
		if opt.NullifyAddresses[i] == tr.To {
			txTo = nil
		}
		if opt.NullifyAddresses[i] == tr.From {
			txFrom = nil
		}
	}

	transfer := transferDBFields{
		chainID:            uint64(tr.ChainID),
		id:                 tr.Hash,
		txHash:             &tr.Hash,
		address:            address,
		blockHash:          blkHash,
		blockNumber:        big.NewInt(tr.BlkNumber),
		sender:             tr.From,
		transferType:       common.Type(tokenType),
		timestamp:          uint64(tr.Timestamp),
		multiTransactionID: tr.MultiTransactionID,
		baseGasFees:        "0x0",
		receiptStatus:      &receiptStatus,
		txValue:            big.NewInt(tr.Value),
		txFrom:             txFrom,
		txTo:               txTo,
		txNonce:            &tr.Nonce,
		tokenAddress:       &opt.TokenAddress,
		contractAddress:    &tr.Contract,
		tokenID:            opt.TokenID,
		transaction:        opt.Tx,
		receipt:            opt.Receipt,
	}
	err = updateOrInsertTransfersDBFields(tx, []transferDBFields{transfer})
	require.NoError(tb, err)
}

func InsertTestPendingTransaction(tb testing.TB, db *sql.DB, tr *TestTransfer) {
	_, err := db.Exec(`
		INSERT INTO pending_transactions (network_id, hash, timestamp, from_address, to_address,
			symbol, gas_price, gas_limit, value, data, type, additional_data, multi_transaction_id
		) VALUES (?, ?, ?, ?, ?, 'ETH', 0, 0, ?, '', 'eth', '', ?)`,
		tr.ChainID, tr.Hash, tr.Timestamp, tr.From, tr.To, (*bigint.SQLBigIntBytes)(big.NewInt(tr.Value)), tr.MultiTransactionID)
	require.NoError(tb, err)
}

func InsertTestMultiTransaction(tb testing.TB, db *sql.DB, tr *TestMultiTransaction) common.MultiTransactionIDType {
	fromTokenType := tr.FromToken
	if tr.FromToken == "" {
		fromTokenType = testutils.EthSymbol
	}
	toTokenType := tr.ToToken
	if tr.ToToken == "" {
		toTokenType = testutils.EthSymbol
	}
	fromAmount := (*hexutil.Big)(big.NewInt(tr.FromAmount))
	toAmount := (*hexutil.Big)(big.NewInt(tr.ToAmount))

	result, err := db.Exec(`
		INSERT INTO multi_transactions (from_address, from_asset, from_amount, to_address, to_asset, to_amount, type, timestamp, from_network_id, to_network_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tr.FromAddress, fromTokenType, fromAmount.String(), tr.ToAddress, toTokenType, toAmount.String(), tr.MultiTransactionType, tr.Timestamp, tr.FromNetworkID, tr.ToNetworkID)
	require.NoError(tb, err)
	rowID, err := result.LastInsertId()
	require.NoError(tb, err)
	tr.MultiTransactionID = common.MultiTransactionIDType(rowID)
	return tr.MultiTransactionID
}

// For using in tests only outside the package
func SaveTransfersMarkBlocksLoaded(database *Database, chainID uint64, address eth_common.Address, transfers []Transfer, blocks []*big.Int) error {
	return saveTransfersMarkBlocksLoaded(database.client, chainID, address, transfers, blocks)
}
