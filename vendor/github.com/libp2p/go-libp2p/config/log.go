package config

import (
	"strings"
	"sync"

	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx/fxevent"
)

var log = logging.Logger("p2p-config")

var (
	fxLogger    fxevent.Logger
	logInitOnce sync.Once
)

type fxLogWriter struct{}

func (l *fxLogWriter) Write(b []byte) (int, error) {
	log.Debug(strings.TrimSuffix(string(b), "\n"))
	return len(b), nil
}

func getFXLogger() fxevent.Logger {
	logInitOnce.Do(func() { fxLogger = &fxevent.ConsoleLogger{W: &fxLogWriter{}} })
	return fxLogger
}
