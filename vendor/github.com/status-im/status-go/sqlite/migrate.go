package sqlite

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/status-im/migrate/v4"
	"github.com/status-im/migrate/v4/database/sqlcipher"
	bindata "github.com/status-im/migrate/v4/source/go_bindata"
)

type CustomMigrationFunc func(tx *sql.Tx) error

type PostStep struct {
	Version         uint
	CustomMigration CustomMigrationFunc
	RollBackVersion uint
}

var migrationTable = "status_go_" + sqlcipher.DefaultMigrationsTable

// Migrate database with option to augment the migration steps with additional processing using the customSteps
// parameter. For each PostStep entry in customSteps the CustomMigration will be called after the migration step
// with the matching Version number has been executed. If the CustomMigration returns an error, the migration process
// is aborted. In case the custom step failures the migrations are run down to RollBackVersion if > 0.
//
// The recommended way to create a custom migration is by providing empty and versioned run/down sql files as markers.
// Then running all the SQL code inside the same transaction to transform and commit provides the possibility
// to completely rollback the migration in case of failure, avoiding to leave the DB in an inconsistent state.
//
// Marker migrations can be created by using PostStep structs with specific Version numbers and a callback function,
// even when no accompanying SQL migration is needed. This can be used to trigger Go code at specific points
// during the migration process.
//
// Caution: This mechanism should be used as a last resort. Prefer data migration using SQL migration files
// whenever possible to ensure consistency and compatibility with standard migration tools.
//
// untilVersion, for testing purposes optional parameter, can be used to limit the migration to a specific version.
// Pass nil to migrate to the latest available version.
func Migrate(db *sql.DB, resources *bindata.AssetSource, customSteps []*PostStep, untilVersion *uint) error {
	source, err := bindata.WithInstance(resources)
	if err != nil {
		return fmt.Errorf("failed to create bindata migration source: %w", err)
	}

	driver, err := sqlcipher.WithInstance(db, &sqlcipher.Config{
		MigrationsTable: migrationTable,
	})
	if err != nil {
		return fmt.Errorf("failed to create sqlcipher driver: %w", err)
	}

	m, err := migrate.NewWithInstance("go-bindata", source, "sqlcipher", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if len(customSteps) == 0 {
		return runRemainingMigrations(m, untilVersion)
	}

	sort.Slice(customSteps, func(i, j int) bool {
		return customSteps[i].Version < customSteps[j].Version
	})

	lastVersion, err := getCurrentVersion(m, db)
	if err != nil {
		return err
	}

	customIndex := 0
	// ignore processed versions
	for customIndex < len(customSteps) && customSteps[customIndex].Version <= lastVersion {
		customIndex++
	}

	if err := runCustomMigrations(m, db, customSteps, customIndex, untilVersion); err != nil {
		return err
	}

	return runRemainingMigrations(m, untilVersion)
}

// runCustomMigrations performs source migrations from current to each custom steps, then runs custom migration callback
// until it executes all custom migrations or an error occurs and it tries to rollback to RollBackVersion if > 0.
func runCustomMigrations(m *migrate.Migrate, db *sql.DB, customSteps []*PostStep, customIndex int, untilVersion *uint) error {
	for customIndex < len(customSteps) && (untilVersion == nil || customSteps[customIndex].Version <= *untilVersion) {
		customStep := customSteps[customIndex]

		if err := m.Migrate(customStep.Version); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to migrate to version %d: %w", customStep.Version, err)
		}

		if err := runCustomMigrationStep(db, customStep, m); err != nil {
			return err
		}

		customIndex++
	}
	return nil
}

func runCustomMigrationStep(db *sql.DB, customStep *PostStep, m *migrate.Migrate) error {

	sqlTx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := customStep.CustomMigration(sqlTx); err != nil {
		_ = sqlTx.Rollback()
		return rollbackCustomMigration(m, customStep, err)
	}

	if err := sqlTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func rollbackCustomMigration(m *migrate.Migrate, customStep *PostStep, customErr error) error {
	if customStep.RollBackVersion > 0 {
		err := m.Migrate(customStep.RollBackVersion)
		newV, _, _ := m.Version()
		if err != nil {
			return fmt.Errorf("failed to rollback migration to version %d: %w", customStep.RollBackVersion, err)
		}
		return fmt.Errorf("custom migration step failed for version %d. Successfully rolled back migration to version %d: %w", customStep.Version, newV, customErr)
	}
	return fmt.Errorf("custom migration step failed for version %d: %w", customStep.Version, customErr)
}

func runRemainingMigrations(m *migrate.Migrate, untilVersion *uint) error {
	if untilVersion != nil {
		if err := m.Migrate(*untilVersion); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to migrate to version %d: %w", *untilVersion, err)
		}
	} else {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			ver, _, _ := m.Version()
			return fmt.Errorf("failed to migrate up: %w, current version: %d", err, ver)
		}
	}
	return nil
}

func getCurrentVersion(m *migrate.Migrate, db *sql.DB) (uint, error) {
	lastVersion, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, fmt.Errorf("failed to get migration version: %w", err)
	}
	if dirty {
		return 0, fmt.Errorf("DB is dirty after migration version %d", lastVersion)
	}
	if err == migrate.ErrNilVersion {
		lastVersion, _, err = GetLastMigrationVersion(db)
		return lastVersion, err
	}
	return lastVersion, nil
}

// GetLastMigrationVersion returns the last migration version stored in the migration table.
// Returns 0 for version in case migrationTableExists is true
func GetLastMigrationVersion(db *sql.DB) (version uint, migrationTableExists bool, err error) {
	// Check if the migration table exists
	row := db.QueryRow("SELECT exists(SELECT name FROM sqlite_master WHERE type='table' AND name=?)", migrationTable)
	migrationTableExists = false
	err = row.Scan(&migrationTableExists)
	if err != nil && err != sql.ErrNoRows {
		return 0, false, err
	}

	var lastMigration uint64 = 0
	if migrationTableExists {
		row = db.QueryRow("SELECT version FROM status_go_schema_migrations")
		err = row.Scan(&lastMigration)
		if err != nil && err != sql.ErrNoRows {
			return 0, true, err
		}
	}
	return uint(lastMigration), migrationTableExists, nil
}
