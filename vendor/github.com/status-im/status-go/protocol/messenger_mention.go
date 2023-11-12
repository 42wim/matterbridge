package protocol

// this is a reimplementation of the mention feature in status-react
// reference implementation: https://github.com/status-im/status-react/blob/972347963498fc4a2bb8f85541e79ed0541698da/src/status_im/chat/models/mentions.cljs#L1
import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"go.uber.org/zap"

	"github.com/status-im/status-go/logutils"

	"github.com/status-im/status-go/api/multiformat"
	"github.com/status-im/status-go/protocol/common"
)

const (
	endingChars = `[\s\.,;:]`

	charAtSign     = "@"
	charQuote      = ">"
	charNewline    = "\n"
	charAsterisk   = "*"
	charUnderscore = "_"
	charTilde      = "~"
	charCodeBlock  = "`"

	intUnknown = -1
)

var (
	specialCharsRegex = regexp.MustCompile("[@~\\\\*_\n>`]{1}")
	endingCharsRegex  = regexp.MustCompile(endingChars)
	wordRegex         = regexp.MustCompile("^[\\w\\d\\-]*" + endingChars + "|[\\S]+")
)

type specialCharLocation struct {
	Index int
	Value string
}

type atSignIndex struct {
	Pending []int
	Checked []int
}

type styleTag struct {
	Len int
	Idx int
}

type textMeta struct {
	atSign         *atSignIndex
	styleTagMap    map[string]*styleTag
	quoteIndex     *int
	newlineIndexes []int
}

type searchablePhrase struct {
	originalName string
	phrase       string
}

type MentionableUser struct {
	*Contact

	searchablePhrases []searchablePhrase

	Key          string // a unique identifier of a mentionable user
	Match        string
	SearchedText string
}

func (c *MentionableUser) MarshalJSON() ([]byte, error) {
	compressedKey, err := multiformat.SerializeLegacyKey(c.ID)
	if err != nil {
		return nil, err
	}

	type MentionableUserJSON struct {
		PrimaryName   string `json:"primaryName"`
		SecondaryName string `json:"secondaryName"`
		ENSVerified   bool   `json:"ensVerified"`
		Added         bool   `json:"added"`
		DisplayName   string `json:"displayName"`
		Key           string `json:"key"`
		Match         string `json:"match"`
		SearchedText  string `json:"searchedText"`
		ID            string `json:"id"`
		CompressedKey string `json:"compressedKey,omitempty"`
	}

	contactJSON := MentionableUserJSON{
		PrimaryName:   c.PrimaryName(),
		SecondaryName: c.SecondaryName(),
		ENSVerified:   c.ENSVerified,
		Added:         c.added(),
		DisplayName:   c.GetDisplayName(),
		Key:           c.Key,
		Match:         c.Match,
		SearchedText:  c.SearchedText,
		ID:            c.ID,
		CompressedKey: compressedKey,
	}

	return json.Marshal(contactJSON)
}

func (c *MentionableUser) GetDisplayName() string {
	if c.ENSVerified && c.EnsName != "" {
		return c.EnsName
	}
	if c.DisplayName != "" {
		return c.DisplayName
	}
	if c.PrimaryName() != "" {
		return c.PrimaryName()
	}
	return c.Alias
}

type SegmentType int

const (
	Text SegmentType = iota
	Mention
)

type InputSegment struct {
	Type  SegmentType `json:"type"`
	Value string      `json:"value"`
}

type MentionState struct {
	AtSignIdx    int
	AtIdxs       []*AtIndexEntry
	MentionEnd   int
	PreviousText string
	NewText      string
	Start        int
	End          int
	operation    textOperation
}

func (ms *MentionState) String() string {
	atIdxsStr := ""
	for i, entry := range ms.AtIdxs {
		if i > 0 {
			atIdxsStr += ", "
		}
		atIdxsStr += fmt.Sprintf("%+v", entry)
	}
	return fmt.Sprintf("MentionState{AtSignIdx: %d, AtIdxs: [%s], MentionEnd: %d, PreviousText: %q, NewText: %s, Start: %d, End: %d}",
		ms.AtSignIdx, atIdxsStr, ms.MentionEnd, ms.PreviousText, ms.NewText, ms.Start, ms.End)
}

type ChatMentionContext struct {
	ChatID             string
	InputSegments      []InputSegment
	MentionSuggestions map[string]*MentionableUser
	MentionState       *MentionState
	PreviousText       string // user input text before the last change
	NewText            string
}

func NewChatMentionContext(chatID string) *ChatMentionContext {
	return &ChatMentionContext{
		ChatID:             chatID,
		MentionSuggestions: make(map[string]*MentionableUser),
		MentionState:       new(MentionState),
	}
}

type mentionableUserGetter interface {
	getMentionableUsers(chatID string) (map[string]*MentionableUser, error)
	getMentionableUser(chatID string, pk string) (*MentionableUser, error)
}

type MentionManager struct {
	mentionContexts map[string]*ChatMentionContext
	*Messenger
	mentionableUserGetter
	logger *zap.Logger
}

