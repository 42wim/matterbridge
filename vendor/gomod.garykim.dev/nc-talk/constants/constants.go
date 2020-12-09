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

package constants

const (
	// BaseEndpoint is the api endpoint for Nextcloud Talk
	BaseEndpoint = "/ocs/v2.php/apps/spreed/api/v1/"
)

// RemoteDavEndpoint returns the endpoint for the Dav API for Nextcloud
func RemoteDavEndpoint(username string, davType string) string {
	return "/remote.php/dav/" + davType + "/" + username + "/"
}
