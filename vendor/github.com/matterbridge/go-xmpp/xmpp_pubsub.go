package xmpp

import (
	"encoding/xml"
	"fmt"
)

const (
	XMPPNS_PUBSUB       = "http://jabber.org/protocol/pubsub"
	XMPPNS_PUBSUB_EVENT = "http://jabber.org/protocol/pubsub#event"
)

type clientPubsubItem struct {
	XMLName xml.Name `xml:"item"`
	ID      string   `xml:"id,attr"`
	Body    []byte   `xml:",innerxml"`
}

type clientPubsubItems struct {
	XMLName xml.Name           `xml:"items"`
	Node    string             `xml:"node,attr"`
	Items   []clientPubsubItem `xml:"item"`
}

type clientPubsub struct {
	XMLName xml.Name          `xml:"pubsub"`
	Items   clientPubsubItems `xml:"items"`
}

type clientPubsubEvent struct {
	XMLName xml.Name          `xml:"event"`
	XMLNS   string            `xml:"xmlns,attr"`
	Items   clientPubsubItems `xml:"items"`
}

type clientPubsubError struct {
	XMLName xml.Name
}

type clientPubsubSubscription struct {
	XMLName xml.Name `xml:"subscription"`
	Node    string   `xml:"node,attr"`
	JID     string   `xml:"jid,attr"`
	SubID   string   `xml:"subid,attr"`
}

type PubsubEvent struct {
	Node  string
	Items []PubsubItem
}

type PubsubSubscription struct {
	SubID  string
	JID    string
	Node   string
	Errors []string
}
type PubsubUnsubscription PubsubSubscription

type PubsubItem struct {
	ID       string
	InnerXML []byte
}

type PubsubItems struct {
	Node  string
	Items []PubsubItem
}

// Converts []clientPubsubItem to []PubsubItem
func pubsubItemsToReturn(items []clientPubsubItem) []PubsubItem {
	var tmp []PubsubItem
	for _, i := range items {
		tmp = append(tmp, PubsubItem{
			ID:       i.ID,
			InnerXML: i.Body,
		})
	}

	return tmp
}

func pubsubClientToReturn(event clientPubsubEvent) PubsubEvent {
	return PubsubEvent{
		Node:  event.Items.Node,
		Items: pubsubItemsToReturn(event.Items.Items),
	}
}

func pubsubStanza(body string) string {
	return fmt.Sprintf("<pubsub xmlns='%s'>%s</pubsub>",
		XMPPNS_PUBSUB, body)
}

func pubsubSubscriptionStanza(node, jid string) string {
	body := fmt.Sprintf("<subscribe node='%s' jid='%s'/>",
		xmlEscape(node),
		xmlEscape(jid))
	return pubsubStanza(body)
}

func pubsubUnsubscriptionStanza(node, jid string) string {
	body := fmt.Sprintf("<unsubscribe node='%s' jid='%s'/>",
		xmlEscape(node),
		xmlEscape(jid))
	return pubsubStanza(body)
}

func (c *Client) PubsubSubscribeNode(node, jid string) {
	c.RawInformation(c.jid,
		jid,
		"sub1",
		"set",
		pubsubSubscriptionStanza(node, c.jid))
}

func (c *Client) PubsubUnsubscribeNode(node, jid string) {
	c.RawInformation(c.jid,
		jid,
		"unsub1",
		"set",
		pubsubUnsubscriptionStanza(node, c.jid))
}

func (c *Client) PubsubRequestLastItems(node, jid string) {
	body := fmt.Sprintf("<items node='%s'/>", node)
	c.RawInformation(c.jid, jid, "items1", "get", pubsubStanza(body))
}

func (c *Client) PubsubRequestItem(node, jid, id string) {
	body := fmt.Sprintf("<items node='%s'><item id='%s'/></items>", node, id)
	c.RawInformation(c.jid, jid, "items3", "get", pubsubStanza(body))
}
