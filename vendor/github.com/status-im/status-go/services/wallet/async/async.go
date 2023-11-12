package async

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

type Command func(context.Context) error

type Commander interface {
	Command(inteval ...time.Duration) Command
}

type Runner interface {
	Run(context.Context) error
}

// SingleShotCommand runs once.
type SingleShotCommand struct {
	Interval time.Duration
	Init     func(context.Context) error
	Runable  func(context.Context) error
}

func (c SingleShotCommand) Run(ctx context.Context) error {
	timer := time.NewTimer(c.Interval)
	if c.Init != nil {
		err := c.Init(ctx)
		if err != nil {
			return err
		}
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			_ = c.Runable(ctx)
		}
	}
}

// FiniteCommand terminates when error is nil.
type FiniteCommand struct {
	Interval time.Duration
	Runable  func(context.Context) error
}

func (c FiniteCommand) Run(ctx context.Context) error {
	err := c.Runable(ctx)
	if err == nil {
		return nil
	}
	ticker := time.NewTicker(c.Interval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			err := c.Runable(ctx)
			if err == nil {
				return nil
			}
		}
	}
}

// InfiniteCommand runs until context is closed.
type InfiniteCommand struct {
	Interval time.Duration
	Runable  func(context.Context) error
}

func (c InfiniteCommand) Run(ctx context.Context) error {
	_ = c.Runable(ctx)
	ticker := time.NewTicker(c.Interval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			_ = c.Runable(ctx)
		}
	}
}

func NewGroup(parent context.Context) *Group {
	ctx, cancel := context.WithCancel(parent)
	return &Group{
		ctx:    ctx,
		cancel: cancel,
	}
}

type Group struct {
	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup
}

func (g *Group) Add(cmd Command) {
	g.wg.Add(1)
	go func() {
		_ = cmd(g.ctx)
		g.wg.Done()
	}()
}

func (g *Group) Stop() {
	g.cancel()
}

func (g *Group) Wait() {
	g.wg.Wait()
}

func (g *Group) WaitAsync() <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		g.Wait()
		close(ch)
	}()
	return ch
}

func NewAtomicGroup(parent context.Context) *AtomicGroup {
	ctx, cancel := context.WithCancel(parent)
	ag := &AtomicGroup{ctx: ctx, cancel: cancel}
	ag.done = ag.onFinish
	return ag
}

// AtomicGroup terminates as soon as first goroutine terminates with error.
type AtomicGroup struct {
	ctx    context.Context
	cancel func()
	done   func()
	wg     sync.WaitGroup

	mu    sync.Mutex
	error error
}

type AtomicGroupKey string

func (d *AtomicGroup) SetName(name string) {
	d.ctx = context.WithValue(d.ctx, AtomicGroupKey("name"), name)
}

func (d *AtomicGroup) Name() string {
	val := d.ctx.Value(AtomicGroupKey("name"))
	if val != nil {
		return val.(string)
	}
	return ""
}

// Go spawns function in a goroutine and stores results or errors.
func (d *AtomicGroup) Add(cmd Command) {
	d.wg.Add(1)
	go func() {
		defer d.done()
		err := cmd(d.ctx)
		d.mu.Lock()
		defer d.mu.Unlock()
		if err != nil {
			// do not overwrite original error by context errors
			if d.error != nil {
				log.Info("async.Command failed", "error", err, "d.error", d.error, "group", d.Name())
				return
			}
			d.error = err
			d.cancel()
			return
		}
	}()
}

// Wait for all downloaders to finish.
func (d *AtomicGroup) Wait() {
	d.wg.Wait()
	if d.Error() == nil {
		d.mu.Lock()
		defer d.mu.Unlock()
		d.cancel()
	}
}

func (d *AtomicGroup) WaitAsync() <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		d.Wait()
		close(ch)
	}()
	return ch
}

// Error stores an error that was reported by any of the downloader. Should be called after Wait.
func (d *AtomicGroup) Error() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.error
}

func (d *AtomicGroup) Stop() {
	d.cancel()
}

func (d *AtomicGroup) onFinish() {
	d.wg.Done()
}

func NewQueuedAtomicGroup(parent context.Context, limit uint32) *QueuedAtomicGroup {
	qag := &QueuedAtomicGroup{NewAtomicGroup(parent), limit, 0, []Command{}, sync.Mutex{}}
	baseDoneFunc := qag.done // save original done function
	qag.AtomicGroup.done = func() {
		baseDoneFunc()
		qag.onFinish()
	}
	return qag
}

type QueuedAtomicGroup struct {
	*AtomicGroup
	limit       uint32
	count       uint32
	pendingCmds []Command
	mu          sync.Mutex
}

func (d *QueuedAtomicGroup) Add(cmd Command) {

	d.mu.Lock()
	if d.limit > 0 && d.count >= d.limit {
		d.pendingCmds = append(d.pendingCmds, cmd)
		d.mu.Unlock()
		return
	}

	d.mu.Unlock()
	d.run(cmd)
}

func (d *QueuedAtomicGroup) run(cmd Command) {
	d.mu.Lock()
	d.count++
	d.mu.Unlock()
	d.AtomicGroup.Add(cmd)
}

func (d *QueuedAtomicGroup) onFinish() {
	d.mu.Lock()
	d.count--

	if d.count < d.limit && len(d.pendingCmds) > 0 {
		cmd := d.pendingCmds[0]
		d.pendingCmds = d.pendingCmds[1:]
		d.mu.Unlock()
		d.run(cmd)
		return
	}

	d.mu.Unlock()
}

func NewErrorCounter(maxErrors int, msg string) *ErrorCounter {
	return &ErrorCounter{maxErrors: maxErrors, msg: msg}
}

type ErrorCounter struct {
	cnt       int
	maxErrors int
	err       error
	msg       string
}

// Returns false in case of counter overflow
func (ec *ErrorCounter) SetError(err error) bool {
	log.Debug("ErrorCounter setError", "msg", ec.msg, "err", err, "cnt", ec.cnt)

	ec.cnt++

	// do not overwrite the first error
	if ec.err == nil {
		ec.err = err
	}

	if ec.cnt >= ec.maxErrors {
		log.Error("ErrorCounter overflow", "msg", ec.msg)
		return false
	}

	return true
}

func (ec *ErrorCounter) Error() error {
	return ec.err
}

func (ec *ErrorCounter) MaxErrors() int {
	return ec.maxErrors
}

type FiniteCommandWithErrorCounter struct {
	FiniteCommand
	*ErrorCounter
}

func (c FiniteCommandWithErrorCounter) Run(ctx context.Context) error {
	f := func(ctx context.Context) (quit bool, err error) {
		err = c.Runable(ctx)
		if err == nil {
			return true, err
		}

		if c.ErrorCounter.SetError(err) {
			return false, err
		}
		return true, err
	}

	quit, err := f(ctx)
	if quit {
		return err
	}

	ticker := time.NewTicker(c.Interval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			quit, err := f(ctx)
			if quit {
				return err
			}
		}
	}
}
