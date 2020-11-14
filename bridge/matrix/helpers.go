package bmatrix

import (
	"encoding/json"
	"errors"
	"html"
	"strings"
	"time"

	matrix "github.com/matrix-org/gomatrix"
)

func newMatrixUsername(username string) *matrixUsername {
	mUsername := new(matrixUsername)

	// check if we have a </tag>. if we have, we don't escape HTML. #696
	if htmlTag.MatchString(username) {
		mUsername.formatted = username
		// remove the HTML formatting for beautiful push messages #1188
		mUsername.plain = htmlReplacementTag.ReplaceAllString(username, "")
	} else {
		mUsername.formatted = html.EscapeString(username)
		mUsername.plain = username
	}

	return mUsername
}

// getRoomID retrieves a matching room ID from the channel name.
func (b *Bmatrix) getRoomID(channel string) string {
	b.RLock()
	defer b.RUnlock()
	for ID, name := range b.RoomMap {
		if name == channel {
			return ID
		}
	}

	return ""
}

// interface2Struct marshals and immediately unmarshals an interface.
// Useful for converting map[string]interface{} to a struct.
func interface2Struct(in interface{}, out interface{}) error {
	jsonObj, err := json.Marshal(in)
	if err != nil {
		return err //nolint:wrapcheck
	}

	return json.Unmarshal(jsonObj, out)
}

// getDisplayName retrieves the displayName for mxid, querying the homserver if the mxid is not in the cache.
func (b *Bmatrix) getDisplayName(mxid string) string {
	if b.GetBool("UseUserName") {
		return mxid[1:]
	}

	b.RLock()
	if val, present := b.NicknameMap[mxid]; present {
		b.RUnlock()

		return val.displayName
	}
	b.RUnlock()

	displayName, err := b.mc.GetDisplayName(mxid)
	var httpError *matrix.HTTPError
	if errors.As(err, &httpError) {
		b.Log.Warnf("Couldn't retrieve the display name for %s", mxid)
	}

	if err != nil {
		return b.cacheDisplayName(mxid, mxid[1:])
	}

	return b.cacheDisplayName(mxid, displayName.DisplayName)
}

// cacheDisplayName stores the mapping between a mxid and a display name, to be reused later without performing a query to the homserver.
// Note that old entries are cleaned when this function is called.
func (b *Bmatrix) cacheDisplayName(mxid string, displayName string) string {
	now := time.Now()

	// scan to delete old entries, to stop memory usage from becoming too high with old entries
	toDelete := []string{}
	b.RLock()
	for k, v := range b.NicknameMap {
		if now.Sub(v.lastUpdated) > 10*time.Minute {
			toDelete = append(toDelete, k)
		}
	}
	b.RUnlock()

	b.Lock()
	for _, v := range toDelete {
		delete(b.NicknameMap, v)
	}
	b.NicknameMap[mxid] = NicknameCacheEntry{
		displayName: displayName,
		lastUpdated: now,
	}
	b.Unlock()

	return displayName
}

// handleError converts errors into httpError.
//nolint:exhaustivestruct
func handleError(err error) *httpError {
	var mErr matrix.HTTPError
	if !errors.As(err, &mErr) {
		return &httpError{
			Err: "not a HTTPError",
		}
	}

	var httpErr httpError

	if err := json.Unmarshal(mErr.Contents, &httpErr); err != nil {
		return &httpError{
			Err: "unmarshal failed",
		}
	}

	return &httpErr
}

func (b *Bmatrix) containsAttachment(content map[string]interface{}) bool {
	// Skip empty messages
	if content["msgtype"] == nil {
		return false
	}

	// Only allow image,video or file msgtypes
	if !(content["msgtype"].(string) == "m.image" ||
		content["msgtype"].(string) == "m.video" ||
		content["msgtype"].(string) == "m.file") {
		return false
	}

	return true
}

// getAvatarURL returns the avatar URL of the specified sender.
func (b *Bmatrix) getAvatarURL(sender string) string {
	urlPath := b.mc.BuildURL("profile", sender, "avatar_url")

	s := struct {
		AvatarURL string `json:"avatar_url"`
	}{}

	err := b.mc.MakeRequest("GET", urlPath, nil, &s)
	if err != nil {
		b.Log.Errorf("getAvatarURL failed: %s", err)

		return ""
	}

	url := strings.ReplaceAll(s.AvatarURL, "mxc://", b.GetString("Server")+"/_matrix/media/r0/thumbnail/")
	if url != "" {
		url += "?width=37&height=37&method=crop"
	}

	return url
}
