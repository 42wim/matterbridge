package protocol

import (
	"context"
	"database/sql"
	"errors"

	"github.com/status-im/status-go/protocol/identity"
)

// Profile showcase preferences
const upsertProfileShowcasePreferencesQuery = "UPDATE profile_showcase_preferences SET clock=? WHERE NOT EXISTS (SELECT 1 FROM profile_showcase_preferences WHERE clock >= ?)"
const selectProfileShowcasePreferencesQuery = "SELECT clock FROM profile_showcase_preferences"

const upsertProfileShowcaseCommunityPreferenceQuery = "INSERT OR REPLACE INTO profile_showcase_communities_preferences(community_id, visibility, sort_order) VALUES (?, ?, ?)" // #nosec G101
const selectProfileShowcaseCommunityPreferenceQuery = "SELECT community_id, visibility, sort_order FROM profile_showcase_communities_preferences"                              // #nosec G101
const deleteProfileShowcaseCommunityPreferenceQuery = "DELETE FROM profile_showcase_communities_preferences WHERE community_id = ?"                                            // #nosec G101
const clearProfileShowcaseCommunitiyPreferencesQuery = "DELETE FROM profile_showcase_communities_preferences"                                                                  // #nosec G101

const upsertProfileShowcaseAccountPreferenceQuery = "INSERT OR REPLACE INTO profile_showcase_accounts_preferences(address, visibility, sort_order) VALUES (?, ?, ?)" // #nosec G101
const selectProfileShowcaseAccountPreferenceQuery = "SELECT address, visibility, sort_order FROM profile_showcase_accounts_preferences"                              // #nosec G101
const selectSpecifiedShowcaseAccountPreferenceQuery = "SELECT address, visibility, sort_order FROM profile_showcase_accounts_preferences WHERE address = ?"          // #nosec G101
const deleteProfileShowcaseAccountPreferenceQuery = "DELETE FROM profile_showcase_accounts_preferences WHERE address = ?"                                            // #nosec G101
const clearProfileShowcaseAccountPreferencesQuery = "DELETE FROM profile_showcase_accounts_preferences"                                                              // #nosec G101

const upsertProfileShowcaseCollectiblePreferenceQuery = "INSERT OR REPLACE INTO profile_showcase_collectibles_preferences(contract_address, chain_id, token_id, visibility, sort_order) VALUES (?, ?, ?, ?, ?)" // #nosec G101
const selectProfileShowcaseCollectiblePreferenceQuery = "SELECT contract_address, chain_id, token_id, visibility, sort_order FROM profile_showcase_collectibles_preferences"                                    // #nosec G101
const clearProfileShowcaseCollectiblePreferencesQuery = "DELETE FROM profile_showcase_collectibles_preferences"                                                                                                 // #nosec G101

const upsertProfileShowcaseVerifiedTokenPreferenceQuery = "INSERT OR REPLACE INTO profile_showcase_verified_tokens_preferences(symbol, visibility, sort_order) VALUES (?, ?, ?)" // #nosec G101
const selectProfileShowcaseVerifiedTokenPreferenceQuery = "SELECT symbol, visibility, sort_order FROM profile_showcase_verified_tokens_preferences"                              // #nosec G101
const clearProfileShowcaseVerifiedTokenPreferencesQuery = "DELETE FROM profile_showcase_verified_tokens_preferences"                                                             // #nosec G101

const upsertProfileShowcaseUnverifiedTokenPreferenceQuery = "INSERT OR REPLACE INTO profile_showcase_unverified_tokens_preferences(contract_address, chain_id, visibility, sort_order) VALUES (?, ?, ?, ?)" // #nosec G101
const selectProfileShowcaseUnverifiedTokenPreferenceQuery = "SELECT contract_address, chain_id, visibility, sort_order FROM profile_showcase_unverified_tokens_preferences"                                 // #nosec G101
const clearProfileShowcaseUnverifiedTokenPreferencesQuery = "DELETE FROM profile_showcase_unverified_tokens_preferences"                                                                                    // #nosec G101

