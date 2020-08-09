// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"regexp"
	"time"
)

type CompliancePost struct {

	// From Team
	TeamName        string
	TeamDisplayName string

	// From Channel
	ChannelName        string
	ChannelDisplayName string
	ChannelType        string

	// From User
	UserUsername string
	UserEmail    string
	UserNickname string

	// From Post
	PostId         string
	PostCreateAt   int64
	PostUpdateAt   int64
	PostDeleteAt   int64
	PostRootId     string
	PostParentId   string
	PostOriginalId string
	PostMessage    string
	PostType       string
	PostProps      string
	PostHashtags   string
	PostFileIds    string

	IsBot bool
}

func CompliancePostHeader() []string {
	return []string{
		"TeamName",
		"TeamDisplayName",

		"ChannelName",
		"ChannelDisplayName",
		"ChannelType",

		"UserUsername",
		"UserEmail",
		"UserNickname",

		"PostId",
		"PostCreateAt",
		"PostUpdateAt",
		"PostDeleteAt",
		"PostRootId",
		"PostParentId",
		"PostOriginalId",
		"PostMessage",
		"PostType",
		"PostProps",
		"PostHashtags",
		"PostFileIds",
		"UserType",
	}
}

func cleanComplianceStrings(in string) string {
	if matched, _ := regexp.MatchString("^\\s*(=|\\+|\\-)", in); matched {
		return "'" + in

	} else {
		return in
	}
}

func (me *CompliancePost) Row() []string {

	postDeleteAt := ""
	if me.PostDeleteAt > 0 {
		postDeleteAt = time.Unix(0, me.PostDeleteAt*int64(1000*1000)).Format(time.RFC3339)
	}

	postUpdateAt := ""
	if me.PostUpdateAt != me.PostCreateAt {
		postUpdateAt = time.Unix(0, me.PostUpdateAt*int64(1000*1000)).Format(time.RFC3339)
	}

	userType := "user"
	if me.IsBot {
		userType = "bot"
	}

	return []string{
		cleanComplianceStrings(me.TeamName),
		cleanComplianceStrings(me.TeamDisplayName),

		cleanComplianceStrings(me.ChannelName),
		cleanComplianceStrings(me.ChannelDisplayName),
		cleanComplianceStrings(me.ChannelType),

		cleanComplianceStrings(me.UserUsername),
		cleanComplianceStrings(me.UserEmail),
		cleanComplianceStrings(me.UserNickname),
		userType,

		me.PostId,
		time.Unix(0, me.PostCreateAt*int64(1000*1000)).Format(time.RFC3339),
		postUpdateAt,
		postDeleteAt,

		me.PostRootId,
		me.PostParentId,
		me.PostOriginalId,
		cleanComplianceStrings(me.PostMessage),
		me.PostType,
		me.PostProps,
		me.PostHashtags,
		me.PostFileIds,
	}
}
