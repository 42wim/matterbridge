package music

import (
	"time"
)

// Music defines Open Graph Music type
type Music struct {
	Musicians   []string   `json:"musicians,omitempty"`
	Creators    []string   `json:"creators,omitempty"`
	Duration    uint64     `json:"duration,omitempty"`
	ReleaseDate *time.Time `json:"release_date,omitempty"`
	Album       *Album     `json:"album"`
	Songs       []*Song    `json:"songs"`
}

type Album struct {
	URL   string `json:"url,omitempty"`
	Disc  uint64 `json:"disc,omitempty"`
	Track uint64 `json:"track,omitempty"`
}

type Song struct {
	URL   string `json:"url,omitempty"`
	Disc  uint64 `json:"disc,omitempty"`
	Track uint64 `json:"track,omitempty"`
}

func NewMusic() *Music {
	return &Music{Album: &Album{}}
}

func (m *Music) AddSongUrl(v string) {
	if len(m.Songs) == 0 || m.Songs[len(m.Songs)-1].URL != "" {
		m.Songs = append(m.Songs, &Song{})
	}
	m.Songs[len(m.Songs)-1].URL = v
}

func (m *Music) AddSongDisc(v uint64) {
	if len(m.Songs) == 0 {
		m.Songs = append(m.Songs, &Song{})
	}
	m.Songs[len(m.Songs)-1].Disc = v
}

func (m *Music) AddSongTrack(v uint64) {
	if len(m.Songs) == 0 {
		m.Songs = append(m.Songs, &Song{})
	}
	m.Songs[len(m.Songs)-1].Track = v
}
