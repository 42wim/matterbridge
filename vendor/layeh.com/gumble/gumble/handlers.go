package gumble

import (
	"crypto/x509"
	"encoding/binary"
	"errors"
	"math"
	"net"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"layeh.com/gumble/gumble/MumbleProto"
	"layeh.com/gumble/gumble/varint"
)

var (
	errUnimplementedHandler = errors.New("gumble: the handler has not been implemented")
	errIncompleteProtobuf   = errors.New("gumble: protobuf message is missing a required field")
	errInvalidProtobuf      = errors.New("gumble: protobuf message has an invalid field")
	errUnsupportedAudio     = errors.New("gumble: unsupported audio codec")
	errNoCodec              = errors.New("gumble: no audio codec")
)

var handlers = [...]func(*Client, []byte) error{
	(*Client).handleVersion,
	(*Client).handleUDPTunnel,
	(*Client).handleAuthenticate,
	(*Client).handlePing,
	(*Client).handleReject,
	(*Client).handleServerSync,
	(*Client).handleChannelRemove,
	(*Client).handleChannelState,
	(*Client).handleUserRemove,
	(*Client).handleUserState,
	(*Client).handleBanList,
	(*Client).handleTextMessage,
	(*Client).handlePermissionDenied,
	(*Client).handleACL,
	(*Client).handleQueryUsers,
	(*Client).handleCryptSetup,
	(*Client).handleContextActionModify,
	(*Client).handleContextAction,
	(*Client).handleUserList,
	(*Client).handleVoiceTarget,
	(*Client).handlePermissionQuery,
	(*Client).handleCodecVersion,
	(*Client).handleUserStats,
	(*Client).handleRequestBlob,
	(*Client).handleServerConfig,
	(*Client).handleSuggestConfig,
}

func parseVersion(packet *MumbleProto.Version) Version {
	var version Version
	if packet.Version != nil {
		version.Version = *packet.Version
	}
	if packet.Release != nil {
		version.Release = *packet.Release
	}
	if packet.Os != nil {
		version.OS = *packet.Os
	}
	if packet.OsVersion != nil {
		version.OSVersion = *packet.OsVersion
	}
	return version
}

func (c *Client) handleVersion(buffer []byte) error {
	var packet MumbleProto.Version
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}
	return nil
}

func (c *Client) handleUDPTunnel(buffer []byte) error {
	if len(buffer) < 1 {
		return errInvalidProtobuf
	}
	audioType := (buffer[0] >> 5) & 0x7
	audioTarget := buffer[0] & 0x1F

	// Opus only
	// TODO: add handling for other packet types
	if audioType != audioCodecIDOpus {
		return errUnsupportedAudio
	}

	// Session
	buffer = buffer[1:]
	session, n := varint.Decode(buffer)
	if n <= 0 {
		return errInvalidProtobuf
	}
	buffer = buffer[n:]
	user := c.Users[uint32(session)]
	if user == nil {
		return errInvalidProtobuf
	}
	decoder := user.decoder
	if decoder == nil {
		// TODO: decoder pool
		// TODO: de-reference after stream is done
		codec := c.audioCodec
		if codec == nil {
			return errNoCodec
		}
		decoder = codec.NewDecoder()
		user.decoder = decoder
	}

	// Sequence
	// TODO: use in jitter buffer
	_, n = varint.Decode(buffer)
	if n <= 0 {
		return errInvalidProtobuf
	}
	buffer = buffer[n:]

	// Length
	length, n := varint.Decode(buffer)
	if n <= 0 {
		return errInvalidProtobuf
	}
	buffer = buffer[n:]
	// Opus audio packets set the 13th bit in the size field as the terminator.
	audioLength := int(length) &^ 0x2000
	if audioLength > len(buffer) {
		return errInvalidProtobuf
	}

	pcm, err := decoder.Decode(buffer[:audioLength], AudioMaximumFrameSize)
	if err != nil {
		return err
	}

	event := AudioPacket{
		Client: c,
		Sender: user,
		Target: &VoiceTarget{
			ID: uint32(audioTarget),
		},
		AudioBuffer: AudioBuffer(pcm),
	}

	if len(buffer)-audioLength == 3*4 {
		// the packet has positional audio data; 3x float32
		buffer = buffer[audioLength:]

		event.X = math.Float32frombits(binary.LittleEndian.Uint32(buffer))
		event.Y = math.Float32frombits(binary.LittleEndian.Uint32(buffer[4:]))
		event.Z = math.Float32frombits(binary.LittleEndian.Uint32(buffer[8:]))
		event.HasPosition = true
	}

	c.volatile.Lock()
	for item := c.Config.AudioListeners.head; item != nil; item = item.next {
		c.volatile.Unlock()
		ch := item.streams[user]
		if ch == nil {
			ch = make(chan *AudioPacket)
			item.streams[user] = ch
			event := AudioStreamEvent{
				Client: c,
				User:   user,
				C:      ch,
			}
			item.listener.OnAudioStream(&event)
		}
		ch <- &event
		c.volatile.Lock()
	}
	c.volatile.Unlock()

	return nil
}

