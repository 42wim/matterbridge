// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package appservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type WebsocketRequest struct {
	ReqID   int         `json:"id,omitempty"`
	Command string      `json:"command"`
	Data    interface{} `json:"data"`

	Deadline time.Duration `json:"-"`
}

type WebsocketCommand struct {
	ReqID   int             `json:"id,omitempty"`
	Command string          `json:"command"`
	Data    json.RawMessage `json:"data"`

	Ctx context.Context `json:"-"`
}

func (wsc *WebsocketCommand) MakeResponse(ok bool, data interface{}) *WebsocketRequest {
	if wsc.ReqID == 0 || wsc.Command == "response" || wsc.Command == "error" {
		return nil
	}
	cmd := "response"
	if !ok {
		cmd = "error"
	}
	if err, isError := data.(error); isError {
		var errorData json.RawMessage
		var jsonErr error
		unwrappedErr := err
		var prefixMessage string
		for unwrappedErr != nil {
			errorData, jsonErr = json.Marshal(unwrappedErr)
			if errorData != nil && len(errorData) > 2 && jsonErr == nil {
				prefixMessage = strings.Replace(err.Error(), unwrappedErr.Error(), "", 1)
				prefixMessage = strings.TrimRight(prefixMessage, ": ")
				break
			}
			unwrappedErr = errors.Unwrap(unwrappedErr)
		}
		if errorData != nil {
			if !gjson.GetBytes(errorData, "message").Exists() {
				errorData, _ = sjson.SetBytes(errorData, "message", err.Error())
			} // else: marshaled error contains a message already
		} else {
			errorData, _ = sjson.SetBytes(nil, "message", err.Error())
		}
		if len(prefixMessage) > 0 {
			errorData, _ = sjson.SetBytes(errorData, "prefix_message", prefixMessage)
		}
		data = errorData
	}
	return &WebsocketRequest{
		ReqID:   wsc.ReqID,
		Command: cmd,
		Data:    data,
	}
}

type WebsocketTransaction struct {
	Status string `json:"status"`
	TxnID  string `json:"txn_id"`
	Transaction
}

type WebsocketTransactionResponse struct {
	TxnID string `json:"txn_id"`
}

type WebsocketMessage struct {
	WebsocketTransaction
	WebsocketCommand
}

const (
	WebsocketCloseConnReplaced       = 4001
	WebsocketCloseTxnNotAcknowledged = 4002
)

type MeowWebsocketCloseCode string

const (
	MeowServerShuttingDown MeowWebsocketCloseCode = "server_shutting_down"
	MeowConnectionReplaced MeowWebsocketCloseCode = "conn_replaced"
	MeowTxnNotAcknowledged MeowWebsocketCloseCode = "transactions_not_acknowledged"
)

var (
	ErrWebsocketManualStop   = errors.New("the websocket was disconnected manually")
	ErrWebsocketOverridden   = errors.New("a new call to StartWebsocket overrode the previous connection")
	ErrWebsocketUnknownError = errors.New("an unknown error occurred")

	ErrWebsocketNotConnected = errors.New("websocket not connected")
	ErrWebsocketClosed       = errors.New("websocket closed before response received")
)

func (mwcc MeowWebsocketCloseCode) String() string {
	switch mwcc {
	case MeowServerShuttingDown:
		return "the server is shutting down"
	case MeowConnectionReplaced:
		return "the connection was replaced by another client"
	case MeowTxnNotAcknowledged:
		return "transactions were not acknowledged"
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
		if closeCommand.Code == WebsocketCloseConnReplaced {
			closeCommand.Status = MeowConnectionReplaced
		} else if closeCommand.Code == websocket.CloseServiceRestart {
			closeCommand.Status = MeowServerShuttingDown
		}
	}
	return &closeCommand
}

func (as *AppService) HasWebsocket() bool {
	return as.ws != nil
}

func (as *AppService) SendWebsocket(cmd *WebsocketRequest) error {
	ws := as.ws
	if cmd == nil {
		return nil
	} else if ws == nil {
		return ErrWebsocketNotConnected
	}
	as.wsWriteLock.Lock()
	defer as.wsWriteLock.Unlock()
	if cmd.Deadline == 0 {
		cmd.Deadline = 3 * time.Minute
	}
	_ = ws.SetWriteDeadline(time.Now().Add(cmd.Deadline))
	return ws.WriteJSON(cmd)
}

func (as *AppService) clearWebsocketResponseWaiters() {
	as.websocketRequestsLock.Lock()
	for _, waiter := range as.websocketRequests {
		waiter <- &WebsocketCommand{Command: "__websocket_closed"}
	}
	as.websocketRequests = make(map[int]chan<- *WebsocketCommand)
	as.websocketRequestsLock.Unlock()
}

func (as *AppService) addWebsocketResponseWaiter(reqID int, waiter chan<- *WebsocketCommand) {
	as.websocketRequestsLock.Lock()
	as.websocketRequests[reqID] = waiter
	as.websocketRequestsLock.Unlock()
}

