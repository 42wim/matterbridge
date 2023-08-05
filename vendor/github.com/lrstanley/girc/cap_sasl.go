// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package girc

import (
	"encoding/base64"
	"fmt"
)

// SASLMech is an representation of what a SASL mechanism should support.
// See SASLExternal and SASLPlain for implementations of this.
type SASLMech interface {
	// Method returns the uppercase version of the SASL mechanism name.
	Method() string
	// Encode returns the response that the SASL mechanism wants to use. If
	// the returned string is empty (e.g. the mechanism gives up), the handler
	// will attempt to panic, as expectation is that if SASL authentication
	// fails, the client will disconnect.
	Encode(params []string) (output string)
}

// SASLExternal implements the "EXTERNAL" SASL type.
type SASLExternal struct {
	// Identity is an optional field which allows the client to specify
	// pre-authentication identification. This means that EXTERNAL will
	// supply this in the initial response. This usually isn't needed (e.g.
	// CertFP).
	Identity string `json:"identity"`
}

// Method identifies what type of SASL this implements.
func (sasl *SASLExternal) Method() string {
	return "EXTERNAL"
}

// Encode for external SALS authentication should really only return a "+",
// unless the user has specified pre-authentication or identification data.
// See https://tools.ietf.org/html/rfc4422#appendix-A for more info.
func (sasl *SASLExternal) Encode(params []string) string {
	if len(params) != 1 || params[0] != "+" {
		return ""
	}

	if sasl.Identity != "" {
		return sasl.Identity
	}

	return "+"
}

// SASLPlain contains the user and password needed for PLAIN SASL authentication.
type SASLPlain struct {
	User string `json:"user"` // User is the username for SASL.
	Pass string `json:"pass"` // Pass is the password for SASL.
}

// Method identifies what type of SASL this implements.
func (sasl *SASLPlain) Method() string {
	return "PLAIN"
}

// Encode encodes the plain user+password into a SASL PLAIN implementation.
// See https://tools.ietf.org/rfc/rfc4422.txt for more info.
func (sasl *SASLPlain) Encode(params []string) string {
	if len(params) != 1 || params[0] != "+" {
		return ""
	}

	in := []byte(sasl.User)

	in = append(in, 0x0)
	in = append(in, []byte(sasl.User)...)
	in = append(in, 0x0)
	in = append(in, []byte(sasl.Pass)...)

	return base64.StdEncoding.EncodeToString(in)
}

const saslChunkSize = 400

func handleSASL(c *Client, e Event) {
	if e.Command == RPL_SASLSUCCESS || e.Command == ERR_SASLALREADY {
		// Let the server know that we're done.
		c.write(&Event{Command: CAP, Params: []string{CAP_END}})
		return
	}

	// Assume they want us to handle sending auth.
	auth := c.Config.SASL.Encode(e.Params)

	if auth == "" {
		// Assume the SASL authentication method doesn't want to respond for
		// some reason. The SASL spec and IRCv3 spec do not define a clear
		// way to abort a SASL exchange, other than to disconnect, or proceed
		// with CAP END.
		c.receive(&Event{Command: ERROR, Params: []string{
			fmt.Sprintf("closing connection: SASL %s failed: %s", c.Config.SASL.Method(), e.Last()),
		}})
		return
	}

	// Send in "saslChunkSize"-length byte chunks. If the last chuck is
	// exactly "saslChunkSize" bytes, send a "AUTHENTICATE +" 0-byte
	// acknowledgement response to let the server know that we're done.
	for {
		if len(auth) > saslChunkSize {
			c.write(&Event{Command: AUTHENTICATE, Params: []string{auth[0 : saslChunkSize-1]}, Sensitive: true})
			auth = auth[saslChunkSize:]
			continue
		}

		if len(auth) <= saslChunkSize {
			c.write(&Event{Command: AUTHENTICATE, Params: []string{auth}, Sensitive: true})

			if len(auth) == 400 {
				c.write(&Event{Command: AUTHENTICATE, Params: []string{"+"}})
			}
			break
		}
	}
}

func handleSASLError(c *Client, e Event) {
	if c.Config.SASL == nil {
		c.write(&Event{Command: CAP, Params: []string{CAP_END}})
		return
	}

	// Authentication failed. The SASL spec and IRCv3 spec do not define a
	// clear way to abort a SASL exchange, other than to disconnect, or
	// proceed with CAP END.
	c.receive(&Event{Command: ERROR, Params: []string{"closing connection: " + e.Last()}})
}
