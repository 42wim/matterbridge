package utp

/*
#include "utp.h"
*/
import "C"

import (
	"net"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/anacrolix/missinggo/inproc"
)

func toSockaddrInet(ip net.IP, port int, zone string) (rsa syscall.RawSockaddrAny, len C.socklen_t) {
	if ip4 := ip.To4(); ip4 != nil && zone == "" {
		rsa4 := (*syscall.RawSockaddrInet4)(unsafe.Pointer(&rsa))
		rsa4.Family = syscall.AF_INET
		rsa4.Port = uint16(port)
		if n := copy(rsa4.Addr[:], ip4); n != 4 {
			panic(n)
		}
		len = C.socklen_t(unsafe.Sizeof(*rsa4))
		return
	}
	rsa6 := (*syscall.RawSockaddrInet6)(unsafe.Pointer(&rsa))
	rsa6.Family = syscall.AF_INET6
	rsa6.Scope_id = zoneToScopeId(zone)
	rsa6.Port = uint16(port)
	if ip != nil {
		if n := copy(rsa6.Addr[:], ip); n != 16 {
			panic(n)
		}
	}
	len = C.socklen_t(unsafe.Sizeof(*rsa6))
	return
}

func zoneToScopeId(zone string) uint32 {
	if zone == "" {
		return 0
	}
	if ifi, err := net.InterfaceByName(zone); err == nil {
		return uint32(ifi.Index)
	}
	ui64, _ := strconv.ParseUint(zone, 10, 32)
	return uint32(ui64)
}

func structSockaddrToUDPAddr(sa *C.struct_sockaddr, udp *net.UDPAddr) error {
	return anySockaddrToUdp((*syscall.RawSockaddrAny)(unsafe.Pointer(sa)), udp)
}

func anySockaddrToUdp(rsa *syscall.RawSockaddrAny, udp *net.UDPAddr) error {
	switch rsa.Addr.Family {
	case syscall.AF_INET:
		sa := (*syscall.RawSockaddrInet4)(unsafe.Pointer(rsa))
		udp.Port = int(sa.Port)
		udp.IP = append(udp.IP[:0], sa.Addr[:]...)
		return nil
	case syscall.AF_INET6:
		sa := (*syscall.RawSockaddrInet6)(unsafe.Pointer(rsa))
		udp.Port = int(sa.Port)
		udp.IP = append(udp.IP[:0], sa.Addr[:]...)
		return nil
	default:
		return syscall.EAFNOSUPPORT
	}
}

func sockaddrToUDP(sa syscall.Sockaddr) net.Addr {
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		return &net.UDPAddr{IP: sa.Addr[0:], Port: sa.Port}
	case *syscall.SockaddrInet6:
		return &net.UDPAddr{IP: sa.Addr[0:], Port: sa.Port /*Zone: zoneToString(int(sa.ZoneId))*/}
	}
	return nil
}

func netAddrToLibSockaddr(na net.Addr) (rsa syscall.RawSockaddrAny, len C.socklen_t) {
	switch v := na.(type) {
	case *net.UDPAddr:
		return toSockaddrInet(v.IP, v.Port, v.Zone)
	case inproc.Addr:
		rsa6 := (*syscall.RawSockaddrInet6)(unsafe.Pointer(&rsa))
		rsa6.Port = uint16(v.Port)
		len = C.socklen_t(unsafe.Sizeof(rsa))
		return
	default:
		panic(na)
	}
}
