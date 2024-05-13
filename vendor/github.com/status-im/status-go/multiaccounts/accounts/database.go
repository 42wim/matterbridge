package accounts

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/multiaccounts/common"
	"github.com/status-im/status-go/multiaccounts/settings"
	notificationssettings "github.com/status-im/status-go/multiaccounts/settings_notifications"
	sociallinkssettings "github.com/status-im/status-go/multiaccounts/settings_social_links"
	walletsettings "github.com/status-im/status-go/multiaccounts/settings_wallet"
	"github.com/status-im/status-go/nodecfg"
	"github.com/status-im/status-go/params"
)

const (
	statusChatPath           = "m/43'/60'/1581'/0'/0"
	statusWalletRootPath     = "m/44'/60'/0'/0/"
	zeroAddress              = "0x0000000000000000000000000000000000000000"
	SyncedFromBackup         = "backup" // means a keypair is coming from backed up data
	ThirtyDaysInMilliseconds = 30 * 24 * 60 * 60 * 1000
)

var (
	errDbPassedParameterIsNil                      = errors.New("accounts: passed parameter is nil")
	errDbTransactionIsNil                          = errors.New("accounts: database transaction is nil")
	ErrDbKeypairNotFound                           = errors.New("accounts: keypair is not found")
	ErrCannotRemoveProfileKeypair                  = errors.New("accounts: cannot remove profile keypair")
	ErrDbAccountNotFound                           = errors.New("accounts: account is not found")
	ErrCannotRemoveProfileAccount                  = errors.New("accounts: cannot remove profile account")
	ErrCannotRemoveDefaultWalletAccount            = errors.New("accounts: cannot remove default wallet account")
	ErrAccountWrongPosition                        = errors.New("accounts: trying to set wrong position to account")
	ErrNotTheSameNumberOdAccountsToApplyReordering = errors.New("accounts: there is different number of accounts between received sync message and db accounts")
	ErrNotTheSameAccountsToApplyReordering         = errors.New("accounts: there are differences between accounts in received sync message and db accounts")
	ErrMovingAccountToWrongPosition                = errors.New("accounts: trying to move account to a wrong position")
	ErrKeypairDifferentAccountsKeyUID              = errors.New("cannot store keypair with different accounts' key uid than keypair's key uid")
	ErrKeypairWithoutAccounts                      = errors.New("cannot store keypair without accounts")
)

type Keypair struct {
	KeyUID                  string      `json:"key-uid"`
	Name                    string      `json:"name"`
	Type                    KeypairType `json:"type"`
	DerivedFrom             string      `json:"derived-from"`
	LastUsedDerivationIndex uint64      `json:"last-used-derivation-index,omitempty"`
	SyncedFrom              string      `json:"synced-from,omitempty"` // keeps an info which device this keypair is added from can be one of two values defined in constants or device name (custom)
	Clock                   uint64      `json:"clock,omitempty"`
	Accounts                []*Account  `json:"accounts,omitempty"`
	Keycards                []*Keycard  `json:"keycards,omitempty"`
	Removed                 bool        `json:"removed,omitempty"`
}

type Account struct {
	Address               types.Address             `json:"address"`
	KeyUID                string                    `json:"key-uid"`
	Wallet                bool                      `json:"wallet"`
	AddressWasNotShown    bool                      `json:"address-was-not-shown,omitempty"`
	Chat                  bool                      `json:"chat"`
	Type                  AccountType               `json:"type,omitempty"`
	Path                  string                    `json:"path,omitempty"`
	PublicKey             types.HexBytes            `json:"public-key,omitempty"`
	Name                  string                    `json:"name"`
	Emoji                 string                    `json:"emoji"`
	ColorID               common.CustomizationColor `json:"colorId,omitempty"`
	Hidden                bool                      `json:"hidden"`
	Clock                 uint64                    `json:"clock,omitempty"`
	Removed               bool                      `json:"removed,omitempty"`
	Operable              AccountOperable           `json:"operable"` // describes an account's operability (check AccountOperable type constants for details)
	CreatedAt             int64                     `json:"createdAt"`
	Position              int64                     `json:"position"`
	ProdPreferredChainIDs string                    `json:"prodPreferredChainIds"`
	TestPreferredChainIDs string                    `json:"testPreferredChainIds"`
}

type KeypairType string
type AccountType string
type AccountOperable string

func (a KeypairType) String() string {
	return string(a)
}

func (a AccountType) String() string {
	return string(a)
}

func (a AccountOperable) String() string {
	return string(a)
}

const (
	KeypairTypeProfile KeypairType = "profile"
	KeypairTypeKey     KeypairType = "key"
	KeypairTypeSeed    KeypairType = "seed"
)

const (
	AccountTypeGenerated AccountType = "generated"
	AccountTypeKey       AccountType = "key"
	AccountTypeSeed      AccountType = "seed"
	AccountTypeWatch     AccountType = "watch"
)

const (
	AccountNonOperable       AccountOperable = "no"        // an account is non operable it is not a keycard account and there is no keystore file for it and no keystore file for the address it is derived from
	AccountPartiallyOperable AccountOperable = "partially" // an account is partially operable if it is not a keycard account and there is created keystore file for the address it is derived from
	AccountFullyOperable     AccountOperable = "fully"     // an account is fully operable if it is not a keycard account and there is a keystore file for it

	ProdPreferredChainIDsDefault        = "1:10:42161"
	TestPreferredChainIDsDefault        = "5:420:421613"
	TestSepoliaPreferredChainIDsDefault = "11155111:11155420:421614"
)

// Returns true if an account is a wallet account that logged in user has a control over, otherwise returns false.
func (a *Account) IsWalletNonWatchOnlyAccount() bool {
	return !a.Chat && len(a.Type) > 0 && a.Type != AccountTypeWatch
}

// Returns true if an account is a wallet account that is ready for sending transactions, otherwise returns false.
func (a *Account) IsWalletAccountReadyForTransaction() bool {
	return a.IsWalletNonWatchOnlyAccount() && a.Operable != AccountNonOperable
}

