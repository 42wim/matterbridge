package community

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/status-im/status-go/services/wallet/thirdparty"
)

type DataDB struct {
	db *sql.DB
}

func NewDataDB(sqlDb *sql.DB) *DataDB {
	return &DataDB{
		db: sqlDb,
	}
}

type InfoState struct {
	LastUpdateTimestamp uint64
	LastUpdateSuccesful bool
}

const communityInfoColumns = "id, name, color, image, image_payload"
const selectCommunityInfoColumns = "name, color, image, image_payload"

const communityInfoStateColumns = "id, last_update_timestamp, last_update_successful"
const selectCommunityInfoStateColumns = "last_update_timestamp, last_update_successful"

func (o *DataDB) SetCommunityInfo(id string, c *thirdparty.CommunityInfo) (err error) {
	tx, err := o.db.Begin()
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

	setState, err := tx.Prepare(fmt.Sprintf(`INSERT OR REPLACE INTO community_data_cache_state (%s) 
		VALUES (?, ?, ?)`, communityInfoStateColumns))
	if err != nil {
		return err
	}

	valid := c != nil
	_, err = setState.Exec(
		id,
		time.Now().Unix(),
		valid,
	)
	if err != nil {
		return err
	}

	if valid {
		setInfo, err := tx.Prepare(fmt.Sprintf(`INSERT OR REPLACE INTO community_data_cache (%s) 
			VALUES (?, ?, ?, ?, ?)`, communityInfoColumns))
		if err != nil {
			return err
		}

		_, err = setInfo.Exec(
			id,
			c.CommunityName,
			c.CommunityColor,
			c.CommunityImage,
			c.CommunityImagePayload,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *DataDB) GetCommunityInfo(id string) (*thirdparty.CommunityInfo, *InfoState, error) {
	if id == "" {
		return nil, nil, nil
	}

	var info thirdparty.CommunityInfo
	var state InfoState
	var row *sql.Row

	getState, err := o.db.Prepare(fmt.Sprintf(`SELECT %s
	FROM community_data_cache_state 
	WHERE id=?`, selectCommunityInfoStateColumns))
	if err != nil {
		return nil, nil, err
	}
	row = getState.QueryRow(id)

	err = row.Scan(
		&state.LastUpdateTimestamp,
		&state.LastUpdateSuccesful,
	)

	if err == sql.ErrNoRows {
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, err
	}

	getInfo, err := o.db.Prepare(fmt.Sprintf(`SELECT %s
		FROM community_data_cache
		WHERE id=?`, selectCommunityInfoColumns))
	if err != nil {
		return nil, nil, err
	}

	row = getInfo.QueryRow(id)

	err = row.Scan(
		&info.CommunityName,
		&info.CommunityColor,
		&info.CommunityImage,
		&info.CommunityImagePayload,
	)

	if err == sql.ErrNoRows {
		return nil, &state, nil
	} else if err != nil {
		return nil, nil, err
	}

	return &info, &state, nil
}
