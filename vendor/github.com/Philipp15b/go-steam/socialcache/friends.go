package socialcache

import (
	"errors"
	"sync"

	. "github.com/Philipp15b/go-steam/protocol/steamlang"
	. "github.com/Philipp15b/go-steam/steamid"
)

// Friends list is a thread safe map
// They can be iterated over like so:
// 	for id, friend := range client.Social.Friends.GetCopy() {
// 		log.Println(id, friend.Name)
// 	}
type FriendsList struct {
	mutex sync.RWMutex
	byId  map[SteamId]*Friend
}

// Returns a new friends list
func NewFriendsList() *FriendsList {
	return &FriendsList{byId: make(map[SteamId]*Friend)}
}

// Adds a friend to the friend list
func (list *FriendsList) Add(friend Friend) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	_, exists := list.byId[friend.SteamId]
	if !exists { //make sure this doesnt already exist
		list.byId[friend.SteamId] = &friend
	}
}

// Removes a friend from the friend list
func (list *FriendsList) Remove(id SteamId) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	delete(list.byId, id)
}

// Returns a copy of the friends map
func (list *FriendsList) GetCopy() map[SteamId]Friend {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	flist := make(map[SteamId]Friend)
	for key, friend := range list.byId {
		flist[key] = *friend
	}
	return flist
}

// Returns a copy of the friend of a given SteamId
func (list *FriendsList) ById(id SteamId) (Friend, error) {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	if val, ok := list.byId[id]; ok {
		return *val, nil
	}
	return Friend{}, errors.New("Friend not found")
}

// Returns the number of friends
func (list *FriendsList) Count() int {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return len(list.byId)
}

//Setter methods
func (list *FriendsList) SetName(id SteamId, name string) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if val, ok := list.byId[id]; ok {
		val.Name = name
	}
}

func (list *FriendsList) SetAvatar(id SteamId, hash []byte) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if val, ok := list.byId[id]; ok {
		val.Avatar = hash
	}
}

func (list *FriendsList) SetRelationship(id SteamId, relationship EFriendRelationship) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if val, ok := list.byId[id]; ok {
		val.Relationship = relationship
	}
}

func (list *FriendsList) SetPersonaState(id SteamId, state EPersonaState) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if val, ok := list.byId[id]; ok {
		val.PersonaState = state
	}
}

func (list *FriendsList) SetPersonaStateFlags(id SteamId, flags EPersonaStateFlag) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if val, ok := list.byId[id]; ok {
		val.PersonaStateFlags = flags
	}
}

func (list *FriendsList) SetGameAppId(id SteamId, gameappid uint32) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if val, ok := list.byId[id]; ok {
		val.GameAppId = gameappid
	}
}

func (list *FriendsList) SetGameId(id SteamId, gameid uint64) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if val, ok := list.byId[id]; ok {
		val.GameId = gameid
	}
}

func (list *FriendsList) SetGameName(id SteamId, name string) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if val, ok := list.byId[id]; ok {
		val.GameName = name
	}
}

// A Friend
type Friend struct {
	SteamId           SteamId `json:",string"`
	Name              string
	Avatar            []byte
	Relationship      EFriendRelationship
	PersonaState      EPersonaState
	PersonaStateFlags EPersonaStateFlag
	GameAppId         uint32
	GameId            uint64 `json:",string"`
	GameName          string
}
