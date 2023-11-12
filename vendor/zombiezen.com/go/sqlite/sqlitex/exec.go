// Copyright (c) 2018 David Crawshaw <david@zentus.com>
// Copyright (c) 2021 Ross Light <rosss@zombiezen.com>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
//
// SPDX-License-Identifier: ISC

// Package sqlitex provides utilities for working with SQLite.
package sqlitex

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/fs"
)

// ExecOptions is the set of optional arguments executing a statement.
type ExecOptions struct {
	// Args is the set of positional arguments to bind to the statement. The first
	// element in the slice is ?1. See https://sqlite.org/lang_expr.html for more
	// details.
	Args []interface{}
	// Named is the set of named arguments to bind to the statement. Keys must
	// start with ':', '@', or '$'. See https://sqlite.org/lang_expr.html for more
	// details.
	Named map[string]interface{}
	// ResultFunc is called for each result row. If ResultFunc returns an error
	// then iteration ceases and Exec returns the error value.
	ResultFunc func(stmt *sqlite.Stmt) error
}

// Exec executes an SQLite query.
//
// For each result row, the resultFn is called.
// Result values can be read by resultFn using stmt.Column* methods.
// If resultFn returns an error then iteration ceases and Exec returns
// the error value.
//
// Any args provided to Exec are bound to numbered parameters of the
// query using the Stmt Bind* methods. Basic reflection on args is used
// to map:
//
//	integers to BindInt64
//	floats   to BindFloat
//	[]byte   to BindBytes
//	string   to BindText
//	bool     to BindBool
//
// All other kinds are printed using fmt.Sprintf("%v", v) and passed
// to BindText.
//
// Exec is implemented using the Stmt prepare mechanism which allows
// better interactions with Go's type system and avoids pitfalls of
// passing a Go closure to cgo.
//
// As Exec is implemented using Conn.Prepare, subsequent calls to Exec
// with the same statement will reuse the cached statement object.
//
// Typical use:
//
//	conn := dbpool.Get()
//	defer dbpool.Put(conn)
//
//	if err := sqlitex.Exec(conn, "INSERT INTO t (a, b, c, d) VALUES (?, ?, ?, ?);", nil, "a1", 1, 42, 1); err != nil {
//		// handle err
//	}
//
//	var a []string
//	var b []int64
//	fn := func(stmt *sqlite.Stmt) error {
//		a = append(a, stmt.ColumnText(0))
//		b = append(b, stmt.ColumnInt64(1))
//		return nil
//	}
//	err := sqlutil.Exec(conn, "SELECT a, b FROM t WHERE c = ? AND d = ?;", fn, 42, 1)
//	if err != nil {
//		// handle err
//	}
func Exec(conn *sqlite.Conn, query string, resultFn func(stmt *sqlite.Stmt) error, args ...interface{}) error {
	stmt, err := conn.Prepare(query)
	if err != nil {
		return annotateErr(err)
	}
	err = exec(stmt, &ExecOptions{
		Args:       args,
		ResultFunc: resultFn,
	})
	resetErr := stmt.Reset()
	if err == nil {
		err = resetErr
	}
	return err
}

// ExecFS executes the single statement in the given SQL file.
// ExecFS is implemented using Conn.Prepare, so subsequent calls to ExecFS with the
// same statement will reuse the cached statement object.
func ExecFS(conn *sqlite.Conn, fsys fs.FS, filename string, opts *ExecOptions) error {
	query, err := readString(fsys, filename)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	stmt, err := conn.Prepare(strings.TrimSpace(query))
	if err != nil {
		return fmt.Errorf("exec %s: %w", filename, err)
	}
	err = exec(stmt, opts)
	resetErr := stmt.Reset()
	if err != nil {
		// Don't strip the error query: we already do this inside exec.
		return fmt.Errorf("exec %s: %w", filename, err)
	}
	if resetErr != nil {
		return fmt.Errorf("exec %s: %w", filename, err)
	}
	return nil
}

// ExecTransient executes an SQLite query without caching the
// underlying query.
// The interface is exactly the same as Exec.
//
// It is the spiritual equivalent of sqlite3_exec.
func ExecTransient(conn *sqlite.Conn, query string, resultFn func(stmt *sqlite.Stmt) error, args ...interface{}) (err error) {
	var stmt *sqlite.Stmt
	var trailingBytes int
	stmt, trailingBytes, err = conn.PrepareTransient(query)
	if err != nil {
		return annotateErr(err)
	}
	defer func() {
		ferr := stmt.Finalize()
		if err == nil {
			err = ferr
		}
	}()
	if trailingBytes != 0 {
		return fmt.Errorf("sqlitex.Exec: query %q has trailing bytes", query)
	}
	return exec(stmt, &ExecOptions{
		Args:       args,
		ResultFunc: resultFn,
	})
}

