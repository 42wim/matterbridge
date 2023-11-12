package rpcstats

import (
	"context"
)

// PublicAPI represents a set of APIs from the namespace.
type PublicAPI struct {
	s *Service
}

// NewAPI creates an instance of the API.
func NewAPI(s *Service) *PublicAPI {
	return &PublicAPI{s: s}
}

// Reset resets RPC usage stats
func (api *PublicAPI) Reset(context context.Context) {
	resetStats()
}

type RPCStats struct {
	Total            uint            `json:"total"`
	CounterPerMethod map[string]uint `json:"methods"`
}

// GetStats retrun RPC usage stats
func (api *PublicAPI) GetStats(context context.Context) (RPCStats, error) {
	total, perMethod := getStats()
	return RPCStats{
		Total:            total,
		CounterPerMethod: perMethod,
	}, nil
}
