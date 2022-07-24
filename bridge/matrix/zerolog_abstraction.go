package bmatrix

import (
	"errors"

	"github.com/sirupsen/logrus"

	"github.com/rs/zerolog"
)

var levels_zerolog2logrus = map[zerolog.Level]logrus.Level{
	zerolog.DebugLevel: logrus.DebugLevel,
	zerolog.InfoLevel:  logrus.InfoLevel,
	zerolog.WarnLevel:  logrus.WarnLevel,
	zerolog.FatalLevel: logrus.FatalLevel,
	zerolog.PanicLevel: logrus.PanicLevel,
	zerolog.ErrorLevel: logrus.ErrorLevel,
	zerolog.TraceLevel: logrus.TraceLevel,
}

// an abstraction for zerolog so we can pipe its output to logrus.Entry, that is used in matterbridge
type zerologWrapper struct {
	inner *logrus.Entry
}

func (w zerologWrapper) Write(p []byte) (n int, err error) {
	return w.inner.Logger.Writer().Write(p)
}

func (w zerologWrapper) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	if logrus_level, present := levels_zerolog2logrus[level]; present {
		return w.inner.Logger.WriterLevel(logrus_level).Write(p)
	}
	// drop the message if we haven't a matching level
	return 0, errors.New("Unsupported logging level")
}

func NewZerologWrapper(entry *logrus.Entry) zerolog.Logger {
	wrapper := zerologWrapper{inner: entry}
	log := zerolog.New(wrapper)

	return log
}
