package socialcache

import (
	"errors"
	"sync"

	. "github.com/Philipp15b/go-steam/protocol/steamlang"
	. "github.com/Philipp15b/go-steam/steamid"
)

// Groups list is a thread safe map
// They can be iterated over like so:
// 	for id, group := range client.Social.Groups.GetCopy() {
// 		log.Println(id, group.Name)
// 	}
type GroupsList struct {
	mutex sync.RWMutex
	byId  map[SteamId]*Group
}

// Returns a new groups list
func NewGroupsList() *GroupsList {
	return &GroupsList{byId: make(map[SteamId]*Group)}
}

// Adds a group to the group list
func (list *GroupsList) Add(group Group) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	_, exists := list.byId[group.SteamId]
	if !exists { //make sure this doesnt already exist
		list.byId[group.SteamId] = &group
	}
}

// Removes a group from the group list
func (list *GroupsList) Remove(id SteamId) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	delete(list.byId, id)
}

// Returns a copy of the groups map
func (list *GroupsList) GetCopy() map[SteamId]Group {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	glist := make(map[SteamId]Group)
	for key, group := range list.byId {
		glist[key] = *group
	}
	return glist
}

// Returns a copy of the group of a given SteamId
func (list *GroupsList) ById(id SteamId) (Group, error) {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	id = id.ChatToClan()
	if val, ok := list.byId[id]; ok {
		return *val, nil
	}
	return Group{}, errors.New("Group not found")
}

// Returns the number of groups
func (list *GroupsList) Count() int {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return len(list.byId)
}

//Setter methods
func (list *GroupsList) SetName(id SteamId, name string) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	id = id.ChatToClan()
	if val, ok := list.byId[id]; ok {
		val.Name = name
	}
}

func (list *GroupsList) SetAvatar(id SteamId, hash []byte) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	id = id.ChatToClan()
	if val, ok := list.byId[id]; ok {
		val.Avatar = hash
	}
}

func (list *GroupsList) SetRelationship(id SteamId, relationship EClanRelationship) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	id = id.ChatToClan()
	if val, ok := list.byId[id]; ok {
		val.Relationship = relationship
	}
}

func (list *GroupsList) SetMemberTotalCount(id SteamId, count uint32) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	id = id.ChatToClan()
	if val, ok := list.byId[id]; ok {
		val.MemberTotalCount = count
	}
}

func (list *GroupsList) SetMemberOnlineCount(id SteamId, count uint32) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	id = id.ChatToClan()
	if val, ok := list.byId[id]; ok {
		val.MemberOnlineCount = count
	}
}

func (list *GroupsList) SetMemberChattingCount(id SteamId, count uint32) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	id = id.ChatToClan()
	if val, ok := list.byId[id]; ok {
		val.MemberChattingCount = count
	}
}

func (list *GroupsList) SetMemberInGameCount(id SteamId, count uint32) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	id = id.ChatToClan()
	if val, ok := list.byId[id]; ok {
		val.MemberInGameCount = count
	}
}

// A Group
type Group struct {
	SteamId             SteamId `json:",string"`
	Name                string
	Avatar              []byte
	Relationship        EClanRelationship
	MemberTotalCount    uint32
	MemberOnlineCount   uint32
	MemberChattingCount uint32
	MemberInGameCount   uint32
}
