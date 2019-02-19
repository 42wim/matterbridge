package ddp

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strings"
	"time"
)

// ----------------------------------------------------------------------
// EJSON document interface
// ----------------------------------------------------------------------
// https://github.com/meteor/meteor/blob/devel/packages/ddp/DDP.md#appendix-ejson

// Doc provides hides the complexity of ejson documents.
type Doc struct {
	root interface{}
}

// NewDoc creates a new document from a generic json parsed document.
func NewDoc(in interface{}) *Doc {
	doc := &Doc{in}
	return doc
}

// Map locates a map[string]interface{} - json object - at a path
// or returns nil if not found.
func (d *Doc) Map(path string) map[string]interface{} {
	item := d.Item(path)
	if item != nil {
		switch m := item.(type) {
		case map[string]interface{}:
			return m
		default:
			return nil
		}
	}
	return nil
}

// Array locates an []interface{} - json array - at a path
// or returns nil if not found.
func (d *Doc) Array(path string) []interface{} {
	item := d.Item(path)
	if item != nil {
		switch m := item.(type) {
		case []interface{}:
			return m
		default:
			return nil
		}
	}
	return nil
}

// StringArray locates an []string - json array of strings - at a path
// or returns nil if not found. The string array will contain all string values
// in the array and skip any non-string entries.
func (d *Doc) StringArray(path string) []string {
	item := d.Item(path)
	if item != nil {
		switch m := item.(type) {
		case []interface{}:
			items := []string{}
			for _, item := range m {
				switch val := item.(type) {
				case string:
					items = append(items, val)
				}
			}
			return items
		case []string:
			return m
		default:
			return nil
		}
	}
	return nil
}

// String returns a string value located at the path or an empty string if not found.
func (d *Doc) String(path string) string {
	item := d.Item(path)
	if item != nil {
		switch m := item.(type) {
		case string:
			return m
		default:
			return ""
		}
	}
	return ""
}

// Bool returns a boolean value located at the path or false if not found.
func (d *Doc) Bool(path string) bool {
	item := d.Item(path)
	if item != nil {
		switch m := item.(type) {
		case bool:
			return m
		default:
			return false
		}
	}
	return false
}

// Float returns a float64 value located at the path or zero if not found.
func (d *Doc) Float(path string) float64 {
	item := d.Item(path)
	if item != nil {
		switch m := item.(type) {
		case float64:
			return m
		default:
			return 0
		}
	}
	return 0
}

// Time returns a time value located at the path or nil if not found.
func (d *Doc) Time(path string) time.Time {
	ticks := d.Float(path + ".$date")
	var t time.Time
	if ticks > 0 {
		sec := int64(ticks / 1000)
		t = time.Unix(int64(sec), 0)
	}
	return t
}

// Item locates a "raw" item at the provided path, returning
// the item found or nil if not found.
func (d *Doc) Item(path string) interface{} {
	item := d.root
	steps := strings.Split(path, ".")
	for _, step := range steps {
		// This is an intermediate step - we must be in a map
		switch m := item.(type) {
		case map[string]interface{}:
			item = m[step]
		case Update:
			item = m[step]
		default:
			return nil
		}
	}
	return item
}

// Set a value for a path. Intermediate items are created as necessary.
func (d *Doc) Set(path string, value interface{}) {
	item := d.root
	steps := strings.Split(path, ".")
	last := steps[len(steps)-1]
	steps = steps[:len(steps)-1]
	for _, step := range steps {
		// This is an intermediate step - we must be in a map
		switch m := item.(type) {
		case map[string]interface{}:
			item = m[step]
			if item == nil {
				item = map[string]interface{}{}
				m[step] = item
			}
		default:
			return
		}
	}
	// Item is now the last map so we just set the value
	switch m := item.(type) {
	case map[string]interface{}:
		m[last] = value
	}
}

// Accounts password login support
type Login struct {
	User     *User     `json:"user"`
	Password *Password `json:"password"`
}

func NewEmailLogin(email, pass string) *Login {
	return &Login{User: &User{Email: email}, Password: NewPassword(pass)}
}

func NewUsernameLogin(user, pass string) *Login {
	return &Login{User: &User{Username: user}, Password: NewPassword(pass)}
}

type LoginResume struct {
	Token string `json:"resume"`
}

func NewLoginResume(token string) *LoginResume {
	return &LoginResume{Token: token}
}

type User struct {
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
}

type Password struct {
	Digest    string `json:"digest"`
	Algorithm string `json:"algorithm"`
}

func NewPassword(pass string) *Password {
	sha := sha256.New()
	io.WriteString(sha, pass)
	digest := sha.Sum(nil)
	return &Password{Digest: hex.EncodeToString(digest), Algorithm: "sha-256"}
}
