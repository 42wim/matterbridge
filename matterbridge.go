package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/gateway"
	"github.com/42wim/matterbridge/gateway/bridgemap"
	"github.com/google/gops/agent"
	prefixed "github.com/matterbridge/logrus-prefixed-formatter"
	"github.com/sirupsen/logrus"
)

var (
	version = "1.16.1"
	githash string

	flagConfig  = flag.String("conf", "matterbridge.toml", "config file")
	flagDebug   = flag.Bool("debug", false, "enable debug")
	flagVersion = flag.Bool("version", false, "show version")
	flagGops    = flag.Bool("gops", false, "enable gops agent")
)

func main() {
	flag.Parse()
	if *flagVersion {
		fmt.Printf("version: %s %s\n", version, githash)
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

	logger.Printf("Running version %s %s", version, githash)
	if strings.Contains(version, "-dev") {
		logger.Println("WARNING: THIS IS A DEVELOPMENT VERSION. Things may break.")
	}

	cfg := config.NewConfig(rootLogger, *flagConfig)
	cfg.BridgeValues().General.Debug = *flagDebug

	r, err := gateway.NewRouter(rootLogger, cfg, bridgemap.FullMap)
	if err != nil {
		logger.Fatalf("Starting gateway failed: %s", err)
	}
	if err = r.Start(); err != nil {
		logger.Fatalf("Starting gateway failed: %s", err)
	}
	logger.Printf("Gateway(s) started succesfully. Now relaying messages")
	select {}
}

func setupLogger() *logrus.Logger {
	logger := &logrus.Logger{
		Out: os.Stdout,
		Formatter: &prefixed.TextFormatter{
			PrefixPadding: 13,
			DisableColors: true,
			FullTimestamp: true,
		},
		Level: logrus.InfoLevel,
	}
	if *flagDebug || os.Getenv("DEBUG") == "1" {
		logger.Formatter = &prefixed.TextFormatter{
			PrefixPadding:   13,
			DisableColors:   true,
			FullTimestamp:   false,
			ForceFormatting: true,
		}
		logger.Level = logrus.DebugLevel
		logger.WithFields(logrus.Fields{"prefix": "main"}).Info("Enabling debug logging.")
	}
	return logger
}
