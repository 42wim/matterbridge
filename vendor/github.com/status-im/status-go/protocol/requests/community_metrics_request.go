package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrNoCommunityID = errors.New("community metrics request has no community id")
var ErrInvalidTimestampIntervals = errors.New("community metrics request invalid time intervals")

type CommunityMetricsRequestType uint

const (
	CommunityMetricsRequestMessagesTimestamps CommunityMetricsRequestType = iota
	CommunityMetricsRequestMessagesCount
	CommunityMetricsRequestMembers
	CommunityMetricsRequestControlNodeUptime
)

type MetricsIntervalRequest struct {
	StartTimestamp uint64 `json:"startTimestamp"`
	EndTimestamp   uint64 `json:"endTimestamp"`
}

type CommunityMetricsRequest struct {
	CommunityID types.HexBytes              `json:"communityId"`
	Type        CommunityMetricsRequestType `json:"type"`
	Intervals   []MetricsIntervalRequest    `json:"intervals"`
}

func (r *CommunityMetricsRequest) Validate() error {
	if len(r.CommunityID) == 0 {
		return ErrNoCommunityID
	}

	for _, interval := range r.Intervals {
		if interval.StartTimestamp == 0 || interval.EndTimestamp == 0 || interval.StartTimestamp >= interval.EndTimestamp {
			return ErrInvalidTimestampIntervals
		}
	}

	return nil
}
