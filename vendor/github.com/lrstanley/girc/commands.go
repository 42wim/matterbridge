// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package girc

import (
	"errors"
	"fmt"
)

// Commands holds a large list of useful methods to interact with the server,
// and wrappers for common events.
type Commands struct {
	c *Client
}

// Nick changes the client nickname.
func (cmd *Commands) Nick(name string) error {
	if !IsValidNick(name) {
		return &ErrInvalidTarget{Target: name}
	}

	cmd.c.Send(&Event{Command: NICK, Params: []string{name}})
	return nil
}

// Join attempts to enter a list of IRC channels, at bulk if possible to
// prevent sending extensive JOIN commands.
func (cmd *Commands) Join(channels ...string) error {
	// We can join multiple channels at once, however we need to ensure that
	// we are not exceeding the line length. (see maxLength)
	max := maxLength - len(JOIN) - 1

	var buffer string

	for i := 0; i < len(channels); i++ {
		if !IsValidChannel(channels[i]) {
			return &ErrInvalidTarget{Target: channels[i]}
		}

		if len(buffer+","+channels[i]) > max {
			cmd.c.Send(&Event{Command: JOIN, Params: []string{buffer}})
			buffer = ""
			continue
		}

		if len(buffer) == 0 {
			buffer = channels[i]
		} else {
			buffer += "," + channels[i]
		}

		if i == len(channels)-1 {
			cmd.c.Send(&Event{Command: JOIN, Params: []string{buffer}})
			return nil
		}
	}

	return nil
}

// JoinKey attempts to enter an IRC channel with a password.
func (cmd *Commands) JoinKey(channel, password string) error {
	if !IsValidChannel(channel) {
		return &ErrInvalidTarget{Target: channel}
	}

	cmd.c.Send(&Event{Command: JOIN, Params: []string{channel, password}})
	return nil
}

// Part leaves an IRC channel.
func (cmd *Commands) Part(channel, message string) error {
	if !IsValidChannel(channel) {
		return &ErrInvalidTarget{Target: channel}
	}

	cmd.c.Send(&Event{Command: JOIN, Params: []string{channel}})
	return nil
}

// PartMessage leaves an IRC channel with a specified leave message.
func (cmd *Commands) PartMessage(channel, message string) error {
	if !IsValidChannel(channel) {
		return &ErrInvalidTarget{Target: channel}
	}

	cmd.c.Send(&Event{Command: JOIN, Params: []string{channel}, Trailing: message})
	return nil
}

// SendCTCP sends a CTCP request to target. Note that this method uses
// PRIVMSG specifically.
func (cmd *Commands) SendCTCP(target, ctcpType, message string) error {
	out := encodeCTCPRaw(ctcpType, message)
	if out == "" {
		return errors.New("invalid CTCP")
	}

	return cmd.Message(target, out)
}

// SendCTCPf sends a CTCP request to target using a specific format. Note that
// this method uses PRIVMSG specifically.
func (cmd *Commands) SendCTCPf(target, ctcpType, format string, a ...interface{}) error {
	return cmd.SendCTCP(target, ctcpType, fmt.Sprintf(format, a...))
}

// SendCTCPReplyf sends a CTCP response to target using a specific format.
// Note that this method uses NOTICE specifically.
func (cmd *Commands) SendCTCPReplyf(target, ctcpType, format string, a ...interface{}) error {
	return cmd.SendCTCPReply(target, ctcpType, fmt.Sprintf(format, a...))
}

// SendCTCPReply sends a CTCP response to target. Note that this method uses
// NOTICE specifically.
func (cmd *Commands) SendCTCPReply(target, ctcpType, message string) error {
	out := encodeCTCPRaw(ctcpType, message)
	if out == "" {
		return errors.New("invalid CTCP")
	}

	return cmd.Notice(target, out)
}

// Message sends a PRIVMSG to target (either channel, service, or user).
func (cmd *Commands) Message(target, message string) error {
	if !IsValidNick(target) && !IsValidChannel(target) {
		return &ErrInvalidTarget{Target: target}
	}

	cmd.c.Send(&Event{Command: PRIVMSG, Params: []string{target}, Trailing: message})
	return nil
}

// Messagef sends a formated PRIVMSG to target (either channel, service, or
// user).
func (cmd *Commands) Messagef(target, format string, a ...interface{}) error {
	return cmd.Message(target, fmt.Sprintf(format, a...))
}

