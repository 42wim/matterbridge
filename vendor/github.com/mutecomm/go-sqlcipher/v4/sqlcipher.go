package sqlite3

import (
	"bytes"
	"errors"
	"os"
)

// sqlite3Header defines the header string used by SQLite 3.
var sqlite3Header = []byte("SQLite format 3\000")

// IsEncrypted returns true, if the database with the given filename is
// encrypted, and false otherwise.
// If the database header cannot be read properly an error is returned.
func IsEncrypted(filename string) (bool, error) {
	// open file
	db, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer db.Close()
	// read header
	var header [16]byte
	n, err := db.Read(header[:])
	if err != nil {
		return false, err
	}
	if n != len(header) {
		return false, errors.New("go-sqlcipher: could not read full header")
	}
	// SQLCipher encrypts also the header, the file is encrypted if the read
	// header does not equal the header string used by SQLite 3.
	encrypted := !bytes.Equal(header[:], sqlite3Header)
	return encrypted, nil
}
