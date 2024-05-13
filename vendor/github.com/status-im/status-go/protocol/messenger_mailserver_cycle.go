package protocol

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/protocol/storenodes"
	"github.com/status-im/status-go/services/mailservers"
	"github.com/status-im/status-go/signal"
)

const defaultBackoff = 10 * time.Second
const graylistBackoff = 3 * time.Minute
const isAndroidEmulator = runtime.GOOS == "android" && runtime.GOARCH == "amd64"
const findNearestMailServer = !isAndroidEmulator

func (m *Messenger) mailserversByFleet(fleet string) []mailservers.Mailserver {
	return mailservers.DefaultMailserversByFleet(fleet)
}

type byRTTMsAndCanConnectBefore []SortedMailserver

func (s byRTTMsAndCanConnectBefore) Len() int {
	return len(s)
}

func (s byRTTMsAndCanConnectBefore) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byRTTMsAndCanConnectBefore) Less(i, j int) bool {
	// Slightly inaccurate as time sensitive sorting, but it does not matter so much
	now := time.Now()
	if s[i].CanConnectAfter.Before(now) && s[j].CanConnectAfter.Before(now) {
		return s[i].RTTMs < s[j].RTTMs
	}
	return s[i].CanConnectAfter.Before(s[j].CanConnectAfter)
}

func (m *Messenger) StartMailserverCycle(mailservers []mailservers.Mailserver) error {
	m.mailserverCycle.allMailservers = mailservers

	version := m.transport.WakuVersion()

	switch version {
	case 1:
		if m.server == nil {
			m.logger.Warn("not starting mailserver cycle: no p2p server is set")
			return nil
		}

		m.mailserverCycle.events = make(chan *p2p.PeerEvent, 20)
		m.mailserverCycle.subscription = m.server.SubscribeEvents(m.mailserverCycle.events)
		go m.updateWakuV1PeerStatus()

	case 2:
		if len(mailservers) == 0 {
			m.logger.Warn("not starting mailserver cycle: empty mailservers list")
			return nil
		}
		for _, storenode := range mailservers {
			_, err := m.transport.AddStorePeer(storenode.Address)
			if err != nil {
				return err
			}
		}
		go m.verifyStorenodeStatus()

	default:
		return fmt.Errorf("unsupported waku version: %d", version)
	}

	m.logger.Debug("starting mailserver cycle",
		zap.Uint("WakuVersion", m.transport.WakuVersion()),
		zap.Any("mailservers", mailservers),
	)

	return nil
}

func (m *Messenger) DisconnectActiveMailserver() {
	m.mailserverCycle.Lock()
	defer m.mailserverCycle.Unlock()
	m.disconnectActiveMailserver()
}

func (m *Messenger) disconnectMailserver() error {
	if m.mailserverCycle.activeMailserver == nil {
		m.logger.Info("no active mailserver")
		return nil
	}
	m.logger.Info("disconnecting active mailserver", zap.String("nodeID", m.mailserverCycle.activeMailserver.ID))
	m.mailPeersMutex.Lock()
	pInfo, ok := m.mailserverCycle.peers[m.mailserverCycle.activeMailserver.ID]
	if ok {
		pInfo.status = disconnected
		pInfo.canConnectAfter = time.Now().Add(graylistBackoff)
		m.mailserverCycle.peers[m.mailserverCycle.activeMailserver.ID] = pInfo
	} else {
		m.mailserverCycle.peers[m.mailserverCycle.activeMailserver.ID] = peerStatus{
			status:          disconnected,
			mailserver:      *m.mailserverCycle.activeMailserver,
			canConnectAfter: time.Now().Add(graylistBackoff),
		}
	}
	m.mailPeersMutex.Unlock()

	// WakuV2 does not keep an active storenode connection

	if m.mailserverCycle.activeMailserver.Version == 1 {
		node, err := m.mailserverCycle.activeMailserver.Enode()
		if err != nil {
			return err
		}
		m.server.RemovePeer(node)
	}

	m.mailserverCycle.activeMailserver = nil
	return nil
}

func (m *Messenger) disconnectActiveMailserver() {
	err := m.disconnectMailserver()
	if err != nil {
		m.logger.Error("failed to disconnect mailserver", zap.Error(err))
	}
	signal.SendMailserverChanged("", "")
}

