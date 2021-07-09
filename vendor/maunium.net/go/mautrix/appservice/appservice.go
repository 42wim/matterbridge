// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package appservice

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"golang.org/x/net/publicsuffix"
	"gopkg.in/yaml.v2"

	"maunium.net/go/maulogger/v2"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// EventChannelSize is the size for the Events channel in Appservice instances.
var EventChannelSize = 64

// Create a blank appservice instance.
func Create() *AppService {
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	return &AppService{
		LogConfig:  CreateLogConfig(),
		clients:    make(map[id.UserID]*mautrix.Client),
		intents:    make(map[id.UserID]*IntentAPI),
		HTTPClient: &http.Client{Timeout: 180 * time.Second, Jar: jar},
		StateStore: NewBasicStateStore(),
		Router:     mux.NewRouter(),
		UserAgent:  mautrix.DefaultUserAgent,
	}
}

// Load an appservice config from a file.
func Load(path string) (*AppService, error) {
	data, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		return nil, readErr
	}

	config := Create()
	return config, yaml.Unmarshal(data, config)
}

// QueryHandler handles room alias and user ID queries from the homeserver.
type QueryHandler interface {
	QueryAlias(alias string) bool
	QueryUser(userID id.UserID) bool
}

type QueryHandlerStub struct{}

func (qh *QueryHandlerStub) QueryAlias(alias string) bool {
	return false
}

func (qh *QueryHandlerStub) QueryUser(userID id.UserID) bool {
	return false
}

// AppService is the main config for all appservices.
// It also serves as the appservice instance struct.
type AppService struct {
	HomeserverDomain string     `yaml:"homeserver_domain"`
	HomeserverURL    string     `yaml:"homeserver_url"`
	RegistrationPath string     `yaml:"registration"`
	Host             HostConfig `yaml:"host"`
	LogConfig        LogConfig  `yaml:"logging"`
	Sync             struct {
		Enabled   bool   `yaml:"enabled"`
		FilterID  string `yaml:"filter_id"`
		NextBatch string `yaml:"next_batch"`
	} `yaml:"sync"`

	Registration *Registration    `yaml:"-"`
	Log          maulogger.Logger `yaml:"-"`

	lastProcessedTransaction string

	Events       chan *event.Event `yaml:"-"`
	QueryHandler QueryHandler      `yaml:"-"`
	StateStore   StateStore        `yaml:"-"`

	Router     *mux.Router `yaml:"-"`
	UserAgent  string      `yaml:"-"`
	server     *http.Server
	HTTPClient *http.Client
	botClient  *mautrix.Client
	botIntent  *IntentAPI

	DefaultHTTPRetries int

	clients     map[id.UserID]*mautrix.Client
	clientsLock sync.RWMutex
	intents     map[id.UserID]*IntentAPI
	intentsLock sync.RWMutex

	ws                *websocket.Conn
	StopWebsocket     func(error)
	WebsocketCommands chan WebsocketCommand
}

func (as *AppService) PrepareWebsocket() {
	if as.WebsocketCommands == nil {
		as.WebsocketCommands = make(chan WebsocketCommand, 32)
	}
}

// HostConfig contains info about how to host the appservice.
type HostConfig struct {
	Hostname string `yaml:"hostname"`
	Port     uint16 `yaml:"port"`
	TLSKey   string `yaml:"tls_key,omitempty"`
	TLSCert  string `yaml:"tls_cert,omitempty"`
}

// Address gets the whole address of the Appservice.
func (hc *HostConfig) Address() string {
	return fmt.Sprintf("%s:%d", hc.Hostname, hc.Port)
}

