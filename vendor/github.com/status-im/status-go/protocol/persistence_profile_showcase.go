package protocol

import (
	"context"
	"database/sql"
	"errors"
)

type ProfileShowcaseVisibility int

const (
	ProfileShowcaseVisibilityNoOne ProfileShowcaseVisibility = iota
	ProfileShowcaseVisibilityIDVerifiedContacts
	ProfileShowcaseVisibilityContacts
	ProfileShowcaseVisibilityEveryone
)

const upsertProfileShowcaseCommunityPreferenceQuery = "INSERT OR REPLACE INTO profile_showcase_communities_preferences(community_id, visibility, sort_order) VALUES (?, ?, ?)" // #nosec G101
const selectProfileShowcaseCommunityPreferenceQuery = "SELECT community_id, visibility, sort_order FROM profile_showcase_communities_preferences"                              // #nosec G101
const deleteProfileShowcaseCommunityPreferenceQuery = "DELETE FROM profile_showcase_communities_preferences WHERE community_id = ?"                                            // #nosec G101

const upsertProfileShowcaseAccountPreferenceQuery = "INSERT OR REPLACE INTO profile_showcase_accounts_preferences(address, name, color_id, emoji, visibility, sort_order) VALUES (?, ?, ?, ?, ?, ?)" // #nosec G101
const selectProfileShowcaseAccountPreferenceQuery = "SELECT address, name, color_id, emoji, visibility, sort_order FROM profile_showcase_accounts_preferences"                                       // #nosec G101
const selectSpecifiedShowcaseAccountPreferenceQuery = "SELECT address, name, color_id, emoji, visibility, sort_order FROM profile_showcase_accounts_preferences WHERE address = ?"                   // #nosec G101
const deleteProfileShowcaseAccountPreferenceQuery = "DELETE FROM profile_showcase_accounts_preferences WHERE address = ?"                                                                            // #nosec G101

const upsertProfileShowcaseCollectiblePreferenceQuery = "INSERT OR REPLACE INTO profile_showcase_collectibles_preferences(contract_address, chain_id, token_id, community_id, account_address, visibility, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?)" // #nosec G101
const selectProfileShowcaseCollectiblePreferenceQuery = "SELECT contract_address, chain_id, token_id, community_id, account_address, visibility, sort_order FROM profile_showcase_collectibles_preferences"                                          // #nosec G101

const upsertProfileShowcaseVerifiedTokenPreferenceQuery = "INSERT OR REPLACE INTO profile_showcase_verified_tokens_preferences(symbol, visibility, sort_order) VALUES (?, ?, ?)" // #nosec G101
const selectProfileShowcaseVerifiedTokenPreferenceQuery = "SELECT symbol, visibility, sort_order FROM profile_showcase_verified_tokens_preferences"                              // #nosec G101

const upsertProfileShowcaseUnverifiedTokenPreferenceQuery = "INSERT OR REPLACE INTO profile_showcase_unverified_tokens_preferences(contract_address, chain_id, community_id, visibility, sort_order) VALUES (?, ?, ?, ?, ?)" // #nosec G101
const selectProfileShowcaseUnverifiedTokenPreferenceQuery = "SELECT contract_address, chain_id, community_id, visibility, sort_order FROM profile_showcase_unverified_tokens_preferences"                                    // #nosec G101

const upsertContactProfileShowcaseCommunityQuery = "INSERT OR REPLACE INTO profile_showcase_communities_contacts(contact_id, community_id, sort_order) VALUES (?, ?, ?)" // #nosec G101
const selectContactProfileShowcaseCommunityQuery = "SELECT community_id, sort_order FROM profile_showcase_communities_contacts WHERE contact_id = ?"                     // #nosec G101
const removeContactProfileShowcaseCommunityQuery = "DELETE FROM profile_showcase_communities_contacts WHERE contact_id = ?"                                              // #nosec G101

const upsertContactProfileShowcaseAccountQuery = "INSERT OR REPLACE INTO profile_showcase_accounts_contacts(contact_id, address, name, color_id, emoji, sort_order) VALUES (?, ?, ?, ?, ?, ?)" // #nosec G101
const selectContactProfileShowcaseAccountQuery = "SELECT * FROM profile_showcase_accounts_contacts WHERE contact_id = ?"                                                                       // #nosec G101
const removeContactProfileShowcaseAccountQuery = "DELETE FROM profile_showcase_accounts_contacts WHERE contact_id = ?"                                                                         // #nosec G101

