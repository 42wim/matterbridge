package common

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/golang/protobuf/proto"

	"github.com/status-im/markdown"
	"github.com/status-im/markdown/ast"

	accountJson "github.com/status-im/status-go/account/json"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/protocol/audio"
	"github.com/status-im/status-go/protocol/protobuf"
)

// QuotedMessage contains the original text of the message replied to
type QuotedMessage struct {
	ID          string `json:"id"`
	ContentType int64  `json:"contentType"`
	// From is a public key of the author of the message.
	From             string          `json:"from"`
	Text             string          `json:"text"`
	ParsedText       json.RawMessage `json:"parsedText,omitempty"`
	AlbumImages      json.RawMessage `json:"albumImages,omitempty"`
	AlbumImagesCount int64           `json:"albumImagesCount"`
	// ImageLocalURL is the local url of the image
	ImageLocalURL string `json:"image,omitempty"`
	// AudioLocalURL is the local url of the audio
	AudioLocalURL string `json:"audio,omitempty"`

	HasSticker bool `json:"sticker,omitempty"`
	// CommunityID is the id of the community advertised
	CommunityID string `json:"communityId,omitempty"`

	Deleted bool `json:"deleted,omitempty"`

	DeletedForMe bool `json:"deletedForMe,omitempty"`

	DiscordMessage *protobuf.DiscordMessage `json:"discordMessage,omitempty"`
}

type CommandState int

const (
	CommandStateRequestAddressForTransaction CommandState = iota + 1
	CommandStateRequestAddressForTransactionDeclined
	CommandStateRequestAddressForTransactionAccepted
	CommandStateRequestTransaction
	CommandStateRequestTransactionDeclined
	CommandStateTransactionPending
	CommandStateTransactionSent
)

type ContactRequestState int

const (
	ContactRequestStatePending ContactRequestState = iota + 1
	ContactRequestStateAccepted
	ContactRequestStateDismissed
)

type ContactVerificationState int

const (
	ContactVerificationStatePending ContactVerificationState = iota + 1
	ContactVerificationStateAccepted
	ContactVerificationStateDeclined
	ContactVerificationStateTrusted
	ContactVerificationStateUntrustworthy
	ContactVerificationStateCanceled
)

const EveryoneMentionTag = "0x00001"

type CommandParameters struct {
	// ID is the ID of the initial message
	ID string `json:"id"`
	// From is the address we are sending the command from
	From string `json:"from"`
	// Address is the address sent with the command
	Address string `json:"address"`
	// Contract is the contract address for ERC20 tokens
	Contract string `json:"contract"`
	// Value is the value as a string sent
	Value string `json:"value"`
	// TransactionHash is the hash of the transaction
	TransactionHash string `json:"transactionHash"`
	// CommandState is the state of the command
	CommandState CommandState `json:"commandState"`
	// The Signature of the pk-bytes+transaction-hash from the wallet
	// address originating
	Signature []byte `json:"signature"`
}

// GapParameters is the From and To indicating the missing period in chat history
type GapParameters struct {
	From uint32 `json:"from,omitempty"`
	To   uint32 `json:"to,omitempty"`
}

func (c *CommandParameters) IsTokenTransfer() bool {
	return len(c.Contract) != 0
}

const (
	OutgoingStatusSending   = "sending"
	OutgoingStatusSent      = "sent"
	OutgoingStatusDelivered = "delivered"
)

type Messages []*Message

func (m Messages) GetClock(i int) uint64 {
	return m[i].Clock
}

