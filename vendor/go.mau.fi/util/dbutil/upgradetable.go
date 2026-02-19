// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dbutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type TxnMode string

const (
	TxnModeOn                   TxnMode = "on"
	TxnModeOff                  TxnMode = "off"
	TxnModeSQLiteForeignKeysOff TxnMode = "sqlite-fkey-off"
)

type UpgradeTable []upgrade

func (ut *UpgradeTable) extend(toSize int) {
	if cap(*ut) >= toSize {
		*ut = (*ut)[:toSize]
	} else {
		resized := make([]upgrade, toSize)
		copy(resized, *ut)
		*ut = resized
	}
}

func (ut *UpgradeTable) Register(from, to, compat int, message string, txn TxnMode, fn upgradeFunc) {
	if from < 0 {
		from += to
	}
	if from < 0 {
		panic("invalid from value in UpgradeTable.Register() call")
	}
	if compat <= 0 {
		compat = to
	}
	upg := upgrade{message: message, fn: fn, upgradesTo: to, compatVersion: compat, transaction: txn}
	if len(*ut) == from {
		*ut = append(*ut, upg)
		return
	} else if len(*ut) < from {
		ut.extend(from + 1)
	} else if (*ut)[from].fn != nil {
		panic(fmt.Errorf("tried to override upgrade at %d ('%s') with '%s'", from, (*ut)[from].message, upg.message))
	}
	(*ut)[from] = upg
}

var upgradeHeaderRegex = regexp.MustCompile(`^-- (?:v(\d+) -> )?v(\d+)(?: \(compatible with v(\d+)\+\))?: (.+)$`)

var transactionDisableRegex = regexp.MustCompile(`^-- transaction: ([a-z-]*)`)

func parseFileHeader(file []byte) (from, to, compat int, message string, txn TxnMode, lines [][]byte, err error) {
	lines = bytes.Split(file, []byte("\n"))
	if len(lines) < 2 {
		err = errors.New("upgrade file too short")
		return
	}
	var maybeFrom int
	match := upgradeHeaderRegex.FindSubmatch(lines[0])
	lines = lines[1:]
	if match == nil {
		err = errors.New("header not found")
	} else if len(match) != 5 {
		err = errors.New("unexpected number of items in regex match")
	} else if maybeFrom, err = strconv.Atoi(string(match[1])); len(match[1]) > 0 && err != nil {
		err = fmt.Errorf("invalid source version: %w", err)
	} else if to, err = strconv.Atoi(string(match[2])); err != nil {
		err = fmt.Errorf("invalid target version: %w", err)
	} else if compat, err = strconv.Atoi(string(match[3])); len(match[3]) > 0 && err != nil {
		err = fmt.Errorf("invalid compatible version: %w", err)
	} else {
		err = nil
		if len(match[1]) > 0 {
			from = maybeFrom
		} else {
			from = -1
		}
		message = string(match[4])
		txn = "on"
		match = transactionDisableRegex.FindSubmatch(lines[0])
		if match != nil {
			lines = lines[1:]
			txn = TxnMode(match[1])
			switch txn {
			case TxnModeOff, TxnModeOn, TxnModeSQLiteForeignKeysOff:
				// ok
			default:
				err = fmt.Errorf("invalid value %q for transaction flag", match[1])
			}
		}
	}
	return
}

var dialectLineFilter = regexp.MustCompile(`^\s*-- only: (postgres|sqlite)(?: for next (\d+) lines| until "(end) only")?(?: \(lines? (commented)\))?`)

// Constants used to make parseDialectFilter clearer
const (
	skipUntilEndTag = -1
	skipNothing     = 0
	skipNextLine    = 1
)

func (db *Database) parseDialectFilter(line []byte) (dialect Dialect, lineCount int, uncomment bool, err error) {
	match := dialectLineFilter.FindSubmatch(line)
	if match == nil {
		return
	}
	dialect, err = ParseDialect(string(match[1]))
	if err != nil {
		return
	}
	uncomment = bytes.Equal(match[4], []byte("commented"))
	if bytes.Equal(match[3], []byte("end")) {
		lineCount = skipUntilEndTag
	} else if len(match[2]) == 0 {
		lineCount = skipNextLine
	} else {
		lineCount, err = strconv.Atoi(string(match[2]))
		if err != nil {
			err = fmt.Errorf("invalid line count %q: %w", match[2], err)
		}
	}
	return
}

var endLineFilter = regexp.MustCompile(`^\s*-- end only (postgres|sqlite)$`)

func (db *Database) Internals() *publishDatabaseInternals {
	return (*publishDatabaseInternals)(db)
}

type publishDatabaseInternals Database

func (di *publishDatabaseInternals) FilterSQLUpgrade(lines [][]byte) (string, error) {
	return (*Database)(di).filterSQLUpgrade(lines)
}

