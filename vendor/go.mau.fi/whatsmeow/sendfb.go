// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package whatsmeow

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/libsignal/groups"
	"go.mau.fi/libsignal/keys/prekey"
	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/session"
	"go.mau.fi/libsignal/signalerror"
	"go.mau.fi/util/random"
	"google.golang.org/protobuf/proto"

	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/binary/armadillo/waCommon"
	"go.mau.fi/whatsmeow/binary/armadillo/waConsumerApplication"
	"go.mau.fi/whatsmeow/binary/armadillo/waMsgApplication"
	"go.mau.fi/whatsmeow/binary/armadillo/waMsgTransport"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

const FBMessageVersion = 3
const FBMessageApplicationVersion = 2
const FBConsumerMessageVersion = 1

// SendFBMessage sends the given v3 message to the given JID.
func (cli *Client) SendFBMessage(
	ctx context.Context,
	to types.JID,
	message *waConsumerApplication.ConsumerApplication,
	metadata *waMsgApplication.MessageApplication_Metadata,
	extra ...SendRequestExtra,
) (resp SendResponse, err error) {
	var req SendRequestExtra
	if len(extra) > 1 {
		err = errors.New("only one extra parameter may be provided to SendMessage")
		return
	} else if len(extra) == 1 {
		req = extra[0]
	}
	consumerMessage, err := proto.Marshal(message)
	if err != nil {
		err = fmt.Errorf("failed to marshal consumer message: %w", err)
		return
	}
	if metadata == nil {
		metadata = &waMsgApplication.MessageApplication_Metadata{}
	}
	metadata.FrankingVersion = 0
	metadata.FrankingKey = random.Bytes(32)
	msgAttrs := getAttrsFromFBMessage(message)
	messageAppProto := &waMsgApplication.MessageApplication{
		Payload: &waMsgApplication.MessageApplication_Payload{
			Content: &waMsgApplication.MessageApplication_Payload_SubProtocol{
				SubProtocol: &waMsgApplication.MessageApplication_SubProtocolPayload{
					SubProtocol: &waMsgApplication.MessageApplication_SubProtocolPayload_ConsumerMessage{
						ConsumerMessage: &waCommon.SubProtocol{
							Payload: consumerMessage,
							Version: FBConsumerMessageVersion,
						},
					},
					FutureProof: waCommon.FutureProofBehavior_PLACEHOLDER,
				},
			},
		},
		Metadata: metadata,
	}
	messageApp, err := proto.Marshal(messageAppProto)
	if err != nil {
		return resp, fmt.Errorf("failed to marshal message application: %w", err)
	}
	frankingHash := hmac.New(sha256.New, metadata.FrankingKey)
	frankingHash.Write(messageApp)
	frankingTag := frankingHash.Sum(nil)
	if to.Device > 0 && !req.Peer {
		err = ErrRecipientADJID
		return
	}
	ownID := cli.getOwnID()
	if ownID.IsEmpty() {
		err = ErrNotLoggedIn
		return
	}

	if req.Timeout == 0 {
		req.Timeout = defaultRequestTimeout
	}
	if len(req.ID) == 0 {
		req.ID = cli.GenerateMessageID()
	}
	resp.ID = req.ID

	start := time.Now()
	// Sending multiple messages at a time can cause weird issues and makes it harder to retry safely
	cli.messageSendLock.Lock()
	resp.DebugTimings.Queue = time.Since(start)
	defer cli.messageSendLock.Unlock()

	respChan := cli.waitResponse(req.ID)
	if !req.Peer {
		cli.addRecentMessage(to, req.ID, nil, messageAppProto)
	}
	var phash string
	var data []byte
	switch to.Server {
	case types.GroupServer:
		phash, data, err = cli.sendGroupV3(ctx, to, ownID, req.ID, messageApp, msgAttrs, frankingTag, &resp.DebugTimings)
	case types.DefaultUserServer, types.MessengerServer:
		if req.Peer {
			err = fmt.Errorf("peer messages to fb are not yet supported")
			//data, err = cli.sendPeerMessage(to, req.ID, message, &resp.DebugTimings)
		} else {
			data, phash, err = cli.sendDMV3(ctx, to, ownID, req.ID, messageApp, msgAttrs, frankingTag, &resp.DebugTimings)
		}
	default:
		err = fmt.Errorf("%w %s", ErrUnknownServer, to.Server)
	}
	start = time.Now()
	if err != nil {
		cli.cancelResponse(req.ID, respChan)
		return
	}
	var respNode *waBinary.Node
	var timeoutChan <-chan time.Time
	if req.Timeout > 0 {
		timeoutChan = time.After(req.Timeout)
	} else {
		timeoutChan = make(<-chan time.Time)
	}
	select {
	case respNode = <-respChan:
	case <-timeoutChan:
		cli.cancelResponse(req.ID, respChan)
		err = ErrMessageTimedOut
		return
	case <-ctx.Done():
		cli.cancelResponse(req.ID, respChan)
		err = ctx.Err()
		return
	}
	resp.DebugTimings.Resp = time.Since(start)
	if isDisconnectNode(respNode) {
		start = time.Now()
		respNode, err = cli.retryFrame("message send", req.ID, data, respNode, ctx, 0)
		resp.DebugTimings.Retry = time.Since(start)
		if err != nil {
			return
		}
	}
	ag := respNode.AttrGetter()
	resp.ServerID = types.MessageServerID(ag.OptionalInt("server_id"))
	resp.Timestamp = ag.UnixTime("t")
	if errorCode := ag.Int("error"); errorCode != 0 {
		err = fmt.Errorf("%w %d", ErrServerReturnedError, errorCode)
	}
	expectedPHash := ag.OptionalString("phash")
	if len(expectedPHash) > 0 && phash != expectedPHash {
		cli.Log.Warnf("Server returned different participant list hash when sending to %s. Some devices may not have received the message.", to)
		// TODO also invalidate device list caches
		cli.groupParticipantsCacheLock.Lock()
		delete(cli.groupParticipantsCache, to)
		cli.groupParticipantsCacheLock.Unlock()
	}
	return
}