// Message represents a message record in the database,
// more specifically in user_messages table.
type Message struct {
	*protobuf.ChatMessage

	// ID calculated as keccak256(compressedAuthorPubKey, data) where data is unencrypted payload.
	ID string `json:"id"`
	// WhisperTimestamp is a timestamp of a Whisper envelope.
	WhisperTimestamp uint64 `json:"whisperTimestamp"`
	// From is a public key of the author of the message.
	From string `json:"from"`
	// Random 3 words name
	Alias string `json:"alias"`
	// Identicon of the author
	Identicon string `json:"identicon"`
	// The chat id to be stored locally
	LocalChatID string `json:"localChatId"`
	// Seen set to true when user have read this message already
	Seen           bool   `json:"seen"`
	OutgoingStatus string `json:"outgoingStatus,omitempty"`

	QuotedMessage *QuotedMessage `json:"quotedMessage"`

	// CommandParameters is the parameters sent with the message
	CommandParameters *CommandParameters `json:"commandParameters"`

	// GapParameters is the value from/to related to the gap
	GapParameters *GapParameters `json:"gapParameters,omitempty"`

	// Computed fields
	// RTL is whether this is a right-to-left message (arabic/hebrew script etc)
	RTL bool `json:"rtl"`
	// ParsedText is the parsed markdown for displaying
	ParsedText []byte `json:"parsedText,omitempty"`
	// ParsedTextAst is the ast of the parsed text
	ParsedTextAst *ast.Node `json:"-"`
	// LineCount is the count of newlines in the message
	LineCount int `json:"lineCount"`
	// Base64Image is the converted base64 image
	Base64Image string `json:"image,omitempty"`
	// ImagePath is the path of the image to be sent
	ImagePath string `json:"imagePath,omitempty"`
	// Base64Audio is the converted base64 audio
	Base64Audio string `json:"audio,omitempty"`
	// AudioPath is the path of the audio to be sent
	AudioPath string `json:"audioPath,omitempty"`
	// ImageLocalURL is the local url of the image
	ImageLocalURL string `json:"imageLocalUrl,omitempty"`
	// AudioLocalURL is the local url of the audio
	AudioLocalURL string `json:"audioLocalUrl,omitempty"`
	// StickerLocalURL is the local url of the sticker
	StickerLocalURL string `json:"stickerLocalUrl,omitempty"`

	// CommunityID is the id of the community to advertise
	CommunityID string `json:"communityId,omitempty"`

	// Replace indicates that this is a replacement of a message
	// that has been updated
	Replace string `json:"replace,omitempty"`
	New     bool   `json:"new,omitempty"`

	SigPubKey *ecdsa.PublicKey `json:"-"`

	// Mentions is an array of mentions for a given message
	Mentions []string

	// Mentioned is whether the user is mentioned in the message
	Mentioned bool `json:"mentioned"`

	// Replied is whether the user is replied to in the message
	Replied bool `json:"replied"`

	// Links is an array of links within given message
	Links              []string
	LinkPreviews       []LinkPreview       `json:"linkPreviews"`
	StatusLinkPreviews []StatusLinkPreview `json:"statusLinkPreviews"`

	// EditedAt indicates the clock value it was edited
	EditedAt uint64 `json:"editedAt"`

	// Deleted indicates if a message was deleted
	Deleted bool `json:"deleted"`

	DeletedBy string `json:"deletedBy,omitempty"`

	DeletedForMe bool `json:"deletedForMe"`

	// ContactRequestState is the state of the contact request message
	ContactRequestState ContactRequestState `json:"contactRequestState,omitempty"`

	// ContactVerificationState is the state of the identity verification process
	ContactVerificationState ContactVerificationState `json:"contactVerificationState,omitempty"`

	DiscordMessage *protobuf.DiscordMessage `json:"discordMessage,omitempty"`
}

