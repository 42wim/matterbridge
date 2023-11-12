package pb

import (
	"errors"
	"fmt"

	"golang.org/x/exp/slices"
)

const MaxContentTopicsPerRequest = 30

var (
	errMissingRequestID   = errors.New("missing RequestId field")
	errMissingPubsubTopic = errors.New("missing PubsubTopic field")
	errNoContentTopics    = errors.New("at least one contenttopic should be specified")
	errMaxContentTopics   = fmt.Errorf("exceeds maximum content topics: %d", MaxContentTopicsPerRequest)
	errEmptyContentTopics = errors.New("one or more content topics specified is empty")
	errMissingMessage     = errors.New("missing WakuMessage field")
)

func (x *FilterSubscribeRequest) Validate() error {
	if x.RequestId == "" {
		return errMissingRequestID
	}

	if x.FilterSubscribeType == FilterSubscribeRequest_SUBSCRIBE || x.FilterSubscribeType == FilterSubscribeRequest_UNSUBSCRIBE {
		if x.PubsubTopic == nil || *x.PubsubTopic == "" {
			return errMissingPubsubTopic
		}

		if len(x.ContentTopics) == 0 {
			return errNoContentTopics
		}

		if slices.Contains[string](x.ContentTopics, "") {
			return errEmptyContentTopics
		}

		if len(x.ContentTopics) > MaxContentTopicsPerRequest {
			return errMaxContentTopics
		}
	}

	return nil
}

func (x *FilterSubscribeResponse) Validate() error {
	if x.RequestId == "" {
		return errMissingRequestID
	}

	return nil
}

func (x *MessagePush) Validate() error {
	if x.WakuMessage == nil {
		return errMissingMessage
	}
	return x.WakuMessage.Validate()
}
