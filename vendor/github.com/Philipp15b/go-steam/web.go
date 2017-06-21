package steam

import (
	"crypto/aes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/Philipp15b/go-steam/cryptoutil"
	. "github.com/Philipp15b/go-steam/protocol"
	. "github.com/Philipp15b/go-steam/protocol/protobuf"
	. "github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/golang/protobuf/proto"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
)

type Web struct {
	// 64 bit alignment
	relogOnNonce uint32

	// The `sessionid` cookie required to use the steam website.
	// This cookie may contain a characters that will need to be URL-escaped, otherwise
	// Steam (probably) interprets is as a string.
	// When used as an URL paramter this is automatically escaped by the Go HTTP package.
	SessionId string
	// The `steamLogin` cookie required to use the steam website. Already URL-escaped.
	// This is only available after calling LogOn().
	SteamLogin string
	// The `steamLoginSecure` cookie required to use the steam website over HTTPs. Already URL-escaped.
	// This is only availbile after calling LogOn().
	SteamLoginSecure string

	webLoginKey string

	client *Client
}

func (w *Web) HandlePacket(packet *Packet) {
	switch packet.EMsg {
	case EMsg_ClientNewLoginKey:
		w.handleNewLoginKey(packet)
	case EMsg_ClientRequestWebAPIAuthenticateUserNonceResponse:
		w.handleAuthNonceResponse(packet)
	}
}

// Fetches the `steamLogin` cookie. This may only be called after the first
// WebSessionIdEvent or it will panic.
func (w *Web) LogOn() {
	if w.webLoginKey == "" {
		panic("Web: webLoginKey not initialized!")
	}

	go func() {
		// retry three times. yes, I know about loops.
		err := w.apiLogOn()
		if err != nil {
			err = w.apiLogOn()
			if err != nil {
				err = w.apiLogOn()
			}
		}
		if err != nil {
			w.client.Emit(WebLogOnErrorEvent(err))
			return
		}
	}()
}

func (w *Web) apiLogOn() error {
	sessionKey := make([]byte, 32)
	rand.Read(sessionKey)

	cryptedSessionKey := cryptoutil.RSAEncrypt(GetPublicKey(EUniverse_Public), sessionKey)
	ciph, _ := aes.NewCipher(sessionKey)
	cryptedLoginKey := cryptoutil.SymmetricEncrypt(ciph, []byte(w.webLoginKey))
	data := make(url.Values)
	data.Add("format", "json")
	data.Add("steamid", strconv.FormatUint(w.client.SteamId().ToUint64(), 10))
	data.Add("sessionkey", string(cryptedSessionKey))
	data.Add("encrypted_loginkey", string(cryptedLoginKey))
	resp, err := http.PostForm("https://api.steampowered.com/ISteamUserAuth/AuthenticateUser/v0001", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		// our web login key has expired, request a new one
		atomic.StoreUint32(&w.relogOnNonce, 1)
		w.client.Write(NewClientMsgProtobuf(EMsg_ClientRequestWebAPIAuthenticateUserNonce, new(CMsgClientRequestWebAPIAuthenticateUserNonce)))
		return nil
	} else if resp.StatusCode != 200 {
		return errors.New("steam.Web.apiLogOn: request failed with status " + resp.Status)
	}

	result := new(struct {
		Authenticateuser struct {
			Token       string
			TokenSecure string
		}
	})
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return err
	}

	w.SteamLogin = result.Authenticateuser.Token
	w.SteamLoginSecure = result.Authenticateuser.TokenSecure

	w.client.Emit(new(WebLoggedOnEvent))
	return nil
}

func (w *Web) handleNewLoginKey(packet *Packet) {
	msg := new(CMsgClientNewLoginKey)
	packet.ReadProtoMsg(msg)

	w.client.Write(NewClientMsgProtobuf(EMsg_ClientNewLoginKeyAccepted, &CMsgClientNewLoginKeyAccepted{
		UniqueId: proto.Uint32(msg.GetUniqueId()),
	}))

	// number -> string -> bytes -> base64
	w.SessionId = base64.StdEncoding.EncodeToString([]byte(strconv.FormatUint(uint64(msg.GetUniqueId()), 10)))

	w.client.Emit(new(WebSessionIdEvent))
}

func (w *Web) handleAuthNonceResponse(packet *Packet) {
	// this has to be the best name for a message yet.
	msg := new(CMsgClientRequestWebAPIAuthenticateUserNonceResponse)
	packet.ReadProtoMsg(msg)
	w.webLoginKey = msg.GetWebapiAuthenticateUserNonce()

	// if the nonce was specifically requested in apiLogOn(),
	// don't emit an event.
	if atomic.CompareAndSwapUint32(&w.relogOnNonce, 1, 0) {
		w.LogOn()
	}
}
