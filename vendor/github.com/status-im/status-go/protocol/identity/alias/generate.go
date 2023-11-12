package alias

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	"github.com/status-im/status-go/eth-node/crypto"
)

const poly uint64 = 0xB8

func generate(seed uint64) string {
	generator := newLSFR(poly, seed)
	adjective1Index := generator.next() % uint64(len(adjectives))
	adjective2Index := generator.next() % uint64(len(adjectives))
	animalIndex := generator.next() % uint64(len(animals))
	adjective1 := adjectives[adjective1Index]
	adjective2 := adjectives[adjective2Index]
	animal := animals[animalIndex]

	return fmt.Sprintf("%s %s %s", adjective1, adjective2, animal)
}

// GenerateFromPublicKey returns the 3 words name given an *ecdsa.PublicKey
func GenerateFromPublicKey(publicKey *ecdsa.PublicKey) string {
	// Here we truncate the public key to the least significant 64 bits
	return generate(uint64(publicKey.X.Int64()))
}

// GenerateFromPublicKeyString returns the 3 words name given a public key
// prefixed with 0x
func GenerateFromPublicKeyString(publicKeyString string) (string, error) {
	publicKeyBytes, err := hex.DecodeString(publicKeyString[2:])
	if err != nil {
		return "", err
	}

	publicKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	if err != nil {
		return "", err
	}

	return GenerateFromPublicKey(publicKey), nil
}
