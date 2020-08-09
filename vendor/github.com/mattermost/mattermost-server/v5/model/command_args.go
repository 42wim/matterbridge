// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"io"

	goi18n "github.com/mattermost/go-i18n/i18n"
)

type CommandArgs struct {
	UserId          string               `json:"user_id"`
	ChannelId       string               `json:"channel_id"`
	TeamId          string               `json:"team_id"`
	RootId          string               `json:"root_id"`
	ParentId        string               `json:"parent_id"`
	TriggerId       string               `json:"trigger_id,omitempty"`
	Command         string               `json:"command"`
	SiteURL         string               `json:"-"`
	T               goi18n.TranslateFunc `json:"-"`
	Session         Session              `json:"-"`
	UserMentions    UserMentionMap       `json:"-"`
	ChannelMentions ChannelMentionMap    `json:"-"`
}

func (o *CommandArgs) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func CommandArgsFromJson(data io.Reader) *CommandArgs {
	var o *CommandArgs
	json.NewDecoder(data).Decode(&o)
	return o
}

// AddUserMention adds or overrides an entry in UserMentions with name username
// and identifier userId
func (o *CommandArgs) AddUserMention(username, userId string) {
	if o.UserMentions == nil {
		o.UserMentions = make(UserMentionMap)
	}

	o.UserMentions[username] = userId
}

// AddChannelMention adds or overrides an entry in ChannelMentions with name
// channelName and identifier channelId
func (o *CommandArgs) AddChannelMention(channelName, channelId string) {
	if o.ChannelMentions == nil {
		o.ChannelMentions = make(ChannelMentionMap)
	}

	o.ChannelMentions[channelName] = channelId
}
