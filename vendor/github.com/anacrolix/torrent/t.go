package torrent

import (
	"strconv"
	"strings"

	"github.com/anacrolix/chansync/events"
	"github.com/anacrolix/missinggo/pubsub"
	"github.com/anacrolix/sync"

	"github.com/anacrolix/torrent/metainfo"
)

// The Torrent's infohash. This is fixed and cannot change. It uniquely identifies a torrent.
func (t *Torrent) InfoHash() metainfo.Hash {
	return t.infoHash
}

// Returns a channel that is closed when the info (.Info()) for the torrent has become available.
func (t *Torrent) GotInfo() events.Done {
	return t.gotMetainfoC
}

// Returns the metainfo info dictionary, or nil if it's not yet available.
func (t *Torrent) Info() (info *metainfo.Info) {
	t.nameMu.RLock()
	info = t.info
	t.nameMu.RUnlock()
	return
}

// Returns a Reader bound to the torrent's data. All read calls block until the data requested is
// actually available. Note that you probably want to ensure the Torrent Info is available first.
func (t *Torrent) NewReader() Reader {
	return t.newReader(0, *t.length)
}

func (t *Torrent) newReader(offset, length int64) Reader {
	r := reader{
		mu:     t.cl.locker(),
		t:      t,
		offset: offset,
		length: length,
	}
	r.readaheadFunc = defaultReadaheadFunc
	t.addReader(&r)
	return &r
}

type PieceStateRuns []PieceStateRun

func (me PieceStateRuns) String() (s string) {
	if len(me) > 0 {
		var sb strings.Builder
		sb.WriteString(me[0].String())
		for i := 1; i < len(me); i += 1 {
			sb.WriteByte(' ')
			sb.WriteString(me[i].String())
		}
		return sb.String()
	}
	return
}

// Returns the state of pieces of the torrent. They are grouped into runs of same state. The sum of
// the state run-lengths is the number of pieces in the torrent.
func (t *Torrent) PieceStateRuns() (runs PieceStateRuns) {
	t.cl.rLock()
	runs = t.pieceStateRuns()
	t.cl.rUnlock()
	return
}

func (t *Torrent) PieceState(piece pieceIndex) (ps PieceState) {
	t.cl.rLock()
	ps = t.pieceState(piece)
	t.cl.rUnlock()
	return
}

// The number of pieces in the torrent. This requires that the info has been
// obtained first.
func (t *Torrent) NumPieces() pieceIndex {
	return t.numPieces()
}

// Get missing bytes count for specific piece.
func (t *Torrent) PieceBytesMissing(piece int) int64 {
	t.cl.lock()
	defer t.cl.unlock()

	return int64(t.pieces[piece].bytesLeft())
}

// Drop the torrent from the client, and close it. It's always safe to do
// this. No data corruption can, or should occur to either the torrent's data,
// or connected peers.
func (t *Torrent) Drop() {
	var wg sync.WaitGroup
	defer wg.Wait()
	t.cl.lock()
	defer t.cl.unlock()
	t.cl.dropTorrent(t.infoHash, &wg)
}

// Number of bytes of the entire torrent we have completed. This is the sum of
// completed pieces, and dirtied chunks of incomplete pieces. Do not use this
// for download rate, as it can go down when pieces are lost or fail checks.
// Sample Torrent.Stats.DataBytesRead for actual file data download rate.
func (t *Torrent) BytesCompleted() int64 {
	t.cl.rLock()
	defer t.cl.rUnlock()
	return t.bytesCompleted()
}

// The subscription emits as (int) the index of pieces as their state changes.
// A state change is when the PieceState for a piece alters in value.
func (t *Torrent) SubscribePieceStateChanges() *pubsub.Subscription {
	return t.pieceStateChanges.Subscribe()
}

// Returns true if the torrent is currently being seeded. This occurs when the
// client is willing to upload without wanting anything in return.
func (t *Torrent) Seeding() (ret bool) {
	t.cl.lock()
	ret = t.seeding()
	t.cl.unlock()
	return
}

// Clobbers the torrent display name if metainfo is unavailable.
// The display name is used as the torrent name while the metainfo is unavailable.
func (t *Torrent) SetDisplayName(dn string) {
	t.nameMu.Lock()
	if !t.haveInfo() {
		t.displayName = dn
	}
	t.nameMu.Unlock()
}

