package metricshelper

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

// RegisterCollectors registers the collectors with reg ignoring
// reregistration error and panics on any other error
func RegisterCollectors(reg prometheus.Registerer, collectors ...prometheus.Collector) {
	for _, c := range collectors {
		err := reg.Register(c)
		if err != nil {
			if ok := errors.As(err, &prometheus.AlreadyRegisteredError{}); !ok {
				panic(err)
			}
		}
	}
}
