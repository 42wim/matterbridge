// Copyright (c) 2021 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package whatsmeow

import (
	"context"
	"errors"
	"fmt"
	"strings"

	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

const InviteLinkPrefix = "https://chat.whatsapp.com/"

func (cli *Client) sendGroupIQ(ctx context.Context, iqType infoQueryType, jid types.JID, content waBinary.Node) (*waBinary.Node, error) {
	return cli.sendIQ(infoQuery{
		Context:   ctx,
		Namespace: "w:g2",
		Type:      iqType,
		To:        jid,
		Content:   []waBinary.Node{content},
	})
}

// CreateGroup creates a group on WhatsApp with the given name and participants.
//
// You don't need to include your own JID in the participants array, the WhatsApp servers will add it implicitly.
//
// Group names are limited to 25 characters. A longer group name will cause a 406 not acceptable error.
//
// Optionally, a create key can be provided to deduplicate the group create notification that will be triggered
// when the group is created. If provided, the JoinedGroup event will contain the same key.
func (cli *Client) CreateGroup(name string, participants []types.JID, createKey types.MessageID) (*types.GroupInfo, error) {
	participantNodes := make([]waBinary.Node, len(participants))
	for i, participant := range participants {
		participantNodes[i] = waBinary.Node{
			Tag:   "participant",
			Attrs: waBinary.Attrs{"jid": participant},
		}
	}
	if createKey == "" {
		createKey = GenerateMessageID()
	}
	// WhatsApp web doesn't seem to include the static prefix for these
	key := strings.TrimPrefix(createKey, "3EB0")
	resp, err := cli.sendGroupIQ(context.TODO(), iqSet, types.GroupServerJID, waBinary.Node{
		Tag: "create",
		Attrs: waBinary.Attrs{
			"subject": name,
			"key":     key,
		},
		Content: participantNodes,
	})
	if err != nil {
		return nil, err
	}
	groupNode, ok := resp.GetOptionalChildByTag("group")
	if !ok {
		return nil, &ElementMissingError{Tag: "group", In: "response to create group query"}
	}
	return cli.parseGroupNode(&groupNode)
}

// LeaveGroup leaves the specified group on WhatsApp.
func (cli *Client) LeaveGroup(jid types.JID) error {
	_, err := cli.sendGroupIQ(context.TODO(), iqSet, types.GroupServerJID, waBinary.Node{
		Tag: "leave",
		Content: []waBinary.Node{{
			Tag:   "group",
			Attrs: waBinary.Attrs{"id": jid},
		}},
	})
	return err
}

type ParticipantChange string

const (
	ParticipantChangeAdd     ParticipantChange = "add"
	ParticipantChangeRemove  ParticipantChange = "remove"
	ParticipantChangePromote ParticipantChange = "promote"
	ParticipantChangeDemote  ParticipantChange = "demote"
)

// UpdateGroupParticipants can be used to add, remove, promote and demote members in a WhatsApp group.
func (cli *Client) UpdateGroupParticipants(jid types.JID, participantChanges map[types.JID]ParticipantChange) (*waBinary.Node, error) {
	content := make([]waBinary.Node, len(participantChanges))
	i := 0
	for participantJID, change := range participantChanges {
		content[i] = waBinary.Node{
			Tag: string(change),
			Content: []waBinary.Node{{
				Tag:   "participant",
				Attrs: waBinary.Attrs{"jid": participantJID},
			}},
		}
		i++
	}
	resp, err := cli.sendIQ(infoQuery{
		Namespace: "w:g2",
		Type:      iqSet,
		To:        jid,
		Content:   content,
	})
	if err != nil {
		return nil, err
	}
	// TODO proper return value?
	return resp, nil
}

// SetGroupPhoto updates the group picture/icon of the given group on WhatsApp.
// The avatar should be a JPEG photo, other formats may be rejected with ErrInvalidImageFormat.
// The bytes can be nil to remove the photo. Returns the new picture ID.
func (cli *Client) SetGroupPhoto(jid types.JID, avatar []byte) (string, error) {
	var content interface{}
	if avatar != nil {
		content = []waBinary.Node{{
			Tag:     "picture",
			Attrs:   waBinary.Attrs{"type": "image"},
			Content: avatar,
		}}
	}
	resp, err := cli.sendIQ(infoQuery{
		Namespace: "w:profile:picture",
		Type:      iqSet,
		To:        types.ServerJID,
		Target:    jid,
		Content:   content,
	})
	if errors.Is(err, ErrIQNotAcceptable) {
		return "", wrapIQError(ErrInvalidImageFormat, err)
	} else if err != nil {
		return "", err
	}
	if avatar == nil {
		return "remove", nil
	}
	pictureID, ok := resp.GetChildByTag("picture").Attrs["id"].(string)
	if !ok {
		return "", fmt.Errorf("didn't find picture ID in response")
	}
	return pictureID, nil
}

// SetGroupName updates the name (subject) of the given group on WhatsApp.
func (cli *Client) SetGroupName(jid types.JID, name string) error {
	_, err := cli.sendGroupIQ(context.TODO(), iqSet, jid, waBinary.Node{
		Tag:     "subject",
		Content: []byte(name),
	})
	return err
}

// SetGroupTopic updates the topic (description) of the given group on WhatsApp.
//
// The previousID and newID fields are optional. If the previous ID is not specified, this will
// automatically fetch the current group info to find the previous topic ID. If the new ID is not
// specified, one will be generated with GenerateMessageID().
func (cli *Client) SetGroupTopic(jid types.JID, previousID, newID, topic string) error {
	if previousID == "" {
		oldInfo, err := cli.GetGroupInfo(jid)
		if err != nil {
			return fmt.Errorf("failed to get old group info to update topic: %v", err)
		}
		previousID = oldInfo.TopicID
	}
	if newID == "" {
		newID = GenerateMessageID()
	}
	_, err := cli.sendGroupIQ(context.TODO(), iqSet, jid, waBinary.Node{
		Tag: "description",
		Attrs: waBinary.Attrs{
			"prev": previousID,
			"id":   newID,
		},
		Content: []waBinary.Node{{
			Tag:     "body",
			Content: []byte(topic),
		}},
	})
	return err
}

// SetGroupLocked changes whether the group is locked (i.e. whether only admins can modify group info).
func (cli *Client) SetGroupLocked(jid types.JID, locked bool) error {
	tag := "locked"
	if !locked {
		tag = "unlocked"
	}
	_, err := cli.sendGroupIQ(context.TODO(), iqSet, jid, waBinary.Node{Tag: tag})
	return err
}

// SetGroupAnnounce changes whether the group is in announce mode (i.e. whether only admins can send messages).
func (cli *Client) SetGroupAnnounce(jid types.JID, announce bool) error {
	tag := "announcement"
	if !announce {
		tag = "not_announcement"
	}
	_, err := cli.sendGroupIQ(context.TODO(), iqSet, jid, waBinary.Node{Tag: tag})
	return err
}

// GetGroupInviteLink requests the invite link to the group from the WhatsApp servers.
//
// If reset is true, then the old invite link will be revoked and a new one generated.
func (cli *Client) GetGroupInviteLink(jid types.JID, reset bool) (string, error) {
	iqType := iqGet
	if reset {
		iqType = iqSet
	}
	resp, err := cli.sendGroupIQ(context.TODO(), iqType, jid, waBinary.Node{Tag: "invite"})
	if errors.Is(err, ErrIQNotAuthorized) {
		return "", wrapIQError(ErrGroupInviteLinkUnauthorized, err)
	} else if errors.Is(err, ErrIQNotFound) {
		return "", wrapIQError(ErrGroupNotFound, err)
	} else if errors.Is(err, ErrIQForbidden) {
		return "", wrapIQError(ErrNotInGroup, err)
	} else if err != nil {
		return "", err
	}
	code, ok := resp.GetChildByTag("invite").Attrs["code"].(string)
	if !ok {
		return "", fmt.Errorf("didn't find invite code in response")
	}
	return InviteLinkPrefix + code, nil
}

// GetGroupInfoFromInvite gets the group info from an invite message.
//
// Note that this is specifically for invite messages, not invite links. Use GetGroupInfoFromLink for resolving chat.whatsapp.com links.
func (cli *Client) GetGroupInfoFromInvite(jid, inviter types.JID, code string, expiration int64) (*types.GroupInfo, error) {
	resp, err := cli.sendGroupIQ(context.TODO(), iqGet, jid, waBinary.Node{
		Tag: "query",
		Content: []waBinary.Node{{
			Tag: "add_request",
			Attrs: waBinary.Attrs{
				"code":       code,
				"expiration": expiration,
				"admin":      inviter,
			},
		}},
	})
	if err != nil {
		return nil, err
	}
	groupNode, ok := resp.GetOptionalChildByTag("group")
	if !ok {
		return nil, &ElementMissingError{Tag: "group", In: "response to invite group info query"}
	}
	return cli.parseGroupNode(&groupNode)
}

// JoinGroupWithInvite joins a group using an invite message.
//
// Note that this is specifically for invite messages, not invite links. Use JoinGroupWithLink for joining with chat.whatsapp.com links.
func (cli *Client) JoinGroupWithInvite(jid, inviter types.JID, code string, expiration int64) error {
	_, err := cli.sendGroupIQ(context.TODO(), iqSet, jid, waBinary.Node{
		Tag: "accept",
		Attrs: waBinary.Attrs{
			"code":       code,
			"expiration": expiration,
			"admin":      inviter,
		},
	})
	return err
}

// GetGroupInfoFromLink resolves the given invite link and asks the WhatsApp servers for info about the group.
// This will not cause the user to join the group.
func (cli *Client) GetGroupInfoFromLink(code string) (*types.GroupInfo, error) {
	code = strings.TrimPrefix(code, InviteLinkPrefix)
	resp, err := cli.sendGroupIQ(context.TODO(), iqGet, types.GroupServerJID, waBinary.Node{
		Tag:   "invite",
		Attrs: waBinary.Attrs{"code": code},
	})
	if errors.Is(err, ErrIQGone) {
		return nil, wrapIQError(ErrInviteLinkRevoked, err)
	} else if errors.Is(err, ErrIQNotAcceptable) {
		return nil, wrapIQError(ErrInviteLinkInvalid, err)
	} else if err != nil {
		return nil, err
	}
	groupNode, ok := resp.GetOptionalChildByTag("group")
	if !ok {
		return nil, &ElementMissingError{Tag: "group", In: "response to group link info query"}
	}
	return cli.parseGroupNode(&groupNode)
}

// JoinGroupWithLink joins the group using the given invite link.
func (cli *Client) JoinGroupWithLink(code string) (types.JID, error) {
	code = strings.TrimPrefix(code, InviteLinkPrefix)
	resp, err := cli.sendGroupIQ(context.TODO(), iqSet, types.GroupServerJID, waBinary.Node{
		Tag:   "invite",
		Attrs: waBinary.Attrs{"code": code},
	})
	if errors.Is(err, ErrIQGone) {
		return types.EmptyJID, wrapIQError(ErrInviteLinkRevoked, err)
	} else if errors.Is(err, ErrIQNotAcceptable) {
		return types.EmptyJID, wrapIQError(ErrInviteLinkInvalid, err)
	} else if err != nil {
		return types.EmptyJID, err
	}
	groupNode, ok := resp.GetOptionalChildByTag("group")
	if !ok {
		return types.EmptyJID, &ElementMissingError{Tag: "group", In: "response to group link join query"}
	}
	return groupNode.AttrGetter().JID("jid"), nil
}

// GetJoinedGroups returns the list of groups the user is participating in.
func (cli *Client) GetJoinedGroups() ([]*types.GroupInfo, error) {
	resp, err := cli.sendGroupIQ(context.TODO(), iqGet, types.GroupServerJID, waBinary.Node{
		Tag: "participating",
		Content: []waBinary.Node{
			{Tag: "participants"},
			{Tag: "description"},
		},
	})
	if err != nil {
		return nil, err
	}
	groups, ok := resp.GetOptionalChildByTag("groups")
	if !ok {
		return nil, &ElementMissingError{Tag: "groups", In: "response to group list query"}
	}
	children := groups.GetChildren()
	infos := make([]*types.GroupInfo, 0, len(children))
	for _, child := range children {
		if child.Tag != "group" {
			cli.Log.Debugf("Unexpected child in group list response: %s", child.XMLString())
			continue
		}
		parsed, parseErr := cli.parseGroupNode(&child)
		if parseErr != nil {
			cli.Log.Warnf("Error parsing group %s: %v", parsed.JID, parseErr)
		}
		infos = append(infos, parsed)
	}
	return infos, nil
}

// GetGroupInfo requests basic info about a group chat from the WhatsApp servers.
func (cli *Client) GetGroupInfo(jid types.JID) (*types.GroupInfo, error) {
	return cli.getGroupInfo(context.TODO(), jid, true)
}

func (cli *Client) getGroupInfo(ctx context.Context, jid types.JID, lockParticipantCache bool) (*types.GroupInfo, error) {
	res, err := cli.sendGroupIQ(ctx, iqGet, jid, waBinary.Node{
		Tag:   "query",
		Attrs: waBinary.Attrs{"request": "interactive"},
	})
	if errors.Is(err, ErrIQNotFound) {
		return nil, wrapIQError(ErrGroupNotFound, err)
	} else if errors.Is(err, ErrIQForbidden) {
		return nil, wrapIQError(ErrNotInGroup, err)
	} else if err != nil {
		return nil, err
	}

	groupNode, ok := res.GetOptionalChildByTag("group")
	if !ok {
		return nil, &ElementMissingError{Tag: "groups", In: "response to group info query"}
	}
	groupInfo, err := cli.parseGroupNode(&groupNode)
	if err != nil {
		return groupInfo, err
	}
	if lockParticipantCache {
		cli.groupParticipantsCacheLock.Lock()
		defer cli.groupParticipantsCacheLock.Unlock()
	}
	participants := make([]types.JID, len(groupInfo.Participants))
	for i, part := range groupInfo.Participants {
		participants[i] = part.JID
	}
	cli.groupParticipantsCache[jid] = participants
	return groupInfo, nil
}

func (cli *Client) getGroupMembers(ctx context.Context, jid types.JID) ([]types.JID, error) {
	cli.groupParticipantsCacheLock.Lock()
	defer cli.groupParticipantsCacheLock.Unlock()
	if _, ok := cli.groupParticipantsCache[jid]; !ok {
		_, err := cli.getGroupInfo(ctx, jid, false)
		if err != nil {
			return nil, err
		}
	}
	return cli.groupParticipantsCache[jid], nil
}

func (cli *Client) parseGroupNode(groupNode *waBinary.Node) (*types.GroupInfo, error) {
	var group types.GroupInfo
	ag := groupNode.AttrGetter()

	group.JID = types.NewJID(ag.String("id"), types.GroupServer)
	group.OwnerJID = ag.OptionalJIDOrEmpty("creator")

	group.Name = ag.String("subject")
	group.NameSetAt = ag.UnixTime("s_t")
	group.NameSetBy = ag.OptionalJIDOrEmpty("s_o")

	group.GroupCreated = ag.UnixTime("creation")

	group.AnnounceVersionID = ag.OptionalString("a_v_id")
	group.ParticipantVersionID = ag.OptionalString("p_v_id")

	for _, child := range groupNode.GetChildren() {
		childAG := child.AttrGetter()
		switch child.Tag {
		case "participant":
			pcpType := childAG.OptionalString("type")
			participant := types.GroupParticipant{
				IsAdmin:      pcpType == "admin" || pcpType == "superadmin",
				IsSuperAdmin: pcpType == "superadmin",
				JID:          childAG.JID("jid"),
			}
			group.Participants = append(group.Participants, participant)
		case "description":
			body, bodyOK := child.GetOptionalChildByTag("body")
			if bodyOK {
				topicBytes, _ := body.Content.([]byte)
				group.Topic = string(topicBytes)
				group.TopicID = childAG.String("id")
				group.TopicSetBy = childAG.OptionalJIDOrEmpty("participant")
				group.TopicSetAt = childAG.UnixTime("t")
			}
		case "announcement":
			group.IsAnnounce = true
		case "locked":
			group.IsLocked = true
		case "ephemeral":
			group.IsEphemeral = true
			group.DisappearingTimer = uint32(childAG.Uint64("expiration"))
		case "member_add_mode":
			modeBytes, _ := child.Content.([]byte)
			group.MemberAddMode = types.GroupMemberAddMode(modeBytes)
		default:
			cli.Log.Debugf("Unknown element in group node %s: %s", group.JID.String(), child.XMLString())
		}
		if !childAG.OK() {
			cli.Log.Warnf("Possibly failed to parse %s element in group node: %+v", child.Tag, childAG.Errors)
		}
	}

	return &group, ag.Error()
}

func parseParticipantList(node *waBinary.Node) (participants []types.JID) {
	children := node.GetChildren()
	participants = make([]types.JID, 0, len(children))
	for _, child := range children {
		jid, ok := child.Attrs["jid"].(types.JID)
		if child.Tag != "participant" || !ok {
			continue
		}
		participants = append(participants, jid)
	}
	return
}

func (cli *Client) parseGroupCreate(node *waBinary.Node) (*events.JoinedGroup, error) {
	groupNode, ok := node.GetOptionalChildByTag("group")
	if !ok {
		return nil, fmt.Errorf("group create notification didn't contain group info")
	}
	var evt events.JoinedGroup
	ag := node.AttrGetter()
	evt.Reason = ag.OptionalString("reason")
	evt.CreateKey = ag.OptionalString("key")
	evt.Type = ag.OptionalString("type")
	info, err := cli.parseGroupNode(&groupNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse group info in create notification: %w", err)
	}
	evt.GroupInfo = *info
	return &evt, nil
}

func (cli *Client) parseGroupChange(node *waBinary.Node) (*events.GroupInfo, error) {
	var evt events.GroupInfo
	ag := node.AttrGetter()
	evt.JID = ag.JID("from")
	evt.Notify = ag.OptionalString("notify")
	evt.Sender = ag.OptionalJID("participant")
	evt.Timestamp = ag.UnixTime("t")
	if !ag.OK() {
		return nil, fmt.Errorf("group change doesn't contain required attributes: %w", ag.Error())
	}

	for _, child := range node.GetChildren() {
		cag := child.AttrGetter()
		if child.Tag == "add" || child.Tag == "remove" || child.Tag == "promote" || child.Tag == "demote" {
			evt.PrevParticipantVersionID = cag.String("prev_v_id")
			evt.ParticipantVersionID = cag.String("v_id")
		}
		switch child.Tag {
		case "add":
			evt.JoinReason = cag.OptionalString("reason")
			evt.Join = parseParticipantList(&child)
		case "remove":
			evt.Leave = parseParticipantList(&child)
		case "promote":
			evt.Promote = parseParticipantList(&child)
		case "demote":
			evt.Demote = parseParticipantList(&child)
		case "locked":
			evt.Locked = &types.GroupLocked{IsLocked: true}
		case "unlocked":
			evt.Locked = &types.GroupLocked{IsLocked: false}
		case "subject":
			evt.Name = &types.GroupName{
				Name:      cag.String("subject"),
				NameSetAt: cag.UnixTime("s_t"),
				NameSetBy: cag.OptionalJIDOrEmpty("s_o"),
			}
		case "description":
			topicChild := child.GetChildByTag("body")
			topicBytes, ok := topicChild.Content.([]byte)
			if !ok {
				return nil, fmt.Errorf("group change description has unexpected body: %s", topicChild.XMLString())
			}
			var setBy types.JID
			if evt.Sender != nil {
				setBy = *evt.Sender
			}
			evt.Topic = &types.GroupTopic{
				Topic:      string(topicBytes),
				TopicID:    cag.String("id"),
				TopicSetAt: evt.Timestamp,
				TopicSetBy: setBy,
			}
		case "announcement":
			evt.Announce = &types.GroupAnnounce{
				IsAnnounce:        true,
				AnnounceVersionID: cag.String("v_id"),
			}
		case "not_announcement":
			evt.Announce = &types.GroupAnnounce{
				IsAnnounce:        false,
				AnnounceVersionID: cag.String("v_id"),
			}
		case "invite":
			link := InviteLinkPrefix + cag.String("code")
			evt.NewInviteLink = &link
		case "ephemeral":
			timer := uint32(cag.Uint64("expiration"))
			evt.Ephemeral = &types.GroupEphemeral{
				IsEphemeral:       true,
				DisappearingTimer: timer,
			}
		case "not_ephemeral":
			evt.Ephemeral = &types.GroupEphemeral{IsEphemeral: false}
		default:
			evt.UnknownChanges = append(evt.UnknownChanges, &child)
		}
		if !cag.OK() {
			return nil, fmt.Errorf("group change %s element doesn't contain required attributes: %w", child.Tag, cag.Error())
		}
	}
	return &evt, nil
}

func (cli *Client) updateGroupParticipantCache(evt *events.GroupInfo) {
	if len(evt.Join) == 0 && len(evt.Leave) == 0 {
		return
	}
	cli.groupParticipantsCacheLock.Lock()
	defer cli.groupParticipantsCacheLock.Unlock()
	cached, ok := cli.groupParticipantsCache[evt.JID]
	if !ok {
		return
	}
Outer:
	for _, jid := range evt.Join {
		for _, existingJID := range cached {
			if jid == existingJID {
				continue Outer
			}
		}
		cached = append(cached, jid)
	}
	for _, jid := range evt.Leave {
		for i, existingJID := range cached {
			if existingJID == jid {
				cached[i] = cached[len(cached)-1]
				cached = cached[:len(cached)-1]
				break
			}
		}
	}
	cli.groupParticipantsCache[evt.JID] = cached
}

func (cli *Client) parseGroupNotification(node *waBinary.Node) (interface{}, error) {
	children := node.GetChildren()
	if len(children) == 1 && children[0].Tag == "create" {
		return cli.parseGroupCreate(&children[0])
	} else {
		groupChange, err := cli.parseGroupChange(node)
		if err != nil {
			return nil, err
		}
		cli.updateGroupParticipantCache(groupChange)
		return groupChange, nil
	}
}
