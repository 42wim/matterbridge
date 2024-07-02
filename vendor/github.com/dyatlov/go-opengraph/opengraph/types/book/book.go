package book

import (
	"time"
)

// Book contains Open Graph Book structure
type Book struct {
	ISBN        string     `json:"isbn"`
	ReleaseDate *time.Time `json:"release_date"`
	Tags        []string   `json:"tags"`
	Authors     []string   `json:"authors"`
}
