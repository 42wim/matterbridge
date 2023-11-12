package walletconnect

import (
	"database/sql"
	"errors"
)

type DbSession struct {
	Topic           Topic  `json:"topic"`
	PairingTopic    Topic  `json:"pairingTopic"`
	Expiry          int64  `json:"expiry"`
	Active          bool   `json:"active"`
	DappName        string `json:"dappName"`
	DappURL         string `json:"dappUrl"`
	DappDescription string `json:"dappDescription"`
	DappIcon        string `json:"dappIcon"`
	DappVerifyURL   string `json:"dappVerifyUrl"`
	DappPublicKey   string `json:"dappPublicKey"`
}

func UpsertSession(db *sql.DB, session DbSession) error {
	insertSQL := `
	INSERT OR IGNORE INTO
		wallet_connect_sessions (topic, pairing_topic, expiry, active)
	VALUES
		(?, ?, ?, ?);

	UPDATE
		wallet_connect_sessions
	SET
		pairing_topic = ?,
		expiry = ?,
		active = ?,
		dapp_name = ?,
		dapp_url = ?,
		dapp_description = ?,
		dapp_icon = ?,
		dapp_verify_url = ?,
		dapp_publicKey = ?
	WHERE
		topic = ?;`

	_, err := db.Exec(insertSQL,
		session.Topic,
		session.PairingTopic,
		session.Expiry,
		session.Active,
		session.PairingTopic,
		session.Expiry,
		session.Active,
		session.DappName,
		session.DappURL,
		session.DappDescription,
		session.DappIcon,
		session.DappVerifyURL,
		session.DappPublicKey,
		session.Topic,
	)
	return err
}

func ChangeSessionState(db *sql.DB, topic Topic, active bool) error {
	stmt, err := db.Prepare("UPDATE wallet_connect_sessions SET active = ? WHERE topic = ?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(active, topic)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("unable to locate session for DB state change")
	}

	return nil
}

func GetSessionByTopic(db *sql.DB, topic Topic) (*DbSession, error) {
	querySQL := `
	SELECT *
	FROM
		wallet_connect_sessions
	WHERE
		topic = ?`

	row := db.QueryRow(querySQL, topic)

	var session DbSession
	err := row.Scan(&session.Topic,
		&session.PairingTopic,
		&session.Expiry,
		&session.Active,
		&session.DappName,
		&session.DappURL,
		&session.DappDescription,
		&session.DappIcon,
		&session.DappVerifyURL,
		&session.DappPublicKey)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func GetSessionsByPairingTopic(db *sql.DB, pairingTopic Topic) ([]DbSession, error) {
	querySQL := `
	SELECT *
	FROM
		wallet_connect_sessions
	WHERE
		pairing_topic = ?`

	rows, err := db.Query(querySQL, pairingTopic)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]DbSession, 0, 2)
	for rows.Next() {
		var session DbSession
		err := rows.Scan(&session.Topic,
			&session.PairingTopic,
			&session.Expiry,
			&session.Active,
			&session.DappName,
			&session.DappURL,
			&session.DappDescription,
			&session.DappIcon,
			&session.DappVerifyURL,
			&session.DappPublicKey)
		if err != nil {
			return nil, err
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// GetActiveSessions returns all active sessions (active and not expired) that have an expiry timestamp newer or equal to the given timestamp.
func GetActiveSessions(db *sql.DB, expiryNotOlderThanTimestamp int64) ([]DbSession, error) {
	querySQL := `
	SELECT *
	FROM
		wallet_connect_sessions
	WHERE
		active != 0 AND
		expiry >= ?
	ORDER BY
		expiry DESC`

	rows, err := db.Query(querySQL, expiryNotOlderThanTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]DbSession, 0, 2)
	for rows.Next() {
		var session DbSession
		err := rows.Scan(&session.Topic,
			&session.PairingTopic,
			&session.Expiry,
			&session.Active,
			&session.DappName,
			&session.DappURL,
			&session.DappDescription,
			&session.DappIcon,
			&session.DappVerifyURL,
			&session.DappPublicKey)
		if err != nil {
			return nil, err
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}