func (a *Account) MarshalJSON() ([]byte, error) {
	item := struct {
		Address               types.Address             `json:"address"`
		MixedcaseAddress      string                    `json:"mixedcase-address"`
		KeyUID                string                    `json:"key-uid"`
		Wallet                bool                      `json:"wallet"`
		Chat                  bool                      `json:"chat"`
		Type                  AccountType               `json:"type"`
		Path                  string                    `json:"path"`
		PublicKey             types.HexBytes            `json:"public-key"`
		Name                  string                    `json:"name"`
		Emoji                 string                    `json:"emoji"`
		ColorID               common.CustomizationColor `json:"colorId"`
		Hidden                bool                      `json:"hidden"`
		Clock                 uint64                    `json:"clock"`
		Removed               bool                      `json:"removed"`
		Operable              AccountOperable           `json:"operable"`
		CreatedAt             int64                     `json:"createdAt"`
		Position              int64                     `json:"position"`
		ProdPreferredChainIDs string                    `json:"prodPreferredChainIds"`
		TestPreferredChainIDs string                    `json:"testPreferredChainIds"`
	}{
		Address:               a.Address,
		MixedcaseAddress:      a.Address.Hex(),
		KeyUID:                a.KeyUID,
		Wallet:                a.Wallet,
		Chat:                  a.Chat,
		Type:                  a.Type,
		Path:                  a.Path,
		PublicKey:             a.PublicKey,
		Name:                  a.Name,
		Emoji:                 a.Emoji,
		ColorID:               a.ColorID,
		Hidden:                a.Hidden,
		Clock:                 a.Clock,
		Removed:               a.Removed,
		Operable:              a.Operable,
		CreatedAt:             a.CreatedAt,
		Position:              a.Position,
		ProdPreferredChainIDs: a.ProdPreferredChainIDs,
		TestPreferredChainIDs: a.TestPreferredChainIDs,
	}

	return json.Marshal(item)
}

func (a *Keypair) MarshalJSON() ([]byte, error) {
	item := struct {
		KeyUID                  string      `json:"key-uid"`
		Name                    string      `json:"name"`
		Type                    KeypairType `json:"type"`
		DerivedFrom             string      `json:"derived-from"`
		LastUsedDerivationIndex uint64      `json:"last-used-derivation-index"`
		SyncedFrom              string      `json:"synced-from"`
		Clock                   uint64      `json:"clock"`
		Accounts                []*Account  `json:"accounts"`
		Keycards                []*Keycard  `json:"keycards"`
		Removed                 bool        `json:"removed"`
	}{
		KeyUID:                  a.KeyUID,
		Name:                    a.Name,
		Type:                    a.Type,
		DerivedFrom:             a.DerivedFrom,
		LastUsedDerivationIndex: a.LastUsedDerivationIndex,
		SyncedFrom:              a.SyncedFrom,
		Clock:                   a.Clock,
		Accounts:                a.Accounts,
		Keycards:                a.Keycards,
		Removed:                 a.Removed,
	}

	return json.Marshal(item)
}

func (a *Keypair) CopyKeypair() *Keypair {
	kp := &Keypair{
		Clock:                   a.Clock,
		KeyUID:                  a.KeyUID,
		Name:                    a.Name,
		Type:                    a.Type,
		DerivedFrom:             a.DerivedFrom,
		LastUsedDerivationIndex: a.LastUsedDerivationIndex,
		SyncedFrom:              a.SyncedFrom,
		Accounts:                make([]*Account, len(a.Accounts)),
		Keycards:                make([]*Keycard, len(a.Keycards)),
		Removed:                 a.Removed,
	}

	for i, acc := range a.Accounts {
		kp.Accounts[i] = &Account{
			Address:               acc.Address,
			KeyUID:                acc.KeyUID,
			Wallet:                acc.Wallet,
			Chat:                  acc.Chat,
			Type:                  acc.Type,
			Path:                  acc.Path,
			PublicKey:             acc.PublicKey,
			Name:                  acc.Name,
			Emoji:                 acc.Emoji,
			ColorID:               acc.ColorID,
			Hidden:                acc.Hidden,
			Clock:                 acc.Clock,
			Removed:               acc.Removed,
			Operable:              acc.Operable,
			CreatedAt:             acc.CreatedAt,
			Position:              acc.Position,
			ProdPreferredChainIDs: acc.ProdPreferredChainIDs,
			TestPreferredChainIDs: acc.TestPreferredChainIDs,
		}
	}

	for i, kc := range a.Keycards {
		kp.Keycards[i] = &Keycard{
			KeycardUID:        kc.KeycardUID,
			KeycardName:       kc.KeycardName,
			KeycardLocked:     kc.KeycardLocked,
			AccountsAddresses: kc.AccountsAddresses,
			KeyUID:            kc.KeyUID,
		}
	}

	return kp
}

func (a *Keypair) GetChatPublicKey() types.HexBytes {
	for _, acc := range a.Accounts {
		if acc.Chat {
			return acc.PublicKey
		}
	}

	return nil
}

func (a *Keypair) MigratedToKeycard() bool {
	return len(a.Keycards) > 0
}

// Returns operability of a keypair:
// - if any of keypair's account is not operable, then a keyapir is considered as non operable
// - if any of keypair's account is partially operable, then a keyapir is considered as partially operable
// - if all accounts are fully operable, then a keyapir is considered as fully operable
func (a *Keypair) Operability() AccountOperable {
	for _, acc := range a.Accounts {
		if acc.Operable == AccountNonOperable {
			return AccountNonOperable
		}
		if acc.Operable == AccountPartiallyOperable {
			return AccountPartiallyOperable
		}
	}

	return AccountFullyOperable
}

// Database sql wrapper for operations with browser objects.
type Database struct {
	settings.DatabaseSettingsManager
	*notificationssettings.NotificationsSettings
	*sociallinkssettings.SocialLinksSettings
	*walletsettings.WalletSettings
	db *sql.DB
}