func (cli *Client) sendGroupV3(
	ctx context.Context,
	to,
	ownID types.JID,
	id types.MessageID,
	messageApp []byte,
	msgAttrs messageAttrs,
	frankingTag []byte,
	timings *MessageDebugTimings,
) (string, []byte, error) {
	var participants []types.JID
	var err error
	start := time.Now()
	if to.Server == types.GroupServer {
		participants, err = cli.getGroupMembers(ctx, to)
		if err != nil {
			return "", nil, fmt.Errorf("failed to get group members: %w", err)
		}
	}
	timings.GetParticipants = time.Since(start)

	start = time.Now()
	builder := groups.NewGroupSessionBuilder(cli.Store, pbSerializer)
	senderKeyName := protocol.NewSenderKeyName(to.String(), ownID.SignalAddress())
	signalSKDMessage, err := builder.Create(senderKeyName)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create sender key distribution message to send %s to %s: %w", id, to, err)
	}
	skdm := &waMsgTransport.MessageTransport_Protocol_Ancillary_SenderKeyDistributionMessage{
		GroupID:                             to.String(),
		AxolotlSenderKeyDistributionMessage: signalSKDMessage.Serialize(),
	}

	cipher := groups.NewGroupCipher(builder, senderKeyName, cli.Store)
	plaintext, err := proto.Marshal(&waMsgTransport.MessageTransport{
		Payload: &waMsgTransport.MessageTransport_Payload{
			ApplicationPayload: &waCommon.SubProtocol{
				Payload: messageApp,
				Version: FBMessageApplicationVersion,
			},
			FutureProof: waCommon.FutureProofBehavior_PLACEHOLDER,
		},
		Protocol: &waMsgTransport.MessageTransport_Protocol{
			Integral: &waMsgTransport.MessageTransport_Protocol_Integral{
				Padding: padMessage(nil),
				DSM:     nil,
			},
			Ancillary: &waMsgTransport.MessageTransport_Protocol_Ancillary{
				Skdm:               nil,
				DeviceListMetadata: nil,
				Icdc:               nil,
				BackupDirective: &waMsgTransport.MessageTransport_Protocol_Ancillary_BackupDirective{
					MessageID:  id,
					ActionType: waMsgTransport.MessageTransport_Protocol_Ancillary_BackupDirective_UPSERT,
				},
			},
		},
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal message transport: %w", err)
	}
	encrypted, err := cipher.Encrypt(plaintext)
	if err != nil {
		return "", nil, fmt.Errorf("failed to encrypt group message to send %s to %s: %w", id, to, err)
	}
	ciphertext := encrypted.SignedSerialize()
	timings.GroupEncrypt = time.Since(start)

	node, allDevices, err := cli.prepareMessageNodeV3(ctx, to, ownID, id, nil, skdm, msgAttrs, frankingTag, participants, timings)
	if err != nil {
		return "", nil, err
	}

	phash := participantListHashV2(allDevices)
	node.Attrs["phash"] = phash
	skMsg := waBinary.Node{
		Tag:     "enc",
		Content: ciphertext,
		Attrs:   waBinary.Attrs{"v": "3", "type": "skmsg"},
	}
	if msgAttrs.MediaType != "" {
		skMsg.Attrs["mediatype"] = msgAttrs.MediaType
	}
	node.Content = append(node.GetChildren(), skMsg)

	start = time.Now()
	data, err := cli.sendNodeAndGetData(*node)
	timings.Send = time.Since(start)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send message node: %w", err)
	}
	return phash, data, nil
}

