package pb

import (
	"errors"
)

// MaxContentFilters is the maximum number of allowed content filters in a query
const MaxContentFilters = 10

var (
	errMissingRequestID   = errors.New("missing RequestId field")
	errMissingQuery       = errors.New("missing Query field")
	errRequestIDMismatch  = errors.New("requestID in response does not match request")
	errMaxContentFilters  = errors.New("exceeds the maximum number of content filters allowed")
	errEmptyContentTopics = errors.New("one or more content topics specified is empty")
)

func (x *HistoryQuery) Validate() error {
	if len(x.ContentFilters) > MaxContentFilters {
		return errMaxContentFilters
	}

	for _, m := range x.ContentFilters {
		if m.ContentTopic == "" {
			return errEmptyContentTopics
		}
	}

	return nil
}

func (x *HistoryRPC) ValidateQuery() error {
	if x.RequestId == "" {
		return errMissingRequestID
	}

	if x.Query == nil {
		return errMissingQuery
	}

	return x.Query.Validate()
}

func (x *HistoryResponse) Validate() error {
	for _, m := range x.Messages {
		if err := m.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (x *HistoryRPC) ValidateResponse(requestID string) error {
	if x.RequestId == "" {
		return errMissingRequestID
	}

	if x.RequestId != requestID {
		return errRequestIDMismatch
	}

	if x.Response != nil {
		return x.Response.Validate()
	}

	return nil
}