// ExecTransientFS executes the single statement in the given SQL file without
// caching the underlying query.
func ExecTransientFS(conn *sqlite.Conn, fsys fs.FS, filename string, opts *ExecOptions) error {
	query, err := readString(fsys, filename)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	stmt, _, err := conn.PrepareTransient(strings.TrimSpace(query))
	if err != nil {
		return fmt.Errorf("exec %s: %w", filename, err)
	}
	defer stmt.Finalize()
	err = exec(stmt, opts)
	resetErr := stmt.Reset()
	if err != nil {
		// Don't strip the error query: we already do this inside exec.
		return fmt.Errorf("exec %s: %w", filename, err)
	}
	if resetErr != nil {
		return fmt.Errorf("exec %s: %w", filename, err)
	}
	return nil
}

// PrepareTransientFS prepares an SQL statement from a file that is not cached by
// the Conn. Subsequent calls with the same query will create new Stmts.
// The caller is responsible for calling Finalize on the returned Stmt when the
// Stmt is no longer needed.
func PrepareTransientFS(conn *sqlite.Conn, fsys fs.FS, filename string) (*sqlite.Stmt, error) {
	query, err := readString(fsys, filename)
	if err != nil {
		return nil, fmt.Errorf("prepare: %w", err)
	}
	stmt, _, err := conn.PrepareTransient(strings.TrimSpace(query))
	if err != nil {
		return nil, fmt.Errorf("prepare %s: %w", filename, err)
	}
	return stmt, nil
}

func exec(stmt *sqlite.Stmt, opts *ExecOptions) (err error) {
	if opts != nil {
		for i, arg := range opts.Args {
			setArg(stmt, i+1, reflect.ValueOf(arg))
		}
		if err := setNamed(stmt, opts.Named); err != nil {
			return err
		}
	}
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return err
		}
		if !hasRow {
			break
		}
		if opts != nil && opts.ResultFunc != nil {
			if err := opts.ResultFunc(stmt); err != nil {
				return err
			}
		}
	}
	return nil
}

func setArg(stmt *sqlite.Stmt, i int, v reflect.Value) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		stmt.BindInt64(i, v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		stmt.BindInt64(i, int64(v.Uint()))
	case reflect.Float32, reflect.Float64:
		stmt.BindFloat(i, v.Float())
	case reflect.String:
		stmt.BindText(i, v.String())
	case reflect.Bool:
		stmt.BindBool(i, v.Bool())
	case reflect.Invalid:
		stmt.BindNull(i)
	default:
		if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Uint8 {
			stmt.BindBytes(i, v.Bytes())
		} else {
			stmt.BindText(i, fmt.Sprint(v.Interface()))
		}
	}
}

func setNamed(stmt *sqlite.Stmt, args map[string]interface{}) error {
	if len(args) == 0 {
		return nil
	}
	for i, count := 1, stmt.BindParamCount(); i <= count; i++ {
		name := stmt.BindParamName(i)
		if name == "" {
			continue
		}
		arg, present := args[name]
		if !present {
			return fmt.Errorf("missing parameter %s", name)
		}
		setArg(stmt, i, reflect.ValueOf(arg))
	}
	return nil
}

func annotateErr(err error) error {
	// TODO(maybe)
	// if err, isError := err.(sqlite.Error); isError {
	// 	if err.Loc == "" {
	// 		err.Loc = "Exec"
	// 	} else {
	// 		err.Loc = "Exec: " + err.Loc
	// 	}
	// 	return err
	// }
	return fmt.Errorf("sqlutil.Exec: %w", err)
}

// ExecScript executes a script of SQL statements.
//
// The script is wrapped in a SAVEPOINT transaction,
// which is rolled back on any error.
func ExecScript(conn *sqlite.Conn, queries string) (err error) {
	defer Save(conn)(&err)

	for {
		queries = strings.TrimSpace(queries)
		if queries == "" {
			break
		}
		var stmt *sqlite.Stmt
		var trailingBytes int
		stmt, trailingBytes, err = conn.PrepareTransient(queries)
		if err != nil {
			return err
		}
		usedBytes := len(queries) - trailingBytes
		queries = queries[usedBytes:]
		_, err := stmt.Step()
		stmt.Finalize()
		if err != nil {
			return err
		}
	}
	return nil
}

// ExecScriptFS executes a script of SQL statements from a file.
//
// The script is wrapped in a SAVEPOINT transaction, which is rolled back on
// any error.
func ExecScriptFS(conn *sqlite.Conn, fsys fs.FS, filename string, opts *ExecOptions) (err error) {
	queries, err := readString(fsys, filename)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	defer Save(conn)(&err)
	for {
		queries = strings.TrimSpace(queries)
		if queries == "" {
			return nil
		}
		stmt, trailingBytes, err := conn.PrepareTransient(queries)
		if err != nil {
			return fmt.Errorf("exec %s: %w", filename, err)
		}
		usedBytes := len(queries) - trailingBytes
		queries = queries[usedBytes:]
		err = exec(stmt, opts)
		stmt.Finalize()
		if err != nil {
			return fmt.Errorf("exec %s: %w", filename, err)
		}
	}
}

func readString(fsys fs.FS, filename string) (string, error) {
	f, err := fsys.Open(filename)
	if err != nil {
		return "", err
	}
	content := new(strings.Builder)
	_, err = io.Copy(content, f)
	f.Close()
	if err != nil {
		return "", fmt.Errorf("%s: %w", filename, err)
	}
	return content.String(), nil
}
