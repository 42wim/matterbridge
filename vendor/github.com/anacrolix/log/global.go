package log

import (
	"fmt"
	"io/ioutil"
	"os"
)

var (
	DefaultHandler = StreamHandler{
		W:   os.Stderr,
		Fmt: LineFormatter,
	}
	Default = Logger{
		nonZero:     true,
		filterLevel: Error,
		Handlers:    []Handler{DefaultHandler},
	}
	DiscardHandler = StreamHandler{
		W:   ioutil.Discard,
		Fmt: func(Record) []byte { return nil },
	}
)

func Levelf(level Level, format string, a ...interface{}) {
	Default.LazyLog(level, func() Msg {
		return Fmsg(format, a...).Skip(1)
	})
}

func Printf(format string, a ...interface{}) {
	Default.Log(Fmsg(format, a...).Skip(1))
}

// Prints the arguments to the Default Logger.
func Print(a ...interface{}) {
	// TODO: There's no "Print" equivalent constructor for a Msg, and I don't know what I'd call it.
	Str(fmt.Sprint(a...)).Skip(1).Log(Default)
}
