package accounts

import (
	"context"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"

	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/multiaccounts/accounts"
	walletsettings "github.com/status-im/status-go/multiaccounts/settings_wallet"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/protocol"
	"github.com/status-im/status-go/services/accounts/accountsevent"
)

func NewAccountsAPI(manager *account.GethManager, config *params.NodeConfig, db *accounts.Database, feed *event.Feed, messenger **protocol.Messenger) *API {
	return &API{manager, config, db, feed, messenger}
}

// API is class with methods available over RPC.
type API struct {
	manager   *account.GethManager
	config    *params.NodeConfig
	db        *accounts.Database
	feed      *event.Feed
	messenger **protocol.Messenger
}

type DerivedAddress struct {
	Address        common.Address `json:"address"`
	Path           string         `json:"path"`
	HasActivity    bool           `json:"hasActivity"`
	AlreadyCreated bool           `json:"alreadyCreated"`
}

func (api *API) SaveAccount(ctx context.Context, account *accounts.Account) error {
	log.Info("[AccountsAPI::SaveAccount]")
	err := (*api.messenger).SaveOrUpdateAccount(account)
	if err != nil {
		return err
	}

	api.feed.Send(accountsevent.Event{
		Type:     accountsevent.EventTypeAdded,
		Accounts: []common.Address{common.Address(account.Address)},
	})
	return nil
}

// Setting `Keypair` without `Accounts` will update keypair only, `Keycards` won't be saved/updated this way.
func (api *API) SaveKeypair(ctx context.Context, keypair *accounts.Keypair) error {
	log.Info("[AccountsAPI::SaveKeypair]")
	err := (*api.messenger).SaveOrUpdateKeypair(keypair)
	if err != nil {
		return err
	}

	commonAddresses := []common.Address{}
	for _, acc := range keypair.Accounts {
		commonAddresses = append(commonAddresses, common.Address(acc.Address))
	}

	api.feed.Send(accountsevent.Event{
		Type:     accountsevent.EventTypeAdded,
		Accounts: commonAddresses,
	})
	return nil
}

func (api *API) HasPairedDevices(ctx context.Context) bool {
	return (*api.messenger).HasPairedDevices()
}

// Setting `Keypair` without `Accounts` will update keypair only.
func (api *API) UpdateKeypairName(ctx context.Context, keyUID string, name string) error {
	return (*api.messenger).UpdateKeypairName(keyUID, name)
}

func (api *API) MoveWalletAccount(ctx context.Context, fromPosition int64, toPosition int64) error {
	return (*api.messenger).MoveWalletAccount(fromPosition, toPosition)
}

func (api *API) UpdateTokenPreferences(ctx context.Context, preferences []walletsettings.TokenPreferences) error {
	return (*api.messenger).UpdateTokenPreferences(preferences)
}

func (api *API) GetTokenPreferences(ctx context.Context) ([]walletsettings.TokenPreferences, error) {
	return (*api.messenger).GetTokenPreferences()
}

func (api *API) UpdateCollectiblePreferences(ctx context.Context, preferences []walletsettings.CollectiblePreferences) error {
	return (*api.messenger).UpdateCollectiblePreferences(preferences)
}

func (api *API) GetCollectiblePreferences(ctx context.Context) ([]walletsettings.CollectiblePreferences, error) {
	return (*api.messenger).GetCollectiblePreferences()
}

func (api *API) GetAccounts(ctx context.Context) ([]*accounts.Account, error) {
	return api.db.GetActiveAccounts()
}

func (api *API) GetWatchOnlyAccounts(ctx context.Context) ([]*accounts.Account, error) {
	return api.db.GetActiveWatchOnlyAccounts()
}

func (api *API) GetKeypairs(ctx context.Context) ([]*accounts.Keypair, error) {
	return api.db.GetActiveKeypairs()
}

func (api *API) GetAccountByAddress(ctx context.Context, address types.Address) (*accounts.Account, error) {
	return api.db.GetAccountByAddress(address)
}

