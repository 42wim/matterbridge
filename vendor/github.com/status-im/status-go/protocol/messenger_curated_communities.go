package protocol

import (
	"context"
	"errors"
	"reflect"
	"time"

	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/communities"
)

const (
	curatedCommunitiesUpdateInterval = time.Hour
	communitiesUpdateFailureInterval = time.Minute
)

// Regularly gets list of curated communities and signals them to client
func (m *Messenger) startCuratedCommunitiesUpdateLoop() {
	logger := m.logger.Named("curatedCommunitiesUpdateLoop")

	if m.contractMaker == nil {
		logger.Warn("not starting curated communities loop: contract maker not initialized")
		return
	}

	go func() {
		// Initialize interval to 0 for immediate execution
		var interval time.Duration = 0

		cache, err := m.communitiesManager.GetCuratedCommunities()
		if err != nil {
			logger.Error("failed to start curated communities loop", zap.Error(err))
			return
		}

		for {
			select {
			case <-time.After(interval):
				// Immediate execution on first run, then set to regular interval
				interval = curatedCommunitiesUpdateInterval

				curatedCommunities, err := m.getCuratedCommunitiesFromContract()
				if err != nil {
					interval = communitiesUpdateFailureInterval
					logger.Error("failed to get curated communities from contract", zap.Error(err))
					continue
				}

				if reflect.DeepEqual(cache.ContractCommunities, curatedCommunities.ContractCommunities) &&
					reflect.DeepEqual(cache.ContractFeaturedCommunities, curatedCommunities.ContractFeaturedCommunities) {
					// nothing changed
					continue
				}

				err = m.communitiesManager.SetCuratedCommunities(curatedCommunities)
				if err == nil {
					cache = curatedCommunities
				} else {
					logger.Error("failed to save curated communities", zap.Error(err))
				}

				response, err := m.fetchCuratedCommunities(curatedCommunities)
				if err != nil {
					interval = communitiesUpdateFailureInterval
					logger.Error("failed to fetch curated communities", zap.Error(err))
					continue
				}

				m.config.messengerSignalsHandler.SendCuratedCommunitiesUpdate(response)

			case <-m.quit:
				return
			}
		}
	}()
}

func (m *Messenger) getCuratedCommunitiesFromContract() (*communities.CuratedCommunities, error) {
	if m.contractMaker == nil {
		return nil, errors.New("contract maker not initialized")
	}

	testNetworksEnabled, err := m.settings.GetTestNetworksEnabled()
	if err != nil {
		return nil, err
	}

	chainID := uint64(10) // Optimism Mainnet
	if testNetworksEnabled {
		chainID = 420 // Optimism Goerli
	}

	directory, err := m.contractMaker.NewDirectory(chainID)
	if err != nil {
		return nil, err
	}

	callOpts := &bind.CallOpts{Context: context.Background(), Pending: false}

	contractCommunities, err := directory.GetCommunities(callOpts)
	if err != nil {
		return nil, err
	}
	var contractCommunityIDs []string
	for _, c := range contractCommunities {
		contractCommunityIDs = append(contractCommunityIDs, types.HexBytes(c).String())
	}

	featuredContractCommunities, err := directory.GetFeaturedCommunities(callOpts)
	if err != nil {
		return nil, err
	}
	var contractFeaturedCommunityIDs []string
	for _, c := range featuredContractCommunities {
		contractFeaturedCommunityIDs = append(contractFeaturedCommunityIDs, types.HexBytes(c).String())
	}

	return &communities.CuratedCommunities{
		ContractCommunities:         contractCommunityIDs,
		ContractFeaturedCommunities: contractFeaturedCommunityIDs,
	}, nil
}

func (m *Messenger) fetchCuratedCommunities(curatedCommunities *communities.CuratedCommunities) (*communities.KnownCommunitiesResponse, error) {
	response, err := m.communitiesManager.GetStoredDescriptionForCommunities(curatedCommunities.ContractCommunities)
	if err != nil {
		return nil, err
	}
	response.ContractFeaturedCommunities = curatedCommunities.ContractFeaturedCommunities

	// TODO: use mechanism to obtain shard from community ID (https://github.com/status-im/status-desktop/issues/12585)
	var unknownCommunities []communities.CommunityShard
	for _, u := range response.UnknownCommunities {
		unknownCommunities = append(unknownCommunities, communities.CommunityShard{
			CommunityID: u,
		})
	}

	go func() {
		_ = m.fetchCommunities(unknownCommunities)
	}()

	return response, nil
}

func (m *Messenger) CuratedCommunities() (*communities.KnownCommunitiesResponse, error) {
	curatedCommunities, err := m.communitiesManager.GetCuratedCommunities()
	if err != nil {
		return nil, err
	}
	return m.fetchCuratedCommunities(curatedCommunities)
}
