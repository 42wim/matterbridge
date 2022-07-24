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
	"io"
	"net"
	"net/http"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// Start starts the HTTP server that listens for calls from the Matrix homeserver.
func (as *AppService) Start() {
	as.server = &http.Server{
		Handler: as.Router,
	}
	var err error
	if as.Host.IsUnixSocket() {
		err = as.listenUnix()
	} else {
		as.server.Addr = as.Host.Address()
		err = as.listenTCP()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		as.Log.Error().Err(err).Msg("Error in HTTP listener")
	} else {
		as.Log.Debug().Msg("HTTP listener stopped")
	}
}

func (as *AppService) listenUnix() error {
	socket := as.Host.Hostname
	_ = syscall.Unlink(socket)
	defer func() {
		_ = syscall.Unlink(socket)
	}()
	listener, err := net.Listen("unix", socket)
	if err != nil {
		return err
	}
	as.Log.Info().Str("socket", socket).Msg("Starting unix socket HTTP listener")
	return as.server.Serve(listener)
}

func (as *AppService) listenTCP() error {
	if len(as.Host.TLSCert) == 0 || len(as.Host.TLSKey) == 0 {
		as.Log.Info().Str("address", as.server.Addr).Msg("Starting HTTP listener")
		return as.server.ListenAndServe()
	} else {
		as.Log.Info().Str("address", as.server.Addr).Msg("Starting HTTP listener with TLS")
		return as.server.ListenAndServeTLS(as.Host.TLSCert, as.Host.TLSKey)
	}
}

func (as *AppService) Stop() {
	if as.server == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = as.server.Shutdown(ctx)
	as.server = nil
}

// CheckServerToken checks if the given request originated from the Matrix homeserver.
func (as *AppService) CheckServerToken(w http.ResponseWriter, r *http.Request) (isValid bool) {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 0 && strings.HasPrefix(authHeader, "Bearer ") {
		isValid = authHeader[len("Bearer "):] == as.Registration.ServerToken
	} else {
		queryToken := r.URL.Query().Get("access_token")
		if len(queryToken) > 0 {
			isValid = queryToken == as.Registration.ServerToken
		} else {
			Error{
				ErrorCode:  ErrUnknownToken,
				HTTPStatus: http.StatusForbidden,
				Message:    "Missing access token",
			}.Write(w)
			return
		}
	}
	if !isValid {
		Error{
			ErrorCode:  ErrUnknownToken,
			HTTPStatus: http.StatusForbidden,
			Message:    "Incorrect access token",
		}.Write(w)
	}
	return
}

// PutTransaction handles a /transactions PUT call from the homeserver.
func (as *AppService) PutTransaction(w http.ResponseWriter, r *http.Request) {
	if !as.CheckServerToken(w, r) {
		return
	}

	vars := mux.Vars(r)
	txnID := vars["txnID"]
	if len(txnID) == 0 {
		Error{
			ErrorCode:  ErrNoTransactionID,
			HTTPStatus: http.StatusBadRequest,
			Message:    "Missing transaction ID",
		}.Write(w)
		return
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		Error{
			ErrorCode:  ErrNotJSON,
			HTTPStatus: http.StatusBadRequest,
			Message:    "Missing request body",
		}.Write(w)
		return
	}
	log := as.Log.With().Str("transaction_id", txnID).Logger()
	ctx := context.Background()
	ctx = log.WithContext(ctx)
	if as.txnIDC.IsProcessed(txnID) {
		// Duplicate transaction ID: no-op
		WriteBlankOK(w)
		log.Debug().Msg("Ignoring duplicate transaction")
		return
	}

	var txn Transaction
	err = json.Unmarshal(body, &txn)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse transaction content")
		Error{
			ErrorCode:  ErrBadJSON,
			HTTPStatus: http.StatusBadRequest,
			Message:    "Failed to parse body JSON",
		}.Write(w)
	} else {
		as.handleTransaction(ctx, txnID, &txn)
		WriteBlankOK(w)
	}
}

func (as *AppService) handleTransaction(ctx context.Context, id string, txn *Transaction) {
	log := zerolog.Ctx(ctx)
	log.Debug().Object("content", txn).Msg("Starting handling of transaction")
	if as.Registration.EphemeralEvents {
		if txn.EphemeralEvents != nil {
			as.handleEvents(ctx, txn.EphemeralEvents, event.EphemeralEventType)
		} else if txn.MSC2409EphemeralEvents != nil {
			as.handleEvents(ctx, txn.MSC2409EphemeralEvents, event.EphemeralEventType)
		}
		if txn.ToDeviceEvents != nil {
			as.handleEvents(ctx, txn.ToDeviceEvents, event.ToDeviceEventType)
		} else if txn.MSC2409ToDeviceEvents != nil {
			as.handleEvents(ctx, txn.MSC2409ToDeviceEvents, event.ToDeviceEventType)
		}
	}
	as.handleEvents(ctx, txn.Events, event.UnknownEventType)
	if txn.DeviceLists != nil {
		as.handleDeviceLists(ctx, txn.DeviceLists)
	} else if txn.MSC3202DeviceLists != nil {
		as.handleDeviceLists(ctx, txn.MSC3202DeviceLists)
	}
	if txn.DeviceOTKCount != nil {
		as.handleOTKCounts(ctx, txn.DeviceOTKCount)
	} else if txn.MSC3202DeviceOTKCount != nil {
		as.handleOTKCounts(ctx, txn.MSC3202DeviceOTKCount)
	}
	as.txnIDC.MarkProcessed(id)
	log.Debug().Msg("Finished dispatching events from transaction")
}

