// Discordgo - Discord bindings for Go
// Available at https://github.com/bwmarrin/discordgo

// Copyright 2015-2016 Bruce Marriner <bruce@sqls.net>.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains code related to the Message struct

package discordgo

import (
	"fmt"
	"regexp"
)

// A Message stores all data related to a specific Discord message.
type Message struct {
	ID              string               `json:"id"`
	ChannelID       string               `json:"channel_id"`
	Content         string               `json:"content"`
	Timestamp       string               `json:"timestamp"`
	EditedTimestamp string               `json:"edited_timestamp"`
	MentionRoles    []string             `json:"mention_roles"`
	Tts             bool                 `json:"tts"`
	MentionEveryone bool                 `json:"mention_everyone"`
	Author          *User                `json:"author"`
	Attachments     []*MessageAttachment `json:"attachments"`
	Embeds          []*MessageEmbed      `json:"embeds"`
	Mentions        []*User              `json:"mentions"`
}

// A MessageAttachment stores data for message attachments.
type MessageAttachment struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	ProxyURL string `json:"proxy_url"`
	Filename string `json:"filename"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Size     int    `json:"size"`
}

// An MessageEmbed stores data for message embeds.
type MessageEmbed struct {
	URL         string `json:"url"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnail   *struct {
		URL      string `json:"url"`
		ProxyURL string `json:"proxy_url"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
	} `json:"thumbnail"`
	Provider *struct {
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"provider"`
	Author *struct {
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"author"`
	Video *struct {
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"video"`
}

// ContentWithMentionsReplaced will replace all @<id> mentions with the
// username of the mention.
func (m *Message) ContentWithMentionsReplaced() string {
	if m.Mentions == nil {
		return m.Content
	}
	content := m.Content
	for _, user := range m.Mentions {
		content = regexp.MustCompile(fmt.Sprintf("<@!?(%s)>", user.ID)).ReplaceAllString(content, "@"+user.Username)
	}
	return content
}
