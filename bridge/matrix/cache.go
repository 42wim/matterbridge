package bmatrix

import (
	"sort"
	"sync"
	"time"

	"maunium.net/go/mautrix/id"
)

type UserInRoomCacheEntry struct {
	displayName *string
	avatarURL   *string
	// for bridged messages that we sent, keep the source URL to know when to upgrade the
	// profile picture (instead of doing it on every message)
	sourceAvatar              *string
	lastUpdated               time.Time
	conflictWithOtherUsername bool
}

type UserCacheEntry struct {
	globalEntry *UserInRoomCacheEntry
	perChannel  map[id.RoomID]UserInRoomCacheEntry
}

type UserInfoCache struct {
	users map[id.UserID]UserCacheEntry
	sync.RWMutex
}

func NewUserInfoCache() *UserInfoCache {
	return &UserInfoCache{
		users:   make(map[id.UserID]UserCacheEntry),
		RWMutex: sync.RWMutex{},
	}
}

// note: the cache is read-locked inside this function
func (c *UserInfoCache) getAttributeFromCache(channelID id.RoomID, mxid id.UserID, attributeIsPresent func(UserInRoomCacheEntry) bool) *UserInRoomCacheEntry {
	c.RLock()
	defer c.RUnlock()

	if user, userPresent := c.users[mxid]; userPresent {
		// try first the name of the user in the room, then globally
		if roomCachedEntry, roomPresent := user.perChannel[channelID]; roomPresent && attributeIsPresent(roomCachedEntry) {
			return &roomCachedEntry
		}

		if user.globalEntry != nil && attributeIsPresent(*user.globalEntry) {
			return user.globalEntry
		}
	}

	return nil
}

// note: cache is locked inside this function
func (b *Bmatrix) cacheEntry(channelID id.RoomID, mxid id.UserID, callback func(UserInRoomCacheEntry) UserInRoomCacheEntry) {
	now := time.Now()

	cache := b.UserCache

	cache.Lock()
	defer cache.Unlock()

	cache.clearObsoleteEntries(mxid)

	var newEntry UserCacheEntry
	if user, userPresent := cache.users[mxid]; userPresent {
		newEntry = user
	} else {
		newEntry = UserCacheEntry{
			globalEntry: nil,
			perChannel:  make(map[id.RoomID]UserInRoomCacheEntry),
		}
	}

	cacheEntry := UserInRoomCacheEntry{
		lastUpdated: now,
	}
	if channelID == "" && newEntry.globalEntry != nil {
		cacheEntry = *newEntry.globalEntry
	} else if channelID != "" {
		if roomCachedEntry, roomPresent := newEntry.perChannel[channelID]; roomPresent {
			cacheEntry = roomCachedEntry
		}
	}

	newCacheEntry := callback(cacheEntry)
	// try first the name of the user in the room, then globally
	if channelID == "" {
		newEntry.globalEntry = &newCacheEntry
	} else {
		// this is a local (room-specific) state, let's cache it as such
		newEntry.perChannel[channelID] = newCacheEntry
	}

	cache.users[mxid] = newEntry
}

// scan to delete old entries, to stop memory usage from becoming high with obsolete entries.
// note: assume the cache is already write-locked
// TODO: should we update the timestamp when the entry is used?
func (c *UserInfoCache) clearObsoleteEntries(mxid id.UserID) {
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

// note: cache is locked inside this function
func (c *UserInfoCache) removeFromCache(roomID id.RoomID, mxid id.UserID) {
	c.Lock()
	defer c.Unlock()

	if user, userPresent := c.users[mxid]; userPresent {
		if _, roomPresent := user.perChannel[roomID]; roomPresent {
			delete(user.perChannel, roomID)
			c.users[mxid] = user
		}
	}
}
