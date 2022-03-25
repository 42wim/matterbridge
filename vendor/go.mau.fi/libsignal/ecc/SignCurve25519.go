package ecc

// Package curve25519sign implements a signature scheme based on Curve25519 keys.
// See https://moderncrypto.org/mail-archive/curves/2014/000205.html for details.

import (
	"crypto/ed25519"
	"crypto/sha512"

	"filippo.io/edwards25519"
	"filippo.io/edwards25519/field"
)

// sign signs the message with privateKey and returns a signature as a byte slice.
func sign(privateKey *[32]byte, message []byte, random [64]byte) *[64]byte {

	// Calculate Ed25519 public key from Curve25519 private key
	var A edwards25519.Point
	privateKeyScalar, _ := edwards25519.NewScalar().SetBytesWithClamping(privateKey[:])
	A.ScalarBaseMult(privateKeyScalar)
	publicKey := *(*[32]byte)(A.Bytes())

	// Calculate r
	diversifier := [32]byte{
		0xFE, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

	var r [64]byte
	hash := sha512.New()
	hash.Write(diversifier[:])
	hash.Write(privateKey[:])
	hash.Write(message)
	hash.Write(random[:])
	hash.Sum(r[:0])

	// Calculate R
	var rReduced *edwards25519.Scalar
	rReduced, _ = edwards25519.NewScalar().SetUniformBytes(r[:])
	var R edwards25519.Point
	R.ScalarBaseMult(rReduced)

	var encodedR [32]byte
	encodedR = *(*[32]byte)(R.Bytes())

	// Calculate S = r + SHA2-512(R || A_ed || msg) * a  (mod L)
	var hramDigest [64]byte
	hash.Reset()
	hash.Write(encodedR[:])
	hash.Write(publicKey[:])
	hash.Write(message)
	hash.Sum(hramDigest[:0])
	hramDigestReduced, _ := edwards25519.NewScalar().SetUniformBytes(hramDigest[:])

	sScalar := edwards25519.NewScalar().MultiplyAdd(hramDigestReduced, privateKeyScalar, rReduced)
	s := *(*[32]byte)(sScalar.Bytes())

	signature := new([64]byte)
	copy(signature[:], encodedR[:])
	copy(signature[32:], s[:])
	signature[63] |= publicKey[31] & 0x80

	return signature
}

// verify checks whether the message has a valid signature.
func verify(publicKey [32]byte, message []byte, signature *[64]byte) bool {

	publicKey[31] &= 0x7F

	/* Convert the Curve25519 public key into an Ed25519 public key.  In
	particular, convert Curve25519's "montgomery" x-coordinate into an
	Ed25519 "edwards" y-coordinate:

	ed_y = (mont_x - 1) / (mont_x + 1)

	NOTE: mont_x=-1 is converted to ed_y=0 since fe_invert is mod-exp

	Then move the sign bit into the pubkey from the signature.
	*/

	var edY, one, montX, montXMinusOne, montXPlusOne field.Element
	_, _ = montX.SetBytes(publicKey[:])
	_ = one.One()
	montXMinusOne.Subtract(&montX, &one)
	montXPlusOne.Add(&montX, &one)
	montXPlusOne.Invert(&montXPlusOne)
	edY.Multiply(&montXMinusOne, &montXPlusOne)

	A_ed := *(*[32]byte)(edY.Bytes())

	A_ed[31] |= signature[63] & 0x80
	signature[63] &= 0x7F

	return ed25519.Verify(A_ed[:], message, signature[:])
}
