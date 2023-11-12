package walletdatabase

import (
	"database/sql"

	"github.com/status-im/status-go/sqlite"
	"github.com/status-im/status-go/walletdatabase/migrations"
)

type DbInitializer struct {
}

func (a DbInitializer) Initialize(path, password string, kdfIterationsNumber int) (*sql.DB, error) {
	return InitializeDB(path, password, kdfIterationsNumber)
}

var walletCustomSteps = []*sqlite.PostStep{}

func doMigration(db *sql.DB) error {
	// Run all the new migrations
	return migrations.Migrate(db, walletCustomSteps)
}

// InitializeDB creates db file at a given path and applies migrations.
func InitializeDB(path, password string, kdfIterationsNumber int) (*sql.DB, error) {
	db, err := sqlite.OpenDB(path, password, kdfIterationsNumber)
	if err != nil {
		return nil, err
	}

	err = doMigration(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}
