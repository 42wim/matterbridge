package torrent

import (
	"fmt"
	"time"

	"github.com/anacrolix/log"
	"github.com/anacrolix/upnp"
)

const UpnpDiscoverLogTag = "upnp-discover"

func (cl *Client) addPortMapping(d upnp.Device, proto upnp.Protocol, internalPort int, upnpID string) {
	logger := cl.logger.WithContextText(fmt.Sprintf("UPnP device at %v: mapping internal %v port %v", d.GetLocalIPAddress(), proto, internalPort))
	externalPort, err := d.AddPortMapping(proto, internalPort, internalPort, upnpID, 0)
	if err != nil {
		logger.WithDefaultLevel(log.Warning).Printf("error: %v", err)
		return
	}
	level := log.Info
	if externalPort != internalPort {
		level = log.Warning
	}
	logger.WithDefaultLevel(level).Printf("success: external port %v", externalPort)
}

func (cl *Client) forwardPort() {
	cl.lock()
	defer cl.unlock()
	if cl.config.NoDefaultPortForwarding {
		return
	}
	cl.unlock()
	ds := upnp.Discover(0, 2*time.Second, cl.logger.WithValues(UpnpDiscoverLogTag))
	cl.lock()
	cl.logger.WithDefaultLevel(log.Debug).Printf("discovered %d upnp devices", len(ds))
	port := cl.incomingPeerPort()
	id := cl.config.UpnpID
	cl.unlock()
	for _, d := range ds {
		go cl.addPortMapping(d, upnp.TCP, port, id)
		go cl.addPortMapping(d, upnp.UDP, port, id)
	}
	cl.lock()
}
