// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO(rsc):
//	More precise error handling.
//	Presence functionality.
// TODO(mattn):
//  Add proxy authentication.

// Package xmpp implements a simple Google Talk client
// using the XMPP protocol described in RFC 3920 and RFC 3921.
package xmpp

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"hash"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/net/proxy"
)

const (
	nsStream       = "http://etherx.jabber.org/streams"
	nsTLS          = "urn:ietf:params:xml:ns:xmpp-tls"
	nsSASL         = "urn:ietf:params:xml:ns:xmpp-sasl"
	nsSASL2        = "urn:xmpp:sasl:2"
	nsBind         = "urn:ietf:params:xml:ns:xmpp-bind"
	nsBind2        = "urn:xmpp:bind:0"
	nsFast         = "urn:xmpp:fast:0"
	nsSASLCB       = "urn:xmpp:sasl-cb:0"
	nsClient       = "jabber:client"
	nsSession      = "urn:ietf:params:xml:ns:xmpp-session"
	nsStreamLimits = "urn:xmpp:stream-limits:0"
)

// Default TLS configuration options
var DefaultConfig = &tls.Config{}

// DebugWriter is the writer used to write debugging output to.
var DebugWriter io.Writer = os.Stderr

// Cookie is a unique XMPP session identifier
type Cookie uint64

func getCookie() Cookie {
	var buf [8]byte
	if _, err := rand.Reader.Read(buf[:]); err != nil {
		panic("Failed to read random bytes: " + err.Error())
	}
	return Cookie(binary.LittleEndian.Uint64(buf[:]))
}

// Fast holds the XEP-0484 fast token, mechanism and expiry date
type Fast struct {
	Token     string
	Mechanism string
	Expiry    time.Time
}

// Client holds XMPP connection options
type Client struct {
	conn             net.Conn // connection to server
	jid              string   // Jabber ID for our connection
	domain           string
	nextMutex        sync.Mutex // Mutex to prevent multiple access to xml.Decoder
	shutdown         bool       // Variable signalling that the stream will be closed
	p                *xml.Decoder
	stanzaWriter     io.Writer
	LimitMaxBytes    int    // Maximum stanza size (XEP-0478: Stream Limits Advertisement)
	LimitIdleSeconds int    // Maximum idle seconds (XEP-0478: Stream Limits Advertisement)
	Mechanism        string // SCRAM mechanism used.
	Fast             Fast   // XEP-0484 FAST Token, mechanism and expiry.
}

func (c *Client) JID() string {
	return c.jid
}

func containsIgnoreCase(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}

func connect(host, user, passwd string, timeout time.Duration) (net.Conn, error) {
	addr := host

	if strings.TrimSpace(host) == "" {
		a := strings.SplitN(user, "@", 2)
		if len(a) == 2 {
			addr = a[1]
		}
	}
	a := strings.SplitN(host, ":", 2)
	if len(a) == 1 {
		addr += ":5222"
	}

	http_proxy := os.Getenv("HTTP_PROXY")
	if http_proxy == "" {
		http_proxy = os.Getenv("http_proxy")
	}
	// test for no proxy, takes a comma separated list with substrings to match
	if http_proxy != "" {
		noproxy := os.Getenv("NO_PROXY")
		if noproxy == "" {
			noproxy = os.Getenv("no_proxy")
		}
		if noproxy != "" {
			nplist := strings.Split(noproxy, ",")
			for _, s := range nplist {
				if containsIgnoreCase(addr, s) {
					http_proxy = ""
					break
				}
			}
		}
	}
	socks5Target, socks5 := strings.CutPrefix(http_proxy, "socks5://")
	if http_proxy != "" && !socks5 {
		url, err := url.Parse(http_proxy)
		if err == nil {
			addr = url.Host
		}
	}
	var c net.Conn
	var err error
	if socks5 {
		dialer, err := proxy.SOCKS5("tcp", socks5Target, nil, nil)
		if err != nil {
			return nil, err
		}
		c, err = dialer.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
	} else {
		c, err = net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			return nil, err
		}
	}

	if http_proxy != "" && !socks5 {
		fmt.Fprintf(c, "CONNECT %s HTTP/1.1\r\n", host)
		fmt.Fprintf(c, "Host: %s\r\n", host)
		fmt.Fprintf(c, "\r\n")
		br := bufio.NewReader(c)
		req, _ := http.NewRequest("CONNECT", host, nil)
		resp, err := http.ReadResponse(br, req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			f := strings.SplitN(resp.Status, " ", 2)
			return nil, errors.New(f[1])
		}
	}
	return c, nil
}

// Options are used to specify additional options for new clients, such as a Resource.
type Options struct {
	// Host specifies what host to connect to, as either "hostname" or "hostname:port"
	// If host is not specified, the  DNS SRV should be used to find the host from the domainpart of the JID.
	// Default the port to 5222.
	Host string

	// User specifies what user to authenticate to the remote server.
	User string

	// Password supplies the password to use for authentication with the remote server.
	Password string

	// DialTimeout is the time limit for establishing a connection. A
	// DialTimeout of zero means no timeout.
	DialTimeout time.Duration

	// Resource specifies an XMPP client resource, like "bot", instead of accepting one
	// from the server.  Use "" to let the server generate one for your client.
	Resource string

	// OAuthScope provides go-xmpp the required scope for OAuth2 authentication.
	OAuthScope string

	// OAuthToken provides go-xmpp with the required OAuth2 token used to authenticate
	OAuthToken string

	// OAuthXmlNs provides go-xmpp with the required namespaced used for OAuth2 authentication.  This is
	// provided to the server as the xmlns:auth attribute of the OAuth2 authentication request.
	OAuthXmlNs string

	// TLS Config
	TLSConfig *tls.Config

	// InsecureAllowUnencryptedAuth permits authentication over a TCP connection that has not been promoted to
	// TLS by STARTTLS; this could leak authentication information over the network, or permit man in the middle
	// attacks.
	InsecureAllowUnencryptedAuth bool

	// NoTLS directs go-xmpp to not use TLS initially to contact the server; instead, a plain old unencrypted
	// TCP connection should be used. (Can be combined with StartTLS to support STARTTLS-based servers.)
	NoTLS bool

	// StartTLS directs go-xmpp to STARTTLS if the server supports it; go-xmpp will automatically STARTTLS
	// if the server requires it regardless of this option.
	StartTLS bool

	// Debug output
	Debug bool

	// Use server sessions
	Session bool

	// Presence Status
	Status string

	// Status message
	StatusMessage string

	// Auth mechanism to use
	Mechanism string

	// XEP-0474: SASL SCRAM Downgrade Protection
	SSDP bool

	// XEP-0388: Extensible SASL Profile
	// Value for software
	UserAgentSW string

	// XEP-0388: XEP-0388: Extensible SASL Profile
	// Value for device
	UserAgentDev string

	// XEP-0388: Extensible SASL Profile
	// Unique stable identifier for the client installation
	// MUST be a valid UUIDv4
	UserAgentID string

	// Enable XEP-0484: Fast Authentication Streamlining Tokens
	Fast bool

	// XEP-0484: Fast Authentication Streamlining Tokens
	// Fast Token
	FastToken string

	// XEP-0484: Fast Authentication Streamlining Tokens
	// Fast Mechanism
	FastMechanism string

	// XEP-0484: Fast Authentication Streamlining Tokens
	// Invalidate the current token
	FastInvalidate bool
}