func (m *Messenger) cycleMailservers() {
	m.logger.Info("Automatically switching mailserver")

	if m.mailserverCycle.activeMailserver != nil {
		m.disconnectActiveMailserver()
	}

	err := m.findNewMailserver()
	if err != nil {
		m.logger.Error("Error getting new mailserver", zap.Error(err))
	}
}

func poolSize(fleetSize int) int {
	return int(math.Ceil(float64(fleetSize) / 4))
}

func (m *Messenger) getFleet() (string, error) {
	var fleet string
	dbFleet, err := m.settings.GetFleet()
	if err != nil {
		return "", err
	}
	if dbFleet != "" {
		fleet = dbFleet
	} else if m.config.clusterConfig.Fleet != "" {
		fleet = m.config.clusterConfig.Fleet
	} else {
		fleet = params.FleetProd
	}
	return fleet, nil
}

func (m *Messenger) allMailservers() ([]mailservers.Mailserver, error) {
	// Get configured fleet
	fleet, err := m.getFleet()
	if err != nil {
		return nil, err
	}

	// Get default mailservers for given fleet
	allMailservers := m.mailserversByFleet(fleet)

	// Add custom configured mailservers
	if m.mailserversDatabase != nil {
		customMailservers, err := m.mailserversDatabase.Mailservers()
		if err != nil {
			return nil, err
		}

		for _, c := range customMailservers {
			if c.Fleet == fleet {
				c.Version = m.transport.WakuVersion()
				allMailservers = append(allMailservers, c)
			}
		}
	}

	// Filter mailservers by configured waku version
	wakuVersion := m.transport.WakuVersion()
	matchingMailservers := make([]mailservers.Mailserver, 0, len(allMailservers))

	for _, ms := range allMailservers {
		if ms.Version == wakuVersion {
			matchingMailservers = append(matchingMailservers, ms)
		}
	}

	return matchingMailservers, nil
}

type SortedMailserver struct {
	Address         string
	RTTMs           int
	CanConnectAfter time.Time
}

func (m *Messenger) findNewMailserver() error {
	pinnedMailserver, err := m.getPinnedMailserver()
	if err != nil {
		m.logger.Error("Could not obtain the pinned mailserver", zap.Error(err))
		return err
	}
	if pinnedMailserver != nil {
		return m.connectToMailserver(*pinnedMailserver)
	}

	allMailservers := m.mailserverCycle.allMailservers

	//	TODO: remove this check once sockets are stable on x86_64 emulators
	if findNearestMailServer {
		m.logger.Info("Finding a new mailserver...")

		var mailserverStr []string
		for _, m := range allMailservers {
			mailserverStr = append(mailserverStr, m.Address)
		}

		if len(allMailservers) == 0 {
			m.logger.Warn("no mailservers available") // Do nothing...
			return nil

		}

		var parseFn func(string) (string, error)
		if allMailservers[0].Version == 2 {
			parseFn = mailservers.MultiAddressToAddress
		} else {
			parseFn = mailservers.EnodeStringToAddr
		}

		pingResult, err := mailservers.DoPing(context.Background(), mailserverStr, 500, parseFn)
		if err != nil {
			// pinging mailservers might fail, but we don't care
			m.logger.Warn("mailservers.DoPing failed with", zap.Error(err))
		}

		var availableMailservers []*mailservers.PingResult
		for _, result := range pingResult {
			if result.Err != nil {
				m.logger.Info("connecting error", zap.String("err", *result.Err))
				continue // The results with error are ignored
			}
			availableMailservers = append(availableMailservers, result)
		}

		if len(availableMailservers) == 0 {
			m.logger.Warn("No mailservers available") // Do nothing...
			return nil
		}

		mailserversByAddress := make(map[string]mailservers.Mailserver)
		for idx := range allMailservers {
			mailserversByAddress[allMailservers[idx].Address] = allMailservers[idx]
		}
		var sortedMailservers []SortedMailserver
		for _, ping := range availableMailservers {
			address := ping.Address
			ms := mailserversByAddress[address]
			sortedMailserver := SortedMailserver{
				Address: address,
				RTTMs:   *ping.RTTMs,
			}
			m.mailPeersMutex.Lock()
			pInfo, ok := m.mailserverCycle.peers[ms.ID]
			m.mailPeersMutex.Unlock()
			if ok {
				if time.Now().Before(pInfo.canConnectAfter) {
					continue // We can't connect to this node yet
				}
			}

			sortedMailservers = append(sortedMailservers, sortedMailserver)

		}
		sort.Sort(byRTTMsAndCanConnectBefore(sortedMailservers))

		// Picks a random mailserver amongs the ones with the lowest latency
		// The pool size is 1/4 of the mailservers were pinged successfully
		pSize := poolSize(len(sortedMailservers) - 1)
		if pSize <= 0 {
			pSize = len(sortedMailservers)
			if pSize <= 0 {
				m.logger.Warn("No mailservers available") // Do nothing...
				return nil
			}
		}

		r, err := rand.Int(rand.Reader, big.NewInt(int64(pSize)))
		if err != nil {
			return err
		}

		msPing := sortedMailservers[r.Int64()]
		ms := mailserversByAddress[msPing.Address]
		m.logger.Info("connecting to mailserver", zap.String("address", ms.Address))
		return m.connectToMailserver(ms)
	}

	mailserversByAddress := make(map[string]mailservers.Mailserver)
	for idx := range allMailservers {
		mailserversByAddress[allMailservers[idx].Address] = allMailservers[idx]
	}

	pSize := poolSize(len(allMailservers) - 1)
	if pSize <= 0 {
		pSize = len(allMailservers)
		if pSize <= 0 {
			m.logger.Warn("No mailservers available") // Do nothing...
			return nil
		}
	}

	r, err := rand.Int(rand.Reader, big.NewInt(int64(pSize)))
	if err != nil {
		return err
	}

	msPing := allMailservers[r.Int64()]
	ms := mailserversByAddress[msPing.Address]
	m.logger.Info("connecting to mailserver", zap.String("address", ms.Address))
	return m.connectToMailserver(ms)

}

