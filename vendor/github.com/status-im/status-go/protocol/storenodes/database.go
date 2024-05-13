package storenodes

import (
	"bytes"
	"database/sql"
	"fmt"
	"time"

	"github.com/status-im/status-go/eth-node/types"
)

type Database struct {
	db *sql.DB
}

func NewDB(db *sql.DB) *Database {
	return &Database{db: db}
}

// syncSave will sync the storenodes in the DB from the snode slice
//   - if a storenode is not in the provided list, it will be soft-deleted
//   - if a storenode is in the provided list, it will be inserted or updated
func (d *Database) syncSave(communityID types.HexBytes, snode []Storenode, clock uint64) (err error) {
	var tx *sql.Tx
	tx, err = d.db.Begin()
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

	now := time.Now().Unix()
	dbNodes, err := d.getByCommunityID(communityID, tx)
	if err != nil {
		return fmt.Errorf("getting storenodes by community id: %w", err)
	}
	// Soft-delete db nodes that are not in the provided list
	for _, dbN := range dbNodes {
		if find(dbN, snode) != nil {
			continue
		}
		if clock != 0 && dbN.Clock >= clock {
			continue
		}
		if err := d.softDelete(communityID, dbN.StorenodeID, now, tx); err != nil {
			return fmt.Errorf("soft deleting existing storenodes: %w", err)
		}

	}
	// Insert or update the nodes in the provided list
	for _, n := range snode {
		// defensively validate the communityID
		if len(n.CommunityID) == 0 || !bytes.Equal(communityID, n.CommunityID) {
			err = fmt.Errorf("communityID mismatch %v != %v", communityID, n.CommunityID)
			return err
		}
		dbN := find(n, dbNodes)
		if dbN != nil && n.Clock != 0 && dbN.Clock >= n.Clock {
			continue
		}
		if err := d.upsert(n, tx); err != nil {
			return fmt.Errorf("upserting storenodes: %w", err)
		}
	}
	// TODO for now only allow one storenode per community
	count, err := d.countByCommunity(communityID, tx)
	if err != nil {
		return err
	}
	if count > 1 {
		err = fmt.Errorf("only one storenode per community is allowed")
		return err
	}
	return nil
}

func (d *Database) getAll() ([]Storenode, error) {
	rows, err := d.db.Query(`
		SELECT community_id, storenode_id, name, address, fleet, version, clock, removed, deleted_at
		FROM community_storenodes
		WHERE removed = 0
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return toStorenodes(rows)
}

func (d *Database) getByCommunityID(communityID types.HexBytes, tx ...*sql.Tx) ([]Storenode, error) {
	var rows *sql.Rows
	var err error
	q := `
	SELECT community_id, storenode_id, name, address, fleet, version, clock, removed, deleted_at
	FROM community_storenodes
	WHERE community_id = ? AND removed = 0
`
	if len(tx) > 0 {
		rows, err = tx[0].Query(q, communityID)
	} else {
		rows, err = d.db.Query(q, communityID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return toStorenodes(rows)
}

func (d *Database) softDelete(communityID types.HexBytes, storenodeID string, deletedAt int64, tx *sql.Tx) error {
	_, err := tx.Exec("UPDATE community_storenodes SET removed = 1, deleted_at = ? WHERE community_id = ? AND storenode_id = ?", deletedAt, communityID, storenodeID)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) upsert(n Storenode, tx *sql.Tx) error {
	_, err := tx.Exec(`INSERT OR REPLACE INTO community_storenodes(
		community_id,
		storenode_id,
		name,
		address,
		fleet,
		version,
		clock,
		removed,
		deleted_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.CommunityID,
		n.StorenodeID,
		n.Name,
		n.Address,
		n.Fleet,
		n.Version,
		n.Clock,
		n.Removed,
		n.DeletedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) countByCommunity(communityID types.HexBytes, tx *sql.Tx) (int, error) {
	var count int
	err := tx.QueryRow(`SELECT COUNT(*) FROM community_storenodes WHERE community_id = ? AND removed = 0`, communityID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func toStorenodes(rows *sql.Rows) ([]Storenode, error) {
	var result []Storenode

	for rows.Next() {
		var m Storenode
		if err := rows.Scan(
			&m.CommunityID,
			&m.StorenodeID,
			&m.Name,
			&m.Address,
			&m.Fleet,
			&m.Version,
			&m.Clock,
			&m.Removed,
			&m.DeletedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, m)
	}

	return result, nil
}

func find(n Storenode, nodes []Storenode) *Storenode {
	for i, node := range nodes {
		if node.StorenodeID == n.StorenodeID && bytes.Equal(node.CommunityID, n.CommunityID) {
			return &nodes[i]
		}
	}
	return nil
}