const upsertContactProfileShowcaseCollectibleQuery = "INSERT OR REPLACE INTO profile_showcase_collectibles_contacts(contact_id, contract_address, chain_id, token_id, community_id, account_address, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?)" // #nosec G101
const selectContactProfileShowcaseCollectibleQuery = "SELECT contract_address, chain_id, token_id, community_id, account_address, sort_order FROM profile_showcase_collectibles_contacts WHERE contact_id = ?"                                 // #nosec G101
const removeContactProfileShowcaseCollectibleQuery = "DELETE FROM profile_showcase_collectibles_contacts WHERE contact_id = ?"                                                                                                                 // #nosec G101

const upsertContactProfileShowcaseVerifiedTokenQuery = "INSERT OR REPLACE INTO profile_showcase_verified_tokens_contacts(contact_id, symbol, sort_order) VALUES (?, ?, ?)" // #nosec G101
const selectContactProfileShowcaseVerifiedTokenQuery = "SELECT symbol, sort_order FROM profile_showcase_verified_tokens_contacts WHERE contact_id = ?"                     // #nosec G101
const removeContactProfileShowcaseVerifiedTokenQuery = "DELETE FROM profile_showcase_verified_tokens_contacts WHERE contact_id = ?"                                        // #nosec G101

const upsertContactProfileShowcaseUnverifiedTokenQuery = "INSERT OR REPLACE INTO profile_showcase_unverified_tokens_contacts(contact_id, contract_address, chain_id, community_id, sort_order) VALUES (?, ?, ?, ?, ?)" // #nosec G101
const selectContactProfileShowcaseUnverifiedTokenQuery = "SELECT contract_address, chain_id, community_id, sort_order FROM profile_showcase_unverified_tokens_contacts WHERE contact_id = ?"                           // #nosec G101
const removeContactProfileShowcaseUnverifiedTokenQuery = "DELETE FROM profile_showcase_unverified_tokens_contacts WHERE contact_id = ?"                                                                                // #nosec G101

const selectProfileShowcaseAccountsWhichMatchTheAddress = `
SELECT psa.*
FROM
	contacts c
LEFT JOIN
	profile_showcase_accounts_contacts psa
ON
	c.id = psa.contact_id
WHERE
	psa.address = ?
`

type ProfileShowcaseCommunityPreference struct {
	CommunityID        string                    `json:"communityId"`
	ShowcaseVisibility ProfileShowcaseVisibility `json:"showcaseVisibility"`
	Order              int                       `json:"order"`
}

type ProfileShowcaseAccountPreference struct {
	Address            string                    `json:"address"`
	Name               string                    `json:"name"`
	ColorID            string                    `json:"colorId"`
	Emoji              string                    `json:"emoji"`
	ShowcaseVisibility ProfileShowcaseVisibility `json:"showcaseVisibility"`
	Order              int                       `json:"order"`
}

type ProfileShowcaseCollectiblePreference struct {
	ContractAddress    string                    `json:"contractAddress"`
	ChainID            uint64                    `json:"chainId"`
	TokenID            string                    `json:"tokenId"`
	CommunityID        string                    `json:"communityId"`
	AccountAddress     string                    `json:"accountAddress"`
	ShowcaseVisibility ProfileShowcaseVisibility `json:"showcaseVisibility"`
	Order              int                       `json:"order"`
}

type ProfileShowcaseVerifiedTokenPreference struct {
	Symbol             string                    `json:"symbol"`
	ShowcaseVisibility ProfileShowcaseVisibility `json:"showcaseVisibility"`
	Order              int                       `json:"order"`
}

type ProfileShowcaseUnverifiedTokenPreference struct {
	ContractAddress    string                    `json:"contractAddress"`
	ChainID            uint64                    `json:"chainId"`
	CommunityID        string                    `json:"communityId"`
	ShowcaseVisibility ProfileShowcaseVisibility `json:"showcaseVisibility"`
	Order              int                       `json:"order"`
}

