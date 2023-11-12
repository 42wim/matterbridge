package protocol

import (
	"fmt"
	"sort"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/requests"
)

type MetricsIntervalResponse struct {
	StartTimestamp uint64   `json:"startTimestamp"`
	EndTimestamp   uint64   `json:"endTimestamp"`
	Timestamps     []uint64 `json:"timestamps"`
	Count          int      `json:"count"`
}

type CommunityMetricsResponse struct {
	Type        requests.CommunityMetricsRequestType `json:"type"`
	CommunityID types.HexBytes                       `json:"communityId"`
	Intervals   []MetricsIntervalResponse            `json:"intervals"`
}

func (m *Messenger) getChatIdsForCommunity(communityID types.HexBytes) ([]string, error) {
	community, err := m.GetCommunityByID(communityID)
	if err != nil {
		return []string{}, err
	}
	return community.ChatIDs(), nil
}

func (m *Messenger) collectCommunityMessagesTimestamps(request *requests.CommunityMetricsRequest) (*CommunityMetricsResponse, error) {
	chatIDs, err := m.getChatIdsForCommunity(request.CommunityID)
	if err != nil {
		return nil, err
	}

	intervals := make([]MetricsIntervalResponse, len(request.Intervals))
	for i, sourceInterval := range request.Intervals {
		// TODO: messages count should be stored in special table, not calculated here
		timestamps, err := m.persistence.SelectMessagesTimestampsForChatsByPeriod(chatIDs, sourceInterval.StartTimestamp, sourceInterval.EndTimestamp)
		if err != nil {
			return nil, err
		}

		// there is no built-in sort for uint64
		sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })

		intervals[i] = MetricsIntervalResponse{
			StartTimestamp: sourceInterval.StartTimestamp,
			EndTimestamp:   sourceInterval.EndTimestamp,
			Timestamps:     timestamps,
			Count:          len(timestamps),
		}
	}

	response := &CommunityMetricsResponse{
		Type:        request.Type,
		CommunityID: request.CommunityID,
		Intervals:   intervals,
	}

	return response, nil
}

func (m *Messenger) collectCommunityMessagesCount(request *requests.CommunityMetricsRequest) (*CommunityMetricsResponse, error) {
	chatIDs, err := m.getChatIdsForCommunity(request.CommunityID)
	if err != nil {
		return nil, err
	}

	intervals := make([]MetricsIntervalResponse, len(request.Intervals))
	for i, sourceInterval := range request.Intervals {
		// TODO: messages count should be stored in special table, not calculated here
		count, err := m.persistence.SelectMessagesCountForChatsByPeriod(chatIDs, sourceInterval.StartTimestamp, sourceInterval.EndTimestamp)
		if err != nil {
			return nil, err
		}
		intervals[i] = MetricsIntervalResponse{
			StartTimestamp: sourceInterval.StartTimestamp,
			EndTimestamp:   sourceInterval.EndTimestamp,
			Count:          count,
		}
	}

	response := &CommunityMetricsResponse{
		Type:        request.Type,
		CommunityID: request.CommunityID,
		Intervals:   intervals,
	}

	return response, nil
}

func (m *Messenger) CollectCommunityMetrics(request *requests.CommunityMetricsRequest) (*CommunityMetricsResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	switch request.Type {
	case requests.CommunityMetricsRequestMessagesTimestamps:
		return m.collectCommunityMessagesTimestamps(request)
	case requests.CommunityMetricsRequestMessagesCount:
		return m.collectCommunityMessagesCount(request)
	default:
		return nil, fmt.Errorf("metrics for %d is not implemented yet", request.Type)
	}
}