const upsertProfileShowcaseSocialLinkPreferenceQuery = "INSERT OR REPLACE INTO profile_showcase_social_links_preferences(url, text, visibility, sort_order) VALUES (?, ?, ?, ?)" // #nosec G101
const selectProfileShowcaseSocialLinkPreferenceQuery = "SELECT url, text, visibility, sort_order FROM profile_showcase_social_links_preferences"                                 // #nosec G101
const clearProfileShowcaseSocialLinkPreferencesQuery = "DELETE FROM profile_showcase_social_links_preferences"                                                                   // #nosec G101

// Profile showcase for a contact
const upsertContactProfileShowcaseCommunityQuery = "INSERT OR REPLACE INTO profile_showcase_communities_contacts(contact_id, community_id, sort_order, grant) VALUES (?, ?, ?, ?)" // #nosec G101
const selectContactProfileShowcaseCommunityQuery = "SELECT community_id, sort_order, grant FROM profile_showcase_communities_contacts WHERE contact_id = ?"                        // #nosec G101
const removeContactProfileShowcaseCommunityQuery = "DELETE FROM profile_showcase_communities_contacts WHERE contact_id = ?"                                                        // #nosec G101

const upsertContactProfileShowcaseAccountQuery = "INSERT OR REPLACE INTO profile_showcase_accounts_contacts(contact_id, address, name, color_id, emoji, sort_order) VALUES (?, ?, ?, ?, ?, ?)" // #nosec G101
const selectContactProfileShowcaseAccountQuery = "SELECT * FROM profile_showcase_accounts_contacts WHERE contact_id = ?"                                                                       // #nosec G101
const removeContactProfileShowcaseAccountQuery = "DELETE FROM profile_showcase_accounts_contacts WHERE contact_id = ?"                                                                         // #nosec G101

const upsertContactProfileShowcaseCollectibleQuery = "INSERT OR REPLACE INTO profile_showcase_collectibles_contacts(contact_id, contract_address, chain_id, token_id, sort_order) VALUES (?, ?, ?, ?, ?)" // #nosec G101
const selectContactProfileShowcaseCollectibleQuery = "SELECT contract_address, chain_id, token_id, sort_order FROM profile_showcase_collectibles_contacts WHERE contact_id = ?"                           // #nosec G101
const removeContactProfileShowcaseCollectibleQuery = "DELETE FROM profile_showcase_collectibles_contacts WHERE contact_id = ?"                                                                            // #nosec G101

const upsertContactProfileShowcaseVerifiedTokenQuery = "INSERT OR REPLACE INTO profile_showcase_verified_tokens_contacts(contact_id, symbol, sort_order) VALUES (?, ?, ?)" // #nosec G101
const selectContactProfileShowcaseVerifiedTokenQuery = "SELECT symbol, sort_order FROM profile_showcase_verified_tokens_contacts WHERE contact_id = ?"                     // #nosec G101
const removeContactProfileShowcaseVerifiedTokenQuery = "DELETE FROM profile_showcase_verified_tokens_contacts WHERE contact_id = ?"                                        // #nosec G101

const upsertContactProfileShowcaseUnverifiedTokenQuery = "INSERT OR REPLACE INTO profile_showcase_unverified_tokens_contacts(contact_id, contract_address, chain_id, sort_order) VALUES (?, ?, ?, ?)" // #nosec G101
const selectContactProfileShowcaseUnverifiedTokenQuery = "SELECT contract_address, chain_id, sort_order FROM profile_showcase_unverified_tokens_contacts WHERE contact_id = ?"                        // #nosec G101
const removeContactProfileShowcaseUnverifiedTokenQuery = "DELETE FROM profile_showcase_unverified_tokens_contacts WHERE contact_id = ?"                                                               // #nosec G101

