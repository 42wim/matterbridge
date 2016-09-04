package wray

import (
	"errors"
)

type Transport interface {
	isUsable(string) bool
	connectionType() string
	send(map[string]interface{}) (Response, error)
	setUrl(string)
}

func SelectTransport(client *FayeClient, transportTypes []string, disabled []string) (Transport, error) {
	for _, transport := range registeredTransports {
		if contains(transport.connectionType(), transportTypes) && transport.isUsable(client.url) {
			return transport, nil
		}
	}
	return nil, errors.New("No usable transports available")
}