func (db *Database) filterSQLUpgrade(lines [][]byte) (string, error) {
	output := make([][]byte, 0, len(lines))
	for i := 0; i < len(lines); i++ {
		dialect, lineCount, uncomment, err := db.parseDialectFilter(lines[i])
		if err != nil {
			return "", err
		} else if lineCount == skipNothing {
			output = append(output, lines[i])
		} else if lineCount == skipUntilEndTag {
			startedAt := i
			startedAtMatch := dialectLineFilter.FindSubmatch(lines[startedAt])
			// Skip filter start line
			i++
			for ; i < len(lines); i++ {
				if match := endLineFilter.FindSubmatch(lines[i]); match != nil {
					if !bytes.Equal(match[1], startedAtMatch[1]) {
						return "", fmt.Errorf(`unexpected end tag %q for %q start at line %d`, string(match[0]), string(startedAtMatch[1]), startedAt)
					}
					break
				}
				if dialect == db.Dialect {
					if uncomment {
						if !bytes.HasPrefix(lines[i], []byte("--")) {
							return "", fmt.Errorf("line %d isn't commented even though the dialect filter claimed it is (-- must be at beginning of line)", i+1)
						}
						output = append(output, bytes.TrimPrefix(lines[i], []byte("--")))
					} else {
						output = append(output, lines[i])
					}
				}
			}
			if i == len(lines) {
				return "", fmt.Errorf(`didn't get end tag matching start %q at line %d`, string(startedAtMatch[1]), startedAt)
			}
		} else if dialect != db.Dialect {
			i += lineCount
		} else {
			// Skip current line, uncomment the specified number of lines
			i++
			targetI := i + lineCount
			for ; i < targetI; i++ {
				if uncomment {
					output = append(output, bytes.TrimPrefix(lines[i], []byte("--")))
				} else {
					output = append(output, lines[i])
				}
			}
			// Decrement counter to avoid skipping the next line
			i--
		}
	}
	return string(bytes.Join(output, []byte("\n"))), nil
}

func sqlUpgradeFunc(fileName string, lines [][]byte) upgradeFunc {
	return func(ctx context.Context, db *Database) error {
		if dialect, skip, _, err := db.parseDialectFilter(lines[0]); err == nil && skip == skipNextLine && dialect != db.Dialect {
			return nil
		} else if upgradeSQL, err := db.filterSQLUpgrade(lines); err != nil {
			panic(fmt.Errorf("failed to parse upgrade %s: %w", fileName, err))
		} else {
			_, err = db.Exec(ctx, upgradeSQL)
			return err
		}
	}
}

func splitSQLUpgradeFunc(sqliteData, postgresData string) upgradeFunc {
	return func(ctx context.Context, db *Database) (err error) {
		switch db.Dialect {
		case SQLite:
			_, err = db.Exec(ctx, sqliteData)
		case Postgres:
			_, err = db.Exec(ctx, postgresData)
		default:
			err = fmt.Errorf("unknown dialect %s", db.Dialect)
		}
		return
	}
}

func parseSplitSQLUpgrade(name string, fs fullFS, skipNames map[string]struct{}) (from, to, compat int, message string, txn TxnMode, fn upgradeFunc) {
	postgresName := fmt.Sprintf("%s.postgres.sql", name)
	sqliteName := fmt.Sprintf("%s.sqlite.sql", name)
	skipNames[postgresName] = struct{}{}
	skipNames[sqliteName] = struct{}{}
	postgresData, err := fs.ReadFile(postgresName)
	if err != nil {
		panic(err)
	}
	sqliteData, err := fs.ReadFile(sqliteName)
	if err != nil {
		panic(err)
	}
	from, to, compat, message, txn, _, err = parseFileHeader(postgresData)
	if err != nil {
		panic(fmt.Errorf("failed to parse header in %s: %w", postgresName, err))
	}
	sqliteFrom, sqliteTo, sqliteCompat, sqliteMessage, sqliteTxn, _, err := parseFileHeader(sqliteData)
	if err != nil {
		panic(fmt.Errorf("failed to parse header in %s: %w", sqliteName, err))
	}
	if from != sqliteFrom || to != sqliteTo || compat != sqliteCompat {
		panic(fmt.Errorf("mismatching versions in postgres and sqlite versions of %s: %d/%d -> %d/%d", name, from, sqliteFrom, to, sqliteTo))
	} else if message != sqliteMessage {
		panic(fmt.Errorf("mismatching message in postgres and sqlite versions of %s: %q != %q", name, message, sqliteMessage))
	} else if txn != sqliteTxn {
		panic(fmt.Errorf("mismatching transaction flag in postgres and sqlite versions of %s: %s != %s", name, txn, sqliteTxn))
	}
	fn = splitSQLUpgradeFunc(string(sqliteData), string(postgresData))
	return
}

type fullFS interface {
	fs.ReadFileFS
	fs.ReadDirFS
}

var splitFileNameRegex = regexp.MustCompile(`^(.+)\.(postgres|sqlite)\.sql$`)

func (ut *UpgradeTable) RegisterFS(fs fullFS) {
	ut.RegisterFSPath(fs, ".")
}

func (ut *UpgradeTable) RegisterFSPath(fs fullFS, dir string) {
	files, err := fs.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	skipNames := map[string]struct{}{}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			// do nothing
		} else if _, skip := skipNames[file.Name()]; skip {
			// also do nothing
		} else if splitName := splitFileNameRegex.FindStringSubmatch(file.Name()); splitName != nil {
			from, to, compat, message, txn, fn := parseSplitSQLUpgrade(splitName[1], fs, skipNames)
			ut.Register(from, to, compat, message, txn, fn)
		} else if data, err := fs.ReadFile(filepath.Join(dir, file.Name())); err != nil {
			panic(err)
		} else if from, to, compat, message, txn, lines, err := parseFileHeader(data); err != nil {
			panic(fmt.Errorf("failed to parse header in %s: %w", file.Name(), err))
		} else {
			ut.Register(from, to, compat, message, txn, sqlUpgradeFunc(file.Name(), lines))
		}
	}
}
