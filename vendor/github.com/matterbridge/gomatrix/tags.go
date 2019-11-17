// Copyright 2019 Sumukha PK
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

// TagContent contains the data for an m.tag message type
// https://matrix.org/docs/spec/client_server/r0.4.0.html#m-tag
type TagContent struct {
	Tags map[string]TagProperties `json:"tags"`
}

// TagProperties contains the properties of a Tag
type TagProperties struct {
	Order float32 `json:"order,omitempty"` // Empty values must be neglected
}
