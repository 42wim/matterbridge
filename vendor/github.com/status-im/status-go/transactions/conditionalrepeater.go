package transactions

import (
	"context"
	"sync"
	"time"
)

// TaskFunc defines the task to be run. The context is canceled when Stop is
// called to early stop scheduled task.
type TaskFunc func(ctx context.Context) (done bool)

const (
	WorkNotDone = false
	WorkDone    = true
)

// ConditionalRepeater runs a task at regular intervals until the task returns
// true. It doesn't allow running task in parallel and can be triggered early
// by call to RunUntilDone.
type ConditionalRepeater struct {
	interval time.Duration
	task     TaskFunc
	// nil if not running
	ctx      context.Context
	ctxMu    sync.Mutex
	cancel   context.CancelFunc
	runNowCh chan bool
	runNowMu sync.Mutex
}

func NewConditionalRepeater(interval time.Duration, task TaskFunc) *ConditionalRepeater {
	return &ConditionalRepeater{
		interval: interval,
		task:     task,
		runNowCh: make(chan bool, 1),
	}
}

// RunUntilDone starts the task immediately and continues to run it at the defined
// interval until the task returns true. Can be called multiple times but it
// does not allow multiple concurrent executions of the task.
func (t *ConditionalRepeater) RunUntilDone() {
	t.ctxMu.Lock()
	defer func() {
		t.runNowMu.Lock()
		if len(t.runNowCh) == 0 {
			t.runNowCh <- true
		}
		t.runNowMu.Unlock()
		t.ctxMu.Unlock()
	}()

	if t.ctx != nil {
		return
	}
	t.ctx, t.cancel = context.WithCancel(context.Background())

	go func() {
		defer func() {
			t.ctxMu.Lock()
			defer t.ctxMu.Unlock()
			t.cancel()
			t.ctx = nil
		}()

		ticker := time.NewTicker(t.interval)
		defer ticker.Stop()

		for {
			select {
			// Stop was called or task returned true
			case <-t.ctx.Done():
				return
			// Scheduled execution
			case <-ticker.C:
				if t.task(t.ctx) {
					return
				}
			// Start right away if requested
			case <-t.runNowCh:
				ticker.Reset(t.interval)
				if t.task(t.ctx) {
					t.runNowMu.Lock()
					if len(t.runNowCh) == 0 {
						t.runNowMu.Unlock()
						return
					}
					t.runNowMu.Unlock()
				}
			}
		}
	}()
}

// Stop forcefully stops the running task by canceling its context.
func (t *ConditionalRepeater) Stop() {
	t.ctxMu.Lock()
	defer t.ctxMu.Unlock()
	if t.ctx != nil {
		t.cancel()
	}
}

func (t *ConditionalRepeater) IsRunning() bool {
	t.ctxMu.Lock()
	defer t.ctxMu.Unlock()
	return t.ctx != nil
}
