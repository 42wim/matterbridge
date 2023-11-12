package protocol

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"sort"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/protobuf"
)

// MembershipUpdateMessage is a message used to propagate information
// about group membership changes.
// For more information, see https://github.com/status-im/specs/blob/master/status-group-chats-spec.md.
type MembershipUpdateMessage struct {
	ChatID        string                  `json:"chatId"` // UUID concatenated with hex-encoded public key of the creator for the chat
	Events        []MembershipUpdateEvent `json:"events"`
	Message       *protobuf.ChatMessage   `json:"-"`
	EmojiReaction *protobuf.EmojiReaction `json:"-"`
}

const signatureLength = 65

func MembershipUpdateEventFromProtobuf(chatID string, raw []byte) (*MembershipUpdateEvent, error) {
	if len(raw) <= signatureLength {
		return nil, errors.New("invalid payload length")
	}
	decodedEvent := protobuf.MembershipUpdateEvent{}
	signature := raw[:signatureLength]
	encodedEvent := raw[signatureLength:]

	signatureMaterial := append([]byte(chatID), encodedEvent...)
	publicKey, err := crypto.ExtractSignature(signatureMaterial, signature)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract signature")
	}

	from := publicKeyToString(publicKey)

	err = proto.Unmarshal(encodedEvent, &decodedEvent)
	if err != nil {
		return nil, err
	}
	return &MembershipUpdateEvent{
		ClockValue: decodedEvent.Clock,
		ChatID:     chatID,
		Members:    decodedEvent.Members,
		Name:       decodedEvent.Name,
		Type:       decodedEvent.Type,
		Color:      decodedEvent.Color,
		Image:      decodedEvent.Image,
		Signature:  signature,
		RawPayload: encodedEvent,
		From:       from,
	}, nil
}

func (m *MembershipUpdateMessage) ToProtobuf() (*protobuf.MembershipUpdateMessage, error) {
	var rawEvents [][]byte
	for _, e := range m.Events {
		var encodedEvent []byte
		encodedEvent = append(encodedEvent, e.Signature...)
		encodedEvent = append(encodedEvent, e.RawPayload...)
		rawEvents = append(rawEvents, encodedEvent)
	}

	mUM := &protobuf.MembershipUpdateMessage{
		ChatId: m.ChatID,
		Events: rawEvents,
	}

	// If message is not piggybacking anything, that's a valid case and we just return
	switch {
	case m.Message != nil:
		mUM.ChatEntity = &protobuf.MembershipUpdateMessage_Message{Message: m.Message}
	case m.EmojiReaction != nil:
		mUM.ChatEntity = &protobuf.MembershipUpdateMessage_EmojiReaction{EmojiReaction: m.EmojiReaction}
	}

	return mUM, nil
}

func MembershipUpdateMessageFromProtobuf(raw *protobuf.MembershipUpdateMessage) (*MembershipUpdateMessage, error) {
	var events []MembershipUpdateEvent
	for _, e := range raw.Events {
		verifiedEvent, err := MembershipUpdateEventFromProtobuf(raw.ChatId, e)
		if err != nil {
			return nil, err
		}
		events = append(events, *verifiedEvent)
	}
	return &MembershipUpdateMessage{
		ChatID:        raw.ChatId,
		Events:        events,
		Message:       raw.GetMessage(),
		EmojiReaction: raw.GetEmojiReaction(),
	}, nil
}

// EncodeMembershipUpdateMessage encodes a MembershipUpdateMessage using protobuf serialization.
func EncodeMembershipUpdateMessage(value MembershipUpdateMessage) ([]byte, error) {
	pb, err := value.ToProtobuf()
	if err != nil {
		return nil, err
	}

	return proto.Marshal(pb)
}

// MembershipUpdateEvent contains an event information.
// Member and Members are hex-encoded values with 0x prefix.
type MembershipUpdateEvent struct {
	Type       protobuf.MembershipUpdateEvent_EventType `json:"type"`
	ClockValue uint64                                   `json:"clockValue"`
	Members    []string                                 `json:"members,omitempty"` // in "members-added" and "admins-added" events
	Name       string                                   `json:"name,omitempty"`    // name of the group chat
	Color      string                                   `json:"color,omitempty"`   // color of the group chat
	Image      []byte                                   `json:"image,omitempty"`   // image of the group chat
	From       string                                   `json:"from,omitempty"`
	Signature  []byte                                   `json:"signature,omitempty"`
	ChatID     string                                   `json:"chatId"`
	RawPayload []byte                                   `json:"rawPayload"`
}

