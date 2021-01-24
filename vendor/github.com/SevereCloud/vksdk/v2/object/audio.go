package object // import "github.com/SevereCloud/vksdk/v2/object"

import (
	"fmt"
)

// AudioAudio struct.
type AudioAudio struct {
	AccessKey           string             `json:"access_key"` // Access key for the audio
	ID                  int                `json:"id"`
	OwnerID             int                `json:"owner_id"`
	Artist              string             `json:"artist"`
	Title               string             `json:"title"`
	Duration            int                `json:"duration"`
	Date                int                `json:"date"`
	URL                 string             `json:"url"`
	IsHq                BaseBoolInt        `json:"is_hq"`
	IsExplicit          BaseBoolInt        `json:"is_explicit"`
	StoriesAllowed      BaseBoolInt        `json:"stories_allowed"`
	ShortVideosAllowed  BaseBoolInt        `json:"short_videos_allowed"`
	IsFocusTrack        BaseBoolInt        `json:"is_focus_track"`
	IsLicensed          BaseBoolInt        `json:"is_licensed"`
	StoriesCoverAllowed BaseBoolInt        `json:"stories_cover_allowed"`
	LyricsID            int                `json:"lyrics_id"`
	AlbumID             int                `json:"album_id"`
	GenreID             int                `json:"genre_id"`
	TrackCode           string             `json:"track_code"`
	NoSearch            int                `json:"no_search"`
	MainArtists         []AudioAudioArtist `json:"main_artists"`
	Ads                 AudioAds           `json:"ads"`
	Subtitle            string             `json:"subtitle"`
}

// ToAttachment return attachment format.
func (audio AudioAudio) ToAttachment() string {
	return fmt.Sprintf("audio%d_%d", audio.OwnerID, audio.ID)
}

// AudioAds struct.
type AudioAds struct {
	ContentID      string `json:"content_id"`
	Duration       string `json:"duration"`
	AccountAgeType string `json:"account_age_type"`
	PUID1          string `json:"puid1"`
	PUID22         string `json:"puid22"`
}

// AudioAudioArtist struct.
type AudioAudioArtist struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Domain string `json:"domain"`
}

// AudioAudioUploadResponse struct.
type AudioAudioUploadResponse struct {
	Audio    string `json:"audio"`
	Hash     string `json:"hash"`
	Redirect string `json:"redirect"`
	Server   int    `json:"server"`
}

// AudioLyrics struct.
type AudioLyrics struct {
	LyricsID int    `json:"lyrics_id"`
	Text     string `json:"text"`
}
