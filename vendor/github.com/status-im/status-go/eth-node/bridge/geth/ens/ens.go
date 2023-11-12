package ens

import (
	"bytes"
	"context"
	"crypto/elliptic"
	"encoding/hex"
	"math/big"
	"time"

	ens "github.com/wealdtech/go-ens/v3"
	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/status-im/status-go/eth-node/crypto"
	enstypes "github.com/status-im/status-go/eth-node/types/ens"
)

const (
	contractQueryTimeout = 5000 * time.Millisecond
)

type Verifier struct {
	logger *zap.Logger
}

// NewVerifier returns a Verifier attached to the specified logger
func NewVerifier(logger *zap.Logger) *Verifier {
	return &Verifier{logger: logger}
}

func (m *Verifier) ReverseResolve(address common.Address, rpcEndpoint string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), contractQueryTimeout)
	defer cancel()

	ethClient, err := ethclient.DialContext(ctx, rpcEndpoint)
	if err != nil {
		return "", err
	}
	return ens.ReverseResolve(ethClient, address)
}

func (m *Verifier) verifyENSName(ensInfo enstypes.ENSDetails, ethclient *ethclient.Client) enstypes.ENSResponse {
	publicKeyStr := ensInfo.PublicKeyString
	ensName := ensInfo.Name
	m.logger.Info("Resolving ENS name", zap.String("name", ensName), zap.String("publicKey", publicKeyStr))
	response := enstypes.ENSResponse{
		Name:            ensName,
		PublicKeyString: publicKeyStr,
		VerifiedAt:      time.Now().Unix(),
	}

	expectedPubKeyBytes, err := hex.DecodeString(publicKeyStr)
	if err != nil {
		response.Error = err
		return response
	}

	publicKey, err := crypto.UnmarshalPubkey(expectedPubKeyBytes)
	if err != nil {
		response.Error = err
		return response
	}

	// Resolve ensName
	resolver, err := ens.NewResolver(ethclient, ensName)
	if err != nil {
		m.logger.Error("error while creating ENS name resolver", zap.String("ensName", ensName), zap.Error(err))
		response.Error = err
		return response
	}
	x, y, err := resolver.PubKey()
	if err != nil {
		m.logger.Error("error while resolving public key from ENS name", zap.String("ensName", ensName), zap.Error(err))
		response.Error = err
		return response
	}

	// Assemble the bytes returned for the pubkey
	pubKeyBytes := elliptic.Marshal(crypto.S256(), new(big.Int).SetBytes(x[:]), new(big.Int).SetBytes(y[:]))

	response.PublicKey = publicKey
	response.Verified = bytes.Equal(pubKeyBytes, expectedPubKeyBytes)
	return response
}

// CheckBatch verifies that a registered ENS name matches the expected public key
func (m *Verifier) CheckBatch(ensDetails []enstypes.ENSDetails, rpcEndpoint, contractAddress string) (map[string]enstypes.ENSResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), contractQueryTimeout)
	defer cancel()

	ch := make(chan enstypes.ENSResponse)
	response := make(map[string]enstypes.ENSResponse)

	ethclient, err := ethclient.DialContext(ctx, rpcEndpoint)
	if err != nil {
		return nil, err
	}

	for _, ensInfo := range ensDetails {
		go func(info enstypes.ENSDetails) { ch <- m.verifyENSName(info, ethclient) }(ensInfo)
	}

	for range ensDetails {
		r := <-ch
		response[r.PublicKeyString] = r
	}
	close(ch)

	return response, nil
}
