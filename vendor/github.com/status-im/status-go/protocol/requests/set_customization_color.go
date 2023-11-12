package requests

import (
	"errors"

	"github.com/status-im/status-go/multiaccounts/common"
)

var ErrSetCustomizationColorInvalidColor = errors.New("customizationColor: invalid color")
var ErrSetCustomizationColorInvalidKeyUID = errors.New("keyUid: invalid id")

type SetCustomizationColor struct {
	CustomizationColor common.CustomizationColor `json:"customizationColor"`
	KeyUID             string                    `json:"keyUid"`
}

func (a *SetCustomizationColor) Validate() error {
	if len(a.CustomizationColor) == 0 {
		return ErrSetCustomizationColorInvalidColor
	}

	if len(a.KeyUID) == 0 {
		return ErrSetCustomizationColorInvalidKeyUID
	}

	return nil
}
