package extkeys

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
)

// Implementation of the following BIPs:
//   - BIP32 (https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki)
//   - BIP39 (https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki)
//   - BIP44 (https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki)
//
// Referencing
// https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki
// https://bitcoin.org/en/developer-guide#hardened-keys

// Reference Implementations
// https://github.com/btcsuite/btcutil/tree/master/hdkeychain
// https://github.com/WeMeetAgain/go-hdwallet

// https://github.com/ConsenSys/eth-lightwallet/blob/master/lib/keystore.js
// https://github.com/bitpay/bitcore-lib/tree/master/lib

// MUST CREATE HARDENED CHILDREN OF THE MASTER PRIVATE KEY (M) TO PREVENT
// A COMPROMISED CHILD KEY FROM COMPROMISING THE MASTER KEY.
// AS THERE ARE NO NORMAL CHILDREN FOR THE MASTER KEYS,
// THE MASTER PUBLIC KEY IS NOT USED IN HD WALLETS.
// ALL OTHER KEYS CAN HAVE NORMAL CHILDREN,
// SO THE CORRESPONDING EXTENDED PUBLIC KEYS MAY BE USED INSTEAD.

// TODO make sure we're doing this ^^^^ !!!!!!

type KeyPurpose int

const (
	KeyPurposeWallet KeyPurpose = iota + 1
	KeyPurposeChat
)

const (
	// HardenedKeyStart defines a starting point for hardened key.
	// Each extended key has 2^31 normal child keys and 2^31 hardened child keys.
	// Thus the range for normal child keys is [0, 2^31 - 1] and the range for hardened child keys is [2^31, 2^32 - 1].
	HardenedKeyStart = 0x80000000 // 2^31

	// MinSeedBytes is the minimum number of bytes allowed for a seed to a master node.
	MinSeedBytes = 16 // 128 bits

	// MaxSeedBytes is the maximum number of bytes allowed for a seed to a master node.
	MaxSeedBytes = 64 // 512 bits

	// serializedKeyLen is the length of a serialized public or private
	// extended key.  It consists of 4 bytes version, 1 byte depth, 4 bytes
	// fingerprint, 4 bytes child number, 32 bytes chain code, and 33 bytes
	// public/private key data.
	serializedKeyLen = 4 + 1 + 4 + 4 + 32 + 33 // 78 bytes

	// CoinTypeBTC is BTC coin type
	CoinTypeBTC = 0 // 0x80000000

	// CoinTypeTestNet is test net coin type
	CoinTypeTestNet = 1 // 0x80000001

	// CoinTypeETH is ETH coin type
	CoinTypeETH = 60 // 0x8000003c

	// EmptyExtendedKeyString marker string for zero extended key
	EmptyExtendedKeyString = "Zeroed extended key"

	// MaxDepth is the maximum depth of an extended key.
	// Extended keys with depth MaxDepth cannot derive child keys.
	MaxDepth = 255
)

// errors
var (
	ErrInvalidKey                 = errors.New("key is invalid")
	ErrInvalidKeyPurpose          = errors.New("key purpose is invalid")
	ErrInvalidSeed                = errors.New("seed is invalid")
	ErrInvalidSeedLen             = fmt.Errorf("the recommended size of seed is %d-%d bits", MinSeedBytes, MaxSeedBytes)
	ErrDerivingHardenedFromPublic = errors.New("cannot derive a hardened key from public key")
	ErrBadChecksum                = errors.New("bad extended key checksum")
	ErrInvalidKeyLen              = errors.New("serialized extended key length is invalid")
	ErrDerivingChild              = errors.New("error deriving child key")
	ErrInvalidMasterKey           = errors.New("invalid master key supplied")
	ErrMaxDepthExceeded           = errors.New("max depth exceeded")
)

var (
	// PrivateKeyVersion is version for private key
	PrivateKeyVersion, _ = hex.DecodeString("0488ADE4")

	// PublicKeyVersion is version for public key
	PublicKeyVersion, _ = hex.DecodeString("0488B21E")

	// EthBIP44ParentPath is BIP44 keys parent's derivation path
	EthBIP44ParentPath = []uint32{
		HardenedKeyStart + 44,          // purpose
		HardenedKeyStart + CoinTypeETH, // cointype set to ETH
		HardenedKeyStart + 0,           // account
		0,                              // 0 - public, 1 - private
	}

	// EIP1581KeyTypeChat is used as chat key_type in the derivation of EIP1581 keys
	EIP1581KeyTypeChat uint32 = 0x00

	// EthEIP1581ChatParentPath is EIP-1581 chat keys parent's derivation path
	EthEIP1581ChatParentPath = []uint32{
		HardenedKeyStart + 43,                 // purpose
		HardenedKeyStart + CoinTypeETH,        // cointype set to ETH
		HardenedKeyStart + 1581,               // EIP-1581 subpurpose
		HardenedKeyStart + EIP1581KeyTypeChat, // key_type (chat)
	}
)