func (cli *Client) sendDMV3(
	ctx context.Context,
	to,
	ownID types.JID,
	id types.MessageID,
	messageApp []byte,
	msgAttrs messageAttrs,
	frankingTag []byte,
	timings *MessageDebugTimings,
) ([]byte, string, error) {
	payload := &waMsgTransport.MessageTransport_Payload{
		ApplicationPayload: &waCommon.SubProtocol{
			Payload: messageApp,
			Version: FBMessageApplicationVersion,
		},
		FutureProof: waCommon.FutureProofBehavior_PLACEHOLDER,
	}

	node, allDevices, err := cli.prepareMessageNodeV3(ctx, to, ownID, id, payload, nil, msgAttrs, frankingTag, []types.JID{to, ownID.ToNonAD()}, timings)
	if err != nil {
		return nil, "", err
	}
	start := time.Now()
	data, err := cli.sendNodeAndGetData(*node)
	timings.Send = time.Since(start)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send message node: %w", err)
	}
	return data, participantListHashV2(allDevices), nil
}

type messageAttrs struct {
	Type        string
	MediaType   string
	Edit        types.EditAttribute
	DecryptFail events.DecryptFailMode
	PollType    string
}

func getAttrsFromFBMessage(msg *waConsumerApplication.ConsumerApplication) (attrs messageAttrs) {
	switch payload := msg.GetPayload().GetPayload().(type) {
	case *waConsumerApplication.ConsumerApplication_Payload_Content:
		switch content := payload.Content.GetContent().(type) {
		case *waConsumerApplication.ConsumerApplication_Content_MessageText,
			*waConsumerApplication.ConsumerApplication_Content_ExtendedTextMessage:
			attrs.Type = "text"
		case *waConsumerApplication.ConsumerApplication_Content_ImageMessage:
			attrs.MediaType = "image"
		case *waConsumerApplication.ConsumerApplication_Content_StickerMessage:
			attrs.MediaType = "sticker"
		case *waConsumerApplication.ConsumerApplication_Content_ViewOnceMessage:
			switch content.ViewOnceMessage.GetViewOnceContent().(type) {
			case *waConsumerApplication.ConsumerApplication_ViewOnceMessage_ImageMessage:
				attrs.MediaType = "image"
			case *waConsumerApplication.ConsumerApplication_ViewOnceMessage_VideoMessage:
				attrs.MediaType = "video"
			}
		case *waConsumerApplication.ConsumerApplication_Content_DocumentMessage:
			attrs.MediaType = "document"
		case *waConsumerApplication.ConsumerApplication_Content_AudioMessage:
			if content.AudioMessage.GetPTT() {
				attrs.MediaType = "ptt"
			} else {
				attrs.MediaType = "audio"
			}
		case *waConsumerApplication.ConsumerApplication_Content_VideoMessage:
			// TODO gifPlayback?
			attrs.MediaType = "video"
		case *waConsumerApplication.ConsumerApplication_Content_LocationMessage:
			attrs.MediaType = "location"
		case *waConsumerApplication.ConsumerApplication_Content_LiveLocationMessage:
			attrs.MediaType = "location"
		case *waConsumerApplication.ConsumerApplication_Content_ContactMessage:
			attrs.MediaType = "vcard"
		case *waConsumerApplication.ConsumerApplication_Content_ContactsArrayMessage:
			attrs.MediaType = "contact_array"
		case *waConsumerApplication.ConsumerApplication_Content_PollCreationMessage:
			attrs.PollType = "creation"
			attrs.Type = "poll"
		case *waConsumerApplication.ConsumerApplication_Content_PollUpdateMessage:
			attrs.PollType = "vote"
			attrs.Type = "poll"
			attrs.DecryptFail = events.DecryptFailHide
		case *waConsumerApplication.ConsumerApplication_Content_ReactionMessage:
			attrs.Type = "reaction"
			attrs.DecryptFail = events.DecryptFailHide
		case *waConsumerApplication.ConsumerApplication_Content_EditMessage:
			attrs.Edit = types.EditAttributeMessageEdit
			attrs.DecryptFail = events.DecryptFailHide
		}
		if attrs.MediaType != "" && attrs.Type == "" {
			attrs.Type = "media"
		}
	case *waConsumerApplication.ConsumerApplication_Payload_ApplicationData:
		switch content := payload.ApplicationData.GetApplicationContent().(type) {
		case *waConsumerApplication.ConsumerApplication_ApplicationData_Revoke:
			if content.Revoke.GetKey().GetFromMe() {
				attrs.Edit = types.EditAttributeSenderRevoke
			} else {
				attrs.Edit = types.EditAttributeAdminRevoke
			}
			attrs.DecryptFail = events.DecryptFailHide
		}
	case *waConsumerApplication.ConsumerApplication_Payload_Signal:
	case *waConsumerApplication.ConsumerApplication_Payload_SubProtocol:
	}
	if attrs.Type == "" {
		attrs.Type = "text"
	}
	return
}