func NewMentionManager(m *Messenger) *MentionManager {
	mm := &MentionManager{
		mentionContexts: make(map[string]*ChatMentionContext),
		Messenger:       m,
		logger:          logutils.ZapLogger().Named("MentionManager"),
	}
	mm.mentionableUserGetter = mm
	return mm
}

func (m *MentionManager) getChatMentionContext(chatID string) *ChatMentionContext {
	ctx, ok := m.mentionContexts[chatID]
	if !ok {
		ctx = NewChatMentionContext(chatID)
		m.mentionContexts[chatID] = ctx
	}
	return ctx
}

func (m *MentionManager) getMentionableUser(chatID string, pk string) (*MentionableUser, error) {
	mentionableUsers, err := m.mentionableUserGetter.getMentionableUsers(chatID)
	if err != nil {
		return nil, err
	}
	user, ok := mentionableUsers[pk]
	if !ok {
		return nil, fmt.Errorf("user not found when getting mentionable user, pk: %s", pk)
	}
	return user, nil
}

func (m *MentionManager) getMentionableUsers(chatID string) (map[string]*MentionableUser, error) {
	mentionableUsers := make(map[string]*MentionableUser)
	chat, _ := m.allChats.Load(chatID)
	if chat == nil {
		return nil, fmt.Errorf("chat not found when getting mentionable users, chatID: %s", chatID)
	}

	var publicKeys []string
	switch {
	case chat.PrivateGroupChat():
		for _, mb := range chat.Members {
			publicKeys = append(publicKeys, mb.ID)
		}
	case chat.OneToOne():
		publicKeys = append(publicKeys, chatID)
	case chat.CommunityChat():
		community, err := m.communitiesManager.GetByIDString(chat.CommunityID)
		if err != nil {
			return nil, err
		}
		for _, pk := range community.GetMemberPubkeys() {
			publicKeys = append(publicKeys, common.PubkeyToHex(pk))
		}
	case chat.Public():
		m.allContacts.Range(func(pk string, _ *Contact) bool {
			publicKeys = append(publicKeys, pk)
			return true
		})
	}

	var me = m.myHexIdentity()
	for _, pk := range publicKeys {
		if pk == me {
			continue
		}
		if err := m.addMentionableUser(mentionableUsers, pk); err != nil {
			return nil, err
		}
	}
	return mentionableUsers, nil
}

func (m *MentionManager) addMentionableUser(mentionableUsers map[string]*MentionableUser, publicKey string) error {
	contact, ok := m.allContacts.Load(publicKey)
	if !ok {
		c, err := buildContactFromPkString(publicKey)
		if err != nil {
			return err
		}
		contact = c
	}
	user := &MentionableUser{
		Contact: contact,
	}
	user = addSearchablePhrases(user)
	if user != nil {
		mentionableUsers[publicKey] = user
	}
	return nil
}

func (m *MentionManager) ReplaceWithPublicKey(chatID, text string) (string, error) {
	chat, _ := m.allChats.Load(chatID)
	if chat == nil {
		return "", fmt.Errorf("chat not found when check mentions, chatID: %s", chatID)
	}
	mentionableUsers, err := m.mentionableUserGetter.getMentionableUsers(chatID)
	if err != nil {
		return "", err
	}
	newText := ReplaceMentions(text, mentionableUsers)
	m.ClearMentions(chatID)
	return newText, nil
}

func (m *MentionManager) OnChangeText(chatID, fullText string) (*ChatMentionContext, error) {
	ctx := m.getChatMentionContext(chatID)
	diff := diffText(ctx.PreviousText, fullText)
	if diff == nil {
		return ctx, nil
	}
	ctx.PreviousText = fullText
	if ctx.MentionState == nil {
		ctx.MentionState = &MentionState{}
	}
	ctx.MentionState.PreviousText = diff.previousText
	ctx.MentionState.NewText = diff.newText
	ctx.MentionState.Start = diff.start
	ctx.MentionState.End = diff.end
	ctx.MentionState.operation = diff.operation

	atIndexes, err := calculateAtIndexEntries(ctx.MentionState)
	if err != nil {
		return ctx, err
	}
	ctx.MentionState.AtIdxs = atIndexes
	m.logger.Debug("OnChangeText", zap.String("chatID", chatID), zap.Any("state", ctx.MentionState))
	return m.calculateSuggestions(chatID, fullText)
}

func (m *MentionManager) calculateSuggestions(chatID, fullText string) (*ChatMentionContext, error) {
	ctx := m.getChatMentionContext(chatID)

	mentionableUsers, err := m.mentionableUserGetter.getMentionableUsers(chatID)
	if err != nil {
		return nil, err
	}
	m.logger.Debug("calculateSuggestions", zap.String("chatID", chatID), zap.String("fullText", fullText), zap.Int("num of mentionable user", len(mentionableUsers)))

	m.calculateSuggestionsWithMentionableUsers(chatID, fullText, mentionableUsers)

	return ctx, nil
}

