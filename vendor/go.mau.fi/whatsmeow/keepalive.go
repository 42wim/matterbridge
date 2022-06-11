// Copyright (c) 2021 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package whatsmeow

import (
	"context"
	"math/rand"
	"time"

	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

var (
	// KeepAliveResponseDeadline specifies the duration to wait for a response to websocket keepalive pings.
	KeepAliveResponseDeadline = 10 * time.Second
	// KeepAliveIntervalMin specifies the minimum interval for websocket keepalive pings.
	KeepAliveIntervalMin = 20 * time.Second
	// KeepAliveIntervalMax specifies the maximum interval for websocket keepalive pings.
	KeepAliveIntervalMax = 30 * time.Second
)

func (cli *Client) keepAliveLoop(ctx context.Context) {
	var lastSuccess time.Time
	var errorCount int
	for {
		interval := rand.Int63n(KeepAliveIntervalMax.Milliseconds()-KeepAliveIntervalMin.Milliseconds()) + KeepAliveIntervalMin.Milliseconds()
		select {
		case <-time.After(time.Duration(interval) * time.Millisecond):
			isSuccess, shouldContinue := cli.sendKeepAlive(ctx)
			if !shouldContinue {
				return
			} else if !isSuccess {
				errorCount++
				go cli.dispatchEvent(&events.KeepAliveTimeout{
					ErrorCount:  errorCount,
					LastSuccess: lastSuccess,
				})
			} else {
				if errorCount > 0 {
					errorCount = 0
					go cli.dispatchEvent(&events.KeepAliveRestored{})
				}
				lastSuccess = time.Now()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (cli *Client) sendKeepAlive(ctx context.Context) (isSuccess, shouldContinue bool) {
	respCh, err := cli.sendIQAsync(infoQuery{
		Namespace: "w:p",
		Type:      "get",
		To:        types.ServerJID,
		Content:   []waBinary.Node{{Tag: "ping"}},
	})
	if err != nil {
		cli.Log.Warnf("Failed to send keepalive: %v", err)
		return false, true
	}
	select {
	case <-respCh:
		// All good
		return true, true
	case <-time.After(KeepAliveResponseDeadline):
		cli.Log.Warnf("Keepalive timed out")
		return false, true
	case <-ctx.Done():
		return false, false
	}
}
