package utils

import (
	"strings"

	logging "github.com/ipfs/go-log/v2"

	"go.uber.org/zap"
)

var log *zap.Logger
var messageLoggers map[string]*zap.Logger

// Logger creates a zap.Logger with some reasonable defaults
func Logger() *zap.Logger {
	if log == nil {
		InitLogger("console", "stdout")
	}
	return log
}

// MessagesLogger returns a logger used for debug logging of sent/received messages
func MessagesLogger(prefix string) *zap.Logger {
	if messageLoggers == nil {
		messageLoggers = make(map[string]*zap.Logger)
	}
	logger := messageLoggers[prefix]
	if logger == nil {
		logger = logging.Logger(prefix + ".messages").Desugar()
		messageLoggers[prefix] = logger
	}

	return logger
}

// InitLogger initializes a global logger using an specific encoding
func InitLogger(encoding string, output string) {
	cfg := logging.GetConfig()

	if encoding == "json" {
		cfg.Format = logging.JSONOutput
	} else if encoding == "nocolor" {
		cfg.Format = logging.PlaintextOutput
	} else {
		cfg.Format = logging.ColorizedOutput
	}

	if output == "stdout" || output == "" {
		cfg.Stdout = true
		cfg.Stderr = false
	} else {
		if encoding == "console" {
			cfg.Format = logging.PlaintextOutput
		}
		cfg.Stdout = false
		cfg.Stderr = false

		outputParts := strings.Split(output, ":")
		if len(outputParts) == 2 {
			cfg.File = outputParts[1]
		} else {
			if len(outputParts) > 2 || outputParts[0] != "file" {
				panic("invalid output format")
			}
			cfg.File = "./waku.log"
		}
	}
	if cfg.Level == logging.LevelError {
		// Override default level setting
		cfg.Level = logging.LevelInfo
	}

	logging.SetupLogging(cfg)

	log = logging.Logger("gowaku").Desugar()
}
