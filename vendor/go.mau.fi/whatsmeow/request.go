// Copyright (c) 2021 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package whatsmeow

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
)

func (cli *Client) generateRequestID() string {
	return cli.uniqueID + strconv.FormatUint(uint64(atomic.AddUint32(&cli.idCounter, 1)), 10)
}

var closedNode = &waBinary.Node{Tag: "xmlstreamend"}

func (cli *Client) clearResponseWaiters() {
	cli.responseWaitersLock.Lock()
	for _, waiter := range cli.responseWaiters {
		select {
		case waiter <- closedNode:
		default:
			close(waiter)
		}
	}
	cli.responseWaiters = make(map[string]chan<- *waBinary.Node)
	cli.responseWaitersLock.Unlock()
}

func (cli *Client) waitResponse(reqID string) chan *waBinary.Node {
	ch := make(chan *waBinary.Node, 1)
	cli.responseWaitersLock.Lock()
	cli.responseWaiters[reqID] = ch
	cli.responseWaitersLock.Unlock()
	return ch
}

func (cli *Client) cancelResponse(reqID string, ch chan *waBinary.Node) {
	cli.responseWaitersLock.Lock()
	close(ch)
	delete(cli.responseWaiters, reqID)
	cli.responseWaitersLock.Unlock()
}

func (cli *Client) receiveResponse(data *waBinary.Node) bool {
	id, ok := data.Attrs["id"].(string)
	if !ok || (data.Tag != "iq" && data.Tag != "ack") {
		return false
	}
	cli.responseWaitersLock.Lock()
	waiter, ok := cli.responseWaiters[id]
	if !ok {
		cli.responseWaitersLock.Unlock()
		return false
	}
	delete(cli.responseWaiters, id)
	cli.responseWaitersLock.Unlock()
	waiter <- data
	return true
}

type infoQueryType string

const (
	iqSet infoQueryType = "set"
	iqGet infoQueryType = "get"
)

type infoQuery struct {
	Namespace string
	Type      infoQueryType
	To        types.JID
	ID        string
	Content   interface{}

	Timeout time.Duration
	Context context.Context
}

func (cli *Client) sendIQAsync(query infoQuery) (<-chan *waBinary.Node, error) {
	if len(query.ID) == 0 {
		query.ID = cli.generateRequestID()
	}
	waiter := cli.waitResponse(query.ID)
	attrs := waBinary.Attrs{
		"id":    query.ID,
		"xmlns": query.Namespace,
		"type":  string(query.Type),
	}
	if !query.To.IsEmpty() {
		attrs["to"] = query.To
	}
	err := cli.sendNode(waBinary.Node{
		Tag:     "iq",
		Attrs:   attrs,
		Content: query.Content,
	})
	if err != nil {
		cli.cancelResponse(query.ID, waiter)
		return nil, err
	}
	return waiter, nil
}

func (cli *Client) sendIQ(query infoQuery) (*waBinary.Node, error) {
	resChan, err := cli.sendIQAsync(query)
	if err != nil {
		return nil, err
	}
	if query.Timeout == 0 {
		query.Timeout = 1 * time.Minute
	}
	if query.Context == nil {
		query.Context = context.Background()
	}
	select {
	case res := <-resChan:
		if res == closedNode {
			return nil, ErrIQDisconnected
		}
		resType, _ := res.Attrs["type"].(string)
		if res.Tag != "iq" || (resType != "result" && resType != "error") {
			return res, &IQError{RawNode: res}
		} else if resType == "error" {
			return res, parseIQError(res)
		}
		return res, nil
	case <-query.Context.Done():
		return nil, query.Context.Err()
	case <-time.After(query.Timeout):
		return nil, ErrIQTimedOut
	}
}
