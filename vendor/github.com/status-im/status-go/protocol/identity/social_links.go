package identity

import (
	"encoding/json"

	"github.com/status-im/status-go/protocol/protobuf"
)

// static links which need to be decorated by the UI clients
const (
	TwitterID      = "__twitter"
	PersonalSiteID = "__personal_site"
	GithubID       = "__github"
	YoutubeID      = "__youtube"
	DiscordID      = "__discord"
	TelegramID     = "__telegram"
)

type SocialLink struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type SocialLinks []*SocialLink

type SocialLinksInfo struct {
	Links   []*SocialLink `json:"links"`
	Removed bool          `json:"removed"`
}

func NewSocialLinks(links []*protobuf.SocialLink) SocialLinks {
	res := SocialLinks{}
	for _, link := range links {
		res = append(res, &SocialLink{Text: link.Text, URL: link.Url})
	}
	return res
}

func (s *SocialLink) ToProtobuf() *protobuf.SocialLink {
	return &protobuf.SocialLink{
		Text: s.Text,
		Url:  s.URL,
	}
}

func (s *SocialLink) Equal(link *SocialLink) bool {
	return s.Text == link.Text && s.URL == link.URL
}

func (s *SocialLinks) ToProtobuf() []*protobuf.SocialLink {
	res := []*protobuf.SocialLink{}
	for _, link := range *s {
		res = append(res, link.ToProtobuf())
	}
	return res
}

func (s *SocialLinks) ToSyncProtobuf(clock uint64) *protobuf.SyncSocialLinks {
	res := &protobuf.SyncSocialLinks{
		Clock: clock,
	}
	for _, link := range *s {
		res.SocialLinks = append(res.SocialLinks, link.ToProtobuf())
	}
	return res
}

// Equal means the same links at the same order
func (s *SocialLinks) Equal(links SocialLinks) bool {
	if len(*s) != len(links) {
		return false
	}
	for i := range *s {
		if !(*s)[i].Equal(links[i]) {
			return false
		}
	}
	return true
}

func (s *SocialLinks) Contains(link *SocialLink) bool {
	if len(*s) == 0 {
		return false
	}
	for _, l := range *s {
		if l.Equal(link) {
			return true
		}
	}
	return false
}

func (s *SocialLinks) Serialize() ([]byte, error) {
	return json.Marshal(*s)
}
