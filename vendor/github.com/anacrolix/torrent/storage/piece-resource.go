package storage

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"sort"
	"strconv"
	"sync"

	"github.com/anacrolix/missinggo/v2/resource"

	"github.com/anacrolix/torrent/metainfo"
)

type piecePerResource struct {
	rp   PieceProvider
	opts ResourcePiecesOpts
}

type ResourcePiecesOpts struct {
	// After marking a piece complete, don't bother deleting its incomplete blobs.
	LeaveIncompleteChunks bool
	// Sized puts require being able to stream from a statement executed on another connection.
	// Without them, we buffer the entire read and then put that.
	NoSizedPuts bool
	Capacity    *int64
}

func NewResourcePieces(p PieceProvider) ClientImpl {
	return NewResourcePiecesOpts(p, ResourcePiecesOpts{})
}

func NewResourcePiecesOpts(p PieceProvider, opts ResourcePiecesOpts) ClientImpl {
	return &piecePerResource{
		rp:   p,
		opts: opts,
	}
}

type piecePerResourceTorrentImpl struct {
	piecePerResource
	locks []sync.RWMutex
}

func (piecePerResourceTorrentImpl) Close() error {
	return nil
}

func (s piecePerResource) OpenTorrent(info *metainfo.Info, infoHash metainfo.Hash) (TorrentImpl, error) {
	t := piecePerResourceTorrentImpl{
		s,
		make([]sync.RWMutex, info.NumPieces()),
	}
	return TorrentImpl{Piece: t.Piece, Close: t.Close}, nil
}

func (s piecePerResourceTorrentImpl) Piece(p metainfo.Piece) PieceImpl {
	return piecePerResourcePiece{
		mp:               p,
		piecePerResource: s.piecePerResource,
		mu:               &s.locks[p.Index()],
	}
}

type PieceProvider interface {
	resource.Provider
}

type ConsecutiveChunkReader interface {
	ReadConsecutiveChunks(prefix string) (io.ReadCloser, error)
}

type piecePerResourcePiece struct {
	mp metainfo.Piece
	piecePerResource
	// This protects operations that move complete/incomplete pieces around, which can trigger read
	// errors that may cause callers to do more drastic things.
	mu *sync.RWMutex
}

var _ io.WriterTo = piecePerResourcePiece{}

func (s piecePerResourcePiece) WriteTo(w io.Writer) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.mustIsComplete() {
		r, err := s.completed().Get()
		if err != nil {
			return 0, fmt.Errorf("getting complete instance: %w", err)
		}
		defer r.Close()
		return io.Copy(w, r)
	}
	if ccr, ok := s.rp.(ConsecutiveChunkReader); ok {
		return s.writeConsecutiveIncompleteChunks(ccr, w)
	}
	return io.Copy(w, io.NewSectionReader(s, 0, s.mp.Length()))
}

func (s piecePerResourcePiece) writeConsecutiveIncompleteChunks(ccw ConsecutiveChunkReader, w io.Writer) (int64, error) {
	r, err := ccw.ReadConsecutiveChunks(s.incompleteDirPath() + "/")
	if err != nil {
		return 0, err
	}
	defer r.Close()
	return io.Copy(w, r)
}

// Returns if the piece is complete. Ok should be true, because we are the definitive source of
// truth here.
func (s piecePerResourcePiece) mustIsComplete() bool {
	completion := s.Completion()
	if !completion.Ok {
		panic("must know complete definitively")
	}
	return completion.Complete
}

func (s piecePerResourcePiece) Completion() Completion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fi, err := s.completed().Stat()
	return Completion{
		Complete: err == nil && fi.Size() == s.mp.Length(),
		Ok:       true,
	}
}

type SizedPutter interface {
	PutSized(io.Reader, int64) error
}

