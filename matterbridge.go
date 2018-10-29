package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/gateway"
	"github.com/42wim/matterbridge/gateway/webhook"
	"github.com/google/gops/agent"
	"github.com/spf13/viper"
	prefixed "github.com/matterbridge/logrus-prefixed-formatter"
	log "github.com/sirupsen/logrus"
)

var (
	version = "1.11.4-dev"
	githash string
)

func main() {
	log.SetFormatter(&prefixed.TextFormatter{PrefixPadding: 13, DisableColors: true, FullTimestamp: true})
	flog := log.WithFields(log.Fields{"prefix": "main"})
	flagConfig := flag.String("conf", "matterbridge.toml", "config file")
	flagDebug := flag.Bool("debug", false, "enable debug")
	flagVersion := flag.Bool("version", false, "show version")
	flagGops := flag.Bool("gops", false, "enable gops agent")
	flagWebhook := flag.Bool("webhook", false, "run webhook serve mode")
	flag.Parse()
	if *flagGops {
		agent.Listen(&agent.Options{})
		defer agent.Close()
	}
	if *flagVersion {
		fmt.Printf("version: %s %s\n", version, githash)
		return
	}
	if *flagDebug || os.Getenv("DEBUG") == "1" {
		log.SetFormatter(&prefixed.TextFormatter{PrefixPadding: 13, DisableColors: true, FullTimestamp: false, ForceFormatting: true})
		flog.Info("Enabling debug")
		log.SetLevel(log.DebugLevel)
	}
	flog.Printf("Running version %s %s", version, githash)
	if strings.Contains(version, "-dev") {
		flog.Println("WARNING: THIS IS A DEVELOPMENT VERSION. Things may break.")
	}
	cfg := config.NewConfig(*flagConfig)
	cfg.General.Debug = *flagDebug
	if *flagWebhook {
		// TODO: Find out why this reverts after config load
		log.SetFormatter(&prefixed.TextFormatter{PrefixPadding: 13, DisableColors: true, FullTimestamp: true})
		flog.Println("Starting webhook for reloading remote config...")
		if viper.GetString("ConfigWebhookToken") == "" {
			flog.Fatalf("Must set config webhook's auth token to use.")
		}
		flog.Println("Serving at: POST /webhook")
		webhook.Serve()
	}
	r, err := gateway.NewRouter(cfg)
	if err != nil {
		flog.Fatalf("Starting gateway failed: %s", err)
	}
	err = r.Start()
	if err != nil {
		flog.Fatalf("Starting gateway failed: %s", err)
	}
	flog.Printf("Gateway(s) started succesfully. Now relaying messages")
	select {}
}
