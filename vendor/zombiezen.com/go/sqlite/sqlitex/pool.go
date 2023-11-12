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

package sqlitex

import (
	"context"
	"fmt"
	"sync"

	"zombiezen.com/go/sqlite"
)

// Pool is a pool of SQLite connections.
//
// It is safe for use by multiple goroutines concurrently.
//
// Typically, a goroutine that needs to use an SQLite *Conn
// Gets it from the pool and defers its return:
//
//	conn := dbpool.Get(nil)
//	defer dbpool.Put(conn)
//
// As Get may block, a context can be used to return if a task
// is cancelled. In this case the Conn returned will be nil:
//
//	conn := dbpool.Get(ctx)
//	if conn == nil {
//		return context.Canceled
//	}
//	defer dbpool.Put(conn)
type Pool struct {
	free   chan *sqlite.Conn
	closed chan struct{}

	mu  sync.Mutex
	all map[*sqlite.Conn]context.CancelFunc
}

// Open opens a fixed-size pool of SQLite connections.
// A flags value of 0 defaults to:
//
//	SQLITE_OPEN_READWRITE
//	SQLITE_OPEN_CREATE
//	SQLITE_OPEN_WAL
//	SQLITE_OPEN_URI
//	SQLITE_OPEN_NOMUTEX
func Open(uri string, flags sqlite.OpenFlags, poolSize int) (pool *Pool, err error) {
	if uri == ":memory:" {
		return nil, strerror{msg: `sqlite: ":memory:" does not work with multiple connections, use "file::memory:?mode=memory"`}
	}

	p := &Pool{
		free:   make(chan *sqlite.Conn, poolSize),
		closed: make(chan struct{}),
	}
	defer func() {
		// If an error occurred, call Close outside the lock so this doesn't deadlock.
		if err != nil {
			p.Close()
		}
	}()

	if flags == 0 {
		flags = sqlite.OpenReadWrite |
			sqlite.OpenCreate |
			sqlite.OpenWAL |
			sqlite.OpenURI |
			sqlite.OpenNoMutex
	}

	// TODO(maybe)
	// sqlitex_pool is also defined in package sqlite
	// const sqlitex_pool = sqlite.OpenFlags(0x01000000)
	// flags |= sqlitex_pool

	p.all = make(map[*sqlite.Conn]context.CancelFunc)
	for i := 0; i < poolSize; i++ {
		conn, err := sqlite.OpenConn(uri, flags)
		if err != nil {
			return nil, err
		}
		p.free <- conn
		p.all[conn] = func() {}
	}

	return p, nil
}

// Get returns an SQLite connection from the Pool.
//
// If no Conn is available, Get will block until at least one Conn is returned
// with Put, or until either the Pool is closed or the context is canceled. If
// no Conn can be obtained, nil is returned.
//
// The provided context is also used to control the execution lifetime of the
// connection. See Conn.SetInterrupt for details.
//
// Applications must ensure that all non-nil Conns returned from Get are
// returned to the same Pool with Put.
//
// Although ctx historically may be nil, this is not a recommended design
// pattern.
func (p *Pool) Get(ctx context.Context) *sqlite.Conn {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case conn := <-p.free:
		ctx, cancel := context.WithCancel(ctx)
		// TODO(maybe)
		// conn.SetTracer(&tracer{ctx: ctx})
		conn.SetInterrupt(ctx.Done())

		p.mu.Lock()
		defer p.mu.Unlock()
		p.all[conn] = cancel

		return conn
	case <-ctx.Done():
	case <-p.closed:
	}
	return nil
}

// Put puts an SQLite connection back into the Pool.
//
// Put will panic if the conn was not originally created by p. Put(nil) is a
// no-op.
//
// Applications must ensure that all non-nil Conns returned from Get are
// returned to the same Pool with Put.
func (p *Pool) Put(conn *sqlite.Conn) {
	if conn == nil {
		// See https://github.com/zombiezen/go-sqlite/issues/17
		return
	}
	query := conn.CheckReset()
	if query != "" {
		panic(fmt.Sprintf(
			"connection returned to pool has active statement: %q",
			query))
	}

	p.mu.Lock()
	cancel, found := p.all[conn]
	if found {
		p.all[conn] = func() {}
	}
	p.mu.Unlock()

	if !found {
		panic("sqlite.Pool.Put: connection not created by this pool")
	}

	conn.SetInterrupt(nil)
	cancel()
	p.free <- conn
}

// Close interrupts and closes all the connections in the Pool,
// blocking until all connections are returned to the Pool.
func (p *Pool) Close() (err error) {
	close(p.closed)

	p.mu.Lock()
	n := len(p.all)
	cancelList := make([]context.CancelFunc, 0, n)
	for conn, cancel := range p.all {
		cancelList = append(cancelList, cancel)
		p.all[conn] = func() {}
	}
	p.mu.Unlock()

	for _, cancel := range cancelList {
		cancel()
	}
	for closed := 0; closed < n; closed++ {
		conn := <-p.free
		if err2 := conn.Close(); err == nil {
			err = err2
		}
	}
	return
}

type strerror struct {
	msg string
}

func (err strerror) Error() string { return err.msg }

// TODO(maybe)

// type tracer struct {
// 	ctx       context.Context
// 	ctxStack  []context.Context
// 	taskStack []*trace.Task
// }

// func (t *tracer) pctx() context.Context {
// 	if len(t.ctxStack) != 0 {
// 		return t.ctxStack[len(t.ctxStack)-1]
// 	}
// 	return t.ctx
// }

// func (t *tracer) Push(name string) {
// 	ctx, task := trace.NewTask(t.pctx(), name)
// 	t.ctxStack = append(t.ctxStack, ctx)
// 	t.taskStack = append(t.taskStack, task)
// }

// func (t *tracer) Pop() {
// 	t.taskStack[len(t.taskStack)-1].End()
// 	t.taskStack = t.taskStack[:len(t.taskStack)-1]
// 	t.ctxStack = t.ctxStack[:len(t.ctxStack)-1]
// }

// func (t *tracer) NewTask(name string) sqlite.TracerTask {
// 	ctx, task := trace.NewTask(t.pctx(), name)
// 	return &tracerTask{
// 		ctx:  ctx,
// 		task: task,
// 	}
// }

// type tracerTask struct {
// 	ctx    context.Context
// 	task   *trace.Task
// 	region *trace.Region
// }

// func (t *tracerTask) StartRegion(regionType string) {
// 	if t.region != nil {
// 		panic("sqlitex.tracerTask.StartRegion: already in region")
// 	}
// 	t.region = trace.StartRegion(t.ctx, regionType)
// }

// func (t *tracerTask) EndRegion() {
// 	t.region.End()
// 	t.region = nil
// }

// func (t *tracerTask) End() {
// 	t.task.End()
// }
