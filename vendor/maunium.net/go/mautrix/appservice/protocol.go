// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package appservice

import (
	"encoding/json"
	"net/http"

	"maunium.net/go/mautrix/event"
)

// EventList contains a list of events.
type EventList struct {
	Events              []*event.Event `json:"events"`
	EphemeralEvents     []*event.Event `json:"ephemeral"`
	SoruEphemeralEvents []*event.Event `json:"de.sorunome.msc2409.ephemeral"`
}

// EventListener is a function that receives events.
type EventListener func(evt *event.Event)

// WriteBlankOK writes a blank OK message as a reply to a HTTP request.
func WriteBlankOK(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{}"))
}

// Respond responds to a HTTP request with a JSON object.
func Respond(w http.ResponseWriter, data interface{}) error {
	w.Header().Add("Content-Type", "application/json")
	dataStr, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(dataStr)
	return err
}

// Error represents a Matrix protocol error.
type Error struct {
	HTTPStatus int       `json:"-"`
	ErrorCode  ErrorCode `json:"errcode"`
	Message    string    `json:"message"`
}

func (err Error) Write(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatus)
	_ = Respond(w, &err)
}

// ErrorCode is the machine-readable code in an Error.
type ErrorCode string

// Native ErrorCodes
const (
	ErrUnknownToken ErrorCode = "M_UNKNOWN_TOKEN"
	ErrBadJSON      ErrorCode = "M_BAD_JSON"
	ErrNotJSON      ErrorCode = "M_NOT_JSON"
	ErrUnknown      ErrorCode = "M_UNKNOWN"
)

// Custom ErrorCodes
const (
	ErrNoTransactionID ErrorCode = "NET.MAUNIUM.NO_TRANSACTION_ID"
)