func (m *Message) MarshalJSON() ([]byte, error) {
	type StickerAlias struct {
		Hash string `json:"hash"`
		Pack int32  `json:"pack"`
		URL  string `json:"url"`
	}

	if m.ChatMessage == nil {
		m.ChatMessage = &protobuf.ChatMessage{}
	}

	type MessageStructType struct {
		ID                       string                           `json:"id"`
		WhisperTimestamp         uint64                           `json:"whisperTimestamp"`
		From                     string                           `json:"from"`
		Alias                    string                           `json:"alias"`
		Identicon                string                           `json:"identicon"`
		Seen                     bool                             `json:"seen"`
		OutgoingStatus           string                           `json:"outgoingStatus,omitempty"`
		QuotedMessage            *QuotedMessage                   `json:"quotedMessage"`
		RTL                      bool                             `json:"rtl"`
		ParsedText               json.RawMessage                  `json:"parsedText,omitempty"`
		LineCount                int                              `json:"lineCount"`
		Text                     string                           `json:"text"`
		ChatID                   string                           `json:"chatId"`
		LocalChatID              string                           `json:"localChatId"`
		Clock                    uint64                           `json:"clock"`
		Replace                  string                           `json:"replace"`
		ResponseTo               string                           `json:"responseTo"`
		New                      bool                             `json:"new,omitempty"`
		EnsName                  string                           `json:"ensName"`
		DisplayName              string                           `json:"displayName"`
		Image                    string                           `json:"image,omitempty"`
		AlbumID                  string                           `json:"albumId,omitempty"`
		ImageWidth               uint32                           `json:"imageWidth,omitempty"`
		ImageHeight              uint32                           `json:"imageHeight,omitempty"`
		AlbumImagesCount         uint32                           `json:"albumImagesCount,omitempty"`
		Audio                    string                           `json:"audio,omitempty"`
		AudioDurationMs          uint64                           `json:"audioDurationMs,omitempty"`
		CommunityID              string                           `json:"communityId,omitempty"`
		Sticker                  *StickerAlias                    `json:"sticker,omitempty"`
		CommandParameters        *CommandParameters               `json:"commandParameters,omitempty"`
		GapParameters            *GapParameters                   `json:"gapParameters,omitempty"`
		Timestamp                uint64                           `json:"timestamp"`
		ContentType              protobuf.ChatMessage_ContentType `json:"contentType"`
		MessageType              protobuf.MessageType             `json:"messageType"`
		Mentions                 []string                         `json:"mentions,omitempty"`
		Mentioned                bool                             `json:"mentioned,omitempty"`
		Replied                  bool                             `json:"replied,omitempty"`
		Links                    []string                         `json:"links,omitempty"`
		LinkPreviews             []LinkPreview                    `json:"linkPreviews,omitempty"`
		StatusLinkPreviews       []StatusLinkPreview              `json:"statusLinkPreviews,omitempty"`
		EditedAt                 uint64                           `json:"editedAt,omitempty"`
		Deleted                  bool                             `json:"deleted,omitempty"`
		DeletedBy                string                           `json:"deletedBy,omitempty"`
		DeletedForMe             bool                             `json:"deletedForMe,omitempty"`
		ContactRequestState      ContactRequestState              `json:"contactRequestState,omitempty"`
		ContactVerificationState ContactVerificationState         `json:"contactVerificationState,omitempty"`
		DiscordMessage           *protobuf.DiscordMessage         `json:"discordMessage,omitempty"`
		BridgeMessage            *protobuf.BridgeMessage          `json:"bridgeMessage,omitempty"`
	}
	item := MessageStructType{
		ID:                       m.ID,
		WhisperTimestamp:         m.WhisperTimestamp,
		From:                     m.From,
		Alias:                    m.Alias,
		Identicon:                m.Identicon,
		Seen:                     m.Seen,
		OutgoingStatus:           m.OutgoingStatus,
		QuotedMessage:            m.QuotedMessage,
		RTL:                      m.RTL,
		ParsedText:               m.ParsedText,
		LineCount:                m.LineCount,
		Text:                     m.Text,
		Replace:                  m.Replace,
		ChatID:                   m.ChatId,
		LocalChatID:              m.LocalChatID,
		Clock:                    m.Clock,
		ResponseTo:               m.ResponseTo,
		New:                      m.New,
		EnsName:                  m.EnsName,
		DisplayName:              m.DisplayName,
		Image:                    m.ImageLocalURL,
		Audio:                    m.AudioLocalURL,
		CommunityID:              m.CommunityID,
		Timestamp:                m.Timestamp,
		ContentType:              m.ContentType,
		Mentions:                 m.Mentions,
		Mentioned:                m.Mentioned,
		Replied:                  m.Replied,
		Links:                    m.Links,
		LinkPreviews:             m.LinkPreviews,
		StatusLinkPreviews:       m.StatusLinkPreviews,
		MessageType:              m.MessageType,
		CommandParameters:        m.CommandParameters,
		GapParameters:            m.GapParameters,
		EditedAt:                 m.EditedAt,
		Deleted:                  m.Deleted,
		DeletedBy:                m.DeletedBy,
		DeletedForMe:             m.DeletedForMe,
		ContactRequestState:      m.ContactRequestState,
		ContactVerificationState: m.ContactVerificationState,
	}

	if sticker := m.GetSticker(); sticker != nil {
		item.Sticker = &StickerAlias{
			Pack: sticker.Pack,
			Hash: sticker.Hash,
			URL:  m.StickerLocalURL,
		}
	}

	if audio := m.GetAudio(); audio != nil {
		item.AudioDurationMs = audio.DurationMs
	}

	if image := m.GetImage(); image != nil {
		item.AlbumID = image.AlbumId
		item.ImageWidth = image.Width
		item.ImageHeight = image.Height
		item.AlbumImagesCount = image.AlbumImagesCount
	}

	if discordMessage := m.GetDiscordMessage(); discordMessage != nil {
		item.DiscordMessage = discordMessage
	}

	if bridgeMessage := m.GetBridgeMessage(); bridgeMessage != nil {
		item.BridgeMessage = bridgeMessage
	}

	if item.From != "" {
		ext, err := accountJson.ExtendStructWithPubKeyData(item.From, item)
		if err != nil {
			return nil, err
		}

		return json.Marshal(ext)
	}

	return json.Marshal(item)
}

