package xmpp

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"strconv"
)

const (
	XMPPNS_AVATAR_PEP_DATA     = "urn:xmpp:avatar:data"
	XMPPNS_AVATAR_PEP_METADATA = "urn:xmpp:avatar:metadata"
)

type clientAvatarData struct {
	XMLName xml.Name `xml:"data"`
	Data    []byte   `xml:",innerxml"`
}

type clientAvatarInfo struct {
	XMLName xml.Name `xml:"info"`
	Bytes   string   `xml:"bytes,attr"`
	Width   string   `xml:"width,attr"`
	Height  string   `xml:"height,attr"`
	ID      string   `xml:"id,attr"`
	Type    string   `xml:"type,attr"`
	URL     string   `xml:"url,attr"`
}

type clientAvatarMetadata struct {
	XMLName xml.Name         `xml:"metadata"`
	XMLNS   string           `xml:"xmlns,attr"`
	Info    clientAvatarInfo `xml:"info"`
}

type AvatarData struct {
	Data []byte
	From string
}

type AvatarMetadata struct {
	From   string
	Bytes  int
	Width  int
	Height int
	ID     string
	Type   string
	URL    string
}

func handleAvatarData(itemsBody []byte, from, id string) (AvatarData, error) {
	var data clientAvatarData
	err := xml.Unmarshal(itemsBody, &data)
	if err != nil {
		return AvatarData{}, err
	}

	// Base64-decode the avatar data to check its SHA1 hash
	dataRaw, err := base64.StdEncoding.DecodeString(
		string(data.Data))
	if err != nil {
		return AvatarData{}, err
	}

	hash := sha1.Sum(dataRaw)
	hashStr := hex.EncodeToString(hash[:])
	if hashStr != id {
		return AvatarData{}, errors.New("SHA1 hashes do not match")
	}

	return AvatarData{
		Data: dataRaw,
		From: from,
	}, nil
}

func handleAvatarMetadata(body []byte, from string) (AvatarMetadata, error) {
	var meta clientAvatarMetadata
	err := xml.Unmarshal(body, &meta)
	if err != nil {
		return AvatarMetadata{}, err
	}

	return AvatarMetadata{
		From:   from,
		Bytes:  atoiw(meta.Info.Bytes),
		Width:  atoiw(meta.Info.Width),
		Height: atoiw(meta.Info.Height),
		ID:     meta.Info.ID,
		Type:   meta.Info.Type,
		URL:    meta.Info.URL,
	}, nil
}

// A wrapper for atoi which just returns -1 if an error occurs
func atoiw(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		return -1
	}

	return i
}

func (c *Client) AvatarSubscribeMetadata(jid string) {
	c.PubsubSubscribeNode(XMPPNS_AVATAR_PEP_METADATA, jid)
}

func (c *Client) AvatarUnsubscribeMetadata(jid string) {
	c.PubsubUnsubscribeNode(XMPPNS_AVATAR_PEP_METADATA, jid)
}

func (c *Client) AvatarRequestData(jid string) {
	c.PubsubRequestLastItems(XMPPNS_AVATAR_PEP_DATA, jid)
}

func (c *Client) AvatarRequestDataByID(jid, id string) {
	c.PubsubRequestItem(XMPPNS_AVATAR_PEP_DATA, jid, id)
}

func (c *Client) AvatarRequestMetadata(jid string) {
	c.PubsubRequestLastItems(XMPPNS_AVATAR_PEP_METADATA, jid)
}
