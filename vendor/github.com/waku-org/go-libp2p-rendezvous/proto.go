package rendezvous

import (
	"errors"
	"fmt"
	"time"

	db "github.com/waku-org/go-libp2p-rendezvous/db"
	pb "github.com/waku-org/go-libp2p-rendezvous/pb"

	logging "github.com/ipfs/go-log/v2"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/record"
)

var log = logging.Logger("rendezvous")

const (
	RendezvousProto = protocol.ID("/rendezvous/1.0.0")

	DefaultTTL = 2 * 3600 // 2hr
)

type RendezvousError struct {
	Status pb.Message_ResponseStatus
	Text   string
}

func (e RendezvousError) Error() string {
	return fmt.Sprintf("Rendezvous error: %s (%s)", e.Text, e.Status.String())
}

func NewRegisterMessage(privKey crypto.PrivKey, ns string, pi peer.AddrInfo, ttl int) (*pb.Message, error) {
	return newRegisterMessage(privKey, ns, pi, ttl)
}

func newRegisterMessage(privKey crypto.PrivKey, ns string, pi peer.AddrInfo, ttl int) (*pb.Message, error) {
	msg := new(pb.Message)
	msg.Type = pb.Message_REGISTER.Enum()
	msg.Register = new(pb.Message_Register)
	if ns != "" {
		msg.Register.Ns = &ns
	}
	if ttl > 0 {
		ttlu64 := uint64(ttl)
		msg.Register.Ttl = &ttlu64
	}

	peerInfo := &peer.PeerRecord{
		PeerID: pi.ID,
		Addrs:  pi.Addrs,
		Seq:    uint64(time.Now().Unix()),
	}

	envelope, err := record.Seal(peerInfo, privKey)
	if err != nil {
		return nil, err
	}

	envPayload, err := envelope.Marshal()
	if err != nil {
		return nil, err
	}

	msg.Register.SignedPeerRecord = envPayload

	return msg, nil
}

func newUnregisterMessage(ns string, pid peer.ID) *pb.Message {
	msg := new(pb.Message)
	msg.Type = pb.Message_UNREGISTER.Enum()
	msg.Unregister = new(pb.Message_Unregister)
	if ns != "" {
		msg.Unregister.Ns = &ns
	}
	return msg
}

func NewDiscoverMessage(ns string, limit int, cookie []byte) *pb.Message {
	return newDiscoverMessage(ns, limit, cookie)
}

func newDiscoverMessage(ns string, limit int, cookie []byte) *pb.Message {
	msg := new(pb.Message)
	msg.Type = pb.Message_DISCOVER.Enum()
	msg.Discover = new(pb.Message_Discover)
	if ns != "" {
		msg.Discover.Ns = &ns
	}
	if limit > 0 {
		limitu64 := uint64(limit)
		msg.Discover.Limit = &limitu64
	}
	if cookie != nil {
		msg.Discover.Cookie = cookie
	}
	return msg
}
func pbToPeerRecord(envelopeBytes []byte) (peer.AddrInfo, error) {
	envelope, rec, err := record.ConsumeEnvelope(envelopeBytes, peer.PeerRecordEnvelopeDomain)
	if err != nil {
		return peer.AddrInfo{}, err
	}

	peerRec, ok := rec.(*peer.PeerRecord)
	if !ok {
		return peer.AddrInfo{}, errors.New("invalid peer record")
	}

	if !peerRec.PeerID.MatchesPublicKey(envelope.PublicKey) {
		return peer.AddrInfo{}, errors.New("signing key does not match peer record")
	}

	return peer.AddrInfo{ID: peerRec.PeerID, Addrs: peerRec.Addrs}, nil
}

func newRegisterResponse(ttl int) *pb.Message_RegisterResponse {
	ttlu64 := uint64(ttl)
	r := new(pb.Message_RegisterResponse)
	r.Status = pb.Message_OK.Enum()
	r.Ttl = &ttlu64
	return r
}

func newRegisterResponseError(status pb.Message_ResponseStatus, text string) *pb.Message_RegisterResponse {
	r := new(pb.Message_RegisterResponse)
	r.Status = status.Enum()
	r.StatusText = &text
	return r
}

func newDiscoverResponse(regs []db.RegistrationRecord, cookie []byte) *pb.Message_DiscoverResponse {
	r := new(pb.Message_DiscoverResponse)
	r.Status = pb.Message_OK.Enum()

	rregs := make([]*pb.Message_Register, len(regs))
	for i, reg := range regs {
		rreg := new(pb.Message_Register)
		rreg.Ns = &reg.Ns
		rreg.SignedPeerRecord = reg.SignedPeerRecord
		rttl := uint64(reg.Ttl)
		rreg.Ttl = &rttl
		rregs[i] = rreg
	}

	r.Registrations = rregs
	r.Cookie = cookie

	return r
}

func newDiscoverResponseError(status pb.Message_ResponseStatus, text string) *pb.Message_DiscoverResponse {
	r := new(pb.Message_DiscoverResponse)
	r.Status = status.Enum()
	r.StatusText = &text
	return r
}
