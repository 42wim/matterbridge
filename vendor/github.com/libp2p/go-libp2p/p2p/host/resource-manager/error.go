package rcmgr

import (
	"errors"

	"github.com/libp2p/go-libp2p/core/network"
)

type ErrStreamOrConnLimitExceeded struct {
	current, attempted, limit int
	err                       error
}

func (e *ErrStreamOrConnLimitExceeded) Error() string { return e.err.Error() }
func (e *ErrStreamOrConnLimitExceeded) Unwrap() error { return e.err }

// edge may be "" if this is not an edge error
func logValuesStreamLimit(scope, edge string, dir network.Direction, stat network.ScopeStat, err error) []interface{} {
	logValues := make([]interface{}, 0, 2*8)
	logValues = append(logValues, "scope", scope)
	if edge != "" {
		logValues = append(logValues, "edge", edge)
	}
	logValues = append(logValues, "direction", dir)
	var e *ErrStreamOrConnLimitExceeded
	if errors.As(err, &e) {
		logValues = append(logValues,
			"current", e.current,
			"attempted", e.attempted,
			"limit", e.limit,
		)
	}
	return append(logValues, "stat", stat, "error", err)
}

// edge may be "" if this is not an edge error
func logValuesConnLimit(scope, edge string, dir network.Direction, usefd bool, stat network.ScopeStat, err error) []interface{} {
	logValues := make([]interface{}, 0, 2*9)
	logValues = append(logValues, "scope", scope)
	if edge != "" {
		logValues = append(logValues, "edge", edge)
	}
	logValues = append(logValues, "direction", dir, "usefd", usefd)
	var e *ErrStreamOrConnLimitExceeded
	if errors.As(err, &e) {
		logValues = append(logValues,
			"current", e.current,
			"attempted", e.attempted,
			"limit", e.limit,
		)
	}
	return append(logValues, "stat", stat, "error", err)
}

type ErrMemoryLimitExceeded struct {
	current, attempted, limit int64
	priority                  uint8
	err                       error
}

func (e *ErrMemoryLimitExceeded) Error() string { return e.err.Error() }
func (e *ErrMemoryLimitExceeded) Unwrap() error { return e.err }

// edge may be "" if this is not an edge error
func logValuesMemoryLimit(scope, edge string, stat network.ScopeStat, err error) []interface{} {
	logValues := make([]interface{}, 0, 2*8)
	logValues = append(logValues, "scope", scope)
	if edge != "" {
		logValues = append(logValues, "edge", edge)
	}
	var e *ErrMemoryLimitExceeded
	if errors.As(err, &e) {
		logValues = append(logValues,
			"current", e.current,
			"attempted", e.attempted,
			"priority", e.priority,
			"limit", e.limit,
		)
	}
	return append(logValues, "stat", stat, "error", err)
}