func (c *Client) handleAuthenticate(buffer []byte) error {
	return errUnimplementedHandler
}

func (c *Client) handlePing(buffer []byte) error {
	var packet MumbleProto.Ping
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	atomic.AddUint32(&c.tcpPacketsReceived, 1)

	if packet.Timestamp != nil {
		diff := time.Since(time.Unix(0, int64(*packet.Timestamp)))

		index := int(c.tcpPacketsReceived) - 1
		if index >= len(c.tcpPingTimes) {
			for i := 1; i < len(c.tcpPingTimes); i++ {
				c.tcpPingTimes[i-1] = c.tcpPingTimes[i]
			}
			index = len(c.tcpPingTimes) - 1
		}

		// average is in milliseconds
		ping := float32(diff.Seconds() * 1000)
		c.tcpPingTimes[index] = ping

		var sum float32
		for i := 0; i <= index; i++ {
			sum += c.tcpPingTimes[i]
		}
		avg := sum / float32(index+1)

		sum = 0
		for i := 0; i <= index; i++ {
			sum += (avg - c.tcpPingTimes[i]) * (avg - c.tcpPingTimes[i])
		}
		variance := sum / float32(index+1)

		atomic.StoreUint32(&c.tcpPingAvg, math.Float32bits(avg))
		atomic.StoreUint32(&c.tcpPingVar, math.Float32bits(variance))
	}
	return nil
}

func (c *Client) handleReject(buffer []byte) error {
	var packet MumbleProto.Reject
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if c.State() != StateConnected {
		return errInvalidProtobuf
	}

	err := &RejectError{}

	if packet.Type != nil {
		err.Type = RejectType(*packet.Type)
	}
	if packet.Reason != nil {
		err.Reason = *packet.Reason
	}
	c.connect <- err
	c.Conn.Close()
	return nil
}

func (c *Client) handleServerSync(buffer []byte) error {
	var packet MumbleProto.ServerSync
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}
	event := ConnectEvent{
		Client: c,
	}

	if packet.Session != nil {
		{
			c.volatile.Lock()

			c.Self = c.Users[*packet.Session]

			c.volatile.Unlock()
		}
	}
	if packet.WelcomeText != nil {
		event.WelcomeMessage = packet.WelcomeText
	}
	if packet.MaxBandwidth != nil {
		val := int(*packet.MaxBandwidth)
		event.MaximumBitrate = &val
	}
	atomic.StoreUint32(&c.state, uint32(StateSynced))
	c.Config.Listeners.onConnect(&event)
	close(c.connect)
	return nil
}

func (c *Client) handleChannelRemove(buffer []byte) error {
	var packet MumbleProto.ChannelRemove
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.ChannelId == nil {
		return errIncompleteProtobuf
	}

	var channel *Channel
	{
		c.volatile.Lock()

		channelID := *packet.ChannelId
		channel = c.Channels[channelID]
		if channel == nil {
			c.volatile.Unlock()
			return errInvalidProtobuf
		}
		channel.client = nil
		delete(c.Channels, channelID)
		delete(c.permissions, channelID)
		if parent := channel.Parent; parent != nil {
			delete(parent.Children, channel.ID)
		}
		for _, link := range channel.Links {
			delete(link.Links, channelID)
		}

		c.volatile.Unlock()
	}

	if c.State() == StateSynced {
		event := ChannelChangeEvent{
			Client:  c,
			Type:    ChannelChangeRemoved,
			Channel: channel,
		}
		c.Config.Listeners.onChannelChange(&event)
	}
	return nil
}