// NewClient establishes a new Client connection based on a set of Options.
func (o Options) NewClient() (*Client, error) {
	host := o.Host
	if strings.TrimSpace(host) == "" {
		a := strings.SplitN(o.User, "@", 2)
		if len(a) == 2 {
			if _, addrs, err := net.LookupSRV("xmpp-client", "tcp", a[1]); err == nil {
				if len(addrs) > 0 {
					// default to first record
					host = fmt.Sprintf("%s:%d", addrs[0].Target, addrs[0].Port)
					defP := addrs[0].Priority
					for _, adr := range addrs {
						if adr.Priority < defP {
							host = fmt.Sprintf("%s:%d", adr.Target, adr.Port)
							defP = adr.Priority
						}
					}
				} else {
					host = a[1]
				}
			} else {
				host = a[1]
			}
		}
	}
	c, err := connect(host, o.User, o.Password, o.DialTimeout)
	if err != nil {
		return nil, err
	}

	if strings.LastIndex(host, ":") > 0 {
		host = host[:strings.LastIndex(host, ":")]
	}

	client := new(Client)
	if o.NoTLS {
		client.conn = c
	} else {
		var tlsconn *tls.Conn
		if o.TLSConfig != nil {
			tlsconn = tls.Client(c, o.TLSConfig)
			host = o.TLSConfig.ServerName
		} else {
			newconfig := DefaultConfig.Clone()
			newconfig.ServerName = host
			tlsconn = tls.Client(c, newconfig)
		}
		if err = tlsconn.Handshake(); err != nil {
			return nil, err
		}
		insecureSkipVerify := DefaultConfig.InsecureSkipVerify
		if o.TLSConfig != nil {
			insecureSkipVerify = o.TLSConfig.InsecureSkipVerify
		}
		if !insecureSkipVerify {
			if err = tlsconn.VerifyHostname(host); err != nil {
				return nil, err
			}
		}
		client.conn = tlsconn
	}

	if err := client.init(&o); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

// NewClient creates a new connection to a host given as "hostname" or "hostname:port".
// If host is not specified, the  DNS SRV should be used to find the host from the domainpart of the JID.
// Default the port to 5222.
func NewClient(host, user, passwd string, debug bool) (*Client, error) {
	opts := Options{
		Host:     host,
		User:     user,
		Password: passwd,
		Debug:    debug,
		Session:  false,
	}
	return opts.NewClient()
}

// NewClientNoTLS creates a new client without TLS
func NewClientNoTLS(host, user, passwd string, debug bool) (*Client, error) {
	opts := Options{
		Host:     host,
		User:     user,
		Password: passwd,
		NoTLS:    true,
		Debug:    debug,
		Session:  false,
	}
	return opts.NewClient()
}

// Close closes the XMPP connection
func (c *Client) Close() error {
	c.shutdown = true
	if c.conn != (*tls.Conn)(nil) {
		fmt.Fprintf(c.stanzaWriter, "</stream:stream>\n")
		go func() {
			<-time.After(10 * time.Second)
			c.conn.Close()
		}()
		// Wait for the server also closing the stream.
		for {
			ee, err := c.nextEnd()
			// If the server already closed the stream it is
			// likely to receive an error when trying to parse
			// the stream. Therefore the connection is also closed
			// if an error is received.
			if err != nil {
				return c.conn.Close()
			}
			if ee.Name.Local == "stream" {
				return c.conn.Close()
			}
		}
	}
	return nil
}

func cnonce() string {
	randSize := big.NewInt(0)
	randSize.Lsh(big.NewInt(1), 64)
	cn, err := rand.Int(rand.Reader, randSize)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%016x", cn)
}

func (c *Client) init(o *Options) error {
	var domain string
	var user string
	a := strings.SplitN(o.User, "@", 2)
	// Check if User is not empty. Otherwise, we'll be attempting ANONYMOUS with Host domain.
	switch {
	case len(o.User) > 0:
		if len(a) != 2 {
			return errors.New("xmpp: invalid username (want user@domain): " + o.User)
		}
		user = a[0]
		domain = a[1]
	case strings.Contains(o.Host, ":"):
		domain = strings.SplitN(o.Host, ":", 2)[0]
	default:
		domain = o.Host
	}

	// Declare intent to be a jabber client and gather stream features.
	f, err := c.startStream(o, domain)
	if err != nil {
		return err
	}
	// Make the max. stanza size limit available.
	if f.Limits.MaxBytes != "" {
		c.LimitMaxBytes, err = strconv.Atoi(f.Limits.MaxBytes)
		if err != nil {
			c.LimitMaxBytes = 0
		}
	}
	// Make the servers time limit after which it might consider the stream idle available.
	if f.Limits.IdleSeconds != "" {
		c.LimitIdleSeconds, err = strconv.Atoi(f.Limits.IdleSeconds)
		if err != nil {
			c.LimitIdleSeconds = 0
		}
	}

	// If the server requires we STARTTLS, attempt to do so.
	if f, err = c.startTLSIfRequired(f, o, domain); err != nil {
		return err
	}
	var mechanism, channelBinding, clientFirstMessage, clientFinalMessageBare, authMessage string
	var bind2Data, resource, userAgentSW, userAgentDev, userAgentID, fastAuth string
	var serverSignature, keyingMaterial []byte
	var scramPlus, ok, tlsConnOK, tls13, serverEndPoint, sasl2, bind2 bool
	var cbsSlice, mechSlice []string
	var tlsConn *tls.Conn
	// Use SASL2 if available
	if f.Authentication.Mechanism != nil && c.IsEncrypted() {
		sasl2 = true
		mechSlice = f.Authentication.Mechanism
		// Detect whether bind2 is available
		if f.Authentication.Inline.Bind.Xmlns != "" {
			bind2 = true
		}
	} else {
		mechSlice = f.Mechanisms.Mechanism
	}
	if o.User == "" && o.Password == "" {
		foundAnonymous := false
		for _, m := range mechSlice {
			if m == "ANONYMOUS" {
				fmt.Fprintf(c.stanzaWriter, "<auth xmlns='%s' mechanism='ANONYMOUS' />\n", nsSASL)
				foundAnonymous = true
				break
			}
		}
		if !foundAnonymous {
			return fmt.Errorf("ANONYMOUS authentication is not an option and username and password were not specified")
		}
	} else {
		// Even digest forms of authentication are unsafe if we do not know that the host
		// we are talking to is the actual server, and not a man in the middle playing
		// proxy.
		if !c.IsEncrypted() && !o.InsecureAllowUnencryptedAuth {
			return errors.New("refusing to authenticate over unencrypted TCP connection")
		}

		tlsConn, ok = c.conn.(*tls.Conn)
		if ok {
			tlsConnOK = true
		}
		mechanism = ""
		if o.Mechanism != "" {
			if slices.Contains(mechSlice, o.Mechanism) {
				mechanism = o.Mechanism
			}
		} else {
			switch {
			case slices.Contains(mechSlice, "SCRAM-SHA-512-PLUS") && tlsConnOK:
				mechanism = "SCRAM-SHA-512-PLUS"
			case slices.Contains(mechSlice, "SCRAM-SHA-256-PLUS") && tlsConnOK:
				mechanism = "SCRAM-SHA-256-PLUS"
			case slices.Contains(mechSlice, "SCRAM-SHA-1-PLUS") && tlsConnOK:
				mechanism = "SCRAM-SHA-1-PLUS"
			case slices.Contains(mechSlice, "SCRAM-SHA-512"):
				mechanism = "SCRAM-SHA-512"
			case slices.Contains(mechSlice, "SCRAM-SHA-256"):
				mechanism = "SCRAM-SHA-256"
			case slices.Contains(mechSlice, "SCRAM-SHA-1"):
				mechanism = "SCRAM-SHA-1"
			case slices.Contains(mechSlice, "X-OAUTH2"):
				mechanism = "X-OAUTH2"
			case slices.Contains(mechSlice, "PLAIN") && tlsConnOK:
				mechanism = "PLAIN"
			}
		}
		if strings.HasPrefix(mechanism, "SCRAM-SHA") {
			if strings.HasSuffix(mechanism, "PLUS") {
				scramPlus = true
			}
			if scramPlus {
				for _, cbs := range f.ChannelBindings.ChannelBinding {
					cbsSlice = append(cbsSlice, cbs.Type)
				}
				tlsState := tlsConn.ConnectionState()
				switch tlsState.Version {
				case tls.VersionTLS13:
					tls13 = true
					if slices.Contains(cbsSlice, "tls-server-end-point") && !slices.Contains(cbsSlice, "tls-exporter") {
						serverEndPoint = true
					} else {
						keyingMaterial, err = tlsState.ExportKeyingMaterial("EXPORTER-Channel-Binding", nil, 32)
						if err != nil {
							return err
						}
					}
				case tls.VersionTLS10, tls.VersionTLS11, tls.VersionTLS12:
					if slices.Contains(cbsSlice, "tls-server-end-point") && !slices.Contains(cbsSlice, "tls-unique") {
						serverEndPoint = true
					} else {
						keyingMaterial = tlsState.TLSUnique
					}
				default:
					return errors.New(mechanism + ": unknown TLS version")
				}
				if serverEndPoint {
					var h hash.Hash
					// This material is not necessary for `tls-server-end-point` binding, but it is required to check that
					// the TLS connection was not renegotiated. This function will fail if that's the case (see
					// https://pkg.go.dev/crypto/tls#ConnectionState.ExportKeyingMaterial
					_, err = tlsState.ExportKeyingMaterial("EXPORTER-Channel-Binding", nil, 32)
					if err != nil {
						return err
					}
					switch tlsState.PeerCertificates[0].SignatureAlgorithm {
					case x509.SHA1WithRSA, x509.SHA256WithRSA, x509.ECDSAWithSHA1,
						x509.ECDSAWithSHA256, x509.SHA256WithRSAPSS:
						h = sha256.New()
					case x509.SHA384WithRSA, x509.ECDSAWithSHA384, x509.SHA384WithRSAPSS:
						h = sha512.New384()
					case x509.SHA512WithRSA, x509.ECDSAWithSHA512, x509.SHA512WithRSAPSS:
						h = sha512.New()
					}
					h.Write(tlsState.PeerCertificates[0].Raw)
					keyingMaterial = h.Sum(nil)
					h.Reset()
				}
				if len(keyingMaterial) == 0 {
					return errors.New(mechanism + ": no keying material")
				}
				switch {
				case tls13 && !serverEndPoint:
					channelBinding = base64.StdEncoding.EncodeToString(append([]byte("p=tls-exporter,,"), keyingMaterial[:]...))
				case serverEndPoint:
					channelBinding = base64.StdEncoding.EncodeToString(append([]byte("p=tls-server-end-point,,"), keyingMaterial[:]...))
				default:
					channelBinding = base64.StdEncoding.EncodeToString(append([]byte("p=tls-unique,,"), keyingMaterial[:]...))
				}
			}
			var shaNewFn func() hash.Hash
			switch mechanism {
			case "SCRAM-SHA-512", "SCRAM-SHA-512-PLUS":
				shaNewFn = sha512.New
			case "SCRAM-SHA-256", "SCRAM-SHA-256-PLUS":
				shaNewFn = sha256.New
			case "SCRAM-SHA-1", "SCRAM-SHA-1-PLUS":
				shaNewFn = sha1.New
			default:
				return errors.New("unsupported auth mechanism")
			}
			clientNonce := cnonce()
			if scramPlus {
				switch {
				case tls13 && !serverEndPoint:
					clientFirstMessage = "p=tls-exporter,,n=" + user + ",r=" + clientNonce
				case serverEndPoint:
					clientFirstMessage = "p=tls-server-end-point,,n=" + user + ",r=" + clientNonce
				default:
					clientFirstMessage = "p=tls-unique,,n=" + user + ",r=" + clientNonce
				}
			} else {
				clientFirstMessage = "n,,n=" + user + ",r=" + clientNonce
			}
			if sasl2 {
				if bind2 {
					if o.UserAgentSW != "" {
						resource = o.UserAgentSW
					} else {
						resource = "go-xmpp"
					}
					bind2Data = fmt.Sprintf("<bind xmlns='%s'><tag>%s</tag></bind>",
						nsBind2, resource)
				}
				if o.UserAgentSW != "" {
					userAgentSW = fmt.Sprintf("<software>%s</software>", o.UserAgentSW)
				} else {
					userAgentSW = "<software>go-xmpp</software>"
				}
				if o.UserAgentDev != "" {
					userAgentDev = fmt.Sprintf("<device>%s</device>", o.UserAgentDev)
				}
				if o.UserAgentID != "" {
					userAgentID = fmt.Sprintf(" id='%s'", o.UserAgentID)
				}
				if o.Fast && f.Authentication.Inline.Fast.Mechanism != nil && o.UserAgentID != "" && c.IsEncrypted() {
					var mech string
					if o.FastToken == "" {
						m := f.Authentication.Inline.Fast.Mechanism
						switch {
						case slices.Contains(m, "HT-SHA-256-EXPR") && tls13:
							mech = "HT-SHA-256-EXPR"
						case slices.Contains(m, "HT-SHA-256-UNIQ") && !tls13:
							mech = "HT-SHA-256-UNIQ"
						case slices.Contains(m, "HT-SHA-256-ENDP"):
							mech = "HT-SHA-256-ENDP"
						case slices.Contains(m, "HT-SHA-256-NONE"):
							mech = "HT-SHA-256-NONE"
						default:
							return fmt.Errorf("fast: unsupported auth mechanism %s", m)
						}
						fastAuth = fmt.Sprintf("<request-token xmlns='%s' mechanism='%s'/>", nsFast, mech)
					} else {
						var fastInvalidate string
						if o.FastInvalidate {
							fastInvalidate = " invalidate='true'"
						}
						fastAuth = fmt.Sprintf("<fast xmlns='%s'%s/>", nsFast, fastInvalidate)
						tlsState := tlsConn.ConnectionState()
						mechanism = o.FastMechanism
						switch mechanism {
						case "HT-SHA-256-EXPR":
							keyingMaterial, err = tlsState.ExportKeyingMaterial("EXPORTER-Channel-Binding", nil, 32)
							if err != nil {
								return err
							}
						case "HT-SHA-256-UNIQ":
							keyingMaterial = tlsState.TLSUnique
						case "HT-SHA-256-ENDP":
							var h hash.Hash
							switch tlsState.PeerCertificates[0].SignatureAlgorithm {
							case x509.SHA1WithRSA, x509.SHA256WithRSA, x509.ECDSAWithSHA1,
								x509.ECDSAWithSHA256, x509.SHA256WithRSAPSS:
								h = sha256.New()
							case x509.SHA384WithRSA, x509.ECDSAWithSHA384, x509.SHA384WithRSAPSS:
								h = sha512.New384()
							case x509.SHA512WithRSA, x509.ECDSAWithSHA512, x509.SHA512WithRSAPSS:
								h = sha512.New()
							}
							h.Write(tlsState.PeerCertificates[0].Raw)
							keyingMaterial = h.Sum(nil)
							h.Reset()
						case "HT-SHA-256-NONE":
							keyingMaterial = []byte("")
						default:
							return fmt.Errorf("fast: unsupported auth mechanism %s", mechanism)
						}
						h := hmac.New(sha256.New, []byte(o.FastToken))
						initiator := append([]byte("Initiator")[:], keyingMaterial[:]...)
						_, err = h.Write(initiator)
						if err != nil {
							return err
						}
						initiatorHashedToken := h.Sum(nil)
						user := strings.Split(o.User, "@")[0]
						clientFirstMessage = user + "\x00" + string(initiatorHashedToken)
					}
				}
				fmt.Fprintf(c.stanzaWriter,
					"<authenticate xmlns='%s' mechanism='%s'><initial-response>%s</initial-response><user-agent%s>%s%s</user-agent>%s%s</authenticate>\n",
					nsSASL2, mechanism, base64.StdEncoding.EncodeToString([]byte(clientFirstMessage)), userAgentID, userAgentSW, userAgentDev, bind2Data, fastAuth)
			} else {
				fmt.Fprintf(c.stanzaWriter, "<auth xmlns='%s' mechanism='%s'>%s</auth>\n",
					nsSASL, mechanism, base64.StdEncoding.EncodeToString([]byte(clientFirstMessage)))
			}
			var sfm string
			_, val, err := c.next()
			if err != nil {
				return err
			}
			switch v := val.(type) {
			case *sasl2Failure:
				errorMessage := v.Text
				if errorMessage == "" {
					// v.Any is type of sub-element in failure,
					// which gives a description of what failed if there was no text element
					errorMessage = v.Any.Local
				}
				return errors.New("auth failure: " + errorMessage)
			case *saslFailure:
				errorMessage := v.Text
				if errorMessage == "" {
					// v.Any is type of sub-element in failure,
					// which gives a description of what failed if there was no text element
					errorMessage = v.Any.Local
				}
				return errors.New("auth failure: " + errorMessage)
			case *sasl2Success:
				if strings.HasPrefix(mechanism, "SCRAM-SHA") {
					successMsg, err := base64.StdEncoding.DecodeString(v.AdditionalData)
					if err != nil {
						return err
					}
					if !strings.HasPrefix(string(successMsg), "v=") {
						return errors.New("server sent unexpected content in SCRAM success message")
					}
					c.Mechanism = mechanism
				}
				if strings.HasPrefix(mechanism, "HT-SHA") {
					// TODO: Check whether server implementations already support
					// https://www.ietf.org/archive/id/draft-schmaus-kitten-sasl-ht-09.html#section-3.3
					h := hmac.New(sha256.New, []byte(o.FastToken))
					responder := append([]byte("Responder")[:], keyingMaterial[:]...)
					_, err = h.Write(responder)
					if err != nil {
						return err
					}
					responderMsgRcv, err := base64.StdEncoding.DecodeString(v.AdditionalData)
					if err != nil {
						return err
					}
					responderMsgCalc := h.Sum(nil)
					if string(responderMsgCalc) != string(responderMsgRcv) {
						return fmt.Errorf("server sent unexpected content in FAST success message")
					}
					c.Mechanism = mechanism
				}
				if bind2 {
					c.jid = v.AuthorizationIdentifier
				}
				if v.Token.Token != "" && v.Token.Token != o.FastToken {
					m := f.Authentication.Inline.Fast.Mechanism
					switch {
					case slices.Contains(m, "HT-SHA-256-EXPR") && tls13:
						c.Fast.Mechanism = "HT-SHA-256-EXPR"
					case slices.Contains(m, "HT-SHA-256-UNIQ") && !tls13:
						c.Fast.Mechanism = "HT-SHA-256-UNIQ"
					case slices.Contains(m, "HT-SHA-256-ENDP"):
						c.Fast.Mechanism = "HT-SHA-256-ENDP"
					case slices.Contains(m, "HT-SHA-256-NONE"):
						c.Fast.Mechanism = "HT-SHA-256-NONE"
					}
					c.Fast.Token = v.Token.Token
					c.Fast.Expiry, _ = time.Parse(time.RFC3339, v.Token.Expiry)
				}
				if o.Session {
					// if server support session, open it
					cookie := getCookie() // generate new id value for session
					fmt.Fprintf(c.stanzaWriter, "<iq to='%s' type='set' id='%x'><session xmlns='%s'/></iq>\n", xmlEscape(domain), cookie, nsSession)
				}

				// We're connected and can now receive and send messages.
				fmt.Fprintf(c.stanzaWriter, "<presence xml:lang='en'><show>%s</show><status>%s</status></presence>\n", o.Status, o.StatusMessage)
				return nil
			case *sasl2Challenge:
				sfm = v.Text
			case *saslChallenge:
				sfm = v.Text
			}
			b, err := base64.StdEncoding.DecodeString(sfm)
			if err != nil {
				return err
			}
			var serverNonce, dgProtect string
			var salt []byte
			var iterations int
			for _, serverReply := range strings.Split(string(b), ",") {
				switch {
				case strings.HasPrefix(serverReply, "r="):
					serverNonce = strings.SplitN(serverReply, "=", 2)[1]
					if !strings.HasPrefix(serverNonce, clientNonce) {
						return errors.New("SCRAM: server nonce didn't start with client nonce")
					}
				case strings.HasPrefix(serverReply, "s="):
					salt, err = base64.StdEncoding.DecodeString(strings.SplitN(serverReply, "=", 2)[1])
					if err != nil {
						return err
					}
					if string(salt) == "" {
						return errors.New("SCRAM: server sent empty salt")
					}
				case strings.HasPrefix(serverReply, "i="):
					iterations, err = strconv.Atoi(strings.SplitN(serverReply,
						"=", 2)[1])
					if err != nil {
						return err
					}
				case strings.HasPrefix(serverReply, "d=") && o.SSDP:
					serverDgProtectHash := strings.SplitN(serverReply, "=", 2)[1]
					slices.Sort(f.Mechanisms.Mechanism)
					for _, mech := range f.Mechanisms.Mechanism {
						if dgProtect == "" {
							dgProtect = mech
						} else {
							dgProtect = dgProtect + "," + mech
						}
					}
					dgProtect = dgProtect + "|"
					slices.Sort(cbsSlice)
					for i, cb := range cbsSlice {
						if i == 0 {
							dgProtect = dgProtect + cb
						} else {
							dgProtect = dgProtect + "," + cb
						}
					}
					dgh := shaNewFn()
					dgh.Write([]byte(dgProtect))
					dHash := dgh.Sum(nil)
					dHashb64 := base64.StdEncoding.EncodeToString(dHash)
					if dHashb64 != serverDgProtectHash {
						return errors.New("SCRAM: downgrade protection hash mismatch")
					}
					dgh.Reset()
				case strings.HasPrefix(serverReply, "m="):
					return errors.New("SCRAM: server sent reserved 'm' attribute.")
				}
			}
			if scramPlus {
				clientFinalMessageBare = "c=" + channelBinding + ",r=" + serverNonce
			} else {
				clientFinalMessageBare = "c=biws,r=" + serverNonce
			}
			saltedPassword := pbkdf2.Key([]byte(o.Password), salt,
				iterations, shaNewFn().Size(), shaNewFn)
			h := hmac.New(shaNewFn, saltedPassword)
			_, err = h.Write([]byte("Client Key"))
			if err != nil {
				return err
			}
			clientKey := h.Sum(nil)
			h.Reset()
			var storedKey []byte
			switch mechanism {
			case "SCRAM-SHA-512", "SCRAM-SHA-512-PLUS":
				storedKey512 := sha512.Sum512(clientKey)
				storedKey = storedKey512[:]
			case "SCRAM-SHA-256", "SCRAM-SH-256-PLUS":
				storedKey256 := sha256.Sum256(clientKey)
				storedKey = storedKey256[:]
			case "SCRAM-SHA-1", "SCRAM-SHA-1-PLUS":
				storedKey1 := sha1.Sum(clientKey)
				storedKey = storedKey1[:]
			}
			_, err = h.Write([]byte("Server Key"))
			if err != nil {
				return err
			}
			serverFirstMessage, err := base64.StdEncoding.DecodeString(sfm)
			if err != nil {
				return err
			}
			authMessage = strings.SplitAfter(clientFirstMessage, ",,")[1] + "," +
				string(serverFirstMessage) + "," + clientFinalMessageBare
			h = hmac.New(shaNewFn, storedKey[:])
			_, err = h.Write([]byte(authMessage))
			if err != nil {
				return err
			}
			clientSignature := h.Sum(nil)
			h.Reset()
			if len(clientKey) != len(clientSignature) {
				return errors.New("SCRAM: client key and signature length mismatch")
			}
			clientProof := make([]byte, len(clientKey))
			for i := range clientKey {
				clientProof[i] = clientKey[i] ^ clientSignature[i]
			}
			h = hmac.New(shaNewFn, saltedPassword)
			_, err = h.Write([]byte("Server Key"))
			if err != nil {
				return err
			}
			serverKey := h.Sum(nil)
			h.Reset()
			h = hmac.New(shaNewFn, serverKey)
			_, err = h.Write([]byte(authMessage))
			if err != nil {
				return err
			}
			serverSignature = h.Sum(nil)
			if string(serverSignature) == "" {
				return errors.New("SCRAM: calculated an empty server signature")
			}
			clientFinalMessage := base64.StdEncoding.EncodeToString([]byte(clientFinalMessageBare +
				",p=" + base64.StdEncoding.EncodeToString(clientProof)))
			if sasl2 {
				fmt.Fprintf(c.stanzaWriter, "<response xmlns='%s'>%s</response>\n", nsSASL2,
					clientFinalMessage)
			} else {
				fmt.Fprintf(c.stanzaWriter, "<response xmlns='%s'>%s</response>\n", nsSASL,
					clientFinalMessage)
			}
		}
		if mechanism == "X-OAUTH2" && o.OAuthToken != "" && o.OAuthScope != "" {
			// Oauth authentication: send base64-encoded \x00 user \x00 token.
			raw := "\x00" + user + "\x00" + o.OAuthToken
			enc := make([]byte, base64.StdEncoding.EncodedLen(len(raw)))
			base64.StdEncoding.Encode(enc, []byte(raw))
			if sasl2 {
				fmt.Fprintf(c.stanzaWriter, "<auth xmlns='%s' mechanism='X-OAUTH2' auth:service='oauth2' "+
					"xmlns:auth='%s'>%s</auth>\n", nsSASL2, o.OAuthXmlNs, enc)
			} else {
				fmt.Fprintf(c.stanzaWriter, "<auth xmlns='%s' mechanism='X-OAUTH2' auth:service='oauth2' "+
					"xmlns:auth='%s'>%s</auth>\n", nsSASL, o.OAuthXmlNs, enc)
			}
		}
		if mechanism == "PLAIN" {
			// Plain authentication: send base64-encoded \x00 user \x00 password.
			raw := "\x00" + user + "\x00" + o.Password
			enc := make([]byte, base64.StdEncoding.EncodedLen(len(raw)))
			base64.StdEncoding.Encode(enc, []byte(raw))
			if sasl2 {
				fmt.Fprintf(c.conn, "<auth xmlns='%s' mechanism='PLAIN'>%s</auth>\n", nsSASL2, enc)
			} else {
				fmt.Fprintf(c.conn, "<auth xmlns='%s' mechanism='PLAIN'>%s</auth>\n", nsSASL, enc)
			}
		}
	}
	if mechanism == "" {
		return fmt.Errorf("no viable authentication method available: %v", f.Mechanisms.Mechanism)
	}
	// Next message should be either success or failure.
	name, val, err := c.next()
	if err != nil {
		return err
	}
	switch v := val.(type) {
	case *sasl2Success:
		if strings.HasPrefix(mechanism, "SCRAM-SHA") {
			successMsg, err := base64.StdEncoding.DecodeString(v.AdditionalData)
			if err != nil {
				return err
			}
			if !strings.HasPrefix(string(successMsg), "v=") {
				return errors.New("server sent unexpected content in SCRAM success message")
			}
			serverSignatureReply := strings.SplitN(string(successMsg), "v=", 2)[1]
			serverSignatureRemote, err := base64.StdEncoding.DecodeString(serverSignatureReply)
			if err != nil {
				return err
			}
			if string(serverSignature) != string(serverSignatureRemote) {
				return errors.New("SCRAM: server signature mismatch")
			}
			c.Mechanism = mechanism
		}
		if bind2 {
			c.jid = v.AuthorizationIdentifier
		}
		if v.Token.Token != "" {
			m := f.Authentication.Inline.Fast.Mechanism
			switch {
			case slices.Contains(m, "HT-SHA-256-EXPR") && tls13:
				c.Fast.Mechanism = "HT-SHA-256-EXPR"
			case slices.Contains(m, "HT-SHA-256-UNIQ") && !tls13:
				c.Fast.Mechanism = "HT-SHA-256-UNIQ"
			case slices.Contains(m, "HT-SHA-256-ENDP"):
				c.Fast.Mechanism = "HT-SHA-256-ENDP"
			case slices.Contains(m, "HT-SHA-256-NONE"):
				c.Fast.Mechanism = "HT-SHA-256-NONE"
			}
			c.Fast.Token = v.Token.Token
			c.Fast.Expiry, _ = time.Parse(time.RFC3339, v.Token.Expiry)
		}
	case *saslSuccess:
		if strings.HasPrefix(mechanism, "SCRAM-SHA") {
			successMsg, err := base64.StdEncoding.DecodeString(v.Text)
			if err != nil {
				return err
			}
			if !strings.HasPrefix(string(successMsg), "v=") {
				return errors.New("server sent unexpected content in SCRAM success message")
			}
			serverSignatureReply := strings.SplitN(string(successMsg), "v=", 2)[1]
			serverSignatureRemote, err := base64.StdEncoding.DecodeString(serverSignatureReply)
			if err != nil {
				return err
			}
			if string(serverSignature) != string(serverSignatureRemote) {
				return errors.New("SCRAM: server signature mismatch")
			}
			c.Mechanism = mechanism
		}
	case *sasl2Failure:
		errorMessage := v.Text
		if errorMessage == "" {
			// v.Any is type of sub-element in failure,
			// which gives a description of what failed if there was no text element
			errorMessage = v.Any.Local
		}
		return errors.New("auth failure: " + errorMessage)
	case *saslFailure:
		errorMessage := v.Text
		if errorMessage == "" {
			// v.Any is type of sub-element in failure,
			// which gives a description of what failed if there was no text element
			errorMessage = v.Any.Local
		}
		return errors.New("auth failure: " + errorMessage)
	default:
		return errors.New("expected <success> or <failure>, got <" + name.Local + "> in " + name.Space)
	}

	if !sasl2 {
		// Now that we're authenticated, we're supposed to start the stream over again.
		// Declare intent to be a jabber client.
		if f, err = c.startStream(o, domain); err != nil {
			return err
		}
	}
	// Make the max. stanza size limit available.
	if f.Limits.MaxBytes != "" {
		c.LimitMaxBytes, err = strconv.Atoi(f.Limits.MaxBytes)
		if err != nil {
			c.LimitMaxBytes = 0
		}
	}
	// Make the servers time limit after which it might consider the stream idle available.
	if f.Limits.IdleSeconds != "" {
		c.LimitIdleSeconds, err = strconv.Atoi(f.Limits.IdleSeconds)
		if err != nil {
			c.LimitIdleSeconds = 0
		}
	}

	if !bind2 {
		// Generate a unique cookie
		cookie := getCookie()

		// Send IQ message asking to bind to the local user name.
		if o.Resource == "" {
			fmt.Fprintf(c.stanzaWriter, "<iq type='set' id='%x'><bind xmlns='%s'></bind></iq>\n", cookie, nsBind)
		} else {
			fmt.Fprintf(c.stanzaWriter, "<iq type='set' id='%x'><bind xmlns='%s'><resource>%s</resource></bind></iq>\n", cookie, nsBind, o.Resource)
		}
		_, val, err = c.next()
		if err != nil {
			return err
		}
		switch v := val.(type) {
		case *streamError:
			errorMessage := v.Text.Text
			if errorMessage == "" {
				// v.Any is type of sub-element in failure,
				// which gives a description of what failed if there was no text element
				errorMessage = v.Any.Space
			}
			return errors.New("stream error: " + errorMessage)
		case *clientIQ:
			if v.Bind.XMLName.Space == nsBind {
				c.jid = v.Bind.Jid // our local id
				c.domain = domain
			} else {
				return errors.New("bind: unexpected reply to xmpp-bind IQ")
			}
		}
	}
	if o.Session {
		// if server support session, open it
		cookie := getCookie() // generate new id value for session
		fmt.Fprintf(c.stanzaWriter, "<iq to='%s' type='set' id='%x'><session xmlns='%s'/></iq>\n", xmlEscape(domain), cookie, nsSession)
	}

	// We're connected and can now receive and send messages.
	fmt.Fprintf(c.stanzaWriter, "<presence xml:lang='en'><show>%s</show><status>%s</status></presence>\n", o.Status, o.StatusMessage)

	return nil
}

// startTlsIfRequired examines the server's stream features and, if STARTTLS is required or supported, performs the TLS handshake.
// f will be updated if the handshake completes, as the new stream's features are typically different from the original.
func (c *Client) startTLSIfRequired(f *streamFeatures, o *Options, domain string) (*streamFeatures, error) {
	// whether we start tls is a matter of opinion: the server's and the user's.
	switch {
	case f.StartTLS == nil:
		// the server does not support STARTTLS
		return f, nil
	case !o.StartTLS && f.StartTLS.Required == nil:
		return f, nil
	case f.StartTLS.Required != nil:
		// the server requires STARTTLS.
	case !o.StartTLS:
		// the user wants STARTTLS and the server supports it.
	}
	var err error

	fmt.Fprintf(c.stanzaWriter, "<starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>\n")
	var k tlsProceed
	if err = c.p.DecodeElement(&k, nil); err != nil {
		return f, errors.New("unmarshal <proceed>: " + err.Error())
	}

	tc := o.TLSConfig
	if tc == nil {
		tc = DefaultConfig.Clone()
		// TODO(scott): we should consider using the server's address or reverse lookup
		tc.ServerName = domain
	}
	t := tls.Client(c.conn, tc)

	if err = t.Handshake(); err != nil {
		return f, errors.New("starttls handshake: " + err.Error())
	}
	c.conn = t

	// restart our declaration of XMPP stream intentions.
	tf, err := c.startStream(o, domain)
	if err != nil {
		return f, err
	}
	return tf, nil
}

// startStream will start a new XML decoder for the connection, signal the start of a stream to the server and verify that the server has
// also started the stream; if o.Debug is true, startStream will tee decoded XML data to stderr.  The features advertised by the server
// will be returned.
func (c *Client) startStream(o *Options, domain string) (*streamFeatures, error) {
	if o.Debug {
		c.p = xml.NewDecoder(tee{c.conn, DebugWriter})
		c.stanzaWriter = io.MultiWriter(c.conn, DebugWriter)
	} else {
		c.p = xml.NewDecoder(c.conn)
		c.stanzaWriter = c.conn
	}

	if c.IsEncrypted() {
		_, err := fmt.Fprintf(c.stanzaWriter, "<?xml version='1.0'?>"+
			"<stream:stream from='%s' to='%s' xmlns='%s'"+
			" xmlns:stream='%s' version='1.0'>\n",
			xmlEscape(o.User), xmlEscape(domain), nsClient, nsStream)
		if err != nil {
			return nil, err
		}
	} else {
		_, err := fmt.Fprintf(c.stanzaWriter, "<?xml version='1.0'?>"+
			"<stream:stream to='%s' xmlns='%s' xmlns:stream='%s' version='1.0'>\n",
			xmlEscape(domain), nsClient, nsStream)
		if err != nil {
			return nil, err
		}
	}

	// We expect the server to start a <stream>.
	se, err := c.nextStart()
	if err != nil {
		return nil, err
	}
	if se.Name.Space != nsStream || se.Name.Local != "stream" {
		return nil, fmt.Errorf("expected <stream> but got <%v> in %v", se.Name.Local, se.Name.Space)
	}

	// Now we're in the stream and can use Unmarshal.
	// Next message should be <features> to tell us authentication options.
	// See section 4.6 in RFC 3920.
	f := new(streamFeatures)
	if err = c.p.DecodeElement(f, nil); err != nil {
		return f, errors.New("unmarshal <features>: " + err.Error())
	}
	return f, nil
}

// IsEncrypted will return true if the client is connected using a TLS transport, either because it used.
// TLS to connect from the outset, or because it successfully used STARTTLS to promote a TCP connection to TLS.
func (c *Client) IsEncrypted() bool {
	_, ok := c.conn.(*tls.Conn)
	return ok
}

// Chat is an incoming or outgoing XMPP chat message.
type Chat struct {
	Remote    string
	Type      string
	Text      string
	Subject   string
	Thread    string
	Ooburl    string
	Oobdesc   string
	Lang      string
	ID        string
	ReplaceID string
	Roster    Roster
	Other     []string
	OtherElem []XMLElement
	Stamp     time.Time
}

type Roster []Contact

type Contact struct {
	Remote string
	Name   string
	Group  []string
}

// Presence is an XMPP presence notification.
type Presence struct {
	From   string
	To     string
	Type   string
	Show   string
	Status string
}

type IQ struct {
	ID    string
	From  string
	To    string
	Type  string
	Query []byte
}

// Recv waits to receive the next XMPP stanza.
func (c *Client) Recv() (stanza interface{}, err error) {
	for {
		_, val, err := c.next()
		if err != nil {
			return Chat{}, err
		}
		switch v := val.(type) {
		case *streamError:
			errorMessage := v.Text.Text
			if errorMessage == "" {
				// v.Any is type of sub-element in failure,
				// which gives a description of what failed if there was no text element
				errorMessage = v.Any.Space
			}
			return Chat{}, errors.New("stream error: " + errorMessage)
		case *clientMessage:
			if v.Event.XMLNS == XMPPNS_PUBSUB_EVENT {
				// Handle Pubsub notifications
				switch v.Event.Items.Node {
				case XMPPNS_AVATAR_PEP_METADATA:
					if len(v.Event.Items.Items) == 0 {
						return AvatarMetadata{}, errors.New("No avatar metadata items available")
					}

					return handleAvatarMetadata(v.Event.Items.Items[0].Body,
						v.From)
				// I am not sure whether this can even happen.
				// XEP-0084 only specifies a subscription to
				// the metadata node.
				/*case XMPPNS_AVATAR_PEP_DATA:
				return handleAvatarData(v.Event.Items.Items[0].Body,
					v.From,
					v.Event.Items.Items[0].ID)*/
				default:
					return pubsubClientToReturn(v.Event), nil
				}
			}

			stamp, _ := time.Parse(
				"2006-01-02T15:04:05Z",
				v.Delay.Stamp,
			)
			chat := Chat{
				Remote:    v.From,
				Type:      v.Type,
				Text:      v.Body,
				Subject:   v.Subject,
				Thread:    v.Thread,
				ID:        v.ID,
				ReplaceID: v.ReplaceID.ID,
				Other:     v.OtherStrings(),
				OtherElem: v.Other,
				Stamp:     stamp,
				Lang:      v.Lang,
			}
			return chat, nil
		case *clientQuery:
			var r Roster
			for _, item := range v.Item {
				r = append(r, Contact{item.Jid, item.Name, item.Group})
			}
			return Chat{Type: "roster", Roster: r}, nil
		case *clientPresence:
			return Presence{v.From, v.To, v.Type, v.Show, v.Status}, nil
		case *clientIQ:
			switch {
			case v.Query.XMLName.Space == "urn:xmpp:ping":
				// TODO check more strictly
				err := c.SendResultPing(v.ID, v.From)
				if err != nil {
					return Chat{}, err
				}
				fallthrough
			case v.Type == "error":
				switch v.ID {
				case "sub1":
					// Pubsub subscription failed
					var errs []clientPubsubError
					err := xml.Unmarshal([]byte(v.Error.InnerXML), &errs)
					if err != nil {
						return PubsubSubscription{}, err
					}

					var errsStr []string
					for _, e := range errs {
						errsStr = append(errsStr, e.XMLName.Local)
					}

					return PubsubSubscription{
						Errors: errsStr,
					}, nil
				default:
					res, err := xml.Marshal(v.Query)
					if err != nil {
						return Chat{}, err
					}

					return IQ{
						ID: v.ID, From: v.From, To: v.To, Type: v.Type,
						Query: res,
					}, nil
				}
			case v.Type == "result":
				switch v.ID {
				case "sub1":
					if v.Query.XMLName.Local == "pubsub" {
						// Subscription or unsubscription was successful
						var sub clientPubsubSubscription
						err := xml.Unmarshal([]byte(v.Query.InnerXML), &sub)
						if err != nil {
							return PubsubSubscription{}, err
						}

						return PubsubSubscription{
							SubID:  sub.SubID,
							JID:    sub.JID,
							Node:   sub.Node,
							Errors: nil,
						}, nil
					}
				case "unsub1":
					if v.Query.XMLName.Local == "pubsub" {
						var sub clientPubsubSubscription
						err := xml.Unmarshal([]byte(v.Query.InnerXML), &sub)
						if err != nil {
							return PubsubUnsubscription{}, err
						}

						return PubsubUnsubscription{
							SubID:  sub.SubID,
							JID:    v.From,
							Node:   sub.Node,
							Errors: nil,
						}, nil
					} else {
						// Unsubscribing MAY contain a pubsub element. But it does
						// not have to
						return PubsubUnsubscription{
							SubID:  "",
							JID:    v.From,
							Node:   "",
							Errors: nil,
						}, nil
					}
				case "info1":
					if v.Query.XMLName.Space == XMPPNS_DISCO_ITEMS {
						var itemsQuery clientDiscoItemsQuery
						err := xml.Unmarshal(v.InnerXML, &itemsQuery)
						if err != nil {
							return []DiscoItem{}, err
						}

						return DiscoItems{
							Jid:   v.From,
							Items: clientDiscoItemsToReturn(itemsQuery.Items),
						}, nil
					}
				case "info3":
					if v.Query.XMLName.Space == XMPPNS_DISCO_INFO {
						var disco clientDiscoQuery
						err := xml.Unmarshal(v.InnerXML, &disco)
						if err != nil {
							return DiscoResult{}, err
						}

						return DiscoResult{
							Features:   clientFeaturesToReturn(disco.Features),
							Identities: clientIdentitiesToReturn(disco.Identities),
						}, nil
					}
				case "items1", "items3":
					if v.Query.XMLName.Local == "pubsub" {
						var p clientPubsubItems
						err := xml.Unmarshal([]byte(v.Query.InnerXML), &p)
						if err != nil {
							return PubsubItems{}, err
						}

						switch p.Node {
						case XMPPNS_AVATAR_PEP_DATA:
							if len(p.Items) == 0 {
								return AvatarData{}, errors.New("No avatar data items available")
							}

							return handleAvatarData(p.Items[0].Body,
								v.From,
								p.Items[0].ID)
						case XMPPNS_AVATAR_PEP_METADATA:
							if len(p.Items) == 0 {
								return AvatarMetadata{}, errors.New("No avatar metadata items available")
							}

							return handleAvatarMetadata(p.Items[0].Body,
								v.From)
						default:
							return PubsubItems{
								p.Node,
								pubsubItemsToReturn(p.Items),
							}, nil
						}
					}
					// Note: XEP-0084 states that metadata and data
					// should be fetched with an id of retrieve1.
					// Since we already have PubSub implemented, we
					// can just use items1 and items3 to do the same
					// as an Avatar node is just a PEP (PubSub) node.
					/*case "retrieve1":
					var p clientPubsubItems
					err := xml.Unmarshal([]byte(v.Query.InnerXML), &p)
					if err != nil {
						return PubsubItems{}, err
					}

					switch p.Node {
					case XMPPNS_AVATAR_PEP_DATA:
						return handleAvatarData(p.Items[0].Body,
							v.From,
							p.Items[0].ID)
					case XMPPNS_AVATAR_PEP_METADATA:
						return handleAvatarMetadata(p.Items[0].Body,
							v
					}*/
				default:
					res, err := xml.Marshal(v.Query)
					if err != nil {
						return Chat{}, err
					}

					return IQ{
						ID: v.ID, From: v.From, To: v.To, Type: v.Type,
						Query: res,
					}, nil
				}
			case v.Query.XMLName.Local == "":
				return IQ{ID: v.ID, From: v.From, To: v.To, Type: v.Type}, nil
			default:
				res, err := xml.Marshal(v.Query)
				if err != nil {
					return Chat{}, err
				}

				return IQ{
					ID: v.ID, From: v.From, To: v.To, Type: v.Type,
					Query: res,
				}, nil
			}
		}
	}
}

// Send sends the message wrapped inside an XMPP message stanza body.
func (c *Client) Send(chat Chat) (n int, err error) {
	var subtext, thdtext, oobtext, msgidtext, msgcorrecttext string
	if chat.Subject != `` {
		subtext = `<subject>` + xmlEscape(chat.Subject) + `</subject>`
	}
	if chat.Thread != `` {
		thdtext = `<thread>` + xmlEscape(chat.Thread) + `</thread>`
	}
	if chat.Ooburl != `` {
		oobtext = `<x xmlns="jabber:x:oob"><url>` + xmlEscape(chat.Ooburl) + `</url>`
		if chat.Oobdesc != `` {
			oobtext += `<desc>` + xmlEscape(chat.Oobdesc) + `</desc>`
		}
		oobtext += `</x>`
	}
	if chat.ID != `` {
		msgidtext = `id='` + xmlEscape(chat.ID) + `'`
	} else {
		msgidtext = `id='` + cnonce() + `'`
	}

	if chat.ReplaceID != `` {
		msgcorrecttext = `<replace id='` + xmlEscape(chat.ReplaceID) + `' xmlns='urn:xmpp:message-correct:0'/>`
	}

	chat.Text = validUTF8(chat.Text)

	stanza := fmt.Sprintf("<message to='%s' type='%s' "+msgidtext+" xml:lang='en'>"+subtext+"<body>%s</body>"+msgcorrecttext+oobtext+thdtext+"</message>",
		xmlEscape(chat.Remote), xmlEscape(chat.Type), xmlEscape(chat.Text))

	if c.LimitMaxBytes != 0 && len(stanza) > c.LimitMaxBytes {
		return 0, fmt.Errorf("stanza size (%v bytes) exceeds server limit (%v bytes)",
			len(stanza), c.LimitMaxBytes)
	}

	return fmt.Fprint(c.stanzaWriter, stanza)
}

// SendOOB sends OOB data wrapped inside an XMPP message stanza, without actual body.
func (c *Client) SendOOB(chat Chat) (n int, err error) {
	var thdtext, oobtext string
	if chat.Thread != `` {
		thdtext = `<thread>` + xmlEscape(chat.Thread) + `</thread>`
	}
	if chat.Ooburl != `` {
		oobtext = `<x xmlns="jabber:x:oob"><url>` + xmlEscape(chat.Ooburl) + `</url>`
		if chat.Oobdesc != `` {
			oobtext += `<desc>` + xmlEscape(chat.Oobdesc) + `</desc>`
		}
		oobtext += `</x>`
	}
	stanza := fmt.Sprintf("<message to='%s' type='%s' id='%s' xml:lang='en'>"+oobtext+thdtext+"</message>\n",
		xmlEscape(chat.Remote), xmlEscape(chat.Type), cnonce())
	if c.LimitMaxBytes != 0 && len(stanza) > c.LimitMaxBytes {
		return 0, fmt.Errorf("stanza size (%v bytes) exceeds server limit (%v bytes)",
			len(stanza), c.LimitMaxBytes)
	}
	return fmt.Fprint(c.stanzaWriter, stanza)
}

// SendOrg sends the original text without being wrapped in an XMPP message stanza.
func (c *Client) SendOrg(org string) (n int, err error) {
	stanza := fmt.Sprint(org + "\n")
	if c.LimitMaxBytes != 0 && len(stanza) > c.LimitMaxBytes {
		return 0, fmt.Errorf("stanza size (%v bytes) exceeds server limit (%v bytes)",
			len(stanza), c.LimitMaxBytes)
	}
	return fmt.Fprint(c.stanzaWriter, stanza)
}

// SendPresence sends Presence wrapped inside XMPP presence stanza.
func (c *Client) SendPresence(presence Presence) (n int, err error) {
	// Forge opening presence tag
	var buf string = "<presence"

	if presence.From != "" {
		buf = buf + fmt.Sprintf(" from='%s'", xmlEscape(presence.From))
	}

	if presence.To != "" {
		buf = buf + fmt.Sprintf(" to='%s'", xmlEscape(presence.To))
	}

	if presence.Type != "" {
		// https://www.ietf.org/rfc/rfc3921.txt, 2.2.1, types can only be
		// unavailable, subscribe, subscribed, unsubscribe, unsubscribed, probe, error
		switch presence.Type {
		case "unavailable", "subscribe", "subscribed", "unsubscribe", "unsubscribed", "probe", "error":
			buf = buf + fmt.Sprintf(" type='%s'", xmlEscape(presence.Type))
		}
	}

	buf = buf + ">"

	// TODO: there may be optional tag "priority", but former presence type does not take this into account
	//       so either we must follow std, change type xmpp.Presence and break backward compatibility
	//       or leave it as-is and potentially break client software

	if presence.Show != "" {
		// https://www.ietf.org/rfc/rfc3921.txt 2.2.2.1, show can be only
		// away, chat, dnd, xa
		switch presence.Show {
		case "away", "chat", "dnd", "xa":
			buf = buf + fmt.Sprintf("<show>%s</show>", xmlEscape(presence.Show))
		}
	}

	if presence.Status != "" {
		buf = buf + fmt.Sprintf("<status>%s</status>", xmlEscape(presence.Status))
	}

	stanza := fmt.Sprintf(buf + "</presence>\n")
	if c.LimitMaxBytes != 0 && len(stanza) > c.LimitMaxBytes {
		return 0, fmt.Errorf("stanza size (%v bytes) exceeds server limit (%v bytes)",
			len(stanza), c.LimitMaxBytes)
	}
	return fmt.Fprint(c.stanzaWriter, stanza)
}

// SendKeepAlive sends a "whitespace keepalive" as described in chapter 4.6.1 of RFC6120.
func (c *Client) SendKeepAlive() (n int, err error) {
	return fmt.Fprintf(c.conn, " ")
}

// SendHtml sends the message as HTML as defined by XEP-0071
func (c *Client) SendHtml(chat Chat) (n int, err error) {
	stanza := fmt.Sprintf("<message to='%s' type='%s' xml:lang='en'><body>%s</body>"+
		"<html xmlns='http://jabber.org/protocol/xhtml-im'><body xmlns='http://www.w3.org/1999/xhtml'>%s</body></html></message>\n",
		xmlEscape(chat.Remote), xmlEscape(chat.Type), xmlEscape(chat.Text), chat.Text)
	if c.LimitMaxBytes != 0 && len(stanza) > c.LimitMaxBytes {
		return 0, fmt.Errorf("stanza size (%v bytes) exceeds server limit (%v bytes)",
			len(stanza), c.LimitMaxBytes)
	}
	return fmt.Fprint(c.stanzaWriter, stanza)
}

// Roster asks for the chat roster.
func (c *Client) Roster() error {
	fmt.Fprintf(c.stanzaWriter, "<iq from='%s' type='get' id='roster1'><query xmlns='jabber:iq:roster'/></iq>\n", xmlEscape(c.jid))
	return nil
}

// RFC 3920  C.1  Streams name space
type streamFeatures struct {
	XMLName         xml.Name `xml:"http://etherx.jabber.org/streams features"`
	Authentication  sasl2Authentication
	StartTLS        *tlsStartTLS
	Mechanisms      saslMechanisms
	ChannelBindings saslChannelBindings
	Bind            bindBind
	Session         bool
	Limits          streamLimits
}

type streamError struct {
	XMLName xml.Name `xml:"http://etherx.jabber.org/streams error"`
	Any     xml.Name
	Text    struct {
		Text  string `xml:",chardata"`
		Lang  string `xml:"lang,attr"`
		Xmlns string `xml:"xmlns,attr"`
	} `xml:"text"`
}

// RFC 3920  C.3  TLS name space
type tlsStartTLS struct {
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-tls starttls"`
	Required *string  `xml:"required"`
}

type tlsProceed struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-tls proceed"`
}

