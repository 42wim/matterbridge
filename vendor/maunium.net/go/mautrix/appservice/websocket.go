// Copyright (c) 2021 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package appservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/gorilla/websocket"

	"maunium.net/go/mautrix/event"
)

type ErrorResponse struct {
	ErrorCode ErrorCode `json:"errcode"`
	Error     string    `json:"error"`
}

type WebsocketCommand struct {
	ReqID   int             `json:"id,omitempty"`
	Command string          `json:"command"`
	Data    json.RawMessage `json:"data"`
}

type WebsocketTransaction struct {
	Status string `json:"status"`
	TxnID  string `json:"txn_id"`
	EventList
}

type WebsocketMessage struct {
	WebsocketTransaction
	WebsocketCommand
}

type MeowWebsocketCloseCode string

const (
	MeowServerShuttingDown MeowWebsocketCloseCode = "server_shutting_down"
	MeowConnectionReplaced MeowWebsocketCloseCode = "conn_replaced"
)

var (
	WebsocketManualStop = errors.New("the websocket was disconnected manually")
	WebsocketOverridden = errors.New("a new call to StartWebsocket overrode the previous connection")
	WebsocketUnknownError = errors.New("an unknown error occurred")
)

func (mwcc MeowWebsocketCloseCode) String() string {
	switch mwcc {
	case MeowServerShuttingDown:
		return "the server is shutting down"
	case MeowConnectionReplaced:
		return "the connection was replaced by another client"
	default:
		return string(mwcc)
	}
}

type CloseCommand struct {
	Code    int                    `json:"-"`
	Command string                 `json:"command"`
	Status  MeowWebsocketCloseCode `json:"status"`
}

func (cc CloseCommand) Error() string {
	return fmt.Sprintf("websocket: close %d: %s", cc.Code, cc.Status.String())
}

func parseCloseError(err error) error {
	closeError := &websocket.CloseError{}
	if !errors.As(err, &closeError) {
		return err
	}
	var closeCommand CloseCommand
	closeCommand.Code = closeError.Code
	closeCommand.Command = "disconnect"
	if len(closeError.Text) > 0 {
		jsonErr := json.Unmarshal([]byte(closeError.Text), &closeCommand)
		if jsonErr != nil {
			return err
		}
	}
	if len(closeCommand.Status) == 0 {
		if closeCommand.Code == 4001 {
			closeCommand.Status = MeowConnectionReplaced
		} else if closeCommand.Code == websocket.CloseServiceRestart {
			closeCommand.Status = MeowServerShuttingDown
		}
	}
	return &closeCommand
}

func (as *AppService) SendWebsocket(cmd WebsocketCommand) error {
	if as.ws == nil {
		return errors.New("websocket not connected")
	}
	return as.ws.WriteJSON(&cmd)
}

func (as *AppService) consumeWebsocket(stopFunc func(error), ws *websocket.Conn) {
	defer stopFunc(WebsocketUnknownError)
	for {
		var msg WebsocketMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			as.Log.Debugln("Error reading from websocket:", err)
			stopFunc(parseCloseError(err))
			return
		}
		if msg.Command == "" || msg.Command == "transaction" {
			if as.Registration.EphemeralEvents && msg.EphemeralEvents != nil {
				as.handleEvents(msg.EphemeralEvents, event.EphemeralEventType)
			}
			as.handleEvents(msg.Events, event.UnknownEventType)
		} else if msg.Command == "connect" {
			as.Log.Debugln("Websocket connect confirmation received")
		} else {
			select {
			case as.WebsocketCommands <- msg.WebsocketCommand:
			default:
				as.Log.Warnln("Dropping websocket command %s %d / %s", msg.Command, msg.ReqID, msg.Data)
			}
		}
	}
}

func (as *AppService) StartWebsocket(baseURL string, onConnect func()) error {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}
	parsed.Path = filepath.Join(parsed.Path, "_matrix/client/unstable/fi.mau.as_sync")
	if parsed.Scheme == "http" {
		parsed.Scheme = "ws"
	} else if parsed.Scheme == "https" {
		parsed.Scheme = "wss"
	}
	ws, resp, err := websocket.DefaultDialer.Dial(parsed.String(), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", as.Registration.AppToken)},
		"User-Agent":    []string{as.BotClient().UserAgent},
	})
	if resp != nil && resp.StatusCode >= 400 {
		var errResp ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		if err != nil {
			return fmt.Errorf("websocket request returned HTTP %d with non-JSON body", resp.StatusCode)
		} else {
			return fmt.Errorf("websocket request returned %s (HTTP %d): %s", errResp.ErrorCode, resp.StatusCode, errResp.Error)
		}
	} else if err != nil {
		return fmt.Errorf("failed to open websocket: %w", err)
	}
	if as.StopWebsocket != nil {
		as.StopWebsocket(WebsocketOverridden)
	}
	closeChan := make(chan error)
	closeChanSync := sync.Once{}
	stopFunc := func(err error) {
		closeChanSync.Do(func() {
			closeChan <- err
		})
	}
	as.ws = ws
	as.StopWebsocket = stopFunc
	as.PrepareWebsocket()
	as.Log.Debugln("Appservice transaction websocket connected")

	go as.consumeWebsocket(stopFunc, ws)

	if onConnect != nil {
		onConnect()
	}

	closeErr := <-closeChan

	err = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, ""))
	if err != nil && err != websocket.ErrCloseSent {
		as.Log.Warnln("Error writing close message to websocket:", err)
	}
	err = ws.Close()
	if err != nil {
		as.Log.Warnln("Error closing websocket:", err)
	}
	return closeErr
}