// ErrInvalidSource is returned when a method needs to know the origin of an
// event, however Event.Source is unknown (e.g. sent by the user, not the
// server.)
var ErrInvalidSource = errors.New("event has nil or invalid source address")

// Reply sends a reply to channel or user, based on where the supplied event
// originated from. See also ReplyTo().
func (cmd *Commands) Reply(event Event, message string) error {
	if event.Source == nil {
		return ErrInvalidSource
	}

	if len(event.Params) > 0 && IsValidChannel(event.Params[0]) {
		return cmd.Message(event.Params[0], message)
	}

	return cmd.Message(event.Source.Name, message)
}

// Replyf sends a reply to channel or user with a format string, based on
// where the supplied event originated from. See also ReplyTof().
func (cmd *Commands) Replyf(event Event, format string, a ...interface{}) error {
	return cmd.Reply(event, fmt.Sprintf(format, a...))
}

// ReplyTo sends a reply to a channel or user, based on where the supplied
// event originated from. ReplyTo(), when originating from a channel will
// default to replying with "<user>, <message>". See also Reply().
func (cmd *Commands) ReplyTo(event Event, message string) error {
	if event.Source == nil {
		return ErrInvalidSource
	}

	if len(event.Params) > 0 && IsValidChannel(event.Params[0]) {
		return cmd.Message(event.Params[0], event.Source.Name+", "+message)
	}

	return cmd.Message(event.Source.Name, message)
}

// ReplyTof sends a reply to a channel or user with a format string, based
// on where the supplied event originated from. ReplyTo(), when originating
// from a channel will default to replying with "<user>, <message>". See
// also Replyf().
func (cmd *Commands) ReplyTof(event Event, format string, a ...interface{}) error {
	return cmd.ReplyTo(event, fmt.Sprintf(format, a...))
}

// Action sends a PRIVMSG ACTION (/me) to target (either channel, service,
// or user).
func (cmd *Commands) Action(target, message string) error {
	if !IsValidNick(target) && !IsValidChannel(target) {
		return &ErrInvalidTarget{Target: target}
	}

	cmd.c.Send(&Event{
		Command:  PRIVMSG,
		Params:   []string{target},
		Trailing: fmt.Sprintf("\001ACTION %s\001", message),
	})
	return nil
}

// Actionf sends a formated PRIVMSG ACTION (/me) to target (either channel,
// service, or user).
func (cmd *Commands) Actionf(target, format string, a ...interface{}) error {
	return cmd.Action(target, fmt.Sprintf(format, a...))
}

// Notice sends a NOTICE to target (either channel, service, or user).
func (cmd *Commands) Notice(target, message string) error {
	if !IsValidNick(target) && !IsValidChannel(target) {
		return &ErrInvalidTarget{Target: target}
	}

	cmd.c.Send(&Event{Command: NOTICE, Params: []string{target}, Trailing: message})
	return nil
}

// Noticef sends a formated NOTICE to target (either channel, service, or
// user).
func (cmd *Commands) Noticef(target, format string, a ...interface{}) error {
	return cmd.Notice(target, fmt.Sprintf(format, a...))
}

// SendRaw sends a raw string back to the server, without carriage returns
// or newlines.
func (cmd *Commands) SendRaw(raw string) error {
	e := ParseEvent(raw)
	if e == nil {
		return errors.New("invalid event: " + raw)
	}

	cmd.c.Send(e)
	return nil
}

// SendRawf sends a formated string back to the server, without carriage
// returns or newlines.
func (cmd *Commands) SendRawf(format string, a ...interface{}) error {
	return cmd.SendRaw(fmt.Sprintf(format, a...))
}

// Topic sets the topic of channel to message. Does not verify the length
// of the topic.
func (cmd *Commands) Topic(channel, message string) {
	cmd.c.Send(&Event{Command: TOPIC, Params: []string{channel}, Trailing: message})
}

// Who sends a WHO query to the server, which will attempt WHOX by default.
// See http://faerion.sourceforge.net/doc/irc/whox.var for more details. This
// sends "%tcuhnr,2" per default. Do not use "1" as this will conflict with
// girc's builtin tracking functionality.
func (cmd *Commands) Who(target string) error {
	if !IsValidNick(target) && !IsValidChannel(target) && !IsValidUser(target) {
		return &ErrInvalidTarget{Target: target}
	}

	cmd.c.Send(&Event{Command: WHO, Params: []string{target, "%tcuhnr,2"}})
	return nil
}

