package tracker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/anacrolix/dht/v2/krpc"
	"github.com/anacrolix/log"
	trHttp "github.com/anacrolix/torrent/tracker/http"
	"github.com/anacrolix/torrent/tracker/shared"
	"github.com/anacrolix/torrent/tracker/udp"
)

const (
	None      = shared.None
	Started   = shared.Started
	Stopped   = shared.Stopped
	Completed = shared.Completed
)

type AnnounceRequest = udp.AnnounceRequest

type AnnounceResponse = trHttp.AnnounceResponse

type Peer = trHttp.Peer

type AnnounceEvent = udp.AnnounceEvent

var ErrBadScheme = errors.New("unknown scheme")

type Announce struct {
	TrackerUrl string
	Request    AnnounceRequest
	HostHeader string
	HTTPProxy  func(*http.Request) (*url.URL, error)
	ServerName string
	UserAgent  string
	UdpNetwork string
	// If the port is zero, it's assumed to be the same as the Request.Port.
	ClientIp4 krpc.NodeAddr
	// If the port is zero, it's assumed to be the same as the Request.Port.
	ClientIp6 krpc.NodeAddr
	Context   context.Context
	Logger    log.Logger
}

// The code *is* the documentation.
const DefaultTrackerAnnounceTimeout = 15 * time.Second

func (me Announce) Do() (res AnnounceResponse, err error) {
	cl, err := NewClient(me.TrackerUrl, NewClientOpts{
		Http: trHttp.NewClientOpts{
			Proxy:      me.HTTPProxy,
			ServerName: me.ServerName,
		},
		UdpNetwork: me.UdpNetwork,
		Logger:     me.Logger.WithContextValue(fmt.Sprintf("tracker client for %q", me.TrackerUrl)),
	})
	if err != nil {
		return
	}
	defer cl.Close()
	if me.Context == nil {
		// This is just to maintain the old behaviour that should be a timeout of 15s. Users can
		// override it by providing their own Context. See comments elsewhere about longer timeouts
		// acting as rate limiting overloaded trackers.
		ctx, cancel := context.WithTimeout(context.Background(), DefaultTrackerAnnounceTimeout)
		defer cancel()
		me.Context = ctx
	}
	return cl.Announce(me.Context, me.Request, trHttp.AnnounceOpt{
		UserAgent:  me.UserAgent,
		HostHeader: me.HostHeader,
		ClientIp4:  me.ClientIp4.IP,
		ClientIp6:  me.ClientIp6.IP,
	})
}