type tlsFailure struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-tls failure"`
}

type sasl2Authentication struct {
	XMLName   xml.Name `xml:"urn:xmpp:sasl:2 authentication"`
	Mechanism []string `xml:"mechanism"`
	Inline    struct {
		Text string `xml:",chardata"`
		Bind struct {
			XMLName xml.Name `xml:"urn:xmpp:bind:0 bind"`
			Xmlns   string   `xml:"xmlns,attr"`
			Text    string   `xml:",chardata"`
		} `xml:"bind"`
		Fast struct {
			XMLName   xml.Name `xml:"urn:xmpp:fast:0 fast"`
			Text      string   `xml:",chardata"`
			Tls0rtt   string   `xml:"tls-0rtt,attr"`
			Mechanism []string `xml:"mechanism"`
		} `xml:"fast"`
	} `xml:"inline"`
}

// RFC 3920  C.4  SASL name space
type saslMechanisms struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl mechanisms"`
	Mechanism []string `xml:"mechanism"`
}

type saslChannelBindings struct {
	XMLName        xml.Name `xml:"sasl-channel-binding"`
	Text           string   `xml:",chardata"`
	Xmlns          string   `xml:"xmlns,attr"`
	ChannelBinding []struct {
		Text string `xml:",chardata"`
		Type string `xml:"type,attr"`
	} `xml:"channel-binding"`
}

