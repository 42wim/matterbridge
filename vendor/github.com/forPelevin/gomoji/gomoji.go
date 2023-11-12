package gomoji

import (
	"bytes"
	"errors"
	"strings"

	"github.com/rivo/uniseg"
)

// errors
var (
	ErrStrNotEmoji = errors.New("the string is not emoji")
)

// Emoji is an entity that represents comprehensive emoji info.
type Emoji struct {
	Slug        string `json:"slug"`
	Character   string `json:"character"`
	UnicodeName string `json:"unicode_name"`
	CodePoint   string `json:"code_point"`
	Group       string `json:"group"`
	SubGroup    string `json:"sub_group"`
}

// ContainsEmoji checks whether given string contains emoji or not. It uses local emoji list as provider.
func ContainsEmoji(s string) bool {
	gr := uniseg.NewGraphemes(s)
	for gr.Next() {
		if _, ok := emojiMap[gr.Str()]; ok {
			return true
		}
	}

	return false
}

// AllEmojis gets all emojis from provider.
func AllEmojis() []Emoji {
	return emojiMapToSlice(emojiMap)
}

// RemoveEmojis removes all emojis from the s string and returns a new string.
func RemoveEmojis(s string) string {
	cleanBuf := bytes.Buffer{}

	gr := uniseg.NewGraphemes(s)
	for gr.Next() {
		if _, ok := emojiMap[gr.Str()]; !ok {
			cleanBuf.Write(gr.Bytes())
		}
	}

	return strings.TrimSpace(cleanBuf.String())
}

// GetInfo returns a gomoji.Emoji model representation of provided emoji.
// If the emoji was not found, it returns the gomoji.ErrStrNotEmoji error
func GetInfo(emoji string) (Emoji, error) {
	em, ok := emojiMap[emoji]
	if !ok {
		return Emoji{}, ErrStrNotEmoji
	}

	return em, nil
}

// FindAll finds all emojis in given string. If there are no emojis it returns a nil-slice.
func FindAll(s string) []Emoji {
	var emojis []Emoji

	gr := uniseg.NewGraphemes(s)
	for gr.Next() {
		if em, ok := emojiMap[gr.Str()]; ok {
			emojis = append(emojis, em)
		}
	}

	return emojis
}

func emojiMapToSlice(em map[string]Emoji) []Emoji {
	emojis := make([]Emoji, 0, len(em))
	for _, emoji := range em {
		emojis = append(emojis, emoji)
	}

	return emojis
}
