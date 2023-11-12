package enr

import (
	"crypto/ecdsa"
	"encoding/binary"
	"errors"
	"math"
	"math/rand"
	"net"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/multiformats/go-multiaddr"
)

func NewLocalnode(priv *ecdsa.PrivateKey) (*enode.LocalNode, error) {
	db, err := enode.OpenDB("")
	if err != nil {
		return nil, err
	}
	return enode.NewLocalNode(db, priv), nil
}

type ENROption func(*enode.LocalNode) error

func WithMultiaddress(multiaddrs ...multiaddr.Multiaddr) ENROption {
	return func(localnode *enode.LocalNode) (err error) {

		// Randomly shuffle multiaddresses
		rand.Shuffle(len(multiaddrs), func(i, j int) { multiaddrs[i], multiaddrs[j] = multiaddrs[j], multiaddrs[i] })

		// Adding extra multiaddresses. Should probably not exceed the enr max size of 300bytes
		failedOnceWritingENR := false
		couldWriteENRatLeastOnce := false
		successIdx := -1
		for i := len(multiaddrs); i > 0; i-- {
			err = writeMultiaddressField(localnode, multiaddrs[0:i])
			if err == nil {
				couldWriteENRatLeastOnce = true
				successIdx = i
				break
			}
			failedOnceWritingENR = true
		}

		if failedOnceWritingENR && couldWriteENRatLeastOnce {
			// Could write a subset of multiaddresses but not all
			err = writeMultiaddressField(localnode, multiaddrs[0:successIdx])
			if err != nil {
				return errors.New("could not write new ENR")
			}
		}

		return nil
	}
}

func WithCapabilities(lightpush, filter, store, relay bool) ENROption {
	return func(localnode *enode.LocalNode) (err error) {
		wakuflags := NewWakuEnrBitfield(lightpush, filter, store, relay)
		return WithWakuBitfield(wakuflags)(localnode)
	}
}

func WithWakuBitfield(flags WakuEnrBitfield) ENROption {
	return func(localnode *enode.LocalNode) (err error) {
		localnode.Set(enr.WithEntry(WakuENRField, flags))
		return nil
	}
}

func WithIP(ipAddr *net.TCPAddr) ENROption {
	return func(localnode *enode.LocalNode) (err error) {
		localnode.SetStaticIP(ipAddr.IP)
		localnode.Set(enr.TCP(uint16(ipAddr.Port))) // TODO: ipv6?
		return nil
	}
}

func WithUDPPort(udpPort uint) ENROption {
	return func(localnode *enode.LocalNode) (err error) {
		if udpPort > math.MaxUint16 {
			return errors.New("invalid udp port number")
		}
		localnode.SetFallbackUDP(int(udpPort))
		return nil
	}
}

func Update(localnode *enode.LocalNode, enrOptions ...ENROption) error {
	for _, opt := range enrOptions {
		err := opt(localnode)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeMultiaddressField(localnode *enode.LocalNode, addrAggr []multiaddr.Multiaddr) (err error) {
	defer func() {
		if e := recover(); e != nil {
			// Deleting the multiaddr entry, as we could not write it succesfully
			localnode.Delete(enr.WithEntry(MultiaddrENRField, struct{}{}))
			err = errors.New("could not write enr record")
		}
	}()

	var fieldRaw []byte
	for _, addr := range addrAggr {
		maRaw := addr.Bytes()
		maSize := make([]byte, 2)
		binary.BigEndian.PutUint16(maSize, uint16(len(maRaw)))

		fieldRaw = append(fieldRaw, maSize...)
		fieldRaw = append(fieldRaw, maRaw...)
	}

	if len(fieldRaw) != 0 && len(fieldRaw) <= 100 { // Max length for multiaddr field before triggering the 300 bytes limit
		localnode.Set(enr.WithEntry(MultiaddrENRField, fieldRaw))
	}

	// This is to trigger the signing record err due to exceeding 300bytes limit
	_ = localnode.Node()

	return nil
}
