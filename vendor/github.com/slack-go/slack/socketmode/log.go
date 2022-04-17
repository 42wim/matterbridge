package socketmode

import "fmt"

// TODO merge logger, ilogger, and internalLogger with the top-level package's equivalents

// logger is a logger interface compatible with both stdlib and some
// 3rd party loggers.
type logger interface {
	Output(int, string) error
}

// ilogger represents the internal logging api we use.
type ilogger interface {
	logger
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})
}

// internalLog implements the additional methods used by our internal logging.
type internalLog struct {
	logger
}

// Println replicates the behaviour of the standard logger.
func (t internalLog) Println(v ...interface{}) {
	t.Output(2, fmt.Sprintln(v...))
}

// Printf replicates the behaviour of the standard logger.
func (t internalLog) Printf(format string, v ...interface{}) {
	t.Output(2, fmt.Sprintf(format, v...))
}

// Print replicates the behaviour of the standard logger.
func (t internalLog) Print(v ...interface{}) {
	t.Output(2, fmt.Sprint(v...))
}

func (smc *Client) Debugf(format string, v ...interface{}) {
	if smc.debug {
		smc.log.Output(2, fmt.Sprintf(format, v...))
	}
}

func (smc *Client) Debugln(v ...interface{}) {
	if smc.debug {
		smc.log.Output(2, fmt.Sprintln(v...))
	}
}
