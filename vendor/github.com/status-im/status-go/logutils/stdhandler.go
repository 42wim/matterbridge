package logutils

import (
	stdlog "log"

	"github.com/ethereum/go-ethereum/log"
)

// NewStdHandler returns handler that uses logger from golang std lib.
func NewStdHandler(fmtr log.Format) log.Handler {
	return log.FuncHandler(func(r *log.Record) error {
		line := fmtr.Format(r)
		// 8 is a number of frames that will be skipped when log is printed.
		// this is needed to show the file (with line number) where call to a logger was made
		return stdlog.Output(8, string(line))
	})
}
