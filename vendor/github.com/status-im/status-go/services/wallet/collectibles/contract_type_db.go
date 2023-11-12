package collectibles

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"

	w_common "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/sqlite"
)

func upsertContractType(creator sqlite.StatementCreator, id thirdparty.ContractID, contractType w_common.ContractType) error {
	if contractType == w_common.ContractTypeUnknown {
		return nil
	}

	q := sq.Replace("contract_type_cache").
		SetMap(sq.Eq{"chain_id": id.ChainID, "contract_address": id.Address, "contract_type": contractType})

	query, args, err := q.ToSql()
	if err != nil {
		return err
	}

	stmt, err := creator.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)

	return err
}

func readContractType(creator sqlite.StatementCreator, id thirdparty.ContractID) (w_common.ContractType, error) {
	q := sq.Select("contract_type").
		From("contract_type_cache").
		Where(sq.Eq{"chain_id": id.ChainID, "contract_address": id.Address})

	query, args, err := q.ToSql()
	if err != nil {
		return w_common.ContractTypeUnknown, err
	}

	stmt, err := creator.Prepare(query)
	if err != nil {
		return w_common.ContractTypeUnknown, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return w_common.ContractTypeUnknown, err
	}

	var transferType w_common.ContractType
	err = stmt.QueryRow(args...).Scan(&transferType)

	if err == sql.ErrNoRows {
		return w_common.ContractTypeUnknown, nil
	} else if err != nil {
		return w_common.ContractTypeUnknown, err
	}

	return transferType, nil
}
