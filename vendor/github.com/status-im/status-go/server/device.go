package server

import (
	"os"
	"strings"
)

var (
	local = ".local"
)

func RemoveSuffix(input, suffix string) string {
	il := len(input)
	sl := len(suffix)
	if il > sl {
		if input[il-sl:] == suffix {
			return input[:il-sl]
		}
	}
	return input
}

func parseHostname(hostname string) string {
	hostname = RemoveSuffix(hostname, local)
	return strings.ReplaceAll(hostname, "-", " ")
}

func GetDeviceName() (string, error) {
	name, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return parseHostname(name), nil
}