func (u *MembershipUpdateEvent) Equal(update MembershipUpdateEvent) bool {
	return bytes.Equal(u.Signature, update.Signature)
}

func (u *MembershipUpdateEvent) Sign(key *ecdsa.PrivateKey) error {
	if len(u.ChatID) == 0 {
		return errors.New("can't sign with empty chatID")
	}
	encodedEvent, err := proto.Marshal(u.ToProtobuf())
	if err != nil {
		return err
	}
	u.RawPayload = encodedEvent
	var signatureMaterial []byte
	signatureMaterial = append(signatureMaterial, []byte(u.ChatID)...)
	signatureMaterial = crypto.Keccak256(append(signatureMaterial, u.RawPayload...))
	signature, err := crypto.Sign(signatureMaterial, key)

	if err != nil {
		return err
	}
	u.Signature = signature
	u.From = publicKeyToString(&key.PublicKey)
	return nil
}

func (u *MembershipUpdateEvent) ToProtobuf() *protobuf.MembershipUpdateEvent {
	return &protobuf.MembershipUpdateEvent{
		Clock:   u.ClockValue,
		Name:    u.Name,
		Color:   u.Color,
		Image:   u.Image,
		Members: u.Members,
		Type:    u.Type,
	}
}

func MergeMembershipUpdateEvents(dest []MembershipUpdateEvent, src []MembershipUpdateEvent) []MembershipUpdateEvent {
	for _, update := range src {
		var exists bool
		for _, existing := range dest {
			if existing.Equal(update) {
				exists = true
				break
			}
		}
		if !exists {
			dest = append(dest, update)
		}
	}
	return dest
}

func NewChatCreatedEvent(name string, color string, clock uint64) MembershipUpdateEvent {
	return MembershipUpdateEvent{
		Type:       protobuf.MembershipUpdateEvent_CHAT_CREATED,
		Name:       name,
		ClockValue: clock,
		Color:      color,
	}
}

func NewNameChangedEvent(name string, clock uint64) MembershipUpdateEvent {
	return MembershipUpdateEvent{
		Type:       protobuf.MembershipUpdateEvent_NAME_CHANGED,
		Name:       name,
		ClockValue: clock,
	}
}

func NewColorChangedEvent(color string, clock uint64) MembershipUpdateEvent {
	return MembershipUpdateEvent{
		Type:       protobuf.MembershipUpdateEvent_COLOR_CHANGED,
		Color:      color,
		ClockValue: clock,
	}
}

func NewImageChangedEvent(image []byte, clock uint64) MembershipUpdateEvent {
	return MembershipUpdateEvent{
		Type:       protobuf.MembershipUpdateEvent_IMAGE_CHANGED,
		Image:      image,
		ClockValue: clock,
	}
}

func NewMembersAddedEvent(members []string, clock uint64) MembershipUpdateEvent {
	return MembershipUpdateEvent{
		Type:       protobuf.MembershipUpdateEvent_MEMBERS_ADDED,
		Members:    members,
		ClockValue: clock,
	}
}

func NewMemberJoinedEvent(clock uint64) MembershipUpdateEvent {
	return MembershipUpdateEvent{
		Type:       protobuf.MembershipUpdateEvent_MEMBER_JOINED,
		ClockValue: clock,
	}
}

func NewAdminsAddedEvent(admins []string, clock uint64) MembershipUpdateEvent {
	return MembershipUpdateEvent{
		Type:       protobuf.MembershipUpdateEvent_ADMINS_ADDED,
		Members:    admins,
		ClockValue: clock,
	}
}

func NewMemberRemovedEvent(member string, clock uint64) MembershipUpdateEvent {
	return MembershipUpdateEvent{
		Type:       protobuf.MembershipUpdateEvent_MEMBER_REMOVED,
		Members:    []string{member},
		ClockValue: clock,
	}
}

func NewAdminRemovedEvent(admin string, clock uint64) MembershipUpdateEvent {
	return MembershipUpdateEvent{
		Type:       protobuf.MembershipUpdateEvent_ADMIN_REMOVED,
		Members:    []string{admin},
		ClockValue: clock,
	}
}

