package record

import (
	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/keys/identity"
	"go.mau.fi/libsignal/util/bytehelper"
)

// NewPendingKeyExchange will return a new PendingKeyExchange object.
func NewPendingKeyExchange(sequence uint32, localBaseKeyPair, localRatchetKeyPair *ecc.ECKeyPair,
	localIdentityKeyPair *identity.KeyPair) *PendingKeyExchange {

	return &PendingKeyExchange{
		sequence:             sequence,
		localBaseKeyPair:     localBaseKeyPair,
		localRatchetKeyPair:  localRatchetKeyPair,
		localIdentityKeyPair: localIdentityKeyPair,
	}
}

// NewPendingKeyExchangeFromStruct will return a PendingKeyExchange object from
// the given structure. This is used to get a deserialized pending prekey exchange
// fetched from persistent storage.
func NewPendingKeyExchangeFromStruct(structure *PendingKeyExchangeStructure) *PendingKeyExchange {
	// Return nil if no structure was provided.
	if structure == nil {
		return nil
	}

	// Alias the SliceToArray method.
	getArray := bytehelper.SliceToArray

	// Convert the bytes in the given structure to ECC objects.
	localBaseKeyPair := ecc.NewECKeyPair(
		ecc.NewDjbECPublicKey(getArray(structure.LocalBaseKeyPublic)),
		ecc.NewDjbECPrivateKey(getArray(structure.LocalBaseKeyPrivate)),
	)
	localRatchetKeyPair := ecc.NewECKeyPair(
		ecc.NewDjbECPublicKey(getArray(structure.LocalRatchetKeyPublic)),
		ecc.NewDjbECPrivateKey(getArray(structure.LocalRatchetKeyPrivate)),
	)
	localIdentityKeyPair := identity.NewKeyPair(
		identity.NewKey(ecc.NewDjbECPublicKey(getArray(structure.LocalIdentityKeyPublic))),
		ecc.NewDjbECPrivateKey(getArray(structure.LocalIdentityKeyPrivate)),
	)

	// Return the PendingKeyExchange with the deserialized keys.
	return &PendingKeyExchange{
		sequence:             structure.Sequence,
		localBaseKeyPair:     localBaseKeyPair,
		localRatchetKeyPair:  localRatchetKeyPair,
		localIdentityKeyPair: localIdentityKeyPair,
	}
}

// PendingKeyExchangeStructure is a serializable structure for pending
// key exchanges. This structure is used for persistent storage of the
// key exchange state.
type PendingKeyExchangeStructure struct {
	Sequence                uint32
	LocalBaseKeyPublic      []byte
	LocalBaseKeyPrivate     []byte
	LocalRatchetKeyPublic   []byte
	LocalRatchetKeyPrivate  []byte
	LocalIdentityKeyPublic  []byte
	LocalIdentityKeyPrivate []byte
}

// PendingKeyExchange is a structure for storing a pending
// key exchange for a session state.
type PendingKeyExchange struct {
	sequence             uint32
	localBaseKeyPair     *ecc.ECKeyPair
	localRatchetKeyPair  *ecc.ECKeyPair
	localIdentityKeyPair *identity.KeyPair
}

// structre will return a serializable structure of a pending key exchange
// so it can be persistently stored.
func (p *PendingKeyExchange) structure() *PendingKeyExchangeStructure {
	getSlice := bytehelper.ArrayToSlice
	return &PendingKeyExchangeStructure{
		Sequence:                p.sequence,
		LocalBaseKeyPublic:      getSlice(p.localBaseKeyPair.PublicKey().PublicKey()),
		LocalBaseKeyPrivate:     getSlice(p.localBaseKeyPair.PrivateKey().Serialize()),
		LocalRatchetKeyPublic:   getSlice(p.localRatchetKeyPair.PublicKey().PublicKey()),
		LocalRatchetKeyPrivate:  getSlice(p.localRatchetKeyPair.PrivateKey().Serialize()),
		LocalIdentityKeyPublic:  getSlice(p.localIdentityKeyPair.PublicKey().PublicKey().PublicKey()),
		LocalIdentityKeyPrivate: getSlice(p.localIdentityKeyPair.PrivateKey().Serialize()),
	}
}
