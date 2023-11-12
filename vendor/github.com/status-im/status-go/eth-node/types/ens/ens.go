package enstypes

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
)

type ENSVerifier interface {
	// CheckBatch verifies that a registered ENS name matches the expected public key
	CheckBatch(ensDetails []ENSDetails, rpcEndpoint, contractAddress string) (map[string]ENSResponse, error)
	ReverseResolve(address common.Address, rpcEndpoint string) (string, error)
}

type ENSDetails struct {
	Name            string `json:"name"`
	PublicKeyString string `json:"publicKey"`
}

type ENSResponse struct {
	Name            string           `json:"name"`
	Verified        bool             `json:"verified"`
	VerifiedAt      int64            `json:"verifiedAt"`
	Error           error            `json:"error"`
	PublicKey       *ecdsa.PublicKey `json:"-"`
	PublicKeyString string           `json:"publicKey"`
}
