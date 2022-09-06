package bmatrix

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	matrix "maunium.net/go/mautrix"
	matrixAppService "maunium.net/go/mautrix/appservice"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	"github.com/42wim/matterbridge/bridge/helper"
)

// arbitrary limit to determine when to cleanup nickname cache entries
const MaxNumberOfUsersInCache = 50_000

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
func (b *Bmatrix) getRoomID(channelName string) id.RoomID {
	b.RLock()
	defer b.RUnlock()
	for ID, channel := range b.RoomMap {
		if channelName == channel.name {
			return ID
		}
	}

	return ""
}

func (b *Bmatrix) retrieveGlobalDisplayname(mxid id.UserID) string {
	displayName, err := b.mc.GetDisplayName(mxid)
	var httpError *matrix.HTTPError
	if errors.As(err, &httpError) {
		b.Log.Warnf("Couldn't retrieve the display name for %s", mxid)
	}

	if err != nil {
		return string(mxid)[1:]
	}

	return displayName.DisplayName
}

// getDisplayName retrieves the displayName for mxid, querying the homeserver if the mxid is not in the cache.
func (b *Bmatrix) getDisplayName(channelID id.RoomID, mxid id.UserID) string {
	if b.GetBool("UseUserName") {
		return string(mxid)[1:]
	}

	cachedEntry := b.UserCache.getAttributeFromCache(channelID, mxid, func(e UserInRoomCacheEntry) bool {
		return e.displayName != nil
	})
	if cachedEntry == nil {
		// retrieve the global display name
		return b.cacheDisplayName("", mxid, b.retrieveGlobalDisplayname(mxid))
	}

	if cachedEntry.conflictWithOtherUsername {
		return fmt.Sprintf("%s (%s)", *cachedEntry.displayName, mxid)
	}

	return *cachedEntry.displayName
}

// to prevent username reuse across matrix rooms - or even inside the same room, if a user uses multiple servers -
// identify users with naming conflicts.
// Note: this function locks the cache
func (c *UserInfoCache) detectDisplayNameConflicts(mxid id.UserID, displayName string) bool {
	conflict := false

	c.RLock()
	defer c.RUnlock()

	for mxidIter, NicknameCacheIter := range c.users {
		// skip conflict detection against ourselves, obviously
		if mxidIter == mxid {
			continue
		}

		for channelID, userInChannelCacheEntry := range NicknameCacheIter.perChannel {
			if userInChannelCacheEntry.displayName != nil && *userInChannelCacheEntry.displayName == displayName {
				userInChannelCacheEntry.conflictWithOtherUsername = true
				c.users[mxidIter].perChannel[channelID] = userInChannelCacheEntry
				conflict = true
			}
		}

		if NicknameCacheIter.globalEntry != nil && NicknameCacheIter.globalEntry.displayName != nil && *NicknameCacheIter.globalEntry.displayName == displayName {
			c.users[mxidIter].globalEntry.conflictWithOtherUsername = true
			conflict = true
		}
	}

	return conflict
}

// cacheDisplayName stores the mapping between a mxid and a display name, to be reused
// later without performing a query to the homeserver.
// Note that old entries are cleaned when this function is called.
func (b *Bmatrix) cacheDisplayName(channelID id.RoomID, mxid id.UserID, displayName string) string {
	conflict := b.UserCache.detectDisplayNameConflicts(mxid, displayName)

	b.cacheEntry(channelID, mxid, func(entry UserInRoomCacheEntry) UserInRoomCacheEntry {
		entry.displayName = &displayName
		entry.conflictWithOtherUsername = conflict
		return entry
	})

	return displayName
}

// retrieveGlobalAvatarURL returns the global avatar URL of the specified user.
func (b *Bmatrix) retrieveGlobalAvatarURL(mxid id.UserID) id.ContentURIString {
	url, err := b.mc.GetAvatarURL(mxid)
	if err != nil {
		b.Log.Errorf("Couldn't retrieve the URL of the avatar for MXID %s", mxid)
		return ""
	}

	return id.ContentURIString(url.String())
}

// getAvatarURL retrieves the avatar URL for mxid, querying the homeserver if the mxid is not in the cache.
func (b *Bmatrix) getAvatarURL(channelID id.RoomID, mxid id.UserID) string {
	cachedEntry := b.UserCache.getAttributeFromCache(channelID, mxid, func(e UserInRoomCacheEntry) bool {
		return e.avatarURL != nil
	})
	if cachedEntry == nil {
		// retrieve the global display name
		return b.cacheAvatarURL("", mxid, b.retrieveGlobalAvatarURL(mxid))
	}

	return *cachedEntry.avatarURL
}

// cacheAvatarURL stores the mapping between a mxid and the URL of the user avatar, to be reused
// later without performing a query to the homeserver.
// Note that old entries are cleaned when this function is called.
func (b *Bmatrix) cacheAvatarURL(channelID id.RoomID, mxid id.UserID, avatarURL id.ContentURIString) string {
	contentURI, err := id.ParseContentURI(string(avatarURL))
	if err != nil {
		return ""
	}

	fullURL := b.mc.GetDownloadURL(contentURI)

	b.cacheEntry(channelID, mxid, func(entry UserInRoomCacheEntry) UserInRoomCacheEntry {
		entry.avatarURL = &fullURL
		return entry
	})

	return fullURL
}

