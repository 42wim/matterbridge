package eventbus

import (
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"
)

type subSettings struct {
	buffer int
	name   string
}

var subCnt atomic.Int64

var subSettingsDefault = subSettings{
	buffer: 16,
}

// newSubSettings returns the settings for a new subscriber
// The default naming strategy is sub-<fileName>-L<lineNum>
func newSubSettings() subSettings {
	settings := subSettingsDefault
	_, file, line, ok := runtime.Caller(2) // skip=1 is eventbus.Subscriber
	if ok {
		file = strings.TrimPrefix(file, "github.com/")
		// remove the version number from the path, for example
		// go-libp2p-package@v0.x.y-some-hash-123/file.go will be shortened go go-libp2p-package/file.go
		if idx1 := strings.Index(file, "@"); idx1 != -1 {
			if idx2 := strings.Index(file[idx1:], "/"); idx2 != -1 {
				file = file[:idx1] + file[idx1+idx2:]
			}
		}
		settings.name = fmt.Sprintf("%s-L%d", file, line)
	} else {
		settings.name = fmt.Sprintf("subscriber-%d", subCnt.Add(1))
	}
	return settings
}

func BufSize(n int) func(interface{}) error {
	return func(s interface{}) error {
		s.(*subSettings).buffer = n
		return nil
	}
}

func Name(name string) func(interface{}) error {
	return func(s interface{}) error {
		s.(*subSettings).name = name
		return nil
	}
}

type emitterSettings struct {
	makeStateful bool
}

// Stateful is an Emitter option which makes the eventbus channel
// 'remember' last event sent, and when a new subscriber joins the
// bus, the remembered event is immediately sent to the subscription
// channel.
//
// This allows to provide state tracking for dynamic systems, and/or
// allows new subscribers to verify that there are Emitters on the channel
func Stateful(s interface{}) error {
	s.(*emitterSettings).makeStateful = true
	return nil
}

type Option func(*basicBus)

func WithMetricsTracer(metricsTracer MetricsTracer) Option {
	return func(bus *basicBus) {
		bus.metricsTracer = metricsTracer
		bus.wildcard.metricsTracer = metricsTracer
	}
}