// Save saves this config into a file at the given path.
func (as *AppService) Save(path string) error {
	data, err := yaml.Marshal(as)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

// YAML returns the config in YAML format.
func (as *AppService) YAML() (string, error) {
	data, err := yaml.Marshal(as)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (as *AppService) BotMXID() id.UserID {
	return id.NewUserID(as.Registration.SenderLocalpart, as.HomeserverDomain)
}

func (as *AppService) makeIntent(userID id.UserID) *IntentAPI {
	as.intentsLock.Lock()
	defer as.intentsLock.Unlock()

	intent, ok := as.intents[userID]
	if ok {
		return intent
	}

	localpart, homeserver, err := userID.Parse()
	if err != nil || len(localpart) == 0 || homeserver != as.HomeserverDomain {
		if err != nil {
			as.Log.Fatalfln("Failed to parse user ID %s: %v", userID, err)
		} else if len(localpart) == 0 {
			as.Log.Fatalfln("Failed to make intent for %s: localpart is empty", userID)
		} else if homeserver != as.HomeserverDomain {
			as.Log.Fatalfln("Failed to make intent for %s: homeserver isn't %s", userID, as.HomeserverDomain)
		}
		return nil
	}
	intent = as.NewIntentAPI(localpart)
	as.intents[userID] = intent
	return intent
}

func (as *AppService) Intent(userID id.UserID) *IntentAPI {
	as.intentsLock.RLock()
	intent, ok := as.intents[userID]
	as.intentsLock.RUnlock()
	if !ok {
		return as.makeIntent(userID)
	}
	return intent
}

func (as *AppService) BotIntent() *IntentAPI {
	if as.botIntent == nil {
		as.botIntent = as.makeIntent(as.BotMXID())
	}
	return as.botIntent
}

func (as *AppService) makeClient(userID id.UserID) *mautrix.Client {
	as.clientsLock.Lock()
	defer as.clientsLock.Unlock()

	client, ok := as.clients[userID]
	if ok {
		return client
	}

	client, err := mautrix.NewClient(as.HomeserverURL, userID, as.Registration.AppToken)
	if err != nil {
		as.Log.Fatalln("Failed to create mautrix client instance:", err)
		return nil
	}
	client.UserAgent = as.UserAgent
	client.Syncer = nil
	client.Store = nil
	client.AppServiceUserID = userID
	client.Logger = as.Log.Sub(string(userID))
	client.Client = as.HTTPClient
	client.DefaultHTTPRetries = as.DefaultHTTPRetries
	as.clients[userID] = client
	return client
}

func (as *AppService) Client(userID id.UserID) *mautrix.Client {
	as.clientsLock.RLock()
	client, ok := as.clients[userID]
	as.clientsLock.RUnlock()
	if !ok {
		return as.makeClient(userID)
	}
	return client
}

func (as *AppService) BotClient() *mautrix.Client {
	if as.botClient == nil {
		as.botClient = as.makeClient(as.BotMXID())
		as.botClient.Logger = as.Log.Sub("Bot")
	}
	return as.botClient
}

// Init initializes the logger and loads the registration of this appservice.
func (as *AppService) Init() (bool, error) {
	as.Events = make(chan *event.Event, EventChannelSize)
	as.QueryHandler = &QueryHandlerStub{}

	if len(as.UserAgent) == 0 {
		as.UserAgent = mautrix.DefaultUserAgent
	}

	as.Log = maulogger.Create()
	as.LogConfig.Configure(as.Log)
	as.Log.Debugln("Logger initialized successfully.")

	if len(as.RegistrationPath) > 0 {
		var err error
		as.Registration, err = LoadRegistration(as.RegistrationPath)
		if err != nil {
			return false, err
		}
	}

	as.Log.Debugln("Appservice initialized successfully.")
	return true, nil
}

// LogConfig contains configs for the logger.
type LogConfig struct {
	Directory       string `yaml:"directory"`
	FileNameFormat  string `yaml:"file_name_format"`
	FileDateFormat  string `yaml:"file_date_format"`
	FileMode        uint32 `yaml:"file_mode"`
	TimestampFormat string `yaml:"timestamp_format"`
	RawPrintLevel   string `yaml:"print_level"`
	JSONStdout      bool   `yaml:"print_json"`
	JSONFile        bool   `yaml:"file_json"`
	PrintLevel      int    `yaml:"-"`
}

type umLogConfig LogConfig

func (lc *LogConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal((*umLogConfig)(lc))
	if err != nil {
		return err
	}

	switch strings.ToUpper(lc.RawPrintLevel) {
	case "TRACE":
		lc.PrintLevel = -10
	case "DEBUG":
		lc.PrintLevel = maulogger.LevelDebug.Severity
	case "INFO":
		lc.PrintLevel = maulogger.LevelInfo.Severity
	case "WARN", "WARNING":
		lc.PrintLevel = maulogger.LevelWarn.Severity
	case "ERR", "ERROR":
		lc.PrintLevel = maulogger.LevelError.Severity
	case "FATAL":
		lc.PrintLevel = maulogger.LevelFatal.Severity
	default:
		return errors.New("invalid print level " + lc.RawPrintLevel)
	}
	return err
}

func (lc *LogConfig) MarshalYAML() (interface{}, error) {
	switch {
	case lc.PrintLevel >= maulogger.LevelFatal.Severity:
		lc.RawPrintLevel = maulogger.LevelFatal.Name
	case lc.PrintLevel >= maulogger.LevelError.Severity:
		lc.RawPrintLevel = maulogger.LevelError.Name
	case lc.PrintLevel >= maulogger.LevelWarn.Severity:
		lc.RawPrintLevel = maulogger.LevelWarn.Name
	case lc.PrintLevel >= maulogger.LevelInfo.Severity:
		lc.RawPrintLevel = maulogger.LevelInfo.Name
	default:
		lc.RawPrintLevel = maulogger.LevelDebug.Name
	}
	return lc, nil
}

// CreateLogConfig creates a basic LogConfig.
func CreateLogConfig() LogConfig {
	return LogConfig{
		Directory:       "./logs",
		FileNameFormat:  "%[1]s-%02[2]d.log",
		TimestampFormat: "Jan _2, 2006 15:04:05",
		FileMode:        0600,
		FileDateFormat:  "2006-01-02",
		PrintLevel:      10,
	}
}

type FileFormatData struct {
	Date  string
	Index int
}

// GetFileFormat returns a mauLogger-compatible logger file format based on the data in the struct.
func (lc LogConfig) GetFileFormat() maulogger.LoggerFileFormat {
	if len(lc.Directory) > 0 {
		_ = os.MkdirAll(lc.Directory, 0700)
	}
	path := filepath.Join(lc.Directory, lc.FileNameFormat)
	tpl, _ := template.New("fileformat").Parse(path)

	return func(now string, i int) string {
		var buf strings.Builder
		_ = tpl.Execute(&buf, FileFormatData{
			Date:  now,
			Index: i,
		})
		return buf.String()
	}
}

// Configure configures a mauLogger instance with the data in this struct.
func (lc LogConfig) Configure(log maulogger.Logger) {
	basicLogger := log.(*maulogger.BasicLogger)
	basicLogger.FileFormat = lc.GetFileFormat()
	basicLogger.FileMode = os.FileMode(lc.FileMode)
	basicLogger.FileTimeFormat = lc.FileDateFormat
	basicLogger.TimeFormat = lc.TimestampFormat
	basicLogger.PrintLevel = lc.PrintLevel
	basicLogger.JSONFile = lc.JSONFile
	if lc.JSONStdout {
		basicLogger.EnableJSONStdout()
	}
}
