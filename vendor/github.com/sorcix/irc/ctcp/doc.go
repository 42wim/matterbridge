// Copyright 2014 Vic Demuzere
//
// Use of this source code is governed by the MIT license.

// Package ctcp implements partial support for the Client-to-Client Protocol.
//
// CTCP defines extended messages using the standard PRIVMSG and NOTICE
// commands in IRC. This means that any CTCP messages are embedded inside the
// normal message text. Clients that don't support CTCP simply show
// the encoded message to the user.
//
// Most IRC clients support only a subset of the protocol, and only a few
// commands are actually used. This package aims to implement the most basic
// CTCP messages: a single command per IRC message. Quoting is not supported.
//
// Example using the irc.Message type:
//
//    m := irc.ParseMessage(...)
//
//    if tag, text, ok := ctcp.Decode(m.Trailing); ok {
//        // This is a CTCP message.
//    } else {
//        // This is not a CTCP message.
//    }
//
// Similar, for encoding messages:
//
//    m.Trailing = ctcp.Encode("ACTION","wants a cookie!")
//
// Do not send a complete IRC message to Decode, it won't work.
package ctcp
