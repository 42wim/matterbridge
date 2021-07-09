package bmatrix

import (
	"errors"

	"maunium.net/go/mautrix/appservice"
	"maunium.net/go/mautrix/event"
)

//nolint: wrapcheck
func (b *Bmatrix) startAppService() error {
	appService := appservice.Create()

	appService.HomeserverURL = b.mc.HomeserverURL.String()
	_, homeServerDomain, _ := b.mc.UserID.Parse()
	appService.HomeserverDomain = homeServerDomain
	//nolint: exhaustivestruct
	appService.Host = appservice.HostConfig{
		Hostname: b.GetString("AppServiceHost"),
		Port:     uint16(b.GetInt("AppServicePort")),
	}
	appService.RegistrationPath = b.GetString("AppServiceConfigPath")

	initSuccess, err := appService.Init()
	if err != nil {
		return err
	} else if !initSuccess {
		return errors.New("couldn't initialise the application service")
	}

	b.appService = appService

	// TODO: detect service completion and rerun automatically
	go b.appService.Start()
	b.Log.Debug("appservice launched")

	processor := appservice.NewEventProcessor(b.appService)
	processor.On(event.EventMessage, func(ev *event.Event) {
		b.handleEvent(originAppService, ev)
	})
	go processor.Start()
	b.Log.Debug("appservice even dispatcher launched")

	// handle service stopping/restarting
	go func(b *Bmatrix, processor *appservice.EventProcessor) {
		<-b.stop

		b.appService.Stop()
		processor.Stop()
		b.stopAck <- struct{}{}
	}(b, processor)

	return nil
}
