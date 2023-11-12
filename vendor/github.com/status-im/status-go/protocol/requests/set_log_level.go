package requests

import (
	"errors"
)

const (
	ErrorLogLevel = "ERROR"
	WarnLogLevel  = "WARN"
	InfoLogLevel  = "INFO"
	DebugLogLevel = "DEBUG"
	TraceLogLevel = "TRACE"
)

var ErrSetLogLevelInvalidLogLevel = errors.New("set-log-level: invalid log level")

type SetLogLevel struct {
	LogLevel string `json:"logLevel"`
}

func (c *SetLogLevel) Validate() error {
	switch c.LogLevel {
	case ErrorLogLevel, WarnLogLevel, InfoLogLevel, DebugLogLevel, TraceLogLevel:
		return nil
	}

	return ErrSetLogLevelInvalidLogLevel
}
