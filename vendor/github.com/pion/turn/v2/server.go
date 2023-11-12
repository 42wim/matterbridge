// Package turn contains the public API for pion/turn, a toolkit for building TURN clients and servers
package turn

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pion/logging"
	"github.com/pion/turn/v2/internal/allocation"
	"github.com/pion/turn/v2/internal/proto"
	"github.com/pion/turn/v2/internal/server"
)

const (
	defaultInboundMTU = 1600
)

// Server is an instance of the Pion TURN Server
type Server struct {
	log                logging.LeveledLogger
	authHandler        AuthHandler
	realm              string
	channelBindTimeout time.Duration
	nonces             *sync.Map

	packetConnConfigs  []PacketConnConfig
	listenerConfigs    []ListenerConfig
	allocationManagers []*allocation.Manager

	inboundMTU int
}

// NewServer creates the Pion TURN server
func NewServer(config ServerConfig) (*Server, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	loggerFactory := config.LoggerFactory
	if loggerFactory == nil {
		loggerFactory = logging.NewDefaultLoggerFactory()
	}

	mtu := defaultInboundMTU
	if config.InboundMTU != 0 {
		mtu = config.InboundMTU
	}

	s := &Server{
		log:                loggerFactory.NewLogger("turn"),
		authHandler:        config.AuthHandler,
		realm:              config.Realm,
		channelBindTimeout: config.ChannelBindTimeout,
		packetConnConfigs:  config.PacketConnConfigs,
		listenerConfigs:    config.ListenerConfigs,
		allocationManagers: make([]*allocation.Manager, len(config.PacketConnConfigs)+len(config.ListenerConfigs)),
		nonces:             &sync.Map{},
		inboundMTU:         mtu,
	}

	if s.channelBindTimeout == 0 {
		s.channelBindTimeout = proto.DefaultLifetime
	}

	for i := range s.packetConnConfigs {
		go func(i int, p PacketConnConfig) {
			allocationManager, err := allocation.NewManager(allocation.ManagerConfig{
				AllocatePacketConn: p.RelayAddressGenerator.AllocatePacketConn,
				AllocateConn:       p.RelayAddressGenerator.AllocateConn,
				LeveledLogger:      s.log,
			})
			if err != nil {
				s.log.Errorf("exit read loop on error: %s", err.Error())
				return
			}
			s.allocationManagers[i] = allocationManager
			defer func() {
				if err := allocationManager.Close(); err != nil {
					s.log.Errorf("Failed to close AllocationManager: %s", err.Error())
				}
			}()

			s.readLoop(p.PacketConn, allocationManager)
		}(i, s.packetConnConfigs[i])
	}

	for i, listener := range s.listenerConfigs {
		go func(i int, l ListenerConfig) {
			allocationManager, err := allocation.NewManager(allocation.ManagerConfig{
				AllocatePacketConn: l.RelayAddressGenerator.AllocatePacketConn,
				AllocateConn:       l.RelayAddressGenerator.AllocateConn,
				LeveledLogger:      s.log,
			})
			if err != nil {
				s.log.Errorf("exit read loop on error: %s", err.Error())
				return
			}
			s.allocationManagers[i] = allocationManager
			defer func() {
				if err := allocationManager.Close(); err != nil {
					s.log.Errorf("Failed to close AllocationManager: %s", err.Error())
				}
			}()

			for {
				conn, err := l.Listener.Accept()
				if err != nil {
					s.log.Debugf("exit accept loop on error: %s", err.Error())
					return
				}

				go s.readLoop(NewSTUNConn(conn), allocationManager)
			}
		}(i+len(s.packetConnConfigs), listener)
	}

	return s, nil
}

// AllocationCount returns the number of active allocations. It can be used to drain the server before closing
func (s *Server) AllocationCount() int {
	allocations := 0
	for _, manager := range s.allocationManagers {
		if manager != nil {
			allocations += manager.AllocationCount()
		}
	}
	return allocations
}

// Close stops the TURN Server. It cleans up any associated state and closes all connections it is managing
func (s *Server) Close() error {
	var errors []error

	for _, p := range s.packetConnConfigs {
		if err := p.PacketConn.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	for _, l := range s.listenerConfigs {
		if err := l.Listener.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) == 0 {
		return nil
	}

	err := errFailedToClose
	for _, e := range errors {
		err = fmt.Errorf("%s; Close error (%v) ", err.Error(), e) //nolint:goerr113
	}

	return err
}

func (s *Server) readLoop(p net.PacketConn, allocationManager *allocation.Manager) {
	buf := make([]byte, s.inboundMTU)
	for {
		n, addr, err := p.ReadFrom(buf)
		switch {
		case err != nil:
			s.log.Debugf("exit read loop on error: %s", err.Error())
			return
		case n >= s.inboundMTU:
			s.log.Debugf("Read bytes exceeded MTU, packet is possibly truncated")
		}

		if err := server.HandleRequest(server.Request{
			Conn:               p,
			SrcAddr:            addr,
			Buff:               buf[:n],
			Log:                s.log,
			AuthHandler:        s.authHandler,
			Realm:              s.realm,
			AllocationManager:  allocationManager,
			ChannelBindTimeout: s.channelBindTimeout,
			Nonces:             s.nonces,
		}); err != nil {
			s.log.Errorf("error when handling datagram: %v", err)
		}
	}
}
