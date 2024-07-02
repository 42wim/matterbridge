package article

import (
	"time"
)

// Article contain Open Graph Article structure
type Article struct {
	PublishedTime  *time.Time `json:"published_time"`
	ModifiedTime   *time.Time `json:"modified_time"`
	ExpirationTime *time.Time `json:"expiration_time"`
	Section        string     `json:"section"`
	Tags           []string   `json:"tags"`
	Authors        []string   `json:"authors"`
}
