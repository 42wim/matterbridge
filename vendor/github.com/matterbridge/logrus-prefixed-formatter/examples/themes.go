package main

import (
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var log = logrus.New()

func init() {
	formatter := new(prefixed.TextFormatter)
	formatter.FullTimestamp = true

	// Set specific colors for prefix and timestamp
	formatter.SetColorScheme(&prefixed.ColorScheme{
		PrefixStyle:    "blue+b",
		TimestampStyle: "white+h",
	})

	log.Formatter = formatter
	log.Level = logrus.DebugLevel
}

func main() {
	log.WithFields(logrus.Fields{
		"prefix": "main",
		"animal": "walrus",
		"number": 8,
	}).Debug("Started observing beach")

	// Or you can simply add prefix in square brackets within message itself
	log.WithFields(logrus.Fields{
		"animal": "walrus",
		"size":   10,
	}).Debug("[main] A group of walrus emerges from the ocean")

	// Warning message
	log.WithFields(logrus.Fields{
		"omg":    true,
		"number": 122,
	}).Warn("[main] The group's number increased tremendously!")

	// Information message
	log.WithFields(logrus.Fields{
		"prefix":      "sensor",
		"temperature": -4,
	}).Info("Temperature changes")
}
