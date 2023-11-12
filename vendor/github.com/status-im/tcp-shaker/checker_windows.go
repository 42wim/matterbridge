// +build windows
// WARNING: This is a dummy package, windows suppor is not implemented yet.

package tcp

import (
	"context"
	"time"
)

// Checker is a fake implementation.
type Checker struct {
	isReady chan struct{}
}

func NewChecker() *Checker {
	return NewCheckerZeroLinger(true)
}

func NewCheckerZeroLinger(zeroLinger bool) *Checker {
	isReady := make(chan struct{})
	close(isReady)
	return &Checker{isReady: isReady}
}

func (c *Checker) CheckingLoop(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (c *Checker) CheckAddr(addr string, timeout time.Duration) error {
	return c.CheckAddrZeroLinger(addr, timeout, false)
}

func (c *Checker) CheckAddrZeroLinger(addr string, timeout time.Duration, zeroLinger bool) error {
	return nil
}

func (c *Checker) IsReady() bool {
	return true
}

func (c *Checker) WaitReady() <-chan struct{} {
	return c.isReady
}

func (c *Checker) Close() error {
	return nil
}
