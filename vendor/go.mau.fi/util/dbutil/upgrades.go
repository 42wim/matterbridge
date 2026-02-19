// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dbutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type upgradeFunc func(context.Context, *Database) error

type upgrade struct {
	message string
	fn      upgradeFunc

	upgradesTo    int
	compatVersion int
	transaction   TxnMode
}

func (u *upgrade) DangerouslyRun(ctx context.Context, db *Database) (upgradesTo, compat int, err error) {
	return u.upgradesTo, u.compatVersion, u.fn(ctx, db)
}

var ErrUnsupportedDatabaseVersion = errors.New("unsupported database schema version")
var ErrForeignTables = errors.New("the database contains foreign tables")
var ErrNotOwned = errors.New("the database is owned by")
var ErrUnsupportedDialect = errors.New("unsupported database dialect")

type NotOwnedError struct {
	Owner string
}

func (e NotOwnedError) Error() string {
	return fmt.Sprintf("%v %s", ErrNotOwned, e.Owner)
}

func (e NotOwnedError) Unwrap() error {
	return ErrNotOwned
}

func DangerousInternalUpgradeVersionTable(ctx context.Context, db *Database) error {
	return db.upgradeVersionTable(ctx)
}

func (db *Database) upgradeVersionTable(ctx context.Context) error {
	if compatColumnExists, err := db.ColumnExists(ctx, db.VersionTable, "compat"); err != nil {
		return fmt.Errorf("failed to check if version table is up to date: %w", err)
	} else if !compatColumnExists {
		if tableExists, err := db.TableExists(ctx, db.VersionTable); err != nil {
			return fmt.Errorf("failed to check if version table exists: %w", err)
		} else if !tableExists {
			_, err = db.Exec(ctx, fmt.Sprintf("CREATE TABLE %s (version INTEGER, compat INTEGER)", db.VersionTable))
			if err != nil {
				return fmt.Errorf("failed to create version table: %w", err)
			}
		} else {
			_, err = db.Exec(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN compat INTEGER", db.VersionTable))
			if err != nil {
				return fmt.Errorf("failed to add compat column to version table: %w", err)
			}
		}
	}
	return nil
}

func (db *Database) getVersion(ctx context.Context) (version, compat int, err error) {
	if err = db.upgradeVersionTable(ctx); err != nil {
		return
	}

	var compatNull sql.NullInt32
	err = db.QueryRow(ctx, fmt.Sprintf("SELECT version, compat FROM %s LIMIT 1", db.VersionTable)).Scan(&version, &compatNull)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if compatNull.Valid && compatNull.Int32 != 0 {
		compat = int(compatNull.Int32)
	} else {
		compat = version
	}
	return
}

const (
	tableExistsPostgres = "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name=$1)"
	tableExistsSQLite   = "SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND tbl_name=?1)"
)

func (db *Database) TableExists(ctx context.Context, table string) (exists bool, err error) {
	switch db.Dialect {
	case SQLite:
		err = db.QueryRow(ctx, tableExistsSQLite, table).Scan(&exists)
	case Postgres:
		err = db.QueryRow(ctx, tableExistsPostgres, table).Scan(&exists)
	default:
		err = ErrUnsupportedDialect
	}
	return
}

const (
	columnExistsPostgres = "SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name=$1 AND column_name=$2)"
	columnExistsSQLite   = "SELECT EXISTS(SELECT 1 FROM pragma_table_info(?1) WHERE name=?2)"
)

func (db *Database) ColumnExists(ctx context.Context, table, column string) (exists bool, err error) {
	switch db.Dialect {
	case SQLite:
		err = db.QueryRow(ctx, columnExistsSQLite, table, column).Scan(&exists)
	case Postgres:
		err = db.QueryRow(ctx, columnExistsPostgres, table, column).Scan(&exists)
	default:
		err = ErrUnsupportedDialect
	}
	return
}

const createOwnerTable = `
CREATE TABLE IF NOT EXISTS database_owner (
	key   INTEGER PRIMARY KEY DEFAULT 0,
	owner TEXT NOT NULL
)
`

func (db *Database) checkDatabaseOwner(ctx context.Context) error {
	var owner string
	if !db.IgnoreForeignTables {
		if exists, err := db.TableExists(ctx, "state_groups_state"); err != nil {
			return fmt.Errorf("failed to check if state_groups_state exists: %w", err)
		} else if exists {
			return fmt.Errorf("%w (found state_groups_state, likely belonging to Synapse)", ErrForeignTables)
		} else if exists, err = db.TableExists(ctx, "roomserver_rooms"); err != nil {
			return fmt.Errorf("failed to check if roomserver_rooms exists: %w", err)
		} else if exists {
			return fmt.Errorf("%w (found roomserver_rooms, likely belonging to Dendrite)", ErrForeignTables)
		}
	}
	if db.Owner == "" {
		return nil
	}
	if _, err := db.Exec(ctx, createOwnerTable); err != nil {
		return fmt.Errorf("failed to ensure database owner table exists: %w", err)
	} else if err = db.QueryRow(ctx, "SELECT owner FROM database_owner WHERE key=0").Scan(&owner); errors.Is(err, sql.ErrNoRows) {
		_, err = db.Exec(ctx, "INSERT INTO database_owner (key, owner) VALUES (0, $1)", db.Owner)
		if err != nil {
			return fmt.Errorf("failed to insert database owner: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check database owner: %w", err)
	} else if owner != db.Owner {
		return NotOwnedError{owner}
	}
	return nil
}

func (db *Database) setVersion(ctx context.Context, version, compat int) error {
	_, err := db.Exec(ctx, fmt.Sprintf("DELETE FROM %s", db.VersionTable))
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, fmt.Sprintf("INSERT INTO %s (version, compat) VALUES ($1, $2)", db.VersionTable), version, compat)
	return err
}

func (db *Database) DoSQLiteTransactionWithoutForeignKeys(ctx context.Context, doUpgrade func(context.Context) error) error {
	conn, err := db.AcquireConn(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	_, err = conn.ExecContext(ctx, "PRAGMA foreign_keys=OFF")
	if err != nil {
		return fmt.Errorf("failed to disable foreign keys: %w", err)
	}
	err = db.DoTxn(ctx, &TxnOptions{Conn: conn}, func(ctx context.Context) error {
		err := doUpgrade(ctx)
		if err != nil {
			return err
		}
		_, err = conn.ExecContext(ctx, "PRAGMA foreign_key_check")
		if err != nil {
			return fmt.Errorf("failed to check foreign keys after upgrade: %w", err)
		}
		return nil
	})
	if err != nil {
		_, _ = conn.ExecContext(ctx, "PRAGMA foreign_keys=ON")
		return err
	}
	_, err = conn.ExecContext(ctx, "PRAGMA foreign_keys=ON")
	if err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	return nil
}

func (db *Database) Upgrade(ctx context.Context) error {
	err := db.checkDatabaseOwner(ctx)
	if err != nil {
		return err
	}

	version, compat, err := db.getVersion(ctx)
	if err != nil {
		return err
	}

	if compat > len(db.UpgradeTable) {
		if db.IgnoreUnsupportedDatabase {
			db.Log.WarnUnsupportedVersion(version, compat, len(db.UpgradeTable))
			return nil
		}
		return fmt.Errorf("%w: currently on v%d (compatible down to v%d), latest known: v%d", ErrUnsupportedDatabaseVersion, version, compat, len(db.UpgradeTable))
	}

	db.Log.PrepareUpgrade(version, compat, len(db.UpgradeTable))
	logVersion := version
	for version < len(db.UpgradeTable) {
		upgradeItem := db.UpgradeTable[version]
		if upgradeItem.fn == nil {
			version++
			continue
		}
		doUpgrade := func(ctx context.Context) error {
			err = upgradeItem.fn(ctx, db)
			if err != nil {
				return fmt.Errorf("failed to run upgrade v%d->v%d: %w", version, upgradeItem.upgradesTo, err)
			}
			version = upgradeItem.upgradesTo
			logVersion = version
			err = db.setVersion(ctx, version, upgradeItem.compatVersion)
			if err != nil {
				return err
			}
			return nil
		}
		db.Log.DoUpgrade(logVersion, upgradeItem.upgradesTo, upgradeItem.message, upgradeItem.transaction)
		switch upgradeItem.transaction {
		case TxnModeOff:
			err = doUpgrade(ctx)
		case TxnModeOn:
			err = db.DoTxn(ctx, nil, doUpgrade)
		case TxnModeSQLiteForeignKeysOff:
			switch db.Dialect {
			case SQLite:
				err = db.DoSQLiteTransactionWithoutForeignKeys(ctx, doUpgrade)
			default:
				err = db.DoTxn(ctx, nil, doUpgrade)
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}