func (c *Client) handleChannelState(buffer []byte) error {
	var packet MumbleProto.ChannelState
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.ChannelId == nil {
		return errIncompleteProtobuf
	}
	event := ChannelChangeEvent{
		Client: c,
	}

	{
		c.volatile.Lock()

		channelID := *packet.ChannelId
		channel := c.Channels[channelID]
		if channel == nil {
			channel = c.Channels.create(channelID)
			channel.client = c

			event.Type |= ChannelChangeCreated
		}
		event.Channel = channel
		if packet.Parent != nil {
			if channel.Parent != nil {
				delete(channel.Parent.Children, channelID)
			}
			newParent := c.Channels[*packet.Parent]
			if newParent != channel.Parent {
				event.Type |= ChannelChangeMoved
			}
			channel.Parent = newParent
			if channel.Parent != nil {
				channel.Parent.Children[channel.ID] = channel
			}
		}
		if packet.Name != nil {
			if *packet.Name != channel.Name {
				event.Type |= ChannelChangeName
			}
			channel.Name = *packet.Name
		}
		if packet.Links != nil {
			channel.Links = make(Channels)
			event.Type |= ChannelChangeLinks
			for _, channelID := range packet.Links {
				if c := c.Channels[channelID]; c != nil {
					channel.Links[channelID] = c
				}
			}
		}
		for _, channelID := range packet.LinksAdd {
			if c := c.Channels[channelID]; c != nil {
				event.Type |= ChannelChangeLinks
				channel.Links[channelID] = c
				c.Links[channel.ID] = channel
			}
		}
		for _, channelID := range packet.LinksRemove {
			if c := c.Channels[channelID]; c != nil {
				event.Type |= ChannelChangeLinks
				delete(channel.Links, channelID)
				delete(c.Links, channel.ID)
			}
		}
		if packet.Description != nil {
			if *packet.Description != channel.Description {
				event.Type |= ChannelChangeDescription
			}
			channel.Description = *packet.Description
			channel.DescriptionHash = nil
		}
		if packet.Temporary != nil {
			channel.Temporary = *packet.Temporary
		}
		if packet.Position != nil {
			if *packet.Position != channel.Position {
				event.Type |= ChannelChangePosition
			}
			channel.Position = *packet.Position
		}
		if packet.DescriptionHash != nil {
			event.Type |= ChannelChangeDescription
			channel.DescriptionHash = packet.DescriptionHash
			channel.Description = ""
		}
		if packet.MaxUsers != nil {
			event.Type |= ChannelChangeMaxUsers
			channel.MaxUsers = *packet.MaxUsers
		}

		c.volatile.Unlock()
	}

	if c.State() == StateSynced {
		c.Config.Listeners.onChannelChange(&event)
	}
	return nil
}

func (c *Client) handleUserRemove(buffer []byte) error {
	var packet MumbleProto.UserRemove
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.Session == nil {
		return errIncompleteProtobuf
	}
	event := UserChangeEvent{
		Client: c,
		Type:   UserChangeDisconnected,
	}

	{
		c.volatile.Lock()

		session := *packet.Session
		event.User = c.Users[session]
		if event.User == nil {
			c.volatile.Unlock()
			return errInvalidProtobuf
		}
		if packet.Actor != nil {
			event.Actor = c.Users[*packet.Actor]
			if event.Actor == nil {
				c.volatile.Unlock()
				return errInvalidProtobuf
			}
			event.Type |= UserChangeKicked
		}

		event.User.client = nil
		if event.User.Channel != nil {
			delete(event.User.Channel.Users, session)
		}
		delete(c.Users, session)
		if packet.Reason != nil {
			event.String = *packet.Reason
		}
		if packet.Ban != nil && *packet.Ban {
			event.Type |= UserChangeBanned
		}
		if event.User == c.Self {
			if packet.Ban != nil && *packet.Ban {
				c.disconnectEvent.Type = DisconnectBanned
			} else {
				c.disconnectEvent.Type = DisconnectKicked
			}
		}

		c.volatile.Unlock()
	}

	if c.State() == StateSynced {
		c.Config.Listeners.onUserChange(&event)
	}
	return nil
}

