package requests

import (
	"errors"
)

var ErrCommunityMemberMessagesCommunityID = errors.New("community member messages: invalid id")
var ErrCommunityMemberMessagesMemberPK = errors.New("community member messages: invalid member PK")

type CommunityMemberMessages struct {
	CommunityID     string `json:"communityId"`
	MemberPublicKey string `json:"memberPublicKey"`
}

func (c *CommunityMemberMessages) Validate() error {
	if len(c.CommunityID) == 0 {
		return ErrCommunityMemberMessagesCommunityID
	}

	if len(c.MemberPublicKey) == 0 {
		return ErrCommunityMemberMessagesMemberPK
	}

	return nil
}
