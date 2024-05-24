package video

import (
	"time"

	"github.com/dyatlov/go-opengraph/opengraph/types/actor"
)

// Video defines Open Graph Video type
type Video struct {
	URL         string         `json:"url"`
	SecureURL   string         `json:"secure_url"`
	Type        string         `json:"type"`
	Width       uint64         `json:"width"`
	Height      uint64         `json:"height"`
	Actors      []*actor.Actor `json:"actors,omitempty"`
	Directors   []string       `json:"directors,omitempty"`
	Writers     []string       `json:"writers,omitempty"`
	Duration    uint64         `json:"duration,omitempty"`
	ReleaseDate *time.Time     `json:"release_date,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
}

func NewVideo() *Video {
	return &Video{}
}

func ensureHasVideo(videos []*Video) []*Video {
	if len(videos) == 0 {
		videos = append(videos, NewVideo())
	}
	return videos
}

func AddURL(videos []*Video, v string) []*Video {
	if len(videos) == 0 || (videos[len(videos)-1].URL != "" && videos[len(videos)-1].URL != v) {
		videos = append(videos, NewVideo())
	}
	videos[len(videos)-1].URL = v
	return videos
}

func AddTag(videos []*Video, v string) []*Video {
	videos = ensureHasVideo(videos)
	videos[len(videos)-1].Tags = append(videos[len(videos)-1].Tags, v)
	return videos
}

func AddDuration(videos []*Video, v uint64) []*Video {
	videos = ensureHasVideo(videos)
	videos[len(videos)-1].Duration = v
	return videos
}

func AddReleaseDate(videos []*Video, v *time.Time) []*Video {
	videos = ensureHasVideo(videos)
	videos[len(videos)-1].ReleaseDate = v
	return videos
}

func AddSecureURL(videos []*Video, v string) []*Video {
	videos = ensureHasVideo(videos)
	videos[len(videos)-1].SecureURL = v
	return videos
}

func AddType(videos []*Video, v string) []*Video {
	videos = ensureHasVideo(videos)
	videos[len(videos)-1].Type = v
	return videos
}

func AddWidth(videos []*Video, v uint64) []*Video {
	videos = ensureHasVideo(videos)
	videos[len(videos)-1].Width = v
	return videos
}

func AddHeight(videos []*Video, v uint64) []*Video {
	videos = ensureHasVideo(videos)
	videos[len(videos)-1].Height = v
	return videos
}
