package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/gateway"
	"github.com/42wim/matterbridge/gateway/bridgemap"
	"github.com/42wim/matterbridge/version"
	"github.com/google/gops/agent"
	prefixed "github.com/matterbridge/logrus-prefixed-formatter"
	"github.com/sirupsen/logrus"
)

var (
	flagConfig  = flag.String("conf", "matterbridge.toml", "config file")
	flagDebug   = flag.Bool("debug", false, "enable debug")
	flagVersion = flag.Bool("version", false, "show version")
	flagGops    = flag.Bool("gops", false, "enable gops agent")
)

func main() {
	flag.Parse()
	if *flagVersion {
		fmt.Printf("version: %s %s\n", version.Release, version.GitHash)
		return
	}

	rootLogger := setupLogger()
	logger := rootLogger.WithFields(logrus.Fields{"prefix": "main"})

	if *flagGops {
		if err := agent.Listen(agent.Options{}); err != nil {
			logger.Errorf("Failed to start gops agent: %#v", err)
		} else {
			defer agent.Close()
		}
	}

	logger.Printf("Running version %s %s", version.Release, version.GitHash)
	if strings.Contains(version.Release, "-dev") {
		logger.Println("WARNING: THIS IS A DEVELOPMENT VERSION. Things may break.")
	}

	cfg := config.NewConfig(rootLogger, *flagConfig)
	cfg.BridgeValues().General.Debug = *flagDebug

	// if logging to a file, ensure it is closed when the program terminates
	// nolint:errcheck
	defer func() {
		if f, ok := rootLogger.Out.(*os.File); ok {
			f.Sync()
			f.Close()
		}
	}()

	r, err := gateway.NewRouter(rootLogger, cfg, bridgemap.FullMap)
	if err != nil {
		logger.Fatalf("Starting gateway failed: %s", err)
	}
	if err = r.Start(); err != nil {
		logger.Fatalf("Starting gateway failed: %s", err)
	}
	logger.Printf("Gateway(s) started successfully. Now relaying messages")
	select {}
}

func setupLogger() *logrus.Logger {
	logger := &logrus.Logger{
		Out: os.Stdout,
		Formatter: &prefixed.TextFormatter{
			PrefixPadding: 13,
			DisableColors: true,
		},
		Level: logrus.InfoLevel,
	}
	if *flagDebug || os.Getenv("DEBUG") == "1" {
		logger.SetReportCaller(true)
		logger.Formatter = &prefixed.TextFormatter{
			PrefixPadding: 13,
			DisableColors: true,
			FullTimestamp: false,

			CallerFormatter: func(function, file string) string {
				return fmt.Sprintf(" [%s:%s]", function, file)
			},
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				sp := strings.SplitAfter(f.File, "/matterbridge/")
				filename := f.File
				if len(sp) > 1 {
					filename = sp[1]
				}
				s := strings.Split(f.Function, ".")
				funcName := s[len(s)-1]
				return funcName, fmt.Sprintf("%s:%d", filename, f.Line)
			},
		}

		logger.Level = logrus.DebugLevel
		logger.WithFields(logrus.Fields{"prefix": "main"}).Info("Enabling debug logging.")
	}
	return logger
}
