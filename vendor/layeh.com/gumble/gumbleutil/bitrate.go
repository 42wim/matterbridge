package gumbleutil

import (
	"time"

	"layeh.com/gumble/gumble"
)

var autoBitrate = &Listener{
	Connect: func(e *gumble.ConnectEvent) {
		if e.MaximumBitrate != nil {
			const safety = 5
			interval := e.Client.Config.AudioInterval
			dataBytes := (*e.MaximumBitrate / (8 * (int(time.Second/interval) + safety))) - 32 - 10

			e.Client.Config.AudioDataBytes = dataBytes
		}
	},
}

// AutoBitrate is a gumble.EventListener that automatically sets the client's
// AudioDataBytes to suitable value, based on the server's bitrate.
var AutoBitrate gumble.EventListener

func init() {
	AutoBitrate = autoBitrate
}
