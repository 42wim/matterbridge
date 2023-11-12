package web3provider

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	signercore "github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/services/rpcfilters"
	"github.com/status-im/status-go/services/typeddata"
	"github.com/status-im/status-go/transactions"
)

// signMessage checks the pwd vs the selected account and signs a message
func (api *API) signMessage(data interface{}, address string, password string) (types.HexBytes, error) {
	account, err := api.getVerifiedWalletAccount(address, password)
	if err != nil {
		return types.HexBytes{}, err
	}

	var dBytes []byte
	switch d := data.(type) {
	case string:
		dBytes = []byte(d)
	case []byte:
		dBytes = d
	case byte:
		dBytes = []byte{d}
	}

	hash := crypto.TextHash(dBytes)

	sig, err := crypto.Sign(hash, account.AccountKey.PrivateKey)
	if err != nil {
		return types.HexBytes{}, err
	}

	sig[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper

	return types.HexBytes(sig), err
}

// signTypedData accepts data and password. Gets verified account and signs typed data.
func (api *API) signTypedData(typed typeddata.TypedData, address string, password string) (types.HexBytes, error) {
	account, err := api.getVerifiedWalletAccount(address, password)
	if err != nil {
		return types.HexBytes{}, err
	}
	chain := new(big.Int).SetUint64(api.s.config.NetworkID)
	sig, err := typeddata.Sign(typed, account.AccountKey.PrivateKey, chain)
	if err != nil {
		return types.HexBytes{}, err
	}
	return types.HexBytes(sig), err
}

// signTypedDataV4 accepts data and password. Gets verified account and signs typed data.
func (api *API) signTypedDataV4(typed signercore.TypedData, address string, password string) (types.HexBytes, error) {
	account, err := api.getVerifiedWalletAccount(address, password)
	if err != nil {
		return types.HexBytes{}, err
	}
	chain := new(big.Int).SetUint64(api.s.config.NetworkID)
	sig, err := typeddata.SignTypedDataV4(typed, account.AccountKey.PrivateKey, chain)
	if err != nil {
		return types.HexBytes{}, err
	}
	return types.HexBytes(sig), err
}

// SendTransaction creates a new transaction and waits until it's complete.
func (api *API) sendTransaction(chainID uint64, sendArgs transactions.SendTxArgs, password string, requestType string) (hash types.Hash, err error) {
	verifiedAccount, err := api.getVerifiedWalletAccount(sendArgs.From.String(), password)
	if err != nil {
		return hash, err
	}

	hash, err = api.s.transactor.SendTransactionWithChainID(chainID, sendArgs, verifiedAccount)
	if err != nil {
		return
	}

	go api.s.rpcFiltersSrvc.TriggerTransactionSentToUpstreamEvent(&rpcfilters.PendingTxInfo{
		Hash:    common.Hash(hash),
		Type:    requestType,
		From:    common.Address(sendArgs.From),
		ChainID: chainID,
	})

	return
}

func (api *API) EcRecover(data hexutil.Bytes, sig hexutil.Bytes) (types.Address, error) {
	if len(sig) != 65 {
		return types.Address{}, fmt.Errorf("signature must be 65 bytes long")
	}
	if sig[64] != 27 && sig[64] != 28 {
		return types.Address{}, fmt.Errorf("invalid Ethereum signature (V is not 27 or 28)")
	}
	sig[64] -= 27 // Transform yellow paper V from 27/28 to 0/1
	hash := crypto.TextHash(data)
	rpk, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return types.Address{}, err
	}
	return crypto.PubkeyToAddress(*rpk), nil
}
