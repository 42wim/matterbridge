package peer

import (
	"context"
	"errors"
)

var (
	// ErrInvalidTopic error returned when the requested topic is not valid.
	ErrInvalidTopic = errors.New("topic not valid")

	// ErrInvalidRange error returned when max-min range is not valid.
	ErrInvalidRange = errors.New("invalid range, Min should be lower or equal to Max")

	// ErrDiscovererNotProvided error when discoverer is not being provided.
	ErrDiscovererNotProvided = errors.New("discoverer not provided")
)

// PublicAPI represents a set of APIs from the `web3.peer` namespace.
type PublicAPI struct {
	s *Service
}

// NewAPI creates an instance of the peer API.
func NewAPI(s *Service) *PublicAPI {
	return &PublicAPI{s: s}
}

// DiscoverRequest json request for peer_discover.
type DiscoverRequest struct {
	Topic string `json:"topic"`
	Max   int    `json:"max"`
	Min   int    `json:"min"`
}

// Discover is an implementation of `peer_discover` or `web3.peer.discover` API.
func (api *PublicAPI) Discover(context context.Context, req DiscoverRequest) (err error) {
	if api.s.d == nil {
		return ErrDiscovererNotProvided
	}
	if len(req.Topic) == 0 {
		return ErrInvalidTopic
	}
	if req.Max < req.Min {
		return ErrInvalidRange
	}
	return api.s.d.Discover(req.Topic, req.Max, req.Min)
}
