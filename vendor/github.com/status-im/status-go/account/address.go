package account

import (
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
)

func CreateAddress() (address, pubKey, privKey string, err error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return "", "", "", err
	}

	privKeyBytes := crypto.FromECDSA(key)
	pubKeyBytes := crypto.FromECDSAPub(&key.PublicKey)
	addressBytes := crypto.PubkeyToAddress(key.PublicKey)

	privKey = types.EncodeHex(privKeyBytes)
	pubKey = types.EncodeHex(pubKeyBytes)
	address = addressBytes.Hex()

	return
}
