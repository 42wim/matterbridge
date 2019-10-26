// Copyright 2017 Jan Christian Grünhage
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gomatrix

import "errors"

//Filter is used by clients to specify how the server should filter responses to e.g. sync requests
//Specified by: https://matrix.org/docs/spec/client_server/r0.2.0.html#filtering
type Filter struct {
	AccountData FilterPart `json:"account_data,omitempty"`
	EventFields []string   `json:"event_fields,omitempty"`
	EventFormat string     `json:"event_format,omitempty"`
	Presence    FilterPart `json:"presence,omitempty"`
	Room        RoomFilter `json:"room,omitempty"`
}

// RoomFilter is used to define filtering rules for room events
type RoomFilter struct {
	AccountData  FilterPart `json:"account_data,omitempty"`
	Ephemeral    FilterPart `json:"ephemeral,omitempty"`
	IncludeLeave bool       `json:"include_leave,omitempty"`
	NotRooms     []string   `json:"not_rooms,omitempty"`
	Rooms        []string   `json:"rooms,omitempty"`
	State        FilterPart `json:"state,omitempty"`
	Timeline     FilterPart `json:"timeline,omitempty"`
}

// FilterPart is used to define filtering rules for specific categories of events
type FilterPart struct {
	NotRooms    []string `json:"not_rooms,omitempty"`
	Rooms       []string `json:"rooms,omitempty"`
	Limit       int      `json:"limit,omitempty"`
	NotSenders  []string `json:"not_senders,omitempty"`
	NotTypes    []string `json:"not_types,omitempty"`
	Senders     []string `json:"senders,omitempty"`
	Types       []string `json:"types,omitempty"`
	ContainsURL *bool    `json:"contains_url,omitempty"`
}

// Validate checks if the filter contains valid property values
func (filter *Filter) Validate() error {
	if filter.EventFormat != "client" && filter.EventFormat != "federation" {
		return errors.New("Bad event_format value. Must be one of [\"client\", \"federation\"]")
	}
	return nil
}

// DefaultFilter returns the default filter used by the Matrix server if no filter is provided in the request
func DefaultFilter() Filter {
	return Filter{
		AccountData: DefaultFilterPart(),
		EventFields: nil,
		EventFormat: "client",
		Presence:    DefaultFilterPart(),
		Room: RoomFilter{
			AccountData:  DefaultFilterPart(),
			Ephemeral:    DefaultFilterPart(),
			IncludeLeave: false,
			NotRooms:     nil,
			Rooms:        nil,
			State:        DefaultFilterPart(),
			Timeline:     DefaultFilterPart(),
		},
	}
}

// DefaultFilterPart returns the default filter part used by the Matrix server if no filter is provided in the request
func DefaultFilterPart() FilterPart {
	return FilterPart{
		NotRooms:   nil,
		Rooms:      nil,
		Limit:      20,
		NotSenders: nil,
		NotTypes:   nil,
		Senders:    nil,
		Types:      nil,
	}
}