type saslAbort struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl abort"`
}

type sasl2Success struct {
	XMLName                 xml.Name `xml:"urn:xmpp:sasl:2 success"`
	Text                    string   `xml:",chardata"`
	AdditionalData          string   `xml:"additional-data"`
	AuthorizationIdentifier string   `xml:"authorization-identifier"`
	Bound                   struct {
		Text  string `xml:",chardata"`
		Xmlns string `xml:"urn:xmpp:bind:0,attr"`
	} `xml:"bound"`
	Token struct {
		Text   string `xml:",chardata"`
		Xmlns  string `xml:"urn:xmpp:fast:0,attr"`
		Expiry string `xml:"expiry,attr"`
		Token  string `xml:"token,attr"`
	} `xml:"token"`
}

type saslSuccess struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl success"`
	Text    string   `xml:",chardata"`
}

type sasl2Failure struct {
	XMLName xml.Name `xml:"urn:xmpp:sasl:2 failure"`
	Any     xml.Name `xml:",any"`
	Text    string   `xml:"text"`
}

type saslFailure struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl failure"`
	Any     xml.Name `xml:",any"`
	Text    string   `xml:"text"`
}

type sasl2Challenge struct {
	XMLName xml.Name `xml:"urn:xmpp:sasl:2 challenge"`
	Text    string   `xml:",chardata"`
}