// Whois sends a WHOIS query to the server, targeted at a specific user.
// as WHOIS is a bit slower, you may want to use WHO for brief user info.
func (cmd *Commands) Whois(nick string) error {
	if !IsValidNick(nick) {
		return &ErrInvalidTarget{Target: nick}
	}

	cmd.c.Send(&Event{Command: WHOIS, Params: []string{nick}})
	return nil
}

// Ping sends a PING query to the server, with a specific identifier that
// the server should respond with.
func (cmd *Commands) Ping(id string) {
	cmd.c.write(&Event{Command: PING, Params: []string{id}})
}

// Pong sends a PONG query to the server, with an identifier which was
// received from a previous PING query received by the client.
func (cmd *Commands) Pong(id string) {
	cmd.c.write(&Event{Command: PONG, Params: []string{id}})
}

// Oper sends a OPER authentication query to the server, with a username
// and password.
func (cmd *Commands) Oper(user, pass string) {
	cmd.c.Send(&Event{Command: OPER, Params: []string{user, pass}, Sensitive: true})
}

// Kick sends a KICK query to the server, attempting to kick nick from
// channel, with reason. If reason is blank, one will not be sent to the
// server.
func (cmd *Commands) Kick(channel, nick, reason string) error {
	if !IsValidChannel(channel) {
		return &ErrInvalidTarget{Target: channel}
	}

	if !IsValidNick(nick) {
		return &ErrInvalidTarget{Target: nick}
	}

	if reason != "" {
		cmd.c.Send(&Event{Command: KICK, Params: []string{channel, nick}, Trailing: reason})
		return nil
	}

	cmd.c.Send(&Event{Command: KICK, Params: []string{channel, nick}})
	return nil
}

// Invite sends a INVITE query to the server, to invite nick to channel.
func (cmd *Commands) Invite(channel, nick string) error {
	if !IsValidChannel(channel) {
		return &ErrInvalidTarget{Target: channel}
	}

	if !IsValidNick(nick) {
		return &ErrInvalidTarget{Target: nick}
	}

	cmd.c.Send(&Event{Command: INVITE, Params: []string{nick, channel}})
	return nil
}

// Away sends a AWAY query to the server, suggesting that the client is no
// longer active. If reason is blank, Client.Back() is called. Also see
// Client.Back().
func (cmd *Commands) Away(reason string) {
	if reason == "" {
		cmd.Back()
		return
	}

	cmd.c.Send(&Event{Command: AWAY, Params: []string{reason}})
}

// Back sends a AWAY query to the server, however the query is blank,
// suggesting that the client is active once again. Also see Client.Away().
func (cmd *Commands) Back() {
	cmd.c.Send(&Event{Command: AWAY})
}

// List sends a LIST query to the server, which will list channels and topics.
// Supports multiple channels at once, in hopes it will reduce extensive
// LIST queries to the server. Supply no channels to run a list against the
// entire server (warning, that may mean LOTS of channels!)
func (cmd *Commands) List(channels ...string) error {
	if len(channels) == 0 {
		cmd.c.Send(&Event{Command: LIST})
		return nil
	}

	// We can LIST multiple channels at once, however we need to ensure that
	// we are not exceeding the line length. (see maxLength)
	max := maxLength - len(JOIN) - 1

	var buffer string

	for i := 0; i < len(channels); i++ {
		if !IsValidChannel(channels[i]) {
			return &ErrInvalidTarget{Target: channels[i]}
		}

		if len(buffer+","+channels[i]) > max {
			cmd.c.Send(&Event{Command: LIST, Params: []string{buffer}})
			buffer = ""
			continue
		}

		if len(buffer) == 0 {
			buffer = channels[i]
		} else {
			buffer += "," + channels[i]
		}

		if i == len(channels)-1 {
			cmd.c.Send(&Event{Command: LIST, Params: []string{buffer}})
			return nil
		}
	}

	return nil
}

// Whowas sends a WHOWAS query to the server. amount is the amount of results
// you want back.
func (cmd *Commands) Whowas(nick string, amount int) error {
	if !IsValidNick(nick) {
		return &ErrInvalidTarget{Target: nick}
	}

	cmd.c.Send(&Event{Command: WHOWAS, Params: []string{nick, string(amount)}})
	return nil
}