type ProfileShowcasePreferences struct {
	Communities      []*ProfileShowcaseCommunityPreference       `json:"communities"`
	Accounts         []*ProfileShowcaseAccountPreference         `json:"accounts"`
	Collectibles     []*ProfileShowcaseCollectiblePreference     `json:"collectibles"`
	VerifiedTokens   []*ProfileShowcaseVerifiedTokenPreference   `json:"verifiedTokens"`
	UnverifiedTokens []*ProfileShowcaseUnverifiedTokenPreference `json:"unverifiedTokens"`
}

type ProfileShowcaseCommunity struct {
	CommunityID string `json:"communityId"`
	Order       int    `json:"order"`
}

type ProfileShowcaseAccount struct {
	ContactID string `json:"contactId"`
	Address   string `json:"address"`
	Name      string `json:"name"`
	ColorID   string `json:"colorId"`
	Emoji     string `json:"emoji"`
	Order     int    `json:"order"`
}

type ProfileShowcaseCollectible struct {
	ContractAddress string `json:"contractAddress"`
	ChainID         uint64 `json:"chainId"`
	TokenID         string `json:"tokenId"`
	CommunityID     string `json:"communityId"`
	AccountAddress  string `json:"accountAddress"`
	Order           int    `json:"order"`
}

type ProfileShowcaseVerifiedToken struct {
	Symbol string `json:"symbol"`
	Order  int    `json:"order"`
}

type ProfileShowcaseUnverifiedToken struct {
	ContractAddress string `json:"contractAddress"`
	ChainID         uint64 `json:"chainId"`
	CommunityID     string `json:"communityId"`
	Order           int    `json:"order"`
}

type ProfileShowcase struct {
	ContactID        string                            `json:"contactId"`
	Communities      []*ProfileShowcaseCommunity       `json:"communities"`
	Accounts         []*ProfileShowcaseAccount         `json:"accounts"`
	Collectibles     []*ProfileShowcaseCollectible     `json:"collectibles"`
	VerifiedTokens   []*ProfileShowcaseVerifiedToken   `json:"verifiedTokens"`
	UnverifiedTokens []*ProfileShowcaseUnverifiedToken `json:"unverifiedTokens"`
}

// Queries for showcase preferences
func (db sqlitePersistence) saveProfileShowcaseCommunityPreference(tx *sql.Tx, community *ProfileShowcaseCommunityPreference) error {
	_, err := tx.Exec(upsertProfileShowcaseCommunityPreferenceQuery,
		community.CommunityID,
		community.ShowcaseVisibility,
		community.Order,
	)

	return err
}

func (db sqlitePersistence) getProfileShowcaseCommunitiesPreferences(tx *sql.Tx) ([]*ProfileShowcaseCommunityPreference, error) {
	rows, err := tx.Query(selectProfileShowcaseCommunityPreferenceQuery)
	if err != nil {
		return nil, err
	}

	communities := []*ProfileShowcaseCommunityPreference{}

	for rows.Next() {
		community := &ProfileShowcaseCommunityPreference{}

		err := rows.Scan(
			&community.CommunityID,
			&community.ShowcaseVisibility,
			&community.Order,
		)

		if err != nil {
			return nil, err
		}

		communities = append(communities, community)
	}
	return communities, nil
}

func (db sqlitePersistence) saveProfileShowcaseAccountPreference(tx *sql.Tx, account *ProfileShowcaseAccountPreference) error {
	_, err := tx.Exec(upsertProfileShowcaseAccountPreferenceQuery,
		account.Address,
		account.Name,
		account.ColorID,
		account.Emoji,
		account.ShowcaseVisibility,
		account.Order,
	)

	return err
}

func (db sqlitePersistence) processProfileShowcaseAccountPreferences(rows *sql.Rows) (result []*ProfileShowcaseAccountPreference, err error) {
	if rows == nil {
		return nil, errors.New("rows is nil")
	}

	for rows.Next() {
		account := &ProfileShowcaseAccountPreference{}

		err := rows.Scan(
			&account.Address,
			&account.Name,
			&account.ColorID,
			&account.Emoji,
			&account.ShowcaseVisibility,
			&account.Order,
		)

		if err != nil {
			return nil, err
		}

		result = append(result, account)
	}

	err = rows.Err()
	return
}

