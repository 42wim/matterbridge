// Copyright 2021 Ross Light
// SPDX-License-Identifier: ISC

package sqlitex

import (
	"errors"

	"zombiezen.com/go/sqlite"
)

var errNoResults = errors.New("sqlite: statement has no results")
var errMultipleResults = errors.New("sqlite: statement has multiple result rows")

func resultSetup(stmt *sqlite.Stmt) error {
	hasRow, err := stmt.Step()
	if err != nil {
		stmt.Reset()
		return err
	}
	if !hasRow {
		stmt.Reset()
		return errNoResults
	}
	return nil
}

func resultTeardown(stmt *sqlite.Stmt) error {
	hasRow, err := stmt.Step()
	if err != nil {
		stmt.Reset()
		return err
	}
	if hasRow {
		stmt.Reset()
		return errMultipleResults
	}
	return stmt.Reset()
}

func ResultBool(stmt *sqlite.Stmt) (bool, error) {
	res, err := ResultInt64(stmt)
	return res != 0, err
}

func ResultInt(stmt *sqlite.Stmt) (int, error) {
	res, err := ResultInt64(stmt)
	return int(res), err
}

func ResultInt64(stmt *sqlite.Stmt) (int64, error) {
	if err := resultSetup(stmt); err != nil {
		return 0, err
	}
	res := stmt.ColumnInt64(0)
	if err := resultTeardown(stmt); err != nil {
		return 0, err
	}
	return res, nil
}

func ResultText(stmt *sqlite.Stmt) (string, error) {
	if err := resultSetup(stmt); err != nil {
		return "", err
	}
	res := stmt.ColumnText(0)
	if err := resultTeardown(stmt); err != nil {
		return "", err
	}
	return res, nil
}

func ResultFloat(stmt *sqlite.Stmt) (float64, error) {
	if err := resultSetup(stmt); err != nil {
		return 0, err
	}
	res := stmt.ColumnFloat(0)
	if err := resultTeardown(stmt); err != nil {
		return 0, err
	}
	return res, nil
}
