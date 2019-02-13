package token

import "fmt"

var SingleByteTokens = [...]string{"", "", "", "200", "400", "404", "500", "501", "502", "action", "add",
	"after", "archive", "author", "available", "battery", "before", "body",
	"broadcast", "chat", "clear", "code", "composing", "contacts", "count",
	"create", "debug", "delete", "demote", "duplicate", "encoding", "error",
	"false", "filehash", "from", "g.us", "group", "groups_v2", "height", "id",
	"image", "in", "index", "invis", "item", "jid", "kind", "last", "leave",
	"live", "log", "media", "message", "mimetype", "missing", "modify", "name",
	"notification", "notify", "out", "owner", "participant", "paused",
	"picture", "played", "presence", "preview", "promote", "query", "raw",
	"read", "receipt", "received", "recipient", "recording", "relay",
	"remove", "response", "resume", "retry", "s.whatsapp.net", "seconds",
	"set", "size", "status", "subject", "subscribe", "t", "text", "to", "true",
	"type", "unarchive", "unavailable", "url", "user", "value", "web", "width",
	"mute", "read_only", "admin", "creator", "short", "update", "powersave",
	"checksum", "epoch", "block", "previous", "409", "replaced", "reason",
	"spam", "modify_tag", "message_info", "delivery", "emoji", "title",
	"description", "canonical-url", "matched-text", "star", "unstar",
	"media_key", "filename", "identity", "unread", "page", "page_count",
	"search", "media_message", "security", "call_log", "profile", "ciphertext",
	"invite", "gif", "vcard", "frequent", "privacy", "blacklist", "whitelist",
	"verify", "location", "document", "elapsed", "revoke_invite", "expiration",
	"unsubscribe", "disable", "vname", "old_jid", "new_jid", "announcement",
	"locked", "prop", "label", "color", "call", "offer", "call-id"}

var doubleByteTokens = [...]string{}

func GetSingleToken(i int) (string, error) {
	if i < 3 || i >= len(SingleByteTokens) {
		return "", fmt.Errorf("index out of single byte token bounds %d", i)
	}

	return SingleByteTokens[i], nil
}

func GetDoubleToken(index1 int, index2 int) (string, error) {
	n := 256*index1 + index2
	if n < 0 || n >= len(doubleByteTokens) {
		return "", fmt.Errorf("index out of double byte token bounds %d", n)
	}

	return doubleByteTokens[n], nil
}

func IndexOfSingleToken(token string) int {
	for i, t := range SingleByteTokens {
		if t == token {
			return i
		}
	}

	return -1
}

const (
	LIST_EMPTY   = 0
	STREAM_END   = 2
	DICTIONARY_0 = 236
	DICTIONARY_1 = 237
	DICTIONARY_2 = 238
	DICTIONARY_3 = 239
	LIST_8       = 248
	LIST_16      = 249
	JID_PAIR     = 250
	HEX_8        = 251
	BINARY_8     = 252
	BINARY_20    = 253
	BINARY_32    = 254
	NIBBLE_8     = 255
)

const (
	PACKED_MAX      = 254
	SINGLE_BYTE_MAX = 256
)
