package ipfs

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/wealdtech/go-multicodec"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/params"
)

const maxRequestsPerSecond = 3

type taskResponse struct {
	err      error
	response []byte
}

type taskRequest struct {
	cid      string
	download bool
	doneChan chan taskResponse
}

type Downloader struct {
	ctx             context.Context
	cancel          func()
	ipfsDir         string
	wg              sync.WaitGroup
	rateLimiterChan chan taskRequest
	inputTaskChan   chan taskRequest
	client          *http.Client

	quit chan struct{}
}

func NewDownloader(rootDir string) *Downloader {
	ipfsDir := filepath.Clean(filepath.Join(rootDir, "./ipfs"))
	if err := os.MkdirAll(ipfsDir, 0700); err != nil {
		panic("could not create IPFSDir")
	}

	ctx, cancel := context.WithCancel(context.TODO())

	d := &Downloader{
		ctx:             ctx,
		cancel:          cancel,
		ipfsDir:         ipfsDir,
		rateLimiterChan: make(chan taskRequest, maxRequestsPerSecond),
		inputTaskChan:   make(chan taskRequest, 1000),
		wg:              sync.WaitGroup{},
		client: &http.Client{
			Timeout: time.Second * 5,
		},

		quit: make(chan struct{}, 1),
	}

	go d.taskDispatcher()
	go d.worker()

	return d
}

func (d *Downloader) Stop() {
	close(d.quit)

	d.cancel()

	d.wg.Wait()

	close(d.inputTaskChan)
	close(d.rateLimiterChan)
}

func (d *Downloader) worker() {
	for request := range d.rateLimiterChan {
		resp, err := d.download(request.cid, request.download)
		request.doneChan <- taskResponse{
			err:      err,
			response: resp,
		}
	}
}

func (d *Downloader) taskDispatcher() {
	ticker := time.NewTicker(time.Second / maxRequestsPerSecond)
	defer ticker.Stop()

	for {
		<-ticker.C
		request, ok := <-d.inputTaskChan
		if !ok {
			return
		}
		d.rateLimiterChan <- request

	}
}

func hashToCid(hash []byte) (string, error) {
	// contract response includes a contenthash, which needs to be decoded to reveal
	// an IPFS identifier. Once decoded, download the content from IPFS. This content
	// is in EDN format, ie https://ipfs.infura.io/ipfs/QmWVVLwVKCwkVNjYJrRzQWREVvEk917PhbHYAUhA1gECTM
	// and it also needs to be decoded in to a nim type

	data, codec, err := multicodec.RemoveCodec(hash)
	if err != nil {
		return "", err
	}

	codecName, err := multicodec.Name(codec)
	if err != nil {
		return "", err
	}

	if codecName != "ipfs-ns" {
		return "", errors.New("codecName is not ipfs-ns")
	}

	thisCID, err := cid.Parse(data)
	if err != nil {
		return "", err
	}

	return thisCID.Hash().B58String(), nil
}

func decodeStringHash(input string) (string, error) {
	hash, err := hexutil.Decode("0x" + input)
	if err != nil {
		return "", err
	}

	cid, err := hashToCid(hash)
	if err != nil {
		return "", err
	}

	return cid, nil
}

// Get checks if an IPFS image exists and returns it from cache
// otherwise downloads it from INFURA's ipfs gateway
func (d *Downloader) Get(hash string, download bool) ([]byte, error) {
	cid, err := decodeStringHash(hash)
	if err != nil {
		return nil, err
	}

	exists, content, err := d.exists(cid)
	if err != nil {
		return nil, err
	}
	if exists {
		return content, nil
	}

	doneChan := make(chan taskResponse, 1)

	d.wg.Add(1)

	d.inputTaskChan <- taskRequest{
		cid:      cid,
		download: download,
		doneChan: doneChan,
	}

	done := <-doneChan
	close(doneChan)

	d.wg.Done()

	return done.response, done.err
}

func (d *Downloader) exists(cid string) (bool, []byte, error) {
	path := filepath.Join(d.ipfsDir, cid)
	_, err := os.Stat(path)
	if err == nil {
		fileContent, err := os.ReadFile(path)
		return true, fileContent, err
	}

	return false, nil, nil
}

func (d *Downloader) download(cid string, download bool) ([]byte, error) {
	path := filepath.Join(d.ipfsDir, cid)

	req, err := http.NewRequest(http.MethodGet, params.IpfsGatewayURL+cid, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(d.ctx)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error("failed to close the stickerpack request body", "err", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Error("could not load data for", "cid", cid, "code", resp.StatusCode)
		return nil, errors.New("could not load ipfs data")
	}

	fileContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if download {
		// #nosec G306
		err = os.WriteFile(path, fileContent, 0700)
		if err != nil {
			return nil, err
		}
	}

	return fileContent, nil
}
