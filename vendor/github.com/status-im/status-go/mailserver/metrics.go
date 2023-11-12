package mailserver

import prom "github.com/prometheus/client_golang/prometheus"

// By default the /metrics endpoint is not available.
// It is exposed only if -metrics flag is set.

var (
	envelopesCounter = prom.NewCounter(prom.CounterOpts{
		Name: "mailserver_envelopes_total",
		Help: "Number of envelopes processed.",
	})
	deliveryFailuresCounter = prom.NewCounterVec(prom.CounterOpts{
		Name: "mailserver_delivery_failures_total",
		Help: "Number of requests that failed processing.",
	}, []string{"type"})
	deliveryAttemptsCounter = prom.NewCounter(prom.CounterOpts{
		Name: "mailserver_delivery_attempts_total",
		Help: "Number of Whisper envelopes processed.",
	})
	requestsBatchedCounter = prom.NewCounter(prom.CounterOpts{
		Name: "mailserver_requests_batched_total",
		Help: "Number of processed batched requests.",
	})
	requestsInBundlesDuration = prom.NewHistogram(prom.HistogramOpts{
		Name: "mailserver_requests_bundle_process_duration_seconds",
		Help: "The time it took to process message bundles.",
	})
	syncFailuresCounter = prom.NewCounterVec(prom.CounterOpts{
		Name: "mailserver_sync_failures_total",
		Help: "Number of failures processing a sync requests.",
	}, []string{"type"})
	syncAttemptsCounter = prom.NewCounter(prom.CounterOpts{
		Name: "mailserver_sync_attempts_total",
		Help: "Number of attempts are processing a sync requests.",
	})
	sendRawEnvelopeDuration = prom.NewHistogram(prom.HistogramOpts{
		Name: "mailserver_send_raw_envelope_duration_seconds",
		Help: "The time it took to send a Whisper envelope.",
	})
	sentEnvelopeBatchSizeMeter = prom.NewHistogram(prom.HistogramOpts{
		Name:    "mailserver_sent_envelope_batch_size_bytes",
		Help:    "Size of processed Whisper envelopes in bytes.",
		Buckets: prom.ExponentialBuckets(1024, 4, 10),
	})
	mailDeliveryDuration = prom.NewHistogram(prom.HistogramOpts{
		Name: "mailserver_delivery_duration_seconds",
		Help: "Time it takes to deliver messages to a Whisper peer.",
	})
	archivedErrorsCounter = prom.NewCounterVec(prom.CounterOpts{
		Name: "mailserver_archived_envelopes_failures_total",
		Help: "Number of failures storing a Whisper envelope.",
	}, []string{"db"})
	archivedEnvelopesGauge = prom.NewGaugeVec(prom.GaugeOpts{
		Name: "mailserver_archived_envelopes_total",
		Help: "Number of envelopes saved in the DB.",
	}, []string{"db"})
	archivedEnvelopeSizeMeter = prom.NewHistogramVec(prom.HistogramOpts{
		Name:    "mailserver_archived_envelope_size_bytes",
		Help:    "Size of envelopes saved.",
		Buckets: prom.ExponentialBuckets(1024, 2, 11),
	}, []string{"db"})
	envelopeQueriesCounter = prom.NewCounterVec(prom.CounterOpts{
		Name: "mailserver_envelope_queries_total",
		Help: "Number of queries for envelopes in the DB.",
	}, []string{"filter", "history"})
)

func init() {
	prom.MustRegister(envelopesCounter)
	prom.MustRegister(deliveryFailuresCounter)
	prom.MustRegister(deliveryAttemptsCounter)
	prom.MustRegister(requestsBatchedCounter)
	prom.MustRegister(requestsInBundlesDuration)
	prom.MustRegister(syncFailuresCounter)
	prom.MustRegister(syncAttemptsCounter)
	prom.MustRegister(sendRawEnvelopeDuration)
	prom.MustRegister(sentEnvelopeBatchSizeMeter)
	prom.MustRegister(mailDeliveryDuration)
	prom.MustRegister(archivedErrorsCounter)
	prom.MustRegister(archivedEnvelopesGauge)
	prom.MustRegister(archivedEnvelopeSizeMeter)
	prom.MustRegister(envelopeQueriesCounter)
}
