package migrationsprevnodecfg

import (
	"database/sql"

	bindata "github.com/status-im/migrate/v4/source/go_bindata"

	"github.com/status-im/status-go/sqlite"
)

// Migrate applies migrations.
func Migrate(db *sql.DB) error {
	return sqlite.Migrate(db, bindata.Resource(
		AssetNames(),
		func(name string) ([]byte, error) {
			return Asset(name)
		},
	), nil, nil)
}
