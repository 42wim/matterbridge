/*
Package longpoll implements Bots Long Poll API.

See more https://vk.com/dev/bots_longpoll
*/
package longpoll // import "github.com/SevereCloud/vksdk/v2/longpoll-bot"

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/SevereCloud/vksdk/v2"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/internal"
)

// Response struct.
type Response struct {
	Ts      string              `json:"ts"`
	Updates []events.GroupEvent `json:"updates"`
	Failed  int                 `json:"failed"`
}

// LongPoll struct.
type LongPoll struct {
	GroupID int
	Server  string
	Key     string
	Ts      string
	Wait    int
	VK      *api.VK
	Client  *http.Client
	cancel  context.CancelFunc

	funcFullResponseList []func(Response)

	events.FuncList
}

// NewLongPoll returns a new LongPoll.
//
// The LongPoll will use the http.DefaultClient.
// This means that if the http.DefaultClient is modified by other components
// of your application the modifications will be picked up by the SDK as well.
func NewLongPoll(vk *api.VK, groupID int) (*LongPoll, error) {
	lp := &LongPoll{
		VK:      vk,
		GroupID: groupID,
		Wait:    25,
		Client:  http.DefaultClient,
	}
	lp.FuncList = *events.NewFuncList()

	err := lp.updateServer(true)

	return lp, err
}

// NewLongPollCommunity returns a new LongPoll for community token.
//
// The LongPoll will use the http.DefaultClient.
// This means that if the http.DefaultClient is modified by other components
// of your application the modifications will be picked up by the SDK as well.
func NewLongPollCommunity(vk *api.VK) (*LongPoll, error) {
	resp, err := vk.GroupsGetByID(nil)
	if err != nil {
		return nil, err
	}

	lp := &LongPoll{
		VK:      vk,
		GroupID: resp[0].ID,
		Wait:    25,
		Client:  http.DefaultClient,
	}
	lp.FuncList = *events.NewFuncList()

	err = lp.updateServer(true)

	return lp, err
}

func (lp *LongPoll) updateServer(updateTs bool) error {
	params := api.Params{
		"group_id": lp.GroupID,
	}

	serverSetting, err := lp.VK.GroupsGetLongPollServer(params)
	if err != nil {
		return err
	}

	lp.Key = serverSetting.Key
	lp.Server = serverSetting.Server

	if updateTs {
		lp.Ts = serverSetting.Ts
	}

	return nil
}

func (lp *LongPoll) check(ctx context.Context) (response Response, err error) {
	u := fmt.Sprintf("%s?act=a_check&key=%s&ts=%s&wait=%d", lp.Server, lp.Key, lp.Ts, lp.Wait)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return response, err
	}

	resp, err := lp.Client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	response, err = parseResponse(resp.Body)
	if err != nil {
		return response, err
	}

	err = lp.checkResponse(response)

	return response, err
}

func parseResponse(reader io.Reader) (response Response, err error) {
	decoder := json.NewDecoder(reader)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return response, err
		}

		t, ok := token.(string)
		if !ok {
			continue
		}

		switch t {
		case "failed":
			raw, err := decoder.Token()
			if err != nil {
				return response, err
			}

			response.Failed = int(raw.(float64))
		case "updates":
			var updates []events.GroupEvent

			err = decoder.Decode(&updates)
			if err != nil {
				return response, err
			}

			response.Updates = updates
		case "ts":
			// can be a number in the response with "failed" field: {"ts":8,"failed":1}
			// or string, e.g. {"ts":"8","updates":[]}
			rawTs, err := decoder.Token()
			if err != nil {
				return response, err
			}

			if ts, isNumber := rawTs.(float64); isNumber {
				response.Ts = strconv.Itoa(int(ts))
			} else {
				response.Ts = rawTs.(string)
			}
		}
	}

	return response, err
}

func (lp *LongPoll) checkResponse(response Response) (err error) {
	switch response.Failed {
	case 0:
		lp.Ts = response.Ts
	case 1:
		lp.Ts = response.Ts
	case 2:
		err = lp.updateServer(false)
	case 3:
		err = lp.updateServer(true)
	default:
		err = &Failed{response.Failed}
	}

	return
}

func (lp *LongPoll) autoSetting(ctx context.Context) error {
	params := api.Params{
		"group_id":    lp.GroupID,
		"enabled":     true,
		"api_version": vksdk.API,
	}.WithContext(ctx)
	for _, event := range lp.ListEvents() {
		params[string(event)] = true
	}

	// Updating LongPoll settings
	_, err := lp.VK.GroupsSetLongPollSettings(params)

	return err
}

// Run handler.
func (lp *LongPoll) Run() error {
	return lp.RunWithContext(context.Background())
}

// RunWithContext handler.
func (lp *LongPoll) RunWithContext(ctx context.Context) error {
	return lp.run(ctx)
}

func (lp *LongPoll) run(ctx context.Context) error {
	ctx, lp.cancel = context.WithCancel(ctx)

	if err := lp.autoSetting(ctx); err != nil {
		return err
	}

	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				return nil
			}
		default:
			resp, err := lp.check(ctx)
			if err != nil {
				return err
			}

			ctx = context.WithValue(ctx, internal.LongPollTsKey, resp.Ts)

			for _, event := range resp.Updates {
				err = lp.Handler(ctx, event)
				if err != nil {
					return err
				}
			}

			for _, f := range lp.funcFullResponseList {
				f(resp)
			}
		}
	}
}

// Shutdown gracefully shuts down the longpoll without interrupting any active connections.
func (lp *LongPoll) Shutdown() {
	if lp.cancel != nil {
		lp.cancel()
	}
}

// FullResponse handler.
func (lp *LongPoll) FullResponse(f func(Response)) {
	lp.funcFullResponseList = append(lp.funcFullResponseList, f)
}
