// +build darwin

package tcp

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// Checker contains an epoll instance for TCP handshake checking
type Checker struct {
	pipePool
	resultPipes
	pollerLock sync.Mutex
	_kqueueFd  int32
	zeroLinger bool
	isReady    chan struct{}
}

// NewChecker creates a Checker with linger set to zero.
func NewChecker() *Checker {
	return NewCheckerZeroLinger(true)
}

// NewCheckerZeroLinger creates a Checker with zeroLinger set to given value.
func NewCheckerZeroLinger(zeroLinger bool) *Checker {
	return &Checker{
		pipePool:    newPipePoolSyncPool(),
		resultPipes: newResultPipesSyncMap(),
		_kqueueFd:   -1,
		zeroLinger:  zeroLinger,
		isReady:     make(chan struct{}),
	}
}

// CheckingLoop must be called before anything else.
// NOTE: this function blocks until ctx got canceled.
func (c *Checker) CheckingLoop(ctx context.Context) error {
	kqueue, err := c.createPoller()
	if err != nil {
		return errors.Wrap(err, "error creating poller")
	}
	defer c.closePoller()

	c.setReady()
	defer c.resetReady()

	return c.pollingLoop(ctx, kqueue)
}

func (c *Checker) createPoller() (int, error) {
	c.pollerLock.Lock()
	defer c.pollerLock.Unlock()

	if c.getKQueueAtomic() > 0 {
		// return if already initialized
		return -1, ErrCheckerAlreadyStarted
	}

	kqueue, err := createPoller()
	if err != nil {
		return -1, err
	}
	c.setKQueueAtomic(kqueue)

	return kqueue, nil
}

func (c *Checker) closePoller() error {
	c.pollerLock.Lock()
	defer c.pollerLock.Unlock()
	var err error
	if c.getKQueueAtomic() > 0 {
		err = unix.Close(c.getKQueueAtomic())
	}
	c.setKQueueAtomic(-1)
	return err
}

func (c *Checker) setReady() {
	close(c.isReady)
}

func (c *Checker) resetReady() {
	c.isReady = make(chan struct{})
}

const pollerTimeout = time.Second

func (c *Checker) pollingLoop(ctx context.Context, kqueue int) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			evts, err := pollEvents(kqueue, pollerTimeout)
			if err != nil {
				// fatal error
				return errors.Wrap(err, "error during polling loop")
			}

			c.handlePollerEvents(evts)
		}
	}
}

func (c *Checker) handlePollerEvents(evts []event) {
	for _, e := range evts {
		if pipe, exists := c.resultPipes.popResultPipe(e.Fd); exists {
			pipe <- e.Err
		}
		// error pipe not found
		// in this case, e.Fd should have been handled in the previous event.
	}
}

func (c *Checker) getKQueueAtomic() int {
	return int(atomic.LoadInt32(&c._kqueueFd))
}

func (c *Checker) setKQueueAtomic(fd int) {
	atomic.StoreInt32(&c._kqueueFd, int32(fd))
}

// CheckAddr performs a TCP check with given TCP address and timeout
// A successful check will result in nil error
// ErrTimeout is returned if timeout
// zeroLinger is an optional parameter indicating if linger should be set to zero
// for this particular connection
// Note: timeout includes domain resolving
func (c *Checker) CheckAddr(addr string, timeout time.Duration) (err error) {
	return c.CheckAddrZeroLinger(addr, timeout, c.zeroLinger)
}

// CheckAddrZeroLinger is like CheckAddr with an extra parameter indicating whether to enable zero linger.
func (c *Checker) CheckAddrZeroLinger(addr string, timeout time.Duration, zeroLinger bool) error {
	// Set deadline
	deadline := time.Now().Add(timeout)

	// Parse address
	rAddr, err := parseSockAddr(addr)
	if err != nil {
		return err
	}
	// Create socket with options set
	fd, err := createSocketZeroLinger(zeroLinger)
	if err != nil {
		return err
	}
	// Socket closes after the right socket event returns
	defer unix.Close(fd)

	// Connect to the address
	if success, cErr := connect(fd, rAddr); cErr != nil {
		// If there was an error, return it.
		return &ErrConnect{cErr}
	} else if success {
		// If the connect was successful, we are done.
		return nil
	}
	// Otherwise wait for the result of connect.
	return c.waitConnectResult(fd, deadline.Sub(time.Now()))
}

func (c *Checker) waitConnectResult(fd int, timeout time.Duration) error {
	// get a pipe of connect result
	resultPipe := c.getPipe()
	defer func() {
		c.resultPipes.deregisterResultPipe(fd)
		c.putBackPipe(resultPipe)
	}()

	// this must be done before registerEvents
	c.resultPipes.registerResultPipe(fd, resultPipe)
	// Register to epoll for later error checking
	err := registerEvents(c.KQueue(), fd)
	if err != nil {
		return err
	}

	// Wait for connect result
	err = c.waitPipeTimeout(resultPipe, timeout)

	return err
}

func (c *Checker) waitPipeTimeout(pipe chan error, timeout time.Duration) error {
	select {
	case ret := <-pipe:
		return ret
	case <-time.After(timeout):
		return ErrTimeout
	}
}

// WaitReady returns a chan which is closed when the Checker is ready for use.
func (c *Checker) WaitReady() <-chan struct{} {
	return c.isReady
}

// IsReady returns a bool indicates whether the Checker is ready for use
func (c *Checker) IsReady() bool {
	return c.getKQueueAtomic() > 0
}

// PollerFd returns the inner fd of poller instance.
// NOTE: Use this only when you really know what you are doing.
func (c *Checker) KQueue() int {
	return c.getKQueueAtomic()
}
