package gozulipbot

import (
	"flag"
	"fmt"
	"time"
)

func (b *Bot) GetConfigFromFlags() error {
	var (
		apiKey  = flag.String("apikey", "", "bot api key")
		apiURL  = flag.String("apiurl", "", "url of zulip server")
		email   = flag.String("email", "", "bot email address")
		backoff = flag.Duration("backoff", 1*time.Second, "backoff base duration")
	)
	flag.Parse()

	if *apiKey == "" {
		return fmt.Errorf("--apikey is required")
	}
	if *apiURL == "" {
		return fmt.Errorf("--apiurl is required")
	}
	if *email == "" {
		return fmt.Errorf("--email is required")
	}
	b.APIKey = *apiKey
	b.APIURL = *apiURL
	b.Email = *email
	b.Backoff = *backoff
	return nil
}
