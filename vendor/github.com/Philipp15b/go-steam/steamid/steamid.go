package steamid

import (
	"fmt"
	"strconv"
	"errors"
	"regexp"
	"strings"
)

type ChatInstanceFlag uint64

const (
	Clan     ChatInstanceFlag = 0x100000 >> 1
	Lobby    ChatInstanceFlag = 0x100000 >> 2
	MMSLobby ChatInstanceFlag = 0x100000 >> 3
)

type SteamId uint64

func NewId(id string) (SteamId, error) {
	valid, err := regexp.MatchString(`STEAM_[0-5]:[01]:\d+`, id)
	if err != nil {
		return SteamId(0), err
	}
	if valid {
		id = strings.Replace(id, "STEAM_", "", -1) // remove STEAM_
		splitid := strings.Split(id, ":")          // split 0:1:00000000 into 0 1 00000000
		universe, _ := strconv.ParseInt(splitid[0], 10, 32)
		if universe == 0 { //EUniverse_Invalid
			universe = 1 //EUniverse_Public
		}
		authServer, _ := strconv.ParseUint(splitid[1], 10, 32)
		accId, _ := strconv.ParseUint(splitid[2], 10, 32)
		accountType := int32(1) //EAccountType_Individual
		accountId := (uint32(accId) << 1) | uint32(authServer)
		return NewIdAdv(uint32(accountId), 1, int32(universe), accountType), nil
	} else {
		newid, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			return SteamId(0), err
		}
		return SteamId(newid), nil
	}
	return SteamId(0), errors.New(fmt.Sprintf("Invalid SteamId: %s\n", id))
}

func NewIdAdv(accountId, instance uint32, universe int32, accountType int32) SteamId {
	s := SteamId(0)
	s = s.SetAccountId(accountId)
	s = s.SetAccountInstance(instance)
	s = s.SetAccountUniverse(universe)
	s = s.SetAccountType(accountType)
	return s
}

func (s SteamId) ToUint64() uint64 {
	return uint64(s)
}

func (s SteamId) ToString() string {
	return strconv.FormatUint(uint64(s), 10)
}

func (s SteamId) String() string {
	switch s.GetAccountType() {
	case 0: // EAccountType_Invalid
		fallthrough
	case 1: // EAccountType_Individual
		if s.GetAccountUniverse() <= 1 { // EUniverse_Public
			return fmt.Sprintf("STEAM_0:%d:%d", s.GetAccountId()&1, s.GetAccountId()>>1)
		} else {
			return fmt.Sprintf("STEAM_%d:%d:%d", s.GetAccountUniverse(), s.GetAccountId()&1, s.GetAccountId()>>1)
		}
	default:
		return strconv.FormatUint(uint64(s), 10)
	}
}

func (s SteamId) get(offset uint, mask uint64) uint64 {
	return (uint64(s) >> offset) & mask
}

func (s SteamId) set(offset uint, mask, value uint64) SteamId {
	return SteamId((uint64(s) & ^(mask << offset)) | (value&mask)<<offset)
}

func (s SteamId) GetAccountId() uint32 {
	return uint32(s.get(0, 0xFFFFFFFF))
}

func (s SteamId) SetAccountId(id uint32) SteamId {
	return s.set(0, 0xFFFFFFFF, uint64(id))
}

func (s SteamId) GetAccountInstance() uint32 {
	return uint32(s.get(32, 0xFFFFF))
}

func (s SteamId) SetAccountInstance(value uint32) SteamId {
	return s.set(32, 0xFFFFF, uint64(value))
}

func (s SteamId) GetAccountType() int32 {
	return int32(s.get(52, 0xF))
}

func (s SteamId) SetAccountType(t int32) SteamId {
	return s.set(52, 0xF, uint64(t))
}

func (s SteamId) GetAccountUniverse() int32 {
	return int32(s.get(56, 0xF))
}

func (s SteamId) SetAccountUniverse(universe int32) SteamId {
	return s.set(56, 0xF, uint64(universe))
}

//used to fix the Clan SteamId to a Chat SteamId
func (s SteamId) ClanToChat() SteamId {
	if s.GetAccountType() == int32(7) { //EAccountType_Clan
		s = s.SetAccountInstance(uint32(Clan))
		s = s.SetAccountType(8) //EAccountType_Chat
	}
	return s
}

//used to fix the Chat SteamId to a Clan SteamId
func (s SteamId) ChatToClan() SteamId {
	if s.GetAccountType() == int32(8) { //EAccountType_Chat
		s = s.SetAccountInstance(0)
		s = s.SetAccountType(int32(7)) //EAccountType_Clan
	}
	return s
}