func (s piecePerResourcePiece) MarkComplete() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	incompleteChunks := s.getChunks()
	r, err := func() (io.ReadCloser, error) {
		if ccr, ok := s.rp.(ConsecutiveChunkReader); ok {
			return ccr.ReadConsecutiveChunks(s.incompleteDirPath() + "/")
		}
		return ioutil.NopCloser(io.NewSectionReader(incompleteChunks, 0, s.mp.Length())), nil
	}()
	if err != nil {
		return fmt.Errorf("getting incomplete chunks reader: %w", err)
	}
	defer r.Close()
	completedInstance := s.completed()
	err = func() error {
		if sp, ok := completedInstance.(SizedPutter); ok && !s.opts.NoSizedPuts {
			return sp.PutSized(r, s.mp.Length())
		} else {
			return completedInstance.Put(r)
		}
	}()
	if err == nil && !s.opts.LeaveIncompleteChunks {
		// I think we do this synchronously here since we don't want callers to act on the completed
		// piece if we're concurrently still deleting chunks. The caller may decide to start
		// downloading chunks again and won't expect us to delete them. It seems to be much faster
		// to let the resource provider do this if possible.
		var wg sync.WaitGroup
		for _, c := range incompleteChunks {
			wg.Add(1)
			go func(c chunk) {
				defer wg.Done()
				c.instance.Delete()
			}(c)
		}
		wg.Wait()
	}
	return err
}

func (s piecePerResourcePiece) MarkNotComplete() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.completed().Delete()
}

func (s piecePerResourcePiece) ReadAt(b []byte, off int64) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.mustIsComplete() {
		return s.completed().ReadAt(b, off)
	}
	return s.getChunks().ReadAt(b, off)
}

func (s piecePerResourcePiece) WriteAt(b []byte, off int64) (n int, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i, err := s.rp.NewInstance(path.Join(s.incompleteDirPath(), strconv.FormatInt(off, 10)))
	if err != nil {
		panic(err)
	}
	r := bytes.NewReader(b)
	if sp, ok := i.(SizedPutter); ok {
		err = sp.PutSized(r, r.Size())
	} else {
		err = i.Put(r)
	}
	n = len(b) - r.Len()
	return
}

type chunk struct {
	offset   int64
	instance resource.Instance
}

type chunks []chunk

func (me chunks) ReadAt(b []byte, off int64) (int, error) {
	for {
		if len(me) == 0 {
			return 0, io.EOF
		}
		if me[0].offset <= off {
			break
		}
		me = me[1:]
	}
	n, err := me[0].instance.ReadAt(b, off-me[0].offset)
	if n == len(b) {
		return n, nil
	}
	if err == nil || err == io.EOF {
		n_, err := me[1:].ReadAt(b[n:], off+int64(n))
		return n + n_, err
	}
	return n, err
}

func (s piecePerResourcePiece) getChunks() (chunks chunks) {
	names, err := s.incompleteDir().Readdirnames()
	if err != nil {
		return
	}
	for _, n := range names {
		offset, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			panic(err)
		}
		i, err := s.rp.NewInstance(path.Join(s.incompleteDirPath(), n))
		if err != nil {
			panic(err)
		}
		chunks = append(chunks, chunk{offset, i})
	}
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].offset < chunks[j].offset
	})
	return
}

func (s piecePerResourcePiece) completedInstancePath() string {
	return path.Join("completed", s.mp.Hash().HexString())
}

func (s piecePerResourcePiece) completed() resource.Instance {
	i, err := s.rp.NewInstance(s.completedInstancePath())
	if err != nil {
		panic(err)
	}
	return i
}

func (s piecePerResourcePiece) incompleteDirPath() string {
	return path.Join("incompleted", s.mp.Hash().HexString())
}

func (s piecePerResourcePiece) incompleteDir() resource.DirInstance {
	i, err := s.rp.NewInstance(s.incompleteDirPath())
	if err != nil {
		panic(err)
	}
	return i.(resource.DirInstance)
}