func (m *MentionManager) calculateSuggestionsWithMentionableUsers(chatID string, fullText string, mentionableUsers map[string]*MentionableUser) {
	ctx := m.getChatMentionContext(chatID)
	state := ctx.MentionState

	if len(state.AtIdxs) == 0 {
		state.AtIdxs = nil
		ctx.MentionSuggestions = nil
		ctx.InputSegments = []InputSegment{{
			Type:  Text,
			Value: fullText,
		}}
		return
	}

	newAtIndexEntries := checkIdxForMentions(fullText, state.AtIdxs, mentionableUsers)
	calculatedInput, success := calculateInput(fullText, newAtIndexEntries)
	if !success {
		m.logger.Warn("calculateSuggestionsWithMentionableUsers: calculateInput failed", zap.String("chatID", chatID), zap.String("fullText", fullText), zap.Any("state", state))
	}

	var end int
	switch state.operation {
	case textOperationAdd:
		end = state.Start + len([]rune(state.NewText))
	case textOperationDelete:
		end = state.Start
	case textOperationReplace:
		end = state.Start + len([]rune(state.NewText))
	default:
		m.logger.Error("calculateSuggestionsWithMentionableUsers: unknown textOperation", zap.String("chatID", chatID), zap.String("fullText", fullText), zap.Any("state", state))
	}

	atSignIdx := lastIndexOfAtSign(fullText, end)
	var suggestions map[string]*MentionableUser
	if atSignIdx != -1 {
		searchedText := strings.ToLower(subs(fullText, atSignIdx+1, end))
		m.logger.Debug("calculateSuggestionsWithMentionableUsers", zap.Int("atSignIdx", atSignIdx), zap.String("searchedText", searchedText), zap.String("fullText", fullText), zap.Any("state", state), zap.Int("end", end))
		if end-atSignIdx <= 100 {
			suggestions = getUserSuggestions(mentionableUsers, searchedText, -1)
		}
	}

	state.AtSignIdx = atSignIdx
	state.AtIdxs = newAtIndexEntries
	state.MentionEnd = end
	ctx.InputSegments = calculatedInput
	ctx.MentionSuggestions = suggestions
}

func (m *MentionManager) SelectMention(chatID, text, primaryName, publicKey string) (*ChatMentionContext, error) {
	ctx := m.getChatMentionContext(chatID)
	state := ctx.MentionState

	atSignIdx := state.AtSignIdx
	mentionEnd := state.MentionEnd

	var nextChar rune
	tr := []rune(text)
	if mentionEnd < len(tr) {
		nextChar = tr[mentionEnd]
	}

	space := ""
	if string(nextChar) == "" || (!unicode.IsSpace(nextChar)) {
		space = " "
	}

	ctx.NewText = string(tr[:atSignIdx+1]) + primaryName + space + string(tr[mentionEnd:])

	ctx, err := m.OnChangeText(chatID, ctx.NewText)
	if err != nil {
		return nil, err
	}
	m.clearSuggestions(chatID)
	return ctx, nil
}

func (m *MentionManager) clearSuggestions(chatID string) {
	m.getChatMentionContext(chatID).MentionSuggestions = nil
}

func (m *MentionManager) ClearMentions(chatID string) {
	ctx := m.getChatMentionContext(chatID)
	ctx.MentionState = nil
	ctx.InputSegments = nil
	ctx.NewText = ""
	ctx.PreviousText = ""
	m.clearSuggestions(chatID)
}

func (m *MentionManager) ToInputField(chatID, text string) (*ChatMentionContext, error) {
	mentionableUsers, err := m.mentionableUserGetter.getMentionableUsers(chatID)
	if err != nil {
		return nil, err
	}
	textWithMentions := toInputField(text)
	newText := ""
	for i, segment := range textWithMentions {
		if segment.Type == Mention {
			mentionableUser := mentionableUsers[segment.Value]
			mention := mentionableUser.GetDisplayName()
			if !strings.HasPrefix(mention, charAtSign) {
				segment.Value = charAtSign + mention
			}
			textWithMentions[i] = segment
		}
		newText += segment.Value
	}
	ctx := m.getChatMentionContext(chatID)
	ctx.InputSegments = textWithMentions
	ctx.MentionState = toInfo(textWithMentions)
	ctx.NewText = newText
	ctx.PreviousText = newText
	return ctx, nil
}

func rePos(s string) []specialCharLocation {
	var res []specialCharLocation
	lastMatch := specialCharsRegex.FindStringIndex(s)
	for lastMatch != nil {
		start, end := lastMatch[0], lastMatch[1]
		c := s[start:end]
		res = append(res, specialCharLocation{utf8.RuneCountInString(s[:start]), c})
		lastMatch = specialCharsRegex.FindStringIndex(s[end:])
		if lastMatch != nil {
			lastMatch[0] += end
			lastMatch[1] += end
		}
	}
	return res
}

func codeTagLen(idxs []specialCharLocation, idx int) int {
	pos, c := idxs[idx].Index, idxs[idx].Value

	next := func(n int) (int, string) {
		if n < len(idxs) {
			return idxs[n].Index, idxs[n].Value
		}
		return 0, ""
	}

	pos2, c2 := next(idx + 1)
	pos3, c3 := next(idx + 2)

	if c == c2 && pos == pos2-1 && c2 == c3 && pos == pos3-2 {
		return 3
	}

	if c == c2 && pos == pos2-1 {
		return 2
	}

	return 1
}

