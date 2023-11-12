package generator

import (
	"bytes"
	"errors"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/extkeys"
)

var (
	// ErrInvalidKeystoreExtendedKey is returned when the decrypted keystore file
	// contains some old Status keys.
	// The old version used to store the BIP44 account at index 0 as PrivateKey,
	// and the BIP44 account at index 1 as ExtendedKey.
	// The current version stores the same key as PrivateKey and ExtendedKey.
	ErrInvalidKeystoreExtendedKey  = errors.New("PrivateKey and ExtendedKey are different")
	ErrInvalidMnemonicPhraseLength = errors.New("invalid mnemonic phrase length; valid lengths are 12, 15, 18, 21, and 24")
)

// ValidateKeystoreExtendedKey validates the keystore keys, checking that
// ExtendedKey is the extended key of PrivateKey.
func ValidateKeystoreExtendedKey(key *types.Key) error {
	if key.ExtendedKey.IsZeroed() {
		return nil
	}

	if !bytes.Equal(crypto.FromECDSA(key.PrivateKey), crypto.FromECDSA(key.ExtendedKey.ToECDSA())) {
		return ErrInvalidKeystoreExtendedKey
	}

	return nil
}

// MnemonicPhraseLengthToEntropyStrength returns the entropy strength for a given mnemonic length
func MnemonicPhraseLengthToEntropyStrength(length int) (extkeys.EntropyStrength, error) {
	if length < 12 || length > 24 || length%3 != 0 {
		return 0, ErrInvalidMnemonicPhraseLength
	}

	bitsLength := length * 11
	checksumLength := bitsLength % 32

	return extkeys.EntropyStrength(bitsLength - checksumLength), nil
}
