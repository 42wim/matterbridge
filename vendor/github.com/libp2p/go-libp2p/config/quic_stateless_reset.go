package config

import (
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/hkdf"

	"github.com/libp2p/go-libp2p/core/crypto"

	"github.com/quic-go/quic-go"
)

const statelessResetKeyInfo = "libp2p quic stateless reset key"

func PrivKeyToStatelessResetKey(key crypto.PrivKey) (quic.StatelessResetKey, error) {
	var statelessResetKey quic.StatelessResetKey
	keyBytes, err := key.Raw()
	if err != nil {
		return statelessResetKey, err
	}
	keyReader := hkdf.New(sha256.New, keyBytes, nil, []byte(statelessResetKeyInfo))
	if _, err := io.ReadFull(keyReader, statelessResetKey[:]); err != nil {
		return statelessResetKey, err
	}
	return statelessResetKey, nil
}