func (db sqlitePersistence) getProfileShowcaseAccountsPreferences(tx *sql.Tx) ([]*ProfileShowcaseAccountPreference, error) {
	rows, err := tx.Query(selectProfileShowcaseAccountPreferenceQuery)
	if err != nil {
		return nil, err
	}

	return db.processProfileShowcaseAccountPreferences(rows)
}

func (db sqlitePersistence) GetProfileShowcaseAccountPreference(accountAddress string) (*ProfileShowcaseAccountPreference, error) {
	rows, err := db.db.Query(selectSpecifiedShowcaseAccountPreferenceQuery, accountAddress)
	if err != nil {
		return nil, err
	}

	accounts, err := db.processProfileShowcaseAccountPreferences(rows)
	if len(accounts) > 0 {
		return accounts[0], err
	}
	return nil, err
}

func (db sqlitePersistence) DeleteProfileShowcaseAccountPreference(accountAddress string) (bool, error) {
	result, err := db.db.Exec(deleteProfileShowcaseAccountPreferenceQuery, accountAddress)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	return rows > 0, err
}

func (db sqlitePersistence) DeleteProfileShowcaseCommunityPreference(communityID string) (bool, error) {
	result, err := db.db.Exec(deleteProfileShowcaseCommunityPreferenceQuery, communityID)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	return rows > 0, err
}

func (db sqlitePersistence) saveProfileShowcaseCollectiblePreference(tx *sql.Tx, collectible *ProfileShowcaseCollectiblePreference) error {
	_, err := tx.Exec(upsertProfileShowcaseCollectiblePreferenceQuery,
		collectible.ContractAddress,
		collectible.ChainID,
		collectible.TokenID,
		collectible.CommunityID,
		collectible.AccountAddress,
		collectible.ShowcaseVisibility,
		collectible.Order,
	)

	return err
}

func (db sqlitePersistence) getProfileShowcaseCollectiblesPreferences(tx *sql.Tx) ([]*ProfileShowcaseCollectiblePreference, error) {
	rows, err := tx.Query(selectProfileShowcaseCollectiblePreferenceQuery)
	if err != nil {
		return nil, err
	}

	collectibles := []*ProfileShowcaseCollectiblePreference{}

	for rows.Next() {
		collectible := &ProfileShowcaseCollectiblePreference{}

		err := rows.Scan(
			&collectible.ContractAddress,
			&collectible.ChainID,
			&collectible.TokenID,
			&collectible.CommunityID,
			&collectible.AccountAddress,
			&collectible.ShowcaseVisibility,
			&collectible.Order,
		)

		if err != nil {
			return nil, err
		}

		collectibles = append(collectibles, collectible)
	}
	return collectibles, nil
}

func (db sqlitePersistence) saveProfileShowcaseVerifiedTokenPreference(tx *sql.Tx, token *ProfileShowcaseVerifiedTokenPreference) error {
	_, err := tx.Exec(upsertProfileShowcaseVerifiedTokenPreferenceQuery,
		token.Symbol,
		token.ShowcaseVisibility,
		token.Order,
	)

	return err
}

func (db sqlitePersistence) saveProfileShowcaseUnverifiedTokenPreference(tx *sql.Tx, token *ProfileShowcaseUnverifiedTokenPreference) error {
	_, err := tx.Exec(upsertProfileShowcaseUnverifiedTokenPreferenceQuery,
		token.ContractAddress,
		token.ChainID,
		token.CommunityID,
		token.ShowcaseVisibility,
		token.Order,
	)

	return err
}

func (db sqlitePersistence) getProfileShowcaseVerifiedTokensPreferences(tx *sql.Tx) ([]*ProfileShowcaseVerifiedTokenPreference, error) {
	rows, err := tx.Query(selectProfileShowcaseVerifiedTokenPreferenceQuery)
	if err != nil {
		return nil, err
	}

	tokens := []*ProfileShowcaseVerifiedTokenPreference{}

	for rows.Next() {
		token := &ProfileShowcaseVerifiedTokenPreference{}

		err := rows.Scan(
			&token.Symbol,
			&token.ShowcaseVisibility,
			&token.Order,
		)

		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
	}
	return tokens, nil
}