type saslChallenge struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl challenge"`
	Text    string   `xml:",chardata"`
}

type streamLimits struct {
	XMLName     xml.Name `xml:"limits"`
	Text        string   `xml:",chardata"`
	Xmlns       string   `xml:"xmlns,attr"`
	MaxBytes    string   `xml:"max-bytes"`
	IdleSeconds string   `xml:"idle-seconds"`
}

// RFC 3920  C.5  Resource binding name space
type bindBind struct {
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-bind bind"`
	Resource string
	Jid      string `xml:"jid"`
}

type clientMessageCorrect struct {
	XMLName xml.Name `xml:"urn:xmpp:message-correct:0 replace"`
	ID      string   `xml:"id,attr"`
}

// RFC 3921  B.1  jabber:client
type clientMessage struct {
	XMLName xml.Name `xml:"jabber:client message"`
	From    string   `xml:"from,attr"`
	ID      string   `xml:"id,attr"`
	To      string   `xml:"to,attr"`
	Type    string   `xml:"type,attr"` // chat, error, groupchat, headline, or normal
	Lang    string   `xml:"lang,attr"`

	// These should technically be []clientText, but string is much more convenient.
	Subject   string `xml:"subject"`
	Body      string `xml:"body"`
	Thread    string `xml:"thread"`
	ReplaceID clientMessageCorrect

	// Pubsub
	Event clientPubsubEvent `xml:"event"`

	// Any hasn't matched element
	Other []XMLElement `xml:",any"`

	Delay Delay `xml:"delay"`
}

