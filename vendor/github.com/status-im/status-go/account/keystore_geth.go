package account

import (
	"os"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"

	gethbridge "github.com/status-im/status-go/eth-node/bridge/geth"
	"github.com/status-im/status-go/eth-node/types"
)

// makeAccountManager creates ethereum accounts.Manager with single disk backend and lightweight kdf.
// If keydir is empty new temporary directory with go-ethereum-keystore will be intialized.
func makeAccountManager(keydir string) (manager *accounts.Manager, err error) {
	if keydir == "" {
		// There is no datadir.
		keydir, err = os.MkdirTemp("", "go-ethereum-keystore")
	}
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(keydir, 0700); err != nil {
		return nil, err
	}
	config := accounts.Config{InsecureUnlockAllowed: false}
	return accounts.NewManager(&config, keystore.NewKeyStore(keydir, keystore.LightScryptN, keystore.LightScryptP)), nil
}

func makeKeyStore(manager *accounts.Manager) (types.KeyStore, error) {
	backends := manager.Backends(keystore.KeyStoreType)
	if len(backends) == 0 {
		return nil, ErrAccountKeyStoreMissing
	}
	keyStore, ok := backends[0].(*keystore.KeyStore)
	if !ok {
		return nil, ErrAccountKeyStoreMissing
	}

	return gethbridge.WrapKeyStore(keyStore), nil
}
