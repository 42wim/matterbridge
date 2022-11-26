// Auto-generated to Go types using avdl-compiler v1.4.10 (https://github.com/keybase/node-avdl-compiler)
//   Input file: ../client/protocol/avdl/keybase1/notify_team.avdl

package keybase1

import (
	"fmt"
)

type TeamChangeSet struct {
	MembershipChanged bool `codec:"membershipChanged" json:"membershipChanged"`
	KeyRotated        bool `codec:"keyRotated" json:"keyRotated"`
	Renamed           bool `codec:"renamed" json:"renamed"`
	Misc              bool `codec:"misc" json:"misc"`
}

func (o TeamChangeSet) DeepCopy() TeamChangeSet {
	return TeamChangeSet{
		MembershipChanged: o.MembershipChanged,
		KeyRotated:        o.KeyRotated,
		Renamed:           o.Renamed,
		Misc:              o.Misc,
	}
}

type TeamChangedSource int

const (
	TeamChangedSource_SERVER       TeamChangedSource = 0
	TeamChangedSource_LOCAL        TeamChangedSource = 1
	TeamChangedSource_LOCAL_RENAME TeamChangedSource = 2
)

func (o TeamChangedSource) DeepCopy() TeamChangedSource { return o }

var TeamChangedSourceMap = map[string]TeamChangedSource{
	"SERVER":       0,
	"LOCAL":        1,
	"LOCAL_RENAME": 2,
}

var TeamChangedSourceRevMap = map[TeamChangedSource]string{
	0: "SERVER",
	1: "LOCAL",
	2: "LOCAL_RENAME",
}

func (e TeamChangedSource) String() string {
	if v, ok := TeamChangedSourceRevMap[e]; ok {
		return v
	}
	return fmt.Sprintf("%v", int(e))
}

type AvatarUpdateType int

const (
	AvatarUpdateType_NONE AvatarUpdateType = 0
	AvatarUpdateType_USER AvatarUpdateType = 1
	AvatarUpdateType_TEAM AvatarUpdateType = 2
)

func (o AvatarUpdateType) DeepCopy() AvatarUpdateType { return o }

var AvatarUpdateTypeMap = map[string]AvatarUpdateType{
	"NONE": 0,
	"USER": 1,
	"TEAM": 2,
}

var AvatarUpdateTypeRevMap = map[AvatarUpdateType]string{
	0: "NONE",
	1: "USER",
	2: "TEAM",
}

func (e AvatarUpdateType) String() string {
	if v, ok := AvatarUpdateTypeRevMap[e]; ok {
		return v
	}
	return fmt.Sprintf("%v", int(e))
}
