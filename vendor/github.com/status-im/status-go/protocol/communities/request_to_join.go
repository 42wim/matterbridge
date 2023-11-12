package communities

import (
	"fmt"
	"strconv"
	"time"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/protobuf"
)

type RequestToJoinState uint

const (
	RequestToJoinStatePending RequestToJoinState = iota + 1
	RequestToJoinStateDeclined
	RequestToJoinStateAccepted
	RequestToJoinStateCanceled
	RequestToJoinStateAcceptedPending
	RequestToJoinStateDeclinedPending
	RequestToJoinStateAwaitingAddresses
)

type RequestToJoin struct {
	ID               types.HexBytes              `json:"id"`
	PublicKey        string                      `json:"publicKey"`
	Clock            uint64                      `json:"clock"`
	ENSName          string                      `json:"ensName,omitempty"`
	ChatID           string                      `json:"chatId"`
	CommunityID      types.HexBytes              `json:"communityId"`
	State            RequestToJoinState          `json:"state"`
	Our              bool                        `json:"our"`
	Deleted          bool                        `json:"deleted"`
	RevealedAccounts []*protobuf.RevealedAccount `json:"revealedAccounts,omitempty"`
}

func (r *RequestToJoin) CalculateID() {
	r.ID = CalculateRequestID(r.PublicKey, r.CommunityID)
}

func (r *RequestToJoin) ToCommunityRequestToJoinProtobuf() *protobuf.CommunityRequestToJoin {
	return &protobuf.CommunityRequestToJoin{
		Clock:            r.Clock,
		EnsName:          r.ENSName,
		CommunityId:      r.CommunityID,
		RevealedAccounts: r.RevealedAccounts,
	}
}

func (r *RequestToJoin) ToSyncProtobuf() *protobuf.SyncCommunityRequestsToJoin {
	return &protobuf.SyncCommunityRequestsToJoin{
		Id:               r.ID,
		PublicKey:        r.PublicKey,
		Clock:            r.Clock,
		EnsName:          r.ENSName,
		ChatId:           r.ChatID,
		CommunityId:      r.CommunityID,
		State:            uint64(r.State),
		RevealedAccounts: r.RevealedAccounts,
	}
}

func (r *RequestToJoin) InitFromSyncProtobuf(proto *protobuf.SyncCommunityRequestsToJoin) {
	r.ID = proto.Id
	r.PublicKey = proto.PublicKey
	r.Clock = proto.Clock
	r.ENSName = proto.EnsName
	r.ChatID = proto.ChatId
	r.CommunityID = proto.CommunityId
	r.State = RequestToJoinState(proto.State)
	r.RevealedAccounts = proto.RevealedAccounts
}

func (r *RequestToJoin) Empty() bool {
	return len(r.ID)+len(r.PublicKey)+int(r.Clock)+len(r.ENSName)+len(r.ChatID)+len(r.CommunityID)+int(r.State) == 0
}

func (r *RequestToJoin) ShouldRetainDeclined(clock uint64) (bool, error) {
	if r.State != RequestToJoinStateDeclined {
		return false, nil
	}

	declineExpiryClock, err := AddTimeoutToRequestToJoinClock(r.Clock)
	if err != nil {
		return false, err
	}

	return clock < declineExpiryClock, nil
}

func AddTimeoutToRequestToJoinClock(clock uint64) (uint64, error) {
	requestToJoinClock, err := strconv.ParseInt(fmt.Sprint(clock), 10, 64)
	if err != nil {
		return 0, err
	}

	// Adding 7 days to the request clock
	requestTimeOutClock := uint64(time.Unix(requestToJoinClock, 0).AddDate(0, 0, 7).Unix())

	return requestTimeOutClock, nil
}
