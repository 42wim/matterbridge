//go:build !test_silent

package tt

import "go.uber.org/zap"

func loggerConfig() zap.Config {
	return zap.NewDevelopmentConfig()
}
