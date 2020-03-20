package bslack

import (
	"errors"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/matterhook"
	"github.com/slack-go/slack"
)

type BLegacy struct {
	*Bslack
}

func NewLegacy(cfg *bridge.Config) bridge.Bridger {
	b := &BLegacy{Bslack: newBridge(cfg)}
	b.legacy = true
	return b
}

func (b *BLegacy) Connect() error {
	b.RLock()
	defer b.RUnlock()
	if b.GetString(incomingWebhookConfig) != "" {
		switch {
		case b.GetString(outgoingWebhookConfig) != "":
			b.Log.Info("Connecting using webhookurl (sending) and webhookbindaddress (receiving)")
			b.mh = matterhook.New(b.GetString(outgoingWebhookConfig), matterhook.Config{
				InsecureSkipVerify: b.GetBool(skipTLSConfig),
				BindAddress:        b.GetString(incomingWebhookConfig),
			})
		case b.GetString(tokenConfig) != "":
			b.Log.Info("Connecting using token (sending)")
			b.sc = slack.New(b.GetString(tokenConfig))
			b.rtm = b.sc.NewRTM()
			go b.rtm.ManageConnection()
			b.Log.Info("Connecting using webhookbindaddress (receiving)")
			b.mh = matterhook.New(b.GetString(outgoingWebhookConfig), matterhook.Config{
				InsecureSkipVerify: b.GetBool(skipTLSConfig),
				BindAddress:        b.GetString(incomingWebhookConfig),
			})
		default:
			b.Log.Info("Connecting using webhookbindaddress (receiving)")
			b.mh = matterhook.New(b.GetString(outgoingWebhookConfig), matterhook.Config{
				InsecureSkipVerify: b.GetBool(skipTLSConfig),
				BindAddress:        b.GetString(incomingWebhookConfig),
			})
		}
		go b.handleSlack()
		return nil
	}
	if b.GetString(outgoingWebhookConfig) != "" {
		b.Log.Info("Connecting using webhookurl (sending)")
		b.mh = matterhook.New(b.GetString(outgoingWebhookConfig), matterhook.Config{
			InsecureSkipVerify: b.GetBool(skipTLSConfig),
			DisableServer:      true,
		})
		if b.GetString(tokenConfig) != "" {
			b.Log.Info("Connecting using token (receiving)")
			b.sc = slack.New(b.GetString(tokenConfig), slack.OptionDebug(b.GetBool("debug")))
			b.channels = newChannelManager(b.Log, b.sc)
			b.users = newUserManager(b.Log, b.sc)
			b.rtm = b.sc.NewRTM()
			go b.rtm.ManageConnection()
			go b.handleSlack()
		}
	} else if b.GetString(tokenConfig) != "" {
		b.Log.Info("Connecting using token (sending and receiving)")
		b.sc = slack.New(b.GetString(tokenConfig), slack.OptionDebug(b.GetBool("debug")))
		b.channels = newChannelManager(b.Log, b.sc)
		b.users = newUserManager(b.Log, b.sc)
		b.rtm = b.sc.NewRTM()
		go b.rtm.ManageConnection()
		go b.handleSlack()
	}
	if b.GetString(incomingWebhookConfig) == "" && b.GetString(outgoingWebhookConfig) == "" && b.GetString(tokenConfig) == "" {
		return errors.New("no connection method found. See that you have WebhookBindAddress, WebhookURL or Token configured")
	}
	return nil
}
