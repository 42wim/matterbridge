package autorelay

import (
	"errors"

	"github.com/libp2p/go-libp2p/p2p/metricshelper"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	pbv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/pb"
	"github.com/prometheus/client_golang/prometheus"
)

const metricNamespace = "libp2p_autorelay"

var (
	status = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: metricNamespace,
		Name:      "status",
		Help:      "relay finder active",
	})
	reservationsOpenedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "reservations_opened_total",
			Help:      "Reservations Opened",
		},
	)
	reservationsClosedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "reservations_closed_total",
			Help:      "Reservations Closed",
		},
	)
	reservationRequestsOutcomeTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "reservation_requests_outcome_total",
			Help:      "Reservation Request Outcome",
		},
		[]string{"request_type", "outcome"},
	)

	relayAddressesUpdatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "relay_addresses_updated_total",
			Help:      "Relay Addresses Updated Count",
		},
	)
	relayAddressesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "relay_addresses_count",
			Help:      "Relay Addresses Count",
		},
	)

	candidatesCircuitV2SupportTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "candidates_circuit_v2_support_total",
			Help:      "Candidiates supporting circuit v2",
		},
		[]string{"support"},
	)
	candidatesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "candidates_total",
			Help:      "Candidates Total",
		},
		[]string{"type"},
	)
	candLoopState = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "candidate_loop_state",
			Help:      "Candidate Loop State",
		},
	)

	scheduledWorkTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "scheduled_work_time",
			Help:      "Scheduled Work Times",
		},
		[]string{"work_type"},
	)

	desiredReservations = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "desired_reservations",
			Help:      "Desired Reservations",
		},
	)

	collectors = []prometheus.Collector{
		status,
		reservationsOpenedTotal,
		reservationsClosedTotal,
		reservationRequestsOutcomeTotal,
		relayAddressesUpdatedTotal,
		relayAddressesCount,
		candidatesCircuitV2SupportTotal,
		candidatesTotal,
		candLoopState,
		scheduledWorkTime,
		desiredReservations,
	}
)

type candidateLoopState int

const (
	peerSourceRateLimited candidateLoopState = iota
	waitingOnPeerChan
	waitingForTrigger
	stopped
)

// MetricsTracer is the interface for tracking metrics for autorelay
type MetricsTracer interface {
	RelayFinderStatus(isActive bool)

	ReservationEnded(cnt int)
	ReservationOpened(cnt int)
	ReservationRequestFinished(isRefresh bool, err error)

	RelayAddressCount(int)
	RelayAddressUpdated()

	CandidateChecked(supportsCircuitV2 bool)
	CandidateAdded(cnt int)
	CandidateRemoved(cnt int)
	CandidateLoopState(state candidateLoopState)

	ScheduledWorkUpdated(scheduledWork *scheduledWorkTimes)

	DesiredReservations(int)
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

	// Initialise these counters to 0 otherwise the first reservation requests aren't handled
	// correctly when using promql increse function
	reservationRequestsOutcomeTotal.WithLabelValues("refresh", "success")
	reservationRequestsOutcomeTotal.WithLabelValues("new", "success")
	candidatesCircuitV2SupportTotal.WithLabelValues("yes")
	candidatesCircuitV2SupportTotal.WithLabelValues("no")
	return &metricsTracer{}
}

func (mt *metricsTracer) RelayFinderStatus(isActive bool) {
	if isActive {
		status.Set(1)
	} else {
		status.Set(0)
	}
}

func (mt *metricsTracer) ReservationEnded(cnt int) {
	reservationsClosedTotal.Add(float64(cnt))
}

func (mt *metricsTracer) ReservationOpened(cnt int) {
	reservationsOpenedTotal.Add(float64(cnt))
}

func (mt *metricsTracer) ReservationRequestFinished(isRefresh bool, err error) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	if isRefresh {
		*tags = append(*tags, "refresh")
	} else {
		*tags = append(*tags, "new")
	}
	*tags = append(*tags, getReservationRequestStatus(err))
	reservationRequestsOutcomeTotal.WithLabelValues(*tags...).Inc()

	if !isRefresh && err == nil {
		reservationsOpenedTotal.Inc()
	}
}

func (mt *metricsTracer) RelayAddressUpdated() {
	relayAddressesUpdatedTotal.Inc()
}

func (mt *metricsTracer) RelayAddressCount(cnt int) {
	relayAddressesCount.Set(float64(cnt))
}

func (mt *metricsTracer) CandidateChecked(supportsCircuitV2 bool) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)
	if supportsCircuitV2 {
		*tags = append(*tags, "yes")
	} else {
		*tags = append(*tags, "no")
	}
	candidatesCircuitV2SupportTotal.WithLabelValues(*tags...).Inc()
}

