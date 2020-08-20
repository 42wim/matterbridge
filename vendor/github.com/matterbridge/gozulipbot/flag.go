package gozulipbot

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func (b *Bot) GetConfigFromFlags() error {
	var (
		apiKey  = flag.String("apikey", "ZULIP_APIKEY", "bot api key or env var")
		apiURL  = flag.String("apiurl", "ZULIP_APIURL", "url of zulip server or env var")
		backoff = flag.Duration("backoff", 1*time.Second, "backoff base duration or env var")
		email   = flag.String("email", "ZULIP_EMAIL", "bot email address or env var")
		env     = flag.Bool("env", false, "get string values from environment variables")
	)
	flag.Parse()

	b.APIKey = *apiKey
	b.APIURL = *apiURL
	b.Email = *email
	b.Backoff = *backoff
	if *env {
		b.GetConfigFromEnvironment()
	}
	return b.checkConfig()
}

func (b *Bot) GetConfigFromEnvironment() error {
	if apiKey, exists := os.LookupEnv(b.APIKey); !exists {
		return fmt.Errorf("--env was set but env var %s did not exist", b.APIKey)
	} else {
		b.APIKey = apiKey
	}
	if apiURL, exists := os.LookupEnv(b.APIURL); !exists {
		return fmt.Errorf("--env was set but env var %s did not exist", b.APIURL)
	} else {
		b.APIURL = apiURL
	}
	if email, exists := os.LookupEnv(b.Email); !exists {
		return fmt.Errorf("--env was set but env var %s did not exist", b.Email)
	} else {
		b.Email = email
	}
	return nil
}

func (b *Bot) checkConfig() error {
	if b.APIKey == "" {
		return fmt.Errorf("--apikey is required")
	}
	if b.APIURL == "" {
		return fmt.Errorf("--apiurl is required")
	}
	if b.Email == "" {
		return fmt.Errorf("--email is required")
	}
	if b.Backoff <= 0 {
		return fmt.Errorf("--backoff must be greater than zero")
	}
	return nil
}
