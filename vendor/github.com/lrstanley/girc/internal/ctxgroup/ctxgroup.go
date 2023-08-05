// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ctxgroup

import (
	"context"
	"sync"
)

// A Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
type Group struct {
	ctx    context.Context
	cancel func()

	wg sync.WaitGroup

	errOnce sync.Once
	err     error
}

// New returns a new Group and an associated context derived from ctx.
// Obtain the derived context from calling Group.Context().
//
// The derived context is canceled the first time a function passed to Go
// returns a non-nil error or the first time Wait returns, whichever occurs
// first.
func New(ctx context.Context) *Group {
	nctx, cancel := context.WithCancel(ctx)
	return &Group{ctx: nctx, cancel: cancel}
}

// Context returns the context for this group. It may be canceled by the first
// function to return a non-nil error.
func (g *Group) Context() context.Context {
	return g.ctx
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}

// Go calls the given function in a new goroutine. The first call to return a
// non-nil error cancels the group; its error will be returned by Wait.
func (g *Group) Go(f func(ctx context.Context) error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()

		if err := f(g.ctx); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}
