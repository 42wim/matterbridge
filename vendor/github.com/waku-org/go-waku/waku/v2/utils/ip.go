package utils

import (
	"net"
	"strings"
)

// IsIPv4 validates if string is a valid IPV4 address
func IsIPv4(str string) bool {
	ip := net.ParseIP(str).To4()
	return ip != nil
}

// IsIPv6 validates if string is a valid IPV6 address
func IsIPv6(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ":")
}