func (m *clientMessage) OtherStrings() []string {
	a := make([]string, len(m.Other))
	for i, e := range m.Other {
		a[i] = e.String()
	}
	return a
}

type XMLElement struct {
	XMLName  xml.Name
	Attr     []xml.Attr `xml:",any,attr"` // Save the attributes of the xml element
	InnerXML string     `xml:",innerxml"`
}

func (e *XMLElement) String() string {
	r := bytes.NewReader([]byte(e.InnerXML))
	d := xml.NewDecoder(r)
	var buf bytes.Buffer
	for {
		tok, err := d.Token()
		if err != nil {
			break
		}
		switch v := tok.(type) {
		case xml.StartElement:
			err = d.Skip()
		case xml.CharData:
			_, err = buf.Write(v)
		}
		if err != nil {
			break
		}
	}
	return buf.String()
}

type Delay struct {
	Stamp string `xml:"stamp,attr"`
}

type clientPresence struct {
	XMLName xml.Name `xml:"jabber:client presence"`
	From    string   `xml:"from,attr"`
	ID      string   `xml:"id,attr"`
	To      string   `xml:"to,attr"`
	Type    string   `xml:"type,attr"` // error, probe, subscribe, subscribed, unavailable, unsubscribe, unsubscribed
	Lang    string   `xml:"lang,attr"`

	Show     string `xml:"show"`   // away, chat, dnd, xa
	Status   string `xml:"status"` // sb []clientText
	Priority string `xml:"priority,attr"`
	Error    *clientError
}

