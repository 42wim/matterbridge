package collectibles

import (
	"database/sql"
	"fmt"
	"math/big"

	"github.com/status-im/status-go/protocol/communities/token"
	"github.com/status-im/status-go/services/wallet/bigint"
	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/sqlite"
)

type CollectibleDataDB struct {
	db *sql.DB
}

func NewCollectibleDataDB(sqlDb *sql.DB) *CollectibleDataDB {
	return &CollectibleDataDB{
		db: sqlDb,
	}
}

const collectibleDataColumns = "chain_id, contract_address, token_id, provider, name, description, permalink, image_url, image_payload, animation_url, animation_media_type, background_color, token_uri, community_id"
const collectibleCommunityDataColumns = "community_privileges_level"
const collectibleTraitsColumns = "chain_id, contract_address, token_id, trait_type, trait_value, display_type, max_value"
const selectCollectibleTraitsColumns = "trait_type, trait_value, display_type, max_value"

func rowsToCollectibleTraits(rows *sql.Rows) ([]thirdparty.CollectibleTrait, error) {
	var traits []thirdparty.CollectibleTrait = make([]thirdparty.CollectibleTrait, 0)
	for rows.Next() {
		var trait thirdparty.CollectibleTrait
		err := rows.Scan(
			&trait.TraitType,
			&trait.Value,
			&trait.DisplayType,
			&trait.MaxValue,
		)
		if err != nil {
			return nil, err
		}
		traits = append(traits, trait)
	}
	return traits, nil
}

func getCollectibleTraits(creator sqlite.StatementCreator, id thirdparty.CollectibleUniqueID) ([]thirdparty.CollectibleTrait, error) {
	// Get traits list
	selectTraits, err := creator.Prepare(fmt.Sprintf(`SELECT %s
		FROM collectible_traits_cache
		WHERE chain_id = ? AND contract_address = ? AND token_id = ?`, selectCollectibleTraitsColumns))
	if err != nil {
		return nil, err
	}

	rows, err := selectTraits.Query(
		id.ContractID.ChainID,
		id.ContractID.Address,
		(*bigint.SQLBigIntBytes)(id.TokenID.Int),
	)
	if err != nil {
		return nil, err
	}

	return rowsToCollectibleTraits(rows)
}

