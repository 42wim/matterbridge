// Imported from github.com/ethereum/go-ethereum/accounts/keystore/keystore.go

package keystore

import (
	"errors"
)

const (
	version = 3
)

var (
	ErrDecrypt = errors.New("could not decrypt key with given password")
)