func (m *Messenger) mailserverStatus(mailserverID string) connStatus {
	m.mailPeersMutex.RLock()
	defer m.mailPeersMutex.RUnlock()
	peer, ok := m.mailserverCycle.peers[mailserverID]
	if !ok {
		return disconnected
	}
	return peer.status
}

func (m *Messenger) connectToMailserver(ms mailservers.Mailserver) error {

	m.logger.Info("connecting to mailserver", zap.Any("peer", ms.ID))

	m.mailserverCycle.activeMailserver = &ms
	signal.SendMailserverChanged(m.mailserverCycle.activeMailserver.Address, m.mailserverCycle.activeMailserver.ID)

	// Adding a peer and marking it as connected can't be executed sync in WakuV1, because
	// There's a delay between requesting a peer being added, and a signal being
	// received after the peer was added. So we first set the peer status as
	// Connecting and once a peerConnected signal is received, we mark it as
	// Connected
	activeMailserverStatus := m.mailserverStatus(ms.ID)
	if ms.Version != m.transport.WakuVersion() {
		return errors.New("mailserver waku version doesn't match")
	}

	if activeMailserverStatus != connected {
		// WakuV2 does not require having the peer connected to query the peer

		// Attempt to connect to mailserver by adding it as a peer
		if ms.Version == 1 {
			node, err := ms.Enode()
			if err != nil {
				return err
			}
			m.server.AddPeer(node)
			if err := m.peerStore.Update([]*enode.Node{node}); err != nil {
				return err
			}
		}

		connectionStatus := connecting
		if ms.Version == 2 {
			connectionStatus = connected
		}

		m.mailPeersMutex.Lock()
		m.mailserverCycle.peers[ms.ID] = peerStatus{
			status:                connectionStatus,
			lastConnectionAttempt: time.Now(),
			canConnectAfter:       time.Now().Add(defaultBackoff),
			mailserver:            ms,
		}
		m.mailPeersMutex.Unlock()

		if ms.Version == 2 {
			m.mailserverCycle.activeMailserver.FailedRequests = 0
			m.logger.Info("mailserver available", zap.String("address", m.mailserverCycle.activeMailserver.UniqueID()))
			m.EmitMailserverAvailable()
			signal.SendMailserverAvailable(m.mailserverCycle.activeMailserver.Address, m.mailserverCycle.activeMailserver.ID)

			// Query mailserver
			if m.config.codeControlFlags.AutoRequestHistoricMessages {
				go func() {
					_, err := m.performMailserverRequest(&ms, func(_ mailservers.Mailserver) (*MessengerResponse, error) {
						return m.RequestAllHistoricMessages(false, false)
					})
					if err != nil {
						m.logger.Error("could not perform mailserver request", zap.Error(err))
					}
				}()
			}
		}
	}
	return nil
}