// cacheSourceAvatarURL stores the mapping between a virtual user and the *source* URL of the user avatar,
// to be reused later without reuploading the same avatar repeatedly.
// Note that old entries are cleaned when this function is called.
func (b *Bmatrix) cacheSourceAvatarURL(channelID id.RoomID, mxid id.UserID, avatarURL string) {
	b.cacheEntry(channelID, mxid, func(entry UserInRoomCacheEntry) UserInRoomCacheEntry {
		entry.sourceAvatar = &avatarURL
		return entry
	})
}

// getSourceAvatarURL retrieves the avatar URL for mxid, querying the homeserver if the mxid is not in the cache.
func (b *Bmatrix) getSourceAvatarURL(channelID id.RoomID, mxid id.UserID) string {
	cachedEntry := b.UserCache.getAttributeFromCache(channelID, mxid, func(e UserInRoomCacheEntry) bool {
		return e.sourceAvatar != nil
	})
	if cachedEntry == nil {
		return ""
	}

	return *cachedEntry.sourceAvatar
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

type SendMessageEventWrapper struct {
	inner *matrix.Client
}

//nolint:wrapcheck
func (w SendMessageEventWrapper) SendMessageEvent(roomID id.RoomID, eventType event.Type, contentJSON interface{}) (*matrix.RespSendEvent, error) {
	return w.inner.SendMessageEvent(roomID, eventType, contentJSON)
}

func (b *Bmatrix) uploadAvatar(channelID id.RoomID, mxid id.UserID, intent *matrixAppService.IntentAPI, avatarURL string) error {
	cachedURL := b.getSourceAvatarURL(channelID, mxid)

	// do we need to update the avatar for that user
	if avatarURL == cachedURL {
		return nil
	}

	var buf bytes.Buffer
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	req, err := http.NewRequest(http.MethodGet, avatarURL, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return err
	}
	data := buf.Bytes()

	existingData, err := helper.DownloadFile(b.getAvatarURL(channelID, mxid))
	if err == nil && existingData != nil && bytes.Equal(*existingData, data) {
		// the existing avatar is already correct, cache the source URL and return
		b.cacheSourceAvatarURL(channelID, mxid, avatarURL)
		return nil
	}

	//nolint:exhaustruct
	matrixResp, err := b.mc.UploadMedia(matrix.ReqUploadMedia{
		ContentBytes: data,
		ContentType:  resp.Header.Get("Content-Type"),
	})
	if err != nil {
		return err
	}

	err = intent.SetAvatarURL(matrixResp.ContentURI)
	if err != nil {
		return err
	}

	b.cacheSourceAvatarURL(channelID, mxid, avatarURL)

	return nil
}

//nolint:wrapcheck
func (b *Bmatrix) sendMessageEventWithRetries(channel id.RoomID, message event.MessageEventContent, username string, avatarURL string) (string, error) {
	var (
		resp   *matrix.RespSendEvent
		client interface {
			SendMessageEvent(roomID id.RoomID, eventType event.Type, contentJSON interface{}) (resp *matrix.RespSendEvent, err error)
		}
		err error
	)

	b.RLock()
	appservice := b.RoomMap[channel].appService
	b.RUnlock()

	client = SendMessageEventWrapper{inner: b.mc}

	// only try to send messages through the app Service *once* we have received
	// events through it (otherwise we don't really know if the appservice works)
	// Additionally, even if we're receiving messages in that room via the appService listener,
	// let's check that the appservice "covers" that room
	if appservice && b.appService.namespaces.containsRoom(channel) && len(b.appService.namespaces.prefixes) > 0 {
		b.Log.Debugf("Sending with appService")
		// we take the first prefix
		bridgeUserID := fmt.Sprintf("@%s%s:%s", b.appService.namespaces.prefixes[0], id.EncodeUserLocalpart(username), b.appService.appService.HomeserverDomain)
		intent := b.appService.appService.Intent(id.UserID(bridgeUserID))
		// if we can't change the display name it's not great but not the end of the world either, ignore it
		// TODO: do not perform this action on every message, with an in-memory cache or something
		_ = intent.SetDisplayName(username)
		//nolint:errcheck
		go b.uploadAvatar(channel, id.UserID(bridgeUserID), intent, avatarURL)
		client = intent
	} else {
		applyUsernametoMessage(&message, username)
	}

	err = b.retry(func() error {
		resp, err = client.SendMessageEvent(channel, event.EventMessage, message)

		return err
	})
	if err != nil {
		return "", err
	}

	return string(resp.EventID), err
}

func applyUsernametoMessage(newMsg *event.MessageEventContent, username string) {
	matrixUsername := newMatrixUsername(username)

	newMsg.Body = matrixUsername.plain + newMsg.Body
	if newMsg.FormattedBody != "" {
		newMsg.FormattedBody = matrixUsername.formatted + newMsg.FormattedBody
	}
}
