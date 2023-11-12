package requests

import (
	"errors"
)

var ErrCreateAccountInvalidDisplayName = errors.New("create-account: invalid display name")
var ErrCreateAccountInvalidPassword = errors.New("create-account: invalid password")
var ErrCreateAccountInvalidCustomizationColor = errors.New("create-account: invalid customization color")
var ErrCreateAccountInvalidRootKeystoreDir = errors.New("create-account: invalid root keystore directory")
var ErrCreateAccountInvalidBackupDisabledDataDir = errors.New("create-account: invalid backup disabled data directory")
var ErrCreateAccountInvalidLogFilePath = errors.New("create-account: invalid log file path")

type CreateAccount struct {
	// BackupDisabledDataDir is the directory where backup is disabled
	BackupDisabledDataDir string `json:"backupDisabledDataDir"`

	DeviceName         string `json:"deviceName"`
	DisplayName        string `json:"displayName"`
	Password           string `json:"password"`
	ImagePath          string `json:"imagePath"`
	CustomizationColor string `json:"customizationColor"`
	Emoji              string `json:"emoji"`

	WakuV2Nameserver  *string `json:"wakuV2Nameserver"`
	WakuV2LightClient bool    `json:"wakuV2LightClient"`

	LogLevel    *string `json:"logLevel"`
	LogFilePath string  `json:"logFilePath"`
	LogEnabled  bool    `json:"logEnabled"`

	PreviewPrivacy bool `json:"previewPrivacy"`

	VerifyTransactionURL     *string `json:"verifyTransactionURL"`
	VerifyENSURL             *string `json:"verifyENSURL"`
	VerifyENSContractAddress *string `json:"verifyENSContractAddress"`
	VerifyTransactionChainID *int64  `json:"verifyTransactionChainID"`
	UpstreamConfig           string  `json:"upstreamConfig"`

	CurrentNetwork string `json:"currentNetwork"`
	NetworkID      uint64 `json:"networkId"`

	WalletSecretsConfig
}

type WalletSecretsConfig struct {
	PoktToken            string `json:"poktToken"`
	InfuraToken          string `json:"infuraToken"`
	InfuraSecret         string `json:"infuraSecret"`
	OpenseaAPIKey        string `json:"openseaApiKey"`
	RaribleMainnetAPIKey string `json:"raribleMainnetApiKey"`
	RaribleTestnetAPIKey string `json:"raribleTestnetApiKey"`

	// Testing
	GanacheURL                  string `json:"ganacheURL"`
	AlchemyEthereumMainnetToken string `json:"alchemyEthereumMainnetToken"`
	AlchemyEthereumGoerliToken  string `json:"alchemyEthereumGoerliToken"`
	AlchemyEthereumSepoliaToken string `json:"alchemyEthereumSepoliaToken"`
	AlchemyArbitrumMainnetToken string `json:"alchemyArbitrumMainnetToken"`
	AlchemyArbitrumGoerliToken  string `json:"alchemyArbitrumGoerliToken"`
	AlchemyArbitrumSepoliaToken string `json:"alchemyArbitrumSepoliaToken"`
	AlchemyOptimismMainnetToken string `json:"alchemyOptimismMainnetToken"`
	AlchemyOptimismGoerliToken  string `json:"alchemyOptimismGoerliToken"`
	AlchemyOptimismSepoliaToken string `json:"alchemyOptimismSepoliaToken"`
}

func (c *CreateAccount) Validate() error {
	return ValidateAccountCreationRequest(*c)
}

func ValidateAccountCreationRequest(c CreateAccount) error {
	// TODO(cammellos): Add proper validation for password/displayname/etc
	if len(c.DisplayName) == 0 {
		return ErrCreateAccountInvalidDisplayName
	}

	if len(c.Password) == 0 {
		return ErrCreateAccountInvalidPassword
	}

	if len(c.CustomizationColor) == 0 {
		return ErrCreateAccountInvalidCustomizationColor
	}

	if len(c.BackupDisabledDataDir) == 0 {
		return ErrCreateAccountInvalidBackupDisabledDataDir
	}

	if len(c.LogFilePath) == 0 {
		return ErrCreateAccountInvalidLogFilePath
	}

	return nil

}
