package metadata

import (
	"context"
	"errors"
	"math"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	libp2pProtocol "github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-msgio/pbio"
	"github.com/multiformats/go-multiaddr"
	"github.com/waku-org/go-waku/logging"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/enr"
	"github.com/waku-org/go-waku/waku/v2/protocol/metadata/pb"
	"go.uber.org/zap"
)

// MetadataID_v1 is the current Waku Metadata protocol identifier
const MetadataID_v1 = libp2pProtocol.ID("/vac/waku/metadata/1.0.0")

// WakuMetadata is the implementation of the Waku Metadata protocol
type WakuMetadata struct {
	network.Notifiee

	h         host.Host
	ctx       context.Context
	cancel    context.CancelFunc
	clusterID uint16
	localnode *enode.LocalNode

	log *zap.Logger
}

// NewWakuMetadata returns a new instance of Waku Metadata struct
// Takes an optional peermanager if WakuLightPush is being created along with WakuNode.
// If using libp2p host, then pass peermanager as nil
func NewWakuMetadata(clusterID uint16, localnode *enode.LocalNode, log *zap.Logger) *WakuMetadata {
	m := new(WakuMetadata)
	m.log = log.Named("metadata")
	m.clusterID = clusterID
	m.localnode = localnode

	return m
}

// Sets the host to be able to mount or consume a protocol
func (wakuM *WakuMetadata) SetHost(h host.Host) {
	wakuM.h = h
}

// Start inits the metadata protocol
func (wakuM *WakuMetadata) Start(ctx context.Context) error {
	if wakuM.clusterID == 0 {
		wakuM.log.Warn("no clusterID is specified. Protocol will not be initialized")
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)

	wakuM.ctx = ctx
	wakuM.cancel = cancel

	wakuM.h.Network().Notify(wakuM)

	wakuM.h.SetStreamHandlerMatch(MetadataID_v1, protocol.PrefixTextMatch(string(MetadataID_v1)), wakuM.onRequest(ctx))
	wakuM.log.Info("metadata protocol started")
	return nil
}

func (wakuM *WakuMetadata) getClusterAndShards() (*uint32, []uint32, error) {
	shard, err := enr.RelaySharding(wakuM.localnode.Node().Record())
	if err != nil {
		return nil, nil, err
	}

	var shards []uint32
	if shard != nil && shard.ClusterID == uint16(wakuM.clusterID) {
		for _, idx := range shard.ShardIDs {
			shards = append(shards, uint32(idx))
		}
	}

	u32ClusterID := uint32(wakuM.clusterID)

	return &u32ClusterID, shards, nil
}

func (wakuM *WakuMetadata) Request(ctx context.Context, peerID peer.ID) (*protocol.RelayShards, error) {
	logger := wakuM.log.With(logging.HostID("peer", peerID))

	stream, err := wakuM.h.NewStream(ctx, peerID, MetadataID_v1)
	if err != nil {
		logger.Error("creating stream to peer", zap.Error(err))
		return nil, err
	}

	clusterID, shards, err := wakuM.getClusterAndShards()
	if err != nil {
		if err := stream.Reset(); err != nil {
			wakuM.log.Error("resetting connection", zap.Error(err))
		}
		return nil, err
	}

	request := &pb.WakuMetadataRequest{}
	request.ClusterId = clusterID
	request.Shards = shards

	writer := pbio.NewDelimitedWriter(stream)
	reader := pbio.NewDelimitedReader(stream, math.MaxInt32)

	err = writer.WriteMsg(request)
	if err != nil {
		logger.Error("writing request", zap.Error(err))
		if err := stream.Reset(); err != nil {
			wakuM.log.Error("resetting connection", zap.Error(err))
		}
		return nil, err
	}

	response := &pb.WakuMetadataResponse{}
	err = reader.ReadMsg(response)
	if err != nil {
		logger.Error("reading response", zap.Error(err))
		if err := stream.Reset(); err != nil {
			wakuM.log.Error("resetting connection", zap.Error(err))
		}
		return nil, err
	}

	stream.Close()

	if response.ClusterId == nil {
		return nil, errors.New("node did not provide a waku clusterid")
	}

	rClusterID := uint16(*response.ClusterId)
	var rShardIDs []uint16
	for _, i := range response.Shards {
		rShardIDs = append(rShardIDs, uint16(i))
	}

	rs, err := protocol.NewRelayShards(rClusterID, rShardIDs...)
	if err != nil {
		return nil, err
	}

	return &rs, nil
}

func (wakuM *WakuMetadata) onRequest(ctx context.Context) func(network.Stream) {
	return func(stream network.Stream) {
		logger := wakuM.log.With(logging.HostID("peer", stream.Conn().RemotePeer()))
		request := &pb.WakuMetadataRequest{}

		writer := pbio.NewDelimitedWriter(stream)
		reader := pbio.NewDelimitedReader(stream, math.MaxInt32)

		err := reader.ReadMsg(request)
		if err != nil {
			logger.Error("reading request", zap.Error(err))
			if err := stream.Reset(); err != nil {
				wakuM.log.Error("resetting connection", zap.Error(err))
			}
			return
		}

		response := new(pb.WakuMetadataResponse)

		clusterID, shards, err := wakuM.getClusterAndShards()
		if err != nil {
			logger.Error("obtaining shard info", zap.Error(err))
		} else {
			response.ClusterId = clusterID
			response.Shards = shards
		}

		err = writer.WriteMsg(response)
		if err != nil {
			logger.Error("writing response", zap.Error(err))
			if err := stream.Reset(); err != nil {
				wakuM.log.Error("resetting connection", zap.Error(err))
			}
			return
		}

		stream.Close()
	}
}

// Stop unmounts the metadata protocol
func (wakuM *WakuMetadata) Stop() {
	if wakuM.cancel == nil {
		return
	}

	wakuM.h.Network().StopNotify(wakuM)
	wakuM.cancel()
	wakuM.h.RemoveStreamHandler(MetadataID_v1)

}

// Listen is called when network starts listening on an addr
func (wakuM *WakuMetadata) Listen(n network.Network, m multiaddr.Multiaddr) {
	// Do nothing
}

// ListenClose is called when network stops listening on an address
func (wakuM *WakuMetadata) ListenClose(n network.Network, m multiaddr.Multiaddr) {
	// Do nothing
}

func (wakuM *WakuMetadata) disconnectPeer(peerID peer.ID, reason error) {
	logger := wakuM.log.With(logging.HostID("peerID", peerID))
	logger.Error("disconnecting from peer", zap.Error(reason))
	wakuM.h.Peerstore().RemovePeer(peerID)
	if err := wakuM.h.Network().ClosePeer(peerID); err != nil {
		logger.Error("could not disconnect from peer", zap.Error(err))
	}
}

// Connected is called when a connection is opened
func (wakuM *WakuMetadata) Connected(n network.Network, cc network.Conn) {
	go func() {
		// Metadata verification is done only if a clusterID is specified
		if wakuM.clusterID == 0 {
			return
		}

		peerID := cc.RemotePeer()

		shard, err := wakuM.Request(wakuM.ctx, peerID)
		if err != nil {
			wakuM.disconnectPeer(peerID, err)
			return
		}

		if shard.ClusterID != wakuM.clusterID {
			wakuM.disconnectPeer(peerID, errors.New("different clusterID reported"))
		}
	}()
}

// Disconnected is called when a connection closed
func (wakuM *WakuMetadata) Disconnected(n network.Network, cc network.Conn) {
	// Do nothing
}
