package upnp

import "github.com/anacrolix/log"

type levelLogger struct {
	log.Logger
}

func (me levelLogger) logf(level log.Level, format string, args ...interface{}) {
	log.Fmsg(format, args...).Skip(2).LogLevel(level, me.Logger)
}

func (me levelLogger) Infof(format string, args ...interface{}) {
	me.logf(log.Info, format, args...)
}

func (me levelLogger) Debugf(format string, args ...interface{}) {
	me.logf(log.Debug, format, args...)
}

func (me levelLogger) Errorf(format string, args ...interface{}) {
	me.logf(log.Error, format, args...)
}