// ExtendedKey represents BIP44-compliant HD key
type ExtendedKey struct {
	Version          []byte // 4 bytes, mainnet: 0x0488B21E public, 0x0488ADE4 private; testnet: 0x043587CF public, 0x04358394 private
	Depth            uint8  // 1 byte,  depth: 0x00 for master nodes, 0x01 for level-1 derived keys, ....
	FingerPrint      []byte // 4 bytes, fingerprint of the parent's key (0x00000000 if master key)
	ChildNumber      uint32 // 4 bytes, This is ser32(i) for i in xi = xpar/i, with xi the key being serialized. (0x00000000 if master key)
	KeyData          []byte // 33 bytes, the public key or private key data (serP(K) for public keys, 0x00 || ser256(k) for private keys)
	ChainCode        []byte // 32 bytes, the chain code
	IsPrivate        bool   // (non-serialized) if false, this chain will only contain a public key and can only create a public key chain.
	CachedPubKeyData []byte // (non-serialized) used for memoization of public key (calculated from a private key)
}

// nolint: gas
const masterSecret = "Bitcoin seed"

// NewMaster creates new master node, root of HD chain/tree.
// Both master and child nodes are of ExtendedKey type, and all the children derive from the root node.
func NewMaster(seed []byte) (*ExtendedKey, error) {
	// Ensure seed is within expected limits
	lseed := len(seed)
	if lseed < MinSeedBytes || lseed > MaxSeedBytes {
		return nil, ErrInvalidSeedLen
	}

	secretKey, chainCode, err := splitHMAC(seed, []byte(masterSecret))
	if err != nil {
		return nil, err
	}

	master := &ExtendedKey{
		Version:     PrivateKeyVersion,
		Depth:       0,
		FingerPrint: []byte{0x00, 0x00, 0x00, 0x00},
		ChildNumber: 0,
		KeyData:     secretKey,
		ChainCode:   chainCode,
		IsPrivate:   true,
	}

	return master, nil
}

// Child derives extended key at a given index i.
// If parent is private, then derived key is also private. If parent is public, then derived is public.
//
// If i >= HardenedKeyStart, then hardened key is generated.
// You can only generate hardened keys from private parent keys.
// If you try generating hardened key form public parent key, ErrDerivingHardenedFromPublic is returned.
//
// There are four CKD (child key derivation) scenarios:
// 1) Private extended key -> Hardened child private extended key
// 2) Private extended key -> Non-hardened child private extended key
// 3) Public extended key -> Non-hardened child public extended key
// 4) Public extended key -> Hardened child public extended key (INVALID!)
func (k *ExtendedKey) Child(i uint32) (*ExtendedKey, error) {
	if k.Depth == MaxDepth {
		return nil, ErrMaxDepthExceeded
	}

	// A hardened child may not be created from a public extended key (Case #4).
	isChildHardened := i >= HardenedKeyStart
	if !k.IsPrivate && isChildHardened {
		return nil, ErrDerivingHardenedFromPublic
	}

	keyLen := 33
	seed := make([]byte, keyLen+4)
	if isChildHardened {
		// Case #1: 0x00 || ser256(parentKey) || ser32(i)
		copy(seed[1:], k.KeyData) // 0x00 || ser256(parentKey)
	} else {
		// Case #2 and #3: serP(parentPubKey) || ser32(i)
		copy(seed, k.pubKeyBytes())
	}
	binary.BigEndian.PutUint32(seed[keyLen:], i)

	secretKey, chainCode, err := splitHMAC(seed, k.ChainCode)
	if err != nil {
		return nil, err
	}

	child := &ExtendedKey{
		ChainCode:   chainCode,
		Depth:       k.Depth + 1,
		ChildNumber: i,
		IsPrivate:   k.IsPrivate,
		// The fingerprint for the derived child is the first 4 bytes of parent's
		FingerPrint: btcutil.Hash160(k.pubKeyBytes())[:4],
	}

	if k.IsPrivate {
		// Case #1 or #2: childKey = parse256(IL) + parentKey
		parentKeyBigInt := new(big.Int).SetBytes(k.KeyData)
		keyBigInt := new(big.Int).SetBytes(secretKey)
		keyBigInt.Add(keyBigInt, parentKeyBigInt)
		keyBigInt.Mod(keyBigInt, btcec.S256().N)

		// Make sure that child.KeyData is 32 bytes of data even if the value is represented with less bytes.
		// When we derive a child of this key, we call splitHMAC that does a sha512 of a seed that is:
		// - 1 byte with 0x00
		// - 32 bytes for the key data
		// - 4 bytes for the child key index
		// If we don't padd the KeyData, it will be shifted to left in that 32 bytes space
		// generating a different seed and different child key.
		// This part fixes a bug we had previously and described at:
		// https://medium.com/@alexberegszaszi/why-do-my-bip32-wallets-disagree-6f3254cc5846#.86inuifuq
		keyData := keyBigInt.Bytes()
		if len(keyData) < 32 {
			extra := make([]byte, 32-len(keyData))
			keyData = append(extra, keyData...)
		}

		child.KeyData = keyData
		child.Version = PrivateKeyVersion
	} else {
		// Case #3: childKey = serP(point(parse256(IL)) + parentKey)

		// Calculate the corresponding intermediate public key for intermediate private key.
		keyx, keyy := btcec.S256().ScalarBaseMult(secretKey)
		if keyx.Sign() == 0 || keyy.Sign() == 0 {
			return nil, ErrInvalidKey
		}

		// Convert the serialized compressed parent public key into X and Y coordinates
		// so it can be added to the intermediate public key.
		pubKey, err := btcec.ParsePubKey(k.KeyData, btcec.S256())
		if err != nil {
			return nil, err
		}

		// childKey = serP(point(parse256(IL)) + parentKey)
		childX, childY := btcec.S256().Add(keyx, keyy, pubKey.X, pubKey.Y)
		pk := btcec.PublicKey{Curve: btcec.S256(), X: childX, Y: childY}
		child.KeyData = pk.SerializeCompressed()
		child.Version = PublicKeyVersion
	}
	return child, nil
}

