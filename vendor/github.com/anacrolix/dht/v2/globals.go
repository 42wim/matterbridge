package dht

import (
	"golang.org/x/time/rate"
)

var DefaultSendLimiter = rate.NewLimiter(25, 25)
