package transfer

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/log"
)

// SetupIterativeDownloader configures IterativeDownloader with last known synced block.
func SetupIterativeDownloader(
	client HeaderReader, downloader BatchDownloader, size *big.Int, to *big.Int, from *big.Int) (*IterativeDownloader, error) {

	if to == nil || from == nil {
		return nil, errors.New("to or from cannot be nil")
	}

	log.Debug("iterative downloader", "from", from, "to", to, "size", size)
	d := &IterativeDownloader{
		client:     client,
		batchSize:  size,
		downloader: downloader,
		from:       from,
		to:         to,
	}
	return d, nil
}

// BatchDownloader interface for loading transfers in batches in speificed range of blocks.
type BatchDownloader interface {
	GetHeadersInRange(ctx context.Context, from, to *big.Int) ([]*DBHeader, error)
}

// IterativeDownloader downloads batches of transfers in a specified size.
type IterativeDownloader struct {
	client HeaderReader

	batchSize *big.Int

	downloader BatchDownloader

	from, to *big.Int
	previous *big.Int
}

// Finished true when earliest block with given sync option is zero.
func (d *IterativeDownloader) Finished() bool {
	return d.from.Cmp(d.to) == 0
}

// Header return last synced header.
func (d *IterativeDownloader) Header() *big.Int {
	return d.previous
}

// Next moves closer to the end on every new iteration.
func (d *IterativeDownloader) Next(parent context.Context) ([]*DBHeader, *big.Int, *big.Int, error) {
	to := d.to
	from := new(big.Int).Sub(to, d.batchSize)
	// if start < 0; start = 0
	if from.Cmp(d.from) == -1 {
		from = d.from
	}
	headers, err := d.downloader.GetHeadersInRange(parent, from, to)
	log.Debug("load erc20 transfers in range", "from", from, "to", to, "batchSize", d.batchSize)
	if err != nil {
		log.Error("failed to get transfer in between two blocks", "from", from, "to", to, "error", err)
		return nil, nil, nil, err
	}

	d.previous, d.to = d.to, from
	return headers, d.from, to, nil
}

// Revert reverts last step progress. Should be used if application failed to process transfers.
// For example failed to persist them.
func (d *IterativeDownloader) Revert() {
	if d.previous != nil {
		d.from = d.previous
	}
}