func (db sqlitePersistence) getProfileShowcaseUnverifiedTokensPreferences(tx *sql.Tx) ([]*ProfileShowcaseUnverifiedTokenPreference, error) {
	rows, err := tx.Query(selectProfileShowcaseUnverifiedTokenPreferenceQuery)
	if err != nil {
		return nil, err
	}

	tokens := []*ProfileShowcaseUnverifiedTokenPreference{}

	for rows.Next() {
		token := &ProfileShowcaseUnverifiedTokenPreference{}

		err := rows.Scan(
			&token.ContractAddress,
			&token.ChainID,
			&token.CommunityID,
			&token.ShowcaseVisibility,
			&token.Order,
		)

		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
	}
	return tokens, nil
}

// Queries for contacts showcase
func (db sqlitePersistence) saveProfileShowcaseCommunityContact(tx *sql.Tx, contactID string, community *ProfileShowcaseCommunity) error {
	_, err := tx.Exec(upsertContactProfileShowcaseCommunityQuery,
		contactID,
		community.CommunityID,
		community.Order,
	)

	return err
}

func (db sqlitePersistence) getProfileShowcaseCommunitiesContact(tx *sql.Tx, contactID string) ([]*ProfileShowcaseCommunity, error) {
	rows, err := tx.Query(selectContactProfileShowcaseCommunityQuery, contactID)
	if err != nil {
		return nil, err
	}

	communities := []*ProfileShowcaseCommunity{}

	for rows.Next() {
		community := &ProfileShowcaseCommunity{}

		err := rows.Scan(&community.CommunityID, &community.Order)
		if err != nil {
			return nil, err
		}

		communities = append(communities, community)
	}
	return communities, nil
}

func (db sqlitePersistence) clearProfileShowcaseCommunityContact(tx *sql.Tx, contactID string) error {
	_, err := tx.Exec(removeContactProfileShowcaseCommunityQuery, contactID)
	if err != nil {
		return err
	}

	return nil
}

func (db sqlitePersistence) saveProfileShowcaseAccountContact(tx *sql.Tx, contactID string, account *ProfileShowcaseAccount) error {
	_, err := tx.Exec(upsertContactProfileShowcaseAccountQuery,
		contactID,
		account.Address,
		account.Name,
		account.ColorID,
		account.Emoji,
		account.Order,
	)

	return err
}

func (db sqlitePersistence) processProfileShowcaseAccounts(rows *sql.Rows) (result []*ProfileShowcaseAccount, err error) {
	if rows == nil {
		return nil, errors.New("rows is nil")
	}

	for rows.Next() {
		account := &ProfileShowcaseAccount{}

		err = rows.Scan(&account.Address, &account.Name, &account.ColorID, &account.Emoji, &account.Order, &account.ContactID)
		if err != nil {
			return
		}

		result = append(result, account)
	}

	err = rows.Err()
	return
}

func (db sqlitePersistence) getProfileShowcaseAccountsContact(tx *sql.Tx, contactID string) ([]*ProfileShowcaseAccount, error) {
	rows, err := tx.Query(selectContactProfileShowcaseAccountQuery, contactID)
	if err != nil {
		return nil, err
	}

	return db.processProfileShowcaseAccounts(rows)
}

func (db sqlitePersistence) GetProfileShowcaseAccountsByAddress(address string) ([]*ProfileShowcaseAccount, error) {
	rows, err := db.db.Query(selectProfileShowcaseAccountsWhichMatchTheAddress, address)
	if err != nil {
		return nil, err
	}

	return db.processProfileShowcaseAccounts(rows)
}

func (db sqlitePersistence) clearProfileShowcaseAccountsContact(tx *sql.Tx, contactID string) error {
	_, err := tx.Exec(removeContactProfileShowcaseAccountQuery, contactID)
	return err
}

func (db sqlitePersistence) saveProfileShowcaseCollectibleContact(tx *sql.Tx, contactID string, collectible *ProfileShowcaseCollectible) error {
	_, err := tx.Exec(upsertContactProfileShowcaseCollectibleQuery,
		contactID,
		collectible.ContractAddress,
		collectible.ChainID,
		collectible.TokenID,
		collectible.CommunityID,
		collectible.AccountAddress,
		collectible.Order,
	)

	return err
}

