package enr

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/waku-org/go-waku/waku/v2/utils"
)

// WakuENRField is the name of the ENR field that contains information about which protocols are supported by the node
const WakuENRField = "waku2"

// MultiaddrENRField is the name of the ENR field that will contain multiaddresses that cannot be described using the
// already available ENR fields (i.e. in the case of websocket connections)
const MultiaddrENRField = "multiaddrs"

const ShardingIndicesListEnrField = "rs"

const ShardingBitVectorEnrField = "rsv"

// WakuEnrBitfield is a8-bit flag field to indicate Waku capabilities. Only the 4 LSBs are currently defined according to RFC31 (https://rfc.vac.dev/spec/31/).
type WakuEnrBitfield = uint8

// NewWakuEnrBitfield creates a WakuEnrBitField whose value will depend on which protocols are enabled in the node
func NewWakuEnrBitfield(lightpush, filter, store, relay bool) WakuEnrBitfield {
	var v uint8

	if lightpush {
		v |= (1 << 3)
	}

	if filter {
		v |= (1 << 2)
	}

	if store {
		v |= (1 << 1)
	}

	if relay {
		v |= (1 << 0)
	}

	return v
}

// EnodeToMultiaddress converts an enode into a multiaddress
func enodeToMultiAddr(node *enode.Node) (multiaddr.Multiaddr, error) {
	pubKey := utils.EcdsaPubKeyToSecp256k1PublicKey(node.Pubkey())
	peerID, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		return nil, err
	}

	ipType := "ip4"
	portNumber := node.TCP()
	if utils.IsIPv6(node.IP().String()) {
		ipType = "ip6"
		var port enr.TCP6
		if err := node.Record().Load(&port); err != nil {
			return nil, err
		}
		portNumber = int(port)
	}

	return multiaddr.NewMultiaddr(fmt.Sprintf("/%s/%s/tcp/%d/p2p/%s", ipType, node.IP(), portNumber, peerID))
}

// Multiaddress is used to extract all the multiaddresses that are part of a ENR record
func Multiaddress(node *enode.Node) (peer.ID, []multiaddr.Multiaddr, error) {
	pubKey := utils.EcdsaPubKeyToSecp256k1PublicKey(node.Pubkey())
	peerID, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		return "", nil, err
	}

	var result []multiaddr.Multiaddr

	addr, err := enodeToMultiAddr(node)
	if err != nil {
		return "", nil, err
	}
	result = append(result, addr)

	var multiaddrRaw []byte
	if err := node.Record().Load(enr.WithEntry(MultiaddrENRField, &multiaddrRaw)); err != nil {
		if !enr.IsNotFound(err) {
			return "", nil, err
		}
		// No multiaddr entry on enr
		return peerID, result, nil
	}

	if len(multiaddrRaw) < 2 {
		// There was no error loading the multiaddr field, but its length is incorrect
		return peerID, result, nil
	}

	offset := 0
	for {
		maSize := binary.BigEndian.Uint16(multiaddrRaw[offset : offset+2])
		if len(multiaddrRaw) < offset+2+int(maSize) {
			return "", nil, errors.New("invalid multiaddress field length")
		}
		maRaw := multiaddrRaw[offset+2 : offset+2+int(maSize)]
		addr, err := multiaddr.NewMultiaddrBytes(maRaw)
		if err != nil {
			return "", nil, fmt.Errorf("invalid multiaddress field length")
		}

		hostInfoStr := fmt.Sprintf("/p2p/%s", peerID.Pretty())
		_, pID := peer.SplitAddr(addr)
		if pID != "" && pID != peerID {
			// Addresses in the ENR that contain a p2p component are circuit relay addr
			hostInfoStr = "/p2p-circuit" + hostInfoStr
		}

		hostInfo, err := multiaddr.NewMultiaddr(hostInfoStr)
		if err != nil {
			return "", nil, err
		}
		result = append(result, addr.Encapsulate(hostInfo))

		offset += 2 + int(maSize)
		if offset >= len(multiaddrRaw) {
			break
		}
	}

	return peerID, result, nil
}

// EnodeToPeerInfo extracts the peer ID and multiaddresses defined in an ENR
func EnodeToPeerInfo(node *enode.Node) (*peer.AddrInfo, error) {
	_, addresses, err := Multiaddress(node)
	if err != nil {
		return nil, err
	}

	res, err := peer.AddrInfosFromP2pAddrs(addresses...)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, errors.New("could not retrieve peer addresses from enr")
	}
	return &res[0], nil
}
