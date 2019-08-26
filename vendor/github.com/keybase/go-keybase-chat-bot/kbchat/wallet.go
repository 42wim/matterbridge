package kbchat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type WalletOutput struct {
	Result WalletResult `json:"result"`
}

type WalletResult struct {
	TxID         string      `json:"txID"`
	Status       string      `json:"status"`
	Amount       string      `json:"amount"`
	Asset        WalletAsset `json:"asset"`
	FromUsername string      `json:"fromUsername"`
	ToUsername   string      `json:"toUsername"`
}

type WalletAsset struct {
	Type   string `json:"type"`
	Code   string `json:"code"`
	Issuer string `json:"issuer"`
}

func (a *API) GetWalletTxDetails(txID string) (wOut WalletOutput, err error) {
	a.Lock()
	defer a.Unlock()

	apiInput := fmt.Sprintf(`{"method": "details", "params": {"options": {"txid": "%s"}}}`, txID)
	cmd := a.runOpts.Command("wallet", "api")
	cmd.Stdin = strings.NewReader(apiInput)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return wOut, err
	}

	if err := json.Unmarshal(out.Bytes(), &wOut); err != nil {
		return wOut, fmt.Errorf("unable to decode wallet output: %s", err.Error())
	}

	return wOut, nil
}
