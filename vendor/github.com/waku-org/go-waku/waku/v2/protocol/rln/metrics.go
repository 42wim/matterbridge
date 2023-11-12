package rln

import (
	"time"

	"github.com/libp2p/go-libp2p/p2p/metricshelper"
	"github.com/prometheus/client_golang/prometheus"
)

var messagesTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "waku_rln_messages_total",
		Help: "number of messages received in the RLN validator",
	})

var spamMessagesTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "waku_rln_spam_messages_total",
		Help: "number of spam messages detected",
	},
	[]string{"contentTopic"})

var invalidMessagesTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "waku_rln_invalid_messages_total",
		Help: "number of invalid messages detected",
	},
	[]string{"type"})

func generateBucketsForHistogram(length int) []float64 {
	// Generate a custom set of 5 buckets for a given length
	numberOfBuckets := 5
	stepSize := length / numberOfBuckets
	var buckets []float64
	for i := 1; i <= 5; i++ {
		buckets = append(buckets, float64(stepSize*i))
	}
	return buckets
}

// This metric will be useful in detecting the index of the root in the acceptable window of roots
var validMessagesTotal = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "waku_rln_valid_messages_total",
	Help:    "number of valid messages with their roots tracked",
	Buckets: generateBucketsForHistogram(acceptableRootWindowSize),
})

var errorsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "waku_rln_errors_total",
		Help: "number of errors detected while operating the rln relay",
	},
	[]string{"type"},
)

var proofVerificationTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "waku_rln_proof_verification_total",
		Help: "number of times the rln proofs are verified",
	})

var proofVerificationDurationSeconds = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name: "waku_rln_proof_verification_duration_seconds",
		Help: "time taken to verify a proof",
	})

var proofGenerationDurationSeconds = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name: "waku_rln_proof_generation_duration_seconds",
		Help: "time taken to generate a proof",
	})

var instanceCreationDurationSeconds = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "waku_rln_instance_creation_duration_seconds",
		Help: "time taken to create an rln instance",
	})

var collectors = []prometheus.Collector{
	messagesTotal,
	spamMessagesTotal,
	invalidMessagesTotal,
	errorsTotal,
	validMessagesTotal,
	proofVerificationTotal,
	proofVerificationDurationSeconds,
	proofGenerationDurationSeconds,
	instanceCreationDurationSeconds,
}

type errCategory string

var (
	proofMetadataExtractionErr errCategory = "proof_metadata_extraction"
	proofVerificationErr       errCategory = "proof_verification"
	duplicateCheckErr          errCategory = "duplicate_check"
	logInsertionErr            errCategory = "log_insertion"
)

type invalidCategory string

var (
	invalidNoProof     invalidCategory = "no_proof"
	invalidEpoch       invalidCategory = "invalid_epoch"
	invalidRoot        invalidCategory = "invalid_root"
	invalidProof       invalidCategory = "invalid_proof"
	proofExtractionErr invalidCategory = "invalid_proof_extract_err"
)

// Metrics exposes the functions required to update prometheus metrics for lightpush protocol
type Metrics interface {
	RecordMessage()
	RecordSpam(contentTopic string)
	RecordInvalidMessage(cause invalidCategory)
	RecordError(err errCategory)
	RecordProofVerification(duration time.Duration)
	RecordProofGeneration(duration time.Duration)
	RecordValidMessages(rootIndex int)
	RecordInstanceCreation(duration time.Duration)
}

type metricsImpl struct {
	reg prometheus.Registerer
}

func newMetrics(reg prometheus.Registerer) Metrics {
	metricshelper.RegisterCollectors(reg, collectors...)
	return &metricsImpl{
		reg: reg,
	}
}

// RecordMessage is used to increase the counter for the number of messages received in the RLN validator
func (m *metricsImpl) RecordMessage() {
	messagesTotal.Inc()
}

// RecordSpam is used to increase the counter for the number of spam of messages received
func (m *metricsImpl) RecordSpam(contentTopic string) {
	spamMessagesTotal.WithLabelValues(contentTopic).Inc()
}

// RecordError increases the counter for different error types
func (m *metricsImpl) RecordError(err errCategory) {
	errorsTotal.WithLabelValues(string(err)).Inc()
}

// RecordProofVerification increases the counter for the number of proof verifications perfomed, and the duration of these
func (m *metricsImpl) RecordProofVerification(duration time.Duration) {
	proofVerificationTotal.Inc()
	proofVerificationDurationSeconds.Observe(duration.Seconds())
}

// RecordProofGeneration measures the duration to generate a proof
func (m *metricsImpl) RecordProofGeneration(duration time.Duration) {
	proofGenerationDurationSeconds.Observe(duration.Seconds())
}

// RecordInstanceCreation records how long did it take to instantiate RLN
func (m *metricsImpl) RecordInstanceCreation(duration time.Duration) {
	instanceCreationDurationSeconds.Set(duration.Seconds())
}

// RecordInvalidMessage increases the counter for different types of invalid messages
func (m *metricsImpl) RecordInvalidMessage(cause invalidCategory) {
	invalidMessagesTotal.WithLabelValues(string(cause)).Inc()
}

// RecordValidMessages records the root index used for valid messages
func (m *metricsImpl) RecordValidMessages(rootIndex int) {
	validMessagesTotal.Observe(float64(rootIndex))
}
