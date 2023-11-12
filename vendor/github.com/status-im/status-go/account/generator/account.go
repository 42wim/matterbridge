package generator

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"time"

	accountJson "github.com/status-im/status-go/account/json"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/extkeys"
	"github.com/status-im/status-go/multiaccounts"
)

type Account struct {
	privateKey  *ecdsa.PrivateKey
	extendedKey *extkeys.ExtendedKey
}

func NewAccount(privateKey *ecdsa.PrivateKey, extKey *extkeys.ExtendedKey) Account {
	if privateKey == nil {
		privateKey = extKey.ToECDSA()
	}

	return Account{
		privateKey:  privateKey,
		extendedKey: extKey,
	}
}

func (a *Account) ToAccountInfo() AccountInfo {
	privateKeyHex := types.EncodeHex(crypto.FromECDSA(a.privateKey))
	publicKeyHex := types.EncodeHex(crypto.FromECDSAPub(&a.privateKey.PublicKey))
	addressHex := crypto.PubkeyToAddress(a.privateKey.PublicKey).Hex()

	return AccountInfo{
		PrivateKey: privateKeyHex,
		PublicKey:  publicKeyHex,
		Address:    addressHex,
	}
}

func (a *Account) ToIdentifiedAccountInfo(id string) IdentifiedAccountInfo {
	info := a.ToAccountInfo()
	keyUID := sha256.Sum256(crypto.FromECDSAPub(&a.privateKey.PublicKey))
	keyUIDHex := types.EncodeHex(keyUID[:])
	return IdentifiedAccountInfo{
		AccountInfo: info,
		ID:          id,
		KeyUID:      keyUIDHex,
	}
}

func (a *Account) ToGeneratedAccountInfo(id string, mnemonic string) GeneratedAccountInfo {
	idInfo := a.ToIdentifiedAccountInfo(id)
	return GeneratedAccountInfo{
		IdentifiedAccountInfo: idInfo,
		Mnemonic:              mnemonic,
	}
}

// AccountInfo contains a PublicKey and an Address of an account.
type AccountInfo struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
	Address    string `json:"address"`
}

func (a AccountInfo) MarshalJSON() ([]byte, error) {
	type Alias AccountInfo
	ext, err := accountJson.ExtendStructWithPubKeyData(a.PublicKey, Alias(a))
	if err != nil {
		return nil, err
	}
	return json.Marshal(ext)
}

// IdentifiedAccountInfo contains AccountInfo and the ID of an account.
type IdentifiedAccountInfo struct {
	AccountInfo
	ID string `json:"id"`
	// KeyUID is calculated as sha256 of the master public key and used for key
	// identification. This is the only information available about the master
	// key stored on a keycard before the card is paired.
	// KeyUID name is chosen over KeyID in order to make it consistent with
	// the name already used in Status and Keycard codebases.
	KeyUID string `json:"keyUid"`
}

func (i IdentifiedAccountInfo) MarshalJSON() ([]byte, error) {
	accountInfoJSON, err := i.AccountInfo.MarshalJSON()
	if err != nil {
		return nil, err
	}
	type info struct {
		ID     string `json:"id"`
		KeyUID string `json:"keyUid"`
	}
	infoJSON, err := json.Marshal(info{
		ID:     i.ID,
		KeyUID: i.KeyUID,
	})
	if err != nil {
		return nil, err
	}
	infoJSON[0] = ','
	return append(accountInfoJSON[:len(accountInfoJSON)-1], infoJSON...), nil
}

func (i *IdentifiedAccountInfo) ToMultiAccount() *multiaccounts.Account {
	return &multiaccounts.Account{
		Timestamp: time.Now().Unix(),
		KeyUID:    i.KeyUID,
	}
}

// GeneratedAccountInfo contains IdentifiedAccountInfo and the mnemonic of an account.
type GeneratedAccountInfo struct {
	IdentifiedAccountInfo
	Mnemonic string `json:"mnemonic"`
}

func (g GeneratedAccountInfo) MarshalJSON() ([]byte, error) {
	accountInfoJSON, err := g.IdentifiedAccountInfo.MarshalJSON()
	if err != nil {
		return nil, err
	}
	type info struct {
		Mnemonic string `json:"mnemonic"`
	}
	infoJSON, err := json.Marshal(info{
		Mnemonic: g.Mnemonic,
	})
	if err != nil {
		return nil, err
	}
	infoJSON[0] = ','
	return append(accountInfoJSON[:len(accountInfoJSON)-1], infoJSON...), nil
}

func (g GeneratedAccountInfo) toGeneratedAndDerived(derived map[string]AccountInfo) GeneratedAndDerivedAccountInfo {
	return GeneratedAndDerivedAccountInfo{
		GeneratedAccountInfo: g,
		Derived:              derived,
	}
}

// GeneratedAndDerivedAccountInfo contains GeneratedAccountInfo and derived AccountInfo mapped by derivation path.
type GeneratedAndDerivedAccountInfo struct {
	GeneratedAccountInfo
	Derived map[string]AccountInfo `json:"derived"`
}

func (g GeneratedAndDerivedAccountInfo) MarshalJSON() ([]byte, error) {
	accountInfoJSON, err := g.GeneratedAccountInfo.MarshalJSON()
	if err != nil {
		return nil, err
	}
	type info struct {
		Derived map[string]AccountInfo `json:"derived"`
	}
	infoJSON, err := json.Marshal(info{
		Derived: g.Derived,
	})
	if err != nil {
		return nil, err
	}
	infoJSON[0] = ','
	return append(accountInfoJSON[:len(accountInfoJSON)-1], infoJSON...), nil
}
