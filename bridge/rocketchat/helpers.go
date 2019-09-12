package brocketchat

import (
	"context"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/42wim/matterbridge/hook/rockethook"
	"github.com/42wim/matterbridge/matterhook"
	"github.com/matterbridge/Rocket.Chat.Go.SDK/models"
	"github.com/matterbridge/Rocket.Chat.Go.SDK/realtime"
	"github.com/matterbridge/Rocket.Chat.Go.SDK/rest"
	"github.com/nelsonken/gomf"
)

func (b *Brocketchat) doConnectWebhookBind() error {
	switch {
	case b.GetString("WebhookURL") != "":
		b.Log.Info("Connecting using webhookurl (sending) and webhookbindaddress (receiving)")
		b.mh = matterhook.New(b.GetString("WebhookURL"),
			matterhook.Config{InsecureSkipVerify: b.GetBool("SkipTLSVerify"),
				DisableServer: true})
		b.rh = rockethook.New(b.GetString("WebhookURL"), rockethook.Config{BindAddress: b.GetString("WebhookBindAddress")})
	case b.GetString("Login") != "":
		b.Log.Info("Connecting using login/password (sending)")
		err := b.apiLogin()
		if err != nil {
			return err
		}
	default:
		b.Log.Info("Connecting using webhookbindaddress (receiving)")
		b.rh = rockethook.New(b.GetString("WebhookURL"), rockethook.Config{BindAddress: b.GetString("WebhookBindAddress")})
	}
	return nil
}

func (b *Brocketchat) doConnectWebhookURL() error {
	b.Log.Info("Connecting using webhookurl (sending)")
	b.mh = matterhook.New(b.GetString("WebhookURL"),
		matterhook.Config{InsecureSkipVerify: b.GetBool("SkipTLSVerify"),
			DisableServer: true})
	if b.GetString("Login") != "" {
		b.Log.Info("Connecting using login/password (receiving)")
		err := b.apiLogin()
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Brocketchat) apiLogin() error {
	b.Log.Debugf("handling apiLogin()")
	credentials := &models.UserCredentials{Email: b.GetString("login"), Password: b.GetString("password")}
	if b.GetString("Token") != "" {
		credentials = &models.UserCredentials{ID: b.GetString("Login"), Token: b.GetString("Token")}
	}
	myURL, err := url.Parse(b.GetString("server"))
	if err != nil {
		return err
	}
	client, err := realtime.NewClient(myURL, b.GetBool("debug"))
	b.c = client
	if err != nil {
		return err
	}
	restclient := rest.NewClient(myURL, b.GetBool("debug"))
	user, err := b.c.Login(credentials)
	if err != nil {
		return err
	}
	b.user = user
	b.r = restclient
	err = b.r.Login(credentials)
	if err != nil {
		return err
	}
	b.Log.Info("Connection succeeded")
	return nil
}

func (b *Brocketchat) getChannelName(id string) string {
	b.RLock()
	defer b.RUnlock()
	if name, ok := b.channelMap[id]; ok {
		return name
	}
	return ""
}

func (b *Brocketchat) getChannelID(name string) string {
	b.RLock()
	defer b.RUnlock()
	for k, v := range b.channelMap {
		if v == name || v == "#"+name {
			return k
		}
	}
	return ""
}

func (b *Brocketchat) skipMessage(message *models.Message) bool {
	return message.User.ID == b.user.ID
}

func (b *Brocketchat) uploadFile(fi *config.FileInfo, channel string) error {
	fb := gomf.New()
	if err := fb.WriteField("description", fi.Comment); err != nil {
		return err
	}
	sp := strings.Split(fi.Name, ".")
	mtype := mime.TypeByExtension("." + sp[len(sp)-1])
	if !strings.Contains(mtype, "image") && !strings.Contains(mtype, "video") {
		return nil
	}
	if err := fb.WriteFile("file", fi.Name, mtype, *fi.Data); err != nil {
		return err
	}
	req, err := fb.GetHTTPRequest(context.TODO(), b.GetString("server")+"/api/v1/rooms.upload/"+channel)
	if err != nil {
		return err
	}
	req.Header.Add("X-Auth-Token", b.user.Token)
	req.Header.Add("X-User-Id", b.user.ID)
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		b.Log.Errorf("failed: %#v", string(body))
	}
	return nil
}

// sendWebhook uses the configured WebhookURL to send the message
func (b *Brocketchat) sendWebhook(msg *config.Message) error {
	// skip events
	if msg.Event != "" {
		return nil
	}

	if b.GetBool("PrefixMessagesWithNick") {
		msg.Text = msg.Username + msg.Text
	}
	if msg.Extra != nil {
		// this sends a message only if we received a config.EVENT_FILE_FAILURE_SIZE
		for _, rmsg := range helper.HandleExtra(msg, b.General) {
			rmsg := rmsg // scopelint
			iconURL := config.GetIconURL(&rmsg, b.GetString("iconurl"))
			matterMessage := matterhook.OMessage{
				IconURL:  iconURL,
				Channel:  rmsg.Channel,
				UserName: rmsg.Username,
				Text:     rmsg.Text,
				Props:    make(map[string]interface{}),
			}
			if err := b.mh.Send(matterMessage); err != nil {
				b.Log.Errorf("sendWebhook failed: %s ", err)
			}
		}

		// webhook doesn't support file uploads, so we add the url manually
		if len(msg.Extra["file"]) > 0 {
			for _, f := range msg.Extra["file"] {
				fi := f.(config.FileInfo)
				if fi.URL != "" {
					msg.Text += fi.URL
				}
			}
		}
	}
	iconURL := config.GetIconURL(msg, b.GetString("iconurl"))
	matterMessage := matterhook.OMessage{
		IconURL:  iconURL,
		Channel:  msg.Channel,
		UserName: msg.Username,
		Text:     msg.Text,
	}
	if msg.Avatar != "" {
		matterMessage.IconURL = msg.Avatar
	}
	err := b.mh.Send(matterMessage)
	if err != nil {
		b.Log.Info(err)
		return err
	}
	return nil
}