func clearPendingAtSigns(data *textMeta, from int) {
	newIdxs := make([]int, 0)
	for _, idx := range data.atSign.Pending {
		if idx < from {
			newIdxs = append(newIdxs, idx)
		}
	}
	data.atSign.Pending = []int{}
	data.atSign.Checked = append(data.atSign.Checked, newIdxs...)
}

func checkStyleTag(text string, idxs []specialCharLocation, idx int) (length int, canBeStart bool, canBeEnd bool) {
	pos, c := idxs[idx].Index, idxs[idx].Value
	tr := []rune(text)
	next := func(n int) (int, string) {
		if n < len(idxs) {
			return idxs[n].Index, idxs[n].Value
		}
		return len(tr), ""
	}

	pos2, c2 := next(idx + 1)
	pos3, c3 := next(idx + 2)

	switch {
	case c == c2 && c2 == c3 && pos == pos2-1 && pos == pos3-2:
		length = 3
	case c == c2 && pos == pos2-1:
		length = 2
	default:
		length = 1
	}

	var prevC, nextC *rune
	if decPos := pos - 1; decPos >= 0 {
		prevC = &tr[decPos]
	}

	nextIdx := idxs[idx+length-1].Index + 1
	if nextIdx < len(tr) {
		nextC = &tr[nextIdx]
	}

	if length == 1 {
		canBeEnd = prevC != nil && !unicode.IsSpace(*prevC) && (nextC == nil || unicode.IsSpace(*nextC))
	} else {
		canBeEnd = prevC != nil && !unicode.IsSpace(*prevC)
	}

	canBeStart = nextC != nil && !unicode.IsSpace(*nextC)
	return length, canBeStart, canBeEnd
}

func applyStyleTag(data *textMeta, idx int, pos int, c string, len int, start bool, end bool) int {
	tag := data.styleTagMap[c]
	tripleTilde := c == charTilde && len == 3

	if tag != nil && end {
		oldLen := (*tag).Len
		var tagLen int
		if tripleTilde && oldLen == 3 {
			tagLen = 2
		} else if oldLen >= len {
			tagLen = len
		} else {
			tagLen = oldLen
		}
		oldIdx := (*tag).Idx
		delete(data.styleTagMap, c)
		clearPendingAtSigns(data, oldIdx)
		return idx + tagLen
	} else if start {
		data.styleTagMap[c] = &styleTag{
			Len: len,
			Idx: pos,
		}
		clearPendingAtSigns(data, pos)
	}
	return idx + len
}

func newTextMeta() *textMeta {
	return &textMeta{
		atSign:      new(atSignIndex),
		styleTagMap: make(map[string]*styleTag),
	}
}

func newDataWithAtSignAndQuoteIndex(atSign *atSignIndex, quoteIndex *int) *textMeta {
	data := newTextMeta()
	data.atSign = atSign
	data.quoteIndex = quoteIndex
	return data
}

func getAtSigns(text string) []int {
	idxs := rePos(text)
	data := newTextMeta()
	nextIdx := 0
	tr := []rune(text)
	for i := range idxs {
		if i != nextIdx {
			continue
		}
		nextIdx = i + 1
		quoteIndex := data.quoteIndex
		c := idxs[i].Value
		pos := idxs[i].Index
		switch {
		case c == charNewline:
			prevNewline := intUnknown
			if len(data.newlineIndexes) > 0 {
				prevNewline = data.newlineIndexes[0]
			}

			data.newlineIndexes = append(data.newlineIndexes, pos)

			if quoteIndex != nil && prevNewline != intUnknown && strings.TrimSpace(string(tr[prevNewline:pos-1])) == "" {
				data.quoteIndex = nil
			}
		case quoteIndex != nil:
			continue
		case c == charQuote:
			prevNewlines := make([]int, 0, 2)
			if len(data.newlineIndexes) > 0 {
				prevNewlines = data.newlineIndexes
			}

			if pos == 0 ||
				(len(prevNewlines) == 1 && strings.TrimSpace(string(tr[:pos-1])) == "") ||
				(len(prevNewlines) == 2 && strings.TrimSpace(string(tr[prevNewlines[0]:pos-1])) == "") {
				data = newDataWithAtSignAndQuoteIndex(data.atSign, &pos)
			}
		case c == charAtSign:
			data.atSign.Pending = append(data.atSign.Pending, pos)
		case c == charCodeBlock:
			length := codeTagLen(idxs, i)
			nextIdx = applyStyleTag(data, i, pos, c, length, true, true)
		case c == charAsterisk || c == charUnderscore || c == charTilde:
			length, canBeStart, canBeEnd := checkStyleTag(text, idxs, i)
			nextIdx = applyStyleTag(data, i, pos, c, length, canBeStart, canBeEnd)
		}
	}

	return append(data.atSign.Checked, data.atSign.Pending...)
}