// getActiveMailserver returns the active mailserver if a communityID is present then it'll return the mailserver
// for that community if it has a mailserver setup otherwise it'll return the global mailserver
func (m *Messenger) getActiveMailserver(communityID ...string) *mailservers.Mailserver {
	if len(communityID) == 0 || communityID[0] == "" {
		return m.mailserverCycle.activeMailserver
	}
	ms, err := m.communityStorenodes.GetStorenodeByCommunnityID(communityID[0])
	if err != nil {
		if !errors.Is(err, storenodes.ErrNotFound) {
			m.logger.Error("getting storenode for community, using global", zap.String("communityID", communityID[0]), zap.Error(err))
		}
		// if we don't find a specific mailserver for the community, we just use the regular mailserverCycle's one
		return m.mailserverCycle.activeMailserver
	}
	return &ms
}

func (m *Messenger) getActiveMailserverID(communityID ...string) string {
	ms := m.getActiveMailserver(communityID...)
	if ms == nil {
		return ""
	}
	return ms.ID
}

func (m *Messenger) isMailserverAvailable(mailserverID string) bool {
	return m.mailserverStatus(mailserverID) == connected
}

func mailserverAddressToID(uniqueID string, allMailservers []mailservers.Mailserver) (string, error) {
	for _, ms := range allMailservers {
		if uniqueID == ms.UniqueID() {
			return ms.ID, nil
		}

	}

	return "", nil
}

type ConnectedPeer struct {
	UniqueID string
}

func (m *Messenger) mailserverPeersInfo() []ConnectedPeer {
	var connectedPeers []ConnectedPeer
	for _, connectedPeer := range m.server.PeersInfo() {
		connectedPeers = append(connectedPeers, ConnectedPeer{
			// This is a bit fragile, but should work
			UniqueID: strings.TrimSuffix(connectedPeer.Enode, "?discport=0"),
		})
	}

	return connectedPeers
}

func (m *Messenger) penalizeMailserver(id string) {
	m.mailPeersMutex.Lock()
	defer m.mailPeersMutex.Unlock()
	pInfo, ok := m.mailserverCycle.peers[id]
	if !ok {
		pInfo.status = disconnected
	}

	pInfo.canConnectAfter = time.Now().Add(graylistBackoff)
	m.mailserverCycle.peers[id] = pInfo
}

