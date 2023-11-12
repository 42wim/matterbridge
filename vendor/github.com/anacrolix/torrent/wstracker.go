package torrent

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/anacrolix/log"
	"github.com/anacrolix/torrent/tracker/http"
	"github.com/gorilla/websocket"

	"github.com/anacrolix/torrent/tracker"
	"github.com/anacrolix/torrent/webtorrent"
	"github.com/pion/datachannel"
)

type websocketTrackerStatus struct {
	url url.URL
	tc  *webtorrent.TrackerClient
}

func (me websocketTrackerStatus) statusLine() string {
	return fmt.Sprintf("%+v", me.tc.Stats())
}

func (me websocketTrackerStatus) URL() *url.URL {
	return &me.url
}

type refCountedWebtorrentTrackerClient struct {
	webtorrent.TrackerClient
	refCount int
}

type websocketTrackers struct {
	PeerId             [20]byte
	Logger             log.Logger
	GetAnnounceRequest func(event tracker.AnnounceEvent, infoHash [20]byte) (tracker.AnnounceRequest, error)
	OnConn             func(datachannel.ReadWriteCloser, webtorrent.DataChannelContext)
	mu                 sync.Mutex
	clients            map[string]*refCountedWebtorrentTrackerClient
	Proxy              http.ProxyFunc
}

func (me *websocketTrackers) Get(url string) (*webtorrent.TrackerClient, func()) {
	me.mu.Lock()
	defer me.mu.Unlock()
	value, ok := me.clients[url]
	if !ok {
		dialer := &websocket.Dialer{Proxy: me.Proxy, HandshakeTimeout: websocket.DefaultDialer.HandshakeTimeout}
		value = &refCountedWebtorrentTrackerClient{
			TrackerClient: webtorrent.TrackerClient{
				Dialer:             dialer,
				Url:                url,
				GetAnnounceRequest: me.GetAnnounceRequest,
				PeerId:             me.PeerId,
				OnConn:             me.OnConn,
				Logger: me.Logger.WithText(func(m log.Msg) string {
					return fmt.Sprintf("tracker client for %q: %v", url, m)
				}),
			},
		}
		value.TrackerClient.Start(func(err error) {
			if err != nil {
				me.Logger.Printf("error running tracker client for %q: %v", url, err)
			}
		})
		if me.clients == nil {
			me.clients = make(map[string]*refCountedWebtorrentTrackerClient)
		}
		me.clients[url] = value
	}
	value.refCount++
	return &value.TrackerClient, func() {
		me.mu.Lock()
		defer me.mu.Unlock()
		value.refCount--
		if value.refCount == 0 {
			value.TrackerClient.Close()
			delete(me.clients, url)
		}
	}
}