func (c *Client) handleUserState(buffer []byte) error {
	var packet MumbleProto.UserState
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.Session == nil {
		return errIncompleteProtobuf
	}
	event := UserChangeEvent{
		Client: c,
	}
	var user, actor *User
	{
		c.volatile.Lock()

		session := *packet.Session
		user = c.Users[session]
		if user == nil {
			user = c.Users.create(session)
			user.Channel = c.Channels[0]
			user.client = c

			event.Type |= UserChangeConnected

			if user.Channel == nil {
				c.volatile.Unlock()
				return errInvalidProtobuf
			}
			event.Type |= UserChangeChannel
			user.Channel.Users[session] = user
		}

		event.User = user
		if packet.Actor != nil {
			actor = c.Users[*packet.Actor]
			if actor == nil {
				c.volatile.Unlock()
				return errInvalidProtobuf
			}
			event.Actor = actor
		}
		if packet.Name != nil {
			if *packet.Name != user.Name {
				event.Type |= UserChangeName
			}
			user.Name = *packet.Name
		}
		if packet.UserId != nil {
			if *packet.UserId != user.UserID && !event.Type.Has(UserChangeConnected) {
				if *packet.UserId != math.MaxUint32 {
					event.Type |= UserChangeRegistered
					user.UserID = *packet.UserId
				} else {
					event.Type |= UserChangeUnregistered
					user.UserID = 0
				}
			} else {
				user.UserID = *packet.UserId
			}
		}
		if packet.ChannelId != nil {
			if user.Channel != nil {
				delete(user.Channel.Users, user.Session)
			}
			newChannel := c.Channels[*packet.ChannelId]
			if newChannel == nil {
				c.volatile.Lock()
				return errInvalidProtobuf
			}
			if newChannel != user.Channel {
				event.Type |= UserChangeChannel
				user.Channel = newChannel
			}
			user.Channel.Users[user.Session] = user
		}
		if packet.Mute != nil {
			if *packet.Mute != user.Muted {
				event.Type |= UserChangeAudio
			}
			user.Muted = *packet.Mute
		}
		if packet.Deaf != nil {
			if *packet.Deaf != user.Deafened {
				event.Type |= UserChangeAudio
			}
			user.Deafened = *packet.Deaf
		}
		if packet.Suppress != nil {
			if *packet.Suppress != user.Suppressed {
				event.Type |= UserChangeAudio
			}
			user.Suppressed = *packet.Suppress
		}
		if packet.SelfMute != nil {
			if *packet.SelfMute != user.SelfMuted {
				event.Type |= UserChangeAudio
			}
			user.SelfMuted = *packet.SelfMute
		}
		if packet.SelfDeaf != nil {
			if *packet.SelfDeaf != user.SelfDeafened {
				event.Type |= UserChangeAudio
			}
			user.SelfDeafened = *packet.SelfDeaf
		}
		if packet.Texture != nil {
			event.Type |= UserChangeTexture
			user.Texture = packet.Texture
			user.TextureHash = nil
		}
		if packet.Comment != nil {
			if *packet.Comment != user.Comment {
				event.Type |= UserChangeComment
			}
			user.Comment = *packet.Comment
			user.CommentHash = nil
		}
		if packet.Hash != nil {
			user.Hash = *packet.Hash
		}
		if packet.CommentHash != nil {
			event.Type |= UserChangeComment
			user.CommentHash = packet.CommentHash
			user.Comment = ""
		}
		if packet.TextureHash != nil {
			event.Type |= UserChangeTexture
			user.TextureHash = packet.TextureHash
			user.Texture = nil
		}
		if packet.PrioritySpeaker != nil {
			if *packet.PrioritySpeaker != user.PrioritySpeaker {
				event.Type |= UserChangePrioritySpeaker
			}
			user.PrioritySpeaker = *packet.PrioritySpeaker
		}
		if packet.Recording != nil {
			if *packet.Recording != user.Recording {
				event.Type |= UserChangeRecording
			}
			user.Recording = *packet.Recording
		}

		c.volatile.Unlock()
	}

	if c.State() == StateSynced {
		c.Config.Listeners.onUserChange(&event)
	}
	return nil
}

