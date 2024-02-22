// Package nack provides interceptors to implement sending and receiving negative acknowledgements
package nack

import "github.com/pion/interceptor"

func streamSupportNack(info *interceptor.StreamInfo) bool {
	for _, fb := range info.RTCPFeedback {
		if fb.Type == "nack" && fb.Parameter == "" {
			return true
		}
	}

	return false
}
