// logging implements custom logging field types for commonly
// logged values like host ID or wallet address.
//
// implementation purposely does as little as possible at field creation time,
// and postpones any transformation to output time by relying on the generic
// zap types like zap.Stringer, zap.Array, zap.Object
package logging

import (
	"encoding/hex"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/waku-org/go-waku/waku/v2/protocol/store/pb"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// List of []byte
type byteArr [][]byte

// HexArray creates a field with an array of bytes that will be shown as a hexadecimal string in logs
func HexArray(key string, byteVal byteArr) zapcore.Field {
	return zap.Array(key, byteVal)
}

func (bArr byteArr) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, b := range bArr {
		encoder.AppendString("0x" + hex.EncodeToString(b))
	}
	return nil
}

// List of multiaddrs
type multiaddrs []multiaddr.Multiaddr

// MultiAddrs creates a field with an array of multiaddrs
func MultiAddrs(key string, addrs ...multiaddr.Multiaddr) zapcore.Field {
	return zap.Array(key, multiaddrs(addrs))
}

func (addrs multiaddrs) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, addr := range addrs {
		encoder.AppendString(addr.String())
	}
	return nil
}

// Host ID/Peer ID
type hostID peer.ID

// HostID creates a field for a peer.ID
func HostID(key string, id peer.ID) zapcore.Field {
	return zap.Stringer(key, hostID(id))
}

func (id hostID) String() string { return peer.ID(id).String() }

// Time - Waku uses Nanosecond Unix Time
type timestamp int64

// Time creates a field for Waku time value
func Time(key string, time int64) zapcore.Field {
	return zap.Stringer(key, timestamp(time))
}

func (t timestamp) String() string {
	return time.Unix(0, int64(t)).Format(time.RFC3339)
}

// History Query Filters
type historyFilters []*pb.ContentFilter

// Filters creates a field with an array of history query filters.
// The assumption is that log entries won't have more than one of these,
// so the field key/name is hardcoded to be "filters" to promote consistency.
func Filters(filters []*pb.ContentFilter) zapcore.Field {
	return zap.Array("filters", historyFilters(filters))
}

func (filters historyFilters) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, filter := range filters {
		encoder.AppendString(filter.ContentTopic)
	}
	return nil
}

// History Paging Info
// Probably too detailed for normal log levels, but useful for debugging.
// Also a good example of nested object value.
type pagingInfo pb.PagingInfo
type index pb.Index

// PagingInfo creates a field with history query paging info.
func PagingInfo(pi *pb.PagingInfo) zapcore.Field {
	return zap.Object("paging_info", (*pagingInfo)(pi))
}

func (pi *pagingInfo) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint64("page_size", pi.PageSize)
	encoder.AddString("direction", pi.Direction.String())
	if pi.Cursor != nil {
		return encoder.AddObject("cursor", (*index)(pi.Cursor))
	}
	return nil
}

func (i *index) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddBinary("digest", i.Digest)
	encoder.AddTime("sent", time.Unix(0, i.SenderTime))
	encoder.AddTime("received", time.Unix(0, i.ReceiverTime))
	return nil
}

// Hex encoded bytes
type hexBytes []byte

// HexBytes creates a field for a byte slice that should be emitted as hex encoded string.
func HexBytes(key string, bytes []byte) zap.Field {
	return zap.Stringer(key, hexBytes(bytes))
}

func (bytes hexBytes) String() string {
	return hexutil.Encode(bytes)
}

// ENode creates a field for ENR node.
func ENode(key string, node *enode.Node) zap.Field {
	return zap.Stringer(key, node)
}

// TCPAddr creates a field for TCP v4/v6 address and port
func TCPAddr(key string, ip net.IP, port int) zap.Field {
	return zap.Stringer(key, &net.TCPAddr{IP: ip, Port: port})
}

// UDPAddr creates a field for UDP v4/v6 address and port
func UDPAddr(key string, ip net.IP, port int) zap.Field {
	return zap.Stringer(key, &net.UDPAddr{IP: ip, Port: port})
}