func (c *Client) handleBanList(buffer []byte) error {
	var packet MumbleProto.BanList
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	event := BanListEvent{
		Client:  c,
		BanList: make(BanList, 0, len(packet.Bans)),
	}

	for _, banPacket := range packet.Bans {
		ban := &Ban{
			Address: net.IP(banPacket.Address),
		}
		if banPacket.Mask != nil {
			size := net.IPv4len * 8
			if len(ban.Address) == net.IPv6len {
				size = net.IPv6len * 8
			}
			ban.Mask = net.CIDRMask(int(*banPacket.Mask), size)
		}
		if banPacket.Name != nil {
			ban.Name = *banPacket.Name
		}
		if banPacket.Hash != nil {
			ban.Hash = *banPacket.Hash
		}
		if banPacket.Reason != nil {
			ban.Reason = *banPacket.Reason
		}
		if banPacket.Start != nil {
			ban.Start, _ = time.Parse(time.RFC3339, *banPacket.Start)
		}
		if banPacket.Duration != nil {
			ban.Duration = time.Duration(*banPacket.Duration) * time.Second
		}
		event.BanList = append(event.BanList, ban)
	}

	c.Config.Listeners.onBanList(&event)
	return nil
}

func (c *Client) handleTextMessage(buffer []byte) error {
	var packet MumbleProto.TextMessage
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	event := TextMessageEvent{
		Client: c,
	}
	if packet.Actor != nil {
		event.Sender = c.Users[*packet.Actor]
	}
	if packet.Session != nil {
		event.Users = make([]*User, 0, len(packet.Session))
		for _, session := range packet.Session {
			if user := c.Users[session]; user != nil {
				event.Users = append(event.Users, user)
			}
		}
	}
	if packet.ChannelId != nil {
		event.Channels = make([]*Channel, 0, len(packet.ChannelId))
		for _, id := range packet.ChannelId {
			if channel := c.Channels[id]; channel != nil {
				event.Channels = append(event.Channels, channel)
			}
		}
	}
	if packet.TreeId != nil {
		event.Trees = make([]*Channel, 0, len(packet.TreeId))
		for _, id := range packet.TreeId {
			if channel := c.Channels[id]; channel != nil {
				event.Trees = append(event.Trees, channel)
			}
		}
	}
	if packet.Message != nil {
		event.Message = *packet.Message
	}

	c.Config.Listeners.onTextMessage(&event)
	return nil
}

func (c *Client) handlePermissionDenied(buffer []byte) error {
	var packet MumbleProto.PermissionDenied
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.Type == nil || *packet.Type == MumbleProto.PermissionDenied_H9K {
		return errInvalidProtobuf
	}

	event := PermissionDeniedEvent{
		Client: c,
		Type:   PermissionDeniedType(*packet.Type),
	}
	if packet.Reason != nil {
		event.String = *packet.Reason
	}
	if packet.Name != nil {
		event.String = *packet.Name
	}
	if packet.Session != nil {
		event.User = c.Users[*packet.Session]
		if event.User == nil {
			return errInvalidProtobuf
		}
	}
	if packet.ChannelId != nil {
		event.Channel = c.Channels[*packet.ChannelId]
		if event.Channel == nil {
			return errInvalidProtobuf
		}
	}
	if packet.Permission != nil {
		event.Permission = Permission(*packet.Permission)
	}

	c.Config.Listeners.onPermissionDenied(&event)
	return nil
}

