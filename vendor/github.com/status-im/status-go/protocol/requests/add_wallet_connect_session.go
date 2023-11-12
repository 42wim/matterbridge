package requests

import (
	"errors"
)

var ErrAddWalletConnectSessionInvalidID = errors.New("add-wallet-connect-session: invalid id")
var ErrAddWalletConnectSessionInvalidDAppName = errors.New("add-wallet-connect-session: invalid dapp name")
var ErrAddWalletConnectSessionInvalidDAppURL = errors.New("add-wallet-connect-session: invalid dapp url")
var ErrAddWalletConnectSessionInvalidInfo = errors.New("add-wallet-connect-session: invalid info")

type AddWalletConnectSession struct {
	PeerID   string `json:"id"`
	DAppName string `json:"dappName"`
	DAppURL  string `json:"dappUrl"`
	Info     string `json:"info"`
}

func (c *AddWalletConnectSession) Validate() error {
	if len(c.PeerID) == 0 {
		return ErrAddWalletConnectSessionInvalidID
	}

	if len(c.DAppName) == 0 {
		return ErrAddWalletConnectSessionInvalidDAppName
	}

	if len(c.DAppURL) == 0 {
		return ErrAddWalletConnectSessionInvalidDAppURL
	}

	if len(c.Info) == 0 {
		return ErrAddWalletConnectSessionInvalidInfo
	}

	return nil
}
