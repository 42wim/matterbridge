package collectibles

import (
	"database/sql"
	"fmt"

	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/sqlite"
)

type CollectionDataDB struct {
	db *sql.DB
}

func NewCollectionDataDB(sqlDb *sql.DB) *CollectionDataDB {
	return &CollectionDataDB{
		db: sqlDb,
	}
}

const collectionDataColumns = "chain_id, contract_address, provider, name, slug, image_url, image_payload, community_id"
const collectionTraitsColumns = "chain_id, contract_address, trait_type, min, max"
const selectCollectionTraitsColumns = "trait_type, min, max"

func rowsToCollectionTraits(rows *sql.Rows) (map[string]thirdparty.CollectionTrait, error) {
	traits := make(map[string]thirdparty.CollectionTrait)
	for rows.Next() {
		var traitType string
		var trait thirdparty.CollectionTrait
		err := rows.Scan(
			&traitType,
			&trait.Min,
			&trait.Max,
		)
		if err != nil {
			return nil, err
		}
		traits[traitType] = trait
	}
	return traits, nil
}

func getCollectionTraits(creator sqlite.StatementCreator, id thirdparty.ContractID) (map[string]thirdparty.CollectionTrait, error) {
	// Get traits list
	selectTraits, err := creator.Prepare(fmt.Sprintf(`SELECT %s
		FROM collection_traits_cache
		WHERE chain_id = ? AND contract_address = ?`, selectCollectionTraitsColumns))
	if err != nil {
		return nil, err
	}

	rows, err := selectTraits.Query(
		id.ChainID,
		id.Address,
	)
	if err != nil {
		return nil, err
	}

	return rowsToCollectionTraits(rows)
}

func upsertCollectionTraits(creator sqlite.StatementCreator, id thirdparty.ContractID, traits map[string]thirdparty.CollectionTrait) error {
	// Rremove old traits list
	deleteTraits, err := creator.Prepare(`DELETE FROM collection_traits_cache WHERE chain_id = ? AND contract_address = ?`)
	if err != nil {
		return err
	}

	_, err = deleteTraits.Exec(
		id.ChainID,
		id.Address,
	)
	if err != nil {
		return err
	}

	// Insert new traits list
	insertTrait, err := creator.Prepare(fmt.Sprintf(`INSERT OR REPLACE INTO collection_traits_cache (%s)
		VALUES (?, ?, ?, ?, ?)`, collectionTraitsColumns))
	if err != nil {
		return err
	}

	for traitType, trait := range traits {
		_, err = insertTrait.Exec(
			id.ChainID,
			id.Address,
			traitType,
			trait.Min,
			trait.Max,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func setCollectionsData(creator sqlite.StatementCreator, collections []thirdparty.CollectionData, allowUpdate bool) error {
	insertCollection, err := creator.Prepare(fmt.Sprintf(`%s INTO collection_data_cache (%s) 
																				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, insertStatement(allowUpdate), collectionDataColumns))
	if err != nil {
		return err
	}

	for _, c := range collections {
		_, err = insertCollection.Exec(
			c.ID.ChainID,
			c.ID.Address,
			c.Provider,
			c.Name,
			c.Slug,
			c.ImageURL,
			c.ImagePayload,
			c.CommunityID,
		)
		if err != nil {
			return err
		}

		err = upsertContractType(creator, c.ID, c.ContractType)
		if err != nil {
			return err
		}

		if allowUpdate {
			err = upsertCollectionTraits(creator, c.ID, c.Traits)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (o *CollectionDataDB) SetData(collections []thirdparty.CollectionData, allowUpdate bool) (err error) {
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

	// Insert new collections data
	err = setCollectionsData(tx, collections, allowUpdate)
	if err != nil {
		return err
	}

	return
}

func scanCollectionsDataRow(row *sql.Row) (*thirdparty.CollectionData, error) {
	c := thirdparty.CollectionData{
		Traits: make(map[string]thirdparty.CollectionTrait),
	}
	err := row.Scan(
		&c.ID.ChainID,
		&c.ID.Address,
		&c.Provider,
		&c.Name,
		&c.Slug,
		&c.ImageURL,
		&c.ImagePayload,
		&c.CommunityID,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (o *CollectionDataDB) GetIDsNotInDB(ids []thirdparty.ContractID) ([]thirdparty.ContractID, error) {
	ret := make([]thirdparty.ContractID, 0, len(ids))

	exists, err := o.db.Prepare(`SELECT EXISTS (
			SELECT 1 FROM collection_data_cache
			WHERE chain_id=? AND contract_address=?
		)`)
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		row := exists.QueryRow(
			id.ChainID,
			id.Address,
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

func (o *CollectionDataDB) GetData(ids []thirdparty.ContractID) (map[string]thirdparty.CollectionData, error) {
	ret := make(map[string]thirdparty.CollectionData)

	getData, err := o.db.Prepare(fmt.Sprintf(`SELECT %s
		FROM collection_data_cache
		WHERE chain_id=? AND contract_address=?`, collectionDataColumns))
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		row := getData.QueryRow(
			id.ChainID,
			id.Address,
		)
		c, err := scanCollectionsDataRow(row)
		if err == sql.ErrNoRows {
			continue
		} else if err != nil {
			return nil, err
		} else {
			// Get traits from different table
			c.Traits, err = getCollectionTraits(o.db, c.ID)
			if err != nil {
				return nil, err
			}

			// Get contract type from different table
			c.ContractType, err = readContractType(o.db, c.ID)
			if err != nil {
				return nil, err
			}

			ret[c.ID.HashKey()] = *c
		}
	}
	return ret, nil
}
