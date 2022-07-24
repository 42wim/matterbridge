// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package appservice

import (
	"encoding/json"
	"runtime/debug"

	"github.com/rs/zerolog"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

type ExecMode uint8

const (
	AsyncHandlers ExecMode = iota
	AsyncLoop
	Sync
)

type EventHandler = func(evt *event.Event)
type OTKHandler = func(otk *mautrix.OTKCount)
type DeviceListHandler = func(lists *mautrix.DeviceLists, since string)

type EventProcessor struct {
	ExecMode ExecMode

	as       *AppService
	stop     chan struct{}
	handlers map[event.Type][]EventHandler

	otkHandlers        []OTKHandler
	deviceListHandlers []DeviceListHandler
}

func NewEventProcessor(as *AppService) *EventProcessor {
	return &EventProcessor{
		ExecMode: AsyncHandlers,
		as:       as,
		stop:     make(chan struct{}, 1),
		handlers: make(map[event.Type][]EventHandler),

		otkHandlers:        make([]OTKHandler, 0),
		deviceListHandlers: make([]DeviceListHandler, 0),
	}
}

func (ep *EventProcessor) On(evtType event.Type, handler EventHandler) {
	handlers, ok := ep.handlers[evtType]
	if !ok {
		handlers = []EventHandler{handler}
	} else {
		handlers = append(handlers, handler)
	}
	ep.handlers[evtType] = handlers
}

func (ep *EventProcessor) PrependHandler(evtType event.Type, handler EventHandler) {
	handlers, ok := ep.handlers[evtType]
	if !ok {
		handlers = []EventHandler{handler}
	} else {
		handlers = append([]EventHandler{handler}, handlers...)
	}
	ep.handlers[evtType] = handlers
}

func (ep *EventProcessor) OnOTK(handler OTKHandler) {
	ep.otkHandlers = append(ep.otkHandlers, handler)
}

func (ep *EventProcessor) OnDeviceList(handler DeviceListHandler) {
	ep.deviceListHandlers = append(ep.deviceListHandlers, handler)
}

func (ep *EventProcessor) recoverFunc(data interface{}) {
	if err := recover(); err != nil {
		d, _ := json.Marshal(data)
		ep.as.Log.Error().
			Str(zerolog.ErrorStackFieldName, string(debug.Stack())).
			Interface(zerolog.ErrorFieldName, err).
			Str("event_content", string(d)).
			Msg("Panic in Matrix event handler")
	}
}

func (ep *EventProcessor) callHandler(handler EventHandler, evt *event.Event) {
	defer ep.recoverFunc(evt)
	handler(evt)
}

func (ep *EventProcessor) callOTKHandler(handler OTKHandler, otk *mautrix.OTKCount) {
	defer ep.recoverFunc(otk)
	handler(otk)
}

func (ep *EventProcessor) callDeviceListHandler(handler DeviceListHandler, dl *mautrix.DeviceLists) {
	defer ep.recoverFunc(dl)
	handler(dl, "")
}

func (ep *EventProcessor) DispatchOTK(otk *mautrix.OTKCount) {
	for _, handler := range ep.otkHandlers {
		go ep.callOTKHandler(handler, otk)
	}
}

func (ep *EventProcessor) DispatchDeviceList(dl *mautrix.DeviceLists) {
	for _, handler := range ep.deviceListHandlers {
		go ep.callDeviceListHandler(handler, dl)
	}
}

func (ep *EventProcessor) Dispatch(evt *event.Event) {
	handlers, ok := ep.handlers[evt.Type]
	if !ok {
		return
	}
	switch ep.ExecMode {
	case AsyncHandlers:
		for _, handler := range handlers {
			go ep.callHandler(handler, evt)
		}
	case AsyncLoop:
		go func() {
			for _, handler := range handlers {
				ep.callHandler(handler, evt)
			}
		}()
	case Sync:
		for _, handler := range handlers {
			ep.callHandler(handler, evt)
		}
	}
}
func (ep *EventProcessor) startEvents() {
	for {
		select {
		case evt := <-ep.as.Events:
			ep.Dispatch(evt)
		case <-ep.stop:
			return
		}
	}
}

func (ep *EventProcessor) startEncryption() {
	for {
		select {
		case evt := <-ep.as.ToDeviceEvents:
			ep.Dispatch(evt)
		case otk := <-ep.as.OTKCounts:
			ep.DispatchOTK(otk)
		case dl := <-ep.as.DeviceLists:
			ep.DispatchDeviceList(dl)
		case <-ep.stop:
			return
		}
	}
}

func (ep *EventProcessor) Start() {
	go ep.startEvents()
	go ep.startEncryption()
}

func (ep *EventProcessor) Stop() {
	close(ep.stop)
}
