package utp

import "github.com/anacrolix/log"

const (
	logCallbacks = false
	utpLogging   = false
)

// The default Socket Logger. Override per Socket by using WithLogger with NewSocket.
var Logger = log.Default.WithContextText("go-libutp")
