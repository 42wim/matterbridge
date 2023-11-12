package relay

import (
	"github.com/libp2p/go-libp2p/p2p/metricshelper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/waku-org/go-waku/logging"
	waku_proto "github.com/waku-org/go-waku/waku/v2/protocol"
	"go.uber.org/zap"
)

var messages = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "waku_node_messages",
		Help: "The number of the messages received",
	},
	[]string{"pubsubTopic"},
)

var messageSize = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "waku_histogram_message_size",
	Help:    "message size histogram in kB",
	Buckets: []float64{0.0, 5.0, 15.0, 50.0, 100.0, 300.0, 700.0, 1000.0},
})

var pubsubTopics = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "waku_pubsub_topics",
		Help: "Number of PubSub Topics node is subscribed to",
	})

var collectors = []prometheus.Collector{
	messages,
	messageSize,
	pubsubTopics,
}

// Metrics exposes the functions required to update prometheus metrics for relay protocol
type Metrics interface {
	RecordMessage(envelope *waku_proto.Envelope)
	SetPubSubTopics(int)
}

type metricsImpl struct {
	log *zap.Logger
	reg prometheus.Registerer
}

func newMetrics(reg prometheus.Registerer, logger *zap.Logger) Metrics {
	metricshelper.RegisterCollectors(reg, collectors...)
	return &metricsImpl{
		log: logger,
		reg: reg,
	}
}

// RecordMessage is used to increase the counter for the number of messages received via waku relay
func (m *metricsImpl) RecordMessage(envelope *waku_proto.Envelope) {
	go func() {
		payloadSizeInBytes := len(envelope.Message().Payload)
		payloadSizeInKb := float64(payloadSizeInBytes) / 1000
		messageSize.Observe(payloadSizeInKb)
		pubsubTopic := envelope.PubsubTopic()
		messages.WithLabelValues(pubsubTopic).Inc()
		m.log.Debug("waku.relay received", zap.String("pubsubTopic", pubsubTopic), logging.HexBytes("hash", envelope.Hash()), zap.Int64("receivedTime", envelope.Index().ReceiverTime), zap.Int("payloadSizeBytes", payloadSizeInBytes))
	}()
}

func (m *metricsImpl) SetPubSubTopics(size int) {
	pubsubTopics.Set(float64(size))
}
