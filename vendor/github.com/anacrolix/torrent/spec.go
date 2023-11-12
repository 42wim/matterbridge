package torrent

import (
	"fmt"

	"github.com/anacrolix/torrent/metainfo"
	pp "github.com/anacrolix/torrent/peer_protocol"
	"github.com/anacrolix/torrent/storage"
)

// Specifies a new torrent for adding to a client, or additions to an existing Torrent. There are
// constructor functions for magnet URIs and torrent metainfo files. TODO: This type should be
// dismantled into a new Torrent option type, and separate Torrent mutate method(s).
type TorrentSpec struct {
	// The tiered tracker URIs.
	Trackers [][]string
	// TODO: Move into a "new" Torrent opt type.
	InfoHash  metainfo.Hash
	InfoBytes []byte
	// The name to use if the Name field from the Info isn't available.
	DisplayName string
	Webseeds    []string
	DhtNodes    []string
	PeerAddrs   []string
	// The combination of the "xs" and "as" fields in magnet links, for now.
	Sources []string

	// The chunk size to use for outbound requests. Defaults to 16KiB if not set. Can only be set
	// for new Torrents. TODO: Move into a "new" Torrent opt type.
	ChunkSize pp.Integer
	// TODO: Move into a "new" Torrent opt type.
	Storage storage.ClientImpl

	DisableInitialPieceCheck bool

	// Whether to allow data download or upload
	DisallowDataUpload   bool
	DisallowDataDownload bool
}

func TorrentSpecFromMagnetUri(uri string) (spec *TorrentSpec, err error) {
	m, err := metainfo.ParseMagnetUri(uri)
	if err != nil {
		return
	}
	spec = &TorrentSpec{
		Trackers:    [][]string{m.Trackers},
		DisplayName: m.DisplayName,
		InfoHash:    m.InfoHash,
		Webseeds:    m.Params["ws"],
		Sources:     append(m.Params["xs"], m.Params["as"]...),
		PeerAddrs:   m.Params["x.pe"], // BEP 9
		// TODO: What's the parameter for DHT nodes?
	}
	return
}

// The error will be from unmarshalling the info bytes. The TorrentSpec is still filled out as much
// as possible in this case.
func TorrentSpecFromMetaInfoErr(mi *metainfo.MetaInfo) (*TorrentSpec, error) {
	info, err := mi.UnmarshalInfo()
	if err != nil {
		err = fmt.Errorf("unmarshalling info: %w", err)
	}
	return &TorrentSpec{
		Trackers:    mi.UpvertedAnnounceList(),
		InfoHash:    mi.HashInfoBytes(),
		InfoBytes:   mi.InfoBytes,
		DisplayName: info.Name,
		Webseeds:    mi.UrlList,
		DhtNodes: func() (ret []string) {
			ret = make([]string, 0, len(mi.Nodes))
			for _, node := range mi.Nodes {
				ret = append(ret, string(node))
			}
			return
		}(),
	}, err
}

// Panics if there was anything missing from the metainfo.
func TorrentSpecFromMetaInfo(mi *metainfo.MetaInfo) *TorrentSpec {
	ts, err := TorrentSpecFromMetaInfoErr(mi)
	if err != nil {
		panic(err)
	}
	return ts
}
