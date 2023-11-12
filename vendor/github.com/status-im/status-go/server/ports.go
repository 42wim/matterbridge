package server

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// portManager is responsible for maintaining segregated access to the port field
type portManger struct {
	logger           *zap.Logger
	port             int
	afterPortChanged func(port int)
}

// newPortManager returns a newly initialised portManager
func newPortManager(logger *zap.Logger, afterPortChanged func(int)) portManger {
	pm := portManger{
		logger:           logger.Named("portManger"),
		afterPortChanged: afterPortChanged,
	}
	return pm
}

// SetPort sets portManger.port field to the given port value
// next triggers any given portManger.afterPortChanged function
func (p *portManger) SetPort(port int) error {
	l := p.logger.Named("SetPort")
	l.Debug("fired", zap.Int("port", port))

	if port == 0 {
		errMsg := "port can not be `0`, use ResetPort() instead"
		l.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	p.port = port
	if p.afterPortChanged != nil {
		l.Debug("p.afterPortChanged != nil")
		p.afterPortChanged(port)
	}
	return nil
}

// ResetPort resets portManger.port to 0
func (p *portManger) ResetPort() {
	l := p.logger.Named("ResetPort")
	l.Debug("fired")

	p.port = 0
}

// GetPort gets the current value of portManager.port without any concern for the state of its value
// and therefore does not wait if portManager.port is 0
func (p *portManger) GetPort() int {
	l := p.logger.Named("GetPort")
	l.Debug("fired")

	return p.port
}

// MustGetPort only returns portManager.port if portManager.port is not 0.
func (p *portManger) MustGetPort() int {
	l := p.logger.Named("MustGetPort")
	l.Debug("fired")

	for {
		if p.port != 0 {
			port := p.port
			if port == 0 {
				panic("port is zero, port has reset")
			}
			return port
		}

		l.Debug("port is zero")
		time.Sleep(20 * time.Millisecond)
	}
}
