package bep44

import (
	"crypto/sha1"
)

type Target = [sha1.Size]byte

func MakeMutableTarget(pubKey [32]byte, salt []byte) Target {
	return sha1.Sum(append(pubKey[:], salt...))
}
