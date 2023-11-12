package communities

import (
	"database/sql"

	"github.com/status-im/status-go/protocol/protobuf"
)

type RawCommunityRow struct {
	ID           []byte
	PrivateKey   []byte
	Description  []byte
	Joined       bool
	JoinedAt     int64
	Spectated    bool
	Verified     bool
	SyncedAt     uint64
	Muted        bool
	LastOpenedAt int64
}

func fromSyncCommunityProtobuf(syncCommProto *protobuf.SyncInstallationCommunity) RawCommunityRow {
	return RawCommunityRow{
		ID:           syncCommProto.Id,
		Description:  syncCommProto.Description,
		Joined:       syncCommProto.Joined,
		JoinedAt:     syncCommProto.JoinedAt,
		Spectated:    syncCommProto.Spectated,
		Verified:     syncCommProto.Verified,
		SyncedAt:     syncCommProto.Clock,
		Muted:        syncCommProto.Muted,
		LastOpenedAt: syncCommProto.LastOpenedAt,
	}
}

func (p *Persistence) scanRowToStruct(rowScan func(dest ...interface{}) error) (*RawCommunityRow, error) {
	rcr := new(RawCommunityRow)
	var syncedAt, muteTill sql.NullTime

	err := rowScan(
		&rcr.ID,
		&rcr.PrivateKey,
		&rcr.Description,
		&rcr.Joined,
		&rcr.JoinedAt,
		&rcr.Verified,
		&rcr.Spectated,
		&rcr.Muted,
		&muteTill,
		&syncedAt,
		&rcr.LastOpenedAt,
	)
	if syncedAt.Valid {
		rcr.SyncedAt = uint64(syncedAt.Time.Unix())
	}

	if err != nil {
		return nil, err
	}

	return rcr, nil
}

func (p *Persistence) getAllCommunitiesRaw() (rcrs []*RawCommunityRow, err error) {
	var rows *sql.Rows
	// Keep "*", if the db table is updated, syncing needs to match, this fail will force us to update syncing.
	rows, err = p.db.Query(`SELECT id, private_key, description, joined, joined_at, verified, spectated, muted, muted_till, synced_at, last_opened_at FROM communities_communities`)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			// Don't shadow original error
			_ = rows.Close()
			return

		}
		err = rows.Close()
	}()

	for rows.Next() {
		rcr, err := p.scanRowToStruct(rows.Scan)
		if err != nil {
			return nil, err
		}

		rcrs = append(rcrs, rcr)
	}
	return rcrs, nil
}

func (p *Persistence) getRawCommunityRow(id []byte) (*RawCommunityRow, error) {
	qr := p.db.QueryRow(`SELECT id, private_key, description, joined, joined_at, verified, spectated, muted, muted_till, synced_at, last_opened_at FROM communities_communities WHERE id = ?`, id)
	return p.scanRowToStruct(qr.Scan)
}

func (p *Persistence) getSyncedRawCommunity(id []byte) (*RawCommunityRow, error) {
	qr := p.db.QueryRow(`SELECT id, private_key, description, joined, joined_at, verified, spectated, muted, muted_till, synced_at, last_opened_at FROM communities_communities WHERE id = ? AND synced_at > 0`, id)
	return p.scanRowToStruct(qr.Scan)
}

func (p *Persistence) saveRawCommunityRow(rawCommRow RawCommunityRow) error {
	_, err := p.db.Exec(
		`INSERT INTO communities_communities ("id", "private_key", "description", "joined", "joined_at", "verified", "synced_at", "muted", "last_opened_at") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rawCommRow.ID,
		rawCommRow.PrivateKey,
		rawCommRow.Description,
		rawCommRow.Joined,
		rawCommRow.JoinedAt,
		rawCommRow.Verified,
		rawCommRow.SyncedAt,
		rawCommRow.Muted,
		rawCommRow.LastOpenedAt,
	)
	return err
}

func (p *Persistence) saveRawCommunityRowWithoutSyncedAt(rawCommRow RawCommunityRow) error {
	_, err := p.db.Exec(
		`INSERT INTO communities_communities ("id", "private_key", "description", "joined", "joined_at", "verified", "muted", "last_opened_at") VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		rawCommRow.ID,
		rawCommRow.PrivateKey,
		rawCommRow.Description,
		rawCommRow.Joined,
		rawCommRow.JoinedAt,
		rawCommRow.Verified,
		rawCommRow.Muted,
		rawCommRow.LastOpenedAt,
	)
	return err
}
