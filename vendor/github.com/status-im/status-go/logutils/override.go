package logutils

import (
	"os"
	"strings"

	logging "github.com/ipfs/go-log/v2"

	"github.com/ethereum/go-ethereum/log"
)

type LogSettings struct {
	Enabled         bool
	MobileSystem    bool
	Level           string
	File            string
	MaxSize         int
	MaxBackups      int
	CompressRotated bool
}

// OverrideWithStdLogger overwrites ethereum's root logger with a logger from golang std lib.
func OverrideWithStdLogger(logLevel string) error {
	return enableRootLog(logLevel, NewStdHandler(log.TerminalFormat(false)))
}

// OverrideRootLogWithConfig derives all configuration from params.NodeConfig and configures logger using it.
func OverrideRootLogWithConfig(settings LogSettings, colors bool) error {
	if !settings.Enabled {
		return nil
	}
	if settings.MobileSystem {
		return OverrideWithStdLogger(settings.Level)
	}
	return OverrideRootLog(settings.Enabled, settings.Level, FileOptions{
		Filename:   settings.File,
		MaxSize:    settings.MaxSize,
		MaxBackups: settings.MaxBackups,
		Compress:   settings.CompressRotated,
	}, colors)

}

// OverrideRootLog overrides root logger with file handler, if defined,
// and log level (defaults to INFO).
func OverrideRootLog(enabled bool, levelStr string, fileOpts FileOptions, terminal bool) error {
	if !enabled {
		disableRootLog()
		return nil
	}
	if os.Getenv("CI") == "true" {
		terminal = false
	}
	var (
		handler log.Handler
	)
	if fileOpts.Filename != "" {
		if fileOpts.MaxBackups == 0 {
			// Setting MaxBackups to 0 causes all log files to be kept. Even setting MaxAge to > 0 doesn't fix it
			// Docs: https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2@v2.0.0#readme-cleaning-up-old-log-files
			fileOpts.MaxBackups = 1
		}
		handler = FileHandlerWithRotation(fileOpts, log.TerminalFormat(terminal))
	} else {
		handler = log.StreamHandler(os.Stderr, log.TerminalFormat(terminal))
	}

	return enableRootLog(levelStr, handler)
}

func disableRootLog() {
	log.Root().SetHandler(log.DiscardHandler())
}

func enableRootLog(levelStr string, handler log.Handler) error {
	if levelStr == "" {
		levelStr = "INFO"
	}

	levelStr = strings.ToLower(levelStr)

	level, err := log.LvlFromString(levelStr)
	if err != nil {
		return err
	}

	filteredHandler := log.LvlFilterHandler(level, handler)
	log.Root().SetHandler(filteredHandler)
	log.PrintOrigins(true)

	// go-libp2p logger
	lvl, err := logging.LevelFromString(levelStr)
	if err != nil {
		return err
	}

	logging.SetAllLoggers(lvl)

	return nil
}
