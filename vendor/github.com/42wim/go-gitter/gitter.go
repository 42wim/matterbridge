// Package gitter is a Go client library for the Gitter API.
//
// Author: sromku
package gitter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/mreiferson/go-httpclient"
)

var (
	apiBaseURL    = "https://api.gitter.im/v1/"
	streamBaseURL = "https://stream.gitter.im/v1/"
	fayeBaseURL   = "https://ws.gitter.im/faye"
)

type Gitter struct {
	config struct {
		apiBaseURL    string
		streamBaseURL string
		token         string
		client        *http.Client
	}
	debug     bool
	logWriter io.Writer
}

// New initializes the Gitter API client
//
// For example:
//  api := gitter.New("YOUR_ACCESS_TOKEN")
func New(token string) *Gitter {

	transport := &httpclient.Transport{
		ConnectTimeout:   5 * time.Second,
		ReadWriteTimeout: 40 * time.Second,
	}
	defer transport.Close()

	s := &Gitter{}
	s.config.apiBaseURL = apiBaseURL
	s.config.streamBaseURL = streamBaseURL
	s.config.token = token
	s.config.client = &http.Client{
		Transport: transport,
	}
	return s
}

// SetClient sets a custom http client. Can be useful in App Engine case.
func (gitter *Gitter) SetClient(client *http.Client) {
	gitter.config.client = client
}

// GetUser returns the current user
func (gitter *Gitter) GetUser() (*User, error) {

	var users []User
	response, err := gitter.get(gitter.config.apiBaseURL + "user")
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	err = json.Unmarshal(response, &users)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	if len(users) > 0 {
		return &users[0], nil
	}

	err = APIError{What: "Failed to retrieve current user"}
	gitter.log(err)
	return nil, err
}

// GetUserRooms returns a list of Rooms the user is part of
func (gitter *Gitter) GetUserRooms(userID string) ([]Room, error) {

	var rooms []Room
	response, err := gitter.get(gitter.config.apiBaseURL + "user/" + userID + "/rooms")
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	err = json.Unmarshal(response, &rooms)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	return rooms, nil
}

// GetRooms returns a list of rooms the current user is in
func (gitter *Gitter) GetRooms() ([]Room, error) {

	var rooms []Room
	response, err := gitter.get(gitter.config.apiBaseURL + "rooms")
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	err = json.Unmarshal(response, &rooms)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	return rooms, nil
}

// GetUsersInRoom returns the users in the room with the passed id
func (gitter *Gitter) GetUsersInRoom(roomID string) ([]User, error) {
	var users []User
	response, err := gitter.get(gitter.config.apiBaseURL + "rooms/" + roomID + "/users")
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	err = json.Unmarshal(response, &users)
	if err != nil {
		gitter.log(err)
		return nil, err
	}
	return users, nil
}

// GetRoom returns a room with the passed id
func (gitter *Gitter) GetRoom(roomID string) (*Room, error) {

	var room Room
	response, err := gitter.get(gitter.config.apiBaseURL + "rooms/" + roomID)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	err = json.Unmarshal(response, &room)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	return &room, nil
}

// GetMessages returns a list of messages in a room.
// Pagination is optional. You can pass nil or specific pagination params.
func (gitter *Gitter) GetMessages(roomID string, params *Pagination) ([]Message, error) {

	var messages []Message
	url := gitter.config.apiBaseURL + "rooms/" + roomID + "/chatMessages"
	if params != nil {
		url += "?" + params.encode()
	}
	response, err := gitter.get(url)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	err = json.Unmarshal(response, &messages)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	return messages, nil
}

// GetMessage returns a message in a room.
func (gitter *Gitter) GetMessage(roomID, messageID string) (*Message, error) {

	var message Message
	response, err := gitter.get(gitter.config.apiBaseURL + "rooms/" + roomID + "/chatMessages/" + messageID)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	err = json.Unmarshal(response, &message)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	return &message, nil
}

// SendMessage sends a message to a room
func (gitter *Gitter) SendMessage(roomID, text string) (*Message, error) {

	message := Message{Text: text}
	body, _ := json.Marshal(message)
	response, err := gitter.post(gitter.config.apiBaseURL+"rooms/"+roomID+"/chatMessages", body)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	err = json.Unmarshal(response, &message)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	return &message, nil
}

// UpdateMessage updates a message in a room
func (gitter *Gitter) UpdateMessage(roomID, msgID, text string) (*Message, error) {

	message := Message{Text: text}
	body, _ := json.Marshal(message)
	response, err := gitter.put(gitter.config.apiBaseURL+"rooms/"+roomID+"/chatMessages/"+msgID, body)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	err = json.Unmarshal(response, &message)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	return &message, nil
}