func (as *AppService) removeWebsocketResponseWaiter(reqID int, waiter chan<- *WebsocketCommand) {
	as.websocketRequestsLock.Lock()
	existingWaiter, ok := as.websocketRequests[reqID]
	if ok && existingWaiter == waiter {
		delete(as.websocketRequests, reqID)
	}
	close(waiter)
	as.websocketRequestsLock.Unlock()
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (er *ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", er.Code, er.Message)
}

func (as *AppService) RequestWebsocket(ctx context.Context, cmd *WebsocketRequest, response interface{}) error {
	cmd.ReqID = int(atomic.AddInt32(&as.websocketRequestID, 1))
	respChan := make(chan *WebsocketCommand, 1)
	as.addWebsocketResponseWaiter(cmd.ReqID, respChan)
	defer as.removeWebsocketResponseWaiter(cmd.ReqID, respChan)
	err := as.SendWebsocket(cmd)
	if err != nil {
		return err
	}
	select {
	case resp := <-respChan:
		if resp.Command == "__websocket_closed" {
			return ErrWebsocketClosed
		} else if resp.Command == "error" {
			var respErr ErrorResponse
			err = json.Unmarshal(resp.Data, &respErr)
			if err != nil {
				return fmt.Errorf("failed to parse error JSON: %w", err)
			}
			return &respErr
		} else if response != nil {
			err = json.Unmarshal(resp.Data, &response)
			if err != nil {
				return fmt.Errorf("failed to parse response JSON: %w", err)
			}
			return nil
		} else {
			return nil
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (as *AppService) unknownCommandHandler(cmd WebsocketCommand) (bool, interface{}) {
	zerolog.Ctx(cmd.Ctx).Warn().Msg("No handler for websocket command")
	return false, fmt.Errorf("unknown request type")
}

func (as *AppService) SetWebsocketCommandHandler(cmd string, handler WebsocketHandler) {
	as.websocketHandlersLock.Lock()
	as.websocketHandlers[cmd] = handler
	as.websocketHandlersLock.Unlock()
}

func (as *AppService) consumeWebsocket(stopFunc func(error), ws *websocket.Conn) {
	defer stopFunc(ErrWebsocketUnknownError)
	ctx := context.Background()
	for {
		var msg WebsocketMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			as.Log.Debug().Err(err).Msg("Error reading from websocket")
			stopFunc(parseCloseError(err))
			return
		}
		with := as.Log.With().
			Int("req_id", msg.ReqID).
			Str("ws_command", msg.Command)
		if msg.TxnID != "" {
			with = with.Str("transaction_id", msg.TxnID)
		}
		log := with.Logger()
		ctx = log.WithContext(ctx)
		if msg.Command == "" || msg.Command == "transaction" {
			if msg.TxnID == "" || !as.txnIDC.IsProcessed(msg.TxnID) {
				as.handleTransaction(ctx, msg.TxnID, &msg.Transaction)
			} else {
				log.Debug().
					Object("content", &msg.Transaction).
					Msg("Ignoring duplicate transaction")
			}
			go func() {
				err = as.SendWebsocket(msg.MakeResponse(true, &WebsocketTransactionResponse{TxnID: msg.TxnID}))
				if err != nil {
					log.Warn().Err(err).Msg("Failed to send response to websocket transaction")
				} else {
					log.Debug().Msg("Sent response to transaction")
				}
			}()
		} else if msg.Command == "connect" {
			log.Debug().Msg("Websocket connect confirmation received")
		} else if msg.Command == "response" || msg.Command == "error" {
			as.websocketRequestsLock.RLock()
			respChan, ok := as.websocketRequests[msg.ReqID]
			if ok {
				select {
				case respChan <- &msg.WebsocketCommand:
				default:
					log.Warn().Msg("Failed to handle response: channel didn't accept response")
				}
			} else {
				log.Warn().Msg("Dropping response to unknown request ID")
			}
			as.websocketRequestsLock.RUnlock()
		} else {
			log.Debug().Msg("Received websocket command")
			as.websocketHandlersLock.RLock()
			handler, ok := as.websocketHandlers[msg.Command]
			as.websocketHandlersLock.RUnlock()
			if !ok {
				handler = as.unknownCommandHandler
			}
			go func() {
				okResp, data := handler(msg.WebsocketCommand)
				err = as.SendWebsocket(msg.MakeResponse(okResp, data))
				if err != nil {
					log.Error().Err(err).Msg("Failed to send response to websocket command")
				} else if okResp {
					log.Debug().Msg("Sent success response to websocket command")
				} else {
					log.Debug().Msg("Sent error response to websocket command")
				}
			}()
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

		"X-Mautrix-Process-ID":        []string{as.ProcessID},
		"X-Mautrix-Websocket-Version": []string{"3"},
	})
	if resp != nil && resp.StatusCode >= 400 {
		var errResp Error
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		if err != nil {
			return fmt.Errorf("websocket request returned HTTP %d with non-JSON body", resp.StatusCode)
		} else {
			return fmt.Errorf("websocket request returned %s (HTTP %d): %s", errResp.ErrorCode, resp.StatusCode, errResp.Message)
		}
	} else if err != nil {
		return fmt.Errorf("failed to open websocket: %w", err)
	}
	if as.StopWebsocket != nil {
		as.StopWebsocket(ErrWebsocketOverridden)
	}
	closeChan := make(chan error)
	closeChanOnce := sync.Once{}
	stopFunc := func(err error) {
		closeChanOnce.Do(func() {
			closeChan <- err
		})
	}
	as.ws = ws
	as.StopWebsocket = stopFunc
	as.PrepareWebsocket()
	as.Log.Debug().Msg("Appservice transaction websocket opened")

	go as.consumeWebsocket(stopFunc, ws)

	if onConnect != nil {
		onConnect()
	}

	closeErr := <-closeChan

	if as.ws == ws {
		as.clearWebsocketResponseWaiters()
		as.ws = nil
	}

	_ = ws.SetWriteDeadline(time.Now().Add(3 * time.Second))
	err = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, ""))
	if err != nil && !errors.Is(err, websocket.ErrCloseSent) {
		as.Log.Warn().Err(err).Msg("Error writing close message to websocket")
	}
	err = ws.Close()
	if err != nil {
		as.Log.Warn().Err(err).Msg("Error closing websocket")
	}
	return closeErr
}
