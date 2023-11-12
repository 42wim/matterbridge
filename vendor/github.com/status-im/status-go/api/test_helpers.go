package api

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/multiaccounts"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/multiaccounts/settings"
	"github.com/status-im/status-go/params"

	"github.com/stretchr/testify/require"
)

func setupWalletTest(t *testing.T, password string) (backend *GethStatusBackend, defersFunc func(), err error) {
	tmpdir := t.TempDir()

	defers := make([]func(), 0)
	defersFunc = func() {
		for _, f := range defers {
			f()
		}
	}
	if err != nil {
		return
	}

	backend = NewGethStatusBackend()
	backend.UpdateRootDataDir(tmpdir)

	err = backend.AccountManager().InitKeystore(filepath.Join(tmpdir, "keystore"))

	if err != nil {
		return
	}

	// Create master account
	const pathWalletRoot = "m/44'/60'/0'/0"
	accs, err := backend.AccountManager().
		AccountsGenerator().
		GenerateAndDeriveAddresses(12, 1, "", []string{pathWalletRoot})
	if err != nil {
		return
	}

	masterAccInfo := accs[0]

	_, err = backend.AccountManager().AccountsGenerator().StoreDerivedAccounts(masterAccInfo.ID, password, []string{pathWalletRoot})

	if err != nil {
		return
	}

	account := multiaccounts.Account{
		Name:           "foo",
		Timestamp:      1,
		KeycardPairing: "pairing",
		KeyUID:         masterAccInfo.KeyUID,
	}

	err = backend.ensureDBsOpened(account, password)
	require.NoError(t, err)

	walletRootAddress := masterAccInfo.Derived[pathWalletRoot].Address

	config, err := params.NewNodeConfig(tmpdir, 178733)
	require.NoError(t, err)
	networks := json.RawMessage("{}")
	s := settings.Settings{
		Address:           types.HexToAddress(walletRootAddress),
		DisplayName:       "UserDisplayName",
		CurrentNetwork:    "mainnet_rpc",
		DappsAddress:      types.HexToAddress(walletRootAddress),
		EIP1581Address:    types.HexToAddress(walletRootAddress),
		InstallationID:    "d3efcff6-cffa-560e-a547-21d3858cbc51",
		KeyUID:            account.KeyUID,
		LatestDerivedPath: 0,
		Name:              "Jittery Cornflowerblue Kingbird",
		Networks:          &networks,
		PhotoPath:         "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADIAAAAyCAIAAACRXR/mAAAAjklEQVR4nOzXwQmFMBAAUZXUYh32ZB32ZB02sxYQQSZGsod55/91WFgSS0RM+SyjA56ZRZhFmEWYRRT6h+M6G16zrxv6fdJpmUWYRbxsYr13dKfanpN0WmYRZhGzXz6AWYRZRIfbaX26fT9Jk07LLMIsosPt9I/dTDotswizCG+nhFmEWYRZhFnEHQAA///z1CFkYamgfQAAAABJRU5ErkJggg==",
		PreviewPrivacy:    false,
		PublicKey:         masterAccInfo.PublicKey,
		SigningPhrase:     "yurt joey vibe",
		WalletRootAddress: types.HexToAddress(walletRootAddress)}

	err = backend.saveAccountsAndSettings(s, config, nil)
	require.Error(t, err)
	require.True(t, err == accounts.ErrKeypairWithoutAccounts)

	// this is for StatusNode().Config() call inside of the getVerifiedWalletAccount
	err = backend.StartNode(config)
	require.NoError(t, err)

	defers = append(defers, func() {
		require.NoError(t, backend.StopNode())
	})

	return
}