const upsertContactProfileShowcaseSocialLinkQuery = "INSERT OR REPLACE INTO profile_showcase_social_links_contacts(contact_id, url, text, sort_order) VALUES (?, ?, ?, ?)" // #nosec G101
const selectContactProfileShowcaseSocialLinkQuery = "SELECT url, text, sort_order FROM profile_showcase_social_links_contacts WHERE contact_id = ?"                        // #nosec G101
const removeContactProfileShowcaseSocialLinkQuery = "DELETE FROM profile_showcase_social_links_contacts WHERE contact_id = ?"                                              // #nosec G101

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

// Queries for the profile showcase preferences

func (db sqlitePersistence) saveProfileShowcasePreferencesClock(tx *sql.Tx, clock uint64) error {
	_, err := tx.Exec(upsertProfileShowcasePreferencesQuery, clock, clock)
	return err
}

func (db sqlitePersistence) getProfileShowcasePreferencesClock(tx *sql.Tx) (uint64, error) {
	var clock uint64
	err := tx.QueryRow(selectProfileShowcasePreferencesQuery).Scan(&clock)
	return clock, err
}

func (db sqlitePersistence) saveProfileShowcaseCommunityPreference(tx *sql.Tx, community *identity.ProfileShowcaseCommunityPreference) error {
	_, err := tx.Exec(upsertProfileShowcaseCommunityPreferenceQuery,
		community.CommunityID,
		community.ShowcaseVisibility,
		community.Order,
	)

	return err
}

