package opengraph

import (
	"encoding/json"
	"io"
	"strconv"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/dyatlov/go-opengraph/opengraph/types/actor"
	"github.com/dyatlov/go-opengraph/opengraph/types/article"
	"github.com/dyatlov/go-opengraph/opengraph/types/audio"
	"github.com/dyatlov/go-opengraph/opengraph/types/book"
	"github.com/dyatlov/go-opengraph/opengraph/types/image"
	"github.com/dyatlov/go-opengraph/opengraph/types/music"
	"github.com/dyatlov/go-opengraph/opengraph/types/profile"
	"github.com/dyatlov/go-opengraph/opengraph/types/video"
)

// OpenGraph contains facebook og data
type OpenGraph struct {
	isArticle        bool
	isBook           bool
	isProfile        bool
	Type             string           `json:"type"`
	URL              string           `json:"url"`
	Title            string           `json:"title"`
	Description      string           `json:"description"`
	Determiner       string           `json:"determiner"`
	SiteName         string           `json:"site_name"`
	Locale           string           `json:"locale"`
	LocalesAlternate []string         `json:"locales_alternate"`
	Images           []*image.Image   `json:"images"`
	Audios           []*audio.Audio   `json:"audios"`
	Videos           []*video.Video   `json:"videos"`
	Article          *article.Article `json:"article,omitempty"`
	Book             *book.Book       `json:"book,omitempty"`
	Profile          *profile.Profile `json:"profile,omitempty"`
	Music            *music.Music     `json:"music,omitempty"`
}

// NewOpenGraph returns new instance of Open Graph structure
func NewOpenGraph() *OpenGraph {
	return &OpenGraph{}
}

// ToJSON a simple wrapper around json.Marshal
func (og *OpenGraph) ToJSON() ([]byte, error) {
	return json.Marshal(og)
}

// String return json representation of structure, or error string
func (og *OpenGraph) String() string {
	data, err := og.ToJSON()

	if err != nil {
		return err.Error()
	}

	return string(data[:])
}

// ProcessHTML parses given html from Reader interface and fills up OpenGraph structure
func (og *OpenGraph) ProcessHTML(buffer io.Reader) error {
	z := html.NewTokenizer(buffer)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				return nil
			}
			return z.Err()
		case html.StartTagToken, html.SelfClosingTagToken, html.EndTagToken:
			name, hasAttr := z.TagName()
			if atom.Lookup(name) != atom.Meta || !hasAttr {
				continue
			}
			m := make(map[string]string)
			var key, val []byte
			for hasAttr {
				key, val, hasAttr = z.TagAttr()
				m[atom.String(key)] = string(val)
			}
			og.ProcessMeta(m)
		}
	}
}

func (og *OpenGraph) ensureHasVideo() {
	if len(og.Videos) > 0 {
		return
	}
	og.Videos = append(og.Videos, video.NewVideo())
}

func (og *OpenGraph) ensureHasMusic() {
	if og.Music == nil {
		og.Music = music.NewMusic()
	}
}

