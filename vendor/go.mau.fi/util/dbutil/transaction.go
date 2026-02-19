// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dbutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/petermattis/goid"
	"github.com/rs/zerolog"

	"go.mau.fi/util/exerrors"
	"go.mau.fi/util/random"
)

var (
	ErrTxn       = errors.New("transaction")
	ErrTxnBegin  = fmt.Errorf("%w: begin", ErrTxn)
	ErrTxnCommit = fmt.Errorf("%w: commit", ErrTxn)
)

type contextKey int64

const (
	ContextKeyDoTxnCallerSkip contextKey = 1
)

var nextContextKeyDatabaseTransaction atomic.Uint64

func init() {
	nextContextKeyDatabaseTransaction.Store(1 << 61)
}

func (db *Database) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.Execable(ctx).ExecContext(ctx, query, args...)
}

func (db *Database) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	return db.Execable(ctx).QueryContext(ctx, query, args...)
}

func (db *Database) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return db.Execable(ctx).QueryRowContext(ctx, query, args...)
}

var ErrTransactionDeadlock = errors.New("attempt to start new transaction in goroutine with transaction")
var ErrQueryDeadlock = errors.New("attempt to query without context in goroutine with transaction")
var ErrAcquireDeadlock = errors.New("attempt to acquire connection without context in goroutine with transaction")

func (db *Database) BeginTx(ctx context.Context, opts *TxnOptions) (*LoggingTxn, error) {
	if ctx == nil {
		panic("BeginTx() called with nil ctx")
	}
	return db.LoggingDB.BeginTx(ctx, opts)
}

func (db *Database) DoTxn(ctx context.Context, opts *TxnOptions, fn func(ctx context.Context) error) error {
	if ctx == nil {
		panic("DoTxn() called with nil ctx")
	}
	if ctx.Value(db.txnCtxKey) != nil {
		zerolog.Ctx(ctx).Trace().Msg("Already in a transaction, not creating a new one")
		return fn(ctx)
	} else if db.DeadlockDetection {
		goroutineID := goid.Get()
		if !db.txnDeadlockMap.Add(goroutineID) {
			panic(ErrTransactionDeadlock)
		}
		defer db.txnDeadlockMap.Remove(goroutineID)
	}

	log := zerolog.Ctx(ctx).With().Str("db_txn_id", random.String(12)).Logger()
	slowLog := log

	callerSkip := 1
	if val := ctx.Value(ContextKeyDoTxnCallerSkip); val != nil {
		callerSkip += val.(int)
	}
	if pc, file, line, ok := runtime.Caller(callerSkip); ok {
		slowLog = log.With().Str(zerolog.CallerFieldName, zerolog.CallerMarshalFunc(pc, file, line)).Logger()
	}

	start := time.Now()
	deadlockCh := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				slowLog.Warn().
					Float64("duration_seconds", time.Since(start).Seconds()).
					Msg("Transaction still running")
			case <-deadlockCh:
				return
			}
		}
	}()
	defer func() {
		close(deadlockCh)
		dur := time.Since(start)
		if dur > time.Second {
			slowLog.Warn().
				Float64("duration_seconds", dur.Seconds()).
				Msg("Transaction took long")
		}
	}()
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		log.Trace().Err(err).Msg("Failed to begin transaction")
		return exerrors.NewDualError(ErrTxnBegin, err)
	}
	logLevel := zerolog.TraceLevel
	if GlobalSafeQueryLog {
		logLevel = zerolog.DebugLevel
	}
	log.WithLevel(logLevel).Msg("Transaction started")
	tx.noTotalLog = true
	ctx = log.WithContext(ctx)
	ctx = context.WithValue(ctx, db.txnCtxKey, tx)
	err = fn(ctx)
	if err != nil {
		log.Trace().Err(err).Msg("Database transaction failed, rolling back")
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Warn().Err(rollbackErr).Msg("Rollback after transaction error failed")
		} else {
			log.WithLevel(logLevel).Msg("Rollback successful")
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		log.WithLevel(logLevel).Err(err).Msg("Commit failed")
		return exerrors.NewDualError(ErrTxnCommit, err)
	}
	log.WithLevel(logLevel).Msg("Commit successful")
	return nil
}

func (db *Database) Execable(ctx context.Context) Execable {
	if ctx == nil {
		panic("Conn() called with nil ctx")
	}
	txn, ok := ctx.Value(db.txnCtxKey).(Transaction)
	if ok {
		return txn
	}
	if db.DeadlockDetection && db.txnDeadlockMap.Has(goid.Get()) {
		panic(ErrQueryDeadlock)
	}
	return &db.LoggingDB
}

func (db *Database) AcquireConn(ctx context.Context) (Conn, error) {
	if ctx == nil {
		return nil, fmt.Errorf("AcquireConn() called with nil ctx")
	}
	_, ok := ctx.Value(db.txnCtxKey).(Transaction)
	if ok {
		return nil, fmt.Errorf("cannot acquire connection while in a transaction")
	}
	if db.DeadlockDetection && db.txnDeadlockMap.Has(goid.Get()) {
		panic(ErrAcquireDeadlock)
	}
	conn, err := db.RawDB.Conn(ctx)
	if err != nil {
		return nil, err
	}
	return &LoggingExecable{
		UnderlyingExecable: conn,
		db:                 db,
	}, nil
}
