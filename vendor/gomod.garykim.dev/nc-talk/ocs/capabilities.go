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

package ocs

// Capabilities describes the response from the capabilities request
type Capabilities struct {
	ocs
	Data struct {
		Capabilities struct {
			SpreedCapabilities SpreedCapabilities `json:"spreed"`
		} `json:"capabilities"`
	} `json:"data"`
}

// SpreedCapabilities describes the Nextcloud Talk capabilities response
type SpreedCapabilities struct {
	Features []string `json:"features"`
	Config   struct {
		Attachments struct {
			Allowed bool   `json:"allowed"`
			Folder  string `json:"folder"`
		} `json:"attachments"`
		Chat struct {
			MaxLength   int `json:"max-length"`
			ReadPrivacy int `json:"read-privacy"`
		} `json:"chat"`
		Conversations struct {
			CanCreate bool `json:"can-create"`
		} `json:"conversations"`
		Previews struct {
			MaxGifSize int `json:"max-gif-size"`
		} `json:"previews"`
	} `json:"config"`
}
