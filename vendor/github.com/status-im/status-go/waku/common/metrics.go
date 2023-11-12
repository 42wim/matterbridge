// Copyright 2019 The Waku Library Authors.
//
// The Waku library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Waku library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty off
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Waku library. If not, see <http://www.gnu.org/licenses/>.
//
// This software uses the go-ethereum library, which is licensed
// under the GNU Lesser General Public Library, version 3 or any later.

package common

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

var (
	EnvelopesReceivedCounter = prom.NewCounter(prom.CounterOpts{
		Name: "waku_envelopes_received_total",
		Help: "Number of envelopes received.",
	})
	EnvelopesValidatedCounter = prom.NewCounter(prom.CounterOpts{
		Name: "waku_envelopes_validated_total",
		Help: "Number of envelopes processed successfully.",
	})
	EnvelopesRejectedCounter = prom.NewCounterVec(prom.CounterOpts{
		Name: "waku_envelopes_rejected_total",
		Help: "Number of envelopes rejected.",
	}, []string{"reason"})
	EnvelopesCacheFailedCounter = prom.NewCounterVec(prom.CounterOpts{
		Name: "waku_envelopes_cache_failures_total",
		Help: "Number of envelopes which failed to be cached.",
	}, []string{"type"})
	EnvelopesCachedCounter = prom.NewCounterVec(prom.CounterOpts{
		Name: "waku_envelopes_cached_total",
		Help: "Number of envelopes cached.",
	}, []string{"cache"})
	EnvelopesSizeMeter = prom.NewHistogram(prom.HistogramOpts{
		Name:    "waku_envelopes_size_bytes",
		Help:    "Size of processed Waku envelopes in bytes.",
		Buckets: prom.ExponentialBuckets(256, 4, 10),
	})
	RateLimitsProcessed = prom.NewCounter(prom.CounterOpts{
		Name: "waku_rate_limits_processed_total",
		Help: "Number of packets Waku rate limiter processed.",
	})
	RateLimitsExceeded = prom.NewCounterVec(prom.CounterOpts{
		Name: "waku_rate_limits_exceeded_total",
		Help: "Number of times the Waku rate limits were exceeded",
	}, []string{"type"})
	BridgeSent = prom.NewCounter(prom.CounterOpts{
		Name: "waku_bridge_sent_total",
		Help: "Number of envelopes bridged from Waku",
	})
	BridgeReceivedSucceed = prom.NewCounter(prom.CounterOpts{
		Name: "waku_bridge_received_success_total",
		Help: "Number of envelopes bridged to Waku and successfully added",
	})
	BridgeReceivedFailed = prom.NewCounter(prom.CounterOpts{
		Name: "waku_bridge_received_failure_total",
		Help: "Number of envelopes bridged to Waku and failed to be added",
	})
)

func init() {
	prom.MustRegister(EnvelopesReceivedCounter)
	prom.MustRegister(EnvelopesRejectedCounter)
	prom.MustRegister(EnvelopesCacheFailedCounter)
	prom.MustRegister(EnvelopesCachedCounter)
	prom.MustRegister(EnvelopesSizeMeter)
	prom.MustRegister(RateLimitsProcessed)
	prom.MustRegister(RateLimitsExceeded)
	prom.MustRegister(BridgeSent)
	prom.MustRegister(BridgeReceivedSucceed)
	prom.MustRegister(BridgeReceivedFailed)
}
