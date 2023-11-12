package utils

import (
	"crypto/ecdsa"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
)

// EcdsaPubKeyToSecp256k1PublicKey converts an `ecdsa.PublicKey` into a libp2p `crypto.Secp256k1PublicKey“
func EcdsaPubKeyToSecp256k1PublicKey(pubKey *ecdsa.PublicKey) *crypto.Secp256k1PublicKey {
	xFieldVal := &btcec.FieldVal{}
	yFieldVal := &btcec.FieldVal{}
	xFieldVal.SetByteSlice(pubKey.X.Bytes())
	yFieldVal.SetByteSlice(pubKey.Y.Bytes())
	return (*crypto.Secp256k1PublicKey)(btcec.NewPublicKey(xFieldVal, yFieldVal))
}

// EcdsaPrivKeyToSecp256k1PrivKey converts an `ecdsa.PrivateKey` into a libp2p `crypto.Secp256k1PrivateKey“
func EcdsaPrivKeyToSecp256k1PrivKey(privKey *ecdsa.PrivateKey) *crypto.Secp256k1PrivateKey {
	privK, _ := btcec.PrivKeyFromBytes(privKey.D.Bytes())
	return (*crypto.Secp256k1PrivateKey)(privK)
}
