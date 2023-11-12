package peers

import (
	"database/sql"
	"errors"

	"github.com/status-im/mvds/state"
)

var (
	ErrPeerNotFound = errors.New("peer not found")
)

type sqlitePersistence struct {
	db *sql.DB
}

func NewSQLitePersistence(db *sql.DB) sqlitePersistence {
	return sqlitePersistence{db: db}
}

func (p sqlitePersistence) Add(groupID state.GroupID, peerID state.PeerID) error {
	_, err := p.db.Exec(`INSERT INTO mvds_peers (group_id, peer_id) VALUES (?, ?)`, groupID[:], peerID[:])
	return err
}

func (p sqlitePersistence) Exists(groupID state.GroupID, peerID state.PeerID) (bool, error) {
	var result bool
	err := p.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM mvds_peers WHERE group_id = ? AND peer_id = ?)`,
		groupID[:],
		peerID[:],
	).Scan(&result)
	switch err {
	case sql.ErrNoRows:
		return false, ErrPeerNotFound
	case nil:
		return result, nil
	default:
		return false, err
	}
}

func (p sqlitePersistence) GetByGroupID(groupID state.GroupID) (result []state.PeerID, err error) {
	rows, err := p.db.Query(`SELECT peer_id FROM mvds_peers WHERE group_id = ?`, groupID[:])
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			peerIDBytes []byte
			peerID      state.PeerID
		)
		if err := rows.Scan(&peerIDBytes); err != nil {
			return nil, err
		}
		copy(peerID[:], peerIDBytes)
		result = append(result, peerID)
	}
	return
}
