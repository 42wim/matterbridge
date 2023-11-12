package connection

import (
	"fmt"
)

// State represents device connection state and type,
// as reported by mobile framework.
//
// Zero value represents default assumption about network (online and unknown type).
type State struct {
	Offline   bool `json:"offline"`
	Type      Type `json:"type"`
	Expensive bool `json:"expensive"`
}

// Type represents description of available
// connection types as reported by React Native (see
// https://facebook.github.io/react-native/docs/netinfo.html)
// We're interested mainly in 'wifi' and 'cellular', but
// other types are also may be used.
type Type byte

const (
	Offline  = "offline"
	Wifi     = "wifi"
	Cellular = "cellular"
	Unknown  = "unknown"
	None     = "none"
)

// NewType creates new Type from string.
func NewType(s string) Type {
	switch s {
	case Cellular:
		return connectionCellular
	case Wifi:
		return connectionWifi
	}

	return connectionUnknown
}

// Type constants
const (
	connectionUnknown  Type = iota
	connectionCellular      // cellular, LTE, 4G, 3G, EDGE, etc.
	connectionWifi          // WIFI or iOS simulator
)

func (c State) IsExpensive() bool {
	return c.Expensive || c.Type == connectionCellular
}

// string formats ConnectionState for logs. Implements Stringer.
func (c State) String() string {
	if c.Offline {
		return Offline
	}

	var typ string
	switch c.Type {
	case connectionWifi:
		typ = Wifi
	case connectionCellular:
		typ = Cellular
	default:
		typ = Unknown
	}

	if c.Expensive {
		return fmt.Sprintf("%s (expensive)", typ)
	}

	return typ
}
