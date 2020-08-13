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

package talk

import (
	"gomod.garykim.dev/nc-talk/room"
	"gomod.garykim.dev/nc-talk/user"
)

// NewUser returns a TalkUser instance
// The url should be the full URL of the Nextcloud instance (e.g. https://cloud.mydomain.me)
//
// Deprecated: Use user.NewUser instead for more options and error checks
func NewUser(url string, username string, password string) *user.TalkUser {
	return &user.TalkUser{
		NextcloudURL: url,
		User:         username,
		Pass:         password,
	}
}

// NewRoom returns a new TalkRoom instance
// Token should be the Nextcloud Room Token (e.g. "d6zoa2zs" if the room URL is https://cloud.mydomain.me/call/d6zoa2zs)
//
// Deprecated: Use room.NewRoom instead for extra error checks.
func NewRoom(tuser *user.TalkUser, token string) *room.TalkRoom {
	tr := &room.TalkRoom{
		User:  tuser,
		Token: token,
	}
	return tr
}
