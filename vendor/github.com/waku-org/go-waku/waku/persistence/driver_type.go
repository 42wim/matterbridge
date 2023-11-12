package persistence

import (
	"database/sql"
	"reflect"
)

const (
	UndefinedDriver = iota
	PostgresDriver
	SQLiteDriver
)

func GetDriverType(db *sql.DB) int {
	switch reflect.TypeOf(db.Driver()).String() {
	case "*sqlite3.SQLiteDriver":
		return SQLiteDriver
	case "*stdlib.Driver":
		return PostgresDriver
	}
	return UndefinedDriver
}
