package eventbus

import (
	"reflect"
	"strings"

	"github.com/libp2p/go-libp2p/p2p/metricshelper"

	"github.com/prometheus/client_golang/prometheus"
)

const metricNamespace = "libp2p_eventbus"

var (
	eventsEmitted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "events_emitted_total",
			Help:      "Events Emitted",
		},
		[]string{"event"},
	)
	totalSubscribers = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "subscribers_total",
			Help:      "Number of subscribers for an event type",
		},
		[]string{"event"},
	)
	subscriberQueueLength = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "subscriber_queue_length",
			Help:      "Subscriber queue length",
		},
		[]string{"subscriber_name"},
	)
	subscriberQueueFull = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "subscriber_queue_full",
			Help:      "Subscriber Queue completely full",
		},
		[]string{"subscriber_name"},
	)
	subscriberEventQueued = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "subscriber_event_queued",
			Help:      "Event Queued for subscriber",
		},
		[]string{"subscriber_name"},
	)
	collectors = []prometheus.Collector{
		eventsEmitted,
		totalSubscribers,
		subscriberQueueLength,
		subscriberQueueFull,
		subscriberEventQueued,
	}
)

// MetricsTracer tracks metrics for the eventbus subsystem
type MetricsTracer interface {

	// EventEmitted counts the total number of events grouped by event type
	EventEmitted(typ reflect.Type)

	// AddSubscriber adds a subscriber for the event type
	AddSubscriber(typ reflect.Type)

	// RemoveSubscriber removes a subscriber for the event type
	RemoveSubscriber(typ reflect.Type)

	// SubscriberQueueLength is the length of the subscribers channel
	SubscriberQueueLength(name string, n int)

	// SubscriberQueueFull tracks whether a subscribers channel if full
	SubscriberQueueFull(name string, isFull bool)

	// SubscriberEventQueued counts the total number of events grouped by subscriber
	SubscriberEventQueued(name string)
}

type metricsTracer struct{}

var _ MetricsTracer = &metricsTracer{}

type metricsTracerSetting struct {
	reg prometheus.Registerer
}

type MetricsTracerOption func(*metricsTracerSetting)

func WithRegisterer(reg prometheus.Registerer) MetricsTracerOption {
	return func(s *metricsTracerSetting) {
		if reg != nil {
			s.reg = reg
		}
	}
}

func NewMetricsTracer(opts ...MetricsTracerOption) MetricsTracer {
	setting := &metricsTracerSetting{reg: prometheus.DefaultRegisterer}
	for _, opt := range opts {
		opt(setting)
	}
	metricshelper.RegisterCollectors(setting.reg, collectors...)
	return &metricsTracer{}
}

func (m *metricsTracer) EventEmitted(typ reflect.Type) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	*tags = append(*tags, strings.TrimPrefix(typ.String(), "event."))
	eventsEmitted.WithLabelValues(*tags...).Inc()
}

func (m *metricsTracer) AddSubscriber(typ reflect.Type) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	*tags = append(*tags, strings.TrimPrefix(typ.String(), "event."))
	totalSubscribers.WithLabelValues(*tags...).Inc()
}

func (m *metricsTracer) RemoveSubscriber(typ reflect.Type) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	*tags = append(*tags, strings.TrimPrefix(typ.String(), "event."))
	totalSubscribers.WithLabelValues(*tags...).Dec()
}

func (m *metricsTracer) SubscriberQueueLength(name string, n int) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	*tags = append(*tags, name)
	subscriberQueueLength.WithLabelValues(*tags...).Set(float64(n))
}

func (m *metricsTracer) SubscriberQueueFull(name string, isFull bool) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	*tags = append(*tags, name)
	observer := subscriberQueueFull.WithLabelValues(*tags...)
	if isFull {
		observer.Set(1)
	} else {
		observer.Set(0)
	}
}

func (m *metricsTracer) SubscriberEventQueued(name string) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	*tags = append(*tags, name)
	subscriberEventQueued.WithLabelValues(*tags...).Inc()
}
