package bmatrix

import (
	"errors"
	"fmt"
	"html"
	"sort"
	"sync"
	"time"

	matrix "maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
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

type NicknameCacheEntry struct {
	displayName               string
	lastUpdated               time.Time
	conflictWithOtherUsername bool
}

type NicknameUserEntry struct {
	globalEntry *NicknameCacheEntry
	perChannel  map[id.RoomID]NicknameCacheEntry
}

type NicknameCache struct {
	users map[id.UserID]NicknameUserEntry
	sync.RWMutex
}

func NewNicknameCache() *NicknameCache {
	return &NicknameCache{
		users:   make(map[id.UserID]NicknameUserEntry),
		RWMutex: sync.RWMutex{},
	}
}

// note: cache is not locked here
func (c *NicknameCache) retrieveDisplaynameFromCache(channelID id.RoomID, mxid id.UserID) string {
	var cachedEntry *NicknameCacheEntry = nil

	c.RLock()
	if user, userPresent := c.users[mxid]; userPresent {
		// try first the name of the user in the room, then globally
		if roomCachedEntry, roomPresent := user.perChannel[channelID]; roomPresent {
			cachedEntry = &roomCachedEntry
		} else if user.globalEntry != nil {
			cachedEntry = user.globalEntry
		}
	}
	c.RUnlock()

	if cachedEntry == nil {
		return ""
	}

	if cachedEntry.conflictWithOtherUsername {
		// TODO: the current behavior is that only users with clashing usernames and *that have
		// spoken since the bridge started* will get their mxids shown, and this doesn't
		// feel right
		return fmt.Sprintf("%s (%s)", cachedEntry.displayName, mxid)
	}

	return cachedEntry.displayName
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

	displayname := b.NicknameCache.retrieveDisplaynameFromCache(channelID, mxid)
	if displayname != "" {
		return displayname
	}

	// retrieve the global display name
	return b.cacheDisplayName("", mxid, b.retrieveGlobalDisplayname(mxid))
}

// scan to delete old entries, to stop memory usage from becoming high with obsolete entries.
// note: assume the cache is already write-locked
// TODO: should we update the timestamp when the entry is used?
func (c *NicknameCache) clearObsoleteEntries(mxid id.UserID) {
	// we have a "off-by-one" to account for when the user being added to the
	// cache already have obsolete cache entries, as we want to keep it because
	// we will be refreshing it in a minute
	if len(c.users) <= MaxNumberOfUsersInCache+1 {
		return
	}

	usersLastTimestamp := make(map[id.UserID]int64, len(c.users))
	// compute the last updated timestamp entry for each user
	for mxidIter, NicknameCacheIter := range c.users {
		userLastTimestamp := time.Unix(0, 0)
		for _, userInChannelCacheEntry := range NicknameCacheIter.perChannel {
			if userInChannelCacheEntry.lastUpdated.After(userLastTimestamp) {
				userLastTimestamp = userInChannelCacheEntry.lastUpdated
			}
		}

		if NicknameCacheIter.globalEntry != nil {
			if NicknameCacheIter.globalEntry.lastUpdated.After(userLastTimestamp) {
				userLastTimestamp = NicknameCacheIter.globalEntry.lastUpdated
			}
		}

		usersLastTimestamp[mxidIter] = userLastTimestamp.UnixNano()
	}

	// get the limit timestamp before which we must clear entries as obsolete
	sortedTimestamps := make([]int64, 0, len(usersLastTimestamp))
	for _, value := range usersLastTimestamp {
		sortedTimestamps = append(sortedTimestamps, value)
	}
	sort.Slice(sortedTimestamps, func(i, j int) bool { return sortedTimestamps[i] < sortedTimestamps[j] })
	limitTimestamp := sortedTimestamps[len(sortedTimestamps)-MaxNumberOfUsersInCache]

	// delete entries older than the limit
	for mxidIter, timestamp := range usersLastTimestamp {
		// do not clear the user that we are adding to the cache
		if timestamp <= limitTimestamp && mxidIter != mxid {
			delete(c.users, mxidIter)
		}
	}
}

// to prevent username reuse across matrix rooms - or even inside the same room, if a user uses multiple servers -
// identify users with naming conflicts
func (c *NicknameCache) detectConflict(mxid id.UserID, displayName string) bool {
	conflict := false

	for mxidIter, NicknameCacheIter := range c.users {
		// skip conflict detection against ourselves, obviously
		if mxidIter == mxid {
			continue
		}

		for channelID, userInChannelCacheEntry := range NicknameCacheIter.perChannel {
			if userInChannelCacheEntry.displayName == displayName {
				userInChannelCacheEntry.conflictWithOtherUsername = true
				c.users[mxidIter].perChannel[channelID] = userInChannelCacheEntry
				conflict = true
			}
		}

		if NicknameCacheIter.globalEntry != nil && NicknameCacheIter.globalEntry.displayName == displayName {
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
	now := time.Now()

	cache := b.NicknameCache

	cache.Lock()
	defer cache.Unlock()

	conflict := cache.detectConflict(mxid, displayName)

	cache.clearObsoleteEntries(mxid)

	var newEntry NicknameUserEntry
	if user, userPresent := cache.users[mxid]; userPresent {
		newEntry = user
	} else {
		newEntry = NicknameUserEntry{
			globalEntry: nil,
			perChannel:  make(map[id.RoomID]NicknameCacheEntry),
		}
	}

	cacheEntry := NicknameCacheEntry{
		displayName:               displayName,
		lastUpdated:               now,
		conflictWithOtherUsername: conflict,
	}

	// this is a local (room-specific) display name, let's cache it as such
	if channelID == "" {
		newEntry.globalEntry = &cacheEntry
	} else {
		globalDisplayName := b.retrieveGlobalDisplayname(mxid)
		// updating the global display name or resetting the room name to the global name
		if globalDisplayName == displayName {
			delete(newEntry.perChannel, channelID)
			newEntry.globalEntry = &cacheEntry
		} else {
			newEntry.perChannel[channelID] = cacheEntry
		}
	}

	cache.users[mxid] = newEntry

	return displayName
}

func (b *Bmatrix) removeDisplayNameFromCache(mxid id.UserID) {
	cache := b.NicknameCache

	cache.Lock()
	defer cache.Unlock()

	delete(cache.users, mxid)
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

type SendMessageEventWrapper struct {
	inner *matrix.Client
}

//nolint: wrapcheck
func (w SendMessageEventWrapper) SendMessageEvent(roomID id.RoomID, eventType event.Type, contentJSON interface{}) (*matrix.RespSendEvent, error) {
	return w.inner.SendMessageEvent(roomID, eventType, contentJSON)
}

//nolint: wrapcheck
func (b *Bmatrix) sendMessageEventWithRetries(channel id.RoomID, message event.MessageEventContent, username string) (string, error) {
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
