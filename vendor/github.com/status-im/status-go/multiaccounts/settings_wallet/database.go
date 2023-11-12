package walletsettings

import (
	"database/sql"
	"errors"
)

type TokenPreferences struct {
	Key           string `json:"key"`
	Position      int    `json:"position"`
	GroupPosition int    `json:"groupPosition"`
	Visible       bool   `json:"visible"`
	CommunityID   string `json:"communityId"`
}

type CollectiblePreferencesType int

const (
	CollectiblePreferencesTypeNonCommunityCollectible CollectiblePreferencesType = iota + 1
	CollectiblePreferencesTypeCommunityCollectible
	CollectiblePreferencesTypeCollection
	CollectiblePreferencesTypeCommunity
)

type CollectiblePreferences struct {
	Type     CollectiblePreferencesType `json:"type"`
	Key      string                     `json:"key"`
	Position int                        `json:"position"`
	Visible  bool                       `json:"visible"`
}

type WalletSettings struct {
	db *sql.DB
}

func NewWalletSettings(db *sql.DB) *WalletSettings {
	return &WalletSettings{
		db: db,
	}
}

// This function should not be used directly, it is called from the functions which update token preferences.
func (ws *WalletSettings) setClockOfLastTokenPreferencesChange(tx *sql.Tx, clock uint64) error {
	if tx == nil {
		return errors.New("database transaction is nil")
	}
	_, err := tx.Exec("UPDATE settings SET wallet_token_preferences_change_clock = ? WHERE synthetic_id = 'id'", clock)
	return err
}

func (ws *WalletSettings) GetClockOfLastTokenPreferencesChange() (result uint64, err error) {
	query := "SELECT wallet_token_preferences_change_clock FROM settings WHERE synthetic_id = 'id'"
	err = ws.db.QueryRow(query).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, err
}

func (ws *WalletSettings) UpdateTokenPreferences(preferences []TokenPreferences, groupByCommunity bool, testNetworksEnabled bool, clock uint64) error {
	if len(preferences) == 0 {
		return errors.New("tokens: trying to create custom order with empty list")
	}

	tx, err := ws.db.Begin()
	if err != nil {
		return err
	}

	var mainError error = nil

	defer func() {
		if mainError == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	_, mainError = tx.Exec("DELETE FROM token_preferences WHERE testnet = ?", testNetworksEnabled)
	if mainError != nil {
		return mainError
	}

	for _, p := range preferences {
		if p.Position < 0 {
			mainError = errors.New("tokens: trying to create custom order with negative position")
			return mainError
		}
		_, err := tx.Exec("INSERT INTO token_preferences (key, position, group_position, visible, community_id, testnet) VALUES (?, ?, ?, ?, ?, ?)", p.Key, p.Position, p.GroupPosition, p.Visible, p.CommunityID, testNetworksEnabled)
		if err != nil {
			mainError = err
			return err
		}
	}

	if groupByCommunity {
		// Find community tokens without group position
		// Group position can be -1 if it wasn't created yet. Values must be consitstent across all tokens
		rows, err := tx.Query(`SELECT COUNT(*) FROM token_preferences WHERE testnet = ? AND group_position = -1 AND community_id != '' AND visible GROUP BY community_id HAVING COUNT(*) > 0`, testNetworksEnabled)
		if err != nil {
			mainError = err
			return err
		}
		if rows.Next() {
			mainError = errors.New("tokens: not all community tokens have assigned the group position")
			return mainError
		}
	}

	mainError = ws.setClockOfLastTokenPreferencesChange(tx, clock)
	if mainError != nil {
		return mainError
	}
	return nil
}

func (ws *WalletSettings) GetTokenPreferences(testNetworksEnabled bool) ([]TokenPreferences, error) {
	rows, err := ws.db.Query("SELECT key, position, group_position, visible, community_id FROM token_preferences WHERE testnet = ?", testNetworksEnabled)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TokenPreferences

	for rows.Next() {
		token := TokenPreferences{}
		err := rows.Scan(&token.Key, &token.Position, &token.GroupPosition, &token.Visible, &token.CommunityID)
		if err != nil {
			return nil, err
		}

		result = append(result, token)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (ws *WalletSettings) setClockOfLastCollectiblePreferencesChange(tx *sql.Tx, clock uint64) error {
	if tx == nil {
		return errors.New("database transaction is nil")
	}
	_, err := tx.Exec("UPDATE settings SET wallet_collectible_preferences_change_clock = ? WHERE synthetic_id = 'id'", clock)
	return err
}

func (ws *WalletSettings) GetClockOfLastCollectiblePreferencesChange() (result uint64, err error) {
	query := "SELECT wallet_collectible_preferences_change_clock FROM settings WHERE synthetic_id = 'id'"
	err = ws.db.QueryRow(query).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, err
}

func (ws *WalletSettings) UpdateCollectiblePreferences(preferences []CollectiblePreferences, groupByCommunity bool, groupByCollection bool, testNetworksEnabled bool, clock uint64) error {
	if len(preferences) == 0 {
		return errors.New("collectibles: trying to create custom order with empty list")
	}

	tx, err := ws.db.Begin()
	if err != nil {
		return err
	}

	var mainError error = nil

	defer func() {
		if mainError == nil {
			err = tx.Commit()
			return
		}
		_ = tx.Rollback()
	}()

	_, mainError = tx.Exec("DELETE FROM collectible_preferences WHERE testnet = ?", testNetworksEnabled)
	if mainError != nil {
		return mainError
	}

	for _, p := range preferences {
		if p.Position < 0 {
			mainError = errors.New("collectibles: trying to create custom order with negative position")
			return mainError
		}
		_, err := tx.Exec("INSERT INTO collectible_preferences (type, key, position, visible, testnet) VALUES (?, ?, ?, ?, ?)", p.Type, p.Key, p.Position, p.Visible, testNetworksEnabled)
		if err != nil {
			mainError = err
			return err
		}
	}

	mainError = ws.setClockOfLastCollectiblePreferencesChange(tx, clock)
	if mainError != nil {
		return mainError
	}
	return nil
}

func (ws *WalletSettings) GetCollectiblePreferences(testNetworksEnabled bool) ([]CollectiblePreferences, error) {
	rows, err := ws.db.Query("SELECT type, key, position, visible FROM collectible_preferences WHERE testnet = ?", testNetworksEnabled)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []CollectiblePreferences

	for rows.Next() {
		p := CollectiblePreferences{}
		err := rows.Scan(&p.Type, &p.Key, &p.Position, &p.Visible)
		if err != nil {
			return nil, err
		}

		result = append(result, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
