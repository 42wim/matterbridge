package requests

import "errors"

var ErrLoginInvalidKeyUID = errors.New("login: invalid key-uid")

type Login struct {
	Password      string `json:"password"`
	KeyUID        string `json:"keyUid"`
	KdfIterations int    `json:"kdfIterations"`

	WakuV2Nameserver string `json:"wakuV2Nameserver"`

	WalletSecretsConfig
}

func (c *Login) Validate() error {
	if c.KeyUID == "" {
		return ErrLoginInvalidKeyUID
	}
	return nil
}
