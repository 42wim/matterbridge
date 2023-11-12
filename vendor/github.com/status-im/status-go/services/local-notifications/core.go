package localnotifications

import (
	"database/sql"
	"encoding/json"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/services/wallet/transfer"
	"github.com/status-im/status-go/signal"
)

type PushCategory string

type NotificationType string

type NotificationBody interface {
	json.Marshaler
}

type Notification struct {
	ID                  common.Hash
	Platform            float32
	Body                NotificationBody
	BodyType            NotificationType
	Title               string
	Message             string
	Category            PushCategory
	Deeplink            string
	Image               string
	IsScheduled         bool
	ScheduledTime       string
	IsConversation      bool
	IsGroupConversation bool
	ConversationID      string
	Timestamp           uint64
	Author              NotificationAuthor
	Deleted             bool
}

type NotificationAuthor struct {
	ID   string `json:"id"`
	Icon string `json:"icon"`
	Name string `json:"name"`
}

// notificationAlias is an interim struct used for json un/marshalling
type notificationAlias struct {
	ID                  common.Hash        `json:"id"`
	Platform            float32            `json:"platform,omitempty"`
	Body                json.RawMessage    `json:"body"`
	BodyType            NotificationType   `json:"bodyType"`
	Title               string             `json:"title,omitempty"`
	Message             string             `json:"message,omitempty"`
	Category            PushCategory       `json:"category,omitempty"`
	Deeplink            string             `json:"deepLink,omitempty"`
	Image               string             `json:"imageUrl,omitempty"`
	IsScheduled         bool               `json:"isScheduled,omitempty"`
	ScheduledTime       string             `json:"scheduleTime,omitempty"`
	IsConversation      bool               `json:"isConversation,omitempty"`
	IsGroupConversation bool               `json:"isGroupConversation,omitempty"`
	ConversationID      string             `json:"conversationId,omitempty"`
	Timestamp           uint64             `json:"timestamp,omitempty"`
	Author              NotificationAuthor `json:"notificationAuthor,omitempty"`
	Deleted             bool               `json:"deleted,omitempty"`
}

// MessageEvent - structure used to pass messages from chat to bus
type MessageEvent struct{}

// CustomEvent - structure used to pass custom user set messages to bus
type CustomEvent struct{}

type transmitter struct {
	publisher *event.Feed

	wg   sync.WaitGroup
	quit chan struct{}
}

// Service keeps the state of message bus
type Service struct {
	started           bool
	WatchingEnabled   bool
	chainID           uint64
	transmitter       *transmitter
	walletTransmitter *transmitter
	db                *Database
	walletDB          *transfer.Database
	accountsDB        *accounts.Database
}

func NewService(appDB *sql.DB, walletDB *transfer.Database, chainID uint64) (*Service, error) {
	db := NewDB(appDB, chainID)
	accountsDB, err := accounts.NewDB(appDB)
	if err != nil {
		return nil, err
	}
	trans := &transmitter{}
	walletTrans := &transmitter{}

	return &Service{
		db:                db,
		chainID:           chainID,
		walletDB:          walletDB,
		accountsDB:        accountsDB,
		transmitter:       trans,
		walletTransmitter: walletTrans,
	}, nil
}

func (n *Notification) MarshalJSON() ([]byte, error) {

	var body json.RawMessage
	if n.Body != nil {
		encodedBody, err := n.Body.MarshalJSON()
		if err != nil {
			return nil, err
		}
		body = encodedBody
	}

	alias := notificationAlias{
		ID:                  n.ID,
		Platform:            n.Platform,
		Body:                body,
		BodyType:            n.BodyType,
		Category:            n.Category,
		Title:               n.Title,
		Message:             n.Message,
		Deeplink:            n.Deeplink,
		Image:               n.Image,
		IsScheduled:         n.IsScheduled,
		ScheduledTime:       n.ScheduledTime,
		IsConversation:      n.IsConversation,
		IsGroupConversation: n.IsGroupConversation,
		ConversationID:      n.ConversationID,
		Timestamp:           n.Timestamp,
		Author:              n.Author,
		Deleted:             n.Deleted,
	}

	return json.Marshal(alias)
}

func PushMessages(ns []*Notification) {
	for _, n := range ns {
		pushMessage(n)
	}
}

func pushMessage(notification *Notification) {
	log.Debug("Pushing a new push notification")
	signal.SendLocalNotifications(notification)
}

// Start Worker which processes all incoming messages
func (s *Service) Start() error {
	s.started = true

	s.transmitter.quit = make(chan struct{})
	s.transmitter.publisher = &event.Feed{}

	events := make(chan TransactionEvent, 10)
	sub := s.transmitter.publisher.Subscribe(events)

	s.transmitter.wg.Add(1)
	go func() {
		defer s.transmitter.wg.Done()
		for {
			select {
			case <-s.transmitter.quit:
				sub.Unsubscribe()
				return
			case err := <-sub.Err():
				if err != nil {
					log.Error("Local notifications transmitter failed with", "error", err)
				}
				return
			case event := <-events:
				s.transactionsHandler(event)
			}
		}
	}()

	log.Info("Successful start")

	return nil
}

// Stop worker
func (s *Service) Stop() error {
	s.started = false

	if s.transmitter.quit != nil {
		close(s.transmitter.quit)
		s.transmitter.wg.Wait()
		s.transmitter.quit = nil
	}

	if s.walletTransmitter.quit != nil {
		close(s.walletTransmitter.quit)
		s.walletTransmitter.wg.Wait()
		s.walletTransmitter.quit = nil
	}

	return nil
}

// APIs returns list of available RPC APIs.
func (s *Service) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "localnotifications",
			Version:   "0.1.0",
			Service:   NewAPI(s),
		},
	}
}

// Protocols returns list of p2p protocols.
func (s *Service) Protocols() []p2p.Protocol {
	return nil
}

func (s *Service) IsStarted() bool {
	return s.started
}
