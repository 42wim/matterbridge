package dynamic

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/waku-org/go-waku/waku/v2/protocol/rln/contracts"
	"github.com/waku-org/go-waku/waku/v2/protocol/rln/group_manager"
	"github.com/waku-org/go-waku/waku/v2/protocol/rln/web3"
	"github.com/waku-org/go-zerokit-rln/rln"
	"go.uber.org/zap"
)

// RegistrationEventHandler represents the types of inputs to this handler matches the MemberRegistered event/proc defined in the MembershipContract interface
type RegistrationEventHandler = func([]*contracts.RLNMemberRegistered, uint64) error

// MembershipFetcher is used for getting membershipRegsitered Events from the eth rpc
type MembershipFetcher struct {
	web3Config  *web3.Config
	rln         *rln.RLN
	log         *zap.Logger
	rootTracker *group_manager.MerkleRootTracker
	wg          sync.WaitGroup
}

func NewMembershipFetcher(web3Config *web3.Config, rln *rln.RLN, rootTracker *group_manager.MerkleRootTracker, log *zap.Logger) MembershipFetcher {
	return MembershipFetcher{
		web3Config:  web3Config,
		rln:         rln,
		log:         log,
		rootTracker: rootTracker,
	}
}

// HandleGroupUpdates mounts the supplied handler for the registration events emitting from the membership contract
// It connects to the eth client, subscribes to the `MemberRegistered` event emitted from the `MembershipContract`
// and collects all the events, for every received event, it calls the `handler`
func (mf *MembershipFetcher) HandleGroupUpdates(ctx context.Context, handler RegistrationEventHandler) error {
	fromBlock := mf.web3Config.RLNContract.DeployedBlockNumber
	metadata, err := mf.GetMetadata()
	if err != nil {
		mf.log.Warn("could not load last processed block from metadata. Starting onchain sync from deployment block", zap.Error(err), zap.Uint64("deploymentBlock", mf.web3Config.RLNContract.DeployedBlockNumber))
	} else {
		if mf.web3Config.ChainID.Cmp(metadata.ChainID) != 0 {
			return errors.New("persisted data: chain id mismatch")
		}

		if !bytes.Equal(mf.web3Config.RegistryContract.Address.Bytes(), metadata.ContractAddress.Bytes()) {
			return errors.New("persisted data: contract address mismatch")
		}

		fromBlock = metadata.LastProcessedBlock + 1
		mf.log.Info("resuming onchain sync", zap.Uint64("fromBlock", fromBlock))
	}

	mf.rootTracker.SetValidRootsPerBlock(metadata.ValidRootsPerBlock)
	//
	latestBlockNumber, err := mf.latestBlockNumber(ctx)
	if err != nil {
		return err
	}
	//

	mf.log.Info("loading old events...")
	t := time.Now()
	err = mf.loadOldEvents(ctx, fromBlock, latestBlockNumber, handler)
	if err != nil {
		return err
	}
	mf.log.Info("events loaded", zap.Duration("timeToLoad", time.Since(t)))

	errCh := make(chan error)

	mf.wg.Add(1)
	go mf.watchNewEvents(ctx, latestBlockNumber+1, handler, errCh) // we have already fetched the events for latestBlocNumber in oldEvents
	return <-errCh
}

func (mf *MembershipFetcher) loadOldEvents(ctx context.Context, fromBlock, toBlock uint64, handler RegistrationEventHandler) error {
	for ; fromBlock+maxBatchSize < toBlock; fromBlock += maxBatchSize + 1 { // check if the end of the batch is within the toBlock range
		t1 := time.Now()
		events, err := mf.getEvents(ctx, fromBlock, fromBlock+maxBatchSize)
		if err != nil {
			return err
		}
		t1Since := time.Since(t1)

		t2 := time.Now()
		if err := handler(events, fromBlock+maxBatchSize); err != nil {
			return err
		}

		mf.log.Info("fetching events", zap.Uint64("from", fromBlock), zap.Uint64("to", fromBlock+maxBatchSize), zap.Int("numEvents", len(events)), zap.Duration("timeToFetch", t1Since), zap.Duration("timeToProcess", time.Since(t2)))
	}

	t1 := time.Now()
	events, err := mf.getEvents(ctx, fromBlock, toBlock)
	if err != nil {
		return err
	}
	t1Since := time.Since(t1)

	// process all the fetched events
	t2 := time.Now()
	err = handler(events, toBlock)
	if err != nil {
		return err
	}

	mf.log.Info("fetching events", zap.Uint64("from", fromBlock), zap.Uint64("to", fromBlock+maxBatchSize), zap.Int("numEvents", len(events)), zap.Duration("timeToFetch", t1Since), zap.Duration("timeToProcess", time.Since(t2)))
	return nil
}

