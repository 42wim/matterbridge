package currency

import (
	"context"
	"database/sql"
)

type DB struct {
	db *sql.DB
}

func NewCurrencyDB(sqlDb *sql.DB) *DB {
	return &DB{
		db: sqlDb,
	}
}

func getCachedFormatsFromDBRows(rows *sql.Rows) (FormatPerSymbol, error) {
	formats := make(FormatPerSymbol)

	for rows.Next() {
		var format Format
		if err := rows.Scan(&format.Symbol, &format.DisplayDecimals, &format.StripTrailingZeroes); err != nil {
			return nil, err
		}

		formats[format.Symbol] = format
	}

	return formats, nil
}

func (cdb *DB) GetCachedFormats() (FormatPerSymbol, error) {
	rows, err := cdb.db.Query("SELECT symbol, display_decimals, strip_trailing_zeroes FROM currency_format_cache")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return getCachedFormatsFromDBRows(rows)
}

func (cdb *DB) UpdateCachedFormats(formats FormatPerSymbol) error {
	tx, err := cdb.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	insert, err := tx.Prepare(`INSERT OR REPLACE INTO currency_format_cache
				(symbol, display_decimals, strip_trailing_zeroes)
				VALUES
				(?, ?, ?)`)
	if err != nil {
		return err
	}

	for _, format := range formats {
		_, err = insert.Exec(format.Symbol, format.DisplayDecimals, format.StripTrailingZeroes)
		if err != nil {
			return err
		}
	}
	return nil
}
