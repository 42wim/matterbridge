package logutils

import (
	"sync"

	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/log"
)

var (
	_zapLogger     *zap.Logger
	_initZapLogger sync.Once
)

// ZapLogger creates a custom zap.Logger which will forward logs
// to status-go logger.
func ZapLogger() *zap.Logger {
	_initZapLogger.Do(func() {
		var err error
		_zapLogger, err = NewZapLoggerWithAdapter(log.Root())
		if err != nil {
			panic(err)
		}
	})
	return _zapLogger
}