func (db sqlitePersistence) getProfileShowcaseCollectiblesContact(tx *sql.Tx, contactID string) ([]*ProfileShowcaseCollectible, error) {
	rows, err := tx.Query(selectContactProfileShowcaseCollectibleQuery, contactID)
	if err != nil {
		return nil, err
	}

	collectibles := []*ProfileShowcaseCollectible{}

	for rows.Next() {
		collectible := &ProfileShowcaseCollectible{}

		err := rows.Scan(
			&collectible.ContractAddress,
			&collectible.ChainID,
			&collectible.TokenID,
			&collectible.CommunityID,
			&collectible.AccountAddress,
			&collectible.Order)
		if err != nil {
			return nil, err
		}

		collectibles = append(collectibles, collectible)
	}
	return collectibles, nil
}

func (db sqlitePersistence) clearProfileShowcaseCollectiblesContact(tx *sql.Tx, contactID string) error {
	_, err := tx.Exec(removeContactProfileShowcaseCollectibleQuery, contactID)
	return err
}

func (db sqlitePersistence) saveProfileShowcaseVerifiedTokenContact(tx *sql.Tx, contactID string, token *ProfileShowcaseVerifiedToken) error {
	_, err := tx.Exec(upsertContactProfileShowcaseVerifiedTokenQuery,
		contactID,
		token.Symbol,
		token.Order,
	)

	return err
}

func (db sqlitePersistence) saveProfileShowcaseUnverifiedTokenContact(tx *sql.Tx, contactID string, token *ProfileShowcaseUnverifiedToken) error {
	_, err := tx.Exec(upsertContactProfileShowcaseUnverifiedTokenQuery,
		contactID,
		token.ContractAddress,
		token.ChainID,
		token.CommunityID,
		token.Order,
	)

	return err
}

func (db sqlitePersistence) getProfileShowcaseVerifiedTokensContact(tx *sql.Tx, contactID string) ([]*ProfileShowcaseVerifiedToken, error) {
	rows, err := tx.Query(selectContactProfileShowcaseVerifiedTokenQuery, contactID)
	if err != nil {
		return nil, err
	}

	tokens := []*ProfileShowcaseVerifiedToken{}

	for rows.Next() {
		token := &ProfileShowcaseVerifiedToken{}

		err := rows.Scan(
			&token.Symbol,
			&token.Order)
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
	}
	return tokens, nil
}

func (db sqlitePersistence) getProfileShowcaseUnverifiedTokensContact(tx *sql.Tx, contactID string) ([]*ProfileShowcaseUnverifiedToken, error) {
	rows, err := tx.Query(selectContactProfileShowcaseUnverifiedTokenQuery, contactID)
	if err != nil {
		return nil, err
	}

	tokens := []*ProfileShowcaseUnverifiedToken{}

	for rows.Next() {
		token := &ProfileShowcaseUnverifiedToken{}

		err := rows.Scan(
			&token.ContractAddress,
			&token.ChainID,
			&token.CommunityID,
			&token.Order)
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
	}
	return tokens, nil
}

func (db sqlitePersistence) clearProfileShowcaseVerifiedTokensContact(tx *sql.Tx, contactID string) error {
	_, err := tx.Exec(removeContactProfileShowcaseVerifiedTokenQuery, contactID)
	return err
}

func (db sqlitePersistence) clearProfileShowcaseUnverifiedTokensContact(tx *sql.Tx, contactID string) error {
	_, err := tx.Exec(removeContactProfileShowcaseUnverifiedTokenQuery, contactID)
	return err
}

