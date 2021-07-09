// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package appservice

import (
	"encoding/json"
	"runtime/debug"

	log "maunium.net/go/maulogger/v2"

	"maunium.net/go/mautrix/event"
)

type ExecMode uint8

const (
	AsyncHandlers ExecMode = iota
	AsyncLoop
	Sync
)

type EventHandler func(evt *event.Event)

type EventProcessor struct {
	ExecMode ExecMode

	as       *AppService
	log      log.Logger
	stop     chan struct{}
	handlers map[event.Type][]EventHandler
}

func NewEventProcessor(as *AppService) *EventProcessor {
	return &EventProcessor{
		ExecMode: AsyncHandlers,
		as:       as,
		log:      as.Log.Sub("Events"),
		stop:     make(chan struct{}, 1),
		handlers: make(map[event.Type][]EventHandler),
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

func (ep *EventProcessor) callHandler(handler EventHandler, evt *event.Event) {
	defer func() {
		if err := recover(); err != nil {
			d, _ := json.Marshal(evt)
			ep.log.Errorfln("Panic in Matrix event handler: %v (event content: %s):\n%s", err, string(d), string(debug.Stack()))
		}
	}()
	handler(evt)
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

func (ep *EventProcessor) Start() {
	for {
		select {
		case evt := <-ep.as.Events:
			ep.Dispatch(evt)
		case <-ep.stop:
			return
		}
	}
}

func (ep *EventProcessor) Stop() {
	ep.stop <- struct{}{}
}