func (db sqlitePersistence) getProfileShowcaseCommunitiesPreferences(tx *sql.Tx) ([]*identity.ProfileShowcaseCommunityPreference, error) {
	rows, err := tx.Query(selectProfileShowcaseCommunityPreferenceQuery)
	if err != nil {
		return nil, err
	}

	communities := []*identity.ProfileShowcaseCommunityPreference{}

	for rows.Next() {
		community := &identity.ProfileShowcaseCommunityPreference{}

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

func (db sqlitePersistence) DeleteProfileShowcaseCommunityPreference(communityID string) (bool, error) {
	result, err := db.db.Exec(deleteProfileShowcaseCommunityPreferenceQuery, communityID)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	return rows > 0, err
}

func (db sqlitePersistence) clearProfileShowcaseCommunityPreferences(tx *sql.Tx) error {
	_, err := tx.Exec(clearProfileShowcaseCommunitiyPreferencesQuery)
	return err
}

func (db sqlitePersistence) saveProfileShowcaseAccountPreference(tx *sql.Tx, account *identity.ProfileShowcaseAccountPreference) error {
	_, err := tx.Exec(upsertProfileShowcaseAccountPreferenceQuery,
		account.Address,
		account.ShowcaseVisibility,
		account.Order,
	)

	return err
}

func (db sqlitePersistence) processProfileShowcaseAccountPreferences(rows *sql.Rows) (result []*identity.ProfileShowcaseAccountPreference, err error) {
	if rows == nil {
		return nil, errors.New("rows is nil")
	}

	for rows.Next() {
		account := &identity.ProfileShowcaseAccountPreference{}

		err := rows.Scan(
			&account.Address,
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

func (db sqlitePersistence) getProfileShowcaseAccountsPreferences(tx *sql.Tx) ([]*identity.ProfileShowcaseAccountPreference, error) {
	rows, err := tx.Query(selectProfileShowcaseAccountPreferenceQuery)
	if err != nil {
		return nil, err
	}

	return db.processProfileShowcaseAccountPreferences(rows)
}

func (db sqlitePersistence) GetProfileShowcaseAccountPreference(accountAddress string) (*identity.ProfileShowcaseAccountPreference, error) {
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

func (db sqlitePersistence) clearProfileShowcaseAccountPreferences(tx *sql.Tx) error {
	_, err := tx.Exec(clearProfileShowcaseAccountPreferencesQuery)
	return err
}

func (db sqlitePersistence) saveProfileShowcaseCollectiblePreference(tx *sql.Tx, collectible *identity.ProfileShowcaseCollectiblePreference) error {
	_, err := tx.Exec(upsertProfileShowcaseCollectiblePreferenceQuery,
		collectible.ContractAddress,
		collectible.ChainID,
		collectible.TokenID,
		collectible.ShowcaseVisibility,
		collectible.Order,
	)

	return err
}

func (db sqlitePersistence) getProfileShowcaseCollectiblesPreferences(tx *sql.Tx) ([]*identity.ProfileShowcaseCollectiblePreference, error) {
	rows, err := tx.Query(selectProfileShowcaseCollectiblePreferenceQuery)
	if err != nil {
		return nil, err
	}

	collectibles := []*identity.ProfileShowcaseCollectiblePreference{}

	for rows.Next() {
		collectible := &identity.ProfileShowcaseCollectiblePreference{}

		err := rows.Scan(
			&collectible.ContractAddress,
			&collectible.ChainID,
			&collectible.TokenID,
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

func (db sqlitePersistence) clearProfileShowcaseCollectiblePreferences(tx *sql.Tx) error {
	_, err := tx.Exec(clearProfileShowcaseCollectiblePreferencesQuery)
	return err
}

func (db sqlitePersistence) saveProfileShowcaseVerifiedTokenPreference(tx *sql.Tx, token *identity.ProfileShowcaseVerifiedTokenPreference) error {
	_, err := tx.Exec(upsertProfileShowcaseVerifiedTokenPreferenceQuery,
		token.Symbol,
		token.ShowcaseVisibility,
		token.Order,
	)

	return err
}

func (db sqlitePersistence) getProfileShowcaseVerifiedTokensPreferences(tx *sql.Tx) ([]*identity.ProfileShowcaseVerifiedTokenPreference, error) {
	rows, err := tx.Query(selectProfileShowcaseVerifiedTokenPreferenceQuery)
	if err != nil {
		return nil, err
	}

	tokens := []*identity.ProfileShowcaseVerifiedTokenPreference{}

	for rows.Next() {
		token := &identity.ProfileShowcaseVerifiedTokenPreference{}

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

func (db sqlitePersistence) clearProfileShowcaseVerifiedTokenPreferences(tx *sql.Tx) error {
	_, err := tx.Exec(clearProfileShowcaseVerifiedTokenPreferencesQuery)
	return err
}

func (db sqlitePersistence) saveProfileShowcaseUnverifiedTokenPreference(tx *sql.Tx, token *identity.ProfileShowcaseUnverifiedTokenPreference) error {
	_, err := tx.Exec(upsertProfileShowcaseUnverifiedTokenPreferenceQuery,
		token.ContractAddress,
		token.ChainID,
		token.ShowcaseVisibility,
		token.Order,
	)

	return err
}

func (db sqlitePersistence) getProfileShowcaseUnverifiedTokensPreferences(tx *sql.Tx) ([]*identity.ProfileShowcaseUnverifiedTokenPreference, error) {
	rows, err := tx.Query(selectProfileShowcaseUnverifiedTokenPreferenceQuery)
	if err != nil {
		return nil, err
	}

	tokens := []*identity.ProfileShowcaseUnverifiedTokenPreference{}

	for rows.Next() {
		token := &identity.ProfileShowcaseUnverifiedTokenPreference{}

		err := rows.Scan(
			&token.ContractAddress,
			&token.ChainID,
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

func (db sqlitePersistence) clearProfileShowcaseUnverifiedTokenPreferences(tx *sql.Tx) error {
	_, err := tx.Exec(clearProfileShowcaseUnverifiedTokenPreferencesQuery)
	return err
}

func (db sqlitePersistence) saveProfileShowcaseSocialLinkPreference(tx *sql.Tx, link *identity.ProfileShowcaseSocialLinkPreference) error {
	_, err := tx.Exec(upsertProfileShowcaseSocialLinkPreferenceQuery,
		link.URL,
		link.Text,
		link.ShowcaseVisibility,
		link.Order,
	)

	return err
}

func (db sqlitePersistence) getProfileShowcaseSocialLinkPreferences(tx *sql.Tx) ([]*identity.ProfileShowcaseSocialLinkPreference, error) {
	rows, err := tx.Query(selectProfileShowcaseSocialLinkPreferenceQuery)
	if err != nil {
		return nil, err
	}

	links := []*identity.ProfileShowcaseSocialLinkPreference{}

	for rows.Next() {
		link := &identity.ProfileShowcaseSocialLinkPreference{}

		err := rows.Scan(
			&link.URL,
			&link.Text,
			&link.ShowcaseVisibility,
			&link.Order,
		)

		if err != nil {
			return nil, err
		}

		links = append(links, link)
	}
	return links, nil
}

func (db sqlitePersistence) clearProfileShowcaseSocialLinkPreferences(tx *sql.Tx) error {
	_, err := tx.Exec(clearProfileShowcaseSocialLinkPreferencesQuery)
	return err
}

// Queries for the profile showcase for a contact
func (db sqlitePersistence) saveProfileShowcaseCommunityContact(tx *sql.Tx, contactID string, community *identity.ProfileShowcaseCommunity) error {
	_, err := tx.Exec(upsertContactProfileShowcaseCommunityQuery,
		contactID,
		community.CommunityID,
		community.Order,
		community.Grant,
	)

	return err
}

func (db sqlitePersistence) getProfileShowcaseCommunitiesContact(tx *sql.Tx, contactID string) ([]*identity.ProfileShowcaseCommunity, error) {
	rows, err := tx.Query(selectContactProfileShowcaseCommunityQuery, contactID)
	if err != nil {
		return nil, err
	}

	communities := []*identity.ProfileShowcaseCommunity{}

	for rows.Next() {
		community := &identity.ProfileShowcaseCommunity{}

		err := rows.Scan(&community.CommunityID, &community.Order, &community.Grant)
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

func (db sqlitePersistence) saveProfileShowcaseAccountContact(tx *sql.Tx, contactID string, account *identity.ProfileShowcaseAccount) error {
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

func (db sqlitePersistence) processProfileShowcaseAccounts(rows *sql.Rows) (result []*identity.ProfileShowcaseAccount, err error) {
	if rows == nil {
		return nil, errors.New("rows is nil")
	}

	for rows.Next() {
		account := &identity.ProfileShowcaseAccount{}
		err = rows.Scan(&account.Address, &account.Name, &account.ColorID, &account.Emoji, &account.Order, &account.ContactID)
		if err != nil {
			return
		}

		result = append(result, account)
	}

	err = rows.Err()
	return
}

func (db sqlitePersistence) getProfileShowcaseAccountsContact(tx *sql.Tx, contactID string) ([]*identity.ProfileShowcaseAccount, error) {
	rows, err := tx.Query(selectContactProfileShowcaseAccountQuery, contactID)
	if err != nil {
		return nil, err
	}

	return db.processProfileShowcaseAccounts(rows)
}

func (db sqlitePersistence) GetProfileShowcaseAccountsByAddress(address string) ([]*identity.ProfileShowcaseAccount, error) {
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

func (db sqlitePersistence) saveProfileShowcaseCollectibleContact(tx *sql.Tx, contactID string, collectible *identity.ProfileShowcaseCollectible) error {
	_, err := tx.Exec(upsertContactProfileShowcaseCollectibleQuery,
		contactID,
		collectible.ContractAddress,
		collectible.ChainID,
		collectible.TokenID,
		collectible.Order,
	)

	return err
}

func (db sqlitePersistence) getProfileShowcaseCollectiblesContact(tx *sql.Tx, contactID string) ([]*identity.ProfileShowcaseCollectible, error) {
	rows, err := tx.Query(selectContactProfileShowcaseCollectibleQuery, contactID)
	if err != nil {
		return nil, err
	}

	collectibles := []*identity.ProfileShowcaseCollectible{}

	for rows.Next() {
		collectible := &identity.ProfileShowcaseCollectible{}

		err := rows.Scan(
			&collectible.ContractAddress,
			&collectible.ChainID,
			&collectible.TokenID,
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

func (db sqlitePersistence) saveProfileShowcaseVerifiedTokenContact(tx *sql.Tx, contactID string, token *identity.ProfileShowcaseVerifiedToken) error {
	_, err := tx.Exec(upsertContactProfileShowcaseVerifiedTokenQuery,
		contactID,
		token.Symbol,
		token.Order,
	)

	return err
}

func (db sqlitePersistence) getProfileShowcaseVerifiedTokensContact(tx *sql.Tx, contactID string) ([]*identity.ProfileShowcaseVerifiedToken, error) {
	rows, err := tx.Query(selectContactProfileShowcaseVerifiedTokenQuery, contactID)
	if err != nil {
		return nil, err
	}

	tokens := []*identity.ProfileShowcaseVerifiedToken{}

	for rows.Next() {
		token := &identity.ProfileShowcaseVerifiedToken{}

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

func (db sqlitePersistence) clearProfileShowcaseVerifiedTokensContact(tx *sql.Tx, contactID string) error {
	_, err := tx.Exec(removeContactProfileShowcaseVerifiedTokenQuery, contactID)
	return err
}

func (db sqlitePersistence) saveProfileShowcaseUnverifiedTokenContact(tx *sql.Tx, contactID string, token *identity.ProfileShowcaseUnverifiedToken) error {
	_, err := tx.Exec(upsertContactProfileShowcaseUnverifiedTokenQuery,
		contactID,
		token.ContractAddress,
		token.ChainID,
		token.Order,
	)

	return err
}

func (db sqlitePersistence) getProfileShowcaseUnverifiedTokensContact(tx *sql.Tx, contactID string) ([]*identity.ProfileShowcaseUnverifiedToken, error) {
	rows, err := tx.Query(selectContactProfileShowcaseUnverifiedTokenQuery, contactID)
	if err != nil {
		return nil, err
	}

	tokens := []*identity.ProfileShowcaseUnverifiedToken{}

	for rows.Next() {
		token := &identity.ProfileShowcaseUnverifiedToken{}

		err := rows.Scan(
			&token.ContractAddress,
			&token.ChainID,
			&token.Order)
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
	}
	return tokens, nil
}

func (db sqlitePersistence) clearProfileShowcaseUnverifiedTokensContact(tx *sql.Tx, contactID string) error {
	_, err := tx.Exec(removeContactProfileShowcaseUnverifiedTokenQuery, contactID)
	return err
}

func (db sqlitePersistence) saveProfileShowcaseSocialLinkContact(tx *sql.Tx, contactID string, link *identity.ProfileShowcaseSocialLink) error {
	_, err := tx.Exec(upsertContactProfileShowcaseSocialLinkQuery,
		contactID,
		link.URL,
		link.Text,
		link.Order,
	)

	return err
}

func (db sqlitePersistence) getProfileShowcaseSocialLinksContact(tx *sql.Tx, contactID string) ([]*identity.ProfileShowcaseSocialLink, error) {
	rows, err := tx.Query(selectContactProfileShowcaseSocialLinkQuery, contactID)
	if err != nil {
		return nil, err
	}

	links := []*identity.ProfileShowcaseSocialLink{}

	for rows.Next() {
		link := &identity.ProfileShowcaseSocialLink{}

		err := rows.Scan(
			&link.URL,
			&link.Text,
			&link.Order)
		if err != nil {
			return nil, err
		}

		links = append(links, link)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return links, nil
}

func (db sqlitePersistence) clearProfileShowcaseSocialLinksContact(tx *sql.Tx, contactID string) error {
	_, err := tx.Exec(removeContactProfileShowcaseSocialLinkQuery, contactID)
	return err
}

// public functions
func (db sqlitePersistence) SaveProfileShowcasePreferences(preferences *identity.ProfileShowcasePreferences) error {
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

	err = db.clearProfileShowcaseCommunityPreferences(tx)
	if err != nil {
		return err
	}

	for _, community := range preferences.Communities {
		err = db.saveProfileShowcaseCommunityPreference(tx, community)
		if err != nil {
			return err
		}
	}

	err = db.clearProfileShowcaseAccountPreferences(tx)
	if err != nil {
		return err
	}

	for _, account := range preferences.Accounts {
		err = db.saveProfileShowcaseAccountPreference(tx, account)
		if err != nil {
			return err
		}
	}

	err = db.clearProfileShowcaseCollectiblePreferences(tx)
	if err != nil {
		return err
	}

	for _, collectible := range preferences.Collectibles {
		err = db.saveProfileShowcaseCollectiblePreference(tx, collectible)
		if err != nil {
			return err
		}
	}

	err = db.clearProfileShowcaseVerifiedTokenPreferences(tx)
	if err != nil {
		return err
	}

	for _, token := range preferences.VerifiedTokens {
		err = db.saveProfileShowcaseVerifiedTokenPreference(tx, token)
		if err != nil {
			return err
		}
	}

	err = db.clearProfileShowcaseUnverifiedTokenPreferences(tx)
	if err != nil {
		return err
	}

	for _, token := range preferences.UnverifiedTokens {
		err = db.saveProfileShowcaseUnverifiedTokenPreference(tx, token)
		if err != nil {
			return err
		}
	}

	err = db.clearProfileShowcaseSocialLinkPreferences(tx)
	if err != nil {
		return err
	}

	for _, link := range preferences.SocialLinks {
		err = db.saveProfileShowcaseSocialLinkPreference(tx, link)
		if err != nil {
			return err
		}
	}

	err = db.saveProfileShowcasePreferencesClock(tx, preferences.Clock)
	if err != nil {
		return err
	}

	return nil
}

func (db sqlitePersistence) SaveProfileShowcaseAccountPreference(account *identity.ProfileShowcaseAccountPreference) error {
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

func (db sqlitePersistence) GetProfileShowcasePreferences() (*identity.ProfileShowcasePreferences, error) {
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

	clock, err := db.getProfileShowcasePreferencesClock(tx)
	if err != nil {
		return nil, err
	}

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

	socialLinks, err := db.getProfileShowcaseSocialLinkPreferences(tx)
	if err != nil {
		return nil, err
	}

	return &identity.ProfileShowcasePreferences{
		Clock:            clock,
		Communities:      communities,
		Accounts:         accounts,
		Collectibles:     collectibles,
		VerifiedTokens:   verifiedTokens,
		UnverifiedTokens: unverifiedTokens,
		SocialLinks:      socialLinks,
	}, nil
}

func (db sqlitePersistence) SaveProfileShowcaseForContact(showcase *identity.ProfileShowcase) error {
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

	for _, link := range showcase.SocialLinks {
		err = db.saveProfileShowcaseSocialLinkContact(tx, showcase.ContactID, link)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db sqlitePersistence) GetProfileShowcaseForContact(contactID string) (*identity.ProfileShowcase, error) {
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

	socialLinks, err := db.getProfileShowcaseSocialLinksContact(tx, contactID)
	if err != nil {
		return nil, err
	}

	return &identity.ProfileShowcase{
		ContactID:        contactID,
		Communities:      communities,
		Accounts:         accounts,
		Collectibles:     collectibles,
		VerifiedTokens:   verifiedTokens,
		UnverifiedTokens: unverifiedTokens,
		SocialLinks:      socialLinks,
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

	err = db.clearProfileShowcaseSocialLinksContact(tx, contactID)
	if err != nil {
		return err
	}

	return nil
}
