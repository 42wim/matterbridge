package rtt

import (
	"context"
	"sync"
	"time"

	errors "github.com/pkg/errors"
	tcp "github.com/status-im/tcp-shaker"
)

type Result struct {
	Addr  string
	RTTMs int
	Err   error
}

// timeoutError indicates an error due to TCP connection timeout.
// tcp-shaker returns an error implementing this interface in such a case.
type timeoutError interface {
	Timeout() bool
}

func runCheck(c *tcp.Checker, address string, timeout time.Duration) Result {
	// mesaure RTT
	start := time.Now()
	// TCP Ping
	err := c.CheckAddr(address, timeout)
	// measure RTT
	elapsed := time.Since(start)
	latency := int(elapsed.Nanoseconds() / 1e6)

	if err != nil { // don't confuse users with valid latency values on error
		latency = -1
		switch err.(type) {
		case timeoutError:
			err = errors.Wrap(err, "tcp check timeout")
		case tcp.ErrConnect:
			err = errors.Wrap(err, "unable to connect")
		}
	}

	return Result{
		Addr:  address,
		RTTMs: latency,
		Err:   err,
	}
}

func waitForResults(errCh <-chan error, resCh <-chan Result) (results []Result, err error) {
	for {
		select {
		case err = <-errCh:
			return nil, err
		case res, ok := <-resCh:
			if !ok {
				return
			}
			results = append(results, res)
		}
	}
}

func CheckHosts(addresses []string, timeout time.Duration) ([]Result, error) {
	c := tcp.NewChecker()

	// channel for receiving possible checking loop failure
	errCh := make(chan error, 1)

	// stop the checking loop when function exists
	ctx, stopChecker := context.WithCancel(context.Background())
	defer stopChecker()

	// loop that queries Epoll and pipes events to CheckAddr() calls
	go func() {
		errCh <- c.CheckingLoop(ctx)
	}()
	// wait for CheckingLoop to prepare the epoll/kqueue
	<-c.WaitReady()

	// channel for returning results from concurrent checks
	resCh := make(chan Result, len(addresses))

	var wg sync.WaitGroup
	for i := 0; i < len(addresses); i++ {
		wg.Add(1)
		go func(address string, resCh chan<- Result) {
			defer wg.Done()
			resCh <- runCheck(c, address, timeout)
		}(addresses[i], resCh)
	}
	// wait for all the routines to finish before closing results channel
	wg.Wait()
	close(resCh)

	// wait for the results for all addresses or a checking loop error
	return waitForResults(errCh, resCh)
}