type Group struct {
	chatID  string
	name    string
	color   string
	image   []byte
	events  []MembershipUpdateEvent
	admins  *stringSet
	members *stringSet
}

func groupChatID(creator *ecdsa.PublicKey) string {
	return uuid.New().String() + "-" + publicKeyToString(creator)
}

func NewGroupWithEvents(chatID string, events []MembershipUpdateEvent) (*Group, error) {
	return newGroup(chatID, events)
}

func NewGroupWithCreator(name string, color string, clock uint64, creator *ecdsa.PrivateKey) (*Group, error) {
	chatID := groupChatID(&creator.PublicKey)
	chatCreated := NewChatCreatedEvent(name, color, clock)
	chatCreated.ChatID = chatID
	err := chatCreated.Sign(creator)
	if err != nil {
		return nil, err
	}
	return newGroup(chatID, []MembershipUpdateEvent{chatCreated})
}

func newGroup(chatID string, events []MembershipUpdateEvent) (*Group, error) {
	g := Group{
		chatID:  chatID,
		events:  events,
		admins:  newStringSet(),
		members: newStringSet(),
	}
	if err := g.init(); err != nil {
		return nil, err
	}
	return &g, nil
}

func (g *Group) init() error {
	g.sortEvents()

	var chatID string

	for _, event := range g.events {
		if chatID == "" {
			chatID = event.ChatID
		} else if event.ChatID != chatID {
			return errors.New("updates contain different chat IDs")
		}
		valid := g.validateEvent(event)
		if !valid {
			return fmt.Errorf("invalid event %#+v from %s", event, event.From)
		}
		g.processEvent(event)
	}

	valid := g.validateChatID(g.chatID)
	if !valid {
		return fmt.Errorf("invalid chat ID: %s", g.chatID)
	}
	if chatID != g.chatID {
		return fmt.Errorf("expected chat ID equal %s, got %s", g.chatID, chatID)
	}

	return nil
}

func (g Group) ChatID() string {
	return g.chatID
}

func (g Group) Name() string {
	return g.name
}

func (g Group) Color() string {
	return g.color
}

func (g Group) Image() []byte {
	return g.image
}

func (g Group) Events() []MembershipUpdateEvent {
	return g.events
}

// AbridgedEvents returns the minimum set of events for a user to publish a post
// The events we want to keep:
// 1) Chat created
// 2) Latest color changed
// 3) Latest image changed
// 4) For each admin, the latest admins added event that contains them
// 5) For each member, the latest members added event that contains them
// 4 & 5, might bring removed admins or removed members, for those, we also need to
// keep the event that removes them
func (g Group) AbridgedEvents() []MembershipUpdateEvent {
	var events []MembershipUpdateEvent
	var nameChangedEventFound bool
	var colorChangedEventFound bool
	var imageChangedEventFound bool
	removedMembers := make(map[string]*MembershipUpdateEvent)
	addedMembers := make(map[string]bool)
	extraMembers := make(map[string]bool)
	admins := make(map[string]bool)
	// Iterate in reverse
	for i := len(g.events) - 1; i >= 0; i-- {
		event := g.events[i]
		switch event.Type {
		case protobuf.MembershipUpdateEvent_CHAT_CREATED:
			events = append(events, event)
		case protobuf.MembershipUpdateEvent_NAME_CHANGED:
			if nameChangedEventFound {
				continue
			}
			events = append(events, event)
			nameChangedEventFound = true
		case protobuf.MembershipUpdateEvent_COLOR_CHANGED:
			if colorChangedEventFound {
				continue
			}
			events = append(events, event)
			colorChangedEventFound = true
		case protobuf.MembershipUpdateEvent_IMAGE_CHANGED:
			if imageChangedEventFound {
				continue
			}
			events = append(events, event)
			imageChangedEventFound = true

		case protobuf.MembershipUpdateEvent_MEMBERS_ADDED:
			var shouldAddEvent bool
			for _, m := range event.Members {
				// If it's adding a current user, and we don't have a more
				// recent event
				// if it's an admin, we track it
				if admins[m] || (g.members.Has(m) && !addedMembers[m]) {
					addedMembers[m] = true
					shouldAddEvent = true
				}
			}
			if shouldAddEvent {
				// Append the event and check the not current members that are also
				// added
				for _, m := range event.Members {
					if !g.members.Has(m) && !admins[m] {
						extraMembers[m] = true
					}
				}
				events = append(events, event)
			}
		case protobuf.MembershipUpdateEvent_ADMIN_REMOVED:
			// We add it always for now
			events = append(events, event)
		case protobuf.MembershipUpdateEvent_ADMINS_ADDED:
			// We track admins in full
			admins[event.Members[0]] = true
			events = append(events, event)
		case protobuf.MembershipUpdateEvent_MEMBER_REMOVED:
			// Save member removed events, as we might need it
			// to remove members who have been added but subsequently left
			if removedMembers[event.Members[0]] == nil || removedMembers[event.Members[0]].ClockValue < event.ClockValue {
				removedMembers[event.Members[0]] = &event
			}

		case protobuf.MembershipUpdateEvent_MEMBER_JOINED:
			if g.members.Has(event.From) {
				events = append(events, event)
			}

		}
	}

	for m := range extraMembers {
		if removedMembers[m] != nil {
			events = append(events, *removedMembers[m])
		}
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].ClockValue < events[j].ClockValue
	})

	return events
}