func (as *AppService) handleOTKCounts(ctx context.Context, otks OTKCountMap) {
	for userID, devices := range otks {
		for deviceID, otkCounts := range devices {
			otkCounts.UserID = userID
			otkCounts.DeviceID = deviceID
			select {
			case as.OTKCounts <- &otkCounts:
			default:
				zerolog.Ctx(ctx).Warn().
					Str("user_id", userID.String()).
					Msg("Dropped OTK count update for user because channel is full")
			}
		}
	}
}

func (as *AppService) handleDeviceLists(ctx context.Context, dl *mautrix.DeviceLists) {
	select {
	case as.DeviceLists <- dl:
	default:
		zerolog.Ctx(ctx).Warn().Msg("Dropped device list update because channel is full")
	}
}

func (as *AppService) handleEvents(ctx context.Context, evts []*event.Event, defaultTypeClass event.TypeClass) {
	log := zerolog.Ctx(ctx)
	for _, evt := range evts {
		evt.Mautrix.ReceivedAt = time.Now()
		if defaultTypeClass != event.UnknownEventType {
			evt.Type.Class = defaultTypeClass
		} else if evt.StateKey != nil {
			evt.Type.Class = event.StateEventType
		} else {
			evt.Type.Class = event.MessageEventType
		}
		err := evt.Content.ParseRaw(evt.Type)
		if errors.Is(err, event.ErrUnsupportedContentType) {
			log.Debug().Str("event_id", evt.ID.String()).Msg("Not parsing content of unsupported event")
		} else if err != nil {
			log.Warn().Err(err).
				Str("event_id", evt.ID.String()).
				Str("event_type", evt.Type.Type).
				Str("event_type_class", evt.Type.Class.Name()).
				Msg("Failed to parse content of event")
		}

		if evt.Type.IsState() {
			// TODO remove this check after making sure the log doesn't happen
			historical, ok := evt.Content.Raw["org.matrix.msc2716.historical"].(bool)
			if ok && historical {
				log.Warn().
					Str("event_id", evt.ID.String()).
					Str("event_type", evt.Type.Type).
					Str("state_key", evt.GetStateKey()).
					Msg("Received historical state event")
			} else {
				mautrix.UpdateStateStore(as.StateStore, evt)
			}
		}
		var ch chan *event.Event
		if evt.Type.Class == event.ToDeviceEventType {
			ch = as.ToDeviceEvents
		} else {
			ch = as.Events
		}
		select {
		case ch <- evt:
		default:
			log.Warn().
				Str("event_id", evt.ID.String()).
				Str("event_type", evt.Type.Type).
				Str("event_type_class", evt.Type.Class.Name()).
				Msg("Event channel is full")
			ch <- evt
		}
	}
}

// GetRoom handles a /rooms GET call from the homeserver.
func (as *AppService) GetRoom(w http.ResponseWriter, r *http.Request) {
	if !as.CheckServerToken(w, r) {
		return
	}

	vars := mux.Vars(r)
	roomAlias := vars["roomAlias"]
	ok := as.QueryHandler.QueryAlias(roomAlias)
	if ok {
		WriteBlankOK(w)
	} else {
		Error{
			ErrorCode:  ErrUnknown,
			HTTPStatus: http.StatusNotFound,
		}.Write(w)
	}
}

// GetUser handles a /users GET call from the homeserver.
func (as *AppService) GetUser(w http.ResponseWriter, r *http.Request) {
	if !as.CheckServerToken(w, r) {
		return
	}

	vars := mux.Vars(r)
	userID := id.UserID(vars["userID"])
	ok := as.QueryHandler.QueryUser(userID)
	if ok {
		WriteBlankOK(w)
	} else {
		Error{
			ErrorCode:  ErrUnknown,
			HTTPStatus: http.StatusNotFound,
		}.Write(w)
	}
}

func (as *AppService) PostPing(w http.ResponseWriter, r *http.Request) {
	if !as.CheckServerToken(w, r) {
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 || !json.Valid(body) {
		Error{
			ErrorCode:  ErrNotJSON,
			HTTPStatus: http.StatusBadRequest,
			Message:    "Missing request body",
		}.Write(w)
		return
	}

	var txn mautrix.ReqAppservicePing
	_ = json.Unmarshal(body, &txn)
	as.Log.Debug().Str("txn_id", txn.TxnID).Msg("Received ping from homeserver")

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func (as *AppService) GetLive(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if as.Live {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write([]byte("{}"))
}

func (as *AppService) GetReady(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if as.Ready {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write([]byte("{}"))
}
