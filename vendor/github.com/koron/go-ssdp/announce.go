package ssdp

import (
	"bytes"
	"fmt"
	"net"

	"github.com/koron/go-ssdp/internal/multicast"
)

// AnnounceAlive sends ssdp:alive message.
// location should be a string or a ssdp.LocationProvider.
func AnnounceAlive(nt, usn string, location interface{}, server string, maxAge int, localAddr string) error {
	locProv, err := toLocationProvider(location)
	if err != nil {
		return err
	}
	// dial multicast UDP packet.
	conn, err := multicast.Listen(&multicast.AddrResolver{Addr: localAddr})
	if err != nil {
		return err
	}
	defer conn.Close()
	// build and send message.
	addr, err := multicast.SendAddr()
	if err != nil {
		return err
	}
	msg := &aliveDataProvider{
		host:     addr,
		nt:       nt,
		usn:      usn,
		location: locProv,
		server:   server,
		maxAge:   maxAge,
	}
	if _, err := conn.WriteTo(msg, addr); err != nil {
		return err
	}
	return nil
}

type aliveDataProvider struct {
	host     net.Addr
	nt       string
	usn      string
	location LocationProvider
	server   string
	maxAge   int
}

func (p *aliveDataProvider) Bytes(ifi *net.Interface) []byte {
	return buildAlive(p.host, p.nt, p.usn, p.location.Location(nil, ifi), p.server, p.maxAge)
}

var _ multicast.DataProvider = (*aliveDataProvider)(nil)

func buildAlive(raddr net.Addr, nt, usn, location, server string, maxAge int) []byte {
	// bytes.Buffer#Write() is never fail, so we can omit error checks.
	b := new(bytes.Buffer)
	b.WriteString("NOTIFY * HTTP/1.1\r\n")
	fmt.Fprintf(b, "HOST: %s\r\n", raddr.String())
	fmt.Fprintf(b, "NT: %s\r\n", nt)
	fmt.Fprintf(b, "NTS: %s\r\n", "ssdp:alive")
	fmt.Fprintf(b, "USN: %s\r\n", usn)
	if location != "" {
		fmt.Fprintf(b, "LOCATION: %s\r\n", location)
	}
	if server != "" {
		fmt.Fprintf(b, "SERVER: %s\r\n", server)
	}
	fmt.Fprintf(b, "CACHE-CONTROL: max-age=%d\r\n", maxAge)
	b.WriteString("\r\n")
	return b.Bytes()
}

// AnnounceBye sends ssdp:byebye message.
func AnnounceBye(nt, usn, localAddr string) error {
	// dial multicast UDP packet.
	conn, err := multicast.Listen(&multicast.AddrResolver{Addr: localAddr})
	if err != nil {
		return err
	}
	defer conn.Close()
	// build and send message.
	addr, err := multicast.SendAddr()
	if err != nil {
		return err
	}
	msg, err := buildBye(addr, nt, usn)
	if err != nil {
		return err
	}
	if _, err := conn.WriteTo(multicast.BytesDataProvider(msg), addr); err != nil {
		return err
	}
	return nil
}

func buildBye(raddr net.Addr, nt, usn string) ([]byte, error) {
	b := new(bytes.Buffer)
	// FIXME: error should be checked.
	b.WriteString("NOTIFY * HTTP/1.1\r\n")
	fmt.Fprintf(b, "HOST: %s\r\n", raddr.String())
	fmt.Fprintf(b, "NT: %s\r\n", nt)
	fmt.Fprintf(b, "NTS: %s\r\n", "ssdp:byebye")
	fmt.Fprintf(b, "USN: %s\r\n", usn)
	b.WriteString("\r\n")
	return b.Bytes(), nil
}
