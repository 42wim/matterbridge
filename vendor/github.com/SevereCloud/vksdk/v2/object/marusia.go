package object // import "github.com/SevereCloud/vksdk/v2/object"

import (
	"encoding/json"
)

// MarusiaPicture struct.
type MarusiaPicture struct {
	ID      int `json:"id"`
	OwnerID int `json:"owner_id"`
}

// MarusiaPictureUploadResponse struct.
type MarusiaPictureUploadResponse struct {
	Hash        string          `json:"hash"`   // Uploading hash
	Photo       json.RawMessage `json:"photo"`  // Uploaded photo data
	Server      int             `json:"server"` // Upload server number
	AID         int             `json:"aid"`
	MessageCode int             `json:"message_code"`
}

// MarusiaAudio struct.
type MarusiaAudio struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	OwnerID int    `json:"owner_id"`
}

// MarusiaAudioUploadResponse struct.
type MarusiaAudioUploadResponse struct {
	Sha       string           `json:"sha"`
	Secret    string           `json:"secret"`
	Meta      MarusiaAudioMeta `json:"meta"`
	Hash      string           `json:"hash"`
	Server    string           `json:"server"`
	UserID    int              `json:"user_id"`
	RequestID string           `json:"request_id"`
}

// MarusiaAudioMeta struct.
type MarusiaAudioMeta struct {
	Album       string `json:"album"`
	Artist      string `json:"artist"`
	Bitrate     string `json:"bitrate"`
	Duration    string `json:"duration"`
	Genre       string `json:"genre"`
	Kad         string `json:"kad"`
	Md5         string `json:"md5"`
	Md5DataSize string `json:"md5_data_size"`
	Samplerate  string `json:"samplerate"`
	Title       string `json:"title"`
}
