package requests

import (
	"errors"
)

var ErrRestoreAccountInvalidMnemonic = errors.New("restore-account: invalid mnemonic")

type RestoreAccount struct {
	Mnemonic string `json:"mnemonic"`
	CreateAccount
}

func (c *RestoreAccount) Validate() error {
	if len(c.Mnemonic) == 0 {
		return ErrRestoreAccountInvalidMnemonic
	}

	return ValidateAccountCreationRequest(c.CreateAccount)
}
