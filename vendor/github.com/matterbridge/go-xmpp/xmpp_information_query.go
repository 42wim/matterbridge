package xmpp

import (
	"fmt"
	"strconv"
)

const IQTypeGet = "get"
const IQTypeSet = "set"
const IQTypeResult = "result"

func (c *Client) Discovery() (string, error) {
	// use getCookie for a pseudo random id.
	reqID := strconv.FormatUint(uint64(getCookie()), 10)
	return c.RawInformationQuery(c.jid, c.domain, reqID, IQTypeGet, XMPPNS_DISCO_ITEMS, "")
}

// Discover information about a node
func (c *Client) DiscoverNodeInfo(node string) (string, error) {
	query := fmt.Sprintf("<query xmlns='%s' node='%s'/>", XMPPNS_DISCO_INFO, node)
	return c.RawInformation(c.jid, c.domain, "info3", IQTypeGet, query)
}

// Discover items that the server exposes
func (c *Client) DiscoverServerItems() (string, error) {
	return c.DiscoverEntityItems(c.domain)
}

// Discover items that an entity exposes
func (c *Client) DiscoverEntityItems(jid string) (string, error) {
	query := fmt.Sprintf("<query xmlns='%s'/>", XMPPNS_DISCO_ITEMS)
	return c.RawInformation(c.jid, jid, "info1", IQTypeGet, query)
}

// RawInformationQuery sends an information query request to the server.
func (c *Client) RawInformationQuery(from, to, id, iqType, requestNamespace, body string) (string, error) {
	const xmlIQ = "<iq from='%s' to='%s' id='%s' type='%s'><query xmlns='%s'>%s</query></iq>"
	_, err := fmt.Fprintf(c.conn, xmlIQ, xmlEscape(from), xmlEscape(to), id, iqType, requestNamespace, body)
	return id, err
}

// rawInformation send a IQ request with the payload body to the server
func (c *Client) RawInformation(from, to, id, iqType, body string) (string, error) {
	const xmlIQ = "<iq from='%s' to='%s' id='%s' type='%s'>%s</iq>"
	_, err := fmt.Fprintf(c.conn, xmlIQ, xmlEscape(from), xmlEscape(to), id, iqType, body)
	return id, err
}