func (c *Client) handleACL(buffer []byte) error {
	var packet MumbleProto.ACL
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	acl := &ACL{
		Inherits: packet.GetInheritAcls(),
	}
	if packet.ChannelId == nil {
		return errInvalidProtobuf
	}
	acl.Channel = c.Channels[*packet.ChannelId]
	if acl.Channel == nil {
		return errInvalidProtobuf
	}

	if packet.Groups != nil {
		acl.Groups = make([]*ACLGroup, 0, len(packet.Groups))
		for _, group := range packet.Groups {
			aclGroup := &ACLGroup{
				Name:         *group.Name,
				Inherited:    group.GetInherited(),
				InheritUsers: group.GetInherit(),
				Inheritable:  group.GetInheritable(),
			}
			if group.Add != nil {
				aclGroup.UsersAdd = make(map[uint32]*ACLUser)
				for _, userID := range group.Add {
					aclGroup.UsersAdd[userID] = &ACLUser{
						UserID: userID,
					}
				}
			}
			if group.Remove != nil {
				aclGroup.UsersRemove = make(map[uint32]*ACLUser)
				for _, userID := range group.Remove {
					aclGroup.UsersRemove[userID] = &ACLUser{
						UserID: userID,
					}
				}
			}
			if group.InheritedMembers != nil {
				aclGroup.UsersInherited = make(map[uint32]*ACLUser)
				for _, userID := range group.InheritedMembers {
					aclGroup.UsersInherited[userID] = &ACLUser{
						UserID: userID,
					}
				}
			}
			acl.Groups = append(acl.Groups, aclGroup)
		}
	}
	if packet.Acls != nil {
		acl.Rules = make([]*ACLRule, 0, len(packet.Acls))
		for _, rule := range packet.Acls {
			aclRule := &ACLRule{
				AppliesCurrent:  rule.GetApplyHere(),
				AppliesChildren: rule.GetApplySubs(),
				Inherited:       rule.GetInherited(),
				Granted:         Permission(rule.GetGrant()),
				Denied:          Permission(rule.GetDeny()),
			}
			if rule.UserId != nil {
				aclRule.User = &ACLUser{
					UserID: *rule.UserId,
				}
			} else if rule.Group != nil {
				var group *ACLGroup
				for _, g := range acl.Groups {
					if g.Name == *rule.Group {
						group = g
						break
					}
				}
				if group == nil {
					group = &ACLGroup{
						Name: *rule.Group,
					}
				}
				aclRule.Group = group
			}
			acl.Rules = append(acl.Rules, aclRule)
		}
	}
	c.tmpACL = acl
	return nil
}

func (c *Client) handleQueryUsers(buffer []byte) error {
	var packet MumbleProto.QueryUsers
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	acl := c.tmpACL
	if acl == nil {
		return errIncompleteProtobuf
	}
	c.tmpACL = nil

	userMap := make(map[uint32]string)
	for i := 0; i < len(packet.Ids) && i < len(packet.Names); i++ {
		userMap[packet.Ids[i]] = packet.Names[i]
	}

	for _, group := range acl.Groups {
		for _, user := range group.UsersAdd {
			user.Name = userMap[user.UserID]
		}
		for _, user := range group.UsersRemove {
			user.Name = userMap[user.UserID]
		}
		for _, user := range group.UsersInherited {
			user.Name = userMap[user.UserID]
		}
	}
	for _, rule := range acl.Rules {
		if rule.User != nil {
			rule.User.Name = userMap[rule.User.UserID]
		}
	}

	event := ACLEvent{
		Client: c,
		ACL:    acl,
	}
	c.Config.Listeners.onACL(&event)
	return nil
}

func (c *Client) handleCryptSetup(buffer []byte) error {
	return errUnimplementedHandler
}

func (c *Client) handleContextActionModify(buffer []byte) error {
	var packet MumbleProto.ContextActionModify
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.Action == nil || packet.Operation == nil {
		return errInvalidProtobuf
	}

	event := ContextActionChangeEvent{
		Client: c,
	}

	{
		c.volatile.Lock()

		switch *packet.Operation {
		case MumbleProto.ContextActionModify_Add:
			if ca := c.ContextActions[*packet.Action]; ca != nil {
				c.volatile.Unlock()
				return nil
			}
			event.Type = ContextActionAdd
			contextAction := c.ContextActions.create(*packet.Action)
			if packet.Text != nil {
				contextAction.Label = *packet.Text
			}
			if packet.Context != nil {
				contextAction.Type = ContextActionType(*packet.Context)
			}
			event.ContextAction = contextAction
		case MumbleProto.ContextActionModify_Remove:
			contextAction := c.ContextActions[*packet.Action]
			if contextAction == nil {
				c.volatile.Unlock()
				return nil
			}
			event.Type = ContextActionRemove
			delete(c.ContextActions, *packet.Action)
			event.ContextAction = contextAction
		default:
			c.volatile.Unlock()
			return errInvalidProtobuf
		}

		c.volatile.Unlock()
	}

	c.Config.Listeners.onContextActionChange(&event)
	return nil
}