// NewDB returns a new instance of *Database
func NewDB(db *sql.DB) (*Database, error) {
	sDB, err := settings.MakeNewDB(db)
	if err != nil {
		return nil, err
	}
	sn := notificationssettings.NewNotificationsSettings(db)
	ssl := sociallinkssettings.NewSocialLinksSettings(db)
	sw := walletsettings.NewWalletSettings(db)

	return &Database{sDB, sn, ssl, sw, db}, nil
}

// DB Gets db sql.DB
func (db *Database) DB() *sql.DB {
	return db.db
}

// Close closes database.
func (db *Database) Close() error {
	return db.db.Close()
}

func GetAccountTypeForKeypairType(kpType KeypairType) AccountType {
	switch kpType {
	case KeypairTypeProfile:
		return AccountTypeGenerated
	case KeypairTypeKey:
		return AccountTypeKey
	case KeypairTypeSeed:
		return AccountTypeSeed
	default:
		return AccountTypeWatch
	}
}

func (db *Database) processRows(rows *sql.Rows) ([]*Keypair, []*Account, error) {
	keypairMap := make(map[string]*Keypair)
	allAccounts := []*Account{}

	var (
		kpKeyUID                  sql.NullString
		kpName                    sql.NullString
		kpType                    sql.NullString
		kpDerivedFrom             sql.NullString
		kpLastUsedDerivationIndex sql.NullInt64
		kpSyncedFrom              sql.NullString
		kpClock                   sql.NullInt64
		kpRemoved                 sql.NullBool
	)

	var (
		accAddress               sql.NullString
		accKeyUID                sql.NullString
		accPath                  sql.NullString
		accName                  sql.NullString
		accColorID               sql.NullString
		accEmoji                 sql.NullString
		accWallet                sql.NullBool
		accChat                  sql.NullBool
		accHidden                sql.NullBool
		accOperable              sql.NullString
		accClock                 sql.NullInt64
		accCreatedAt             sql.NullTime
		accPosition              sql.NullInt64
		accRemoved               sql.NullBool
		accProdPreferredChainIDs sql.NullString
		accTestPreferredChainIDs sql.NullString
		accAddressWasNotShown    sql.NullBool
	)

	for rows.Next() {
		kp := &Keypair{}
		acc := &Account{}
		pubkey := []byte{}
		err := rows.Scan(
			&kpKeyUID, &kpName, &kpType, &kpDerivedFrom, &kpLastUsedDerivationIndex, &kpSyncedFrom, &kpClock, &kpRemoved,
			&accAddress, &accKeyUID, &pubkey, &accPath, &accName, &accColorID, &accEmoji,
			&accWallet, &accChat, &accHidden, &accOperable, &accClock, &accCreatedAt, &accPosition, &accRemoved,
			&accProdPreferredChainIDs, &accTestPreferredChainIDs, &accAddressWasNotShown)
		if err != nil {
			return nil, nil, err
		}

		// check keypair fields
		if kpKeyUID.Valid {
			kp.KeyUID = kpKeyUID.String
		}
		if kpName.Valid {
			kp.Name = kpName.String
		}
		if kpType.Valid {
			kp.Type = KeypairType(kpType.String)
		}
		if kpDerivedFrom.Valid {
			kp.DerivedFrom = kpDerivedFrom.String
		}
		if kpLastUsedDerivationIndex.Valid {
			kp.LastUsedDerivationIndex = uint64(kpLastUsedDerivationIndex.Int64)
		}
		if kpSyncedFrom.Valid {
			kp.SyncedFrom = kpSyncedFrom.String
		}
		if kpClock.Valid {
			kp.Clock = uint64(kpClock.Int64)
		}
		if kpRemoved.Valid {
			kp.Removed = kpRemoved.Bool
		}
		// check keypair accounts fields
		if accAddress.Valid {
			acc.Address = types.BytesToAddress([]byte(accAddress.String))
		}
		if accKeyUID.Valid {
			acc.KeyUID = accKeyUID.String
		}
		if accPath.Valid {
			acc.Path = accPath.String
		}
		if accName.Valid {
			acc.Name = accName.String
		}
		if accColorID.Valid {
			acc.ColorID = common.CustomizationColor(accColorID.String)
		}
		if accEmoji.Valid {
			acc.Emoji = accEmoji.String
		}
		if accWallet.Valid {
			acc.Wallet = accWallet.Bool
		}
		if accChat.Valid {
			acc.Chat = accChat.Bool
		}
		if accHidden.Valid {
			acc.Hidden = accHidden.Bool
		}
		if accOperable.Valid {
			acc.Operable = AccountOperable(accOperable.String)
		}
		if accClock.Valid {
			acc.Clock = uint64(accClock.Int64)
		}
		if accCreatedAt.Valid {
			acc.CreatedAt = accCreatedAt.Time.UnixMilli()
		}
		if accPosition.Valid {
			acc.Position = accPosition.Int64
		}
		if accProdPreferredChainIDs.Valid {
			acc.ProdPreferredChainIDs = accProdPreferredChainIDs.String
		}
		if accTestPreferredChainIDs.Valid {
			acc.TestPreferredChainIDs = accTestPreferredChainIDs.String
		}
		if accAddressWasNotShown.Valid {
			acc.AddressWasNotShown = accAddressWasNotShown.Bool
		}
		if lth := len(pubkey); lth > 0 {
			acc.PublicKey = make(types.HexBytes, lth)
			copy(acc.PublicKey, pubkey)
		}
		if accRemoved.Valid {
			acc.Removed = accRemoved.Bool
		}
		acc.Type = GetAccountTypeForKeypairType(kp.Type)

		if kp.KeyUID != "" {
			if _, ok := keypairMap[kp.KeyUID]; !ok {
				keypairMap[kp.KeyUID] = kp
			}
			keypairMap[kp.KeyUID].Accounts = append(keypairMap[kp.KeyUID].Accounts, acc)
		}
		allAccounts = append(allAccounts, acc)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// Convert map to list
	keypairs := make([]*Keypair, 0, len(keypairMap))
	for _, keypair := range keypairMap {
		keypairs = append(keypairs, keypair)
	}

	return keypairs, allAccounts, nil
}

// If `includeRemoved` is false and `keyUID` is not empty, then keypairs which are not flagged as removed and match the `keyUID` will be returned.
// If `includeRemoved` is true and `keyUID` is not empty, then keypairs which match the `keyUID` will be returned (regardless how they are flagged).
// If `includeRemoved` is false and `keyUID` is empty, then all keypairs which are not flagged as removed will be returned.
// If `includeRemoved` is true and `keyUID` is empty, then all keypairs will be returned (regardless how they are flagged).
func (db *Database) getKeypairs(tx *sql.Tx, keyUID string, includeRemoved bool) ([]*Keypair, error) {
	var (
		rows           *sql.Rows
		err            error
		mainQueryWhere string
		subQueryWhere  string
	)
	if tx == nil {
		tx, err = db.db.Begin()
		if err != nil {
			return nil, err
		}
		defer func() {
			if err == nil {
				err = tx.Commit()
				return
			}
			_ = tx.Rollback()
		}()
	}

	if keyUID != "" {
		mainQueryWhere = "WHERE k.key_uid = ?"
		if !includeRemoved {
			mainQueryWhere += " AND k.removed = 0"
		}
	} else if !includeRemoved {
		mainQueryWhere = "WHERE k.removed = 0"
	}

	if !includeRemoved {
		subQueryWhere = "WHERE removed = 0"
	}

	query := fmt.Sprintf( // nolint: gosec
		`
		SELECT
			k.*,
			ka.address,
			ka.key_uid,
			ka.pubkey,
			ka.path,
			ka.name,
			ka.color,
			ka.emoji,
			ka.wallet,
			ka.chat,
			ka.hidden,
			ka.operable,
			ka.clock,
			ka.created_at,
			ka.position,
			ka.removed,
			ka.prod_preferred_chain_ids,
			ka.test_preferred_chain_ids,
                        ka.address_was_not_shown
		FROM
			keypairs k
		LEFT JOIN
			(
				SELECT *
				FROM
					keypairs_accounts
				%s
			) AS ka
		ON
			k.key_uid = ka.key_uid
		%s
		ORDER BY
			ka.position`, subQueryWhere, mainQueryWhere)

	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	if keyUID != "" {
		rows, err = stmt.Query(keyUID)
	} else {
		rows, err = stmt.Query()
	}
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	keypairs, _, err := db.processRows(rows)
	if err != nil {
		return nil, err
	}

	for _, kp := range keypairs {
		keycards, err := db.getKeycards(tx, kp.KeyUID, "")
		if err != nil {
			return nil, err
		}

		kp.Keycards = keycards
	}

	return keypairs, nil
}

func (db *Database) getKeypairByKeyUID(tx *sql.Tx, keyUID string, includeRemoved bool) (*Keypair, error) {
	keypairs, err := db.getKeypairs(tx, keyUID, includeRemoved)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if len(keypairs) == 0 {
		return nil, ErrDbKeypairNotFound
	}
	return keypairs[0], nil
}

// If `includeRemoved` is false and `address` is not zero address, then accounts which are not flagged as removed and match the `address` will be returned.
// If `includeRemoved` is true and `address` is not zero address, then accounts which match the `address` will be returned (regardless how they are flagged).
// If `includeRemoved` is false and `address` is zero address, then all accounts which are not flagged as removed will be returned.
// If `includeRemoved` is true and `address` is zero address, then all accounts will be returned (regardless how they are flagged).
func (db *Database) getAccounts(tx *sql.Tx, address types.Address, includeRemoved bool) ([]*Account, error) {
	var (
		rows  *sql.Rows
		err   error
		where string
	)
	filterByAddress := address.String() != zeroAddress
	if filterByAddress {
		where = "WHERE ka.address = ?"
		if !includeRemoved {
			where += " AND ka.removed = 0"
		}
	} else if !includeRemoved {
		where = "WHERE ka.removed = 0"
	}

	query := fmt.Sprintf( // nolint: gosec
		`
		SELECT
			k.*,
			ka.address,
			ka.key_uid,
			ka.pubkey,
			ka.path,
			ka.name,
			ka.color,
			ka.emoji,
			ka.wallet,
			ka.chat,
			ka.hidden,
			ka.operable,
			ka.clock,
			ka.created_at,
			ka.position,
			ka.removed,
			ka.prod_preferred_chain_ids,
			ka.test_preferred_chain_ids,
			ka.address_was_not_shown
		FROM
			keypairs_accounts ka
		LEFT JOIN
			keypairs k
		ON
			ka.key_uid = k.key_uid
		%s
		ORDER BY
			ka.position`, where)

	if tx == nil {
		if filterByAddress {
			rows, err = db.db.Query(query, address)
		} else {
			rows, err = db.db.Query(query)
		}
		if err != nil {
			return nil, err
		}
	} else {
		stmt, err := tx.Prepare(query)
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		if filterByAddress {
			rows, err = stmt.Query(address)
		} else {
			rows, err = stmt.Query()
		}
		if err != nil {
			return nil, err
		}
	}

	defer rows.Close()
	_, allAccounts, err := db.processRows(rows)
	if err != nil {
		return nil, err
	}

	return allAccounts, nil
}

func (db *Database) getAccountByAddress(tx *sql.Tx, address types.Address) (*Account, error) {
	accounts, err := db.getAccounts(tx, address, false)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if len(accounts) == 0 {
		return nil, ErrDbAccountNotFound
	}

	return accounts[0], nil
}

func (db *Database) markAccountRemoved(tx *sql.Tx, address types.Address, clock uint64) error {
	if tx == nil {
		return errDbTransactionIsNil
	}

	_, err := db.getAccountByAddress(tx, address)
	if err != nil {
		return err
	}

	query, err := tx.Prepare(`
		UPDATE
			keypairs_accounts
		SET
			removed = 1,
			clock = ?
		WHERE
			address = ?
	`)
	if err != nil {
		return err
	}
	defer query.Close()

	_, err = query.Exec(clock, address)
	return err
}

// Marking keypair as removed, will delete related keycards.
func (db *Database) markKeypairRemoved(tx *sql.Tx, keyUID string, clock uint64) error {
	if tx == nil {
		return errDbTransactionIsNil
	}

	keypair, err := db.getKeypairByKeyUID(tx, keyUID, false)
	if err != nil {
		return err
	}

	for _, acc := range keypair.Accounts {
		if acc.Removed {
			continue
		}
		err = db.markAccountRemoved(tx, acc.Address, clock)
		if err != nil {
			return err
		}
	}

	query := `
		UPDATE
			keypairs
		SET
			removed = 1,
			clock = ?
		WHERE
			key_uid = ?
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(clock, keyUID)
	if err != nil {
		return err
	}

	err = db.deleteAllKeycardsWithKeyUID(tx, keyUID)
	return err
}

// Returns active keypairs (excluding removed and excluding removed accounts).
func (db *Database) GetActiveKeypairs() ([]*Keypair, error) {
	return db.getKeypairs(nil, "", false)
}

// Returns all keypairs (including removed and removed accounts).
func (db *Database) GetAllKeypairs() ([]*Keypair, error) {
	return db.getKeypairs(nil, "", true)
}

// Returns keypair if it is not marked as removed and its accounts which are not marked as removed.
func (db *Database) GetKeypairByKeyUID(keyUID string) (*Keypair, error) {
	return db.getKeypairByKeyUID(nil, keyUID, false)
}

// Returns active accounts (excluding removed).
func (db *Database) GetActiveAccounts() ([]*Account, error) {
	return db.getAccounts(nil, types.Address{}, false)
}

// Returns all accounts (including removed).
func (db *Database) GetAllAccounts() ([]*Account, error) {
	return db.getAccounts(nil, types.Address{}, true)
}

// Returns account if it is not marked as removed.
func (db *Database) GetAccountByAddress(address types.Address) (*Account, error) {
	return db.getAccountByAddress(nil, address)
}

// Returns active watch only accounts (excluding removed).
func (db *Database) GetActiveWatchOnlyAccounts() (res []*Account, err error) {
	accounts, err := db.getAccounts(nil, types.Address{}, false)
	if err != nil {
		return nil, err
	}
	for _, acc := range accounts {
		if acc.Type == AccountTypeWatch {
			res = append(res, acc)
		}
	}
	return
}

// Returns all watch only accounts (including removed).
func (db *Database) GetAllWatchOnlyAccounts() (res []*Account, err error) {
	accounts, err := db.getAccounts(nil, types.Address{}, true)
	if err != nil {
		return nil, err
	}
	for _, acc := range accounts {
		if acc.Type == AccountTypeWatch {
			res = append(res, acc)
		}
	}
	return
}

func (db *Database) IsAnyAccountPartiallyOrFullyOperableForKeyUID(keyUID string) (bool, error) {
	kp, err := db.getKeypairByKeyUID(nil, keyUID, false)
	if err != nil {
		return false, err
	}

	for _, acc := range kp.Accounts {
		if acc.Operable != AccountNonOperable {
			return true, nil
		}
	}
	return false, nil
}

func (db *Database) RemoveKeypair(keyUID string, clock uint64) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	return db.markKeypairRemoved(tx, keyUID, clock)
}

func (db *Database) RemoveAccount(address types.Address, clock uint64) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	acc, err := db.getAccountByAddress(tx, address)
	if err != nil {
		return err
	}

	kp, err := db.getKeypairByKeyUID(tx, acc.KeyUID, false)
	if err != nil && err != ErrDbKeypairNotFound {
		return err
	}

	if kp != nil {
		lastAccOfKepairToBeRemoved := true
		for _, kpAcc := range kp.Accounts {
			if !kpAcc.Removed && kpAcc.Address != address {
				lastAccOfKepairToBeRemoved = false
			}
		}

		if lastAccOfKepairToBeRemoved {
			return db.markKeypairRemoved(tx, acc.KeyUID, clock)
		}
	}

	err = db.markAccountRemoved(tx, address, clock)
	if err != nil {
		return err
	}

	// Update keypair clock if any but the watch only account was deleted.
	if kp != nil {
		err = db.updateKeypairClock(tx, acc.KeyUID, clock)
		return err
	}

	return nil
}

func updateKeypairLastUsedIndex(tx *sql.Tx, keyUID string, index uint64, clock uint64, updateKeypairClock bool) error {
	if tx == nil {
		return errDbTransactionIsNil
	}
	var (
		err      error
		setClock string
	)
	if updateKeypairClock {
		setClock = ", clock = ?"
	}

	query := fmt.Sprintf( // nolint: gosec
		`
		UPDATE
				keypairs
			SET
				last_used_derivation_index = ?
				%s
			WHERE
				key_uid = ?`, setClock)

	if setClock != "" {
		_, err = tx.Exec(query, index, clock, keyUID)
	} else {
		_, err = tx.Exec(query, index, keyUID)
	}

	return err
}

func (db *Database) updateKeypairClock(tx *sql.Tx, keyUID string, clock uint64) error {
	if tx == nil {
		return errDbTransactionIsNil
	}

	_, err := tx.Exec(`
			UPDATE
				keypairs
			SET
				clock = ?
			WHERE
				key_uid = ?`,
		clock, keyUID)

	return err
}

func (db *Database) saveOrUpdateAccounts(tx *sql.Tx, accounts []*Account, updateKeypairClock, isGoerliEnabled bool) (err error) {
	if tx == nil {
		return errDbTransactionIsNil
	}

	for _, acc := range accounts {
		var relatedKeypair *Keypair
		// only watch only accounts have an empty `KeyUID` field
		var keyUID *string
		if acc.KeyUID != "" {
			relatedKeypair, err = db.getKeypairByKeyUID(tx, acc.KeyUID, true)
			if err != nil {
				if err == sql.ErrNoRows {
					// all accounts, except watch only accounts, must have a row in `keypairs` table with the same key uid
					continue
				}
				return err
			}
			keyUID = &acc.KeyUID
		}
		var exists bool
		err = tx.QueryRow("SELECT EXISTS (SELECT 1 FROM keypairs_accounts WHERE address = ? AND removed = 0)", acc.Address).Scan(&exists)
		if err != nil {
			return err
		}

		// Apply default values if account is new and not a watch only
		if !exists && acc.Type != AccountTypeWatch {
			if acc.ProdPreferredChainIDs == "" {
				acc.ProdPreferredChainIDs = ProdPreferredChainIDsDefault
			}

			if acc.TestPreferredChainIDs == "" {
				if isGoerliEnabled {
					acc.TestPreferredChainIDs = TestPreferredChainIDsDefault
				} else {
					acc.TestPreferredChainIDs = TestSepoliaPreferredChainIDsDefault
				}
			}
		}

		_, err = tx.Exec(`
			INSERT OR IGNORE INTO
				keypairs_accounts (address, key_uid, pubkey, path, wallet, address_was_not_shown, chat, created_at, updated_at)
			VALUES
				(?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'));

			UPDATE
				keypairs_accounts
			SET
				name = ?,
				color = ?,
				emoji = ?,
				hidden = ?,
				operable = ?,
				clock = ?,
				position = ?,
				updated_at = datetime('now'),
				removed = ?,
				prod_preferred_chain_ids = ?,
				test_preferred_chain_ids = ?
			WHERE
				address = ?;
		`,
			acc.Address, keyUID, acc.PublicKey, acc.Path, acc.Wallet, acc.AddressWasNotShown, acc.Chat,
			acc.Name, acc.ColorID, acc.Emoji, acc.Hidden, acc.Operable, acc.Clock, acc.Position, acc.Removed,
			acc.ProdPreferredChainIDs, acc.TestPreferredChainIDs, acc.Address)

		if err != nil {
			return err
		}

		// Update positions change clock when adding new/updating account
		err = db.setClockOfLastAccountsPositionChange(tx, acc.Clock)
		if err != nil {
			return err
		}

		// Update keypair clock if any but the watch only account has changed.
		if relatedKeypair != nil && updateKeypairClock {
			err = db.updateKeypairClock(tx, acc.KeyUID, acc.Clock)
			if err != nil {
				return err
			}
		}

		if !acc.Removed && strings.HasPrefix(acc.Path, statusWalletRootPath) {
			accIndex, err := strconv.ParseUint(acc.Path[len(statusWalletRootPath):], 0, 64)
			if err != nil {
				return err
			}

			accountsContainPath := func(accounts []*Account, path string) bool {
				for _, acc := range accounts {
					if acc.Path == path {
						return true
					}
				}
				return false
			}

			expectedNewKeypairIndex := uint64(0)
			if relatedKeypair != nil {
				expectedNewKeypairIndex = relatedKeypair.LastUsedDerivationIndex
				for {
					expectedNewKeypairIndex++
					if !accountsContainPath(relatedKeypair.Accounts, statusWalletRootPath+strconv.FormatUint(expectedNewKeypairIndex, 10)) {
						break
					}
				}
			}

			if accIndex == expectedNewKeypairIndex {
				err = updateKeypairLastUsedIndex(tx, acc.KeyUID, accIndex, acc.Clock, updateKeypairClock)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Saves accounts, if an account already exists, it will be updated.
func (db *Database) SaveOrUpdateAccounts(accounts []*Account, updateKeypairClock bool) error {
	if len(accounts) == 0 {
		return errors.New("no provided accounts to save/update")
	}
	isGoerliEnabled, err := db.GetIsGoerliEnabled()
	if err != nil {
		return err
	}

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()
	err = db.saveOrUpdateAccounts(tx, accounts, updateKeypairClock, isGoerliEnabled)
	return err
}

// Saves a keypair and its accounts, if a keypair with `key_uid` already exists, it will be updated,
// if any of its accounts exists it will be updated as well, otherwise it will be added.
// Since keypair type contains `Keycards` as well, they are excluded from the saving/updating this way regardless they
// are set or not.
func (db *Database) SaveOrUpdateKeypair(keypair *Keypair) error {
	if keypair == nil {
		return errDbPassedParameterIsNil
	}

	isGoerliEnabled, err := db.GetIsGoerliEnabled()
	if err != nil {
		return err
	}

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	// If keypair is being saved, not updated, then it must be at least one account and all accounts must have the same key uid.
	dbKeypair, err := db.getKeypairByKeyUID(tx, keypair.KeyUID, true)
	if err != nil && err != ErrDbKeypairNotFound {
		return err
	}
	if dbKeypair == nil {
		if len(keypair.Accounts) == 0 {
			return ErrKeypairWithoutAccounts
		}
		for _, acc := range keypair.Accounts {
			if acc.KeyUID == "" || acc.KeyUID != keypair.KeyUID {
				return ErrKeypairDifferentAccountsKeyUID
			}
		}
	}

	_, err = tx.Exec(`
		INSERT OR IGNORE INTO
			keypairs (key_uid, type, derived_from)
		VALUES
			(?, ?, ?);

		UPDATE
			keypairs
		SET
			name = ?,
			last_used_derivation_index = ?,
			synced_from = ?,
			clock = ?,
			removed = ?
		WHERE
			key_uid = ?;
	`, keypair.KeyUID, keypair.Type, keypair.DerivedFrom,
		keypair.Name, keypair.LastUsedDerivationIndex, keypair.SyncedFrom, keypair.Clock, keypair.Removed, keypair.KeyUID)
	if err != nil {
		return err
	}
	return db.saveOrUpdateAccounts(tx, keypair.Accounts, false, isGoerliEnabled)
}

func (db *Database) UpdateKeypairName(keyUID string, name string, clock uint64, updateChatAccountName bool) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	_, err = db.getKeypairByKeyUID(tx, keyUID, false)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE
			keypairs
		SET
			name = ?,
			clock = ?
		WHERE
			key_uid = ?;
	`, name, clock, keyUID)
	if err != nil {
		return err
	}

	if updateChatAccountName {
		_, err = tx.Exec(`
			UPDATE
				keypairs_accounts
			SET
				name = ?,
				clock = ?
			WHERE
				key_uid = ?
			AND
				path = ?;
		`, name, clock, keyUID, statusChatPath)
		return err
	}

	return nil
}

func (db *Database) GetWalletAddress() (rst types.Address, err error) {
	err = db.db.QueryRow("SELECT address FROM keypairs_accounts WHERE wallet = 1").Scan(&rst)
	return
}

func (db *Database) GetProfileKeypair() (*Keypair, error) {
	keypairs, err := db.getKeypairs(nil, "", false)
	if err != nil {
		return nil, err
	}

	for _, kp := range keypairs {
		if kp.Type == KeypairTypeProfile {
			return kp, nil
		}
	}

	panic("no profile keypair among known keypairs")
}

func (db *Database) GetWalletAddresses() (rst []types.Address, err error) {
	rows, err := db.db.Query("SELECT address FROM keypairs_accounts WHERE chat = 0 AND removed = 0 ORDER BY created_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		addr := types.Address{}
		err = rows.Scan(&addr)
		if err != nil {
			return nil, err
		}
		rst = append(rst, addr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rst, nil
}

func (db *Database) GetChatAddress() (rst types.Address, err error) {
	err = db.db.QueryRow("SELECT address FROM keypairs_accounts WHERE chat = 1").Scan(&rst)
	return
}

func (db *Database) GetAddresses() (rst []types.Address, err error) {
	rows, err := db.db.Query("SELECT address FROM keypairs_accounts WHERE removed = 0 ORDER BY created_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		addr := types.Address{}
		err = rows.Scan(&addr)
		if err != nil {
			return nil, err
		}
		rst = append(rst, addr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rst, nil
}

func (db *Database) keypairExists(tx *sql.Tx, keyUID string) (exists bool, err error) {
	query := `SELECT EXISTS (SELECT 1 FROM keypairs WHERE key_uid = ? AND removed = 0)`

	if tx == nil {
		err = db.db.QueryRow(query, keyUID).Scan(&exists)
	} else {
		err = tx.QueryRow(query, keyUID).Scan(&exists)
	}

	return exists, err
}

// KeypairExists returns true if given address is stored in database.
func (db *Database) KeypairExists(keyUID string) (exists bool, err error) {
	return db.keypairExists(nil, keyUID)
}

// AddressExists returns true if given address is stored in database.
func (db *Database) AddressExists(address types.Address) (exists bool, err error) {
	err = db.db.QueryRow("SELECT EXISTS (SELECT 1 FROM keypairs_accounts WHERE address = ? AND removed = 0)", address).Scan(&exists)
	return exists, err
}

// GetPath returns true if account with given address was recently key and doesn't have a key yet
func (db *Database) GetPath(address types.Address) (path string, err error) {
	err = db.db.QueryRow("SELECT path FROM keypairs_accounts WHERE address = ? AND removed = 0", address).Scan(&path)
	return path, err
}

// NOTE: This should not be used to retrieve `Networks`.
// NetworkManager should be used instead, otherwise RPCURL will be empty
func (db *Database) GetNodeConfig() (*params.NodeConfig, error) {
	return nodecfg.GetNodeConfigFromDB(db.db)
}

// Basically this function should not update the clock, cause it marks keypair/accounts locally. But...
// we need to cover the case when user recovers a Status account from waku, then pairs another device via
// local pairing and then imports seed/private key for the non profile keypair on one of those two devices
// to make that keypair fully operable. In that case we need to inform other device about the change, that
// other device may offer other options for importing that keypair on it.
// If the clock is set to -1, do not update it.
func (db *Database) MarkKeypairFullyOperable(keyUID string, clock uint64, updateKeypairClock bool) (err error) {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	kp, err := db.getKeypairByKeyUID(tx, keyUID, false)
	if err != nil {
		return err
	}

	for _, acc := range kp.Accounts {
		_, err = tx.Exec(`UPDATE keypairs_accounts SET operable = ?	WHERE address = ?`, AccountFullyOperable, acc.Address)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`UPDATE keypairs SET synced_from = "" WHERE key_uid = ?`, keyUID)
	if err != nil {
		return err
	}

	if updateKeypairClock {
		return db.updateKeypairClock(tx, keyUID, clock)
	}

	return nil
}

func (db *Database) MarkAccountFullyOperable(address types.Address) (err error) {
	_, err = db.db.Exec(`UPDATE keypairs_accounts SET operable = ?	WHERE address = ?`, AccountFullyOperable, address)
	return err
}

// This function should not update the clock, cause it marks a keypair locally.
func (db *Database) SetKeypairSyncedFrom(address types.Address, operable AccountOperable) (err error) {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	_, err = db.getAccountByAddress(tx, address)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE keypairs_accounts SET operable = ?	WHERE address = ?`, operable, address)
	return err
}

func (db *Database) GetPositionForNextNewAccount() (int64, error) {
	var pos sql.NullInt64
	err := db.db.QueryRow("SELECT MAX(position) FROM keypairs_accounts WHERE removed = 0").Scan(&pos)
	if err != nil {
		return 0, err
	}
	if pos.Valid {
		return pos.Int64 + 1, nil
	}
	return 0, nil
}

// This function should not be used directly, it is called from the functions which reorders accounts.
func (db *Database) setClockOfLastAccountsPositionChange(tx *sql.Tx, clock uint64) error {
	if tx == nil {
		return nil
	}
	_, err := tx.Exec("UPDATE settings SET wallet_accounts_position_change_clock = ? WHERE synthetic_id = 'id'", clock)
	return err
}

func (db *Database) GetClockOfLastAccountsPositionChange() (result uint64, err error) {
	query := "SELECT wallet_accounts_position_change_clock FROM settings WHERE synthetic_id = 'id'"
	err = db.db.QueryRow(query).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, err
}

// Updates positions of accounts respecting current order.
func (db *Database) ResolveAccountsPositions(clock uint64) (err error) {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	// returns all accounts ordered by position
	dbAccounts, err := db.getAccounts(tx, types.Address{}, false)
	if err != nil {
		return err
	}

	// starting from -1, cause `getAccounts` returns chat account as well
	for i := 0; i < len(dbAccounts); i++ {
		expectedPosition := int64(i - 1)
		if dbAccounts[i].Position != expectedPosition {
			_, err = tx.Exec("UPDATE keypairs_accounts SET position = ? WHERE address = ?", expectedPosition, dbAccounts[i].Address)
			if err != nil {
				return err
			}
		}
	}

	return db.setClockOfLastAccountsPositionChange(tx, clock)
}

// Sets positions for passed accounts.
func (db *Database) SetWalletAccountsPositions(accounts []*Account, clock uint64) (err error) {
	if len(accounts) == 0 {
		return nil
	}
	for _, acc := range accounts {
		if acc.Position < 0 {
			return ErrAccountWrongPosition
		}
	}
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	dbAccounts, err := db.getAccounts(tx, types.Address{}, false)
	if err != nil {
		return err
	}

	// we need to subtract 1, because of the chat account
	if len(dbAccounts)-1 != len(accounts) {
		return ErrNotTheSameNumberOdAccountsToApplyReordering
	}

	for _, dbAcc := range dbAccounts {
		if dbAcc.Chat {
			continue
		}
		found := false
		for _, acc := range accounts {
			if dbAcc.Address == acc.Address {
				found = true
				break
			}
		}
		if !found {
			return ErrNotTheSameAccountsToApplyReordering
		}
	}

	for _, acc := range accounts {
		_, err = tx.Exec("UPDATE keypairs_accounts SET position = ? WHERE address = ?", acc.Position, acc.Address)
		if err != nil {
			return err
		}
	}

	return db.setClockOfLastAccountsPositionChange(tx, clock)
}

// Moves wallet account fromPosition to toPosition.
func (db *Database) MoveWalletAccount(fromPosition int64, toPosition int64, clock uint64) (err error) {
	if fromPosition < 0 || toPosition < 0 || fromPosition == toPosition {
		return ErrMovingAccountToWrongPosition
	}
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	var (
		newMaxPosition int64
		newMinPosition int64
	)
	err = tx.QueryRow("SELECT MAX(position), MIN(position) FROM keypairs_accounts WHERE removed = 0").Scan(&newMaxPosition, &newMinPosition)
	if err != nil {
		return err
	}
	newMaxPosition++
	newMinPosition--

	if toPosition > fromPosition {
		_, err = tx.Exec("UPDATE keypairs_accounts SET position = ? WHERE position = ? AND removed = 0", newMaxPosition, fromPosition)
		if err != nil {
			return err
		}
		for i := fromPosition + 1; i <= toPosition; i++ {
			_, err = tx.Exec("UPDATE keypairs_accounts SET position = ? WHERE position = ? AND removed = 0", i-1, i)
			if err != nil {
				return err
			}
		}
		_, err = tx.Exec("UPDATE keypairs_accounts SET position = ? WHERE position = ? AND removed = 0", toPosition, newMaxPosition)
		if err != nil {
			return err
		}
	} else {
		_, err = tx.Exec("UPDATE keypairs_accounts SET position = ? WHERE position = ? AND removed = 0", newMinPosition, fromPosition)
		if err != nil {
			return err
		}
		for i := fromPosition - 1; i >= toPosition; i-- {
			_, err = tx.Exec("UPDATE keypairs_accounts SET position = ? WHERE position = ? AND removed = 0", i+1, i)
			if err != nil {
				return err
			}
		}
		_, err = tx.Exec("UPDATE keypairs_accounts SET position = ? WHERE position = ? AND removed = 0", toPosition, newMinPosition)
		if err != nil {
			return err
		}
	}

	return db.setClockOfLastAccountsPositionChange(tx, clock)
}

func (db *Database) CheckAndDeleteExpiredKeypairsAndAccounts(time uint64) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	// Check keypairs first
	dbKeypairs, err := db.getKeypairs(tx, "", true)
	if err != nil {
		return err
	}

	for _, dbKp := range dbKeypairs {
		if dbKp.Type == KeypairTypeProfile ||
			!dbKp.Removed ||
			time-dbKp.Clock < ThirtyDaysInMilliseconds {
			continue
		}
		query := `
				DELETE
				FROM
					keypairs
				WHERE
					key_uid = ?
			`
		_, err := tx.Exec(query, dbKp.KeyUID)
		if err != nil {
			return err
		}
	}

	// Check accounts (keypair related and watch only as well)
	dbAccounts, err := db.getAccounts(tx, types.Address{}, true)
	if err != nil {
		return err
	}

	for _, dbAcc := range dbAccounts {
		if dbAcc.Chat ||
			dbAcc.Wallet ||
			!dbAcc.Removed ||
			time-dbAcc.Clock < ThirtyDaysInMilliseconds {
			continue
		}

		query := `
			DELETE
			FROM
				keypairs_accounts
			WHERE
				address = ?
		`
		_, err := tx.Exec(query, dbAcc.Address)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Database) AddressWasShown(address types.Address) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	_, err = tx.Exec(`UPDATE keypairs_accounts SET address_was_not_shown = 0 WHERE address = ?`, address)
	return err
}