// public functions
func (db sqlitePersistence) SaveProfileShowcasePreferences(preferences *ProfileShowcasePreferences) error {
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	for _, community := range preferences.Communities {
		err = db.saveProfileShowcaseCommunityPreference(tx, community)
		if err != nil {
			return err
		}
	}

	for _, account := range preferences.Accounts {
		err = db.saveProfileShowcaseAccountPreference(tx, account)
		if err != nil {
			return err
		}
	}

	for _, collectible := range preferences.Collectibles {
		err = db.saveProfileShowcaseCollectiblePreference(tx, collectible)
		if err != nil {
			return err
		}
	}

	for _, token := range preferences.VerifiedTokens {
		err = db.saveProfileShowcaseVerifiedTokenPreference(tx, token)
		if err != nil {
			return err
		}
	}

	for _, token := range preferences.UnverifiedTokens {
		err = db.saveProfileShowcaseUnverifiedTokenPreference(tx, token)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db sqlitePersistence) SaveProfileShowcaseAccountPreference(account *ProfileShowcaseAccountPreference) error {
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()
	return db.saveProfileShowcaseAccountPreference(tx, account)
}

func (db sqlitePersistence) GetProfileShowcasePreferences() (*ProfileShowcasePreferences, error) {
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	communities, err := db.getProfileShowcaseCommunitiesPreferences(tx)
	if err != nil {
		return nil, err
	}

	accounts, err := db.getProfileShowcaseAccountsPreferences(tx)
	if err != nil {
		return nil, err
	}

	collectibles, err := db.getProfileShowcaseCollectiblesPreferences(tx)
	if err != nil {
		return nil, err
	}

	verifiedTokens, err := db.getProfileShowcaseVerifiedTokensPreferences(tx)
	if err != nil {
		return nil, err
	}

	unverifiedTokens, err := db.getProfileShowcaseUnverifiedTokensPreferences(tx)
	if err != nil {
		return nil, err
	}

	return &ProfileShowcasePreferences{
		Communities:      communities,
		Accounts:         accounts,
		Collectibles:     collectibles,
		VerifiedTokens:   verifiedTokens,
		UnverifiedTokens: unverifiedTokens,
	}, nil
}

func (db sqlitePersistence) SaveProfileShowcaseForContact(showcase *ProfileShowcase) error {
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	for _, community := range showcase.Communities {
		err = db.saveProfileShowcaseCommunityContact(tx, showcase.ContactID, community)
		if err != nil {
			return err
		}
	}

	for _, account := range showcase.Accounts {
		err = db.saveProfileShowcaseAccountContact(tx, showcase.ContactID, account)
		if err != nil {
			return err
		}
	}

	for _, collectible := range showcase.Collectibles {
		err = db.saveProfileShowcaseCollectibleContact(tx, showcase.ContactID, collectible)
		if err != nil {
			return err
		}
	}

	for _, token := range showcase.VerifiedTokens {
		err = db.saveProfileShowcaseVerifiedTokenContact(tx, showcase.ContactID, token)
		if err != nil {
			return err
		}
	}

	for _, token := range showcase.UnverifiedTokens {
		err = db.saveProfileShowcaseUnverifiedTokenContact(tx, showcase.ContactID, token)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db sqlitePersistence) GetProfileShowcaseForContact(contactID string) (*ProfileShowcase, error) {
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	communities, err := db.getProfileShowcaseCommunitiesContact(tx, contactID)
	if err != nil {
		return nil, err
	}

	accounts, err := db.getProfileShowcaseAccountsContact(tx, contactID)
	if err != nil {
		return nil, err
	}

	collectibles, err := db.getProfileShowcaseCollectiblesContact(tx, contactID)
	if err != nil {
		return nil, err
	}

	verifiedTokens, err := db.getProfileShowcaseVerifiedTokensContact(tx, contactID)
	if err != nil {
		return nil, err
	}

	unverifiedTokens, err := db.getProfileShowcaseUnverifiedTokensContact(tx, contactID)
	if err != nil {
		return nil, err
	}

	return &ProfileShowcase{
		ContactID:        contactID,
		Communities:      communities,
		Accounts:         accounts,
		Collectibles:     collectibles,
		VerifiedTokens:   verifiedTokens,
		UnverifiedTokens: unverifiedTokens,
	}, nil
}

func (db sqlitePersistence) ClearProfileShowcaseForContact(contactID string) error {
	tx, err := db.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	err = db.clearProfileShowcaseCommunityContact(tx, contactID)
	if err != nil {
		return err
	}

	err = db.clearProfileShowcaseAccountsContact(tx, contactID)
	if err != nil {
		return err
	}

	err = db.clearProfileShowcaseCollectiblesContact(tx, contactID)
	if err != nil {
		return err
	}

	err = db.clearProfileShowcaseVerifiedTokensContact(tx, contactID)
	if err != nil {
		return err
	}

	err = db.clearProfileShowcaseUnverifiedTokensContact(tx, contactID)
	if err != nil {
		return err
	}

	return nil
}