func (c *Client) handleContextAction(buffer []byte) error {
	return errUnimplementedHandler
}

func (c *Client) handleUserList(buffer []byte) error {
	var packet MumbleProto.UserList
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	event := UserListEvent{
		Client:   c,
		UserList: make(RegisteredUsers, 0, len(packet.Users)),
	}

	for _, user := range packet.Users {
		registeredUser := &RegisteredUser{
			UserID: *user.UserId,
		}
		if user.Name != nil {
			registeredUser.Name = *user.Name
		}
		if user.LastSeen != nil {
			registeredUser.LastSeen, _ = time.ParseInLocation(time.RFC3339, *user.LastSeen, nil)
		}
		if user.LastChannel != nil {
			if lastChannel := c.Channels[*user.LastChannel]; lastChannel != nil {
				registeredUser.LastChannel = lastChannel
			}
		}
		event.UserList = append(event.UserList, registeredUser)
	}

	c.Config.Listeners.onUserList(&event)
	return nil
}

func (c *Client) handleVoiceTarget(buffer []byte) error {
	return errUnimplementedHandler
}

func (c *Client) handlePermissionQuery(buffer []byte) error {
	var packet MumbleProto.PermissionQuery
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	var singleChannel *Channel
	if packet.ChannelId != nil && packet.Permissions != nil {
		singleChannel = c.Channels[*packet.ChannelId]
		if singleChannel == nil {
			return errInvalidProtobuf
		}
	}

	var changedChannels []*Channel

	{
		c.volatile.Lock()

		if packet.GetFlush() {
			oldPermissions := c.permissions
			c.permissions = make(map[uint32]*Permission)
			changedChannels = make([]*Channel, 0, len(oldPermissions))
			for channelID := range oldPermissions {
				changedChannels = append(changedChannels, c.Channels[channelID])
			}
		}

		if singleChannel != nil {
			p := Permission(*packet.Permissions)
			c.permissions[singleChannel.ID] = &p
			changedChannels = append(changedChannels, singleChannel)
		}

		c.volatile.Unlock()
	}

	for _, channel := range changedChannels {
		event := ChannelChangeEvent{
			Client:  c,
			Type:    ChannelChangePermission,
			Channel: channel,
		}
		c.Config.Listeners.onChannelChange(&event)
	}

	return nil
}

func (c *Client) handleCodecVersion(buffer []byte) error {
	var packet MumbleProto.CodecVersion
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}
	event := ServerConfigEvent{
		Client: c,
	}
	event.CodecAlpha = packet.Alpha
	event.CodecBeta = packet.Beta
	{
		val := packet.GetPreferAlpha()
		event.CodecPreferAlpha = &val
	}
	{
		val := packet.GetOpus()
		event.CodecOpus = &val
	}

	var codec AudioCodec
	switch {
	case *event.CodecOpus:
		codec = getAudioCodec(audioCodecIDOpus)
	}
	if codec != nil {
		c.audioCodec = codec

		{
			c.volatile.Lock()

			c.AudioEncoder = codec.NewEncoder()

			c.volatile.Unlock()
		}
	}

	c.Config.Listeners.onServerConfig(&event)
	return nil
}

