package object

// PodcastsItem struct.
type PodcastsItem struct {
	OwnerID int `json:"owner_id"`
}

// PodcastsCategory struct.
type PodcastsCategory struct {
	ID    int         `json:"id"`
	Title string      `json:"title"`
	Cover []BaseImage `json:"cover"`
}

// PodcastsEpisode struct.
type PodcastsEpisode struct {
	ID                  int                 `json:"id"`
	OwnerID             int                 `json:"owner_id"`
	Artist              string              `json:"artist"`
	Title               string              `json:"title"`
	Duration            int                 `json:"duration"`
	Date                int                 `json:"date"`
	URL                 string              `json:"url"`
	LyricsID            int                 `json:"lyrics_id"`
	NoSearch            int                 `json:"no_search"`
	TrackCode           string              `json:"track_code"`
	IsHq                BaseBoolInt         `json:"is_hq"`
	IsFocusTrack        BaseBoolInt         `json:"is_focus_track"`
	IsExplicit          BaseBoolInt         `json:"is_explicit"`
	ShortVideosAllowed  BaseBoolInt         `json:"short_videos_allowed"`
	StoriesAllowed      BaseBoolInt         `json:"stories_allowed"`
	StoriesCoverAllowed BaseBoolInt         `json:"stories_cover_allowed"`
	PodcastInfo         PodcastsPodcastInfo `json:"podcast_info"`
}

// PodcastsPodcastInfo struct.
type PodcastsPodcastInfo struct {
	Cover struct {
		Sizes []BaseImage `json:"cover"`
	}
	Plays       int         `json:"plays"`
	IsFavorite  BaseBoolInt `json:"is_favorite"`
	Description string      `json:"description"`
	Position    int         `json:"position"`
}