func (m *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	aux := struct {
		*Alias
		ResponseTo       string                           `json:"responseTo"`
		EnsName          string                           `json:"ensName"`
		DisplayName      string                           `json:"displayName"`
		ChatID           string                           `json:"chatId"`
		Sticker          *protobuf.StickerMessage         `json:"sticker"`
		AudioDurationMs  uint64                           `json:"audioDurationMs"`
		ParsedText       json.RawMessage                  `json:"parsedText"`
		ContentType      protobuf.ChatMessage_ContentType `json:"contentType"`
		AlbumID          string                           `json:"albumId"`
		ImageWidth       uint32                           `json:"imageWidth"`
		ImageHeight      uint32                           `json:"imageHeight"`
		AlbumImagesCount uint32                           `json:"albumImagesCount"`
		From             string                           `json:"from"`
		Deleted          bool                             `json:"deleted,omitempty"`
		DeletedForMe     bool                             `json:"deletedForMe,omitempty"`
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.ContentType == protobuf.ChatMessage_STICKER {
		m.Payload = &protobuf.ChatMessage_Sticker{Sticker: aux.Sticker}
	}
	if aux.ContentType == protobuf.ChatMessage_AUDIO {
		m.Payload = &protobuf.ChatMessage_Audio{
			Audio: &protobuf.AudioMessage{DurationMs: aux.AudioDurationMs},
		}
	}

	if aux.ContentType == protobuf.ChatMessage_IMAGE {
		m.Payload = &protobuf.ChatMessage_Image{
			Image: &protobuf.ImageMessage{
				AlbumId:          aux.AlbumID,
				Width:            aux.ImageWidth,
				Height:           aux.ImageHeight,
				AlbumImagesCount: aux.AlbumImagesCount},
		}
	}

	m.ResponseTo = aux.ResponseTo
	m.EnsName = aux.EnsName
	m.DisplayName = aux.DisplayName
	m.ChatId = aux.ChatID
	m.ContentType = aux.ContentType
	m.ParsedText = aux.ParsedText
	m.From = aux.From
	m.Deleted = aux.Deleted
	m.DeletedForMe = aux.DeletedForMe
	return nil
}

// Check if the first character is Hebrew or Arabic or the RTL character
func isRTL(s string) bool {
	first, _ := utf8.DecodeRuneInString(s)
	return unicode.Is(unicode.Hebrew, first) ||
		unicode.Is(unicode.Arabic, first) ||
		// RTL character
		first == '\u200f'
}

// parseImage check the message contains an image, and if so
// it creates the a base64 encoded version of it.
func (m *Message) parseImage() error {
	if m.ContentType != protobuf.ChatMessage_IMAGE {
		return nil
	}
	image := m.GetImage()
	if image == nil {
		return errors.New("image empty")
	}

	payload := image.Payload

	e64 := base64.StdEncoding

	maxEncLen := e64.EncodedLen(len(payload))
	encBuf := make([]byte, maxEncLen)

	e64.Encode(encBuf, payload)

	mime, err := images.GetMimeType(image.Payload)

	if err != nil {
		return err
	}

	m.Base64Image = fmt.Sprintf("data:image/%s;base64,%s", mime, encBuf)

	return nil
}

// parseAudio check the message contains an audio, and if so
// it creates a base64 encoded version of it.
func (m *Message) parseAudio() error {
	if m.ContentType != protobuf.ChatMessage_AUDIO {
		return nil
	}
	audio := m.GetAudio()
	if audio == nil {
		return errors.New("audio empty")
	}

	payload := audio.Payload

	e64 := base64.StdEncoding

	maxEncLen := e64.EncodedLen(len(payload))
	encBuf := make([]byte, maxEncLen)

	e64.Encode(encBuf, payload)

	mime, err := getAudioMessageMIME(audio)

	if err != nil {
		return err
	}

	m.Base64Audio = fmt.Sprintf("data:audio/%s;base64,%s", mime, encBuf)

	return nil
}

// implement interface of https://github.com/status-im/markdown/blob/b9fe921681227b1dace4b56364e15edb3b698308/ast/node.go#L701
type SimplifiedTextVisitor struct {
	text           string
	canonicalNames map[string]string
}

func (v *SimplifiedTextVisitor) Visit(node ast.Node, entering bool) ast.WalkStatus {
	// only on entering we fetch, otherwise we go on
	if !entering {
		return ast.GoToNext
	}

	switch n := node.(type) {
	case *ast.Mention:
		literal := string(n.Literal)
		canonicalName, ok := v.canonicalNames[literal]
		if ok {
			v.text += canonicalName
		} else {
			v.text += literal
		}
	case *ast.Link:
		destination := string(n.Destination)
		v.text += destination
	default:
		var literal string

		leaf := node.AsLeaf()
		container := node.AsContainer()
		if leaf != nil {
			literal = string(leaf.Literal)
		} else if container != nil {
			literal = string(container.Literal)
		}
		v.text += literal
	}

	return ast.GoToNext
}

// implement interface of https://github.com/status-im/markdown/blob/b9fe921681227b1dace4b56364e15edb3b698308/ast/node.go#L701
type MentionsAndLinksVisitor struct {
	identity  string
	mentioned bool
	mentions  []string
	links     []string
}

type LinksVisitor struct {
	Links []string
}

func (v *MentionsAndLinksVisitor) Visit(node ast.Node, entering bool) ast.WalkStatus {
	// only on entering we fetch, otherwise we go on
	if !entering {
		return ast.GoToNext
	}
	switch n := node.(type) {
	case *ast.Mention:
		mention := string(n.Literal)
		if mention == v.identity || mention == EveryoneMentionTag {
			v.mentioned = true
		}
		v.mentions = append(v.mentions, mention)
	case *ast.Link:
		v.links = append(v.links, string(n.Destination))
	}

	return ast.GoToNext
}

func (v *LinksVisitor) Visit(node ast.Node, entering bool) ast.WalkStatus {
	if !entering {
		return ast.GoToNext
	}

	switch n := node.(type) {
	case *ast.Link:
		v.Links = append(v.Links, string(n.Destination))
	}

	return ast.GoToNext
}

func runMentionsAndLinksVisitor(parsedText ast.Node, identity string) *MentionsAndLinksVisitor {
	visitor := &MentionsAndLinksVisitor{identity: identity}
	ast.Walk(parsedText, visitor)
	return visitor
}

func RunLinksVisitor(parsedText ast.Node) *LinksVisitor {
	visitor := &LinksVisitor{}
	ast.Walk(parsedText, visitor)
	return visitor
}

// PrepareContent return the parsed content of the message, the line-count and whether
// is a right-to-left message
func (m *Message) PrepareContent(identity string) error {
	var parsedText ast.Node
	switch m.ContentType {
	case protobuf.ChatMessage_DISCORD_MESSAGE:
		parsedText = markdown.Parse([]byte(m.GetDiscordMessage().Content), nil)
	default:
		parsedText = markdown.Parse([]byte(m.Text), nil)
	}

	visitor := runMentionsAndLinksVisitor(parsedText, identity)
	m.Mentions = visitor.mentions
	m.Links = visitor.links
	// Leave it set if already set, as sometimes we might run this without
	// an identity
	if !m.Mentioned || identity != "" {
		m.Mentioned = visitor.mentioned
	}
	jsonParsedText, err := json.Marshal(parsedText)
	if err != nil {
		return err
	}
	m.ParsedTextAst = &parsedText
	m.ParsedText = jsonParsedText
	m.LineCount = strings.Count(m.Text, "\n")
	m.RTL = isRTL(m.Text)
	if err := m.parseImage(); err != nil {
		return err
	}
	return m.parseAudio()
}

// GetSimplifiedText returns a the text stripped of all the markdown and with mentions
// replaced by canonical names
func (m *Message) GetSimplifiedText(identity string, canonicalNames map[string]string) (string, error) {

	if m.ContentType == protobuf.ChatMessage_AUDIO {
		return "Audio", nil
	}
	if m.ContentType == protobuf.ChatMessage_STICKER {
		return "Sticker", nil
	}
	if m.ContentType == protobuf.ChatMessage_IMAGE {
		return "Image", nil
	}
	if m.ContentType == protobuf.ChatMessage_COMMUNITY {
		return "Community", nil
	}
	if m.ContentType == protobuf.ChatMessage_SYSTEM_MESSAGE_CONTENT_PRIVATE_GROUP {
		return "Group", nil
	}

	if m.ParsedTextAst == nil {
		err := m.PrepareContent(identity)
		if err != nil {
			return "", err
		}
	}
	visitor := &SimplifiedTextVisitor{canonicalNames: canonicalNames}
	ast.Walk(*m.ParsedTextAst, visitor)
	return visitor.text, nil
}

func getAudioMessageMIME(i *protobuf.AudioMessage) (string, error) {
	switch i.Type {
	case protobuf.AudioMessage_AAC:
		return "aac", nil
	case protobuf.AudioMessage_AMR:
		return "amr", nil
	}

	return "", errors.New("audio format not supported")
}

// GetSigPubKey returns an ecdsa encoded public key
// this function is required to implement the ChatEntity interface
func (m *Message) GetSigPubKey() *ecdsa.PublicKey {
	return m.SigPubKey
}

// GetProtoBuf returns the struct's embedded protobuf struct
// this function is required to implement the ChatEntity interface
func (m *Message) GetProtobuf() proto.Message {
	return m.ChatMessage
}

// SetMessageType a setter for the MessageType field
// this function is required to implement the ChatEntity interface
func (m *Message) SetMessageType(messageType protobuf.MessageType) {
	m.MessageType = messageType
}

// WrapGroupMessage indicates whether we should wrap this in membership information
func (m *Message) WrapGroupMessage() bool {
	return true
}

// GetPublicKey attempts to return or recreate the *ecdsa.PublicKey of the Message sender.
// If the m.SigPubKey is set this will be returned
// If the m.From is present the string is decoded and unmarshalled into a *ecdsa.PublicKey, the m.SigPubKey is set and returned
// Else an error is thrown
// This function differs from GetSigPubKey() as this function may return an error
func (m *Message) GetSenderPubKey() (*ecdsa.PublicKey, error) {
	// TODO requires tests

	if m.SigPubKey != nil {
		return m.SigPubKey, nil
	}

	if len(m.From) > 0 {
		fromB, err := hex.DecodeString(m.From[2:])
		if err != nil {
			return nil, err
		}

		senderPubKey, err := crypto.UnmarshalPubkey(fromB)
		if err != nil {
			return nil, err
		}

		m.SigPubKey = senderPubKey
		return senderPubKey, nil
	}

	return nil, errors.New("no Message.SigPubKey or Message.From set unable to get public key")
}

func (m *Message) LoadAudio() error {
	file, err := os.Open(m.AudioPath)
	if err != nil {
		return err
	}
	defer file.Close()

	payload, err := ioutil.ReadAll(file)
	if err != nil {
		return err

	}
	audioMessage := m.GetAudio()
	if audioMessage == nil {
		return errors.New("no audio has been passed")
	}
	audioMessage.Payload = payload
	audioMessage.Type = audio.Type(payload)
	m.Payload = &protobuf.ChatMessage_Audio{Audio: audioMessage}
	return os.Remove(m.AudioPath)
}

func (m *Message) LoadImage() error {
	payload, err := images.OpenAndAdjustImage(images.CroppedImage{ImagePath: m.ImagePath}, false)

	if err != nil {
		return err
	}
	imageMessage := m.GetImage()
	imageMessage.Payload = payload
	imageMessage.Format = images.GetProtobufImageFormat(payload)
	m.Payload = &protobuf.ChatMessage_Image{Image: imageMessage}

	return nil
}

func (m *Message) SetAlbumIDAndImagesCount(albumID string, imagesCount uint32) error {
	imageMessage := m.GetImage()
	if imageMessage == nil {
		return errors.New("Image is empty")
	}
	imageMessage.AlbumId = albumID
	imageMessage.AlbumImagesCount = imagesCount
	m.Payload = &protobuf.ChatMessage_Image{Image: imageMessage}

	return nil
}

func NewMessage() *Message {
	return &Message{
		ChatMessage: &protobuf.ChatMessage{},
	}
}
