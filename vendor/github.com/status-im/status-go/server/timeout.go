package server

import (
	"sync"
	"time"
)

// timeoutManager represents a discrete encapsulation of timeout functionality.
// this struct expose 3 functions:
//   - SetTimeout
//   - StartTimeout
//   - StopTimeout
type timeoutManager struct {
	// timeout number of milliseconds the timeout operation will run before executing the `terminate` func()
	// 0 represents an inactive timeout
	timeout uint

	// exitQueue handles the cancel signal channels that circumvent timeout operations and prevent the
	// execution of any `terminate` func()
	exitQueue *exitQueueManager
}

// newTimeoutManager returns a fully qualified and initialised timeoutManager
func newTimeoutManager() *timeoutManager {
	return &timeoutManager{
		exitQueue: &exitQueueManager{queue: []chan struct{}{}},
	}
}

// SetTimeout sets the value of the timeoutManager.timeout
func (t *timeoutManager) SetTimeout(milliseconds uint) {
	t.timeout = milliseconds
}

// StartTimeout starts a timeout operation based on the set timeoutManager.timeout value
// the given terminate func() will be executed once the timeout duration has passed
func (t *timeoutManager) StartTimeout(terminate func()) {
	if t.timeout == 0 {
		return
	}
	t.StopTimeout()

	exit := make(chan struct{}, 1)
	t.exitQueue.add(exit)
	go t.run(terminate, exit)
}

// StopTimeout terminates a timeout operation and exits gracefully
func (t *timeoutManager) StopTimeout() {
	if t.timeout == 0 {
		return
	}
	t.exitQueue.empty()
}

// run inits the main timeout run function that awaits for the exit command to be triggered or for the
// timeout duration to elapse and trigger the parameter terminate function.
func (t *timeoutManager) run(terminate func(), exit chan struct{}) {
	select {
	case <-exit:
		return
	case <-time.After(time.Duration(t.timeout) * time.Millisecond):
		terminate()
		// TODO fire signal to let UI know
		//  https://github.com/status-im/status-go/issues/3305
		return
	}
}

// exitQueueManager
type exitQueueManager struct {
	queue     []chan struct{}
	queueLock sync.Mutex
}

// add handles new exit channels adding them to the exit queue
func (e *exitQueueManager) add(exit chan struct{}) {
	e.queueLock.Lock()
	defer e.queueLock.Unlock()

	e.queue = append(e.queue, exit)
}

// empty sends a signal to every exit channel in the queue and then resets the queue
func (e *exitQueueManager) empty() {
	e.queueLock.Lock()
	defer e.queueLock.Unlock()

	for i := range e.queue {
		e.queue[i] <- struct{}{}
	}

	e.queue = []chan struct{}{}
}
