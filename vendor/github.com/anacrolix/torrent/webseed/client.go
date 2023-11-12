package webseed

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/RoaringBitmap/roaring"
	"github.com/anacrolix/torrent/common"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/segments"
)

type RequestSpec = segments.Extent

type requestPartResult struct {
	resp *http.Response
	err  error
}

type requestPart struct {
	req    *http.Request
	e      segments.Extent
	result chan requestPartResult
	start  func()
	// Wrap http response bodies for such things as download rate limiting.
	responseBodyWrapper ResponseBodyWrapper
}

type Request struct {
	cancel func()
	Result chan RequestResult
}

func (r Request) Cancel() {
	r.cancel()
}

type Client struct {
	HttpClient *http.Client
	Url        string
	fileIndex  segments.Index
	info       *metainfo.Info
	// The pieces we can request with the Url. We're more likely to ban/block at the file-level
	// given that's how requests are mapped to webseeds, but the torrent.Client works at the piece
	// level. We can map our file-level adjustments to the pieces here. This probably need to be
	// private in the future, if Client ever starts removing pieces.
	Pieces              roaring.Bitmap
	ResponseBodyWrapper ResponseBodyWrapper
}

type ResponseBodyWrapper func(io.Reader) io.Reader

func (me *Client) SetInfo(info *metainfo.Info) {
	if !strings.HasSuffix(me.Url, "/") && info.IsDir() {
		// In my experience, this is a non-conforming webseed. For example the
		// http://ia600500.us.archive.org/1/items URLs in archive.org torrents.
		return
	}
	me.fileIndex = segments.NewIndex(common.LengthIterFromUpvertedFiles(info.UpvertedFiles()))
	me.info = info
	me.Pieces.AddRange(0, uint64(info.NumPieces()))
}

type RequestResult struct {
	Bytes []byte
	Err   error
}

func (ws *Client) NewRequest(r RequestSpec) Request {
	ctx, cancel := context.WithCancel(context.Background())
	var requestParts []requestPart
	if !ws.fileIndex.Locate(r, func(i int, e segments.Extent) bool {
		req, err := NewRequest(ws.Url, i, ws.info, e.Start, e.Length)
		if err != nil {
			panic(err)
		}
		req = req.WithContext(ctx)
		part := requestPart{
			req:                 req,
			result:              make(chan requestPartResult, 1),
			e:                   e,
			responseBodyWrapper: ws.ResponseBodyWrapper,
		}
		part.start = func() {
			go func() {
				resp, err := ws.HttpClient.Do(req)
				part.result <- requestPartResult{
					resp: resp,
					err:  err,
				}
			}()
		}
		requestParts = append(requestParts, part)
		return true
	}) {
		panic("request out of file bounds")
	}
	req := Request{
		cancel: cancel,
		Result: make(chan RequestResult, 1),
	}
	go func() {
		b, err := readRequestPartResponses(ctx, requestParts)
		req.Result <- RequestResult{
			Bytes: b,
			Err:   err,
		}
	}()
	return req
}

type ErrBadResponse struct {
	Msg      string
	Response *http.Response
}

func (me ErrBadResponse) Error() string {
	return me.Msg
}

func recvPartResult(ctx context.Context, buf io.Writer, part requestPart) error {
	result := <-part.result
	// Make sure there's no further results coming, it should be a one-shot channel.
	close(part.result)
	if result.err != nil {
		return result.err
	}
	defer result.resp.Body.Close()
	var body io.Reader = result.resp.Body
	if part.responseBodyWrapper != nil {
		body = part.responseBodyWrapper(body)
	}
	// Prevent further accidental use
	result.resp.Body = nil
	if ctx.Err() != nil {
		return ctx.Err()
	}
	switch result.resp.StatusCode {
	case http.StatusPartialContent:
		copied, err := io.Copy(buf, body)
		if err != nil {
			return err
		}
		if copied != part.e.Length {
			return fmt.Errorf("got %v bytes, expected %v", copied, part.e.Length)
		}
		return nil
	case http.StatusOK:
		// This number is based on
		// https://archive.org/download/BloodyPitOfHorror/BloodyPitOfHorror.asr.srt. It seems that
		// archive.org might be using a webserver implementation that refuses to do partial
		// responses to small files.
		if part.e.Start < 48<<10 {
			if part.e.Start != 0 {
				log.Printf("resp status ok but requested range [url=%q, range=%q]",
					part.req.URL,
					part.req.Header.Get("Range"))
			}
			// Instead of discarding, we could try receiving all the chunks present in the response
			// body. I don't know how one would handle multiple chunk requests resulting in an OK
			// response for the same file. The request algorithm might be need to be smarter for
			// that.
			discarded, _ := io.CopyN(io.Discard, body, part.e.Start)
			if discarded != 0 {
				log.Printf("discarded %v bytes in webseed request response part", discarded)
			}
			_, err := io.CopyN(buf, body, part.e.Length)
			return err
		} else {
			return ErrBadResponse{"resp status ok but requested range", result.resp}
		}
	case http.StatusServiceUnavailable:
		return ErrTooFast
	default:
		return ErrBadResponse{
			fmt.Sprintf("unhandled response status code (%v)", result.resp.StatusCode),
			result.resp,
		}
	}
}

var ErrTooFast = errors.New("making requests too fast")

func readRequestPartResponses(ctx context.Context, parts []requestPart) (_ []byte, err error) {
	var buf bytes.Buffer
	for _, part := range parts {
		part.start()
		err = recvPartResult(ctx, &buf, part)
		if err != nil {
			err = fmt.Errorf("reading %q at %q: %w", part.req.URL, part.req.Header.Get("Range"), err)
			break
		}
	}
	return buf.Bytes(), err
}
