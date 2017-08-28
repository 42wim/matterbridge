package gitter

import (
	"encoding/json"
	"fmt"

	"github.com/mrexodia/wray"
)

type Faye struct {
	endpoint string
	Event    chan Event
	client   *wray.FayeClient
	gitter   *Gitter
}

func (gitter *Gitter) Faye(roomID string) *Faye {
	wray.RegisterTransports([]wray.Transport{
		&wray.HttpTransport{
			SendHook: func(data map[string]interface{}) {
				if channel, ok := data["channel"]; ok && channel == "/meta/handshake" {
					data["ext"] = map[string]interface{}{"token": gitter.config.token}
				}
			},
		},
	})
	return &Faye{
		endpoint: "/api/v1/rooms/" + roomID + "/chatMessages",
		Event:    make(chan Event),
		client:   wray.NewFayeClient(fayeBaseURL),
		gitter:   gitter,
	}
}

func (faye *Faye) Listen() {
	defer faye.destroy()

	faye.client.Subscribe(faye.endpoint, false, func(message wray.Message) {
		dataBytes, err := json.Marshal(message.Data["model"])
		if err != nil {
			fmt.Printf("JSON Marshal error: %v\n", err)
			return
		}
		var gitterMessage Message
		err = json.Unmarshal(dataBytes, &gitterMessage)
		if err != nil {
			fmt.Printf("JSON Unmarshal error: %v\n", err)
			return
		}
		faye.Event <- Event{
			Data: &MessageReceived{
				Message: gitterMessage,
			},
		}
	})

	//TODO: this might be needed in the future
	/*go func() {
		for {
			faye.client.Publish("/api/v1/ping2", map[string]interface{}{"reason": "ping"})
			time.Sleep(60 * time.Second)
		}
	}()*/

	faye.client.Listen()
}

func (faye *Faye) destroy() {
	close(faye.Event)
}
