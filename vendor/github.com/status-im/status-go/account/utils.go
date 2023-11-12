package account

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/extkeys"
	"github.com/status-im/status-go/multiaccounts"
)

// errors
var (
	ErrInvalidAccountAddressOrKey  = errors.New("cannot parse address or key to valid account address")
	ErrInvalidMnemonicPhraseLength = errors.New("invalid mnemonic phrase length; valid lengths are 12, 15, 18, 21, and 24")
)

type LoginParams struct {
	ChatAddress    types.Address          `json:"chatAddress"`
	Password       string                 `json:"password"`
	MainAccount    types.Address          `json:"mainAccount"`
	WatchAddresses []types.Address        `json:"watchAddresses"`
	MultiAccount   *multiaccounts.Account `json:"multiAccount"`
}

type ErrZeroAddress struct {
	field string
}

func (e *ErrZeroAddress) Error() string {
	return fmt.Sprintf("%s contains an empty address", e.field)
}

func newErrZeroAddress(field string) *ErrZeroAddress {
	return &ErrZeroAddress{
		field: field,
	}
}

func ParseLoginParams(paramsJSON string) (LoginParams, error) {
	var (
		params      LoginParams
		zeroAddress types.Address
	)
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return params, err
	}

	if params.ChatAddress == zeroAddress {
		return params, newErrZeroAddress("ChatAddress")
	}

	if params.MainAccount == zeroAddress {
		return params, newErrZeroAddress("MainAccount")
	}

	for _, address := range params.WatchAddresses {
		if address == zeroAddress {
			return params, newErrZeroAddress("WatchAddresses")
		}
	}
	return params, nil
}

// Info contains wallet and chat addresses and public keys of an account.
type Info struct {
	WalletAddress string
	WalletPubKey  string
	ChatAddress   string
	ChatPubKey    string
}

// SelectedExtKey is a container for the selected (logged in) external account.
type SelectedExtKey struct {
	Address     types.Address
	AccountKey  *types.Key
	SubAccounts []types.Account
}

// Hex dumps address of a given extended key as hex string.
func (k *SelectedExtKey) Hex() string {
	if k == nil {
		return "0x0"
	}

	return k.Address.Hex()
}

// ParseAccountString parses hex encoded string and returns is as types.Account.
func ParseAccountString(account string) (types.Account, error) {
	// valid address, convert to account
	if types.IsHexAddress(account) {
		return types.Account{Address: types.HexToAddress(account)}, nil
	}

	return types.Account{}, ErrInvalidAccountAddressOrKey
}

// FromAddress converts account address from string to types.Address.
// The function is useful to format "From" field of send transaction struct.
func FromAddress(accountAddress string) types.Address {
	from, err := ParseAccountString(accountAddress)
	if err != nil {
		return types.Address{}
	}

	return from.Address
}

// ToAddress converts account address from string to *common.Address.
// The function is useful to format "To" field of send transaction struct.
func ToAddress(accountAddress string) *types.Address {
	to, err := ParseAccountString(accountAddress)
	if err != nil {
		return nil
	}

	return &to.Address
}

func GetRandomMnemonic() (string, error) {
	// generate mnemonic phrase
	mn := extkeys.NewMnemonic()
	mnemonic, err := mn.MnemonicPhrase(extkeys.EntropyStrength128, extkeys.EnglishLanguage)
	if err != nil {
		return "", fmt.Errorf("can not create mnemonic seed: %v", err)
	}
	return mnemonic, nil
}
