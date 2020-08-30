// Copyright (c) 2020 Gary Kim <gary@garykim.dev>, All Rights Reserved
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// File describes Nextcloud's Rich Object Strings (https://github.com/nextcloud/server/issues/1706)

package ocs

// RichObjectString describes the content of placeholders in TalkRoomMessageData
type RichObjectString struct {
	Type RichObjectStringType `json:"type"`
	ID   string               `json:"id"`
	Name string               `json:"name"`
	Path string               `json:"path"`
	Link string               `json:"link"`
}

// RichObjectStringType describes what a rich object string is describing
type RichObjectStringType string

const (
	// ROSTypeUser describes a rich object string that is a user
	ROSTypeUser RichObjectStringType = "user"
	// ROSTypeGroup describes a rich object string that is a group
	ROSTypeGroup RichObjectStringType = "group"
	// ROSTypeFile describes a rich object string that is a file
	ROSTypeFile RichObjectStringType = "file"
)