// handleMailserverCycleEvent runs every 1 second or when updating peers to keep the data of the active mailserver updated
func (m *Messenger) handleMailserverCycleEvent(connectedPeers []ConnectedPeer) error {
	m.logger.Debug("mailserver cycle event",
		zap.Any("connected", connectedPeers),
		zap.Any("peer-info", m.mailserverCycle.peers))

	m.mailPeersMutex.Lock()

	for pID, pInfo := range m.mailserverCycle.peers {
		if pInfo.status == disconnected {
			continue
		}

		// Removing disconnected

		found := false
		for _, connectedPeer := range connectedPeers {
			id, err := mailserverAddressToID(connectedPeer.UniqueID, m.mailserverCycle.allMailservers)
			if err != nil {
				m.logger.Error("failed to convert id to hex", zap.Error(err))
				return err
			}

			if pID == id {
				found = true
				break
			}
		}
		if !found && (pInfo.status == connected || (pInfo.status == connecting && pInfo.lastConnectionAttempt.Add(8*time.Second).Before(time.Now()))) {
			m.logger.Info("peer disconnected", zap.String("peer", pID))
			pInfo.status = disconnected
			pInfo.canConnectAfter = time.Now().Add(defaultBackoff)
		}

		m.mailserverCycle.peers[pID] = pInfo
	}
	m.mailPeersMutex.Unlock()

	// Only evaluate connected peers once a mailserver has been set
	// otherwise, we would attempt to retrieve history and end up with a mailserver
	// not available error
	if m.mailserverCycle.activeMailserver != nil {
		for _, connectedPeer := range connectedPeers {
			id, err := mailserverAddressToID(connectedPeer.UniqueID, m.mailserverCycle.allMailservers)
			if err != nil {
				m.logger.Error("failed to convert id to hex", zap.Error(err))
				return err
			}
			if id == "" {
				continue
			}

			m.mailPeersMutex.Lock()
			pInfo, ok := m.mailserverCycle.peers[id]
			if !ok || pInfo.status != connected {
				m.logger.Info("peer connected", zap.String("peer", connectedPeer.UniqueID))
				pInfo.status = connected
				if pInfo.canConnectAfter.Before(time.Now()) {
					pInfo.canConnectAfter = time.Now().Add(defaultBackoff)
				}
				m.mailserverCycle.peers[id] = pInfo
				m.mailPeersMutex.Unlock()

				if id == m.mailserverCycle.activeMailserver.ID {
					m.mailserverCycle.activeMailserver.FailedRequests = 0
					m.logger.Info("mailserver available", zap.String("address", connectedPeer.UniqueID))
					m.EmitMailserverAvailable()
					signal.SendMailserverAvailable(m.mailserverCycle.activeMailserver.Address, m.mailserverCycle.activeMailserver.ID)
				}
				// Query mailserver
				if m.config.codeControlFlags.AutoRequestHistoricMessages {
					go func() {
						_, err := m.RequestAllHistoricMessages(false, true)
						if err != nil {
							m.logger.Error("failed to request historic messages", zap.Error(err))
						}
					}()
				}
			} else {
				m.mailPeersMutex.Unlock()
			}
		}
	}

	// Check whether we want to disconnect the mailserver
	if m.mailserverCycle.activeMailserver != nil {
		if m.mailserverCycle.activeMailserver.FailedRequests >= mailserverMaxFailedRequests {
			m.penalizeMailserver(m.mailserverCycle.activeMailserver.ID)
			signal.SendMailserverNotWorking()
			m.logger.Info("connecting too many failed requests")
			m.mailserverCycle.activeMailserver.FailedRequests = 0

			return m.connectToNewMailserverAndWait()
		}

		m.mailPeersMutex.Lock()
		pInfo, ok := m.mailserverCycle.peers[m.mailserverCycle.activeMailserver.ID]
		m.mailPeersMutex.Unlock()

		if ok {
			if pInfo.status != connected && pInfo.lastConnectionAttempt.Add(20*time.Second).Before(time.Now()) {
				m.logger.Info("penalizing mailserver & disconnecting connecting", zap.String("id", m.mailserverCycle.activeMailserver.ID))

				signal.SendMailserverNotWorking()
				m.penalizeMailserver(m.mailserverCycle.activeMailserver.ID)
				m.disconnectActiveMailserver()
			}
		}

	} else {
		m.cycleMailservers()
	}

	m.logger.Debug("updated-peers", zap.Any("peers", m.mailserverCycle.peers))

	return nil
}

func (m *Messenger) asyncRequestAllHistoricMessages() {
	m.logger.Debug("asyncRequestAllHistoricMessages")
	go func() {
		_, err := m.RequestAllHistoricMessages(false, true)
		if err != nil {
			m.logger.Error("failed to request historic messages", zap.Error(err))
		}
	}()
}

func (m *Messenger) updateWakuV1PeerStatus() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := m.handleMailserverCycleEvent(m.mailserverPeersInfo())
			if err != nil {
				m.logger.Error("failed to handle mailserver cycle event", zap.Error(err))
				continue
			}

			ms := m.getActiveMailserver()
			if ms != nil {
				node, err := ms.Enode()
				if err != nil {
					m.logger.Error("failed to parse enode", zap.Error(err))
					continue
				}
				m.server.AddPeer(node)
				if err := m.peerStore.Update([]*enode.Node{node}); err != nil {
					m.logger.Error("failed to update peers", zap.Error(err))
					continue
				}
			}

		case <-m.mailserverCycle.events:
			err := m.handleMailserverCycleEvent(m.mailserverPeersInfo())
			if err != nil {
				m.logger.Error("failed to handle mailserver cycle event", zap.Error(err))
				return
			}
		case <-m.quit:
			close(m.mailserverCycle.events)
			m.mailserverCycle.subscription.Unsubscribe()
			return
		}
	}
}

