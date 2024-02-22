// Copyright (C) 2015 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package upnp

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type MappingChangeSubscriber func(*Mapping, []Address, []Address)

type Mapping struct {
	protocol Protocol
	address  Address

	extAddresses map[string]Address // NAT ID -> Address
	expires      time.Time
	subscribers  []MappingChangeSubscriber
	mut          sync.RWMutex

	ll levelLogger
}

func (m *Mapping) setAddress(id string, address Address) {
	m.mut.Lock()
	if existing, ok := m.extAddresses[id]; !ok || !existing.Equal(address) {
		m.ll.Infof("New NAT port mapping: external %s address %s to local address %s.", m.protocol, address, m.address)
		m.extAddresses[id] = address
	}
	m.mut.Unlock()
}

func (m *Mapping) removeAddress(id string) {
	m.mut.Lock()
	addr, ok := m.extAddresses[id]
	if ok {
		m.ll.Infof("Removing NAT port mapping: external %s address %s, NAT %s is no longer available.", m.protocol, addr, id)
		delete(m.extAddresses, id)
	}
	m.mut.Unlock()
}

func (m *Mapping) clearAddresses() {
	m.mut.Lock()
	var removed []Address
	for id, addr := range m.extAddresses {
		m.ll.Infof("Clearing mapping %s: ID: %s Address: %s", m, id, addr)
		removed = append(removed, addr)
		delete(m.extAddresses, id)
	}
	m.expires = time.Time{}
	m.mut.Unlock()
	if len(removed) > 0 {
		m.notify(nil, removed)
	}
}

func (m *Mapping) notify(added, removed []Address) {
	m.mut.RLock()
	for _, subscriber := range m.subscribers {
		subscriber(m, added, removed)
	}
	m.mut.RUnlock()
}

func (m *Mapping) addressMap() map[string]Address {
	m.mut.RLock()
	addrMap := m.extAddresses
	m.mut.RUnlock()
	return addrMap
}

func (m *Mapping) Protocol() Protocol {
	return m.protocol
}

func (m *Mapping) Address() Address {
	return m.address
}

func (m *Mapping) ExternalAddresses() []Address {
	m.mut.RLock()
	addrs := make([]Address, 0, len(m.extAddresses))
	for _, addr := range m.extAddresses {
		addrs = append(addrs, addr)
	}
	m.mut.RUnlock()
	return addrs
}

func (m *Mapping) OnChanged(subscribed MappingChangeSubscriber) {
	m.mut.Lock()
	m.subscribers = append(m.subscribers, subscribed)
	m.mut.Unlock()
}

func (m *Mapping) String() string {
	return fmt.Sprintf("%s %s", m.protocol, m.address)
}

func (m *Mapping) GoString() string {
	return m.String()
}

// Checks if the mappings local IP address matches the IP address of the gateway
// For example, if we are explicitly listening on 192.168.0.12, there is no
// point trying to acquire a mapping on a gateway to which the local IP is
// 10.0.0.1. Fallback to true if any of the IPs is not there.
func (m *Mapping) validGateway(ip net.IP) bool {
	if m.address.IP == nil || ip == nil || m.address.IP.IsUnspecified() || ip.IsUnspecified() {
		return true
	}
	return m.address.IP.Equal(ip)
}

// Address is essentially net.TCPAddr yet is more general, and has a few helper
// methods which reduce boilerplate code.
type Address struct {
	IP   net.IP
	Port int
}

func (a Address) Equal(b Address) bool {
	return a.Port == b.Port && a.IP.Equal(b.IP)
}

func (a Address) String() string {
	var ipStr string
	if a.IP == nil {
		ipStr = net.IPv4zero.String()
	} else {
		ipStr = a.IP.String()
	}
	return net.JoinHostPort(ipStr, fmt.Sprintf("%d", a.Port))
}

func (a Address) GoString() string {
	return a.String()
}
