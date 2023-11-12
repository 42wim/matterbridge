package dynamic

import (
	"time"

	"github.com/libp2p/go-libp2p/p2p/metricshelper"
	"github.com/prometheus/client_golang/prometheus"
)

var numberRegisteredMemberships = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "waku_rln_number_registered_memberships",
		Help: "number of registered and active rln memberships",
	})

var membershipInsertionDurationSeconds = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "waku_rln_membership_insertion_duration_seconds",
		Help: "time taken to insert a new member into the local merkle tree",
	})

var membershipCredentialsImportDurationSeconds = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "waku_rln_membership_credentials_import_duration_seconds",
		Help: "time taken to import membership credentials",
	})

var collectors = []prometheus.Collector{
	numberRegisteredMemberships,
	membershipInsertionDurationSeconds,
	membershipCredentialsImportDurationSeconds,
}

// Metrics exposes the functions required to update prometheus metrics for lightpush protocol
type Metrics interface {
	RecordRegisteredMembership(num uint)
	RecordMembershipInsertionDuration(duration time.Duration)
	RecordMembershipCredentialsImportDuration(duration time.Duration)
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

// RecordMembershipInsertionDuration records how long did it take to insert members into th merkle tree
func (m *metricsImpl) RecordMembershipInsertionDuration(duration time.Duration) {
	membershipInsertionDurationSeconds.Set(duration.Seconds())
}

// RecordMembershipCredentialsImport records how long did it take to import the membership credentials
func (m *metricsImpl) RecordMembershipCredentialsImportDuration(duration time.Duration) {
	membershipCredentialsImportDurationSeconds.Set(duration.Seconds())
}

// RecordRegisteredMembership records the number of registered memberships
func (m *metricsImpl) RecordRegisteredMembership(num uint) {
	numberRegisteredMemberships.Set(float64(num))
}