func (cli *Client) prepareMessageNodeV3(
	ctx context.Context,
	to,
	ownID types.JID,
	id types.MessageID,
	payload *waMsgTransport.MessageTransport_Payload,
	skdm *waMsgTransport.MessageTransport_Protocol_Ancillary_SenderKeyDistributionMessage,
	msgAttrs messageAttrs,
	frankingTag []byte,
	participants []types.JID,
	timings *MessageDebugTimings,
) (*waBinary.Node, []types.JID, error) {
	start := time.Now()
	allDevices, err := cli.GetUserDevicesContext(ctx, participants)
	timings.GetDevices = time.Since(start)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get device list: %w", err)
	}

	encAttrs := waBinary.Attrs{}
	attrs := waBinary.Attrs{
		"id":   id,
		"type": msgAttrs.Type,
		"to":   to,
	}
	// Only include mediatype on DMs, for groups it's in the skmsg node
	if payload != nil && msgAttrs.MediaType != "" {
		encAttrs["mediatype"] = msgAttrs.MediaType
	}
	if msgAttrs.Edit != "" {
		attrs["edit"] = string(msgAttrs.Edit)
	}
	if msgAttrs.DecryptFail != "" {
		encAttrs["decrypt-fail"] = string(msgAttrs.DecryptFail)
	}

	dsm := &waMsgTransport.MessageTransport_Protocol_Integral_DeviceSentMessage{
		DestinationJID: to.String(),
		Phash:          "",
	}

	start = time.Now()
	participantNodes := cli.encryptMessageForDevicesV3(ctx, allDevices, ownID, id, payload, skdm, dsm, encAttrs)
	timings.PeerEncrypt = time.Since(start)
	content := make([]waBinary.Node, 0, 4)
	content = append(content, waBinary.Node{
		Tag:     "participants",
		Content: participantNodes,
	})
	metaAttrs := make(waBinary.Attrs)
	if msgAttrs.PollType != "" {
		metaAttrs["polltype"] = msgAttrs.PollType
	}
	if msgAttrs.DecryptFail != "" {
		metaAttrs["decrypt-fail"] = string(msgAttrs.DecryptFail)
	}
	if len(metaAttrs) > 0 {
		content = append(content, waBinary.Node{
			Tag:   "meta",
			Attrs: metaAttrs,
		})
	}
	traceRequestID := uuid.New()
	content = append(content, waBinary.Node{
		Tag: "franking",
		Content: []waBinary.Node{{
			Tag:     "franking_tag",
			Content: frankingTag,
		}},
	}, waBinary.Node{
		Tag: "trace",
		Content: []waBinary.Node{{
			Tag:     "request_id",
			Content: traceRequestID[:],
		}},
	})
	return &waBinary.Node{
		Tag:     "message",
		Attrs:   attrs,
		Content: content,
	}, allDevices, nil
}

