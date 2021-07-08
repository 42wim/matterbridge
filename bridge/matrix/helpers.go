package bmatrix

import (
	"errors"
	"fmt"
	"html"
	"time"

	matrix "maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
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
func (b *Bmatrix) getRoomID(channel string) id.RoomID {
	b.RLock()
	defer b.RUnlock()
	for ID, name := range b.RoomMap {
		if name == channel {
			return ID
		}
	}

	return ""
}

// getDisplayName retrieves the displayName for mxid, querying the homeserver if the mxid is not in the cache.
func (b *Bmatrix) getDisplayName(channelID id.RoomID, mxid id.UserID) string {
	if b.GetBool("UseUserName") {
		return string(mxid)[1:]
	}

	b.RLock()
	if channel, channelPresent := b.NicknameMap[channelID]; channelPresent {
		if val, present := channel[mxid]; present {
			b.RUnlock()

			return val.displayName
		}
	}
	b.RUnlock()

	displayName, err := b.mc.GetDisplayName(mxid)
	var httpError *matrix.HTTPError
	if errors.As(err, &httpError) {
		b.Log.Warnf("Couldn't retrieve the display name for %s", mxid)
	}

	if err != nil {
		return b.cacheDisplayName(channelID, mxid, string(mxid)[1:])
	}

	return b.cacheDisplayName(channelID, mxid, displayName.DisplayName)
}

// cacheDisplayName stores the mapping between a mxid and a display name, to be reused later without performing a query to the homeserver.
// Note that old entries are cleaned when this function is called.
func (b *Bmatrix) cacheDisplayName(channelID id.RoomID, mxid id.UserID, displayName string) string {
	now := time.Now()

	// scan to delete old entries, to stop memory usage from becoming high with obsolete entries.
	// In addition, we detect if another user have the same username, and if so, we append their mxids to their usernames to differentiate them.
	toDelete := map[id.RoomID]id.UserID{}
	conflict := false

	b.Lock()
	for channelIDIter, channelEntriesIter := range b.NicknameMap {
		for mxidIter, NicknameCacheIter := range channelEntriesIter {
			// to prevent username reuse across matrix rooms - or even inside the same room, if a user uses multiple servers -
			// append the mxid to the username when there is a conflict
			if NicknameCacheIter.displayName == displayName {
				conflict = true
				// TODO: it would be nice to be able to rename previous messages from this user.
				// The current behavior is that only users with clashing usernames and *that have spoken since the bridge last started* will get their mxids shown, and I don't know if that's the expected behavior.
				NicknameCacheIter.displayName = fmt.Sprintf("%s (%s)", displayName, mxidIter)
				b.NicknameMap[channelIDIter][mxidIter] = NicknameCacheIter
			}

			if now.Sub(NicknameCacheIter.lastUpdated) > 10*time.Minute {
				toDelete[channelIDIter] = mxidIter
			}
		}
	}

	for channelIDIter, mxidIter := range toDelete {
		delete(b.NicknameMap[channelIDIter], mxidIter)
	}

	if conflict {
		displayName = fmt.Sprintf("%s (%s)", displayName, mxid)
	}

	if _, channelPresent := b.NicknameMap[channelID]; !channelPresent {
		b.NicknameMap[channelID] = make(map[id.UserID]NicknameCacheEntry)
	}

	b.NicknameMap[channelID][mxid] = NicknameCacheEntry{
		displayName: displayName,
		lastUpdated: now,
	}
	b.Unlock()

	return displayName
}

// getAvatarURL returns the avatar URL of the specified sender.
func (b *Bmatrix) getAvatarURL(sender id.UserID) string {
	url, err := b.mc.GetAvatarURL(sender)
	if err != nil {
		b.Log.Errorf("Couldn't retrieve the URL of the avatar for MXID %s", sender)
		return ""
	}

	return url.String()
}

// handleRatelimit handles the ratelimit errors and return if we're ratelimited and the amount of time to sleep
func (b *Bmatrix) handleRatelimit(err error) (time.Duration, bool) {
	var mErr matrix.HTTPError
	if !errors.As(err, &mErr) {
		b.Log.Errorf("Received a non-HTTPError, don't know what to make of it:\n%#v", err)
		return 0, false
	}

	if mErr.RespError.ErrCode != "M_LIMIT_EXCEEDED" {
		return 0, false
	}

	b.Log.Debugf("ratelimited: %s", mErr.RespError.Err)

	// fallback to a one-second delay
	retryDelayMs := 1000

	if retryDelayString, present := mErr.RespError.ExtraData["retry_after_ms"]; present {
		if retryDelayInt, correct := retryDelayString.(int); correct && retryDelayInt > retryDelayMs {
			retryDelayMs = retryDelayInt
		}
	}

	b.Log.Infof("getting ratelimited by matrix, sleeping approx %d seconds before retrying", retryDelayMs/1000)

	return time.Duration(retryDelayMs) * time.Millisecond, true
}

// retry function will check if we're ratelimited and retries again when backoff time expired
// returns original error if not 429 ratelimit
func (b *Bmatrix) retry(f func() error) error {
	b.rateMutex.Lock()
	defer b.rateMutex.Unlock()

	for {
		if err := f(); err != nil {
			if backoff, ok := b.handleRatelimit(err); ok {
				time.Sleep(backoff)
			} else {
				return err
			}
		} else {
			return nil
		}
	}
}