func getUserSuggestions(users map[string]*MentionableUser, searchedText string, limit int) map[string]*MentionableUser {
	result := make(map[string]*MentionableUser)
	for pk, user := range users {
		match := findMatch(user, searchedText)
		if match != "" {
			result[pk] = &MentionableUser{
				searchablePhrases: user.searchablePhrases,
				Contact:           user.Contact,
				Key:               pk,
				Match:             match,
				SearchedText:      searchedText,
			}
		}
		if limit != -1 && len(result) >= limit {
			break
		}
	}
	return result
}

// findMatch searches for a matching phrase in MentionableUser's searchable phrases or names.
func findMatch(user *MentionableUser, searchedText string) string {
	if len(user.searchablePhrases) > 0 {
		return findMatchInPhrases(user, searchedText)
	}
	return findMatchInNames(user, searchedText)
}

// findMatchInPhrases searches for a matching phrase in MentionableUser's searchable phrases.
func findMatchInPhrases(user *MentionableUser, searchedText string) string {
	var match string

	for _, p := range user.searchablePhrases {
		if searchedText == "" || strings.HasPrefix(strings.ToLower(p.phrase), searchedText) {
			match = p.originalName
			break
		}
	}

	return match
}

// findMatchInNames searches for a matching phrase in MentionableUser's primary and secondary names.
func findMatchInNames(user *MentionableUser, searchedText string) string {
	var match string
	for _, name := range user.names() {
		if hasMatchingPrefix(name, searchedText) {
			match = name
		}
	}
	return match
}

// hasMatchingPrefix checks if the given text has a matching prefix with the searched text.
func hasMatchingPrefix(text, searchedText string) bool {
	return text != "" && (searchedText == "" || strings.HasPrefix(strings.ToLower(text), searchedText))
}

func isMentioned(user *MentionableUser, text string) bool {
	regexStr := ""
	for i, name := range user.names() {
		if name == "" {
			continue
		}
		name = strings.ToLower(name)
		if i != 0 {
			regexStr += "|"
		}
		regexStr += "^" + name + endingChars + "|" + "^" + name + "$"
	}
	regex := regexp.MustCompile(regexStr)
	lCaseText := strings.ToLower(text)
	return regex.MatchString(lCaseText)
}

func MatchMention(text string, users map[string]*MentionableUser, mentionKeyIdx int) *MentionableUser {
	return matchMention(text, users, mentionKeyIdx, mentionKeyIdx+1, nil)
}

func matchMention(text string, users map[string]*MentionableUser, mentionKeyIdx int, nextWordIdx int, words []string) *MentionableUser {
	tr := []rune(text)
	if nextWordIdx >= len(tr) {
		return nil
	}
	if word := wordRegex.FindString(string(tr[nextWordIdx:])); word != "" {
		newWords := append(words, word)

		t := strings.TrimSpace(strings.ToLower(strings.Join(newWords, "")))
		tt := []rune(t)
		searchedText := t
		if lastChar := len(tt) - 1; lastChar >= 0 && endingCharsRegex.MatchString(string(tt[lastChar:])) {
			searchedText = string(tt[:lastChar])
		}

		userSuggestions := getUserSuggestions(users, searchedText, -1)
		userSuggestionsCnt := len(userSuggestions)
		switch {
		case userSuggestionsCnt == 0:
			return nil
		case userSuggestionsCnt == 1:
			user := getFirstUser(userSuggestions)
			// maybe len(users) == 1 and user input `@` so we need to recheck if the user is really mentioned
			if isMentioned(user, string(tr[mentionKeyIdx+1:])) {
				return user
			}
		case userSuggestionsCnt > 1:
			wordLen := len([]rune(word))
			textLen := len(tr)
			nextWordStart := nextWordIdx + wordLen
			if textLen > nextWordStart {
				user := matchMention(text, users, mentionKeyIdx, nextWordStart, newWords)
				if user != nil {
					return user
				}
			}
			return filterWithFullMatch(userSuggestions, searchedText)
		}
	}
	return nil
}

func filterWithFullMatch(userSuggestions map[string]*MentionableUser, text string) *MentionableUser {
	if text == "" {
		return nil
	}
	result := make(map[string]*MentionableUser)
	for pk, user := range userSuggestions {
		for _, name := range user.names() {
			if strings.ToLower(name) == text {
				result[pk] = user
			}
		}
	}
	return getFirstUser(result)
}

func getFirstUser(userSuggestions map[string]*MentionableUser) *MentionableUser {
	for _, user := range userSuggestions {
		return user
	}
	return nil
}

func ReplaceMentions(text string, users map[string]*MentionableUser) string {
	idxs := getAtSigns(text)
	return replaceMentions(text, users, idxs, 0)
}

func replaceMentions(text string, users map[string]*MentionableUser, idxs []int, diff int) string {
	if strings.TrimSpace(text) == "" || len(idxs) == 0 {
		return text
	}

	mentionKeyIdx := idxs[0] - diff

	if len(users) == 0 {
		return text
	}

	matchUser := MatchMention(text, users, mentionKeyIdx)
	if matchUser == nil {
		return replaceMentions(text, users, idxs[1:], diff)
	}

	tr := []rune(text)
	newText := string(tr[:mentionKeyIdx+1]) + matchUser.ID + string(tr[mentionKeyIdx+1+len([]rune(matchUser.Match)):])
	newDiff := diff + len(tr) - len([]rune(newText))

	return replaceMentions(newText, users, idxs[1:], newDiff)
}

