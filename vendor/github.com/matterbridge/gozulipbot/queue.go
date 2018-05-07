package gozulipbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	HeartbeatError    = fmt.Errorf("EventMessage is a heartbeat")
	UnauthorizedError = fmt.Errorf("Request is unauthorized")
	BackoffError      = fmt.Errorf("Too many requests")
	UnknownError      = fmt.Errorf("Error was unknown")
)

type Queue struct {
	ID           string `json:"queue_id"`
	LastEventID  int    `json:"last_event_id"`
	MaxMessageID int    `json:"max_message_id"`
	Bot          *Bot   `json:"-"`
}

func (q *Queue) EventsChan() (chan EventMessage, func()) {
	end := false
	endFunc := func() {
		end = true
	}

	out := make(chan EventMessage)
	go func() {
		defer close(out)
		for {
			backoffTime := time.Now().Add(q.Bot.Backoff * time.Duration(math.Pow10(int(atomic.LoadInt64(&q.Bot.Retries)))))
			minTime := time.Now().Add(q.Bot.Backoff)
			if end {
				return
			}
			ems, err := q.GetEvents()
			switch {
			case err == HeartbeatError:
				time.Sleep(time.Until(minTime))
				continue
			case err == BackoffError:
				time.Sleep(time.Until(backoffTime))
				atomic.AddInt64(&q.Bot.Retries, 1)
			case err == UnauthorizedError:
				// TODO? have error channel when ending the continuously running process?
				return
			default:
				atomic.StoreInt64(&q.Bot.Retries, 0)
			}
			if err != nil {
				// TODO: handle unknown error
				// For now, handle this like an UnauthorizedError and end the func.
				return
			}
			for _, em := range ems {
				out <- em
			}
			// Always make sure we wait the minimum time before asking again.
			time.Sleep(time.Until(minTime))
		}
	}()

	return out, endFunc
}

// EventsCallback will repeatedly call provided callback function with
// the output of continual queue.GetEvents calls.
// It returns a function which can be called to end the calls.
//
// It will end early if it receives an UnauthorizedError, or an unknown error.
// Note, it will never return a HeartbeatError.
func (q *Queue) EventsCallback(fn func(EventMessage, error)) func() {
	end := false
	endFunc := func() {
		end = true
	}
	go func() {
		for {
			backoffTime := time.Now().Add(q.Bot.Backoff * time.Duration(math.Pow10(int(atomic.LoadInt64(&q.Bot.Retries)))))
			minTime := time.Now().Add(q.Bot.Backoff)
			if end {
				return
			}
			ems, err := q.GetEvents()
			switch {
			case err == HeartbeatError:
				time.Sleep(time.Until(minTime))
				continue
			case err == BackoffError:
				time.Sleep(time.Until(backoffTime))
				atomic.AddInt64(&q.Bot.Retries, 1)
			case err == UnauthorizedError:
				// TODO? have error channel when ending the continuously running process?
				return
			default:
				atomic.StoreInt64(&q.Bot.Retries, 0)
			}
			if err != nil {
				// TODO: handle unknown error
				// For now, handle this like an UnauthorizedError and end the func.
				return
			}
			for _, em := range ems {
				fn(em, err)
			}
			// Always make sure we wait the minimum time before asking again.
			time.Sleep(time.Until(minTime))
		}
	}()

	return endFunc
}

// GetEvents is a blocking call that waits for and parses a list of EventMessages.
// There will usually only be one EventMessage returned.
// When a heartbeat is returned, GetEvents will return a HeartbeatError.
// When an http status code above 400 is returned, one of a BackoffError,
// UnauthorizedError, or UnknownError will be returned.
func (q *Queue) GetEvents() ([]EventMessage, error) {
	resp, err := q.RawGetEvents()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == 429:
		return nil, BackoffError
	case resp.StatusCode == 403:
		return nil, UnauthorizedError
	case resp.StatusCode >= 400:
		return nil, UnknownError
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	msgs, err := q.ParseEventMessages(body)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

// RawGetEvents is a blocking call that receives a response containing a list
// of events (a.k.a. received messages) since the last message id in the queue.
func (q *Queue) RawGetEvents() (*http.Response, error) {
	values := url.Values{}
	values.Set("queue_id", q.ID)
	values.Set("last_event_id", strconv.Itoa(q.LastEventID))

	url := "events?" + values.Encode()

	req, err := q.Bot.constructRequest("GET", url, "")
	if err != nil {
		return nil, err
	}

	return q.Bot.Client.Do(req)
}

func (q *Queue) ParseEventMessages(rawEventResponse []byte) ([]EventMessage, error) {
	rawResponse := map[string]json.RawMessage{}
	err := json.Unmarshal(rawEventResponse, &rawResponse)
	if err != nil {
		return nil, err
	}

	events := []map[string]json.RawMessage{}
	err = json.Unmarshal(rawResponse["events"], &events)
	if err != nil {
		return nil, err
	}

	messages := []EventMessage{}
	for _, event := range events {
		// if the event is a heartbeat, return a special error
		if string(event["type"]) == `"heartbeat"` {
			return nil, HeartbeatError
		}
		var msg EventMessage
		err = json.Unmarshal(event["message"], &msg)
		// TODO? should this check be here
		if err != nil {
			return nil, err
		}
		msg.Queue = q
		messages = append(messages, msg)
	}

	return messages, nil
}
