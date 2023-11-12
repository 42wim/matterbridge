package rendezvous

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	rvs "github.com/waku-org/go-libp2p-rendezvous"
	"github.com/waku-org/go-waku/waku/v2/peerstore"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/service"
	"go.uber.org/zap"
)

// RendezvousID is the current protocol ID used for Rendezvous
const RendezvousID = rvs.RendezvousProto

// RegisterDefaultTTL indicates the TTL used by default when registering a node in a rendezvous point
// TODO: Register* functions should allow setting up a custom TTL
const RegisterDefaultTTL = rvs.DefaultTTL * time.Second

// Rendezvous is the implementation containing the logic to registering a node and discovering new peers using rendezvous protocol
type Rendezvous struct {
	host host.Host

	db            *DB
	rendezvousSvc *rvs.RendezvousService

	peerConnector PeerConnector

	log *zap.Logger
	*service.CommonDiscoveryService
}

// PeerConnector will subscribe to a channel containing the information for all peers found by this discovery protocol
type PeerConnector interface {
	Subscribe(context.Context, <-chan service.PeerData)
}

// NewRendezvous creates an instance of Rendezvous struct
func NewRendezvous(db *DB, peerConnector PeerConnector, log *zap.Logger) *Rendezvous {
	logger := log.Named("rendezvous")
	return &Rendezvous{
		db:                     db,
		peerConnector:          peerConnector,
		log:                    logger,
		CommonDiscoveryService: service.NewCommonDiscoveryService(),
	}
}

// Sets the host to be able to mount or consume a protocol
func (r *Rendezvous) SetHost(h host.Host) {
	r.host = h
}

func (r *Rendezvous) Start(ctx context.Context) error {
	return r.CommonDiscoveryService.Start(ctx, r.start)
}

func (r *Rendezvous) start() error {
	if r.db != nil {
		if err := r.db.Start(r.Context()); err != nil {
			return err
		}
	}
	if r.peerConnector != nil {
		r.peerConnector.Subscribe(r.Context(), r.GetListeningChan())
	}

	r.rendezvousSvc = rvs.NewRendezvousService(r.host, r.db)

	r.log.Info("rendezvous protocol started")
	return nil
}

const registerBackoff = 200 * time.Millisecond
const registerMaxRetries = 7

// Discover is used to find a number of peers that use the default pubsub topic
func (r *Rendezvous) Discover(ctx context.Context, rp *RendezvousPoint, numPeers int) {
	r.DiscoverWithNamespace(ctx, protocol.DefaultPubsubTopic{}.String(), rp, numPeers)
}

// DiscoverShard is used to find a number of peers that support an specific cluster and shard index
func (r *Rendezvous) DiscoverShard(ctx context.Context, rp *RendezvousPoint, cluster uint16, shard uint16, numPeers int) {
	namespace := ShardToNamespace(cluster, shard)
	r.DiscoverWithNamespace(ctx, namespace, rp, numPeers)
}

// DiscoverWithNamespace is used to find a number of peers using a custom namespace (usually a pubsub topic)
func (r *Rendezvous) DiscoverWithNamespace(ctx context.Context, namespace string, rp *RendezvousPoint, numPeers int) {
	rendezvousClient := rvs.NewRendezvousClient(r.host, rp.id)

	addrInfo, cookie, err := rendezvousClient.Discover(ctx, namespace, numPeers, rp.cookie)
	if err != nil {
		r.log.Error("could not discover new peers", zap.Error(err))
		rp.Delay()
		return
	}

	if len(addrInfo) != 0 {
		rp.SetSuccess(cookie)

		for _, p := range addrInfo {
			peer := service.PeerData{
				Origin:       peerstore.Rendezvous,
				AddrInfo:     p,
				PubsubTopics: []string{namespace},
			}
			if !r.PushToChan(peer) {
				r.log.Error("could push to closed channel/context completed")
				return
			}
		}
	} else {
		rp.Delay()
	}

}

func (r *Rendezvous) callRegister(ctx context.Context, namespace string, rendezvousClient rvs.RendezvousClient, retries int) (<-chan time.Time, int) {
	ttl, err := rendezvousClient.Register(ctx, namespace, rvs.DefaultTTL)
	var t <-chan time.Time
	if err != nil {
		r.log.Error("registering rendezvous client", zap.Error(err))
		backoff := registerBackoff * time.Duration(math.Exp2(float64(retries)))
		t = time.After(backoff)
		retries++
	} else {
		t = time.After(ttl)
	}

	return t, retries
}

// Register registers the node in the rendezvous points using the default pubsub topic as namespace
func (r *Rendezvous) Register(ctx context.Context, rendezvousPoints []*RendezvousPoint) {
	r.RegisterWithNamespace(ctx, protocol.DefaultPubsubTopic{}.String(), rendezvousPoints)
}

// RegisterShard registers the node in the rendezvous points using a shard as namespace
func (r *Rendezvous) RegisterShard(ctx context.Context, cluster uint16, shard uint16, rendezvousPoints []*RendezvousPoint) {
	namespace := ShardToNamespace(cluster, shard)
	r.RegisterWithNamespace(ctx, namespace, rendezvousPoints)
}

// RegisterRelayShards registers the node in the rendezvous point by specifying a RelayShards struct (more than one shard index can be registered)
func (r *Rendezvous) RegisterRelayShards(ctx context.Context, rs protocol.RelayShards, rendezvousPoints []*RendezvousPoint) {
	for _, idx := range rs.ShardIDs {
		go r.RegisterShard(ctx, rs.ClusterID, idx, rendezvousPoints)
	}
}

// RegisterWithNamespace registers the node in the rendezvous point by using an specific namespace (usually a pubsub topic)
func (r *Rendezvous) RegisterWithNamespace(ctx context.Context, namespace string, rendezvousPoints []*RendezvousPoint) {
	for _, m := range rendezvousPoints {
		r.WaitGroup().Add(1)
		go func(m *RendezvousPoint) {
			r.WaitGroup().Done()

			rendezvousClient := rvs.NewRendezvousClient(r.host, m.id)
			retries := 0
			var t <-chan time.Time

			t, retries = r.callRegister(ctx, namespace, rendezvousClient, retries)
			for {
				select {
				case <-ctx.Done():
					return
				case <-t:
					t, retries = r.callRegister(ctx, namespace, rendezvousClient, retries)
					if retries >= registerMaxRetries {
						return
					}
				}
			}
		}(m)
	}
}

func (r *Rendezvous) Stop() {
	r.CommonDiscoveryService.Stop(func() {
		r.host.RemoveStreamHandler(rvs.RendezvousProto)
		r.rendezvousSvc = nil
	})
}

// ShardToNamespace translates a cluster and shard index into a rendezvous namespace
func ShardToNamespace(cluster uint16, shard uint16) string {
	return fmt.Sprintf("rs/%d/%d", cluster, shard)
}
