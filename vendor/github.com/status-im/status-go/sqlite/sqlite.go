package sqlite

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"

	sqlcipher "github.com/mutecomm/go-sqlcipher/v4" // We require go sqlcipher that overrides default implementation

	"github.com/status-im/status-go/common/dbsetup"
)

const (
	// The reduced number of kdf iterations (for performance reasons) which is
	// used as the default value
	// https://github.com/status-im/status-go/pull/1343
	// https://notes.status.im/i8Y_l7ccTiOYq09HVgoFwA
	ReducedKDFIterationsNumber = 3200

	// WALMode for sqlite.
	WALMode          = "wal"
	InMemoryPath     = ":memory:"
	V4CipherPageSize = 8192
	V3CipherPageSize = 1024
	sqlMainDatabase  = "main"
)

// DecryptDB completely removes the encryption from the db
func DecryptDB(oldPath string, newPath string, key string, kdfIterationsNumber int) error {

	db, err := openDB(oldPath, key, kdfIterationsNumber, V4CipherPageSize)
	if err != nil {
		return err
	}

	_, err = db.Exec(`ATTACH DATABASE '` + newPath + `' AS plaintext KEY ''`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`SELECT sqlcipher_export('plaintext')`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`DETACH DATABASE plaintext`)
	return err
}

func encryptDB(db *sql.DB, encryptedPath string, key string, kdfIterationsNumber int, onStart func(), onEnd func()) error {
	if onStart != nil {
		onStart()
	}
	if onEnd != nil {
		defer onEnd()
	}

	attachedDbName := "encrypted"
	err := attachDatabaseWithDefaultSettings(db, encryptedPath, attachedDbName, key, kdfIterationsNumber)
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf(`SELECT sqlcipher_export('%s')`, attachedDbName))
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf(`DETACH DATABASE %s`, attachedDbName))
	return err
}

func attachDatabaseWithDefaultSettings(db *sql.DB, attachedDbPath string, attachedDbName string, key string, kdfIterationsNumber int) error {
	_, err := db.Exec(fmt.Sprintf(`ATTACH DATABASE '%s' AS %s KEY '%s'`, attachedDbPath, attachedDbName, key))
	if err != nil {
		return err
	}

	if kdfIterationsNumber <= 0 {
		kdfIterationsNumber = dbsetup.ReducedKDFIterationsNumber
	}

	if _, err := db.Exec(fmt.Sprintf(`PRAGMA %s.busy_timeout = 60000`, attachedDbName)); err != nil {
		return errors.New("failed to set `busy_timeout` pragma on attached db")
	}

	return setDatabaseCipherSettings(db, kdfIterationsNumber, attachedDbName)
}

func setDatabaseCipherSettings(db *sql.DB, kdfIterationsNumber int, dbNameOpt ...string) error {
	dbName := sqlMainDatabase
	if len(dbNameOpt) > 0 {
		dbName = dbNameOpt[0]
	}

	_, err := db.Exec(fmt.Sprintf("PRAGMA %s.kdf_iter = '%d'", dbName, kdfIterationsNumber))
	if err != nil {
		return err
	}

	if _, err := db.Exec(fmt.Sprintf("PRAGMA %s.cipher_page_size = %d", dbName, V4CipherPageSize)); err != nil {
		fmt.Println("failed to set cipher_page_size pragma")
		return err
	}
	if _, err := db.Exec(fmt.Sprintf("PRAGMA %s.cipher_hmac_algorithm = HMAC_SHA1", dbName)); err != nil {
		fmt.Println("failed to set cipher_hmac_algorithm pragma")
		return err
	}

	if _, err := db.Exec(fmt.Sprintf("PRAGMA %s.cipher_kdf_algorithm = PBKDF2_HMAC_SHA1", dbName)); err != nil {
		fmt.Println("failed to set cipher_kdf_algorithm pragma")
		return err
	}

	return nil
}

// EncryptDB takes a plaintext database and adds encryption
func EncryptDB(unencryptedPath string, encryptedPath string, key string, kdfIterationsNumber int, onStart func(), onEnd func()) error {
	_ = os.Remove(encryptedPath)

	db, err := OpenUnecryptedDB(unencryptedPath)
	if err != nil {
		return err
	}
	return encryptDB(db, encryptedPath, key, kdfIterationsNumber, onStart, onEnd)
}

// Export takes an encrypted database and re-encrypts it in a new file, with a new key
func ExportDB(encryptedPath string, key string, kdfIterationsNumber int, newPath string, newKey string, onStart func(), onEnd func()) error {
	db, err := openDB(encryptedPath, key, kdfIterationsNumber, V4CipherPageSize)
	if err != nil {
		return err
	}
	defer db.Close()
	return encryptDB(db, newPath, newKey, kdfIterationsNumber, onStart, onEnd)
}

func buildSqlcipherDSN(path string) (string, error) {
	if path == InMemoryPath {
		return InMemoryPath, nil
	}

	// Adding sqlcipher query parameter to the DSN
	queryOperator := "?"

	if queryStart := strings.IndexRune(path, '?'); queryStart != -1 {
		params, err := url.ParseQuery(path[queryStart+1:])
		if err != nil {
			return "", err
		}

		if len(params) > 0 {
			queryOperator = "&"
		}
	}

	// We need to set txlock=immediate to avoid "database is locked" errors during concurrent write operations
	// This could happen when a read transaction is promoted to write transaction
	// https://www.sqlite.org/lang_transaction.html
	return path + queryOperator + "_txlock=immediate", nil
}

