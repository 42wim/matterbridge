package keystore

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/waku-org/go-zerokit-rln/rln"
	"go.uber.org/zap"
)

// MembershipContractInfo contains information about a membership smart contract address and the chain in which it is deployed
type MembershipContractInfo struct {
	ChainID ChainID         `json:"chainId"`
	Address ContractAddress `json:"address"`
}

// NewMembershipContractInfo generates a new MembershipContract instance
func NewMembershipContractInfo(chainID *big.Int, address common.Address) MembershipContractInfo {
	return MembershipContractInfo{
		ChainID: ChainID{
			chainID,
		},
		Address: ContractAddress(address),
	}
}

// ContractAddress is a common.Address created to comply with the expected marshalling for the credentials
type ContractAddress common.Address

// MarshalText is used to convert a ContractAddress into a valid value expected by the json encoder
func (c ContractAddress) MarshalText() ([]byte, error) {
	return []byte(common.Address(c).Hex()), nil
}

// UnmarshalText converts a byte slice into a ContractAddress
func (c *ContractAddress) UnmarshalText(text []byte) error {
	b, err := hexutil.Decode(string(text))
	if err != nil {
		return err
	}
	copy(c[:], b[:])
	return nil
}

// ChainID is a helper struct created to comply with the expected marshalling for the credentials
type ChainID struct {
	*big.Int
}

// String returns a string with the expected chainId format for the credentials
func (c ChainID) String() string {
	return fmt.Sprintf(`"%s"`, hexutil.EncodeBig(c.Int))
}

// MarshalJSON is used to convert a ChainID into a valid value expected by the json encoder
func (c ChainID) MarshalJSON() (text []byte, err error) {
	return []byte(c.String()), nil
}

// UnmarshalJSON converts a byte slice into a ChainID
func (c *ChainID) UnmarshalJSON(text []byte) error {
	hexVal := strings.ReplaceAll(string(text), `"`, "")
	b, err := hexutil.DecodeBig(hexVal)
	if err != nil {
		return err
	}

	c.Int = b
	return nil
}

// Equals is used to compare MembershipContract
func (m MembershipContractInfo) Equals(other MembershipContractInfo) bool {
	return m.Address == other.Address && m.ChainID.Int64() == other.ChainID.Int64()
}

// MembershipCredentials contains all the information about an RLN Identity Credential and membership group it belongs to
type MembershipCredentials struct {
	IdentityCredential     *rln.IdentityCredential `json:"identityCredential"`
	MembershipContractInfo MembershipContractInfo  `json:"membershipContract"`
	TreeIndex              rln.MembershipIndex     `json:"treeIndex"`
}

// Equals is used to compare MembershipCredentials
func (m MembershipCredentials) Equals(other MembershipCredentials) bool {
	return rln.IdentityCredentialEquals(*m.IdentityCredential, *other.IdentityCredential) && m.MembershipContractInfo.Equals(other.MembershipContractInfo) && m.TreeIndex == other.TreeIndex
}

// AppInfo is a helper structure that contains information about the application that uses these credentials
type AppInfo struct {
	Application   string `json:"application"`
	AppIdentifier string `json:"appIdentifier"`
	Version       string `json:"version"`
}

// Key is a helper type created to represent the key in a map of credentials
type Key string

// AppKeystore represents the membership credentials to be used in RLN
type AppKeystore struct {
	Application   string                        `json:"application"`
	AppIdentifier string                        `json:"appIdentifier"`
	Credentials   map[Key]appKeystoreCredential `json:"credentials"`
	Version       string                        `json:"version"`

	path   string
	logger *zap.Logger
}

type appKeystoreCredential struct {
	Crypto keystore.CryptoJSON `json:"crypto"`
}

const defaultSeparator = "\n"