func (api *API) GetKeypairByKeyUID(ctx context.Context, keyUID string) (*accounts.Keypair, error) {
	return api.db.GetKeypairByKeyUID(keyUID)
}

func (api *API) DeleteAccount(ctx context.Context, address types.Address) error {
	err := (*api.messenger).DeleteAccount(address)
	if err != nil {
		return err
	}

	api.feed.Send(accountsevent.Event{
		Type:     accountsevent.EventTypeRemoved,
		Accounts: []common.Address{common.Address(address)},
	})

	return nil
}

func (api *API) DeleteKeypair(ctx context.Context, keyUID string) error {
	keypair, err := api.db.GetKeypairByKeyUID(keyUID)
	if err != nil {
		return err
	}

	err = (*api.messenger).DeleteKeypair(keyUID)
	if err != nil {
		return err
	}

	var addresses []common.Address
	for _, acc := range keypair.Accounts {
		if acc.Chat {
			continue
		}
		addresses = append(addresses, common.Address(acc.Address))
	}

	api.feed.Send(accountsevent.Event{
		Type:     accountsevent.EventTypeRemoved,
		Accounts: addresses,
	})

	return nil
}

func (api *API) AddKeypair(ctx context.Context, password string, keypair *accounts.Keypair) error {
	if len(keypair.KeyUID) == 0 {
		return errors.New("`KeyUID` field of a keypair must be set")
	}

	if len(keypair.Name) == 0 {
		return errors.New("`Name` field of a keypair must be set")
	}

	if len(keypair.Type) == 0 {
		return errors.New("`Type` field of a keypair must be set")
	}

	if keypair.Type != accounts.KeypairTypeKey {
		if len(keypair.DerivedFrom) == 0 {
			return errors.New("`DerivedFrom` field of a keypair must be set")
		}
	}

	for _, acc := range keypair.Accounts {
		if acc.KeyUID != keypair.KeyUID {
			return errors.New("all accounts of a keypair must have the same `KeyUID` as keypair key uid")
		}

		err := api.checkAccountValidity(acc)
		if err != nil {
			return err
		}
	}

	err := api.SaveKeypair(ctx, keypair)
	if err != nil {
		return err
	}

	if len(password) > 0 {
		for _, acc := range keypair.Accounts {
			if acc.Type == accounts.AccountTypeGenerated || acc.Type == accounts.AccountTypeSeed {
				err = api.createKeystoreFileForAccount(keypair.DerivedFrom, password, acc)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (api *API) checkAccountValidity(account *accounts.Account) error {
	if len(account.Address) == 0 {
		return errors.New("`Address` field of an account must be set")
	}

	if len(account.Type) == 0 {
		return errors.New("`Type` field of an account must be set")
	}

	if account.Wallet || account.Chat {
		return errors.New("default wallet and chat account cannot be added this way")
	}

	if len(account.Name) == 0 {
		return errors.New("`Name` field of an account must be set")
	}

	if len(account.Emoji) == 0 {
		return errors.New("`Emoji` field of an account must be set")
	}

	if len(account.ColorID) == 0 {
		return errors.New("`ColorID` field of an account must be set")
	}

	if account.Type != accounts.AccountTypeWatch {

		if len(account.KeyUID) == 0 {
			return errors.New("`KeyUID` field of an account must be set")
		}

		if len(account.PublicKey) == 0 {
			return errors.New("`PublicKey` field of an account must be set")
		}

		if account.Type != accounts.AccountTypeKey {
			if len(account.Path) == 0 {
				return errors.New("`Path` field of an account must be set")
			}
		}
	}

	addressExists, err := api.db.AddressExists(account.Address)
	if err != nil {
		return err
	}

	if addressExists {
		return errors.New("account already exists")
	}

	return nil
}

func (api *API) createKeystoreFileForAccount(masterAddress string, password string, account *accounts.Account) error {
	if account.Type != accounts.AccountTypeGenerated && account.Type != accounts.AccountTypeSeed {
		return errors.New("cannot create keystore file if account is not of `generated` or `seed` type")
	}
	if masterAddress == "" {
		return errors.New("cannot create keystore file if master address is empty")
	}
	if password == "" {
		return errors.New("cannot create keystore file if password is empty")
	}

	info, err := api.manager.AccountsGenerator().LoadAccount(masterAddress, password)
	if err != nil {
		return err
	}

	_, err = api.manager.AccountsGenerator().StoreDerivedAccounts(info.ID, password, []string{account.Path})
	return err
}

func (api *API) AddAccount(ctx context.Context, password string, account *accounts.Account) error {
	err := api.checkAccountValidity(account)
	if err != nil {
		return err
	}

	if account.Type != accounts.AccountTypeWatch {
		kp, err := api.db.GetKeypairByKeyUID(account.KeyUID)
		if err != nil {
			if err == accounts.ErrDbKeypairNotFound {
				return errors.New("cannot add an account for an unknown keypair")
			}
			return err
		}

		// we need to create local keystore file only if password is provided and the account is being added is of
		// "generated" or "seed" type.
		if (account.Type == accounts.AccountTypeGenerated || account.Type == accounts.AccountTypeSeed) && len(password) > 0 {
			err = api.createKeystoreFileForAccount(kp.DerivedFrom, password, account)
			if err != nil {
				return err
			}
		}
	}

	if account.Type == accounts.AccountTypeGenerated {
		account.AddressWasNotShown = true
	}

	return api.SaveAccount(ctx, account)
}

// Imports a new private key and creates local keystore file.
func (api *API) ImportPrivateKey(ctx context.Context, privateKey string, password string) error {
	info, err := api.manager.AccountsGenerator().ImportPrivateKey(privateKey)
	if err != nil {
		return err
	}

	kp, err := api.db.GetKeypairByKeyUID(info.KeyUID)
	if err != nil && err != accounts.ErrDbKeypairNotFound {
		return err
	}

	if kp != nil {
		return errors.New("provided private key was already imported")
	}

	_, err = api.manager.AccountsGenerator().StoreAccount(info.ID, password)
	return err
}

// Creates all keystore files for a keypair and mark it in db as fully operable.
func (api *API) MakePrivateKeyKeypairFullyOperable(ctx context.Context, privateKey string, password string) error {
	info, err := api.manager.AccountsGenerator().ImportPrivateKey(privateKey)
	if err != nil {
		return err
	}

	kp, err := api.db.GetKeypairByKeyUID(info.KeyUID)
	if err != nil {
		return err
	}

	if kp == nil {
		return errors.New("keypair for the provided private key is not known")
	}

	_, err = api.manager.AccountsGenerator().StoreAccount(info.ID, password)
	if err != nil {
		return err
	}

	return (*api.messenger).MarkKeypairFullyOperable(info.KeyUID)
}

func (api *API) MakePartiallyOperableAccoutsFullyOperable(ctx context.Context, password string) (addresses []types.Address, err error) {
	profileKeypair, err := api.db.GetProfileKeypair()
	if err != nil {
		return
	}

	if !profileKeypair.MigratedToKeycard() && !api.VerifyPassword(password) {
		err = errors.New("wrong password provided")
		return
	}

	keypairs, err := api.db.GetActiveKeypairs()
	if err != nil {
		return
	}

	for _, kp := range keypairs {
		for _, acc := range kp.Accounts {
			if acc.Operable != accounts.AccountPartiallyOperable {
				continue
			}
			err = api.createKeystoreFileForAccount(kp.DerivedFrom, password, acc)
			if err != nil {
				return
			}
			err = api.db.MarkAccountFullyOperable(acc.Address)
			if err != nil {
				return
			}
			addresses = append(addresses, acc.Address)
		}
	}
	return
}

// Imports a new mnemonic and creates local keystore file.
func (api *API) ImportMnemonic(ctx context.Context, mnemonic string, password string) error {
	mnemonicNoExtraSpaces := strings.Join(strings.Fields(mnemonic), " ")

	generatedAccountInfo, err := api.manager.AccountsGenerator().ImportMnemonic(mnemonicNoExtraSpaces, "")
	if err != nil {
		return err
	}

	kp, err := api.db.GetKeypairByKeyUID(generatedAccountInfo.KeyUID)
	if err != nil && err != accounts.ErrDbKeypairNotFound {
		return err
	}

	if kp != nil {
		return errors.New("provided mnemonic was already imported, to add new account use `AddAccount` endpoint")
	}

	_, err = api.manager.AccountsGenerator().StoreAccount(generatedAccountInfo.ID, password)
	return err
}

// Creates all keystore files for a keypair and mark it in db as fully operable.
func (api *API) MakeSeedPhraseKeypairFullyOperable(ctx context.Context, mnemonic string, password string) error {
	mnemonicNoExtraSpaces := strings.Join(strings.Fields(mnemonic), " ")

	generatedAccountInfo, err := api.manager.AccountsGenerator().ImportMnemonic(mnemonicNoExtraSpaces, "")
	if err != nil {
		return err
	}

	kp, err := api.db.GetKeypairByKeyUID(generatedAccountInfo.KeyUID)
	if err != nil {
		return err
	}

	if kp == nil {
		return errors.New("keypair for the provided seed phrase is not known")
	}

	_, err = api.manager.AccountsGenerator().StoreAccount(generatedAccountInfo.ID, password)
	if err != nil {
		return err
	}

	var paths []string
	for _, acc := range kp.Accounts {
		paths = append(paths, acc.Path)
	}

	_, err = api.manager.AccountsGenerator().StoreDerivedAccounts(generatedAccountInfo.ID, password, paths)
	if err != nil {
		return err
	}

	return (*api.messenger).MarkKeypairFullyOperable(generatedAccountInfo.KeyUID)
}

// Creates a random new mnemonic.
func (api *API) GetRandomMnemonic(ctx context.Context) (string, error) {
	return account.GetRandomMnemonic()
}

func (api *API) VerifyKeystoreFileForAccount(address types.Address, password string) bool {
	_, err := api.manager.VerifyAccountPassword(api.config.KeyStoreDir, address.Hex(), password)
	return err == nil
}

func (api *API) VerifyPassword(password string) bool {
	address, err := api.db.GetChatAddress()
	if err != nil {
		return false
	}
	return api.VerifyKeystoreFileForAccount(address, password)
}

func (api *API) MigrateNonProfileKeycardKeypairToApp(ctx context.Context, mnemonic string, password string) error {
	mnemonicNoExtraSpaces := strings.Join(strings.Fields(mnemonic), " ")

	generatedAccountInfo, err := api.manager.AccountsGenerator().ImportMnemonic(mnemonicNoExtraSpaces, "")
	if err != nil {
		return err
	}

	kp, err := api.db.GetKeypairByKeyUID(generatedAccountInfo.KeyUID)
	if err != nil {
		return err
	}

	if kp.Type == accounts.KeypairTypeProfile {
		return errors.New("cannot migrate profile keypair")
	}

	if !kp.MigratedToKeycard() {
		return errors.New("keypair being migrated is not a keycard keypair")
	}

	profileKeypair, err := api.db.GetProfileKeypair()
	if err != nil {
		return err
	}

	if !profileKeypair.MigratedToKeycard() && !api.VerifyPassword(password) {
		return errors.New("wrong password provided")
	}

	_, err = api.manager.AccountsGenerator().StoreAccount(generatedAccountInfo.ID, password)
	if err != nil {
		return err
	}

	for _, acc := range kp.Accounts {
		err = api.createKeystoreFileForAccount(kp.DerivedFrom, password, acc)
		if err != nil {
			return err
		}
	}

	// this will emit SyncKeypair message
	return (*api.messenger).DeleteAllKeycardsWithKeyUID(ctx, generatedAccountInfo.KeyUID)
}

// If keypair is migrated from keycard to app, then `accountsComingFromKeycard` should be set to true, otherwise false.
// If keycard is new `Position` will be determined and set by the backend and `KeycardLocked` will be set to false.
// If keycard is already added, `Position` and `KeycardLocked` will be unchanged.
func (api *API) SaveOrUpdateKeycard(ctx context.Context, keycard *accounts.Keycard, accountsComingFromKeycard bool) error {
	if len(keycard.AccountsAddresses) == 0 {
		return errors.New("cannot migrate a keypair without accounts")
	}

	kpDb, err := api.db.GetKeypairByKeyUID(keycard.KeyUID)
	if err != nil {
		if err == accounts.ErrDbKeypairNotFound {
			return errors.New("cannot migrate an unknown keypair")
		}
		return err
	}

	err = (*api.messenger).SaveOrUpdateKeycard(ctx, keycard)
	if err != nil {
		return err
	}

	if !accountsComingFromKeycard {
		// Once we migrate a keypair, corresponding keystore files need to be deleted
		// if the keypair being migrated is not already migrated (in case user is creating a copy of an existing Keycard)
		// and if keypair operability is different from non operable (otherwise there are not keystore files to be deleted).
		if !kpDb.MigratedToKeycard() && kpDb.Operability() != accounts.AccountNonOperable {
			for _, acc := range kpDb.Accounts {
				if acc.Operable != accounts.AccountFullyOperable {
					continue
				}
				err = api.manager.DeleteAccount(acc.Address)
				if err != nil {
					return err
				}
			}

			err = api.manager.DeleteAccount(types.Address(common.HexToAddress(kpDb.DerivedFrom)))
			if err != nil {
				return err
			}
		}

		err = (*api.messenger).MarkKeypairFullyOperable(keycard.KeyUID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (api *API) GetAllKnownKeycards(ctx context.Context) ([]*accounts.Keycard, error) {
	return api.db.GetAllKnownKeycards()
}

func (api *API) GetKeycardsWithSameKeyUID(ctx context.Context, keyUID string) ([]*accounts.Keycard, error) {
	return api.db.GetKeycardsWithSameKeyUID(keyUID)
}

func (api *API) GetKeycardByKeycardUID(ctx context.Context, keycardUID string) (*accounts.Keycard, error) {
	return api.db.GetKeycardByKeycardUID(keycardUID)
}

func (api *API) SetKeycardName(ctx context.Context, keycardUID string, kpName string) error {
	return (*api.messenger).SetKeycardName(ctx, keycardUID, kpName)
}

func (api *API) KeycardLocked(ctx context.Context, keycardUID string) error {
	return (*api.messenger).KeycardLocked(ctx, keycardUID)
}

func (api *API) KeycardUnlocked(ctx context.Context, keycardUID string) error {
	return (*api.messenger).KeycardUnlocked(ctx, keycardUID)
}

func (api *API) DeleteKeycardAccounts(ctx context.Context, keycardUID string, accountAddresses []types.Address) error {
	return (*api.messenger).DeleteKeycardAccounts(ctx, keycardUID, accountAddresses)
}

func (api *API) DeleteKeycard(ctx context.Context, keycardUID string) error {
	return (*api.messenger).DeleteKeycard(ctx, keycardUID)
}

func (api *API) DeleteAllKeycardsWithKeyUID(ctx context.Context, keyUID string) error {
	return (*api.messenger).DeleteAllKeycardsWithKeyUID(ctx, keyUID)
}

func (api *API) UpdateKeycardUID(ctx context.Context, oldKeycardUID string, newKeycardUID string) error {
	return (*api.messenger).UpdateKeycardUID(ctx, oldKeycardUID, newKeycardUID)
}

func (api *API) AddressWasShown(address types.Address) error {
	return api.db.AddressWasShown(address)
}
