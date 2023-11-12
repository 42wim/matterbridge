package ens

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func nameHash(name string) common.Hash {
	node := common.Hash{}

	if len(name) > 0 {
		labels := strings.Split(name, ".")

		for i := len(labels) - 1; i >= 0; i-- {
			labelSha := crypto.Keccak256Hash([]byte(labels[i]))
			node = crypto.Keccak256Hash(node.Bytes(), labelSha.Bytes())
		}
	}

	return node
}

func validateENSUsername(username string) error {
	if !strings.HasSuffix(username, ".eth") {
		return fmt.Errorf("username must end with .eth")
	}

	return nil
}

func usernameToLabel(username string) [32]byte {
	usernameHashed := crypto.Keccak256([]byte(username))
	var label [32]byte
	copy(label[:], usernameHashed)

	return label
}

func extractCoordinates(pubkey string) ([32]byte, [32]byte) {
	x, _ := hex.DecodeString(pubkey[4:68])
	y, _ := hex.DecodeString(pubkey[68:132])

	var xByte [32]byte
	copy(xByte[:], x)

	var yByte [32]byte
	copy(yByte[:], y)

	return xByte, yByte
}