// ChildForPurpose derives the child key at index i using a derivation path based on the purpose.
func (k *ExtendedKey) ChildForPurpose(p KeyPurpose, i uint32) (*ExtendedKey, error) {
	switch p {
	case KeyPurposeWallet:
		return k.EthBIP44Child(i)
	case KeyPurposeChat:
		return k.EthEIP1581ChatChild(i)
	default:
		return nil, ErrInvalidKeyPurpose
	}
}

// BIP44Child returns Status CKD#i (where i is child index).
// BIP44 format is used: m / purpose' / coin_type' / account' / change / address_index
// BIP44Child is depracated in favour of EthBIP44Child
// Param coinType is deprecated; we override it to always use CoinTypeETH.
func (k *ExtendedKey) BIP44Child(coinType, i uint32) (*ExtendedKey, error) {
	return k.EthBIP44Child(i)
}

// BIP44Child returns Status CKD#i (where i is child index).
// BIP44 format is used: m / purpose' / coin_type' / account' / change / address_index
func (k *ExtendedKey) EthBIP44Child(i uint32) (*ExtendedKey, error) {
	if !k.IsPrivate {
		return nil, ErrInvalidMasterKey
	}

	if k.Depth != 0 {
		return nil, ErrInvalidMasterKey
	}

	// m/44'/60'/0'/0/index
	extKey, err := k.Derive(append(EthBIP44ParentPath, i))
	if err != nil {
		return nil, err
	}

	return extKey, nil
}

// EthEIP1581ChatChild returns the whisper key #i (where i is child index).
// EthEIP1581ChatChild format is used is the one defined in the EIP-1581:
// m / 43' / coin_type' / 1581' / key_type / index
func (k *ExtendedKey) EthEIP1581ChatChild(i uint32) (*ExtendedKey, error) {
	if !k.IsPrivate {
		return nil, ErrInvalidMasterKey
	}

	if k.Depth != 0 {
		return nil, ErrInvalidMasterKey
	}

	// m/43'/60'/1581'/0/index
	extKey, err := k.Derive(append(EthEIP1581ChatParentPath, i))
	if err != nil {
		return nil, err
	}

	return extKey, nil
}

// Derive returns a derived child key at a given path
func (k *ExtendedKey) Derive(path []uint32) (*ExtendedKey, error) {
	var err error
	extKey := k
	for _, i := range path {
		extKey, err = extKey.Child(i)
		if err != nil {
			return nil, ErrDerivingChild
		}
	}

	return extKey, nil
}

// Neuter returns a new extended public key from a give extended private key.
// If the input extended key is already public, it will be returned unaltered.
func (k *ExtendedKey) Neuter() (*ExtendedKey, error) {
	// Already an extended public key.
	if !k.IsPrivate {
		return k, nil
	}

	// Get the associated public extended key version bytes.
	version, err := chaincfg.HDPrivateKeyToPublicKeyID(k.Version)
	if err != nil {
		return nil, err
	}

	// Convert it to an extended public key.  The key for the new extended
	// key will simply be the pubkey of the current extended private key.
	return &ExtendedKey{
		Version:     version,
		KeyData:     k.pubKeyBytes(),
		ChainCode:   k.ChainCode,
		FingerPrint: k.FingerPrint,
		Depth:       k.Depth,
		ChildNumber: k.ChildNumber,
		IsPrivate:   false,
	}, nil
}