func (m *Messenger) verifyStorenodeStatus() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := m.disconnectStorenodeIfRequired()
			if err != nil {
				m.logger.Error("failed to handle mailserver cycle event", zap.Error(err))
				continue
			}

		case <-m.quit:
			return
		}
	}
}

func (m *Messenger) getPinnedMailserver() (*mailservers.Mailserver, error) {
	fleet, err := m.getFleet()
	if err != nil {
		return nil, err
	}

	pinnedMailservers, err := m.settings.GetPinnedMailservers()
	if err != nil {
		return nil, err
	}

	pinnedMailserver, ok := pinnedMailservers[fleet]
	if !ok {
		return nil, nil
	}

	fleetMailservers := mailservers.DefaultMailservers()

	for _, c := range fleetMailservers {
		if c.Fleet == fleet && c.ID == pinnedMailserver {
			return &c, nil
		}
	}

	if m.mailserversDatabase != nil {
		customMailservers, err := m.mailserversDatabase.Mailservers()
		if err != nil {
			return nil, err
		}

		for _, c := range customMailservers {
			if c.Fleet == fleet && c.ID == pinnedMailserver {
				c.Version = m.transport.WakuVersion()
				return &c, nil
			}
		}
	}

	return nil, nil
}

func (m *Messenger) EmitMailserverAvailable() {
	for _, s := range m.mailserverCycle.availabilitySubscriptions {
		s <- struct{}{}
		close(s)
		l := len(m.mailserverCycle.availabilitySubscriptions)
		m.mailserverCycle.availabilitySubscriptions = m.mailserverCycle.availabilitySubscriptions[:l-1]
	}
}

func (m *Messenger) SubscribeMailserverAvailable() chan struct{} {
	c := make(chan struct{})
	m.mailserverCycle.availabilitySubscriptions = append(m.mailserverCycle.availabilitySubscriptions, c)
	return c
}

func (m *Messenger) disconnectStorenodeIfRequired() error {
	m.logger.Debug("wakuV2 storenode status verification")

	if m.mailserverCycle.activeMailserver == nil {
		// No active storenode, find a new one
		m.cycleMailservers()
		return nil
	}

	// Check whether we want to disconnect the active storenode
	if m.mailserverCycle.activeMailserver.FailedRequests >= mailserverMaxFailedRequests {
		m.penalizeMailserver(m.mailserverCycle.activeMailserver.ID)
		signal.SendMailserverNotWorking()
		m.logger.Info("too many failed requests", zap.String("storenode", m.mailserverCycle.activeMailserver.UniqueID()))
		m.mailserverCycle.activeMailserver.FailedRequests = 0
		return m.connectToNewMailserverAndWait()
	}

	return nil
}

func (m *Messenger) waitForAvailableStoreNode(timeout time.Duration) bool {
	// Add 1 second to timeout, because the mailserver cycle has 1 second ticker, which doesn't tick on start.
	// This can be improved after merging https://github.com/status-im/status-go/pull/4380.
	// NOTE: https://stackoverflow.com/questions/32705582/how-to-get-time-tick-to-tick-immediately
	timeout += time.Second

	finish := make(chan struct{})
	cancel := make(chan struct{})

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer func() {
			wg.Done()
		}()
		for !m.isMailserverAvailable(m.getActiveMailserverID()) {
			select {
			case <-m.SubscribeMailserverAvailable():
			case <-cancel:
				return
			}
		}
	}()

	go func() {
		defer func() {
			close(finish)
		}()
		wg.Wait()
	}()

	select {
	case <-finish:
	case <-time.After(timeout):
		close(cancel)
	case <-m.ctx.Done():
		close(cancel)
	}

	return m.isMailserverAvailable(m.getActiveMailserverID())
}
