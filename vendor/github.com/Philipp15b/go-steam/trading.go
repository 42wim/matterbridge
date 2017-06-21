package steam

import (
	. "github.com/Philipp15b/go-steam/protocol"
	. "github.com/Philipp15b/go-steam/protocol/protobuf"
	. "github.com/Philipp15b/go-steam/protocol/steamlang"
	. "github.com/Philipp15b/go-steam/steamid"
	"github.com/golang/protobuf/proto"
)

// Provides access to the Steam client's part of Steam Trading, that is bootstrapping
// the trade.
// The trade itself is not handled by the Steam client itself, but it's a part of
// the Steam website.
//
// You'll receive a TradeProposedEvent when a friend proposes a trade. You can accept it with
// the RespondRequest method. You can request a trade yourself with RequestTrade.
type Trading struct {
	client *Client
}

type TradeRequestId uint32

func (t *Trading) HandlePacket(packet *Packet) {
	switch packet.EMsg {
	case EMsg_EconTrading_InitiateTradeProposed:
		msg := new(CMsgTrading_InitiateTradeRequest)
		packet.ReadProtoMsg(msg)
		t.client.Emit(&TradeProposedEvent{
			RequestId: TradeRequestId(msg.GetTradeRequestId()),
			Other:     SteamId(msg.GetOtherSteamid()),
		})
	case EMsg_EconTrading_InitiateTradeResult:
		msg := new(CMsgTrading_InitiateTradeResponse)
		packet.ReadProtoMsg(msg)
		t.client.Emit(&TradeResultEvent{
			RequestId: TradeRequestId(msg.GetTradeRequestId()),
			Response:  EEconTradeResponse(msg.GetResponse()),
			Other:     SteamId(msg.GetOtherSteamid()),

			NumDaysSteamGuardRequired:            msg.GetSteamguardRequiredDays(),
			NumDaysNewDeviceCooldown:             msg.GetNewDeviceCooldownDays(),
			DefaultNumDaysPasswordResetProbation: msg.GetDefaultPasswordResetProbationDays(),
			NumDaysPasswordResetProbation:        msg.GetPasswordResetProbationDays(),
		})
	case EMsg_EconTrading_StartSession:
		msg := new(CMsgTrading_StartSession)
		packet.ReadProtoMsg(msg)
		t.client.Emit(&TradeSessionStartEvent{
			Other: SteamId(msg.GetOtherSteamid()),
		})
	}
}

// Requests a trade. You'll receive a TradeResultEvent if the request fails or
// if the friend accepted the trade.
func (t *Trading) RequestTrade(other SteamId) {
	t.client.Write(NewClientMsgProtobuf(EMsg_EconTrading_InitiateTradeRequest, &CMsgTrading_InitiateTradeRequest{
		OtherSteamid: proto.Uint64(uint64(other)),
	}))
}

// Responds to a TradeProposedEvent.
func (t *Trading) RespondRequest(requestId TradeRequestId, accept bool) {
	var resp uint32
	if accept {
		resp = 0
	} else {
		resp = 1
	}

	t.client.Write(NewClientMsgProtobuf(EMsg_EconTrading_InitiateTradeResponse, &CMsgTrading_InitiateTradeResponse{
		TradeRequestId: proto.Uint32(uint32(requestId)),
		Response:       proto.Uint32(resp),
	}))
}

// This cancels a request made with RequestTrade.
func (t *Trading) CancelRequest(other SteamId) {
	t.client.Write(NewClientMsgProtobuf(EMsg_EconTrading_CancelTradeRequest, &CMsgTrading_CancelTradeRequest{
		OtherSteamid: proto.Uint64(uint64(other)),
	}))
}