func openDB(path string, key string, kdfIterationsNumber int, cipherPageSize int) (*sql.DB, error) {
	driverName := fmt.Sprintf("sqlcipher_with_extensions-%d", len(sql.Drivers()))
	sql.Register(driverName, &sqlcipher.SQLiteDriver{
		ConnectHook: func(conn *sqlcipher.SQLiteConn) error {
			if _, err := conn.Exec("PRAGMA foreign_keys=ON", []driver.Value{}); err != nil {
				return errors.New("failed to set `foreign_keys` pragma")
			}

			if _, err := conn.Exec(fmt.Sprintf("PRAGMA key = '%s'", key), []driver.Value{}); err != nil {
				return errors.New("failed to set `key` pragma")
			}

			if kdfIterationsNumber <= 0 {
				kdfIterationsNumber = dbsetup.ReducedKDFIterationsNumber
			}

			if _, err := conn.Exec(fmt.Sprintf("PRAGMA cipher_page_size = %d", cipherPageSize), nil); err != nil {
				fmt.Println("failed to set cipher_page_size pragma")
				return err
			}
			if _, err := conn.Exec("PRAGMA cipher_hmac_algorithm = HMAC_SHA1", nil); err != nil {
				fmt.Println("failed to set cipher_hmac_algorithm pragma")
				return err
			}

			if _, err := conn.Exec("PRAGMA cipher_kdf_algorithm = PBKDF2_HMAC_SHA1", nil); err != nil {
				fmt.Println("failed to set cipher_kdf_algorithm pragma")
				return err
			}

			if _, err := conn.Exec(fmt.Sprintf("PRAGMA kdf_iter = '%d'", kdfIterationsNumber), []driver.Value{}); err != nil {
				return errors.New("failed to set `kdf_iter` pragma")
			}

			// readers do not block writers and faster i/o operations
			if _, err := conn.Exec("PRAGMA journal_mode=WAL", []driver.Value{}); err != nil && path != InMemoryPath {
				return fmt.Errorf("failed to set `journal_mode` pragma: %w", err)
			}

			// workaround to mitigate the issue of "database is locked" errors during concurrent write operations
			if _, err := conn.Exec("PRAGMA busy_timeout=60000", []driver.Value{}); err != nil {
				return errors.New("failed to set `busy_timeout` pragma")
			}

			return nil
		},
	})

	dsn, err := buildSqlcipherDSN(path)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}

	if path == InMemoryPath {
		db.SetMaxOpenConns(1)
	} else {
		nproc := func() int {
			maxProcs := runtime.GOMAXPROCS(0)
			numCPU := runtime.NumCPU()
			if maxProcs < numCPU {
				return maxProcs
			}
			return numCPU
		}()
		db.SetMaxOpenConns(nproc)
		db.SetMaxIdleConns(nproc)
	}

	// Dummy select to check if the key is correct. Will return last error from initialization
	if _, err := db.Exec("SELECT 'Key check'"); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// OpenDB opens encrypted database.
func OpenDB(path string, key string, kdfIterationsNumber int) (*sql.DB, error) {
	return openDB(path, key, kdfIterationsNumber, V4CipherPageSize)
}

// OpenUnecryptedDB opens database with setting PRAGMA key.
func OpenUnecryptedDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// Disable concurrent access as not supported by the driver
	db.SetMaxOpenConns(1)

	if _, err = db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, err
	}
	// readers do not block writers and faster i/o operations
	// https://www.sqlite.org/draft/wal.html
	// must be set after db is encrypted
	if path != InMemoryPath {
		var mode string
		err = db.QueryRow("PRAGMA journal_mode=WAL").Scan(&mode)
		if err != nil {
			return nil, err
		}
		if mode != WALMode {
			return nil, fmt.Errorf("unable to set journal_mode to WAL. actual mode %s", mode)
		}
	}

	return db, nil
}

func ChangeEncryptionKey(path string, key string, kdfIterationsNumber int, newKey string, onStart func(), onEnd func()) error {
	if onStart != nil {
		onStart()
	}

	if onEnd != nil {
		defer onEnd()
	}

	if kdfIterationsNumber <= 0 {
		kdfIterationsNumber = dbsetup.ReducedKDFIterationsNumber
	}

	db, err := openDB(path, key, kdfIterationsNumber, V4CipherPageSize)

	if err != nil {
		return err
	}

	resetKeyString := fmt.Sprintf("PRAGMA rekey = '%s'", newKey)
	if _, err = db.Exec(resetKeyString); err != nil {
		return errors.New("failed to set rekey pragma")
	}

	return nil
}

// MigrateV3ToV4 migrates database from v3 to v4 format with encryption.
func MigrateV3ToV4(v3Path string, v4Path string, key string, kdfIterationsNumber int, onStart func(), onEnd func()) error {

	db, err := openDB(v3Path, key, kdfIterationsNumber, V3CipherPageSize)

	if err != nil {
		fmt.Println("failed to open db", err)
		return err
	}
	defer db.Close()

	return encryptDB(db, v4Path, key, kdfIterationsNumber, onStart, onEnd)
}
