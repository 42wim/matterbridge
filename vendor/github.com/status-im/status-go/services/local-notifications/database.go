package localnotifications

import (
	"database/sql"
)

type Database struct {
	db      *sql.DB
	network uint64
}

type NotificationPreference struct {
	Enabled    bool   `json:"enabled"`
	Service    string `json:"service"`
	Event      string `json:"event,omitempty"`
	Identifier string `json:"identifier,omitempty"`
}

func NewDB(db *sql.DB, network uint64) *Database {
	return &Database{db: db, network: network}
}

func (db *Database) GetPreferences() (rst []NotificationPreference, err error) {
	rows, err := db.db.Query("SELECT service, event, identifier, enabled FROM local_notifications_preferences")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		pref := NotificationPreference{}
		err = rows.Scan(&pref.Service, &pref.Event, &pref.Identifier, &pref.Enabled)
		if err != nil {
			return nil, err
		}
		rst = append(rst, pref)
	}
	return rst, nil
}

func (db *Database) GetWalletPreference() (rst NotificationPreference, err error) {
	pref := db.db.QueryRow("SELECT service, event, identifier, enabled FROM local_notifications_preferences WHERE service = 'wallet' AND event = 'transaction' AND identifier = 'all'")

	err = pref.Scan(&rst.Service, &rst.Event, &rst.Identifier, &rst.Enabled)
	if err == sql.ErrNoRows {
		return rst, nil
	}
	return
}

func (db *Database) ChangeWalletPreference(preference bool) error {
	_, err := db.db.Exec("INSERT OR REPLACE INTO local_notifications_preferences (service, event, identifier, enabled) VALUES ('wallet', 'transaction', 'all', ?)", preference)
	return err
}
