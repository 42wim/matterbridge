package peersyncing

import "errors"

type SyncMessageType int

type SyncMessage struct {
	ID        []byte
	Type      SyncMessageType
	GroupID   []byte
	Payload   []byte
	Timestamp uint64
}

var ErrSyncMessageNotValid = errors.New("sync message not valid")

func (s *SyncMessage) Valid() error {
	valid := len(s.ID) != 0 && s.Type != SyncMessageNoType && len(s.GroupID) != 0 && len(s.Payload) != 0 && s.Timestamp != 0
	if !valid {
		return ErrSyncMessageNotValid
	}
	return nil
}

const (
	SyncMessageNoType SyncMessageType = iota
	SyncMessageCommunityType
	SyncMessageOneToOneType
	SyncMessagePrivateGroup
)
