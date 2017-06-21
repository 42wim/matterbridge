package community

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

const cookiePath = "https://steamcommunity.com/"

func SetCookies(client *http.Client, sessionId, steamLogin, steamLoginSecure string) {
	if client.Jar == nil {
		client.Jar, _ = cookiejar.New(new(cookiejar.Options))
	}
	base, err := url.Parse(cookiePath)
	if err != nil {
		panic(err)
	}
	client.Jar.SetCookies(base, []*http.Cookie{
		// It seems that, for some reason, Steam tries to URL-decode the cookie.
		&http.Cookie{
			Name:  "sessionid",
			Value: url.QueryEscape(sessionId),
		},
		// steamLogin is already URL-encoded.
		&http.Cookie{
			Name:  "steamLogin",
			Value: steamLogin,
		},
		&http.Cookie{
			Name:  "steamLoginSecure",
			Value: steamLoginSecure,
		},
	})
}
