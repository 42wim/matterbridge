package images

import (
	"encoding/json"
	"errors"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/protocol/protobuf"
)

type IdentityImage struct {
	KeyUID       string `json:"keyUID"`
	Name         string `json:"name"`
	Payload      []byte `json:"payload"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     int    `json:"fileSize"`
	ResizeTarget int    `json:"resizeTarget"`
	Clock        uint64 `json:"clock"`
	LocalURL     string `json:"localUrl,omitempty"`
}

func (i IdentityImage) GetType() (ImageType, error) {
	it := GetType(i.Payload)
	if it == UNKNOWN {
		return it, errors.New("unsupported file type")
	}

	return it, nil
}

func (i IdentityImage) Hash() []byte {
	return crypto.Keccak256(i.Payload)
}

func (i IdentityImage) GetDataURI() (string, error) {
	return GetPayloadDataURI(i.Payload)
}

func (i IdentityImage) MarshalJSON() ([]byte, error) {
	uri, err := i.GetDataURI()
	if err != nil {
		return nil, err
	}

	temp := struct {
		KeyUID       string `json:"keyUid"`
		Name         string `json:"type"`
		URI          string `json:"uri"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		FileSize     int    `json:"fileSize"`
		ResizeTarget int    `json:"resizeTarget"`
		Clock        uint64 `json:"clock"`
		LocalURL     string `json:"localUrl,omitempty"`
	}{
		KeyUID:       i.KeyUID,
		Name:         i.Name,
		URI:          uri,
		Width:        i.Width,
		Height:       i.Height,
		FileSize:     i.FileSize,
		ResizeTarget: i.ResizeTarget,
		Clock:        i.Clock,
		LocalURL:     i.LocalURL,
	}

	return json.Marshal(temp)
}

func (i *IdentityImage) ToProtobuf() *protobuf.MultiAccount_IdentityImage {
	return &protobuf.MultiAccount_IdentityImage{
		KeyUid:       i.KeyUID,
		Name:         i.Name,
		Payload:      i.Payload,
		Width:        int64(i.Width),
		Height:       int64(i.Height),
		Filesize:     int64(i.FileSize),
		ResizeTarget: int64(i.ResizeTarget),
		Clock:        i.Clock,
	}
}

func (i *IdentityImage) FromProtobuf(ii *protobuf.MultiAccount_IdentityImage) {
	i.KeyUID = ii.KeyUid
	i.Name = ii.Name
	i.Payload = ii.Payload
	i.Width = int(ii.Width)
	i.Height = int(ii.Height)
	i.FileSize = int(ii.Filesize)
	i.ResizeTarget = int(ii.ResizeTarget)
	i.Clock = ii.Clock
}

func (i IdentityImage) IsEmpty() bool {
	return i.KeyUID == "" && i.Name == "" && len(i.Payload) == 0 && i.Width == 0 && i.Height == 0 && i.FileSize == 0 && i.ResizeTarget == 0 && i.Clock == 0
}
