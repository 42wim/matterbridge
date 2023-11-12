package sqlite

import (
	"database/sql"

	_ "github.com/mutecomm/go-sqlcipher/v4" // We require go sqlcipher that overrides default implementation
	"github.com/pkg/errors"
	"github.com/status-im/migrate/v4"
)

const communitiesMigrationVersion uint = 1605075346

// FixCommunitiesMigration fixes an issue with a released migration
// In some instances if it was interrupted the migration would be skipped
// but marked as completed.
// What we do here is that we check whether we are at that migration, if
// so we check that the communities table is present, if not we re-run that
// migration.
func FixCommunitiesMigration(version uint, dirty bool, m *migrate.Migrate, db *sql.DB) error {
	// If the version is not the same, ignore
	if version != communitiesMigrationVersion {
		return nil
	}

	// If it's dirty, it will be replayed anyway
	if dirty {
		return nil
	}

	// Otherwise we check whether it actually succeeded by checking for the
	// presence of the communities_communities table

	var name string

	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='communities_communities'`).Scan(&name)

	// If the err is nil, it means the migration went through fine
	if err == nil {
		return nil
	}

	// If any other other, we return the error as that's unexpected
	if err != sql.ErrNoRows {
		return errors.Wrap(err, "failed to find the communities table")
	}

	// We replay the migration then
	return ReplayLastMigration(version, m)
}

func ReplayLastMigration(version uint, m *migrate.Migrate) error {
	// Force version if dirty so it's not dirty anymore
	if err := m.Force(int(version)); err != nil {
		return errors.Wrap(err, "failed to force migration")
	}

	// Step down 1 and we retry
	if err := m.Steps(-1); err != nil {
		return errors.Wrap(err, "failed to step down")
	}

	return nil
}

func ApplyAdHocMigrations(version uint, dirty bool, m *migrate.Migrate, db *sql.DB) error {
	return FixCommunitiesMigration(version, dirty, m, db)
}