type clientIQ struct {
	// info/query
	XMLName xml.Name   `xml:"jabber:client iq"`
	From    string     `xml:"from,attr"`
	ID      string     `xml:"id,attr"`
	To      string     `xml:"to,attr"`
	Type    string     `xml:"type,attr"` // error, get, result, set
	Query   XMLElement `xml:",any"`
	Error   clientError
	Bind    bindBind

	InnerXML []byte `xml:",innerxml"`
}

type clientError struct {
	XMLName  xml.Name `xml:"jabber:client error"`
	Code     string   `xml:",attr"`
	Type     string   `xml:"type,attr"`
	Any      xml.Name
	InnerXML []byte `xml:",innerxml"`
	Text     string
}

type clientQuery struct {
	Item []rosterItem
}

type rosterItem struct {
	XMLName      xml.Name `xml:"jabber:iq:roster item"`
	Jid          string   `xml:",attr"`
	Name         string   `xml:",attr"`
	Subscription string   `xml:",attr"`
	Group        []string
}

// Scan XML token stream to find next StartElement.
func (c *Client) nextStart() (xml.StartElement, error) {
	for {
		// Do not read from the stream if it's
		// going to be closed.
		if c.shutdown {
			return xml.StartElement{}, io.EOF
		}
		c.nextMutex.Lock()
		to, err := c.p.Token()
		if err != nil || to == nil {
			c.nextMutex.Unlock()
			return xml.StartElement{}, err
		}
		t := xml.CopyToken(to)
		switch t := t.(type) {
		case xml.StartElement:
			c.nextMutex.Unlock()
			return t, nil
		}
		c.nextMutex.Unlock()
	}
}

