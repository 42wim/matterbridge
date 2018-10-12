package bslack

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nlopes/slack"
)

func (b *Bslack) userName(id string) string {
	for _, u := range b.Users {
		if u.ID == id {
			if u.Profile.DisplayName != "" {
				return u.Profile.DisplayName
			}
			return u.Name
		}
	}
	return ""
}

func (b *Bslack) getAvatar(userid string) string {
	var avatar string
	if b.Users != nil {
		for _, u := range b.Users {
			if userid == u.ID {
				return u.Profile.Image48
			}
		}
	}
	return avatar
}

/*
func (b *Bslack) getChannelByName(name string) (*slack.Channel, error) {
	if b.channels == nil {
		return nil, fmt.Errorf("%s: channel %s not found (no channels found)", b.Account, name)
	}
	for _, channel := range b.channels {
		if channel.Name == name {
			return &channel, nil
		}
	}
	return nil, fmt.Errorf("%s: channel %s not found", b.Account, name)
}
*/

func (b *Bslack) getChannelByID(ID string) (*slack.Channel, error) {
	if b.channels == nil {
		return nil, fmt.Errorf("%s: channel %s not found (no channels found)", b.Account, ID)
	}
	for _, channel := range b.channels {
		if channel.ID == ID {
			return &channel, nil
		}
	}
	return nil, fmt.Errorf("%s: channel %s not found", b.Account, ID)
}

func (b *Bslack) getChannelID(name string) string {
	idcheck := strings.Split(name, "ID:")
	if len(idcheck) > 1 {
		return idcheck[1]
	}
	for _, channel := range b.channels {
		if channel.Name == name {
			return channel.ID
		}
	}
	return ""
}

// @see https://api.slack.com/docs/message-formatting#linking_to_channels_and_users
func (b *Bslack) replaceMention(text string) string {
	results := regexp.MustCompile(`<@([a-zA-Z0-9]+)>`).FindAllStringSubmatch(text, -1)
	for _, r := range results {
		text = strings.Replace(text, "<@"+r[1]+">", "@"+b.userName(r[1]), -1)
	}
	return text
}

// @see https://api.slack.com/docs/message-formatting#linking_to_channels_and_users
func (b *Bslack) replaceChannel(text string) string {
	results := regexp.MustCompile(`<#[a-zA-Z0-9]+\|(.+?)>`).FindAllStringSubmatch(text, -1)
	for _, r := range results {
		text = strings.Replace(text, r[0], "#"+r[1], -1)
	}
	return text
}

// @see https://api.slack.com/docs/message-formatting#variables
func (b *Bslack) replaceVariable(text string) string {
	results := regexp.MustCompile(`<!((?:subteam\^)?[a-zA-Z0-9]+)(?:\|@?(.+?))?>`).FindAllStringSubmatch(text, -1)
	for _, r := range results {
		if r[2] != "" {
			text = strings.Replace(text, r[0], "@"+r[2], -1)
		} else {
			text = strings.Replace(text, r[0], "@"+r[1], -1)
		}
	}
	return text
}

// @see https://api.slack.com/docs/message-formatting#linking_to_urls
func (b *Bslack) replaceURL(text string) string {
	results := regexp.MustCompile(`<(.*?)(\|.*?)?>`).FindAllStringSubmatch(text, -1)
	for _, r := range results {
		if len(strings.TrimSpace(r[2])) == 1 { // A display text separator was found, but the text was blank
			text = strings.Replace(text, r[0], "", -1)
		} else {
			text = strings.Replace(text, r[0], r[1], -1)
		}
	}
	return text
}