func addSearchablePhrases(user *MentionableUser) *MentionableUser {
	if !user.Blocked {
		searchablePhrases := user.names()
		for _, s := range searchablePhrases {
			if s != "" {
				newWords := []string{s}
				newWords = append(newWords, strings.Split(s, " ")[1:]...)
				var phrases []searchablePhrase
				for _, w := range newWords {
					phrases = append(phrases, searchablePhrase{s, w})
				}
				user.searchablePhrases = append(user.searchablePhrases, phrases...)
			}
		}
		return user
	}
	return nil
}

type AtIndexEntry struct {
	From    int
	To      int
	Checked bool

	Mentioned bool
	NextAtIdx int
}

func (e *AtIndexEntry) String() string {
	return fmt.Sprintf("{From: %d, To: %d, Checked: %t, Mentioned: %t, NextAtIdx: %d}", e.From, e.To, e.Checked, e.Mentioned, e.NextAtIdx)
}

func calculateAtIndexEntries(state *MentionState) ([]*AtIndexEntry, error) {
	var keptAtIndexEntries []*AtIndexEntry
	var oldRunes []rune
	var newRunes []rune
	var previousRunes = []rune(state.PreviousText)
	switch state.operation {
	case textOperationAdd:
		newRunes = []rune(state.NewText)
	case textOperationDelete:
		oldRunes = previousRunes[state.Start : state.End+1]
	case textOperationReplace:
		oldRunes = previousRunes[state.Start : state.End+1]
		newRunes = []rune(state.NewText)
	default:
		return nil, fmt.Errorf("unknown text operation: %d", state.operation)
	}

	oldLen := len(oldRunes)
	newLen := len(newRunes)
	diff := newLen - oldLen
	oldAtSignIndexes := getAtSignIdxs(string(oldRunes), state.Start)
	newAtSignIndexes := getAtSignIdxs(state.NewText, state.Start)
	for _, entry := range state.AtIdxs {
		deleted := false
		for _, idx := range oldAtSignIndexes {
			if idx == entry.From {
				deleted = true
			}
		}
		if !deleted {
			if entry.From >= state.Start { // update range with diff
				entry.From += diff
				entry.To += diff
			}
			if entry.From < state.Start && entry.To+1 >= state.Start { // impacted after user edit so need to be rechecked
				entry.Checked = false
			}
			keptAtIndexEntries = append(keptAtIndexEntries, entry)
		}
	}
	return addNewAtSignIndexes(keptAtIndexEntries, newAtSignIndexes), nil
}

func addNewAtSignIndexes(keptAtIdxs []*AtIndexEntry, newAtSignIndexes []int) []*AtIndexEntry {
	var newAtIndexEntries []*AtIndexEntry
	var added bool
	var lastNewIdx int
	newAtSignIndexesCount := len(newAtSignIndexes)
	if newAtSignIndexesCount > 0 {
		lastNewIdx = newAtSignIndexes[newAtSignIndexesCount-1]
	}
	for _, entry := range keptAtIdxs {
		if newAtSignIndexesCount > 0 && !added && entry.From > lastNewIdx {
			newAtIndexEntries = append(newAtIndexEntries, makeAtIdxs(newAtSignIndexes)...)
			newAtIndexEntries = append(newAtIndexEntries, entry)
			added = true
		} else {
			newAtIndexEntries = append(newAtIndexEntries, entry)
		}
	}
	if !added {
		newAtIndexEntries = append(newAtIndexEntries, makeAtIdxs(newAtSignIndexes)...)
	}
	return newAtIndexEntries
}

func makeAtIdxs(idxs []int) []*AtIndexEntry {
	result := make([]*AtIndexEntry, len(idxs))
	for i, idx := range idxs {
		result[i] = &AtIndexEntry{
			From:    idx,
			Checked: false,
		}
	}
	return result
}

// getAtSignIdxs returns the indexes of all @ signs in the text.
// delta is the offset of the text within the original text.
func getAtSignIdxs(text string, delta int) []int {
	return getAtSignIdxsHelper(text, delta, 0, []int{})
}

func getAtSignIdxsHelper(text string, delta int, from int, idxs []int) []int {
	tr := []rune(text)
	idx := strings.IndexRune(string(tr[from:]), '@')
	if idx != -1 {
		idx = utf8.RuneCountInString(text[:idx])
		idx += from
		idxs = append(idxs, delta+idx)
		return getAtSignIdxsHelper(text, delta, idx+1, idxs)
	}
	return idxs
}