// The current working name for the torrent. Either the name in the info dict,
// or a display name given such as by the dn value in a magnet link, or "".
func (t *Torrent) Name() string {
	return t.name()
}

// The completed length of all the torrent data, in all its files. This is
// derived from the torrent info, when it is available.
func (t *Torrent) Length() int64 {
	return *t.length
}

// Returns a run-time generated metainfo for the torrent that includes the
// info bytes and announce-list as currently known to the client.
func (t *Torrent) Metainfo() metainfo.MetaInfo {
	t.cl.lock()
	defer t.cl.unlock()
	return t.newMetaInfo()
}

func (t *Torrent) addReader(r *reader) {
	t.cl.lock()
	defer t.cl.unlock()
	if t.readers == nil {
		t.readers = make(map[*reader]struct{})
	}
	t.readers[r] = struct{}{}
	r.posChanged()
}

func (t *Torrent) deleteReader(r *reader) {
	delete(t.readers, r)
	t.readersChanged()
}

// Raise the priorities of pieces in the range [begin, end) to at least Normal
// priority. Piece indexes are not the same as bytes. Requires that the info
// has been obtained, see Torrent.Info and Torrent.GotInfo.
func (t *Torrent) DownloadPieces(begin, end pieceIndex) {
	t.cl.lock()
	t.downloadPiecesLocked(begin, end)
	t.cl.unlock()
}

func (t *Torrent) downloadPiecesLocked(begin, end pieceIndex) {
	for i := begin; i < end; i++ {
		if t.pieces[i].priority.Raise(PiecePriorityNormal) {
			t.updatePiecePriority(i, "Torrent.DownloadPieces")
		}
	}
}

func (t *Torrent) CancelPieces(begin, end pieceIndex) {
	t.cl.lock()
	t.cancelPiecesLocked(begin, end, "Torrent.CancelPieces")
	t.cl.unlock()
}

func (t *Torrent) cancelPiecesLocked(begin, end pieceIndex, reason string) {
	for i := begin; i < end; i++ {
		p := &t.pieces[i]
		if p.priority == PiecePriorityNone {
			continue
		}
		p.priority = PiecePriorityNone
		t.updatePiecePriority(i, reason)
	}
}

func (t *Torrent) initFiles() {
	var offset int64
	t.files = new([]*File)
	for _, fi := range t.info.UpvertedFiles() {
		var path []string
		if len(fi.PathUTF8) != 0 {
			path = fi.PathUTF8
		} else {
			path = fi.Path
		}
		dp := t.info.Name
		if len(fi.Path) != 0 {
			dp = strings.Join(fi.Path, "/")
		}
		*t.files = append(*t.files, &File{
			t,
			strings.Join(append([]string{t.info.Name}, path...), "/"),
			offset,
			fi.Length,
			fi,
			dp,
			PiecePriorityNone,
		})
		offset += fi.Length
	}
}

// Returns handles to the files in the torrent. This requires that the Info is
// available first.
func (t *Torrent) Files() []*File {
	return *t.files
}

func (t *Torrent) AddPeers(pp []PeerInfo) (n int) {
	t.cl.lock()
	n = t.addPeers(pp)
	t.cl.unlock()
	return
}

// Marks the entire torrent for download. Requires the info first, see
// GotInfo. Sets piece priorities for historical reasons.
func (t *Torrent) DownloadAll() {
	t.DownloadPieces(0, t.numPieces())
}

func (t *Torrent) String() string {
	s := t.name()
	if s == "" {
		return t.infoHash.HexString()
	} else {
		return strconv.Quote(s)
	}
}

func (t *Torrent) AddTrackers(announceList [][]string) {
	t.cl.lock()
	defer t.cl.unlock()
	t.addTrackers(announceList)
}

func (t *Torrent) Piece(i pieceIndex) *Piece {
	return t.piece(i)
}

func (t *Torrent) PeerConns() []*PeerConn {
	t.cl.rLock()
	defer t.cl.rUnlock()
	ret := make([]*PeerConn, 0, len(t.conns))
	for c := range t.conns {
		ret = append(ret, c)
	}
	return ret
}
