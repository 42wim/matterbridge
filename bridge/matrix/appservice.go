package bmatrix

import (
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"

	"maunium.net/go/mautrix/appservice"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type AppServiceNamespaces struct {
	rooms     []*regexp.Regexp
	usernames []*regexp.Regexp
	prefixes  []string
}

type AppServiceWrapper struct {
	appService *appservice.AppService
	namespaces AppServiceNamespaces
	stop       chan struct{}
	stopAck    chan struct{}
}

func (w *AppServiceWrapper) ParseNamespaces(logger *logrus.Entry) error {
	if w.appService.Registration != nil {
		// TODO: handle non-exclusive registrations
		for _, v := range w.appService.Registration.Namespaces.RoomIDs {
			re, err := regexp.Compile(v.Regex)
			if err != nil {
				logger.Warnf("couldn't parse the appservice regex '%s'", v.Regex)
				continue
			}

			w.namespaces.rooms = append(w.namespaces.rooms, re)
		}

		for _, v := range w.appService.Registration.Namespaces.UserIDs {
			re, err := regexp.Compile(v.Regex)
			if err != nil {
				logger.Warnf("couldn't parse the appservice regex '%s'", v.Regex)
				continue
			}

			// we assume that the user regexes will be of the form '@<some prefix>.*'
			// where '.*' will be replaced by the username we spoof
			prefix, _ := re.LiteralPrefix()
			if prefix == "" || prefix == "@" {
				logger.Warnf("couldn't find an acceptable prefix in the appservice regex '%s'", v.Regex)
				continue
			}

			if v.Regex != fmt.Sprintf("%s.*", prefix) {
				logger.Warnf("complex regexpes are not supported for appServices, the regexp '%s' does not match the format '@<prefix>.*'", v.Regex)
				continue
			}

			w.namespaces.usernames = append(w.namespaces.usernames, re)
			// drop the '@' in the prefix
			w.namespaces.prefixes = append(w.namespaces.prefixes, prefix[1:])
		}
	}

	return nil
}

func (b *Bmatrix) NewAppService() (*AppServiceWrapper, error) {
	w := &AppServiceWrapper{
		appService: appservice.Create(),
		namespaces: AppServiceNamespaces{
			rooms:     []*regexp.Regexp{},
			usernames: []*regexp.Regexp{},
			prefixes:  []string{},
		},
		stop:    make(chan struct{}, 1),
		stopAck: make(chan struct{}, 1),
	}

	err := w.appService.SetHomeserverURL(b.mc.HomeserverURL.String())
	if err != nil {
		return nil, err
	}

	_, homeServerDomain, _ := b.mc.UserID.Parse()
	w.appService.HomeserverDomain = homeServerDomain
	//nolint:exhaustruct
	w.appService.Host = appservice.HostConfig{
		Hostname: b.GetString("AppServiceHost"),
		Port:     uint16(b.GetInt("AppServicePort")),
	}
	w.appService.Registration, err = appservice.LoadRegistration(b.GetString("AppServiceConfigPath"))
	if err != nil {
		return nil, err
	}

	// forward logs from the appService to the matterbridge logger
	w.appService.Log = NewZerologWrapper(b.Log)

	if err = w.ParseNamespaces(b.Log); err != nil {
		return nil, err
	}

	return w, nil
}

func (a *AppServiceNamespaces) containsRoom(roomID id.RoomID) bool {
	// no room specified: we check all the rooms
	if len(a.rooms) == 0 {
		return true
	}

	for _, room := range a.rooms {
		if room.MatchString(roomID.String()) {
			return true
		}
	}

	return false
}

// nolint: wrapcheck
func (b *Bmatrix) startAppService() error {
	wrapper := b.appService
	// TODO: detect service completion and rerun automatically
	go wrapper.appService.Start()
	b.Log.Debug("appservice launched")

	processor := appservice.NewEventProcessor(wrapper.appService)
	for _, eventType := range []event.Type{event.EventRedaction, event.EventMessage} {
		processor.On(eventType, func(ev *event.Event) {
			b.handleEvent(originAppService, ev)
		})
	}
	go processor.Start()
	b.Log.Debug("appservice event dispatcher launched")

	// handle service stopping/restarting
	go func(appService *appservice.AppService, processor *appservice.EventProcessor) {
		<-wrapper.stop

		appService.Stop()
		processor.Stop()
		wrapper.stopAck <- struct{}{}
	}(wrapper.appService, processor)

	return nil
}
