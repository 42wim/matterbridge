package gethbridge

import (
	"crypto/ecdsa"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/extkeys"
)

type gethKeyStoreAdapter struct {
	keystore *keystore.KeyStore
}

// WrapKeyStore creates a types.KeyStore wrapper over a keystore.KeyStore object
func WrapKeyStore(keystore *keystore.KeyStore) types.KeyStore {
	return &gethKeyStoreAdapter{keystore: keystore}
}

func (k *gethKeyStoreAdapter) ImportECDSA(priv *ecdsa.PrivateKey, passphrase string) (types.Account, error) {
	gethAccount, err := k.keystore.ImportECDSA(priv, passphrase)
	return accountFrom(gethAccount), err
}

func (k *gethKeyStoreAdapter) ImportSingleExtendedKey(extKey *extkeys.ExtendedKey, passphrase string) (types.Account, error) {
	gethAccount, err := k.keystore.ImportSingleExtendedKey(extKey, passphrase)
	return accountFrom(gethAccount), err
}

func (k *gethKeyStoreAdapter) ImportExtendedKeyForPurpose(keyPurpose extkeys.KeyPurpose, extKey *extkeys.ExtendedKey, passphrase string) (types.Account, error) {
	gethAccount, err := k.keystore.ImportExtendedKeyForPurpose(keyPurpose, extKey, passphrase)
	return accountFrom(gethAccount), err
}

func (k *gethKeyStoreAdapter) AccountDecryptedKey(a types.Account, auth string) (types.Account, *types.Key, error) {
	gethAccount, err := gethAccountFrom(a)
	if err != nil {
		return types.Account{}, nil, err
	}

	var gethKey *keystore.Key
	gethAccount, gethKey, err = k.keystore.AccountDecryptedKey(gethAccount, auth)
	return accountFrom(gethAccount), keyFrom(gethKey), err
}

func (k *gethKeyStoreAdapter) Delete(a types.Account) error {
	gethAccount, err := gethAccountFrom(a)
	if err != nil {
		return err
	}

	return k.keystore.Delete(gethAccount)
}

// parseGethURL converts a user supplied URL into the accounts specific structure.
func parseGethURL(url string) (accounts.URL, error) {
	parts := strings.Split(url, "://")
	if len(parts) != 2 || parts[0] == "" {
		return accounts.URL{}, errors.New("protocol scheme missing")
	}
	return accounts.URL{
		Scheme: parts[0],
		Path:   parts[1],
	}, nil
}

func gethAccountFrom(account types.Account) (accounts.Account, error) {
	var (
		gethAccount accounts.Account
		err         error
	)
	gethAccount.Address = common.Address(account.Address)
	if account.URL != "" {
		gethAccount.URL, err = parseGethURL(account.URL)
	}
	return gethAccount, err
}

func accountFrom(gethAccount accounts.Account) types.Account {
	return types.Account{
		Address: types.Address(gethAccount.Address),
		URL:     gethAccount.URL.String(),
	}
}

func keyFrom(k *keystore.Key) *types.Key {
	if k == nil {
		return nil
	}

	return &types.Key{
		ID:              k.Id,
		Address:         types.Address(k.Address),
		PrivateKey:      k.PrivateKey,
		ExtendedKey:     k.ExtendedKey,
		SubAccountIndex: k.SubAccountIndex,
	}
}
