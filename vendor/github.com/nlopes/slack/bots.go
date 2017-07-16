package slack

import (
	"context"
	"errors"
	"net/url"
)

// Bot contains information about a bot
type Bot struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Deleted bool   `json:"deleted"`
	Icons   Icons  `json:"icons"`
}

type botResponseFull struct {
	Bot `json:"bot,omitempty"` // GetBotInfo
	SlackResponse
}

func botRequest(ctx context.Context, path string, values url.Values, debug bool) (*botResponseFull, error) {
	response := &botResponseFull{}
	err := post(ctx, path, values, response, debug)
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, errors.New(response.Error)
	}
	return response, nil
}

// GetBotInfo will retrieve the complete bot information
func (api *Client) GetBotInfo(bot string) (*Bot, error) {
	return api.GetBotInfoContext(context.Background(), bot)
}

// GetBotInfoContext will retrieve the complete bot information using a custom context
func (api *Client) GetBotInfoContext(ctx context.Context, bot string) (*Bot, error) {
	values := url.Values{
		"token": {api.config.token},
		"bot":   {bot},
	}
	response, err := botRequest(ctx, "bots.info", values, api.debug)
	if err != nil {
		return nil, err
	}
	return &response.Bot, nil
}
