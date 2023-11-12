package types

import (
	"crypto/ecdsa"

	"github.com/status-im/status-go/extkeys"
)

type KeyStore interface {
	// ImportAccount imports the account specified with privateKey.
	ImportECDSA(priv *ecdsa.PrivateKey, passphrase string) (Account, error)
	// ImportSingleExtendedKey imports an extended key setting it in both the PrivateKey and ExtendedKey fields
	// of the Key struct.
	// ImportExtendedKey is used in older version of Status where PrivateKey is set to be the BIP44 key at index 0,
	// and ExtendedKey is the extended key of the BIP44 key at index 1.
	ImportSingleExtendedKey(extKey *extkeys.ExtendedKey, passphrase string) (Account, error)
	// ImportExtendedKeyForPurpose stores ECDSA key (obtained from extended key) along with CKD#2 (root for sub-accounts)
	// If key file is not found, it is created. Key is encrypted with the given passphrase.
	// Deprecated: status-go is now using ImportSingleExtendedKey
	ImportExtendedKeyForPurpose(keyPurpose extkeys.KeyPurpose, extKey *extkeys.ExtendedKey, passphrase string) (Account, error)
	// AccountDecryptedKey returns decrypted key for account (provided that password is correct).
	AccountDecryptedKey(a Account, auth string) (Account, *Key, error)
	// Delete deletes the key matched by account if the passphrase is correct.
	// If the account contains no filename, the address must match a unique key.
	Delete(a Account) error
}
