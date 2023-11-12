package pb

import "errors"

var (
	errMissingRequestID   = errors.New("missing RequestId field")
	errMissingQuery       = errors.New("missing Query field")
	errMissingMessage     = errors.New("missing Message field")
	errMissingPubsubTopic = errors.New("missing PubsubTopic field")
	errRequestIDMismatch  = errors.New("requestID in response does not match request")
	errMissingResponse    = errors.New("missing Response field")
)

func (x *PushRpc) ValidateRequest() error {
	if x.RequestId == "" {
		return errMissingRequestID
	}

	if x.Request == nil {
		return errMissingQuery
	}

	if x.Request.PubsubTopic == "" {
		return errMissingPubsubTopic
	}

	if x.Request.Message == nil {
		return errMissingMessage
	}

	return x.Request.Message.Validate()
}

func (x *PushRpc) ValidateResponse(requestID string) error {
	if x.RequestId == "" {
		return errMissingRequestID
	}

	if x.RequestId != requestID {
		return errRequestIDMismatch
	}

	if x.Response == nil {
		return errMissingResponse
	}

	return nil
}