// ProcessMeta processes meta attributes and adds them to Open Graph structure if they are suitable for that
func (og *OpenGraph) ProcessMeta(metaAttrs map[string]string) {
	switch metaAttrs["property"] {
	case "og:description":
		og.Description = metaAttrs["content"]
	case "og:type":
		og.Type = metaAttrs["content"]
		switch og.Type {
		case "article":
			og.isArticle = true
		case "book":
			og.isBook = true
		case "profile":
			og.isProfile = true
		}
	case "og:title":
		og.Title = metaAttrs["content"]
	case "og:url":
		og.URL = metaAttrs["content"]
	case "og:determiner":
		og.Determiner = metaAttrs["content"]
	case "og:site_name":
		og.SiteName = metaAttrs["content"]
	case "og:locale":
		og.Locale = metaAttrs["content"]
	case "og:locale:alternate":
		og.LocalesAlternate = append(og.LocalesAlternate, metaAttrs["content"])
	case "og:audio":
		og.Audios = audio.AddUrl(og.Audios, metaAttrs["content"])
	case "og:audio:secure_url":
		og.Audios = audio.AddSecureUrl(og.Audios, metaAttrs["content"])
	case "og:audio:type":
		og.Audios = audio.AddType(og.Audios, metaAttrs["content"])
	case "og:image":
		og.Images = image.AddURL(og.Images, metaAttrs["content"])
	case "og:image:url":
		og.Images = image.AddURL(og.Images, metaAttrs["content"])
	case "og:image:secure_url":
		og.Images = image.AddSecureURL(og.Images, metaAttrs["content"])
	case "og:image:type":
		og.Images = image.AddType(og.Images, metaAttrs["content"])
	case "og:image:width":
		w, err := strconv.ParseUint(metaAttrs["content"], 10, 64)
		if err == nil {
			og.Images = image.AddWidth(og.Images, w)
		}
	case "og:image:height":
		h, err := strconv.ParseUint(metaAttrs["content"], 10, 64)
		if err == nil {
			og.Images = image.AddHeight(og.Images, h)
		}
	case "og:video":
		og.Videos = video.AddURL(og.Videos, metaAttrs["content"])
	case "og:video:tag":
		og.Videos = video.AddTag(og.Videos, metaAttrs["content"])
	case "og:video:duration":
		if i, err := strconv.ParseUint(metaAttrs["content"], 10, 64); err == nil {
			og.Videos = video.AddDuration(og.Videos, i)
		}
	case "og:video:release_date":
		if t, err := time.Parse(time.RFC3339, metaAttrs["content"]); err == nil {
			og.Videos = video.AddReleaseDate(og.Videos, &t)
		}
	case "og:video:url":
		og.Videos = video.AddURL(og.Videos, metaAttrs["content"])
	case "og:video:secure_url":
		og.Videos = video.AddSecureURL(og.Videos, metaAttrs["content"])
	case "og:video:type":
		og.Videos = video.AddTag(og.Videos, metaAttrs["content"])
	case "og:video:width":
		w, err := strconv.ParseUint(metaAttrs["content"], 10, 64)
		if err == nil {
			og.Videos = video.AddWidth(og.Videos, w)
		}
	case "og:video:height":
		h, err := strconv.ParseUint(metaAttrs["content"], 10, 64)
		if err == nil {
			og.Videos = video.AddHeight(og.Videos, h)
		}
	case "og:video:actor":
		og.ensureHasVideo()
		og.Videos[len(og.Videos)-1].Actors = actor.AddProfile(og.Videos[len(og.Videos)-1].Actors, metaAttrs["content"])
	case "og:video:actor:role":
		og.ensureHasVideo()
		og.Videos[len(og.Videos)-1].Actors = actor.AddRole(og.Videos[len(og.Videos)-1].Actors, metaAttrs["content"])
	case "og:video:director":
		og.ensureHasVideo()
		og.Videos[len(og.Videos)-1].Directors = append(og.Videos[len(og.Videos)-1].Directors, metaAttrs["content"])
	case "og:video:writer":
		og.ensureHasVideo()
		og.Videos[len(og.Videos)-1].Writers = append(og.Videos[len(og.Videos)-1].Writers, metaAttrs["content"])
	case "og:music:duration":
		og.ensureHasMusic()
		if i, err := strconv.ParseUint(metaAttrs["content"], 10, 64); err == nil {
			og.Music.Duration = i
		}
	case "og:music:release_date":
		og.ensureHasMusic()
		if t, err := time.Parse(time.RFC3339, metaAttrs["content"]); err == nil {
			og.Music.ReleaseDate = &t
		}
	case "og:music:album":
		og.ensureHasMusic()
		og.Music.Album.URL = metaAttrs["content"]
	case "og:music:album:disc":
		og.ensureHasMusic()
		if i, err := strconv.ParseUint(metaAttrs["content"], 10, 64); err == nil {
			og.Music.Album.Disc = i
		}
	case "og:music:album:track":
		og.ensureHasMusic()
		if i, err := strconv.ParseUint(metaAttrs["content"], 10, 64); err == nil {
			og.Music.Album.Track = i
		}
	case "og:music:musician":
		og.ensureHasMusic()
		og.Music.Musicians = append(og.Music.Musicians, metaAttrs["content"])
	case "og:music:creator":
		og.ensureHasMusic()
		og.Music.Creators = append(og.Music.Creators, metaAttrs["content"])
	case "og:music:song":
		og.ensureHasMusic()
		og.Music.AddSongUrl(metaAttrs["content"])
	case "og:music:disc":
		og.ensureHasMusic()
		if i, err := strconv.ParseUint(metaAttrs["content"], 10, 64); err == nil {
			og.Music.AddSongDisc(i)
		}
	case "og:music:track":
		og.ensureHasMusic()
		if i, err := strconv.ParseUint(metaAttrs["content"], 10, 64); err == nil {
			og.Music.AddSongTrack(i)
		}
	default:
		if og.isArticle {
			og.processArticleMeta(metaAttrs)
		} else if og.isBook {
			og.processBookMeta(metaAttrs)
		} else if og.isProfile {
			og.processProfileMeta(metaAttrs)
		}
	}
}

func (og *OpenGraph) processArticleMeta(metaAttrs map[string]string) {
	if og.Article == nil {
		og.Article = &article.Article{}
	}
	switch metaAttrs["property"] {
	case "og:article:published_time":
		t, err := time.Parse(time.RFC3339, metaAttrs["content"])
		if err == nil {
			og.Article.PublishedTime = &t
		}
	case "og:article:modified_time":
		t, err := time.Parse(time.RFC3339, metaAttrs["content"])
		if err == nil {
			og.Article.ModifiedTime = &t
		}
	case "og:article:expiration_time":
		t, err := time.Parse(time.RFC3339, metaAttrs["content"])
		if err == nil {
			og.Article.ExpirationTime = &t
		}
	case "og:article:section":
		og.Article.Section = metaAttrs["content"]
	case "og:article:tag":
		og.Article.Tags = append(og.Article.Tags, metaAttrs["content"])
	case "og:article:author":
		og.Article.Authors = append(og.Article.Authors, metaAttrs["content"])
	}
}

func (og *OpenGraph) processBookMeta(metaAttrs map[string]string) {
	if og.Book == nil {
		og.Book = &book.Book{}
	}
	switch metaAttrs["property"] {
	case "og:book:release_date":
		t, err := time.Parse(time.RFC3339, metaAttrs["content"])
		if err == nil {
			og.Book.ReleaseDate = &t
		}
	case "og:book:isbn":
		og.Book.ISBN = metaAttrs["content"]
	case "og:book:tag":
		og.Book.Tags = append(og.Book.Tags, metaAttrs["content"])
	case "og:book:author":
		og.Book.Authors = append(og.Book.Authors, metaAttrs["content"])
	}
}

func (og *OpenGraph) processProfileMeta(metaAttrs map[string]string) {
	if og.Profile == nil {
		og.Profile = &profile.Profile{}
	}
	switch metaAttrs["property"] {
	case "og:profile:first_name":
		og.Profile.FirstName = metaAttrs["content"]
	case "og:profile:last_name":
		og.Profile.LastName = metaAttrs["content"]
	case "og:profile:username":
		og.Profile.Username = metaAttrs["content"]
	case "og:profile:gender":
		og.Profile.Gender = metaAttrs["content"]
	}
}