func (cli *Client) encryptMessageForDevicesV3(
	ctx context.Context,
	allDevices []types.JID,
	ownID types.JID,
	id string,
	payload *waMsgTransport.MessageTransport_Payload,
	skdm *waMsgTransport.MessageTransport_Protocol_Ancillary_SenderKeyDistributionMessage,
	dsm *waMsgTransport.MessageTransport_Protocol_Integral_DeviceSentMessage,
	encAttrs waBinary.Attrs,
) []waBinary.Node {
	participantNodes := make([]waBinary.Node, 0, len(allDevices))
	var retryDevices []types.JID
	for _, jid := range allDevices {
		var dsmForDevice *waMsgTransport.MessageTransport_Protocol_Integral_DeviceSentMessage
		if jid.User == ownID.User {
			if jid == ownID {
				continue
			}
			dsmForDevice = dsm
		}
		encrypted, err := cli.encryptMessageForDeviceAndWrapV3(payload, skdm, dsmForDevice, jid, nil, encAttrs)
		if errors.Is(err, ErrNoSession) {
			retryDevices = append(retryDevices, jid)
			continue
		} else if err != nil {
			cli.Log.Warnf("Failed to encrypt %s for %s: %v", id, jid, err)
			continue
		}
		participantNodes = append(participantNodes, *encrypted)
	}
	if len(retryDevices) > 0 {
		bundles, err := cli.fetchPreKeys(ctx, retryDevices)
		if err != nil {
			cli.Log.Warnf("Failed to fetch prekeys for %v to retry encryption: %v", retryDevices, err)
		} else {
			for _, jid := range retryDevices {
				resp := bundles[jid]
				if resp.err != nil {
					cli.Log.Warnf("Failed to fetch prekey for %s: %v", jid, resp.err)
					continue
				}
				var dsmForDevice *waMsgTransport.MessageTransport_Protocol_Integral_DeviceSentMessage
				if jid.User == ownID.User {
					dsmForDevice = dsm
				}
				encrypted, err := cli.encryptMessageForDeviceAndWrapV3(payload, skdm, dsmForDevice, jid, resp.bundle, encAttrs)
				if err != nil {
					cli.Log.Warnf("Failed to encrypt %s for %s (retry): %v", id, jid, err)
					continue
				}
				participantNodes = append(participantNodes, *encrypted)
			}
		}
	}
	return participantNodes
}

func (cli *Client) encryptMessageForDeviceAndWrapV3(
	payload *waMsgTransport.MessageTransport_Payload,
	skdm *waMsgTransport.MessageTransport_Protocol_Ancillary_SenderKeyDistributionMessage,
	dsm *waMsgTransport.MessageTransport_Protocol_Integral_DeviceSentMessage,
	to types.JID,
	bundle *prekey.Bundle,
	encAttrs waBinary.Attrs,
) (*waBinary.Node, error) {
	node, err := cli.encryptMessageForDeviceV3(payload, skdm, dsm, to, bundle, encAttrs)
	if err != nil {
		return nil, err
	}
	return &waBinary.Node{
		Tag:     "to",
		Attrs:   waBinary.Attrs{"jid": to},
		Content: []waBinary.Node{*node},
	}, nil
}

func (cli *Client) encryptMessageForDeviceV3(
	payload *waMsgTransport.MessageTransport_Payload,
	skdm *waMsgTransport.MessageTransport_Protocol_Ancillary_SenderKeyDistributionMessage,
	dsm *waMsgTransport.MessageTransport_Protocol_Integral_DeviceSentMessage,
	to types.JID,
	bundle *prekey.Bundle,
	extraAttrs waBinary.Attrs,
) (*waBinary.Node, error) {
	builder := session.NewBuilderFromSignal(cli.Store, to.SignalAddress(), pbSerializer)
	if bundle != nil {
		cli.Log.Debugf("Processing prekey bundle for %s", to)
		err := builder.ProcessBundle(bundle)
		if cli.AutoTrustIdentity && errors.Is(err, signalerror.ErrUntrustedIdentity) {
			cli.Log.Warnf("Got %v error while trying to process prekey bundle for %s, clearing stored identity and retrying", err, to)
			cli.clearUntrustedIdentity(to)
			err = builder.ProcessBundle(bundle)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to process prekey bundle: %w", err)
		}
	} else if !cli.Store.ContainsSession(to.SignalAddress()) {
		return nil, ErrNoSession
	}
	cipher := session.NewCipher(builder, to.SignalAddress())
	plaintext, err := proto.Marshal(&waMsgTransport.MessageTransport{
		Payload: payload,
		Protocol: &waMsgTransport.MessageTransport_Protocol{
			Integral: &waMsgTransport.MessageTransport_Protocol_Integral{
				Padding: padMessage(nil),
				DSM:     dsm,
			},
			Ancillary: &waMsgTransport.MessageTransport_Protocol_Ancillary{
				Skdm:               skdm,
				DeviceListMetadata: nil,
				Icdc:               nil,
				BackupDirective:    nil,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message transport: %w", err)
	}
	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		return nil, fmt.Errorf("cipher encryption failed: %w", err)
	}

	encAttrs := waBinary.Attrs{
		"v":    FBMessageVersion,
		"type": "msg",
	}
	if ciphertext.Type() == protocol.PREKEY_TYPE {
		encAttrs["type"] = "pkmsg"
	}
	copyAttrs(extraAttrs, encAttrs)

	return &waBinary.Node{
		Tag:     "enc",
		Attrs:   encAttrs,
		Content: ciphertext.Serialize(),
	}, nil
}