// Scan XML token stream to find next EndElement
func (c *Client) nextEnd() (xml.EndElement, error) {
	c.p.Strict = false
	for {
		c.nextMutex.Lock()
		to, err := c.p.Token()
		if err != nil || to == nil {
			c.nextMutex.Unlock()
			return xml.EndElement{}, err
		}
		t := xml.CopyToken(to)
		switch t := t.(type) {
		case xml.EndElement:
			// Do not unlock mutex if the stream is closed to
			// prevent further reading on the stream.
			if t.Name.Local == "stream" {
				return t, nil
			}
			c.nextMutex.Unlock()
			return t, nil
		}
		c.nextMutex.Unlock()
	}
}

// Scan XML token stream for next element and save into val.
// If val == nil, allocate new element based on proto map.
// Either way, return val.
func (c *Client) next() (xml.Name, interface{}, error) {
	// Read start element to find out what type we want.
	se, err := c.nextStart()
	if err != nil {
		return xml.Name{}, nil, err
	}

	// Put it in an interface and allocate one.
	var nv interface{}
	switch se.Name.Space + " " + se.Name.Local {
	case nsStream + " features":
		nv = &streamFeatures{}
	case nsStream + " error":
		nv = &streamError{}
	case nsTLS + " starttls":
		nv = &tlsStartTLS{}
	case nsTLS + " proceed":
		nv = &tlsProceed{}
	case nsTLS + " failure":
		nv = &tlsFailure{}
	case nsSASL + " mechanisms":
		nv = &saslMechanisms{}
	case nsSASL2 + " challenge":
		nv = &sasl2Challenge{}
	case nsSASL + " challenge":
		nv = &saslChallenge{}
	case nsSASL + " response":
		nv = ""
	case nsSASL + " abort":
		nv = &saslAbort{}
	case nsSASL2 + " success":
		nv = &sasl2Success{}
	case nsSASL + " success":
		nv = &saslSuccess{}
	case nsSASL2 + " failure":
		nv = &sasl2Failure{}
	case nsSASL + " failure":
		nv = &saslFailure{}
	case nsSASLCB + " sasl-channel-binding":
		nv = &saslChannelBindings{}
	case nsBind + " bind":
		nv = &bindBind{}
	case nsClient + " message":
		nv = &clientMessage{}
	case nsClient + " presence":
		nv = &clientPresence{}
	case nsClient + " iq":
		nv = &clientIQ{}
	case nsClient + " error":
		nv = &clientError{}
	default:
		return xml.Name{}, nil, errors.New("unexpected XMPP message " +
			se.Name.Space + " <" + se.Name.Local + "/>")
	}

	// Unmarshal into that storage.
	c.nextMutex.Lock()
	if err = c.p.DecodeElement(nv, &se); err != nil {
		return xml.Name{}, nil, err
	}
	c.nextMutex.Unlock()

	return se.Name, nv, err
}

func xmlEscape(s string) string {
	var b bytes.Buffer
	xml.Escape(&b, []byte(s))

	return b.String()
}

type tee struct {
	r io.Reader
	w io.Writer
}

func (t tee) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		t.w.Write(p[0:n])
		t.w.Write([]byte("\n"))
	}
	return
}

func validUTF8(s string) string {
	// Remove invalid code points.
	s = strings.ToValidUTF8(s, "�")
	reg := regexp.MustCompile(`[\x{0000}-\x{0008}\x{000B}\x{000C}\x{000E}-\x{001F}]`)
	s = reg.ReplaceAllString(s, "�")

	return s
}
