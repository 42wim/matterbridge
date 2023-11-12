package node

import (
	"database/sql"

	"github.com/status-im/mvds/state"
)

type epochSQLitePersistence struct {
	db *sql.DB
}

func newEpochSQLitePersistence(db *sql.DB) *epochSQLitePersistence {
	return &epochSQLitePersistence{db: db}
}

func (p *epochSQLitePersistence) Get(nodeID state.PeerID) (epoch int64, err error) {
	row := p.db.QueryRow(`SELECT epoch FROM mvds_epoch WHERE peer_id = ?`, nodeID[:])
	err = row.Scan(&epoch)
	if err == sql.ErrNoRows {
		err = nil
	}
	return
}

func (p *epochSQLitePersistence) Set(nodeID state.PeerID, epoch int64) error {
	_, err := p.db.Exec(`
		INSERT OR REPLACE INTO mvds_epoch (peer_id, epoch) VALUES (?, ?)`,
		nodeID[:],
		epoch,
	)
	return err
}
