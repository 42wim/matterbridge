package torrent

import "strings"

func LoopbackListenHost(network string) string {
	if strings.IndexByte(network, '4') != -1 {
		return "127.0.0.1"
	} else {
		return "::1"
	}
}