func upsertCollectibleTraits(creator sqlite.StatementCreator, id thirdparty.CollectibleUniqueID, traits []thirdparty.CollectibleTrait) error {
	// Remove old traits list
	deleteTraits, err := creator.Prepare(`DELETE FROM collectible_traits_cache WHERE chain_id = ? AND contract_address = ? AND token_id = ?`)
	if err != nil {
		return err
	}

	_, err = deleteTraits.Exec(
		id.ContractID.ChainID,
		id.ContractID.Address,
		(*bigint.SQLBigIntBytes)(id.TokenID.Int),
	)
	if err != nil {
		return err
	}

	// Insert new traits list
	insertTrait, err := creator.Prepare(fmt.Sprintf(`INSERT INTO collectible_traits_cache (%s)
																				VALUES (?, ?, ?, ?, ?, ?, ?)`, collectibleTraitsColumns))
	if err != nil {
		return err
	}

	for _, t := range traits {
		_, err = insertTrait.Exec(
			id.ContractID.ChainID,
			id.ContractID.Address,
			(*bigint.SQLBigIntBytes)(id.TokenID.Int),
			t.TraitType,
			t.Value,
			t.DisplayType,
			t.MaxValue,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func setCollectiblesData(creator sqlite.StatementCreator, collectibles []thirdparty.CollectibleData, allowUpdate bool) error {
	insertCollectible, err := creator.Prepare(fmt.Sprintf(`%s INTO collectible_data_cache (%s) 
																				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, insertStatement(allowUpdate), collectibleDataColumns))
	if err != nil {
		return err
	}

	for _, c := range collectibles {
		_, err = insertCollectible.Exec(
			c.ID.ContractID.ChainID,
			c.ID.ContractID.Address,
			(*bigint.SQLBigIntBytes)(c.ID.TokenID.Int),
			c.Provider,
			c.Name,
			c.Description,
			c.Permalink,
			c.ImageURL,
			c.ImagePayload,
			c.AnimationURL,
			c.AnimationMediaType,
			c.BackgroundColor,
			c.TokenURI,
			c.CommunityID,
		)
		if err != nil {
			return err
		}

		err = upsertContractType(creator, c.ID.ContractID, c.ContractType)
		if err != nil {
			return err
		}

		if allowUpdate {
			err = upsertCollectibleTraits(creator, c.ID, c.Traits)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (o *CollectibleDataDB) SetData(collectibles []thirdparty.CollectibleData, allowUpdate bool) (err error) {
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

	// Insert new collectibles data
	err = setCollectiblesData(tx, collectibles, allowUpdate)
	if err != nil {
		return err
	}

	return
}

func scanCollectiblesDataRow(row *sql.Row) (*thirdparty.CollectibleData, error) {
	c := thirdparty.CollectibleData{
		ID: thirdparty.CollectibleUniqueID{
			TokenID: &bigint.BigInt{Int: big.NewInt(0)},
		},
		Traits: make([]thirdparty.CollectibleTrait, 0),
	}
	err := row.Scan(
		&c.ID.ContractID.ChainID,
		&c.ID.ContractID.Address,
		(*bigint.SQLBigIntBytes)(c.ID.TokenID.Int),
		&c.Provider,
		&c.Name,
		&c.Description,
		&c.Permalink,
		&c.ImageURL,
		&c.ImagePayload,
		&c.AnimationURL,
		&c.AnimationMediaType,
		&c.BackgroundColor,
		&c.TokenURI,
		&c.CommunityID,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (o *CollectibleDataDB) GetIDsNotInDB(ids []thirdparty.CollectibleUniqueID) ([]thirdparty.CollectibleUniqueID, error) {
	ret := make([]thirdparty.CollectibleUniqueID, 0, len(ids))
	idMap := make(map[string]thirdparty.CollectibleUniqueID, len(ids))

	// Ensure we don't have duplicates
	for _, id := range ids {
		idMap[id.HashKey()] = id
	}

	exists, err := o.db.Prepare(`SELECT EXISTS (
			SELECT 1 FROM collectible_data_cache
			WHERE chain_id=? AND contract_address=? AND token_id=?
		)`)
	if err != nil {
		return nil, err
	}

	for _, id := range idMap {
		row := exists.QueryRow(
			id.ContractID.ChainID,
			id.ContractID.Address,
			(*bigint.SQLBigIntBytes)(id.TokenID.Int),
		)
		var exists bool
		err = row.Scan(&exists)
		if err != nil {
			return nil, err
		}
		if !exists {
			ret = append(ret, id)
		}
	}

	return ret, nil
}

func (o *CollectibleDataDB) GetData(ids []thirdparty.CollectibleUniqueID) (map[string]thirdparty.CollectibleData, error) {
	ret := make(map[string]thirdparty.CollectibleData)

	getData, err := o.db.Prepare(fmt.Sprintf(`SELECT %s
		FROM collectible_data_cache
		WHERE chain_id=? AND contract_address=? AND token_id=?`, collectibleDataColumns))
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		row := getData.QueryRow(
			id.ContractID.ChainID,
			id.ContractID.Address,
			(*bigint.SQLBigIntBytes)(id.TokenID.Int),
		)
		c, err := scanCollectiblesDataRow(row)
		if err == sql.ErrNoRows {
			continue
		} else if err != nil {
			return nil, err
		} else {
			// Get traits from different table
			c.Traits, err = getCollectibleTraits(o.db, c.ID)
			if err != nil {
				return nil, err
			}

			// Get contract type from different table
			c.ContractType, err = readContractType(o.db, c.ID.ContractID)
			if err != nil {
				return nil, err
			}

			ret[c.ID.HashKey()] = *c
		}
	}
	return ret, nil
}

func (o *CollectibleDataDB) SetCommunityInfo(id thirdparty.CollectibleUniqueID, communityInfo thirdparty.CollectibleCommunityInfo) (err error) {
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

	update, err := tx.Prepare(`UPDATE collectible_data_cache 
		SET community_privileges_level=?
		WHERE chain_id=? AND contract_address=? AND token_id=?`)
	if err != nil {
		return err
	}

	_, err = update.Exec(
		communityInfo.PrivilegesLevel,
		id.ContractID.ChainID,
		id.ContractID.Address,
		(*bigint.SQLBigIntBytes)(id.TokenID.Int),
	)

	return err
}

func (o *CollectibleDataDB) GetCommunityInfo(id thirdparty.CollectibleUniqueID) (*thirdparty.CollectibleCommunityInfo, error) {
	ret := thirdparty.CollectibleCommunityInfo{
		PrivilegesLevel: token.CommunityLevel,
	}

	getData, err := o.db.Prepare(fmt.Sprintf(`SELECT %s
		FROM collectible_data_cache
		WHERE chain_id=? AND contract_address=? AND token_id=?`, collectibleCommunityDataColumns))
	if err != nil {
		return nil, err
	}

	row := getData.QueryRow(
		id.ContractID.ChainID,
		id.ContractID.Address,
		(*bigint.SQLBigIntBytes)(id.TokenID.Int),
	)

	var dbPrivilegesLevel sql.NullByte

	err = row.Scan(
		&dbPrivilegesLevel,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if dbPrivilegesLevel.Valid {
		ret.PrivilegesLevel = token.PrivilegesLevel(dbPrivilegesLevel.Byte)
	}

	return &ret, nil
}
