package sqlite

import "database/sql"

// statementCreator allows to pass transaction or database to use in consumer.
type StatementCreator interface {
	Prepare(query string) (*sql.Stmt, error)
}