func checkAtIndexEntry(fullText string, entry *AtIndexEntry, mentionableUsers map[string]*MentionableUser) *AtIndexEntry {
	if entry.Checked {
		return entry
	}
	result := MatchMention(fullText, mentionableUsers, entry.From)
	if result != nil && result.Match != "" {
		return &AtIndexEntry{
			From:      entry.From,
			To:        entry.From + len([]rune(result.Match)),
			Checked:   true,
			Mentioned: true,
		}
	}
	return &AtIndexEntry{
		From:    entry.From,
		To:      len([]rune(fullText)),
		Checked: true,
	}
}

func checkIdxForMentions(fullText string, currentAtIndexEntries []*AtIndexEntry, mentionableUsers map[string]*MentionableUser) []*AtIndexEntry {
	var newIndexEntries []*AtIndexEntry
	for _, entry := range currentAtIndexEntries {
		previousEntryIdx := len(newIndexEntries) - 1
		newEntry := checkAtIndexEntry(fullText, entry, mentionableUsers)
		if previousEntryIdx >= 0 && !newIndexEntries[previousEntryIdx].Mentioned {
			newIndexEntries[previousEntryIdx].To = entry.From - 1
		}
		if previousEntryIdx >= 0 {
			newIndexEntries[previousEntryIdx].NextAtIdx = entry.From
		}
		newEntry.NextAtIdx = intUnknown
		newIndexEntries = append(newIndexEntries, newEntry)
	}

	if len(newIndexEntries) > 0 {
		lastIdx := len(newIndexEntries) - 1
		if newIndexEntries[lastIdx].Mentioned {
			return newIndexEntries
		}
		newIndexEntries[lastIdx].To = len([]rune(fullText)) - 1
		newIndexEntries[lastIdx].Checked = false
		return newIndexEntries
	}

	return nil
}

func appendInputSegment(result *[]InputSegment, typ SegmentType, value string, fullText *string) {
	if value != "" {
		*result = append(*result, InputSegment{Type: typ, Value: value})
		*fullText += value
	}
}

func calculateInput(text string, atIndexEntries []*AtIndexEntry) ([]InputSegment, bool) {
	if len(atIndexEntries) == 0 {
		return []InputSegment{{Type: Text, Value: text}}, true
	}
	idxCount := len(atIndexEntries)
	lastFrom := atIndexEntries[idxCount-1].From

	var result []InputSegment
	fullText := ""
	if atIndexEntries[0].From != 0 {
		t := subs(text, 0, atIndexEntries[0].From)
		appendInputSegment(&result, Text, t, &fullText)
	}

	for _, entry := range atIndexEntries {
		from := entry.From
		to := entry.To
		nextAtIdx := entry.NextAtIdx
		mentioned := entry.Mentioned

		if mentioned && nextAtIdx != intUnknown {
			t := subs(text, from, to+1)
			appendInputSegment(&result, Mention, t, &fullText)

			t = subs(text, to+1, nextAtIdx)
			appendInputSegment(&result, Text, t, &fullText)
		} else if mentioned && lastFrom == from {
			t := subs(text, from, to+1)
			appendInputSegment(&result, Mention, t, &fullText)

			t = subs(text, to+1)
			appendInputSegment(&result, Text, t, &fullText)
		} else {
			t := subs(text, from, to+1)
			appendInputSegment(&result, Text, t, &fullText)
		}
	}

	return result, fullText == text
}

func subs(s string, start int, end ...int) string {
	tr := []rune(s)
	e := len(tr)

	if len(end) > 0 {
		e = end[0]
	}

	if start < 0 {
		start = 0
	}

	if e > len(tr) {
		e = len(tr)
	}

	if e < 0 {
		e = 0
	}

	if start > e {
		start, e = e, start
		if e > len(tr) {
			e = len(tr)
		}
	}

	return string(tr[start:e])
}

func isValidTerminatingCharacter(c rune) bool {
	switch c {
	case '\t': // tab
		return true
	case '\n': // newline
		return true
	case '\f': // new page
		return true
	case '\r': // carriage return
		return true
	case ' ': // whitespace
		return true
	case ',':
		return true
	case '.':
		return true
	case ':':
		return true
	case ';':
		return true
	default:
		return false
	}
}

var hexReg = regexp.MustCompile("[0-9a-f]")

func isPublicKeyCharacter(c rune) bool {
	return hexReg.MatchString(string(c))
}

const mentionLength = 133

func toInputField(text string) []InputSegment {
	// Initialize the variables
	currentMentionLength := 0
	currentText := ""
	currentMention := ""
	var inputFieldEntries []InputSegment

	// Iterate through each character in the input text
	for _, character := range text {
		isPKCharacter := isPublicKeyCharacter(character)
		isTerminationCharacter := isValidTerminatingCharacter(character)

		switch {
		// It's a valid mention.
		// Add any text that is before if present
		// and add the mention.
		// Set the text to the new termination character
		case currentMentionLength == mentionLength && isTerminationCharacter:
			if currentText != "" {
				inputFieldEntries = append(inputFieldEntries, InputSegment{Type: Text, Value: currentText})
			}
			inputFieldEntries = append(inputFieldEntries, InputSegment{Type: Mention, Value: currentMention})
			currentMentionLength = 0
			currentMention = ""
			currentText = string(character)

		// It's either a pk character, or the `x` in the pk
		// in this case add the text to the mention and continue
		case (isPKCharacter && currentMentionLength > 0) || (currentMentionLength == 2 && character == 'x'):
			currentMentionLength++
			currentMention += string(character)

		// The beginning of a mention, discard the @ sign
		// and start following a mention
		case character == '@':
			currentMentionLength = 1
			currentMention = ""

		// Not a mention character, but we were following a mention
		// discard everything up to now and count as text
		case !isPKCharacter && currentMentionLength > 0:
			currentText += "@" + currentMention + string(character)
			currentMentionLength = 0
			currentMention = ""

		// Just a normal text character
		default:
			currentText += string(character)
		}
	}

	// Process any remaining mention/text
	if currentText != "" {
		inputFieldEntries = append(inputFieldEntries, InputSegment{Type: Text, Value: currentText})
	}
	if currentMentionLength == mentionLength {
		inputFieldEntries = append(inputFieldEntries, InputSegment{Type: Mention, Value: currentMention})
	}

	return inputFieldEntries
}