// IsZeroed returns true if key is nil or empty
func (k *ExtendedKey) IsZeroed() bool {
	return k == nil || len(k.KeyData) == 0
}

// String returns the extended key as a human-readable base58-encoded string.
func (k *ExtendedKey) String() string {
	if k.IsZeroed() {
		return EmptyExtendedKeyString
	}

	var childNumBytes [4]byte
	binary.BigEndian.PutUint32(childNumBytes[:], k.ChildNumber)

	// The serialized format is:
	//   version (4) || depth (1) || parent fingerprint (4)) ||
	//   child num (4) || chain code (32) || key data (33) || checksum (4)
	serializedBytes := make([]byte, 0, serializedKeyLen+4)
	serializedBytes = append(serializedBytes, k.Version...)
	serializedBytes = append(serializedBytes, k.Depth)
	serializedBytes = append(serializedBytes, k.FingerPrint...)
	serializedBytes = append(serializedBytes, childNumBytes[:]...)
	serializedBytes = append(serializedBytes, k.ChainCode...)
	if k.IsPrivate {
		serializedBytes = append(serializedBytes, 0x00)
		serializedBytes = paddedAppend(32, serializedBytes, k.KeyData)
	} else {
		serializedBytes = append(serializedBytes, k.pubKeyBytes()...)
	}

	checkSum := chainhash.DoubleHashB(serializedBytes)[:4]
	serializedBytes = append(serializedBytes, checkSum...)
	return base58.Encode(serializedBytes)
}

// pubKeyBytes returns bytes for the serialized compressed public key associated
// with this extended key in an efficient manner including memoization as
// necessary.
//
// When the extended key is already a public key, the key is simply returned as
// is since it's already in the correct form.  However, when the extended key is
// a private key, the public key will be calculated and memoized so future
// accesses can simply return the cached result.
func (k *ExtendedKey) pubKeyBytes() []byte {
	// Just return the key if it's already an extended public key.
	if !k.IsPrivate {
		return k.KeyData
	}

	pkx, pky := btcec.S256().ScalarBaseMult(k.KeyData)
	pubKey := btcec.PublicKey{Curve: btcec.S256(), X: pkx, Y: pky}
	return pubKey.SerializeCompressed()
}

// ToECDSA returns the key data as ecdsa.PrivateKey
func (k *ExtendedKey) ToECDSA() *ecdsa.PrivateKey {
	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), k.KeyData)
	return privKey.ToECDSA()
}

// NewKeyFromString returns a new extended key instance from a base58-encoded
// extended key.
func NewKeyFromString(key string) (*ExtendedKey, error) {
	if key == EmptyExtendedKeyString || len(key) == 0 {
		return &ExtendedKey{}, nil
	}

	// The base58-decoded extended key must consist of a serialized payload
	// plus an additional 4 bytes for the checksum.
	decoded := base58.Decode(key)
	if len(decoded) != serializedKeyLen+4 {
		return nil, ErrInvalidKeyLen
	}

	// The serialized format is:
	//   version (4) || depth (1) || parent fingerprint (4)) ||
	//   child num (4) || chain code (32) || key data (33) || checksum (4)

	// Split the payload and checksum up and ensure the checksum matches.
	payload := decoded[:len(decoded)-4]
	checkSum := decoded[len(decoded)-4:]
	expectedCheckSum := chainhash.DoubleHashB(payload)[:4]
	if !bytes.Equal(checkSum, expectedCheckSum) {
		return nil, ErrBadChecksum
	}

	// Deserialize each of the payload fields.
	version := payload[:4]
	depth := payload[4:5][0]
	fingerPrint := payload[5:9]
	childNumber := binary.BigEndian.Uint32(payload[9:13])
	chainCode := payload[13:45]
	keyData := payload[45:78]

	// The key data is a private key if it starts with 0x00.  Serialized
	// compressed pubkeys either start with 0x02 or 0x03.
	isPrivate := keyData[0] == 0x00
	if isPrivate {
		// Ensure the private key is valid.  It must be within the range
		// of the order of the secp256k1 curve and not be 0.
		keyData = keyData[1:]
		keyNum := new(big.Int).SetBytes(keyData)
		if keyNum.Cmp(btcec.S256().N) >= 0 || keyNum.Sign() == 0 {
			return nil, ErrInvalidSeed
		}
	} else {
		// Ensure the public key parses correctly and is actually on the
		// secp256k1 curve.
		_, err := btcec.ParsePubKey(keyData, btcec.S256())
		if err != nil {
			return nil, err
		}
	}

	return &ExtendedKey{
		Version:     version,
		KeyData:     keyData,
		ChainCode:   chainCode,
		FingerPrint: fingerPrint,
		Depth:       depth,
		ChildNumber: childNumber,
		IsPrivate:   isPrivate,
	}, nil
}
