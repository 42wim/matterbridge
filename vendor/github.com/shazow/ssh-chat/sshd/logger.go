package sshd

import "io"
import stdlog "log"

var logger *stdlog.Logger

func SetLogger(w io.Writer) {
	flags := stdlog.Flags()
	prefix := "[sshd] "
	logger = stdlog.New(w, prefix, flags)
}

type nullWriter struct{}

func (nullWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func init() {
	SetLogger(nullWriter{})
}