func toInfo(inputSegments []InputSegment) *MentionState {
	newText := ""
	state := &MentionState{
		AtSignIdx:    intUnknown,
		End:          intUnknown,
		AtIdxs:       []*AtIndexEntry{},
		MentionEnd:   0,
		PreviousText: "",
		NewText:      newText,
		Start:        intUnknown,
	}

	for _, segment := range inputSegments {
		t := segment.Type
		text := segment.Value
		tr := []rune(text)

		if t == Mention {
			newMention := &AtIndexEntry{
				Checked:   true,
				Mentioned: true,
				From:      state.MentionEnd,
				To:        state.Start + len(tr),
			}

			if len(state.AtIdxs) > 0 {
				lastIdx := state.AtIdxs[len(state.AtIdxs)-1]
				state.AtIdxs = state.AtIdxs[:len(state.AtIdxs)-1]
				lastIdx.NextAtIdx = state.MentionEnd
				state.AtIdxs = append(state.AtIdxs, lastIdx)
			}
			state.AtIdxs = append(state.AtIdxs, newMention)
			state.AtSignIdx = state.MentionEnd
		}

		state.MentionEnd += len(tr)
		state.NewText = string(tr[len(tr)-1])
		state.Start += len(tr)
		state.End += len(tr)
	}

	return state
}

// lastIndexOfAtSign returns the index of the last occurrence of substr in s starting from index start.
// If substr is not present in s, it returns -1.
func lastIndexOfAtSign(s string, start int) int {
	if start < 0 {
		return -1
	}

	t := []rune(s)
	if start >= len(t) {
		start = len(t) - 1
	}

	// Reverse the input strings to find the first occurrence of the reversed substr in the reversed s.
	reversedS := reverse(t[:start+1])

	idx := strings.IndexRune(reversedS, '@')

	if idx == -1 {
		return -1
	}

	// Calculate the index in the original string.
	idx = utf8.RuneCountInString(reversedS[:idx])
	return start - idx
}

// reverse returns the reversed string of input s.
func reverse(r []rune) string {
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

type textOperation int

const (
	textOperationAdd textOperation = iota + 1
	textOperationDelete
	textOperationReplace
)

type TextDiff struct {
	previousText string
	newText      string // if add operation, newText is the added text; if replace operation, newText is the text used to replace the previousText
	start        int    // start index of the operation relate to previousText
	end          int    // end index of the operation relate to previousText, always the same as start if the operation is add, range: start<=end<=len(previousText)-1
	operation    textOperation
}

func diffText(oldText, newText string) *TextDiff {
	if oldText == newText {
		return nil
	}
	t1 := []rune(oldText)
	t2 := []rune(newText)
	oldLen := len(t1)
	newLen := len(t2)
	if oldLen == 0 {
		return &TextDiff{previousText: oldText, newText: newText, start: 0, end: 0, operation: textOperationAdd}
	}
	if newLen == 0 {
		return &TextDiff{previousText: oldText, newText: "", start: 0, end: oldLen, operation: textOperationReplace}
	}

	// if we reach here, t1 and t2 are not empty
	start := 0
	for start < oldLen && start < newLen && t1[start] == t2[start] {
		start++
	}

	oldEnd, newEnd := oldLen, newLen
	for oldEnd > start && newEnd > start && t1[oldEnd-1] == t2[newEnd-1] {
		oldEnd--
		newEnd--
	}

	diff := &TextDiff{previousText: oldText, start: start}
	if newLen > oldLen && (start == oldLen || oldEnd == 0 || start == oldEnd) {
		diff.operation = textOperationAdd
		diff.end = start
		diff.newText = string(t2[start:newEnd])
	} else if newLen < oldLen && (start == newLen || newEnd == 0 || start == newEnd) {
		diff.operation = textOperationDelete
		diff.end = oldEnd - 1
	} else {
		diff.operation = textOperationReplace
		if start == 0 && oldEnd == oldLen { // full replace
			diff.end = oldLen - 1
			diff.newText = newText
		} else { // partial replace
			diff.end = oldEnd - 1
			diff.newText = string(t2[start:newEnd])
		}
	}
	return diff
}
