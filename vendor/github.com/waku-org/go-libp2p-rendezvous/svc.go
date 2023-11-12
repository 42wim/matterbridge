package rendezvous

import (
	"github.com/libp2p/go-libp2p/core/host"
	inet "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"

	db "github.com/waku-org/go-libp2p-rendezvous/db"
	pb "github.com/waku-org/go-libp2p-rendezvous/pb"
)

const (
	MaxTTL               = 72 * 3600 // 72hr
	MaxNamespaceLength   = 256
	MaxPeerAddressLength = 2048
	MaxRegistrations     = 1000
	MaxDiscoverLimit     = 1000
)

type RendezvousService struct {
	DB db.DB
}

func NewRendezvousService(host host.Host, db db.DB) *RendezvousService {
	rz := &RendezvousService{DB: db}
	host.SetStreamHandler(RendezvousProto, rz.handleStream)
	return rz
}

func (rz *RendezvousService) handleStream(s inet.Stream) {
	defer s.Reset()

	pid := s.Conn().RemotePeer()
	log.Debugf("New stream from %s", pid.Pretty())

	r := pbio.NewDelimitedReader(s, inet.MessageSizeMax)
	w := pbio.NewDelimitedWriter(s)

	for {
		var req pb.Message
		var res pb.Message

		err := r.ReadMsg(&req)
		if err != nil {
			return
		}

		t := req.GetType()
		switch t {
		case pb.Message_REGISTER:
			r := rz.handleRegister(pid, req.GetRegister())
			res.Type = pb.Message_REGISTER_RESPONSE.Enum()
			res.RegisterResponse = r
			err = w.WriteMsg(&res)
			if err != nil {
				log.Debugf("Error writing response: %s", err.Error())
				return
			}

		case pb.Message_UNREGISTER:
			err := rz.handleUnregister(pid, req.GetUnregister())
			if err != nil {
				log.Debugf("Error unregistering peer: %s", err.Error())
			}

		case pb.Message_DISCOVER:
			r := rz.handleDiscover(pid, req.GetDiscover())
			res.Type = pb.Message_DISCOVER_RESPONSE.Enum()
			res.DiscoverResponse = r
			err = w.WriteMsg(&res)
			if err != nil {
				log.Debugf("Error writing response: %s", err.Error())
				return
			}

		default:
			log.Debugf("Unexpected message: %s", t.String())
			return
		}
	}
}

func (rz *RendezvousService) handleRegister(p peer.ID, m *pb.Message_Register) *pb.Message_RegisterResponse {
	ns := m.GetNs()
	if ns == "" {
		return newRegisterResponseError(pb.Message_E_INVALID_NAMESPACE, "unspecified namespace")
	}

	if len(ns) > MaxNamespaceLength {
		return newRegisterResponseError(pb.Message_E_INVALID_NAMESPACE, "namespace too long")
	}

	signedPeerRecord := m.GetSignedPeerRecord()
	if signedPeerRecord == nil {
		return newRegisterResponseError(pb.Message_E_INVALID_SIGNED_PEER_RECORD, "missing signed peer record")
	}

	peerRecord, err := pbToPeerRecord(signedPeerRecord)
	if err != nil {
		return newRegisterResponseError(pb.Message_E_INVALID_SIGNED_PEER_RECORD, "invalid peer record")
	}

	if peerRecord.ID != p {
		return newRegisterResponseError(pb.Message_E_INVALID_SIGNED_PEER_RECORD, "peer id mismatch")
	}

	if len(peerRecord.Addrs) == 0 {
		return newRegisterResponseError(pb.Message_E_INVALID_SIGNED_PEER_RECORD, "missing peer addresses")
	}

	mlen := 0
	for _, maddr := range peerRecord.Addrs {
		mlen += len(maddr.Bytes())
	}
	if mlen > MaxPeerAddressLength {
		return newRegisterResponseError(pb.Message_E_INVALID_SIGNED_PEER_RECORD, "peer info too long")
	}

	// Note:
	// We don't validate the addresses, because they could include protocols we don't understand
	// Perhaps we should though.

	mttl := m.GetTtl()
	if mttl > MaxTTL {
		return newRegisterResponseError(pb.Message_E_INVALID_TTL, "bad ttl")
	}

	ttl := DefaultTTL
	if mttl > 0 {
		ttl = int(mttl)
	}

	// now check how many registrations we have for this peer -- simple limit to defend
	// against trivial DoS attacks (eg a peer connects and keeps registering until it
	// fills our db)
	rcount, err := rz.DB.CountRegistrations(p)
	if err != nil {
		log.Errorf("Error counting registrations: %s", err.Error())
		return newRegisterResponseError(pb.Message_E_INTERNAL_ERROR, "database error")
	}

	if rcount > MaxRegistrations {
		log.Warningf("Too many registrations for %s", p)
		return newRegisterResponseError(pb.Message_E_NOT_AUTHORIZED, "too many registrations")
	}

	// ok, seems like we can register
	_, err = rz.DB.Register(p, ns, signedPeerRecord, ttl)
	if err != nil {
		log.Errorf("Error registering: %s", err.Error())
		return newRegisterResponseError(pb.Message_E_INTERNAL_ERROR, "database error")
	}

	log.Infof("registered peer %s %s (%d)", p, ns, ttl)

	return newRegisterResponse(ttl)
}

func (rz *RendezvousService) handleUnregister(p peer.ID, m *pb.Message_Unregister) error {
	ns := m.GetNs()

	err := rz.DB.Unregister(p, ns)
	if err != nil {
		return err
	}

	log.Infof("unregistered peer %s %s", p, ns)

	return nil
}

func (rz *RendezvousService) handleDiscover(p peer.ID, m *pb.Message_Discover) *pb.Message_DiscoverResponse {
	ns := m.GetNs()

	if len(ns) > MaxNamespaceLength {
		return newDiscoverResponseError(pb.Message_E_INVALID_NAMESPACE, "namespace too long")
	}

	limit := MaxDiscoverLimit
	mlimit := m.GetLimit()
	if mlimit > 0 && mlimit < uint64(limit) {
		limit = int(mlimit)
	}

	cookie := m.GetCookie()
	if cookie != nil && !rz.DB.ValidCookie(ns, cookie) {
		return newDiscoverResponseError(pb.Message_E_INVALID_COOKIE, "bad cookie")
	}

	regs, cookie, err := rz.DB.Discover(ns, cookie, limit)
	if err != nil {
		log.Errorf("Error in query: %s", err.Error())
		return newDiscoverResponseError(pb.Message_E_INTERNAL_ERROR, "database error")
	}

	log.Debugf("discover query: %s %s -> %d", p, ns, len(regs))

	return newDiscoverResponse(regs, cookie)
}
