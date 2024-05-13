//go:build test_silent

package tt

import "go.uber.org/zap"

func loggerConfig() zap.Config {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	return config
}
