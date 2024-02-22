package vnet

import (
	"fmt"
	"net"
)

// Deliver directly send packet to vnet or real-server.
// For example, we can use this API to simulate the REPLAY ATTACK.
func (v *UDPProxy) Deliver(sourceAddr, destAddr net.Addr, b []byte) (nn int, err error) {
	v.workers.Range(func(key, value interface{}) bool {
		if nn, err = value.(*aUDPProxyWorker).Deliver(sourceAddr, destAddr, b); err != nil {
			return false // Fail, abort.
		} else if nn == len(b) {
			return false // Done.
		}

		return true // Deliver by next worker.
	})
	return
}

func (v *aUDPProxyWorker) Deliver(sourceAddr, destAddr net.Addr, b []byte) (nn int, err error) {
	addr, ok := sourceAddr.(*net.UDPAddr)
	if !ok {
		return 0, fmt.Errorf("invalid addr %v", sourceAddr) // nolint:goerr113
	}

	// nolint:godox // TODO: Support deliver packet from real server to vnet.
	// If packet is from vnet, proxy to real server.
	var realSocket *net.UDPConn
	if value, ok := v.endpoints.Load(addr.String()); !ok {
		return 0, nil
	} else { // nolint:golint
		realSocket = value.(*net.UDPConn)
	}

	// Send to real server.
	if _, err := realSocket.Write(b); err != nil {
		return 0, err
	}

	return len(b), nil
}
