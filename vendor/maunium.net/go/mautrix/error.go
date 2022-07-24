// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package mautrix

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Common error codes from https://matrix.org/docs/spec/client_server/latest#api-standards
//
// Can be used with errors.Is() to check the response code without casting the error:
//
//	err := client.Sync()
//	if errors.Is(err, MUnknownToken) {
//		// logout
//	}
var (
	// Forbidden access, e.g. joining a room without permission, failed login.
	MForbidden = RespError{ErrCode: "M_FORBIDDEN"}
	// Unrecognized request, e.g. the endpoint does not exist or is not implemented.
	MUnrecognized = RespError{ErrCode: "M_UNRECOGNIZED"}
	// The access token specified was not recognised.
	MUnknownToken = RespError{ErrCode: "M_UNKNOWN_TOKEN"}
	// No access token was specified for the request.
	MMissingToken = RespError{ErrCode: "M_MISSING_TOKEN"}
	// Request contained valid JSON, but it was malformed in some way, e.g. missing required keys, invalid values for keys.
	MBadJSON = RespError{ErrCode: "M_BAD_JSON"}
	// Request did not contain valid JSON.
	MNotJSON = RespError{ErrCode: "M_NOT_JSON"}
	// No resource was found for this request.
	MNotFound = RespError{ErrCode: "M_NOT_FOUND"}
	// Too many requests have been sent in a short period of time. Wait a while then try again.
	MLimitExceeded = RespError{ErrCode: "M_LIMIT_EXCEEDED"}
	// The user ID associated with the request has been deactivated.
	// Typically for endpoints that prove authentication, such as /login.
	MUserDeactivated = RespError{ErrCode: "M_USER_DEACTIVATED"}
	// Encountered when trying to register a user ID which has been taken.
	MUserInUse = RespError{ErrCode: "M_USER_IN_USE"}
	// Encountered when trying to register a user ID which is not valid.
	MInvalidUsername = RespError{ErrCode: "M_INVALID_USERNAME"}
	// Sent when the room alias given to the createRoom API is already in use.
	MRoomInUse = RespError{ErrCode: "M_ROOM_IN_USE"}
	// The state change requested cannot be performed, such as attempting to unban a user who is not banned.
	MBadState = RespError{ErrCode: "M_BAD_STATE"}
	// The request or entity was too large.
	MTooLarge = RespError{ErrCode: "M_TOO_LARGE"}
	// The resource being requested is reserved by an application service, or the application service making the request has not created the resource.
	MExclusive = RespError{ErrCode: "M_EXCLUSIVE"}
	// The client's request to create a room used a room version that the server does not support.
	MUnsupportedRoomVersion = RespError{ErrCode: "M_UNSUPPORTED_ROOM_VERSION"}
	// The client attempted to join a room that has a version the server does not support.
	// Inspect the room_version property of the error response for the room's version.
	MIncompatibleRoomVersion = RespError{ErrCode: "M_INCOMPATIBLE_ROOM_VERSION"}
	// The client specified a parameter that has the wrong value.
	MInvalidParam = RespError{ErrCode: "M_INVALID_PARAM"}

	MSC2659URLNotSet         = RespError{ErrCode: "FI.MAU.MSC2659_URL_NOT_SET"}
	MSC2659BadStatus         = RespError{ErrCode: "FI.MAU.MSC2659_BAD_STATUS"}
	MSC2659ConnectionTimeout = RespError{ErrCode: "FI.MAU.MSC2659_CONNECTION_TIMEOUT"}
	MSC2659ConnectionFailed  = RespError{ErrCode: "FI.MAU.MSC2659_CONNECTION_FAILED"}
)

// HTTPError An HTTP Error response, which may wrap an underlying native Go Error.
type HTTPError struct {
	Request      *http.Request
	Response     *http.Response
	ResponseBody string

	WrappedError error
	RespError    *RespError
	Message      string
}

func (e HTTPError) Is(err error) bool {
	return (e.RespError != nil && errors.Is(e.RespError, err)) || (e.WrappedError != nil && errors.Is(e.WrappedError, err))
}

func (e HTTPError) IsStatus(code int) bool {
	return e.Response != nil && e.Response.StatusCode == code
}

func (e HTTPError) Error() string {
	if e.WrappedError != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.WrappedError)
	} else if e.RespError != nil {
		return fmt.Sprintf("failed to %s %s: %s (HTTP %d): %s", e.Request.Method, e.Request.URL.Path,
			e.RespError.ErrCode, e.Response.StatusCode, e.RespError.Err)
	} else {
		msg := fmt.Sprintf("failed to %s %s: %s", e.Request.Method, e.Request.URL.Path, e.Response.Status)
		if len(e.ResponseBody) > 0 {
			msg = fmt.Sprintf("%s\n%s", msg, e.ResponseBody)
		}
		return msg
	}
}

func (e HTTPError) Unwrap() error {
	if e.WrappedError != nil {
		return e.WrappedError
	} else if e.RespError != nil {
		return *e.RespError
	}
	return nil
}

// RespError is the standard JSON error response from Homeservers. It also implements the Golang "error" interface.
// See https://spec.matrix.org/v1.2/client-server-api/#api-standards
type RespError struct {
	ErrCode   string
	Err       string
	ExtraData map[string]interface{}
}

func (e *RespError) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &e.ExtraData)
	if err != nil {
		return err
	}
	e.ErrCode, _ = e.ExtraData["errcode"].(string)
	e.Err, _ = e.ExtraData["error"].(string)
	return nil
}

func (e *RespError) MarshalJSON() ([]byte, error) {
	if e.ExtraData == nil {
		e.ExtraData = make(map[string]interface{})
	}
	e.ExtraData["errcode"] = e.ErrCode
	e.ExtraData["error"] = e.Err
	return json.Marshal(&e.ExtraData)
}

// Error returns the errcode and error message.
func (e RespError) Error() string {
	return e.ErrCode + ": " + e.Err
}

func (e RespError) Is(err error) bool {
	e2, ok := err.(RespError)
	if !ok {
		return false
	}
	if e.ErrCode == "M_UNKNOWN" && e2.ErrCode == "M_UNKNOWN" {
		return e.Err == e2.Err
	}
	return e2.ErrCode == e.ErrCode
}
