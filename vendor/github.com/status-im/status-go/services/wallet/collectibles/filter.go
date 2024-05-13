package collectibles

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ethereum/go-ethereum/common"

	sq "github.com/Masterminds/squirrel"

	"github.com/status-im/status-go/protocol/communities/token"
	"github.com/status-im/status-go/services/wallet/bigint"
	wcommon "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/thirdparty"
)

func allCollectibleIDsFilter() []thirdparty.CollectibleUniqueID {
	return []thirdparty.CollectibleUniqueID{}
}

func allCommunityIDsFilter() []string {
	return []string{}
}

func allCommunityPrivilegesLevelsFilter() []token.PrivilegesLevel {
	return []token.PrivilegesLevel{}
}

func allFilter() Filter {
	return Filter{
		CollectibleIDs:            allCollectibleIDsFilter(),
		CommunityIDs:              allCommunityIDsFilter(),
		CommunityPrivilegesLevels: allCommunityPrivilegesLevelsFilter(),
		FilterCommunity:           All,
	}
}

type FilterCommunityType int

const (
	All FilterCommunityType = iota
	OnlyNonCommunity
	OnlyCommunity
)

type Filter struct {
	CollectibleIDs            []thirdparty.CollectibleUniqueID `json:"collectible_ids"`
	CommunityIDs              []string                         `json:"community_ids"`
	CommunityPrivilegesLevels []token.PrivilegesLevel          `json:"community_privileges_levels"`

	FilterCommunity FilterCommunityType `json:"filter_community"`
}

func filterOwnedCollectibles(ctx context.Context, db *sql.DB, chainIDs []wcommon.ChainID, addresses []common.Address, filter Filter, offset int, limit int) ([]thirdparty.CollectibleUniqueID, error) {
	if len(addresses) == 0 {
		return nil, errors.New("no addresses provided")
	}
	if len(chainIDs) == 0 {
		return nil, errors.New("no chainIDs provided")
	}

	q := sq.Select("ownership.chain_id,ownership.contract_address,ownership.token_id").Distinct()
	q = q.From("collectibles_ownership_cache ownership").
		LeftJoin(`collectible_data_cache data ON 
		ownership.chain_id = data.chain_id AND 
		ownership.contract_address = data.contract_address AND 
		ownership.token_id = data.token_id`)

	qConditions := sq.And{}
	qConditions = append(qConditions, sq.Eq{"ownership.chain_id": chainIDs})
	qConditions = append(qConditions, sq.Eq{"ownership.owner_address": addresses})

	if len(filter.CollectibleIDs) > 0 {
		collectibleIDConditions := sq.Or{}
		for _, collectibleID := range filter.CollectibleIDs {
			collectibleIDConditions = append(collectibleIDConditions,
				sq.And{
					sq.Eq{"ownership.chain_id": collectibleID.ContractID.ChainID},
					sq.Eq{"ownership.contract_address": collectibleID.ContractID.Address},
					sq.Eq{"ownership.token_id": (*bigint.SQLBigIntBytes)(collectibleID.TokenID.Int)},
				})
		}
		qConditions = append(qConditions, collectibleIDConditions)
	}

	switch filter.FilterCommunity {
	case All:
		// nothing to do
	case OnlyNonCommunity:
		qConditions = append(qConditions, sq.Eq{"data.community_id": ""})
	case OnlyCommunity:
		qConditions = append(qConditions, sq.NotEq{"data.community_id": ""})
	}

	if len(filter.CommunityIDs) > 0 {
		qConditions = append(qConditions, sq.Eq{"data.community_id": filter.CommunityIDs})
	}

	if len(filter.CommunityPrivilegesLevels) > 0 {
		qConditions = append(qConditions, sq.Eq{"data.community_privileges_level": filter.CommunityPrivilegesLevels})
	}

	q = q.Where(qConditions)

	q = q.Limit(uint64(limit))
	q = q.Offset(uint64(offset))

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return thirdparty.RowsToCollectibles(rows)
}
