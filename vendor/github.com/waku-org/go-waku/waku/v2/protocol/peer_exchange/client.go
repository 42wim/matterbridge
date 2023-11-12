package peer_exchange

import (
	"bytes"
	"context"
	"math"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
	"github.com/waku-org/go-waku/waku/v2/peermanager"
	"github.com/waku-org/go-waku/waku/v2/peerstore"
	wenr "github.com/waku-org/go-waku/waku/v2/protocol/enr"
	"github.com/waku-org/go-waku/waku/v2/protocol/peer_exchange/pb"
	"github.com/waku-org/go-waku/waku/v2/service"
	"go.uber.org/zap"
)

func (wakuPX *WakuPeerExchange) Request(ctx context.Context, numPeers int, opts ...PeerExchangeOption) error {
	params := new(PeerExchangeParameters)
	params.host = wakuPX.h
	params.log = wakuPX.log
	params.pm = wakuPX.pm

	optList := DefaultOptions(wakuPX.h)
	optList = append(optList, opts...)
	for _, opt := range optList {
		err := opt(params)
		if err != nil {
			return err
		}
	}

	if params.pm != nil && params.peerAddr != nil {
		pData, err := wakuPX.pm.AddPeer(params.peerAddr, peerstore.Static, []string{}, PeerExchangeID_v20alpha1)
		if err != nil {
			return err
		}
		wakuPX.pm.Connect(pData)
		params.selectedPeer = pData.AddrInfo.ID
	}

	if params.pm != nil && params.selectedPeer == "" {
		var err error
		params.selectedPeer, err = wakuPX.pm.SelectPeer(
			peermanager.PeerSelectionCriteria{
				SelectionType: params.peerSelectionType,
				Proto:         PeerExchangeID_v20alpha1,
				SpecificPeers: params.preferredPeers,
				Ctx:           ctx,
			},
		)
		if err != nil {
			return err
		}
	}
	if params.selectedPeer == "" {
		wakuPX.metrics.RecordError(dialFailure)
		return ErrNoPeersAvailable
	}

	requestRPC := &pb.PeerExchangeRPC{
		Query: &pb.PeerExchangeQuery{
			NumPeers: uint64(numPeers),
		},
	}

	stream, err := wakuPX.h.NewStream(ctx, params.selectedPeer, PeerExchangeID_v20alpha1)
	if err != nil {
		return err
	}

	writer := pbio.NewDelimitedWriter(stream)
	err = writer.WriteMsg(requestRPC)
	if err != nil {
		if err := stream.Reset(); err != nil {
			wakuPX.log.Error("resetting connection", zap.Error(err))
		}
		return err
	}

	reader := pbio.NewDelimitedReader(stream, math.MaxInt32)
	responseRPC := &pb.PeerExchangeRPC{}
	err = reader.ReadMsg(responseRPC)
	if err != nil {
		if err := stream.Reset(); err != nil {
			wakuPX.log.Error("resetting connection", zap.Error(err))
		}
		return err
	}

	stream.Close()

	return wakuPX.handleResponse(ctx, responseRPC.Response)
}

func (wakuPX *WakuPeerExchange) handleResponse(ctx context.Context, response *pb.PeerExchangeResponse) error {
	var discoveredPeers []struct {
		addrInfo peer.AddrInfo
		enr      *enode.Node
	}

	for _, p := range response.PeerInfos {
		enrRecord := &enr.Record{}
		buf := bytes.NewBuffer(p.Enr)

		err := enrRecord.DecodeRLP(rlp.NewStream(buf, uint64(len(p.Enr))))
		if err != nil {
			wakuPX.log.Error("converting bytes to enr", zap.Error(err))
			return err
		}

		enodeRecord, err := enode.New(enode.ValidSchemes, enrRecord)
		if err != nil {
			wakuPX.log.Error("creating enode record", zap.Error(err))
			return err
		}

		addrInfo, err := wenr.EnodeToPeerInfo(enodeRecord)
		if err != nil {
			return err
		}

		discoveredPeers = append(discoveredPeers, struct {
			addrInfo peer.AddrInfo
			enr      *enode.Node
		}{
			addrInfo: *addrInfo,
			enr:      enodeRecord,
		})
	}

	if len(discoveredPeers) != 0 {
		wakuPX.log.Info("connecting to newly discovered peers", zap.Int("count", len(discoveredPeers)))
		wakuPX.WaitGroup().Add(1)
		go func() {
			defer wakuPX.WaitGroup().Done()

			peerCh := make(chan service.PeerData)
			defer close(peerCh)
			wakuPX.peerConnector.Subscribe(ctx, peerCh)
			for _, p := range discoveredPeers {
				peer := service.PeerData{
					Origin:   peerstore.PeerExchange,
					AddrInfo: p.addrInfo,
					ENR:      p.enr,
				}
				select {
				case <-ctx.Done():
					return
				case peerCh <- peer:
				}
			}
		}()
	}

	return nil
}
