package dbsetup

import (
	"database/sql"
	"errors"

	"github.com/ethereum/go-ethereum/log"
)

const InMemoryPath = ":memory:"

// The reduced number of kdf iterations (for performance reasons) which is
// currently used for derivation of the database key
// https://github.com/status-im/status-go/pull/1343
// https://notes.status.im/i8Y_l7ccTiOYq09HVgoFwA
const ReducedKDFIterationsNumber = 3200

type DatabaseInitializer interface {
	Initialize(path, password string, kdfIterationsNumber int) (*sql.DB, error)
}

// GetDBFilename takes an instance of sql.DB and returns the filename of the "main" database
func GetDBFilename(db *sql.DB) (string, error) {
	if db == nil {
		logger := log.New()
		logger.Warn("GetDBFilename was passed a nil pointer sql.DB")
		return "", nil
	}

	var i, category, filename string
	rows, err := db.Query("PRAGMA database_list;")
	if err != nil {
		return "", err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&i, &category, &filename)
		if err != nil {
			return "", err
		}

		// The "main" database is the one we care about
		if category == "main" {
			return filename, nil
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	return "", errors.New("no main database found")
}