func (g Group) Members() []string {
	return g.members.List()
}

func (g Group) MemberPublicKeys() ([]*ecdsa.PublicKey, error) {
	var publicKeys = make([]*ecdsa.PublicKey, 0, len(g.Members()))
	for _, memberPublicKey := range g.Members() {
		publicKey, err := hexToPubkey(memberPublicKey)
		if err != nil {
			return nil, err
		}
		publicKeys = append(publicKeys, publicKey)
	}
	return publicKeys, nil
}

func hexToPubkey(pk string) (*ecdsa.PublicKey, error) {
	bytes, err := types.DecodeHex(pk)
	if err != nil {
		return nil, err
	}
	return crypto.UnmarshalPubkey(bytes)
}

func (g Group) Admins() []string {
	return g.admins.List()
}

func (g *Group) ProcessEvents(events []MembershipUpdateEvent) error {
	for _, event := range events {
		err := g.ProcessEvent(event)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Group) ProcessEvent(event MembershipUpdateEvent) error {
	if !g.validateEvent(event) {
		return fmt.Errorf("invalid event %#+v", event)
	}
	// Check if exists
	g.events = append(g.events, event)
	g.processEvent(event)
	return nil
}

func (g Group) LastClockValue() uint64 {
	if len(g.events) == 0 {
		return 0
	}
	return g.events[len(g.events)-1].ClockValue
}

func (g Group) Creator() (string, error) {
	if len(g.events) == 0 {
		return "", errors.New("no events in the group")
	}
	first := g.events[0]
	if first.Type != protobuf.MembershipUpdateEvent_CHAT_CREATED {
		return "", fmt.Errorf("expected first event to be 'chat-created', got %s", first.Type)
	}
	return first.From, nil
}

func (g Group) isCreator(id string) (bool, error) {
	c, err := g.Creator()
	if err != nil {
		return false, err
	}

	return id == c, nil
}

func (g Group) validateChatID(chatID string) bool {
	creator, err := g.Creator()
	if err != nil || creator == "" {
		return false
	}
	// TODO: It does not verify that the prefix is a valid UUID.
	//       Improve it so that the prefix follows UUIDv4 spec.
	return strings.HasSuffix(chatID, creator) && chatID != creator
}

func (g Group) IsMember(id string) bool {
	return g.members.Has(id)
}

func (g Group) WasEverMember(id string) (bool, error) {
	isCreator, err := g.isCreator(id)
	if err != nil {
		return false, err
	}

	if isCreator {
		return true, nil
	}

	for _, event := range g.events {
		if event.Type == protobuf.MembershipUpdateEvent_MEMBERS_ADDED {
			for _, member := range event.Members {
				if member == id {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// validateEvent returns true if a given event is valid.
func (g Group) validateEvent(event MembershipUpdateEvent) bool {
	if len(event.From) == 0 {
		return false
	}
	switch event.Type {
	case protobuf.MembershipUpdateEvent_CHAT_CREATED:
		return g.admins.Empty() && g.members.Empty()
	case protobuf.MembershipUpdateEvent_NAME_CHANGED:
		return (g.admins.Has(event.From) || g.members.Has(event.From)) && len(event.Name) > 0
	case protobuf.MembershipUpdateEvent_COLOR_CHANGED:
		return (g.admins.Has(event.From) || g.members.Has(event.From)) && len(event.Color) > 0
	case protobuf.MembershipUpdateEvent_IMAGE_CHANGED:
		return (g.admins.Has(event.From) || g.members.Has(event.From)) && len(event.Image) > 0
	case protobuf.MembershipUpdateEvent_MEMBERS_ADDED:
		return g.admins.Has(event.From) || g.members.Has(event.From)
	case protobuf.MembershipUpdateEvent_MEMBER_JOINED:
		return g.members.Has(event.From)
	case protobuf.MembershipUpdateEvent_MEMBER_REMOVED:
		// Member can remove themselves or admin can remove a member.
		return len(event.Members) == 1 && (event.From == event.Members[0] || (g.admins.Has(event.From) && !g.admins.Has(event.Members[0])))
	case protobuf.MembershipUpdateEvent_ADMINS_ADDED:
		return g.admins.Has(event.From) && stringSliceSubset(event.Members, g.members.List())
	case protobuf.MembershipUpdateEvent_ADMIN_REMOVED:
		return len(event.Members) == 1 && g.admins.Has(event.From) && event.From == event.Members[0]
	default:
		return false
	}
}

func (g *Group) processEvent(event MembershipUpdateEvent) {
	switch event.Type {
	case protobuf.MembershipUpdateEvent_CHAT_CREATED:
		g.name = event.Name
		g.color = event.Color
		g.members.Add(event.From)
		g.admins.Add(event.From)
	case protobuf.MembershipUpdateEvent_NAME_CHANGED:
		g.name = event.Name
	case protobuf.MembershipUpdateEvent_COLOR_CHANGED:
		g.color = event.Color
	case protobuf.MembershipUpdateEvent_IMAGE_CHANGED:
		g.image = event.Image
	case protobuf.MembershipUpdateEvent_ADMINS_ADDED:
		g.admins.Add(event.Members...)
	case protobuf.MembershipUpdateEvent_ADMIN_REMOVED:
		g.admins.Remove(event.Members[0])
	case protobuf.MembershipUpdateEvent_MEMBERS_ADDED:
		g.members.Add(event.Members...)
	case protobuf.MembershipUpdateEvent_MEMBER_REMOVED:
		g.admins.Remove(event.Members[0])
		g.members.Remove(event.Members[0])
	}
}

func (g *Group) sortEvents() {
	sort.Slice(g.events, func(i, j int) bool {
		return g.events[i].ClockValue < g.events[j].ClockValue
	})
}

func stringSliceSubset(subset []string, set []string) bool {
	for _, item1 := range set {
		var found bool
		for _, item2 := range subset {
			if item1 == item2 {
				found = true
				break
			}
		}
		if found {
			return true
		}
	}
	return false
}

func publicKeyToString(publicKey *ecdsa.PublicKey) string {
	return types.EncodeHex(crypto.FromECDSAPub(publicKey))
}

type stringSet struct {
	m     map[string]struct{}
	items []string
}

func newStringSet() *stringSet {
	return &stringSet{
		m: make(map[string]struct{}),
	}
}

func newStringSetFromSlice(s []string) *stringSet {
	set := newStringSet()
	if len(s) > 0 {
		set.Add(s...)
	}
	return set
}

func (s *stringSet) Add(items ...string) {
	for _, item := range items {
		if _, ok := s.m[item]; !ok {
			s.m[item] = struct{}{}
			s.items = append(s.items, item)
		}
	}
}

func (s *stringSet) Remove(items ...string) {
	for _, item := range items {
		if _, ok := s.m[item]; ok {
			delete(s.m, item)
			s.removeFromItems(item)
		}
	}
}

func (s *stringSet) Has(item string) bool {
	_, ok := s.m[item]
	return ok
}

func (s *stringSet) Empty() bool {
	return len(s.items) == 0
}

func (s *stringSet) List() []string {
	return s.items
}

func (s *stringSet) removeFromItems(dropped string) {
	n := 0
	for _, item := range s.items {
		if item != dropped {
			s.items[n] = item
			n++
		}
	}
	s.items = s.items[:n]
}
