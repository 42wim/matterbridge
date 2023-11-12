package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

type Server struct {
	isRunning bool
	server    *http.Server
	logger    *zap.Logger
	cert      *tls.Certificate
	hostname  string
	handlers  HandlerPatternMap

	portManger
	*timeoutManager
}

func NewServer(cert *tls.Certificate, hostname string, afterPortChanged func(int), logger *zap.Logger) Server {
	return Server{
		logger:         logger,
		cert:           cert,
		hostname:       hostname,
		portManger:     newPortManager(logger.Named("Server"), afterPortChanged),
		timeoutManager: newTimeoutManager(),
	}
}

func (s *Server) getHost() string {
	return fmt.Sprintf("%s:%d", s.hostname, s.GetPort())
}

func (s *Server) GetHostname() string {
	return s.hostname
}

func (s *Server) GetCert() *tls.Certificate {
	return s.cert
}

func (s *Server) GetLogger() *zap.Logger {
	return s.logger
}

func (s *Server) mustGetHost() string {
	return fmt.Sprintf("%s:%d", s.hostname, s.MustGetPort())
}

func (s *Server) listenAndServe() {
	cfg := &tls.Config{Certificates: []tls.Certificate{*s.cert}, ServerName: s.hostname, MinVersion: tls.VersionTLS12}

	// in case of restart, we should use the same port as the first start in order not to break existing links
	listener, err := tls.Listen("tcp", s.getHost(), cfg)
	if err != nil {
		s.logger.Error("failed to start server, retrying", zap.Error(err))
		s.ResetPort()
		err = s.Start()
		if err != nil {
			s.logger.Error("server start failed, giving up", zap.Error(err))
		}
		return
	}

	err = s.SetPort(listener.Addr().(*net.TCPAddr).Port)
	if err != nil {
		s.logger.Error("failed to set Server.port", zap.Error(err))
		return
	}

	s.isRunning = true

	s.StartTimeout(func() {
		err := s.Stop()
		if err != nil {
			s.logger.Error("server termination fail", zap.Error(err))
		}
	})

	err = s.server.Serve(listener)
	if err != http.ErrServerClosed {
		s.logger.Error("server failed unexpectedly, restarting", zap.Error(err))
		err = s.Start()
		if err != nil {
			s.logger.Error("server start failed, giving up", zap.Error(err))
		}
		return
	}
	s.isRunning = false
}

func (s *Server) resetServer() {
	s.StopTimeout()
	s.server = new(http.Server)
	s.ResetPort()
}

func (s *Server) applyHandlers() {
	if s.server == nil {
		s.server = new(http.Server)
	}
	mux := http.NewServeMux()

	for p, h := range s.handlers {
		mux.HandleFunc(p, h)
	}
	s.server.Handler = mux
}

func (s *Server) Start() error {
	// Once Shutdown has been called on a server, it may not be reused;
	s.resetServer()
	s.applyHandlers()
	go s.listenAndServe()
	return nil
}

func (s *Server) Stop() error {
	s.StopTimeout()
	if s.server != nil {
		return s.server.Shutdown(context.Background())
	}

	return nil
}

func (s *Server) IsRunning() bool {
	return s.isRunning
}

func (s *Server) ToForeground() {
	if !s.isRunning && (s.server != nil) {
		err := s.Start()
		if err != nil {
			s.logger.Error("server start failed during foreground transition", zap.Error(err))
		}
	}
}

func (s *Server) ToBackground() {
	if s.isRunning {
		err := s.Stop()
		if err != nil {
			s.logger.Error("server stop failed during background transition", zap.Error(err))
		}
	}
}

func (s *Server) SetHandlers(handlers HandlerPatternMap) {
	s.handlers = handlers
}

func (s *Server) AddHandlers(handlers HandlerPatternMap) {
	if s.handlers == nil {
		s.handlers = make(HandlerPatternMap)
	}

	for name := range handlers {
		s.handlers[name] = handlers[name]
	}
}

func (s *Server) MakeBaseURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   s.mustGetHost(),
	}
}