func (mt *metricsTracer) CandidateAdded(cnt int) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)
	*tags = append(*tags, "added")
	candidatesTotal.WithLabelValues(*tags...).Add(float64(cnt))
}

func (mt *metricsTracer) CandidateRemoved(cnt int) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)
	*tags = append(*tags, "removed")
	candidatesTotal.WithLabelValues(*tags...).Add(float64(cnt))
}

func (mt *metricsTracer) CandidateLoopState(state candidateLoopState) {
	candLoopState.Set(float64(state))
}

func (mt *metricsTracer) ScheduledWorkUpdated(scheduledWork *scheduledWorkTimes) {
	tags := metricshelper.GetStringSlice()
	defer metricshelper.PutStringSlice(tags)

	*tags = append(*tags, "allowed peer source call")
	scheduledWorkTime.WithLabelValues(*tags...).Set(float64(scheduledWork.nextAllowedCallToPeerSource.Unix()))
	*tags = (*tags)[:0]

	*tags = append(*tags, "reservation refresh")
	scheduledWorkTime.WithLabelValues(*tags...).Set(float64(scheduledWork.nextRefresh.Unix()))
	*tags = (*tags)[:0]

	*tags = append(*tags, "clear backoff")
	scheduledWorkTime.WithLabelValues(*tags...).Set(float64(scheduledWork.nextBackoff.Unix()))
	*tags = (*tags)[:0]

	*tags = append(*tags, "old candidate check")
	scheduledWorkTime.WithLabelValues(*tags...).Set(float64(scheduledWork.nextOldCandidateCheck.Unix()))
}

func (mt *metricsTracer) DesiredReservations(cnt int) {
	desiredReservations.Set(float64(cnt))
}

func getReservationRequestStatus(err error) string {
	if err == nil {
		return "success"
	}

	status := "err other"
	var re client.ReservationError
	if errors.As(err, &re) {
		switch re.Status {
		case pbv2.Status_CONNECTION_FAILED:
			return "connection failed"
		case pbv2.Status_MALFORMED_MESSAGE:
			return "malformed message"
		case pbv2.Status_RESERVATION_REFUSED:
			return "reservation refused"
		case pbv2.Status_PERMISSION_DENIED:
			return "permission denied"
		case pbv2.Status_RESOURCE_LIMIT_EXCEEDED:
			return "resource limit exceeded"
		}
	}
	return status
}

// wrappedMetricsTracer wraps MetricsTracer and ignores all calls when mt is nil
type wrappedMetricsTracer struct {
	mt MetricsTracer
}

var _ MetricsTracer = &wrappedMetricsTracer{}

func (mt *wrappedMetricsTracer) RelayFinderStatus(isActive bool) {
	if mt.mt != nil {
		mt.mt.RelayFinderStatus(isActive)
	}
}

func (mt *wrappedMetricsTracer) ReservationEnded(cnt int) {
	if mt.mt != nil {
		mt.mt.ReservationEnded(cnt)
	}
}

func (mt *wrappedMetricsTracer) ReservationOpened(cnt int) {
	if mt.mt != nil {
		mt.mt.ReservationOpened(cnt)
	}
}

func (mt *wrappedMetricsTracer) ReservationRequestFinished(isRefresh bool, err error) {
	if mt.mt != nil {
		mt.mt.ReservationRequestFinished(isRefresh, err)
	}
}

func (mt *wrappedMetricsTracer) RelayAddressUpdated() {
	if mt.mt != nil {
		mt.mt.RelayAddressUpdated()
	}
}

func (mt *wrappedMetricsTracer) RelayAddressCount(cnt int) {
	if mt.mt != nil {
		mt.mt.RelayAddressCount(cnt)
	}
}

func (mt *wrappedMetricsTracer) CandidateChecked(supportsCircuitV2 bool) {
	if mt.mt != nil {
		mt.mt.CandidateChecked(supportsCircuitV2)
	}
}

func (mt *wrappedMetricsTracer) CandidateAdded(cnt int) {
	if mt.mt != nil {
		mt.mt.CandidateAdded(cnt)
	}
}

func (mt *wrappedMetricsTracer) CandidateRemoved(cnt int) {
	if mt.mt != nil {
		mt.mt.CandidateRemoved(cnt)
	}
}

func (mt *wrappedMetricsTracer) ScheduledWorkUpdated(scheduledWork *scheduledWorkTimes) {
	if mt.mt != nil {
		mt.mt.ScheduledWorkUpdated(scheduledWork)
	}
}

func (mt *wrappedMetricsTracer) DesiredReservations(cnt int) {
	if mt.mt != nil {
		mt.mt.DesiredReservations(cnt)
	}
}

func (mt *wrappedMetricsTracer) CandidateLoopState(state candidateLoopState) {
	if mt.mt != nil {
		mt.mt.CandidateLoopState(state)
	}
}
