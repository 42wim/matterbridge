package tt

import (
	"sync"

	"github.com/status-im/status-go/protocol/zaputil"

	"go.uber.org/zap"
)

var registerOnce sync.Once

// MustCreateTestLogger returns a logger based on the passed flags.
func MustCreateTestLogger() *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	return MustCreateTestLoggerWithConfig(cfg)
}

func MustCreateTestLoggerWithConfig(cfg zap.Config) *zap.Logger {
	registerOnce.Do(func() {
		if err := zaputil.RegisterConsoleHexEncoder(); err != nil {
			panic(err)
		}
	})
	cfg.Encoding = "console-hex"
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return l
}
