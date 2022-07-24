// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package appservice

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type OTKCountMap = map[id.UserID]map[id.DeviceID]mautrix.OTKCount
type FallbackKeyMap = map[id.UserID]map[id.DeviceID][]id.KeyAlgorithm

// Transaction contains a list of events.
type Transaction struct {
	Events          []*event.Event `json:"events"`
	EphemeralEvents []*event.Event `json:"ephemeral,omitempty"`
	ToDeviceEvents  []*event.Event `json:"to_device,omitempty"`

	DeviceLists    *mautrix.DeviceLists `json:"device_lists,omitempty"`
	DeviceOTKCount OTKCountMap          `json:"device_one_time_keys_count,omitempty"`
	FallbackKeys   FallbackKeyMap       `json:"device_unused_fallback_key_types,omitempty"`

	MSC2409EphemeralEvents []*event.Event       `json:"de.sorunome.msc2409.ephemeral,omitempty"`
	MSC2409ToDeviceEvents  []*event.Event       `json:"de.sorunome.msc2409.to_device,omitempty"`
	MSC3202DeviceLists     *mautrix.DeviceLists `json:"org.matrix.msc3202.device_lists,omitempty"`
	MSC3202DeviceOTKCount  OTKCountMap          `json:"org.matrix.msc3202.device_one_time_keys_count,omitempty"`
	MSC3202FallbackKeys    FallbackKeyMap       `json:"org.matrix.msc3202.device_unused_fallback_key_types,omitempty"`
}

func (txn *Transaction) MarshalZerologObject(ctx *zerolog.Event) {
	ctx.Int("pdu", len(txn.Events))
	if txn.EphemeralEvents != nil {
		ctx.Int("edu", len(txn.EphemeralEvents))
	} else if txn.MSC2409EphemeralEvents != nil {
		ctx.Int("unstable_edu", len(txn.MSC2409EphemeralEvents))
	}
	if txn.ToDeviceEvents != nil {
		ctx.Int("to_device", len(txn.ToDeviceEvents))
	} else if txn.MSC2409ToDeviceEvents != nil {
		ctx.Int("unstable_to_device", len(txn.MSC2409ToDeviceEvents))
	}
	if len(txn.DeviceOTKCount) > 0 {
		ctx.Int("otk_count_users", len(txn.DeviceOTKCount))
	} else if len(txn.MSC3202DeviceOTKCount) > 0 {
		ctx.Int("unstable_otk_count_users", len(txn.MSC3202DeviceOTKCount))
	}
	if txn.DeviceLists != nil {
		ctx.Int("device_changes", len(txn.DeviceLists.Changed))
	} else if txn.MSC3202DeviceLists != nil {
		ctx.Int("unstable_device_changes", len(txn.MSC3202DeviceLists.Changed))
	}
	if txn.FallbackKeys != nil {
		ctx.Int("fallback_key_users", len(txn.FallbackKeys))
	} else if txn.MSC3202FallbackKeys != nil {
		ctx.Int("unstable_fallback_key_users", len(txn.MSC3202FallbackKeys))
	}
}

func (txn *Transaction) ContentString() string {
	var parts []string
	if len(txn.Events) > 0 {
		parts = append(parts, fmt.Sprintf("%d PDUs", len(txn.Events)))
	}
	if len(txn.EphemeralEvents) > 0 {
		parts = append(parts, fmt.Sprintf("%d EDUs", len(txn.EphemeralEvents)))
	} else if len(txn.MSC2409EphemeralEvents) > 0 {
		parts = append(parts, fmt.Sprintf("%d EDUs (unstable)", len(txn.MSC2409EphemeralEvents)))
	}
	if len(txn.ToDeviceEvents) > 0 {
		parts = append(parts, fmt.Sprintf("%d to-device events", len(txn.ToDeviceEvents)))
	} else if len(txn.MSC2409ToDeviceEvents) > 0 {
		parts = append(parts, fmt.Sprintf("%d to-device events (unstable)", len(txn.MSC2409ToDeviceEvents)))
	}
	if len(txn.DeviceOTKCount) > 0 {
		parts = append(parts, fmt.Sprintf("OTK counts for %d users", len(txn.DeviceOTKCount)))
	} else if len(txn.MSC3202DeviceOTKCount) > 0 {
		parts = append(parts, fmt.Sprintf("OTK counts for %d users (unstable)", len(txn.MSC3202DeviceOTKCount)))
	}
	if txn.DeviceLists != nil {
		parts = append(parts, fmt.Sprintf("%d device list changes", len(txn.DeviceLists.Changed)))
	} else if txn.MSC3202DeviceLists != nil {
		parts = append(parts, fmt.Sprintf("%d device list changes (unstable)", len(txn.MSC3202DeviceLists.Changed)))
	}
	if txn.FallbackKeys != nil {
		parts = append(parts, fmt.Sprintf("unused fallback key counts for %d users", len(txn.FallbackKeys)))
	} else if txn.MSC3202FallbackKeys != nil {
		parts = append(parts, fmt.Sprintf("unused fallback key counts for %d users (unstable)", len(txn.MSC3202FallbackKeys)))
	}
	return strings.Join(parts, ", ")
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
	Message    string    `json:"error"`
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
