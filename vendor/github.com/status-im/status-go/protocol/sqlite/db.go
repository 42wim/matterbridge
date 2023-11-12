package sqlite

import (
	"database/sql"

	"github.com/pkg/errors"

	_ "github.com/mutecomm/go-sqlcipher/v4" // We require go sqlcipher that overrides default implementation
	"github.com/status-im/migrate/v4"
	"github.com/status-im/migrate/v4/database/sqlcipher"
	bindata "github.com/status-im/migrate/v4/source/go_bindata"
	mvdsmigrations "github.com/status-im/mvds/persistenceutil"
)

var migrationsTable = "status_protocol_go_" + sqlcipher.DefaultMigrationsTable

// applyMigrations allows to apply bindata migrations on the current *sql.DB.
// `assetNames` is a list of assets with migrations and `assetGetter` is responsible
// for returning the content of the asset with a given name.
func applyMigrations(db *sql.DB, assetNames []string, assetGetter func(name string) ([]byte, error)) error {
	resources := bindata.Resource(
		assetNames,
		assetGetter,
	)

	source, err := bindata.WithInstance(resources)
	if err != nil {
		return errors.Wrap(err, "failed to create migration source")
	}

	driver, err := sqlcipher.WithInstance(db, &sqlcipher.Config{
		MigrationsTable: migrationsTable,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create driver")
	}

	m, err := migrate.NewWithInstance(
		"go-bindata",
		source,
		"sqlcipher",
		driver,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create migration instance")
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return errors.Wrap(err, "could not get version")
	}

	err = ApplyAdHocMigrations(version, dirty, m, db)
	if err != nil {
		return errors.Wrap(err, "failed to apply ad-hoc migrations")
	}

	if dirty {
		err = ReplayLastMigration(version, m)
		if err != nil {
			return errors.Wrap(err, "failed to replay last migration")
		}
	}

	if err = m.Up(); err != migrate.ErrNoChange {
		return errors.Wrap(err, "failed to migrate")
	}

	return nil
}

func Migrate(database *sql.DB) error {
	// Apply migrations for all components.
	err := mvdsmigrations.Migrate(database)
	if err != nil {
		return errors.Wrap(err, "failed to apply mvds migrations")
	}

	migrationNames, migrationGetter, err := prepareMigrations(defaultMigrations)
	if err != nil {
		return errors.Wrap(err, "failed to prepare status-go/protocol migrations")
	}
	err = applyMigrations(database, migrationNames, migrationGetter)
	if err != nil {
		return errors.Wrap(err, "failed to apply status-go/protocol migrations")
	}
	return nil
}