func (c *Client) handleUserStats(buffer []byte) error {
	var packet MumbleProto.UserStats
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}

	if packet.Session == nil {
		return errIncompleteProtobuf
	}
	user := c.Users[*packet.Session]
	if user == nil {
		return errInvalidProtobuf
	}

	{
		c.volatile.Lock()

		if user.Stats == nil {
			user.Stats = &UserStats{}
		}
		*user.Stats = UserStats{
			User: user,
		}
		stats := user.Stats

		if packet.FromClient != nil {
			if packet.FromClient.Good != nil {
				stats.FromClient.Good = *packet.FromClient.Good
			}
			if packet.FromClient.Late != nil {
				stats.FromClient.Late = *packet.FromClient.Late
			}
			if packet.FromClient.Lost != nil {
				stats.FromClient.Lost = *packet.FromClient.Lost
			}
			if packet.FromClient.Resync != nil {
				stats.FromClient.Resync = *packet.FromClient.Resync
			}
		}
		if packet.FromServer != nil {
			if packet.FromServer.Good != nil {
				stats.FromServer.Good = *packet.FromServer.Good
			}
			if packet.FromClient.Late != nil {
				stats.FromServer.Late = *packet.FromServer.Late
			}
			if packet.FromClient.Lost != nil {
				stats.FromServer.Lost = *packet.FromServer.Lost
			}
			if packet.FromClient.Resync != nil {
				stats.FromServer.Resync = *packet.FromServer.Resync
			}
		}

		if packet.UdpPackets != nil {
			stats.UDPPackets = *packet.UdpPackets
		}
		if packet.UdpPingAvg != nil {
			stats.UDPPingAverage = *packet.UdpPingAvg
		}
		if packet.UdpPingVar != nil {
			stats.UDPPingVariance = *packet.UdpPingVar
		}
		if packet.TcpPackets != nil {
			stats.TCPPackets = *packet.TcpPackets
		}
		if packet.TcpPingAvg != nil {
			stats.TCPPingAverage = *packet.TcpPingAvg
		}
		if packet.TcpPingVar != nil {
			stats.TCPPingVariance = *packet.TcpPingVar
		}

		if packet.Version != nil {
			stats.Version = parseVersion(packet.Version)
		}
		if packet.Onlinesecs != nil {
			stats.Connected = time.Now().Add(time.Duration(*packet.Onlinesecs) * -time.Second)
		}
		if packet.Idlesecs != nil {
			stats.Idle = time.Duration(*packet.Idlesecs) * time.Second
		}
		if packet.Bandwidth != nil {
			stats.Bandwidth = int(*packet.Bandwidth)
		}
		if packet.Address != nil {
			stats.IP = net.IP(packet.Address)
		}
		if packet.Certificates != nil {
			stats.Certificates = make([]*x509.Certificate, 0, len(packet.Certificates))
			for _, data := range packet.Certificates {
				if data != nil {
					if cert, err := x509.ParseCertificate(data); err == nil {
						stats.Certificates = append(stats.Certificates, cert)
					}
				}
			}
		}
		stats.StrongCertificate = packet.GetStrongCertificate()
		stats.CELTVersions = packet.GetCeltVersions()
		if packet.Opus != nil {
			stats.Opus = *packet.Opus
		}

		c.volatile.Unlock()
	}

	event := UserChangeEvent{
		Client: c,
		Type:   UserChangeStats,
		User:   user,
	}

	c.Config.Listeners.onUserChange(&event)
	return nil
}

func (c *Client) handleRequestBlob(buffer []byte) error {
	return errUnimplementedHandler
}

func (c *Client) handleServerConfig(buffer []byte) error {
	var packet MumbleProto.ServerConfig
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}
	event := ServerConfigEvent{
		Client: c,
	}
	if packet.MaxBandwidth != nil {
		val := int(*packet.MaxBandwidth)
		event.MaximumBitrate = &val
	}
	if packet.WelcomeText != nil {
		event.WelcomeMessage = packet.WelcomeText
	}
	if packet.AllowHtml != nil {
		event.AllowHTML = packet.AllowHtml
	}
	if packet.MessageLength != nil {
		val := int(*packet.MessageLength)
		event.MaximumMessageLength = &val
	}
	if packet.ImageMessageLength != nil {
		val := int(*packet.ImageMessageLength)
		event.MaximumImageMessageLength = &val
	}
	if packet.MaxUsers != nil {
		val := int(*packet.MaxUsers)
		event.MaximumUsers = &val
	}
	c.Config.Listeners.onServerConfig(&event)
	return nil
}

func (c *Client) handleSuggestConfig(buffer []byte) error {
	var packet MumbleProto.SuggestConfig
	if err := proto.Unmarshal(buffer, &packet); err != nil {
		return err
	}
	event := ServerConfigEvent{
		Client: c,
	}
	if packet.Version != nil {
		event.SuggestVersion = &Version{
			Version: packet.GetVersion(),
		}
	}
	if packet.Positional != nil {
		event.SuggestPositional = packet.Positional
	}
	if packet.PushToTalk != nil {
		event.SuggestPushToTalk = packet.PushToTalk
	}
	c.Config.Listeners.onServerConfig(&event)
	return nil
}