func (mf *MembershipFetcher) watchNewEvents(ctx context.Context, fromBlock uint64, handler RegistrationEventHandler, errCh chan<- error) {
	defer mf.wg.Done()

	// Watch for new events
	firstErr := true
	headerCh := make(chan *types.Header)
	subs := event.Resubscribe(2*time.Second, func(ctx context.Context) (event.Subscription, error) {
		s, err := mf.web3Config.ETHClient.SubscribeNewHead(ctx, headerCh)
		if err != nil {
			if err == rpc.ErrNotificationsUnsupported {
				err = errors.New("notifications not supported. The node must support websockets")
			}
			mf.log.Error("subscribing to rln events", zap.Error(err))
		}
		if firstErr { // errCh can be closed only once
			errCh <- err
			close(errCh)
			firstErr = false
		}
		return s, err
	})

	defer subs.Unsubscribe()
	defer close(headerCh)

	for {
		select {
		case h := <-headerCh:
			toBlock := h.Number.Uint64()
			events, err := mf.getEvents(ctx, fromBlock, toBlock)
			if err != nil {
				mf.log.Error("obtaining rln events", zap.Error(err))
			} else {
				// update the last processed block
				fromBlock = toBlock + 1
			}

			err = handler(events, toBlock)
			if err != nil {
				mf.log.Error("processing rln log", zap.Error(err))
			}
		case <-ctx.Done():
			return
		case err := <-subs.Err():
			if err != nil {
				mf.log.Error("watching new events", zap.Error(err))
			}
			return
		}
	}
}

const maxBatchSize = uint64(5000)

func tooMuchDataRequestedError(err error) bool {
	// this error is only infura specific (other providers might have different error messages)
	return err.Error() == "query returned more than 10000 results"
}

func (mf *MembershipFetcher) latestBlockNumber(ctx context.Context) (uint64, error) {
	block, err := mf.web3Config.ETHClient.BlockByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}

	return block.Number().Uint64(), nil
}

func (mf *MembershipFetcher) getEvents(ctx context.Context, fromBlock uint64, toBlock uint64) ([]*contracts.RLNMemberRegistered, error) {
	evts, err := mf.fetchEvents(ctx, fromBlock, toBlock)
	if err != nil {
		if tooMuchDataRequestedError(err) { // divide the range and try again
			mid := (fromBlock + toBlock) / 2
			firstHalfEvents, err := mf.getEvents(ctx, fromBlock, mid)
			if err != nil {
				return nil, err
			}
			secondHalfEvents, err := mf.getEvents(ctx, mid+1, toBlock)
			if err != nil {
				return nil, err
			}
			return append(firstHalfEvents, secondHalfEvents...), nil
		}
		return nil, err
	}
	return evts, nil
}

func (mf *MembershipFetcher) fetchEvents(ctx context.Context, from uint64, to uint64) ([]*contracts.RLNMemberRegistered, error) {
	return retry.DoWithData(
		func() ([]*contracts.RLNMemberRegistered, error) {
			logIterator, err := mf.web3Config.RLNContract.FilterMemberRegistered(&bind.FilterOpts{Start: from, End: &to, Context: ctx})
			if err != nil {
				return nil, err
			}

			var results []*contracts.RLNMemberRegistered

			for {
				if !logIterator.Next() {
					break
				}

				if logIterator.Error() != nil {
					return nil, logIterator.Error()
				}

				results = append(results, logIterator.Event)
			}

			return results, nil
		}, retry.Attempts(3),
	)
}

// GetMetadata retrieves metadata from the zerokit's RLN database
func (mf *MembershipFetcher) GetMetadata() (RLNMetadata, error) {
	b, err := mf.rln.GetMetadata()
	if err != nil {
		return RLNMetadata{}, err
	}

	return DeserializeMetadata(b)
}

func (mf *MembershipFetcher) Stop() {
	mf.web3Config.ETHClient.Close()
	// wait for the watchNewEvents goroutine to finish
	mf.wg.Wait()
}