// JoinRoom joins a room
func (gitter *Gitter) JoinRoom(roomID, userID string) (*Room, error) {

	message := Room{ID: roomID}
	body, _ := json.Marshal(message)
	response, err := gitter.post(gitter.config.apiBaseURL+"user/"+userID+"/rooms", body)

	if err != nil {
		gitter.log(err)
		return nil, err
	}

	var room Room
	err = json.Unmarshal(response, &room)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	return &room, nil
}

// LeaveRoom removes a user from the room
func (gitter *Gitter) LeaveRoom(roomID, userID string) error {

	_, err := gitter.delete(gitter.config.apiBaseURL + "rooms/" + roomID + "/users/" + userID)
	if err != nil {
		gitter.log(err)
		return err
	}

	return nil
}

// SetDebug traces errors if it's set to true.
func (gitter *Gitter) SetDebug(debug bool, logWriter io.Writer) {
	gitter.debug = debug
	gitter.logWriter = logWriter
}

// SearchRooms queries the Rooms resources of gitter API
func (gitter *Gitter) SearchRooms(room string) ([]Room, error) {

	var rooms struct {
		Results []Room `json:"results"`
	}

	response, err := gitter.get(gitter.config.apiBaseURL + "rooms?q=" + room)

	if err != nil {
		gitter.log(err)
		return nil, err
	}

	err = json.Unmarshal(response, &rooms)
	if err != nil {
		gitter.log(err)
		return nil, err
	}
	return rooms.Results, nil
}

// GetRoomId returns the room ID of a given URI
func (gitter *Gitter) GetRoomId(uri string) (string, error) {

	rooms, err := gitter.SearchRooms(uri)
	if err != nil {
		gitter.log(err)
		return "", err
	}

	for _, element := range rooms {
		if element.URI == uri {
			return element.ID, nil
		}
	}
	return "", APIError{What: "Room not found."}
}

// Pagination params
type Pagination struct {

	// Skip n messages
	Skip int

	// Get messages before beforeId
	BeforeID string

	// Get messages after afterId
	AfterID string

	// Maximum number of messages to return
	Limit int

	// Search query
	Query string
}

func (messageParams *Pagination) encode() string {
	values := url.Values{}

	if messageParams.AfterID != "" {
		values.Add("afterId", messageParams.AfterID)
	}

	if messageParams.BeforeID != "" {
		values.Add("beforeId", messageParams.BeforeID)
	}

	if messageParams.Skip > 0 {
		values.Add("skip", strconv.Itoa(messageParams.Skip))
	}

	if messageParams.Limit > 0 {
		values.Add("limit", strconv.Itoa(messageParams.Limit))
	}

	return values.Encode()
}

func (gitter *Gitter) getResponse(url string, stream *Stream) (*http.Response, error) {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		gitter.log(err)
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+gitter.config.token)
	if stream != nil {
		stream.streamConnection.request = r
	}
	response, err := gitter.config.client.Do(r)
	if err != nil {
		gitter.log(err)
		return nil, err
	}
	return response, nil
}

func (gitter *Gitter) get(url string) ([]byte, error) {
	resp, err := gitter.getResponse(url, nil)
	if err != nil {
		gitter.log(err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = APIError{What: fmt.Sprintf("Status code: %v", resp.StatusCode)}
		gitter.log(err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	return body, nil
}

func (gitter *Gitter) post(url string, body []byte) ([]byte, error) {
	r, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+gitter.config.token)

	resp, err := gitter.config.client.Do(r)
	if err != nil {
		gitter.log(err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = APIError{What: fmt.Sprintf("Status code: %v", resp.StatusCode)}
		gitter.log(err)
		return nil, err
	}

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	return result, nil
}

func (gitter *Gitter) put(url string, body []byte) ([]byte, error) {
	r, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+gitter.config.token)

	resp, err := gitter.config.client.Do(r)
	if err != nil {
		gitter.log(err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = APIError{What: fmt.Sprintf("Status code: %v", resp.StatusCode)}
		gitter.log(err)
		return nil, err
	}

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	return result, nil
}

func (gitter *Gitter) delete(url string) ([]byte, error) {
	r, err := http.NewRequest("delete", url, nil)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+gitter.config.token)

	resp, err := gitter.config.client.Do(r)
	if err != nil {
		gitter.log(err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = APIError{What: fmt.Sprintf("Status code: %v", resp.StatusCode)}
		gitter.log(err)
		return nil, err
	}

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		gitter.log(err)
		return nil, err
	}

	return result, nil
}

func (gitter *Gitter) log(a interface{}) {
	if gitter.debug {
		log.Println(a)
		if gitter.logWriter != nil {
			timestamp := time.Now().Format(time.RFC3339)
			msg := fmt.Sprintf("%v: %v", timestamp, a)
			fmt.Fprintln(gitter.logWriter, msg)
		}
	}
}

// APIError holds data of errors returned from the API.
type APIError struct {
	What string
}

func (e APIError) Error() string {
	return fmt.Sprintf("%v", e.What)
}
