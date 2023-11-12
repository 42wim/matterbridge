package wallet

import (
	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/services/wallet/thirdparty/fourbyte"
	"github.com/status-im/status-go/services/wallet/thirdparty/fourbytegithub"
)

type Decoder struct {
	Main     *fourbytegithub.Client
	Fallback *fourbyte.Client
}

func NewDecoder() *Decoder {
	return &Decoder{
		Main:     fourbytegithub.NewClient(),
		Fallback: fourbyte.NewClient(),
	}
}

func (d *Decoder) Decode(data string) (*thirdparty.DataParsed, error) {
	parsed, err := d.Main.Run(data)
	if err == nil {
		return parsed, nil
	}

	return d.Fallback.Run(data)
}
